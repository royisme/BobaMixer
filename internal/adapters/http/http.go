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
)

type Client struct {
	name       string
	provider   string // anthropic, openai, openrouter, etc.
	endpoint   string
	headers    map[string]string
	httpClient *http.Client
}

// UsageResponse represents common usage response structure
type UsageResponse struct {
	Usage struct {
		InputTokens       int `json:"input_tokens"`
		OutputTokens      int `json:"output_tokens"`
		PromptTokens      int `json:"prompt_tokens"`       // OpenAI format
		CompletionTokens  int `json:"completion_tokens"`   // OpenAI format
		TotalTokens       int `json:"total_tokens"`
	} `json:"usage"`
}

func New(name, endpoint string, headers map[string]string) *Client {
	return &Client{
		name:       name,
		endpoint:   endpoint,
		headers:    headers,
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

func NewWithProvider(name, provider, endpoint string, headers map[string]string) *Client {
	return &Client{
		name:       name,
		provider:   strings.ToLower(provider),
		endpoint:   endpoint,
		headers:    headers,
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

func (c *Client) Name() string { return c.name }

func (c *Client) Execute(ctx context.Context, req adapters.Request) (adapters.Result, error) {
	if c.endpoint == "" {
		return adapters.Result{}, errors.New("endpoint not configured")
	}

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
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

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
