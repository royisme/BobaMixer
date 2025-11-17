// Package httpadapter provides HTTP-based adapter implementation for API providers.
package httpadapter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/royisme/bobamixer/internal/adapters"
	"github.com/royisme/bobamixer/internal/logging"
)

// Client is an HTTP-based adapter for communicating with AI provider APIs.
type Client struct {
	headers    map[string]string
	httpClient *http.Client
	name       string
	provider   string // anthropic, openai, openrouter, etc.
	endpoint   string
}

// UsageResponse represents common usage response structure
type UsageResponse struct {
	Usage struct {
		InputTokens      int `json:"input_tokens"`
		OutputTokens     int `json:"output_tokens"`
		PromptTokens     int `json:"prompt_tokens"`     // OpenAI format
		CompletionTokens int `json:"completion_tokens"` // OpenAI format
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// New creates a new HTTP adapter client with the given name, endpoint, and headers.
func New(name, endpoint string, headers map[string]string) *Client {
	return &Client{
		name:       name,
		endpoint:   endpoint,
		headers:    headers,
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

// NewWithProvider creates a new HTTP adapter client with provider-specific configuration.
func NewWithProvider(name, provider, endpoint string, headers map[string]string) *Client {
	return &Client{
		name:       name,
		provider:   strings.ToLower(provider),
		endpoint:   endpoint,
		headers:    headers,
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

// Name returns the name of this adapter client.
func (c *Client) Name() string { return c.name }

// Execute sends an HTTP request to the configured endpoint and returns the result.
func (c *Client) Execute(ctx context.Context, req adapters.Request) (adapters.Result, error) {
	if c.endpoint == "" {
		return adapters.Result{}, errors.New("endpoint not configured")
	}

	// Execute with retry logic (max 2 retries = 3 total attempts)
	return c.executeWithRetry(ctx, req, 2)
}

func (c *Client) executeWithRetry(ctx context.Context, req adapters.Request, maxRetries int) (adapters.Result, error) {
	var lastResult adapters.Result
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Wait before retry (exponential backoff: 1s, 2s)
		if attempt > 0 {
			backoffDuration := time.Duration(attempt) * time.Second
			logging.Info("HTTP retry backoff",
				logging.String("client", c.name),
				logging.Int("attempt", attempt),
				logging.String("backoff", backoffDuration.String()))
			fmt.Printf("[Retry %d after %s]\n", attempt, backoffDuration)
			select {
			case <-time.After(backoffDuration):
			case <-ctx.Done():
				logging.Warn("HTTP request canceled during retry", logging.Err(ctx.Err()))
				return adapters.Result{}, ctx.Err()
			}
		}

		// Log attempt
		if attempt == 0 {
			logging.Info("HTTP request attempt", logging.String("client", c.name))
			fmt.Println("[Attempt 1]")
		}

		// Execute the request
		result, err := c.executeOnce(ctx, req)
		lastResult = result
		lastErr = err

		// If request creation or network error, retry
		if err != nil {
			logging.Error("HTTP request failed",
				logging.String("client", c.name),
				logging.Int("attempt", attempt),
				logging.Err(err))
			if attempt < maxRetries {
				fmt.Printf("Network error: %v\n", err)
				continue
			}
			return result, err
		}

		// Check if we should retry based on status code
		if result.Success {
			logging.Info("HTTP request succeeded",
				logging.String("client", c.name),
				logging.Int64("latency_ms", result.Usage.LatencyMS),
				logging.Int("input_tokens", result.Usage.InputTokens),
				logging.Int("output_tokens", result.Usage.OutputTokens))
			return result, nil
		}

		// Parse status code from error message
		shouldRetry := c.shouldRetry(result.Error)
		if !shouldRetry || attempt >= maxRetries {
			logging.Warn("HTTP request failed, not retrying",
				logging.String("client", c.name),
				logging.String("error", result.Error),
				logging.Bool("should_retry", shouldRetry),
				logging.Int("attempt", attempt))
			return result, nil
		}

		logging.Warn("HTTP request failed, will retry",
			logging.String("client", c.name),
			logging.String("error", result.Error))
		fmt.Printf("%s\n", result.Error)
	}

	return lastResult, lastErr
}

func (c *Client) executeOnce(ctx context.Context, req adapters.Request) (adapters.Result, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, strings.NewReader(string(req.Payload)))
	if err != nil {
		return adapters.Result{}, err
	}

	for k, v := range c.headers {
		httpReq.Header.Set(k, v)
	}

	start := time.Now()
	resp, err := c.httpClient.Do(httpReq)
	latencyMS := time.Since(start).Milliseconds()

	if err != nil {
		return adapters.Result{
			Success: false,
			Error:   err.Error(),
			Usage:   adapters.Usage{Estimate: adapters.EstimateHeuristic, LatencyMS: latencyMS},
		}, nil
	}
	defer func() {
		//nolint:errcheck,gosec // Best effort cleanup, error irrelevant in defer
		resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return adapters.Result{
			Success: false,
			Error:   fmt.Sprintf("read response: %v", err),
			Usage:   adapters.Usage{Estimate: adapters.EstimateHeuristic, LatencyMS: latencyMS},
		}, nil
	}

	// Try to parse usage from response
	usage := c.parseUsage(body)
	usage.LatencyMS = latencyMS

	return adapters.Result{
		Success: resp.StatusCode >= 200 && resp.StatusCode < 300,
		Output:  body,
		Error:   errorText(resp.StatusCode, body),
		Usage:   usage,
	}, nil
}

// shouldRetry determines if a request should be retried based on the error
func (c *Client) shouldRetry(errorMsg string) bool {
	if errorMsg == "" {
		return false
	}

	// Check for 5xx server errors (should retry)
	if strings.Contains(errorMsg, "status 5") {
		return true
	}

	// Check for timeout errors (should retry)
	if strings.Contains(errorMsg, "timeout") ||
		strings.Contains(errorMsg, "Timeout") ||
		strings.Contains(errorMsg, "deadline exceeded") {
		return true
	}

	// Check for 4xx client errors (should NOT retry)
	if strings.Contains(errorMsg, "status 4") {
		return false
	}

	// For other errors, don't retry
	return false
}

// parseUsage attempts to extract usage information from API response
func (c *Client) parseUsage(body []byte) adapters.Usage {
	usage := adapters.Usage{
		InputTokens:  0,
		OutputTokens: 0,
		Estimate:     adapters.EstimateHeuristic,
	}

	if len(body) == 0 {
		return usage
	}

	// Try to parse as JSON
	var usageResp UsageResponse
	if err := json.Unmarshal(body, &usageResp); err != nil {
		// Not JSON or parsing failed, will need estimation
		return usage
	}

	// Check for usage information
	if usageResp.Usage.InputTokens > 0 || usageResp.Usage.OutputTokens > 0 {
		// Anthropic format or similar
		usage.InputTokens = usageResp.Usage.InputTokens
		usage.OutputTokens = usageResp.Usage.OutputTokens
		usage.Estimate = adapters.EstimateExact
		return usage
	}

	if usageResp.Usage.PromptTokens > 0 || usageResp.Usage.CompletionTokens > 0 {
		// OpenAI/OpenRouter format
		usage.InputTokens = usageResp.Usage.PromptTokens
		usage.OutputTokens = usageResp.Usage.CompletionTokens
		usage.Estimate = adapters.EstimateExact
		return usage
	}

	// No usage found, will need estimation
	return usage
}

// SetTimeout sets the HTTP client timeout
func (c *Client) SetTimeout(timeout time.Duration) {
	c.httpClient.Timeout = timeout
}

// EstimateUsageFromPayload estimates token usage from request/response when exact usage unavailable
func (c *Client) EstimateUsageFromPayload(requestBody, responseBody []byte) adapters.Usage {
	// This would be called by higher-level code that has a tokenizer
	return adapters.Usage{
		InputTokens:  0,
		OutputTokens: 0,
		Estimate:     adapters.EstimateHeuristic,
	}
}

func errorText(status int, body []byte) string {
	if status >= 200 && status < 300 {
		return ""
	}

	// Try to extract error message from JSON
	var errorResp struct {
		Error struct {
			Message string `json:"message"`
			Type    string `json:"type"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &errorResp); err == nil && errorResp.Error.Message != "" {
		return fmt.Sprintf("status %d: %s", status, errorResp.Error.Message)
	}

	// Fallback to raw body (truncated)
	bodyStr := string(body)
	if len(bodyStr) > 200 {
		bodyStr = bodyStr[:200] + "..."
	}
	return fmt.Sprintf("status %d: %s", status, bodyStr)
}
