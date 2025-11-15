package pricing

import (
	"testing"
	"time"
)

func TestNewPricingSchema(t *testing.T) {
	schema := NewPricingSchema()

	if schema.Version != SchemaVersion {
		t.Errorf("Version: got %d, want %d", schema.Version, SchemaVersion)
	}

	if schema.Currency != "USD" {
		t.Errorf("Currency: got %s, want USD", schema.Currency)
	}

	if schema.Models == nil {
		t.Error("Models should be initialized")
	}

	if schema.FetchedAt.IsZero() {
		t.Error("FetchedAt should be set")
	}
}

func TestToLegacyTable(t *testing.T) {
	schema := &PricingSchema{
		Version:  1,
		Currency: "USD",
		Models: []ModelPricing{
			{
				Provider:    "openai",
				ID:          "gpt-4",
				DisplayName: "GPT-4",
				Pricing: PricingTiers{
					Token: &TokenPricing{
						Input:  15.0,  // $15 per 1M tokens
						Output: 30.0,  // $30 per 1M tokens
					},
				},
			},
			{
				Provider:    "anthropic",
				ID:          "claude-3-5-sonnet",
				DisplayName: "Claude 3.5 Sonnet",
				Pricing: PricingTiers{
					Token: &TokenPricing{
						Input:  3.0,  // $3 per 1M tokens
						Output: 15.0, // $15 per 1M tokens
					},
				},
			},
		},
	}

	table := schema.ToLegacyTable()

	if len(table.Models) != 2 {
		t.Fatalf("Expected 2 models, got %d", len(table.Models))
	}

	// Check GPT-4 pricing (should be divided by 1000)
	gpt4, ok := table.Models["gpt-4"]
	if !ok {
		t.Fatal("gpt-4 not found in table")
	}

	if gpt4.InputPer1K != 0.015 {
		t.Errorf("GPT-4 InputPer1K: got %f, want 0.015", gpt4.InputPer1K)
	}

	if gpt4.OutputPer1K != 0.030 {
		t.Errorf("GPT-4 OutputPer1K: got %f, want 0.030", gpt4.OutputPer1K)
	}

	// Check Claude pricing
	claude, ok := table.Models["claude-3-5-sonnet"]
	if !ok {
		t.Fatal("claude-3-5-sonnet not found in table")
	}

	if claude.InputPer1K != 0.003 {
		t.Errorf("Claude InputPer1K: got %f, want 0.003", claude.InputPer1K)
	}

	if claude.OutputPer1K != 0.015 {
		t.Errorf("Claude OutputPer1K: got %f, want 0.015", claude.OutputPer1K)
	}
}

func TestToLegacyTableWithMissingTokenPricing(t *testing.T) {
	schema := &PricingSchema{
		Version:  1,
		Currency: "USD",
		Models: []ModelPricing{
			{
				Provider:    "test",
				ID:          "test-model",
				DisplayName: "Test Model",
				Pricing: PricingTiers{
					// No token pricing
					Image: &ImagePricing{
						InputPerImage: floatPtr(0.5),
					},
				},
			},
		},
	}

	table := schema.ToLegacyTable()

	// Should not include models without token pricing
	if len(table.Models) != 0 {
		t.Errorf("Expected 0 models in legacy table, got %d", len(table.Models))
	}
}

func TestCachedPricingExpiry(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		expiresAt time.Time
		wantFresh bool
	}{
		{
			name:      "fresh cache",
			expiresAt: now.Add(1 * time.Hour),
			wantFresh: true,
		},
		{
			name:      "expired cache",
			expiresAt: now.Add(-1 * time.Hour),
			wantFresh: false,
		},
		{
			name:      "just expired",
			expiresAt: now.Add(-1 * time.Second),
			wantFresh: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cached := &CachedPricing{
				Metadata: CacheMetadata{
					ExpiresAt: tt.expiresAt,
				},
			}

			if cached.IsFresh() != tt.wantFresh {
				t.Errorf("IsFresh(): got %v, want %v", cached.IsFresh(), tt.wantFresh)
			}

			if cached.IsExpired() == tt.wantFresh {
				t.Errorf("IsExpired(): got %v, want %v", cached.IsExpired(), !tt.wantFresh)
			}
		})
	}
}

func floatPtr(f float64) *float64 {
	return &f
}
