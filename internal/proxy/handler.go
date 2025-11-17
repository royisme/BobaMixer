package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/royisme/bobamixer/internal/domain/budget"
	"github.com/royisme/bobamixer/internal/domain/pricing"
	"github.com/royisme/bobamixer/internal/domain/routing"
	"github.com/royisme/bobamixer/internal/logging"
	"github.com/royisme/bobamixer/internal/store/config"
	"github.com/royisme/bobamixer/internal/store/sqlite"
)

const (
	providerOpenAI    = "openai"
	providerAnthropic = "anthropic"
)

// Handler handles HTTP proxy requests
type Handler struct {
	db            *sqlite.DB
	stats         *Stats
	pricingTable  *pricing.Table
	budgetTracker *budget.Tracker
	routingEngine *routing.Engine
	mu            sync.RWMutex
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

// UsageRecord represents a usage record to be saved
type UsageRecord struct {
	SessionID    string
	Timestamp    int64
	Tool         string
	Model        string
	Provider     string
	InputTokens  int
	OutputTokens int
	InputCost    float64
	OutputCost   float64
	LatencyMS    int64
}

// NewHandler creates a new proxy handler
func NewHandler(dbPath string) (*Handler, error) {
	db, err := sqlite.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Load pricing table
	// For now, use empty table if pricing load fails
	// In production, we might want to fail if pricing is not available
	pricingTable := &pricing.Table{
		Models: make(map[string]pricing.ModelPrice),
	}

	// Initialize budget tracker
	budgetTracker := budget.NewTracker(db)

	// Initialize routing engine (optional, for future use)
	// This allows for dynamic routing based on request content
	var routingEngine *routing.Engine
	// Note: Routing engine initialization would load routes.yaml here
	// For now, we keep it nil and use URL-based routing

	return &Handler{
		db:            db,
		stats:         &Stats{},
		pricingTable:  pricingTable,
		budgetTracker: budgetTracker,
		routingEngine: routingEngine,
	}, nil
}

// SetPricingTable updates the pricing table
func (h *Handler) SetPricingTable(table *pricing.Table) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.pricingTable = table
}

// SetRoutingEngine updates the routing engine
func (h *Handler) SetRoutingEngine(engine *routing.Engine) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.routingEngine = engine
}

// evaluateRouting evaluates routing decision for logging purposes
// This is currently used for debugging and future unified endpoint support
func (h *Handler) evaluateRouting(reqBody []byte) *routing.RoutingDecision {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.routingEngine == nil {
		return nil
	}

	// Extract features from request
	var req map[string]interface{}
	if err := json.Unmarshal(reqBody, &req); err != nil {
		return nil
	}

	// Build routing features
	features := routing.Features{
		Intent:     "api_request",
		TextSample: extractTextSample(req),
		CtxChars:   len(reqBody),
	}

	// Execute routing
	decision, trace, err := h.routingEngine.Match(context.Background(), features)
	if err != nil {
		logging.Warn("Routing evaluation failed", logging.Err(err))
		return nil
	}

	// Log routing decision for debugging
	if trace.Matched {
		logging.Info("Routing decision",
			logging.String("profile", decision.Profile),
			logging.String("rule_id", trace.RuleID),
			logging.String("explain", trace.Explain),
			logging.Bool("explore", decision.Explore))
	}

	return decision
}

