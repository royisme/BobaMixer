// Package pricing provides model pricing information and cost calculation.
package pricing

import "time"

// SchemaVersion represents the version of the pricing schema
const SchemaVersion = 1

// PricingSchema represents the complete pricing data structure
type PricingSchema struct {
	Version  int             `json:"version"`
	Currency string          `json:"currency"`
	Models   []ModelPricing  `json:"models"`
	FetchedAt time.Time      `json:"fetched_at,omitempty"`
}

// ModelPricing represents comprehensive pricing information for a model
type ModelPricing struct {
	Provider     string        `json:"provider"`               // e.g., "openai", "anthropic"
	ID           string        `json:"id"`                     // Unique model identifier
	DisplayName  string        `json:"display_name,omitempty"` // Human-readable name
	ContextTokens int          `json:"context_tokens,omitempty"`
	Pricing      PricingTiers  `json:"pricing"`
	Source       SourceMeta    `json:"source"`
}

// PricingTiers contains all pricing dimensions for a model
type PricingTiers struct {
	Token         *TokenPricing   `json:"token,omitempty"`
	PerRequest    *RequestPricing `json:"per_request,omitempty"`
	Image         *ImagePricing   `json:"image,omitempty"`
	Audio         *AudioPricing   `json:"audio,omitempty"`
	Tools         *ToolsPricing   `json:"tools,omitempty"`
}

// TokenPricing represents token-based pricing (per million tokens)
type TokenPricing struct {
	Input              float64  `json:"input"`                          // Required: input price per 1M tokens
	Output             float64  `json:"output"`                         // Required: output price per 1M tokens
	CachedInputRead    *float64 `json:"cached_input_read,omitempty"`    // Optional: cache hit read price per 1M tokens
	CachedInputWrite   *float64 `json:"cached_input_write,omitempty"`   // Optional: cache write price per 1M tokens
	InternalReasoning  *float64 `json:"internal_reasoning,omitempty"`   // Optional: internal reasoning tokens (e.g., o1 series)
}

// RequestPricing represents per-request pricing
type RequestPricing struct {
	Request *float64 `json:"request,omitempty"` // Per-request fixed fee
}

// ImagePricing represents image-based pricing
type ImagePricing struct {
	InputPerImage           *float64 `json:"input_per_image,omitempty"`             // Per image input cost
	InputPerMillionTokens   *float64 `json:"input_per_million_tokens,omitempty"`    // Image tokens (if applicable)
}

// AudioPricing represents audio-based pricing
type AudioPricing struct {
	InputPerMillionTokens  *float64 `json:"input_per_million_tokens,omitempty"`   // Audio input tokens
	OutputPerMillionTokens *float64 `json:"output_per_million_tokens,omitempty"`  // Audio output tokens
	InputPerMinute         *float64 `json:"input_per_minute,omitempty"`           // Per minute pricing (e.g., transcription)
	OutputPerMinute        *float64 `json:"output_per_minute,omitempty"`          // Per minute pricing
}

// ToolsPricing represents tool/feature-based pricing
type ToolsPricing struct {
	FileSearchPer1KCalls      *float64 `json:"file_search_per_1k_calls,omitempty"`      // Azure: File Search per 1K calls
	VectorStoreGBDay          *float64 `json:"vector_store_gb_day,omitempty"`           // Azure: Vector storage per GB/day
	ComputerUsePerMillionTokens *float64 `json:"computer_use_per_million_tokens,omitempty"` // Anthropic: Computer Use
	WebSearchPerRequest       *float64 `json:"web_search_per_request,omitempty"`        // Web search per request
}

// SourceMeta contains metadata about the pricing source
type SourceMeta struct {
	Kind      string    `json:"kind"`                // "openrouter", "vendor_json", "html_ref", "profile_fallback"
	URL       string    `json:"url,omitempty"`       // Source URL
	FetchedAt time.Time `json:"fetched_at"`          // When this data was fetched
	Note      string    `json:"note,omitempty"`      // Additional notes (e.g., "unit=per_1M_tokens")
	Partial   bool      `json:"partial,omitempty"`   // True if some fields are missing
}

// CacheMetadata contains cache-related metadata
type CacheMetadata struct {
	FetchedAt   time.Time `json:"fetched_at"`
	TTLHours    int       `json:"ttl_hours"`
	ExpiresAt   time.Time `json:"expires_at"`
	SourceKind  string    `json:"source_kind"`  // Which source was used
}

// CachedPricing wraps PricingSchema with cache metadata
type CachedPricing struct {
	Schema   PricingSchema `json:"schema"`
	Metadata CacheMetadata `json:"metadata"`
}

// IsExpired checks if the cached pricing has expired
func (c *CachedPricing) IsExpired() bool {
	return time.Now().After(c.Metadata.ExpiresAt)
}

// IsFresh checks if the cache is still valid
func (c *CachedPricing) IsFresh() bool {
	return !c.IsExpired()
}

// ToLegacyTable converts PricingSchema to the legacy Table format
// This ensures backward compatibility with existing code
// Only models with valid (non-zero) pricing are included to avoid underestimating costs
func (ps *PricingSchema) ToLegacyTable() *Table {
	table := &Table{
		Models: make(map[string]ModelPrice),
	}

	for _, model := range ps.Models {
		if model.Pricing.Token != nil {
			// Only include models with valid pricing (both input and output > 0)
			// This prevents zero-priced models from being used when pricing data is incomplete
			if model.Pricing.Token.Input > 0 && model.Pricing.Token.Output > 0 {
				// Convert per million to per 1K (divide by 1000)
				table.Models[model.ID] = ModelPrice{
					InputPer1K:  model.Pricing.Token.Input / 1000.0,
					OutputPer1K: model.Pricing.Token.Output / 1000.0,
				}
			}
		}
	}

	return table
}

// NewPricingSchema creates a new pricing schema with default values
func NewPricingSchema() *PricingSchema {
	return &PricingSchema{
		Version:  SchemaVersion,
		Currency: "USD",
		Models:   make([]ModelPricing, 0),
		FetchedAt: time.Now(),
	}
}
