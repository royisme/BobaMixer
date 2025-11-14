package httpx_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/royisme/bobamixer/internal/httpx"
)

//nolint:gocyclo // Test function with multiple subtests is acceptable
func TestExecute(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		// Given: server that returns 200
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprintln(w, `{"usage":{"input_tokens":100,"output_tokens":200}}`) //nolint:errcheck // test handler
		}))
		defer server.Close()

		ctx := context.Background()
		req := httpx.HTTPRequest{
			SessionID: "test-session",
			Endpoint:  server.URL,
			Headers:   map[string]string{"Content-Type": "application/json"},
			Payload:   []byte(`{"test":"data"}`),
			Timeout:   5 * time.Second,
			Retries:   2,
		}

		// When: Execute is called
		result, err := httpx.Execute(ctx, req)

		// Then: request succeeds
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
		if !result.Success {
			t.Error("expected success=true")
		}
		if result.StatusCode != 200 {
			t.Errorf("status code = %d, want 200", result.StatusCode)
		}
		if result.ErrorClass != "" {
			t.Errorf("error class = %s, want empty", result.ErrorClass)
		}
	})

	t.Run("5xx retry succeeds on second attempt", func(t *testing.T) {
		// Given: server that fails first time, succeeds second time
		var attemptCount int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			count := atomic.AddInt32(&attemptCount, 1)
			if count == 1 {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = fmt.Fprintln(w, "Server error") //nolint:errcheck // test handler
			} else {
				w.WriteHeader(http.StatusOK)
				_, _ = fmt.Fprintln(w, `{"usage":{"input_tokens":50,"output_tokens":100}}`) //nolint:errcheck // test handler
			}
		}))
		defer server.Close()

		ctx := context.Background()
		req := httpx.HTTPRequest{
			SessionID: "test-session",
			Endpoint:  server.URL,
			Retries:   2,
		}

		// When: Execute is called
		result, err := httpx.Execute(ctx, req)

		// Then: succeeds on retry
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
		if !result.Success {
			t.Error("expected success=true after retry")
		}
		if attemptCount != 2 {
			t.Errorf("attempt count = %d, want 2", attemptCount)
		}
	})

	t.Run("401 does not retry", func(t *testing.T) {
		// Given: server that returns 401
		var attemptCount int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&attemptCount, 1)
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = fmt.Fprintln(w, `{"error":"unauthorized"}`) //nolint:errcheck // test handler
		}))
		defer server.Close()

		ctx := context.Background()
		req := httpx.HTTPRequest{
			SessionID: "test-session",
			Endpoint:  server.URL,
			Retries:   2,
		}

		// When: Execute is called
		result, err := httpx.Execute(ctx, req)

		// Then: fails without retry
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
		if result.Success {
			t.Error("expected success=false for 401")
		}
		if result.StatusCode != 401 {
			t.Errorf("status code = %d, want 401", result.StatusCode)
		}
		if result.ErrorClass != "4xx" {
			t.Errorf("error class = %s, want 4xx", result.ErrorClass)
		}
		if attemptCount != 1 {
			t.Errorf("attempt count = %d, want 1 (no retry)", attemptCount)
		}
	})

	t.Run("403 does not retry", func(t *testing.T) {
		// Given: server that returns 403
		var attemptCount int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&attemptCount, 1)
			w.WriteHeader(http.StatusForbidden)
			_, _ = fmt.Fprintln(w, `{"error":"forbidden"}`) //nolint:errcheck // test handler
		}))
		defer server.Close()

		ctx := context.Background()
		req := httpx.HTTPRequest{
			SessionID: "test-session",
			Endpoint:  server.URL,
			Retries:   2,
		}

		// When: Execute is called
		result, err := httpx.Execute(ctx, req)

		// Then: fails without retry
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
		if result.Success {
			t.Error("expected success=false for 403")
		}
		if attemptCount != 1 {
			t.Errorf("attempt count = %d, want 1 (no retry)", attemptCount)
		}
	})

	t.Run("timeout retries", func(t *testing.T) {
		// Given: server that delays response
		var attemptCount int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			count := atomic.AddInt32(&attemptCount, 1)
			if count == 1 {
				time.Sleep(200 * time.Millisecond)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprintln(w, `{"test":"ok"}`) //nolint:errcheck // test handler
		}))
		defer server.Close()

		ctx := context.Background()
		req := httpx.HTTPRequest{
			SessionID: "test-session",
			Endpoint:  server.URL,
			Timeout:   50 * time.Millisecond,
			Retries:   2,
		}

		// When: Execute is called
		_, err := httpx.Execute(ctx, req)

		// Then: retries on timeout
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
		// Should have attempted at least twice (first timeout, second success)
		if attemptCount < 2 {
			t.Errorf("attempt count = %d, want >= 2", attemptCount)
		}
	})

	t.Run("network error retries", func(t *testing.T) {
		// Given: invalid endpoint
		ctx := context.Background()
		req := httpx.HTTPRequest{
			SessionID: "test-session",
			Endpoint:  "http://invalid-host-that-does-not-exist.local:9999",
			Timeout:   1 * time.Second,
			Retries:   2,
		}

		// When: Execute is called
		result, err := httpx.Execute(ctx, req)

		// Then: retries on network error and eventually fails
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
		if result.Success {
			t.Error("expected success=false for network error")
		}
		if result.ErrorClass != "network" {
			t.Errorf("error class = %s, want network", result.ErrorClass)
		}
	})

	t.Run("parses usage from response", func(t *testing.T) {
		// Given: server that returns usage in Anthropic format
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprintln(w, `{"usage":{"input_tokens":150,"output_tokens":250}}`) //nolint:errcheck // test handler
		}))
		defer server.Close()

		ctx := context.Background()
		req := httpx.HTTPRequest{
			SessionID: "test-session",
			Endpoint:  server.URL,
		}

		// When: Execute is called
		result, err := httpx.Execute(ctx, req)

		// Then: usage is parsed as exact
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
		if result.Usage.Estimate != "exact" {
			t.Errorf("usage estimate = %s, want exact", result.Usage.Estimate)
		}
		if result.Usage.InputTokens != 150 {
			t.Errorf("input tokens = %d, want 150", result.Usage.InputTokens)
		}
		if result.Usage.OutputTokens != 250 {
			t.Errorf("output tokens = %d, want 250", result.Usage.OutputTokens)
		}
	})

	t.Run("parses usage from OpenAI format", func(t *testing.T) {
		// Given: server that returns usage in OpenAI format
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprintln(w, `{"usage":{"prompt_tokens":200,"completion_tokens":300}}`) //nolint:errcheck // test handler
		}))
		defer server.Close()

		ctx := context.Background()
		req := httpx.HTTPRequest{
			SessionID: "test-session",
			Endpoint:  server.URL,
		}

		// When: Execute is called
		result, err := httpx.Execute(ctx, req)

		// Then: usage is parsed correctly
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
		if result.Usage.Estimate != "exact" {
			t.Errorf("usage estimate = %s, want exact", result.Usage.Estimate)
		}
		if result.Usage.InputTokens != 200 {
			t.Errorf("input tokens = %d, want 200", result.Usage.InputTokens)
		}
		if result.Usage.OutputTokens != 300 {
			t.Errorf("output tokens = %d, want 300", result.Usage.OutputTokens)
		}
	})

	t.Run("respects custom headers", func(t *testing.T) {
		// Given: server that checks headers
		var receivedHeaders http.Header
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedHeaders = r.Header
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprintln(w, `{"status":"ok"}`) //nolint:errcheck // test handler
		}))
		defer server.Close()

		ctx := context.Background()
		req := httpx.HTTPRequest{
			SessionID: "test-session",
			Endpoint:  server.URL,
			Headers: map[string]string{
				"X-Custom-Header": "test-value",
				"Authorization":   "Bearer test-token",
			},
		}

		// When: Execute is called
		result, err := httpx.Execute(ctx, req)

		// Then: headers are sent
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
		if !result.Success {
			t.Error("expected success=true")
		}
		if receivedHeaders.Get("X-Custom-Header") != "test-value" {
			t.Error("custom header not received")
		}
		if receivedHeaders.Get("Authorization") != "Bearer test-token" {
			t.Error("authorization header not received")
		}
	})

	t.Run("includes latency measurement", func(t *testing.T) {
		// Given: server with slight delay
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(10 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprintln(w, `{"status":"ok"}`) //nolint:errcheck // test handler
		}))
		defer server.Close()

		ctx := context.Background()
		req := httpx.HTTPRequest{
			SessionID: "test-session",
			Endpoint:  server.URL,
		}

		// When: Execute is called
		result, err := httpx.Execute(ctx, req)

		// Then: latency is measured
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
		if result.Usage.LatencyMS <= 0 {
			t.Errorf("latency = %d, want > 0", result.Usage.LatencyMS)
		}
	})
}

