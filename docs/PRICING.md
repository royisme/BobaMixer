# Pricing System

BobaMixer's pricing system provides automated, up-to-date model pricing information with support for multiple pricing dimensions and sources.

## Features

- **Multi-source pricing**: OpenRouter API (automatic), vendor JSON (manual), and fallback to profiles
- **Comprehensive pricing dimensions**: Token pricing, cache pricing, per-request fees, image/audio pricing, and tool-based pricing
- **Automatic updates**: Configurable cache with TTL-based refresh
- **Validation**: Built-in validation against official pricing references
- **Unit standardization**: All prices normalized to "per million tokens" (USD)

## Architecture

### Data Flow

```
OpenRouter API → Cache (24h TTL) → Vendor JSON → Profile Fallback
     ↓                                    ↓
  Adapters                           Validators
     ↓                                    ↓
   Unified Schema                   Reference URLs
```

### Components

1. **Schema** (`schema.go`): Unified pricing data structure
2. **Adapters**:
   - `OpenRouterAdapter`: Fetches from OpenRouter Models API
   - `VendorJSONAdapter`: Loads from local vendor JSON file
3. **Cache Manager**: Handles caching with TTL and metadata
4. **Loader**: Orchestrates fallback chain
5. **Validator**: Validates pricing against known references

## Configuration

### pricing.yaml

```yaml
# Enable remote sources (optional)
sources:
  - type: "http-json"
    url: "https://openrouter.ai/api/v1/models"
    priority: 10

# Refresh settings
refresh:
  on_startup: true      # Fetch on startup
  interval_hours: 24    # Cache TTL in hours

# Manual pricing overrides (optional)
models:
  custom-model:
    input_per_1k: 0.01
    output_per_1k: 0.02
```

### Pricing Schema

The unified pricing schema supports multiple dimensions:

```json
{
  "version": 1,
  "currency": "USD",
  "models": [
    {
      "provider": "openai",
      "id": "gpt-4-turbo",
      "display_name": "GPT-4 Turbo",
      "context_tokens": 128000,
      "pricing": {
        "token": {
          "input": 10.0,                    // Per 1M tokens
          "output": 30.0,
          "cached_input_read": 5.0,         // Cache hit pricing
          "cached_input_write": 5.0,        // Cache write pricing
          "internal_reasoning": null        // For o1-style models
        },
        "per_request": {
          "request": null                   // Fixed per-request fee
        },
        "image": {
          "input_per_image": null,
          "input_per_million_tokens": null
        },
        "audio": {
          "input_per_million_tokens": null,
          "output_per_million_tokens": null,
          "input_per_minute": null          // For transcription
        },
        "tools": {
          "file_search_per_1k_calls": null,
          "vector_store_gb_day": null,
          "computer_use_per_million_tokens": null,
          "web_search_per_request": null
        }
      },
      "source": {
        "kind": "openrouter",
        "url": "https://openrouter.ai/api/v1/models",
        "fetched_at": "2025-11-14T00:00:00Z",
        "note": "unit=per_1M_tokens",
        "partial": false
      }
    }
  ]
}
```

## Usage

### Loading Pricing

```go
import "github.com/royisme/bobamixer/internal/domain/pricing"

// Load pricing with automatic fallback
table, err := pricing.Load(homeDir)
if err != nil {
    log.Fatal(err)
}

// Get price for a model
price := table.GetPrice("gpt-4-turbo", profileCost)
fmt.Printf("Input: $%f per 1K tokens\n", price.InputPer1K)
```

### Using the New Loader (Advanced)

```go
// Create loader with custom config
config := pricing.LoaderConfig{
    EnableOpenRouter: true,
    EnableVendorJSON: true,
    CacheTTLHours:    24,
    RefreshOnStartup: true,
}

loader := pricing.NewLoader(homeDir, config)

// Load with fallback
schema, err := loader.LoadWithFallback(context.Background())

// Force refresh
err = loader.Refresh(context.Background())

// Check cache status
isFresh, meta, err := loader.GetCacheStatus()
```

### Validation

```go
validator := pricing.NewPricingValidator()

// Validate schema
warnings := validator.ValidateAgainstRefs(schema)
if len(warnings) > 0 {
    fmt.Println(pricing.FormatWarnings(warnings))
}

// Get reference list
refList := validator.GetReferenceList()
fmt.Println(refList)
```

