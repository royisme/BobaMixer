package httpadapter

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/vantagecraft-dev/bobamixer/internal/adapters"
)

type Client struct {
	name       string
	endpoint   string
	headers    map[string]string
	httpClient *http.Client
}

func New(name, endpoint string, headers map[string]string) *Client {
	return &Client{
		name:       name,
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
	if err != nil {
		return adapters.Result{Success: false, Error: err.Error(), Usage: adapters.Usage{Estimate: adapters.EstimateHeuristic}}, nil
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	usage := adapters.Usage{Estimate: adapters.EstimateHeuristic, LatencyMS: time.Since(start).Milliseconds()}
	return adapters.Result{
		Success: resp.StatusCode >= 200 && resp.StatusCode < 300,
		Output:  body,
		Error:   errorText(resp.StatusCode, body),
		Usage:   usage,
	}, nil
}

func errorText(status int, body []byte) string {
	if status >= 200 && status < 300 {
		return ""
	}
	return fmt.Sprintf("status %d: %s", status, string(body))
}