// extractTextSample extracts a text sample from the request for routing
func extractTextSample(req map[string]interface{}) string {
	// Try to extract from common fields
	if messages, ok := req["messages"].([]interface{}); ok && len(messages) > 0 {
		if msg, ok := messages[0].(map[string]interface{}); ok {
			if content, ok := msg["content"].(string); ok {
				if len(content) > 200 {
					return content[:200]
				}
				return content
			}
		}
	}

	if prompt, ok := req["prompt"].(string); ok {
		if len(prompt) > 200 {
			return prompt[:200]
		}
		return prompt
	}

	return ""
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
	if providerType != providerOpenAI && providerType != providerAnthropic {
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
	case providerOpenAI:
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

	// Check budget before forwarding request
	if err := h.checkBudgetBeforeRequest(bodyBytes); err != nil {
		http.Error(w, fmt.Sprintf("Budget check failed: %s", err.Error()), http.StatusTooManyRequests)
		logging.Warn("Budget check failed", logging.String("error", err.Error()))
		return fmt.Errorf("budget check: %w", err)
	}

	// Evaluate routing (for debugging and future use)
	// Currently logs routing decisions but doesn't change behavior
	h.evaluateRouting(bodyBytes)

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

	// Parse token usage from response
	model, inputTokens, outputTokens := h.parseTokenUsage(providerType, reqBody, respBody)

	// Calculate cost
	inputCost, outputCost := float64(0), float64(0)
	if model != "" && (inputTokens > 0 || outputTokens > 0) {
		h.mu.RLock()
		profileCost := config.Cost{Input: 0, Output: 0} // Default zero cost
		inputCost, outputCost = h.pricingTable.CalculateCost(model, profileCost, inputTokens, outputTokens)
		h.mu.RUnlock()
	}

	// Log basic metrics
	logging.Info("Proxied request",
		logging.String("tool", toolID),
		logging.String("provider", providerType),
		logging.String("path", path),
		logging.String("model", model),
		logging.Int("status", statusCode),
		logging.Int("input_tokens", inputTokens),
		logging.Int("output_tokens", outputTokens),
		logging.String("input_cost", fmt.Sprintf("%.6f", inputCost)),
		logging.String("output_cost", fmt.Sprintf("%.6f", outputCost)),
		logging.Int("req_bytes", len(reqBody)),
		logging.Int("resp_bytes", len(respBody)),
		logging.Int64("latency_ms", latencyMS))

	// Save to database if we have token information
	if model != "" && (inputTokens > 0 || outputTokens > 0) {
		record := &UsageRecord{
			SessionID:    generateSessionID(),
			Timestamp:    startTime.Unix(),
			Tool:         toolID,
			Model:        model,
			Provider:     providerType,
			InputTokens:  inputTokens,
			OutputTokens: outputTokens,
			InputCost:    inputCost,
			OutputCost:   outputCost,
			LatencyMS:    latencyMS,
		}

		if err := h.saveUsageRecord(record); err != nil {
			logging.Error("Failed to save usage record", logging.Err(err))
		}
	}
}

// parseTokenUsage extracts model and token usage from request/response
func (h *Handler) parseTokenUsage(providerType string, reqBody, respBody []byte) (model string, inputTokens, outputTokens int) {
	switch providerType {
	case providerOpenAI:
		return h.parseOpenAIUsage(reqBody, respBody)
	case providerAnthropic:
		return h.parseAnthropicUsage(reqBody, respBody)
	default:
		return "", 0, 0
	}
}

// parseOpenAIUsage parses OpenAI API response for token usage
func (h *Handler) parseOpenAIUsage(reqBody, respBody []byte) (model string, inputTokens, outputTokens int) {
	// Parse request for model
	var req map[string]interface{}
	if err := json.Unmarshal(reqBody, &req); err == nil {
		if m, ok := req["model"].(string); ok {
			model = m
		}
	}

	// Parse response for usage
	var resp map[string]interface{}
	if err := json.Unmarshal(respBody, &resp); err == nil {
		if usage, ok := resp["usage"].(map[string]interface{}); ok {
			if prompt, ok := usage["prompt_tokens"].(float64); ok {
				inputTokens = int(prompt)
			}
			if completion, ok := usage["completion_tokens"].(float64); ok {
				outputTokens = int(completion)
			}
		}
	}

	return model, inputTokens, outputTokens
}

// parseAnthropicUsage parses Anthropic API response for token usage
func (h *Handler) parseAnthropicUsage(reqBody, respBody []byte) (model string, inputTokens, outputTokens int) {
	// Parse request for model
	var req map[string]interface{}
	if err := json.Unmarshal(reqBody, &req); err == nil {
		if m, ok := req["model"].(string); ok {
			model = m
		}
	}

	// Parse response for usage
	var resp map[string]interface{}
	if err := json.Unmarshal(respBody, &resp); err == nil {
		if usage, ok := resp["usage"].(map[string]interface{}); ok {
			if input, ok := usage["input_tokens"].(float64); ok {
				inputTokens = int(input)
			}
			if output, ok := usage["output_tokens"].(float64); ok {
				outputTokens = int(output)
			}
		}
	}

	return model, inputTokens, outputTokens
}

// saveUsageRecord saves a usage record to the database
func (h *Handler) saveUsageRecord(record *UsageRecord) error {
	// First, ensure session exists
	sessionQuery := fmt.Sprintf(`
		INSERT OR IGNORE INTO sessions (id, started_at, ended_at, success, latency_ms)
		VALUES ('%s', %d, %d, 1, %d);
	`, record.SessionID, record.Timestamp, record.Timestamp+record.LatencyMS/1000, record.LatencyMS)

	if err := h.db.Exec(sessionQuery); err != nil {
		return fmt.Errorf("insert session: %w", err)
	}

	// Insert usage record
	usageQuery := fmt.Sprintf(`
		INSERT INTO usage_records (id, session_id, ts, input_tokens, output_tokens, input_cost, output_cost, tool, model, estimate_level)
		VALUES ('%s', '%s', %d, %d, %d, %.6f, %.6f, '%s', '%s', 'exact');
	`, generateRecordID(), record.SessionID, record.Timestamp,
		record.InputTokens, record.OutputTokens,
		record.InputCost, record.OutputCost,
		escapeSQLString(record.Tool), escapeSQLString(record.Model))

	if err := h.db.Exec(usageQuery); err != nil {
		return fmt.Errorf("insert usage record: %w", err)
	}

	return nil
}

// generateSessionID generates a unique session ID
func generateSessionID() string {
	return fmt.Sprintf("proxy_%d", time.Now().UnixNano())
}

// generateRecordID generates a unique record ID
func generateRecordID() string {
	return fmt.Sprintf("rec_%d", time.Now().UnixNano())
}

// escapeSQLString escapes SQL special characters
func escapeSQLString(s string) string {
	// Simple escape for single quotes
	// In production, use parameterized queries
	return strings.ReplaceAll(s, "'", "''")
}

// checkBudgetBeforeRequest checks if the request would exceed budget limits
func (h *Handler) checkBudgetBeforeRequest(reqBody []byte) error {
	// Parse request to extract model
	var req map[string]interface{}
	if err := json.Unmarshal(reqBody, &req); err != nil {
		// If we can't parse the request, allow it to proceed
		// Budget check is best-effort
		return nil
	}

	model, ok := req["model"].(string)
	if !ok || model == "" {
		// No model specified, allow request
		return nil
	}

	// Estimate token usage based on average request
	// For budget checking, we use conservative estimates:
	// - Average input: 1000 tokens
	// - Average output: 500 tokens
	estimatedInputTokens := 1000
	estimatedOutputTokens := 500

	// Calculate estimated cost
	h.mu.RLock()
	profileCost := config.Cost{Input: 0, Output: 0}
	inputCost, outputCost := h.pricingTable.CalculateCost(model, profileCost, estimatedInputTokens, estimatedOutputTokens)
	h.mu.RUnlock()

	estimatedTotalCost := inputCost + outputCost

	// Check budget (global scope for now)
	// In a real implementation, we might extract project/profile from headers
	allowed, message, err := h.budgetTracker.CheckBudget("global", "", estimatedTotalCost)
	if err != nil {
		// If budget check fails (e.g., no budget configured), allow the request
		logging.Info("Budget check error (allowing request)", logging.Err(err))
		return nil
	}

	if !allowed {
		return fmt.Errorf("%s", message)
	}

	return nil
}

// updateProviderStats updates provider-specific counters
func (h *Handler) updateProviderStats(providerType string) {
	h.stats.mu.Lock()
	defer h.stats.mu.Unlock()

	switch providerType {
	case providerOpenAI:
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
		case providerOpenAI:
			return "http://127.0.0.1:7777/openai", nil
		case providerAnthropic:
			return "http://127.0.0.1:7777/anthropic", nil
		default:
			return "", fmt.Errorf("unknown provider type: %s", providerType)
		}
	}

	return rawURL, nil
}