## Vendor JSON Maintenance

The vendor JSON file (`~/.boba/pricing.vendor.json`) is used for:
- Models not available in OpenRouter
- Manual overrides
- Offline pricing data

### Creating/Updating Vendor JSON

1. Use the example file: `configs/examples/pricing.vendor.json`
2. Copy to `~/.boba/pricing.vendor.json`
3. Update pricing from official sources:
   - OpenAI: https://openai.com/api/pricing/
   - Anthropic: https://www.anthropic.com/pricing
   - DeepSeek: https://platform.deepseek.com/api-docs/pricing/
   - Google: https://ai.google.dev/pricing
   - Azure: https://azure.microsoft.com/pricing/details/cognitive-services/openai-service/

### Saving Vendor JSON Programmatically

```go
adapter := pricing.NewVendorJSONAdapter(homeDir)

// Create or load schema
schema := pricing.NewPricingSchema()

// Add models...
schema.Models = append(schema.Models, pricing.ModelPricing{
    Provider: "openai",
    ID:       "gpt-4-turbo",
    // ... pricing details
})

// Save
err := adapter.Save(schema)
```

## Official Pricing References

The system validates against these official sources:

| Provider   | URL                                                                      | Unit            |
|------------|--------------------------------------------------------------------------|-----------------|
| OpenAI     | https://openai.com/api/pricing/                                         | per_1M_tokens   |
| Anthropic  | https://www.anthropic.com/pricing                                       | per_1M_tokens   |
| DeepSeek   | https://platform.deepseek.com/api-docs/pricing/                         | per_1M_tokens   |
| Google     | https://ai.google.dev/pricing                                           | per_1M_tokens   |
| Azure      | https://azure.microsoft.com/pricing/details/cognitive-services/openai-service/ | varies  |
| Mistral    | https://mistral.ai/technology/#pricing                                  | per_1M_tokens   |
| Cohere     | https://cohere.com/pricing                                              | per_1M_tokens   |

## Cache Management

### Cache Location

- Cache file: `~/.boba/pricing.cache.json`
- Default TTL: 24 hours
- Format: Includes both schema and metadata

### Cache Structure

```json
{
  "schema": { /* PricingSchema */ },
  "metadata": {
    "fetched_at": "2025-11-14T00:00:00Z",
    "ttl_hours": 24,
    "expires_at": "2025-11-15T00:00:00Z",
    "source_kind": "openrouter"
  }
}
```

### Manual Cache Operations

```go
cacheManager := pricing.NewCacheManager(homeDir, 24)

// Load cache
schema, err := cacheManager.Load()

// Save cache
err = cacheManager.Save(schema, "openrouter")

// Clear cache
err = cacheManager.Clear()

// Check if fresh
isFresh := cacheManager.IsFresh()

// Get metadata only
meta, err := cacheManager.GetMetadata()
```

## Best Practices

1. **Use OpenRouter as primary source**: It provides the most comprehensive and up-to-date pricing
2. **Maintain vendor JSON for gaps**: Add models not available in OpenRouter
3. **Validate periodically**: Run validation to catch pricing drift
4. **Set appropriate TTL**: 24 hours is recommended for production
5. **Monitor cache status**: Check cache freshness before critical operations
6. **Document pricing sources**: Always include source URL and fetched_at timestamp

## Troubleshooting

### Pricing not updating

1. Check cache status: `loader.GetCacheStatus()`
2. Force refresh: `loader.Refresh(ctx)`
3. Clear cache: `loader.ClearCache()`
4. Verify network access to OpenRouter API

### Missing models

1. Check if model is in OpenRouter: https://openrouter.ai/models
2. Add to vendor JSON if not available
3. Ensure model ID matches provider format

### Validation warnings

1. Review warnings: `pricing.FormatWarnings(warnings)`
2. Check official pricing pages
3. Update vendor JSON if needed
4. Verify unit conversions (1K → 1M)

## Future Enhancements

- Regional pricing support (Azure)
- Time-based pricing (DeepSeek off-peak)
- Batch pricing endpoints
- Automatic vendor JSON updates
- Price history tracking
- Cost forecasting

## Related Documentation

- [Configuration Guide](./CONFIGURATION.md)
- [CLI Commands](./CLI.md)
- [API Reference](./API.md)