func TestErrorClassification(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		wantSuccess    bool
		wantErrorClass string
		wantRetry      bool
	}{
		{name: "200 OK", statusCode: 200, wantSuccess: true, wantErrorClass: "", wantRetry: false},
		{name: "400 Bad Request", statusCode: 400, wantSuccess: false, wantErrorClass: "4xx", wantRetry: false},
		{name: "401 Unauthorized", statusCode: 401, wantSuccess: false, wantErrorClass: "4xx", wantRetry: false},
		{name: "403 Forbidden", statusCode: 403, wantSuccess: false, wantErrorClass: "4xx", wantRetry: false},
		{name: "404 Not Found", statusCode: 404, wantSuccess: false, wantErrorClass: "4xx", wantRetry: false},
		{name: "500 Internal Server Error", statusCode: 500, wantSuccess: false, wantErrorClass: "5xx", wantRetry: true},
		{name: "502 Bad Gateway", statusCode: 502, wantSuccess: false, wantErrorClass: "5xx", wantRetry: true},
		{name: "503 Service Unavailable", statusCode: 503, wantSuccess: false, wantErrorClass: "5xx", wantRetry: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var attemptCount int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				atomic.AddInt32(&attemptCount, 1)
				w.WriteHeader(tt.statusCode)
				_, _ = fmt.Fprintf(w, `{"error":"test error"}`) //nolint:errcheck // test handler
			}))
			defer server.Close()

			ctx := context.Background()
			req := httpx.HTTPRequest{
				SessionID: "test-session",
				Endpoint:  server.URL,
				Retries:   2,
			}

			result, err := httpx.Execute(ctx, req)
			if err != nil {
				t.Fatalf("Execute failed: %v", err)
			}

			if result.Success != tt.wantSuccess {
				t.Errorf("success = %v, want %v", result.Success, tt.wantSuccess)
			}
			if !strings.Contains(result.ErrorClass, tt.wantErrorClass) {
				t.Errorf("error class = %s, want to contain %s", result.ErrorClass, tt.wantErrorClass)
			}

			// Check retry behavior
			expectedAttempts := int32(1)
			if tt.wantRetry {
				expectedAttempts = 3 // 1 initial + 2 retries
			}
			if attemptCount != expectedAttempts {
				t.Errorf("attempt count = %d, want %d", attemptCount, expectedAttempts)
			}
		})
	}
}
