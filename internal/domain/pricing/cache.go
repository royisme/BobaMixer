package pricing

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	// DefaultCacheTTL is the default cache TTL in hours
	DefaultCacheTTL = 24
)

// CacheManager handles pricing cache operations
type CacheManager struct {
	cacheDir string
	ttlHours int
}

// NewCacheManager creates a new cache manager
func NewCacheManager(cacheDir string, ttlHours int) *CacheManager {
	if ttlHours <= 0 {
		ttlHours = DefaultCacheTTL
	}

	return &CacheManager{
		cacheDir: cacheDir,
		ttlHours: ttlHours,
	}
}

// Load loads pricing from cache if it's fresh
func (cm *CacheManager) Load() (*PricingSchema, error) {
	cachePath := filepath.Join(cm.cacheDir, "pricing.cache.json")

	// Check if cache file exists
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("cache not found")
	}

	// #nosec G304 -- path is from safe home directory structure
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, fmt.Errorf("read cache: %w", err)
	}

	var cached CachedPricing
	if err := json.Unmarshal(data, &cached); err != nil {
		return nil, fmt.Errorf("parse cache: %w", err)
	}

	// Check if cache is expired
	if cached.IsExpired() {
		return nil, fmt.Errorf("cache expired at %s", cached.Metadata.ExpiresAt.Format(time.RFC3339))
	}

	return &cached.Schema, nil
}

// Save saves pricing schema to cache with metadata
func (cm *CacheManager) Save(schema *PricingSchema, sourceKind string) error {
	cachePath := filepath.Join(cm.cacheDir, "pricing.cache.json")

	now := time.Now()
	cached := CachedPricing{
		Schema: *schema,
		Metadata: CacheMetadata{
			FetchedAt:  now,
			TTLHours:   cm.ttlHours,
			ExpiresAt:  now.Add(time.Duration(cm.ttlHours) * time.Hour),
			SourceKind: sourceKind,
		},
	}

	data, err := json.MarshalIndent(cached, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal cache: %w", err)
	}

	if err := os.WriteFile(cachePath, data, 0600); err != nil {
		return fmt.Errorf("write cache: %w", err)
	}

	return nil
}

// Clear removes the cache file
func (cm *CacheManager) Clear() error {
	cachePath := filepath.Join(cm.cacheDir, "pricing.cache.json")

	if err := os.Remove(cachePath); err != nil {
		if os.IsNotExist(err) {
			return nil // Already cleared
		}
		return fmt.Errorf("clear cache: %w", err)
	}

	return nil
}

// GetMetadata returns cache metadata without loading the full schema
func (cm *CacheManager) GetMetadata() (*CacheMetadata, error) {
	cachePath := filepath.Join(cm.cacheDir, "pricing.cache.json")

	// #nosec G304 -- path is from safe home directory structure
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, fmt.Errorf("read cache: %w", err)
	}

	var cached CachedPricing
	if err := json.Unmarshal(data, &cached); err != nil {
		return nil, fmt.Errorf("parse cache: %w", err)
	}

	return &cached.Metadata, nil
}

// IsFresh checks if cache exists and is fresh without loading the full data
func (cm *CacheManager) IsFresh() bool {
	meta, err := cm.GetMetadata()
	if err != nil {
		return false
	}

	return time.Now().Before(meta.ExpiresAt)
}
