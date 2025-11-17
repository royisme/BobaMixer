package pricing

import (
	"fmt"
	"strings"
)

// ValidationWarning represents a pricing validation warning
type ValidationWarning struct {
	Provider string
	ModelID  string
	Field    string
	Message  string
	RefURL   string
}

// PricingValidator validates pricing data against known references.
//
//nolint:revive // PricingValidator is the established API name
type PricingValidator struct {
	references map[string]PricingReference
}

// PricingReference contains reference information for a provider.
//
//nolint:revive // PricingReference is the established API name
type PricingReference struct {
	Provider    string
	URL         string
	Unit        string // e.g., "per_1M_tokens"
	Description string
	LastChecked string // ISO date when this reference was last verified
}

// NewPricingValidator creates a new pricing validator
func NewPricingValidator() *PricingValidator {
	return &PricingValidator{
		references: getOfficialReferences(),
	}
}

// getOfficialReferences returns the list of official pricing reference URLs
func getOfficialReferences() map[string]PricingReference {
	return map[string]PricingReference{
		"openai": {
			Provider:    "openai",
			URL:         "https://openai.com/api/pricing/",
			Unit:        "per_1M_tokens",
			Description: "Official OpenAI pricing page with cached input pricing",
			LastChecked: "2025-11-14",
		},
		"anthropic": {
			Provider:    "anthropic",
			URL:         "https://www.anthropic.com/pricing",
			Unit:        "per_1M_tokens",
			Description: "Official Anthropic pricing for Claude models",
			LastChecked: "2025-11-14",
		},
		"deepseek": {
			Provider:    "deepseek",
			URL:         "https://platform.deepseek.com/api-docs/pricing/",
			Unit:        "per_1M_tokens",
			Description: "DeepSeek pricing with cache hit/miss details",
			LastChecked: "2025-11-14",
		},
		"google": {
			Provider:    "google",
			URL:         "https://ai.google.dev/pricing",
			Unit:        "per_1M_tokens",
			Description: "Google Gemini pricing with multimodal pricing (text/audio/image/video)",
			LastChecked: "2025-11-14",
		},
		"azure": {
			Provider:    "azure",
			URL:         "https://azure.microsoft.com/en-us/pricing/details/cognitive-services/openai-service/",
			Unit:        "varies",
			Description: "Azure OpenAI pricing with regional and tool-based pricing (File Search, Vector Store, etc.)",
			LastChecked: "2025-11-14",
		},
		"mistral": {
			Provider:    "mistral",
			URL:         "https://mistral.ai/technology/#pricing",
			Unit:        "per_1M_tokens",
			Description: "Mistral AI pricing page",
			LastChecked: "2025-11-14",
		},
		"cohere": {
			Provider:    "cohere",
			URL:         "https://cohere.com/pricing",
			Unit:        "per_1M_tokens",
			Description: "Cohere pricing page",
			LastChecked: "2025-11-14",
		},
	}
}

// ValidateAgainstRefs validates pricing schema and returns warnings
// This does NOT fetch/parse HTML - it only provides reference checks
func (v *PricingValidator) ValidateAgainstRefs(schema *PricingSchema) []ValidationWarning {
	var warnings []ValidationWarning

	if schema == nil || len(schema.Models) == 0 {
		warnings = append(warnings, ValidationWarning{
			Message: "Pricing schema is empty - no models found",
		})
		return warnings
	}

	// Group models by provider
	providerModels := make(map[string][]ModelPricing)
	for _, model := range schema.Models {
		providerModels[model.Provider] = append(providerModels[model.Provider], model)
	}

	// Check each provider
	for provider, models := range providerModels {
		ref, exists := v.references[provider]
		if !exists {
			// Provider not in our reference list
			warnings = append(warnings, ValidationWarning{
				Provider: provider,
				Message:  fmt.Sprintf("Provider '%s' not in official reference list (may need manual verification)", provider),
			})
			continue
		}

		// Validate models for this provider
		for _, model := range models {
			warnings = append(warnings, v.validateModel(model, ref)...)
		}
	}

	return warnings
}

