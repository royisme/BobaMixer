// Package httpx provides HTTP adapter with retry logic and error classification.
package httpx

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// HTTPRequest represents an HTTP request specification.
type HTTPRequest struct {
	SessionID string
	Endpoint  string
	Headers   map[string]string
	Payload   []byte
	Method    string
	Timeout   time.Duration
	Retries   int // Maximum number of retries (default 2)
}

// Usage represents token usage information.
type Usage struct {
	InputTokens  int
	OutputTokens int
	LatencyMS    int64
	Estimate     string // "exact" | "mapped" | "heuristic"
}

// HTTPResult represents the result of an HTTP request.
type HTTPResult struct {
	Success    bool
	StatusCode int
	Body       []byte
	Usage      Usage
	ErrorClass string // "timeout" | "5xx" | "4xx" | "network" | ""
}

// Execute performs an HTTP request with retry logic and error classification.
// Retries only on: timeout, 5xx errors, and network errors.
// Does NOT retry on: 4xx errors (client errors).
func Execute(ctx context.Context, req HTTPRequest) (*HTTPResult, error) {
	// Set defaults
	if req.Timeout == 0 {
		req.Timeout = 30 * time.Second
	}
	if req.Retries == 0 {
		req.Retries = 2
	}
	if req.Method == "" {
		req.Method = http.MethodPost
	}

	var lastResult *HTTPResult
	maxAttempts := req.Retries + 1 // retries + initial attempt

	for attempt := 0; attempt < maxAttempts; attempt++ {
		// Apply exponential backoff on retries
		if attempt > 0 {
			backoff := time.Duration(attempt) * time.Second
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return &HTTPResult{
					Success:    false,
					ErrorClass: "timeout",
				}, nil
			}
		}

		// Execute single attempt
		result := executeOnce(ctx, req)
		lastResult = result

		// Check if we should retry
		if result.Success {
			return result, nil
		}

		// Determine if we should retry based on error class
		shouldRetry := shouldRetryError(result.ErrorClass)
		if !shouldRetry || attempt >= req.Retries {
			return result, nil
		}
	}

	return lastResult, nil
}

// executeOnce performs a single HTTP request attempt.
func executeOnce(ctx context.Context, req HTTPRequest) *HTTPResult {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: req.Timeout,
	}

	// Create HTTP request
	var bodyReader io.Reader = bytes.NewReader(req.Payload)
	if len(req.Payload) == 0 && (req.Method == http.MethodGet || req.Method == http.MethodHead) {
		bodyReader = http.NoBody
	}

	httpReq, err := http.NewRequestWithContext(ctx, req.Method, req.Endpoint, bodyReader)
	if err != nil {
		return &HTTPResult{
			Success:    false,
			ErrorClass: "network",
		}
	}

	// Set headers
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// Execute request
	start := time.Now()
	resp, err := client.Do(httpReq)
	latencyMS := time.Since(start).Milliseconds()

	// Handle network errors
	if err != nil {
		errorClass := "network"
		if strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "deadline exceeded") {
			errorClass = "timeout"
		}
		return &HTTPResult{
			Success:    false,
			ErrorClass: errorClass,
			Usage: Usage{
				LatencyMS: latencyMS,
				Estimate:  "heuristic",
			},
		}
	}
	defer func() {
		//nolint:errcheck,gosec // Best effort cleanup
		resp.Body.Close()
	}()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &HTTPResult{
			Success:    false,
			StatusCode: resp.StatusCode,
			ErrorClass: "network",
			Usage: Usage{
				LatencyMS: latencyMS,
				Estimate:  "heuristic",
			},
		}
	}

	// Parse usage from response
	usage := parseUsage(body)
	usage.LatencyMS = latencyMS

	// Determine success and error class
	success := resp.StatusCode >= 200 && resp.StatusCode < 300
	errorClass := ""
	if !success {
		errorClass = classifyHTTPError(resp.StatusCode)
	}

	return &HTTPResult{
		Success:    success,
		StatusCode: resp.StatusCode,
		Body:       body,
		Usage:      usage,
		ErrorClass: errorClass,
	}
}

// parseUsage extracts usage information from API response.
// Supports both Anthropic (input_tokens/output_tokens) and OpenAI (prompt_tokens/completion_tokens) formats.
func parseUsage(body []byte) Usage {
	usage := Usage{
		Estimate: "heuristic",
	}

	if len(body) == 0 {
		return usage
	}

	// Try to parse JSON response
	var response struct {
		Usage struct {
			InputTokens      int `json:"input_tokens"`
			OutputTokens     int `json:"output_tokens"`
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return usage
	}

	// Check Anthropic format
	if response.Usage.InputTokens > 0 || response.Usage.OutputTokens > 0 {
		usage.InputTokens = response.Usage.InputTokens
		usage.OutputTokens = response.Usage.OutputTokens
		usage.Estimate = "exact"
		return usage
	}

	// Check OpenAI format
	if response.Usage.PromptTokens > 0 || response.Usage.CompletionTokens > 0 {
		usage.InputTokens = response.Usage.PromptTokens
		usage.OutputTokens = response.Usage.CompletionTokens
		usage.Estimate = "exact"
		return usage
	}

	return usage
}

// classifyHTTPError classifies HTTP errors by status code.
func classifyHTTPError(statusCode int) string {
	switch {
	case statusCode >= 500:
		return "5xx"
	case statusCode >= 400:
		return "4xx"
	default:
		return fmt.Sprintf("http_%d", statusCode)
	}
}

// shouldRetryError determines if an error should trigger a retry.
func shouldRetryError(errorClass string) bool {
	switch errorClass {
	case "5xx", "timeout", "network":
		return true
	case "4xx":
		return false
	default:
		return false
	}
}
