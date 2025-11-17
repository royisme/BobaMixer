package proxy

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/royisme/bobamixer/internal/logging"
	"github.com/royisme/bobamixer/internal/store/sqlite"
)

const (
	providerAnthropic = "anthropic"
)

// Handler handles HTTP proxy requests
type Handler struct {
	db    *sqlite.DB
	stats *Stats
}

// Stats tracks proxy statistics
type Stats struct {
	TotalRequests     int64
	OpenAIRequests    int64
	AnthropicRequests int64
	ErrorCount        int64
	BytesProxied      int64
	LastRequest       time.Time
	mu                sync.RWMutex
}

// NewHandler creates a new proxy handler
func NewHandler(dbPath string) (*Handler, error) {
	db, err := sqlite.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	return &Handler{
		db:    db,
		stats: &Stats{},
	}, nil
}

// ServeHTTP implements http.Handler
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Handle health check endpoint
	if r.URL.Path == "/health" {
		h.handleHealth(w, r)
		return
	}

	// Update stats
	h.stats.mu.Lock()
	h.stats.TotalRequests++
	h.stats.LastRequest = startTime
	h.stats.mu.Unlock()

	// Parse provider type from path
	providerType, targetPath := h.parseRoute(r.URL.Path)
	if providerType == "" {
		http.Error(w, "Invalid proxy route", http.StatusBadRequest)
		h.incrementErrorCount()
		return
	}

	// Update provider-specific stats
	h.updateProviderStats(providerType)

	// Get target base URL from request headers or configuration
	targetURL := h.getTargetURL(r, providerType)
	if targetURL == "" {
		http.Error(w, "No target URL configured", http.StatusBadRequest)
		h.incrementErrorCount()
		return
	}

	// Forward the request
	if err := h.forwardRequest(w, r, targetURL, targetPath, providerType, startTime); err != nil {
		logging.Error("Failed to forward request",
			logging.String("error", err.Error()),
			logging.String("provider", providerType),
			logging.String("path", targetPath))
		h.incrementErrorCount()
	}
}

// parseRoute extracts provider type and target path from proxy URL
func (h *Handler) parseRoute(path string) (providerType, targetPath string) {
	// Expected format: /openai/v1/* or /anthropic/v1/*
	parts := strings.SplitN(strings.TrimPrefix(path, "/"), "/", 2)
	if len(parts) < 2 {
		return "", ""
	}

	providerType = parts[0]
	targetPath = "/" + parts[1]

	// Validate provider type
	if providerType != "openai" && providerType != providerAnthropic {
		return "", ""
	}

	return providerType, targetPath
}

// getTargetURL determines the upstream API URL
func (h *Handler) getTargetURL(r *http.Request, providerType string) string {
	// Check for custom header first
	if target := r.Header.Get("X-Proxy-Target"); target != "" {
		return target
	}

	// Default upstream URLs
	switch providerType {
	case "openai":
		return "https://api.openai.com"
	case providerAnthropic:
		return "https://api.anthropic.com"
	default:
		return ""
	}
}

// forwardRequest forwards the request to the upstream provider
func (h *Handler) forwardRequest(w http.ResponseWriter, r *http.Request, targetURL, targetPath, providerType string, startTime time.Time) error {
	// Read request body
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return fmt.Errorf("read body: %w", err)
	}
	defer func() {
		if cerr := r.Body.Close(); cerr != nil {
			h.stats.ErrorCount++
		}
	}()

	// Build upstream URL
	upstreamURL := targetURL + targetPath
	if r.URL.RawQuery != "" {
		upstreamURL += "?" + r.URL.RawQuery
	}

	// Create upstream request
	upstreamReq, err := http.NewRequestWithContext(r.Context(), r.Method, upstreamURL, bytes.NewReader(bodyBytes))
	if err != nil {
		http.Error(w, "Failed to create upstream request", http.StatusInternalServerError)
		return fmt.Errorf("create request: %w", err)
	}

	// Copy headers (except hop-by-hop headers)
	h.copyHeaders(upstreamReq.Header, r.Header)

	// Remove proxy-specific headers
	upstreamReq.Header.Del("X-Proxy-Target")
	upstreamReq.Header.Del("X-Tool-ID")

	// Send request
	client := &http.Client{
		Timeout: 60 * time.Second,
	}
	resp, err := client.Do(upstreamReq)
	if err != nil {
		http.Error(w, "Failed to reach upstream provider", http.StatusBadGateway)
		return fmt.Errorf("do request: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			h.stats.ErrorCount++
		}
	}()

	// Read response body
	respBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read upstream response", http.StatusBadGateway)
		return fmt.Errorf("read response: %w", err)
	}

	// Log request/response
	h.logRequest(r, providerType, targetPath, bodyBytes, respBodyBytes, resp.StatusCode, startTime)

	// Update bytes proxied
	h.stats.mu.Lock()
	h.stats.BytesProxied += int64(len(bodyBytes) + len(respBodyBytes))
	h.stats.mu.Unlock()

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Write response
	w.WriteHeader(resp.StatusCode)
	if _, err := w.Write(respBodyBytes); err != nil {
		return fmt.Errorf("write response: %w", err)
	}

	return nil
}