// validateModel validates a single model against reference
func (v *PricingValidator) validateModel(model ModelPricing, ref PricingReference) []ValidationWarning {
	var warnings []ValidationWarning

	// Check if token pricing exists
	if model.Pricing.Token == nil {
		warnings = append(warnings, ValidationWarning{
			Provider: model.Provider,
			ModelID:  model.ID,
			Field:    "token",
			Message:  "Missing token pricing",
			RefURL:   ref.URL,
		})
		return warnings
	}

	// Check for zero or negative prices
	if model.Pricing.Token.Input <= 0 {
		warnings = append(warnings, ValidationWarning{
			Provider: model.Provider,
			ModelID:  model.ID,
			Field:    "token.input",
			Message:  fmt.Sprintf("Input price is zero or negative: %f", model.Pricing.Token.Input),
			RefURL:   ref.URL,
		})
	}

	if model.Pricing.Token.Output <= 0 {
		warnings = append(warnings, ValidationWarning{
			Provider: model.Provider,
			ModelID:  model.ID,
			Field:    "token.output",
			Message:  fmt.Sprintf("Output price is zero or negative: %f", model.Pricing.Token.Output),
			RefURL:   ref.URL,
		})
	}

	// Check for unrealistic prices (too high or too low)
	// Typical range: $0.01 to $100 per 1M tokens
	if model.Pricing.Token.Input > 100 {
		warnings = append(warnings, ValidationWarning{
			Provider: model.Provider,
			ModelID:  model.ID,
			Field:    "token.input",
			Message:  fmt.Sprintf("Input price seems unusually high: $%f per 1M tokens", model.Pricing.Token.Input),
			RefURL:   ref.URL,
		})
	}

	if model.Pricing.Token.Input < 0.001 {
		warnings = append(warnings, ValidationWarning{
			Provider: model.Provider,
			ModelID:  model.ID,
			Field:    "token.input",
			Message:  fmt.Sprintf("Input price seems unusually low: $%f per 1M tokens (may be correct for some models)", model.Pricing.Token.Input),
			RefURL:   ref.URL,
		})
	}

	// Check output is typically >= input
	if model.Pricing.Token.Output < model.Pricing.Token.Input {
		warnings = append(warnings, ValidationWarning{
			Provider: model.Provider,
			ModelID:  model.ID,
			Field:    "token.output",
			Message:  "Output price is lower than input price (unusual but may be intentional)",
			RefURL:   ref.URL,
		})
	}

	// Check for missing source metadata
	if model.Source.Kind == "" {
		warnings = append(warnings, ValidationWarning{
			Provider: model.Provider,
			ModelID:  model.ID,
			Field:    "source.kind",
			Message:  "Missing source kind metadata",
			RefURL:   ref.URL,
		})
	}

	// Warn if source is marked as partial
	if model.Source.Partial {
		warnings = append(warnings, ValidationWarning{
			Provider: model.Provider,
			ModelID:  model.ID,
			Field:    "source",
			Message:  "Pricing data is marked as partial (some fields may be missing)",
			RefURL:   ref.URL,
		})
	}

	return warnings
}

// GetReferenceList returns a formatted list of official pricing references
func (v *PricingValidator) GetReferenceList() string {
	var sb strings.Builder

	sb.WriteString("Official Pricing References:\n")
	sb.WriteString("============================\n\n")

	providers := []string{"openai", "anthropic", "deepseek", "google", "azure", "mistral", "cohere"}
	for _, provider := range providers {
		if ref, ok := v.references[provider]; ok {
			sb.WriteString(fmt.Sprintf("Provider: %s\n", ref.Provider))
			sb.WriteString(fmt.Sprintf("  URL: %s\n", ref.URL))
			sb.WriteString(fmt.Sprintf("  Unit: %s\n", ref.Unit))
			sb.WriteString(fmt.Sprintf("  Description: %s\n", ref.Description))
			sb.WriteString(fmt.Sprintf("  Last Checked: %s\n", ref.LastChecked))
			sb.WriteString("\n")
		}
	}

	sb.WriteString("Note: These references should be manually verified periodically.\n")
	sb.WriteString("Use 'boba doctor --pricing' to check for potential issues.\n")

	return sb.String()
}

// FormatWarnings formats validation warnings for display
func FormatWarnings(warnings []ValidationWarning) string {
	if len(warnings) == 0 {
		return "No pricing validation warnings."
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d pricing validation warning(s):\n\n", len(warnings)))

	for i, w := range warnings {
		sb.WriteString(fmt.Sprintf("%d. ", i+1))
		if w.Provider != "" {
			sb.WriteString(fmt.Sprintf("[%s", w.Provider))
			if w.ModelID != "" {
				sb.WriteString(fmt.Sprintf("/%s", w.ModelID))
			}
			sb.WriteString("] ")
		}
		if w.Field != "" {
			sb.WriteString(fmt.Sprintf("(%s) ", w.Field))
		}
		sb.WriteString(w.Message)
		if w.RefURL != "" {
			sb.WriteString(fmt.Sprintf("\n   Reference: %s", w.RefURL))
		}
		sb.WriteString("\n\n")
	}

	return sb.String()
}
