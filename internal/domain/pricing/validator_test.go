package pricing

import (
	"strings"
	"testing"
)

func TestValidatorEmptySchema(t *testing.T) {
	validator := NewPricingValidator()

	schema := &PricingSchema{
		Version:  1,
		Currency: "USD",
		Models:   []ModelPricing{},
	}

	warnings := validator.ValidateAgainstRefs(schema)

	if len(warnings) == 0 {
		t.Error("Expected warnings for empty schema")
	}

	if !strings.Contains(warnings[0].Message, "empty") {
		t.Errorf("Expected 'empty' in warning message, got: %s", warnings[0].Message)
	}
}

func TestValidatorValidModel(t *testing.T) {
	validator := NewPricingValidator()

	schema := &PricingSchema{
		Version:  1,
		Currency: "USD",
		Models: []ModelPricing{
			{
				Provider:      "openai",
				ID:            "gpt-4",
				DisplayName:   "GPT-4",
				ContextTokens: 8192,
				Pricing: PricingTiers{
					Token: &TokenPricing{
						Input:  30.0,
						Output: 60.0,
					},
				},
				Source: SourceMeta{
					Kind: "openrouter",
					URL:  "https://openrouter.ai/api/v1/models",
				},
			},
		},
	}

	warnings := validator.ValidateAgainstRefs(schema)

	// Should have no warnings for valid model
	if len(warnings) > 0 {
		t.Errorf("Expected no warnings for valid model, got %d warnings", len(warnings))
		for i, w := range warnings {
			t.Logf("Warning %d: %s", i+1, w.Message)
		}
	}
}

func TestValidatorMissingTokenPricing(t *testing.T) {
	validator := NewPricingValidator()

	schema := &PricingSchema{
		Version:  1,
		Currency: "USD",
		Models: []ModelPricing{
			{
				Provider:    "openai",
				ID:          "gpt-4",
				DisplayName: "GPT-4",
				Pricing:     PricingTiers{
					// Missing Token pricing
				},
			},
		},
	}

	warnings := validator.ValidateAgainstRefs(schema)

	if len(warnings) == 0 {
		t.Fatal("Expected warnings for missing token pricing")
	}

	foundTokenWarning := false
	for _, w := range warnings {
		if strings.Contains(w.Message, "token pricing") {
			foundTokenWarning = true
			break
		}
	}

	if !foundTokenWarning {
		t.Error("Expected warning about missing token pricing")
	}
}

func TestValidatorZeroOrNegativePrices(t *testing.T) {
	validator := NewPricingValidator()

	tests := []struct {
		name   string
		input  float64
		output float64
		want   string // Expected warning substring
	}{
		{
			name:   "zero input price",
			input:  0,
			output: 10.0,
			want:   "zero or negative",
		},
		{
			name:   "negative input price",
			input:  -1.0,
			output: 10.0,
			want:   "zero or negative",
		},
		{
			name:   "zero output price",
			input:  10.0,
			output: 0,
			want:   "zero or negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := &PricingSchema{
				Version:  1,
				Currency: "USD",
				Models: []ModelPricing{
					{
						Provider: "openai",
						ID:       "test-model",
						Pricing: PricingTiers{
							Token: &TokenPricing{
								Input:  tt.input,
								Output: tt.output,
							},
						},
						Source: SourceMeta{Kind: "test"},
					},
				},
			}

			warnings := validator.ValidateAgainstRefs(schema)

			if len(warnings) == 0 {
				t.Fatal("Expected warnings for zero/negative prices")
			}

			foundWarning := false
			for _, w := range warnings {
				if strings.Contains(strings.ToLower(w.Message), strings.ToLower(tt.want)) {
					foundWarning = true
					break
				}
			}

			if !foundWarning {
				t.Errorf("Expected warning containing '%s'", tt.want)
			}
		})
	}
}

func TestValidatorUnrealisticPrices(t *testing.T) {
	validator := NewPricingValidator()

	tests := []struct {
		name   string
		input  float64
		output float64
		want   string
	}{
		{
			name:   "unusually high input",
			input:  150.0,
			output: 200.0,
			want:   "unusually high",
		},
		{
			name:   "unusually low input",
			input:  0.0001,
			output: 0.0002,
			want:   "unusually low",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := &PricingSchema{
				Version:  1,
				Currency: "USD",
				Models: []ModelPricing{
					{
						Provider: "openai",
						ID:       "test-model",
						Pricing: PricingTiers{
							Token: &TokenPricing{
								Input:  tt.input,
								Output: tt.output,
							},
						},
						Source: SourceMeta{Kind: "test"},
					},
				},
			}

			warnings := validator.ValidateAgainstRefs(schema)

			foundWarning := false
			for _, w := range warnings {
				if strings.Contains(strings.ToLower(w.Message), strings.ToLower(tt.want)) {
					foundWarning = true
					break
				}
			}

			if !foundWarning {
				t.Logf("Warnings: %v", warnings)
				t.Errorf("Expected warning containing '%s'", tt.want)
			}
		})
	}
}

