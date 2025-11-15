package pricing

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	// OpenRouterModelsAPI is the official OpenRouter models endpoint
	OpenRouterModelsAPI = "https://openrouter.ai/api/v1/models"

	// OpenRouterTimeout is the timeout for fetching from OpenRouter
	OpenRouterTimeout = 15 * time.Second
)

// OpenRouterAdapter fetches pricing from OpenRouter Models API
type OpenRouterAdapter struct {
	client  *http.Client
	apiURL  string
}

// NewOpenRouterAdapter creates a new OpenRouter adapter
func NewOpenRouterAdapter() *OpenRouterAdapter {
	return &OpenRouterAdapter{
		client: &http.Client{
			Timeout: OpenRouterTimeout,
		},
		apiURL: OpenRouterModelsAPI,
	}
}

// OpenRouterResponse represents the response from OpenRouter API
type OpenRouterResponse struct {
	Data []OpenRouterModel `json:"data"`
}

// OpenRouterModel represents a model from OpenRouter API
type OpenRouterModel struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	ContextLength   int                    `json:"context_length"`
	Pricing         OpenRouterPricing      `json:"pricing"`
	Architecture    map[string]interface{} `json:"architecture,omitempty"`
}

// OpenRouterPricing represents pricing from OpenRouter
type OpenRouterPricing struct {
	Prompt           string `json:"prompt"`            // Price per token (as string)
	Completion       string `json:"completion"`        // Price per token (as string)
	Request          string `json:"request,omitempty"` // Price per request
	Image            string `json:"image,omitempty"`   // Price per image
	WebSearch        string `json:"web_search,omitempty"`
	InputCacheRead   string `json:"input_cache_read,omitempty"`
	InputCacheWrite  string `json:"input_cache_write,omitempty"`
	InternalReasoning string `json:"internal_reasoning,omitempty"`
}

// Fetch fetches pricing data from OpenRouter
func (a *OpenRouterAdapter) Fetch(ctx context.Context) (*PricingSchema, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "BobaMixer/1.0")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch from OpenRouter: %w", err)
	}
	defer func() {
		//nolint:errcheck,gosec // Best effort cleanup
		resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OpenRouter API returned status %d: %s", resp.StatusCode, string(body))
	}

	var openRouterResp OpenRouterResponse
	if err := json.NewDecoder(resp.Body).Decode(&openRouterResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	// Convert to our schema
	schema := NewPricingSchema()
	schema.FetchedAt = time.Now()

	for _, model := range openRouterResp.Data {
		pricing, partial := a.convertModel(model)
		if pricing != nil {
			pricing.Source.Partial = partial
			schema.Models = append(schema.Models, *pricing)
		}
	}

	return schema, nil
}

// convertModel converts an OpenRouter model to our ModelPricing format
func (a *OpenRouterAdapter) convertModel(model OpenRouterModel) (*ModelPricing, bool) {
	partial := false

	// Parse provider from model ID (format: "provider/model-name")
	provider := "unknown"
	for i, ch := range model.ID {
		if ch == '/' {
			provider = model.ID[:i]
			break
		}
	}

	pricing := &ModelPricing{
		Provider:      provider,
		ID:            model.ID,
		DisplayName:   model.Name,
		ContextTokens: model.ContextLength,
		Pricing: PricingTiers{
			Token: &TokenPricing{},
		},
		Source: SourceMeta{
			Kind:      "openrouter",
			URL:       a.apiURL,
			FetchedAt: time.Now(),
			Note:      "unit=per_1M_tokens",
		},
	}

	// Convert token pricing (OpenRouter returns price per token, we store per million)
	// OpenRouter pricing is in string format, need to parse
	if input, err := parsePrice(model.Pricing.Prompt); err == nil {
		pricing.Pricing.Token.Input = input * 1_000_000
	} else {
		partial = true
	}

	if output, err := parsePrice(model.Pricing.Completion); err == nil {
		pricing.Pricing.Token.Output = output * 1_000_000
	} else {
		partial = true
	}

	// Optional fields
	if cacheRead, err := parsePrice(model.Pricing.InputCacheRead); err == nil {
		val := cacheRead * 1_000_000
		pricing.Pricing.Token.CachedInputRead = &val
	}

	if cacheWrite, err := parsePrice(model.Pricing.InputCacheWrite); err == nil {
		val := cacheWrite * 1_000_000
		pricing.Pricing.Token.CachedInputWrite = &val
	}

	if reasoning, err := parsePrice(model.Pricing.InternalReasoning); err == nil {
		val := reasoning * 1_000_000
		pricing.Pricing.Token.InternalReasoning = &val
	}

	// Per-request pricing
	if request, err := parsePrice(model.Pricing.Request); err == nil {
		pricing.Pricing.PerRequest = &RequestPricing{
			Request: &request,
		}
	}

	// Image pricing
	if image, err := parsePrice(model.Pricing.Image); err == nil {
		pricing.Pricing.Image = &ImagePricing{
			InputPerImage: &image,
		}
	}

	// Tools pricing (web search)
	if webSearch, err := parsePrice(model.Pricing.WebSearch); err == nil {
		if pricing.Pricing.Tools == nil {
			pricing.Pricing.Tools = &ToolsPricing{}
		}
		pricing.Pricing.Tools.WebSearchPerRequest = &webSearch
	}

	// Set partial flag in source metadata
	pricing.Source.Partial = partial

	return pricing, partial
}

// parsePrice parses a price string to float64
// OpenRouter returns prices as strings (e.g., "0.000003")
func parsePrice(priceStr string) (float64, error) {
	if priceStr == "" {
		return 0, fmt.Errorf("empty price")
	}

	var price float64
	_, err := fmt.Sscanf(priceStr, "%f", &price)
	if err != nil {
		return 0, fmt.Errorf("parse price: %w", err)
	}

	return price, nil
}
