package pricing

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOpenRouterAdapterFetch(t *testing.T) {
	// Create a mock server
	mockResponse := `{
		"data": [
			{
				"id": "openai/gpt-4",
				"name": "GPT-4",
				"context_length": 8192,
				"pricing": {
					"prompt": "0.00003",
					"completion": "0.00006",
					"request": "0.01",
					"input_cache_read": "0.000015",
					"input_cache_write": "0.0000375"
				}
			},
			{
				"id": "anthropic/claude-3-5-sonnet",
				"name": "Claude 3.5 Sonnet",
				"context_length": 200000,
				"pricing": {
					"prompt": "0.000003",
					"completion": "0.000015"
				}
			}
		]
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(mockResponse)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	// Create adapter with mock URL
	adapter := NewOpenRouterAdapter()
	adapter.apiURL = server.URL

	// Fetch pricing
	ctx := context.Background()
	schema, err := adapter.Fetch(ctx)
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	// Verify results
	if len(schema.Models) != 2 {
		t.Fatalf("Expected 2 models, got %d", len(schema.Models))
	}

	// Check GPT-4
	gpt4 := schema.Models[0]
	if gpt4.Provider != "openai" {
		t.Errorf("GPT-4 provider: got %s, want openai", gpt4.Provider)
	}

	if gpt4.ID != "openai/gpt-4" {
		t.Errorf("GPT-4 ID: got %s, want openai/gpt-4", gpt4.ID)
	}

	if gpt4.ContextTokens != 8192 {
		t.Errorf("GPT-4 context: got %d, want 8192", gpt4.ContextTokens)
	}

	if gpt4.Pricing.Token == nil {
		t.Fatal("GPT-4 token pricing is nil")
	}

	// Check prices are converted to per million
	// 0.00003 per token * 1,000,000 = 30
	if gpt4.Pricing.Token.Input != 30.0 {
		t.Errorf("GPT-4 input price: got %f, want 30.0", gpt4.Pricing.Token.Input)
	}

	if gpt4.Pricing.Token.Output != 60.0 {
		t.Errorf("GPT-4 output price: got %f, want 60.0", gpt4.Pricing.Token.Output)
	}

	// Check cache pricing
	if gpt4.Pricing.Token.CachedInputRead == nil {
		t.Fatal("GPT-4 cached input read is nil")
	}

	if *gpt4.Pricing.Token.CachedInputRead != 15.0 {
		t.Errorf("GPT-4 cached input read: got %f, want 15.0", *gpt4.Pricing.Token.CachedInputRead)
	}

	// Check per-request pricing
	if gpt4.Pricing.PerRequest == nil {
		t.Fatal("GPT-4 per-request pricing is nil")
	}

	if *gpt4.Pricing.PerRequest.Request != 0.01 {
		t.Errorf("GPT-4 per-request: got %f, want 0.01", *gpt4.Pricing.PerRequest.Request)
	}

	// Check Claude
	claude := schema.Models[1]
	if claude.Provider != "anthropic" {
		t.Errorf("Claude provider: got %s, want anthropic", claude.Provider)
	}

	if claude.Pricing.Token.Input != 3.0 {
		t.Errorf("Claude input price: got %f, want 3.0", claude.Pricing.Token.Input)
	}

	if claude.Pricing.Token.Output != 15.0 {
		t.Errorf("Claude output price: got %f, want 15.0", claude.Pricing.Token.Output)
	}

	// Claude shouldn't have per-request pricing in this mock
	if claude.Pricing.PerRequest != nil {
		t.Error("Claude shouldn't have per-request pricing")
	}

	// Check source metadata
	if gpt4.Source.Kind != "openrouter" {
		t.Errorf("GPT-4 source kind: got %s, want openrouter", gpt4.Source.Kind)
	}

	if gpt4.Source.URL != server.URL {
		t.Errorf("GPT-4 source URL: got %s, want %s", gpt4.Source.URL, server.URL)
	}
}

func TestOpenRouterAdapterFetchError(t *testing.T) {
	// Test server error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := w.Write([]byte("Internal Server Error")); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	adapter := NewOpenRouterAdapter()
	adapter.apiURL = server.URL

	ctx := context.Background()
	_, err := adapter.Fetch(ctx)
	if err == nil {
		t.Error("Expected error for 500 response, got nil")
	}
}

func TestOpenRouterAdapterFetchInvalidJSON(t *testing.T) {
	// Test invalid JSON response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte("invalid json")); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	adapter := NewOpenRouterAdapter()
	adapter.apiURL = server.URL

	ctx := context.Background()
	_, err := adapter.Fetch(ctx)
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestParsePrice(t *testing.T) {
	tests := []struct {
		name      string
		priceStr  string
		wantPrice float64
		wantErr   bool
	}{
		{
			name:      "valid price",
			priceStr:  "0.00003",
			wantPrice: 0.00003,
			wantErr:   false,
		},
		{
			name:      "zero price",
			priceStr:  "0",
			wantPrice: 0,
			wantErr:   false,
		},
		{
			name:      "large price",
			priceStr:  "123.456",
			wantPrice: 123.456,
			wantErr:   false,
		},
		{
			name:     "empty string",
			priceStr: "",
			wantErr:  true,
		},
		{
			name:     "invalid string",
			priceStr: "abc",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			price, err := parsePrice(tt.priceStr)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if price != tt.wantPrice {
				t.Errorf("Price: got %f, want %f", price, tt.wantPrice)
			}
		})
	}
}

func TestConvertModelPartial(t *testing.T) {
	adapter := NewOpenRouterAdapter()

	// Model with missing completion price (should be marked as partial)
	model := OpenRouterModel{
		ID:            "test/model",
		Name:          "Test Model",
		ContextLength: 4096,
		Pricing: OpenRouterPricing{
			Prompt: "0.00001",
			// Missing completion price
		},
	}

	pricing, partial := adapter.convertModel(model)

	if !partial {
		t.Error("Expected partial=true when completion price is missing")
	}

	if pricing.Source.Partial != partial {
		t.Error("Source.Partial should match partial return value")
	}
}