// copyHeaders copies HTTP headers, excluding hop-by-hop headers
func (h *Handler) copyHeaders(dst, src http.Header) {
	// Hop-by-hop headers that should not be copied
	hopByHop := map[string]bool{
		"Connection":          true,
		"Keep-Alive":          true,
		"Proxy-Authenticate":  true,
		"Proxy-Authorization": true,
		"Te":                  true,
		"Trailers":            true,
		"Transfer-Encoding":   true,
		"Upgrade":             true,
	}

	for key, values := range src {
		if !hopByHop[key] {
			for _, value := range values {
				dst.Add(key, value)
			}
		}
	}
}

// logRequest logs the proxied request to the database
func (h *Handler) logRequest(r *http.Request, providerType, path string, reqBody, respBody []byte, statusCode int, startTime time.Time) {
	toolID := r.Header.Get("X-Tool-ID")
	if toolID == "" {
		toolID = "unknown"
	}

	latencyMS := time.Since(startTime).Milliseconds()

	// Log basic metrics (we'll enhance this with token parsing later)
	logging.Info("Proxied request",
		logging.String("tool", toolID),
		logging.String("provider", providerType),
		logging.String("path", path),
		logging.Int("status", statusCode),
		logging.Int("req_bytes", len(reqBody)),
		logging.Int("resp_bytes", len(respBody)),
		logging.Int64("latency_ms", latencyMS))

	// TODO: Parse request/response for token counts and save to usage_records table
	// This will be implemented in Epic 7-4
}

// updateProviderStats updates provider-specific counters
func (h *Handler) updateProviderStats(providerType string) {
	h.stats.mu.Lock()
	defer h.stats.mu.Unlock()

	switch providerType {
	case "openai":
		h.stats.OpenAIRequests++
	case providerAnthropic:
		h.stats.AnthropicRequests++
	}
}

// incrementErrorCount increments the error counter
func (h *Handler) incrementErrorCount() {
	h.stats.mu.Lock()
	h.stats.ErrorCount++
	h.stats.mu.Unlock()
}

// Stats returns current statistics
func (h *Handler) Stats() *Stats {
	h.stats.mu.RLock()
	defer h.stats.mu.RUnlock()

	// Return a copy to avoid race conditions
	return &Stats{
		TotalRequests:     h.stats.TotalRequests,
		OpenAIRequests:    h.stats.OpenAIRequests,
		AnthropicRequests: h.stats.AnthropicRequests,
		ErrorCount:        h.stats.ErrorCount,
		BytesProxied:      h.stats.BytesProxied,
		LastRequest:       h.stats.LastRequest,
	}
}

// handleHealth handles health check requests
func (h *Handler) handleHealth(w http.ResponseWriter, _ *http.Request) {
	stats := h.Stats()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := fmt.Sprintf(`{
  "status": "ok",
  "total_requests": %d,
  "openai_requests": %d,
  "anthropic_requests": %d,
  "error_count": %d,
  "bytes_proxied": %d,
  "last_request": "%s"
}`, stats.TotalRequests, stats.OpenAIRequests, stats.AnthropicRequests,
		stats.ErrorCount, stats.BytesProxied, stats.LastRequest.Format(time.RFC3339))

	if _, err := fmt.Fprint(w, response); err != nil {
		// Log error but don't fail - client may have disconnected
		h.stats.ErrorCount++
	}
}

// ParseProxyURL extracts the target base URL from a proxy-style URL.
func ParseProxyURL(rawURL string) (baseURL string, err error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	// Check if this is a local proxy URL
	if u.Host == "127.0.0.1:7777" || u.Host == "localhost:7777" {
		// Extract provider type from path
		parts := strings.SplitN(strings.TrimPrefix(u.Path, "/"), "/", 2)
		if len(parts) < 1 {
			return "", fmt.Errorf("invalid proxy URL: %s", rawURL)
		}

		providerType := parts[0]
		switch providerType {
		case "openai":
			return "http://127.0.0.1:7777/openai", nil
		case providerAnthropic:
			return "http://127.0.0.1:7777/anthropic", nil
		default:
			return "", fmt.Errorf("unknown provider type: %s", providerType)
		}
	}

	return rawURL, nil
}
