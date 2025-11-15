package pricing

import (
	"context"
	"fmt"

	"github.com/royisme/bobamixer/internal/logging"
)

// LoaderConfig contains configuration for the pricing loader
type LoaderConfig struct {
	EnableOpenRouter bool
	EnableVendorJSON bool
	CacheTTLHours    int
	RefreshOnStartup bool
}

// DefaultLoaderConfig returns the default loader configuration
func DefaultLoaderConfig() LoaderConfig {
	return LoaderConfig{
		EnableOpenRouter: true,
		EnableVendorJSON: true,
		CacheTTLHours:    DefaultCacheTTL,
		RefreshOnStartup: false,
	}
}

// Loader orchestrates loading pricing from various sources
type Loader struct {
	homeDir           string
	config            LoaderConfig
	cache             *CacheManager
	openRouterAdapter *OpenRouterAdapter
	vendorAdapter     *VendorJSONAdapter
}

// NewLoader creates a new pricing loader
func NewLoader(homeDir string, config LoaderConfig) *Loader {
	return &Loader{
		homeDir:           homeDir,
		config:            config,
		cache:             NewCacheManager(homeDir, config.CacheTTLHours),
		openRouterAdapter: NewOpenRouterAdapter(),
		vendorAdapter:     NewVendorJSONAdapter(homeDir),
	}
}

// LoadWithFallback loads pricing with the following fallback chain:
// 1. Try OpenRouter API (if enabled)
// 2. Try cache (if fresh)
// 3. Try vendor JSON
// 4. Return empty schema (will fallback to profiles)
func (l *Loader) LoadWithFallback(ctx context.Context) (*PricingSchema, error) {
	var schema *PricingSchema
	var err error

	// Step 1: Try OpenRouter if enabled and (cache expired OR refresh on startup)
	if l.config.EnableOpenRouter {
		shouldFetch := !l.cache.IsFresh() || l.config.RefreshOnStartup

		if shouldFetch {
			logging.Info("Fetching pricing from OpenRouter API")
			schema, err = l.openRouterAdapter.Fetch(ctx)
			if err != nil {
				logging.Warn("Failed to fetch from OpenRouter", logging.Err(err))
			} else if schema != nil && len(schema.Models) > 0 {
				logging.Info("Successfully fetched pricing from OpenRouter",
					logging.Int("models", len(schema.Models)))

				// Try to merge with vendor JSON for additional models
				if l.config.EnableVendorJSON {
					if vendorSchema, vendorErr := l.vendorAdapter.LoadLocal(); vendorErr == nil {
						logging.Info("Merging with vendor JSON",
							logging.Int("vendor_models", len(vendorSchema.Models)))
						schema = MergeSchemas(schema, vendorSchema)
					}
				}

				// Save to cache
				if saveErr := l.cache.Save(schema, "openrouter"); saveErr != nil {
					logging.Warn("Failed to save cache", logging.Err(saveErr))
				}

				return schema, nil
			}
		}
	}

	// Step 2: Try cache
	logging.Info("Trying to load from cache")
	schema, err = l.cache.Load()
	if err == nil && schema != nil && len(schema.Models) > 0 {
		logging.Info("Successfully loaded pricing from cache",
			logging.Int("models", len(schema.Models)))
		return schema, nil
	}
	if err != nil {
		logging.Info("Cache load failed", logging.Err(err))
	}

	// Step 3: Try vendor JSON
	if l.config.EnableVendorJSON {
		logging.Info("Trying to load from vendor JSON")
		schema, err = l.vendorAdapter.LoadLocal()
		if err == nil && schema != nil && len(schema.Models) > 0 {
			logging.Info("Successfully loaded pricing from vendor JSON",
				logging.Int("models", len(schema.Models)))
			return schema, nil
		}
		if err != nil {
			logging.Info("Vendor JSON load failed", logging.Err(err))
		}
	}

	// Step 4: Return empty schema (will fallback to profiles)
	logging.Info("All pricing sources failed, returning empty schema (will use profile fallback)")
	return NewPricingSchema(), nil
}

// Refresh forces a refresh from OpenRouter
func (l *Loader) Refresh(ctx context.Context) error {
	if !l.config.EnableOpenRouter {
		return fmt.Errorf("OpenRouter is disabled")
	}

	logging.Info("Force refreshing pricing from OpenRouter")
	schema, err := l.openRouterAdapter.Fetch(ctx)
	if err != nil {
		return fmt.Errorf("fetch from OpenRouter: %w", err)
	}

	if len(schema.Models) == 0 {
		return fmt.Errorf("no models returned from OpenRouter")
	}

	// Merge with vendor JSON if enabled
	if l.config.EnableVendorJSON {
		if vendorSchema, vendorErr := l.vendorAdapter.LoadLocal(); vendorErr == nil {
			logging.Info("Merging with vendor JSON",
				logging.Int("vendor_models", len(vendorSchema.Models)))
			schema = MergeSchemas(schema, vendorSchema)
		}
	}

	// Save to cache
	if err := l.cache.Save(schema, "openrouter"); err != nil {
		logging.Warn("Failed to save cache", logging.Err(err))
	}

	logging.Info("Successfully refreshed pricing",
		logging.Int("models", len(schema.Models)))

	return nil
}

// ClearCache clears the pricing cache
func (l *Loader) ClearCache() error {
	return l.cache.Clear()
}

// GetCacheStatus returns cache status information
func (l *Loader) GetCacheStatus() (isFresh bool, metadata *CacheMetadata, err error) {
	isFresh = l.cache.IsFresh()
	metadata, err = l.cache.GetMetadata()
	return isFresh, metadata, err
}