func TestValidatorOutputLowerThanInput(t *testing.T) {
	validator := NewPricingValidator()

	schema := &PricingSchema{
		Version:  1,
		Currency: "USD",
		Models: []ModelPricing{
			{
				Provider: "openai", // Use a known provider to avoid "unknown provider" warning
				ID:       "test-model",
				Pricing: PricingTiers{
					Token: &TokenPricing{
						Input:  10.0,
						Output: 5.0, // Lower than input (unusual)
					},
				},
				Source: SourceMeta{Kind: "test"},
			},
		},
	}

	warnings := validator.ValidateAgainstRefs(schema)

	foundWarning := false
	for _, w := range warnings {
		t.Logf("Warning: %s - %s", w.Field, w.Message)
		if strings.Contains(w.Message, "lower than input") {
			foundWarning = true
			break
		}
	}

	if !foundWarning {
		t.Errorf("Expected warning about output price lower than input, got %d warnings", len(warnings))
	}
}

func TestValidatorPartialSource(t *testing.T) {
	validator := NewPricingValidator()

	schema := &PricingSchema{
		Version:  1,
		Currency: "USD",
		Models: []ModelPricing{
			{
				Provider: "openai",
				ID:       "test-model",
				Pricing: PricingTiers{
					Token: &TokenPricing{
						Input:  10.0,
						Output: 20.0,
					},
				},
				Source: SourceMeta{
					Kind:    "openrouter",
					Partial: true, // Marked as partial
				},
			},
		},
	}

	warnings := validator.ValidateAgainstRefs(schema)

	foundPartialWarning := false
	for _, w := range warnings {
		if strings.Contains(w.Message, "partial") {
			foundPartialWarning = true
			break
		}
	}

	if !foundPartialWarning {
		t.Error("Expected warning about partial source")
	}
}

func TestValidatorUnknownProvider(t *testing.T) {
	validator := NewPricingValidator()

	schema := &PricingSchema{
		Version:  1,
		Currency: "USD",
		Models: []ModelPricing{
			{
				Provider: "unknown-provider",
				ID:       "test-model",
				Pricing: PricingTiers{
					Token: &TokenPricing{
						Input:  10.0,
						Output: 20.0,
					},
				},
				Source: SourceMeta{Kind: "test"},
			},
		},
	}

	warnings := validator.ValidateAgainstRefs(schema)

	foundProviderWarning := false
	for _, w := range warnings {
		if strings.Contains(w.Message, "not in official reference") {
			foundProviderWarning = true
			break
		}
	}

	if !foundProviderWarning {
		t.Error("Expected warning about unknown provider")
	}
}

func TestGetReferenceList(t *testing.T) {
	validator := NewPricingValidator()

	refList := validator.GetReferenceList()

	// Check that all expected providers are in the list
	expectedProviders := []string{"openai", "anthropic", "deepseek", "google", "azure", "mistral", "cohere"}
	for _, provider := range expectedProviders {
		if !strings.Contains(refList, provider) {
			t.Errorf("Reference list missing provider: %s", provider)
		}
	}

	// Check that URLs are included
	if !strings.Contains(refList, "https://") {
		t.Error("Reference list should contain URLs")
	}

	// Check formatting
	if !strings.Contains(refList, "Official Pricing References") {
		t.Error("Reference list should have proper header")
	}
}

func TestFormatWarnings(t *testing.T) {
	warnings := []ValidationWarning{
		{
			Provider: "openai",
			ModelID:  "gpt-4",
			Field:    "token.input",
			Message:  "Price seems high",
			RefURL:   "https://openai.com/api/pricing/",
		},
		{
			Provider: "anthropic",
			ModelID:  "claude-3",
			Message:  "Missing cache pricing",
		},
	}

	formatted := FormatWarnings(warnings)

	// Check that all warnings are included
	if !strings.Contains(formatted, "openai") {
		t.Error("Formatted warnings should contain 'openai'")
	}

	if !strings.Contains(formatted, "gpt-4") {
		t.Error("Formatted warnings should contain 'gpt-4'")
	}

	if !strings.Contains(formatted, "Price seems high") {
		t.Error("Formatted warnings should contain warning message")
	}

	if !strings.Contains(formatted, "https://openai.com/api/pricing/") {
		t.Error("Formatted warnings should contain reference URL")
	}

	// Check count
	if !strings.Contains(formatted, "2 pricing validation warning") {
		t.Error("Formatted warnings should show correct count")
	}
}

func TestFormatWarningsEmpty(t *testing.T) {
	formatted := FormatWarnings([]ValidationWarning{})

	if !strings.Contains(formatted, "No pricing validation warnings") {
		t.Errorf("Expected 'No warnings' message, got: %s", formatted)
	}
}
