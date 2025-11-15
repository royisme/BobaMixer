package pricing

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCacheManagerSaveAndLoad(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Create cache manager
	cm := NewCacheManager(tmpDir, 24)

	// Create test schema
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
						Input:  30.0,
						Output: 60.0,
					},
				},
			},
		},
		FetchedAt: time.Now(),
	}

	// Save to cache
	err := cm.Save(schema, "openrouter")
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify cache file exists
	cachePath := filepath.Join(tmpDir, "pricing.cache.json")
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		t.Fatal("Cache file was not created")
	}

	// Load from cache
	loaded, err := cm.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify loaded data
	if len(loaded.Models) != 1 {
		t.Fatalf("Expected 1 model, got %d", len(loaded.Models))
	}

	if loaded.Models[0].ID != "gpt-4" {
		t.Errorf("Model ID: got %s, want gpt-4", loaded.Models[0].ID)
	}

	if loaded.Models[0].Pricing.Token.Input != 30.0 {
		t.Errorf("Input price: got %f, want 30.0", loaded.Models[0].Pricing.Token.Input)
	}
}

func TestCacheManagerExpiry(t *testing.T) {
	tmpDir := t.TempDir()

	// Create cache manager with 1 hour TTL
	cm := NewCacheManager(tmpDir, 1)

	schema := NewPricingSchema()
	schema.Models = []ModelPricing{
		{
			Provider: "test",
			ID:       "test-model",
			Pricing:  PricingTiers{Token: &TokenPricing{Input: 1.0, Output: 2.0}},
		},
	}

	// Save cache
	if err := cm.Save(schema, "test"); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Should be fresh
	if !cm.IsFresh() {
		t.Error("Cache should be fresh immediately after save")
	}

	// Manually modify cache to be expired
	cachePath := filepath.Join(tmpDir, "pricing.cache.json")
	// #nosec G304 -- test file
	data, _ := os.ReadFile(cachePath)

	// Create expired cache
	expiredCache := CachedPricing{
		Schema: *schema,
		Metadata: CacheMetadata{
			FetchedAt:  time.Now().Add(-2 * time.Hour),
			TTLHours:   1,
			ExpiresAt:  time.Now().Add(-1 * time.Hour),
			SourceKind: "test",
		},
	}

	// Save expired cache
	cm2 := NewCacheManager(tmpDir, 1)
	if err := cm2.Save(&expiredCache.Schema, "test"); err == nil {
		// Manually set old expiry
		cached := &CachedPricing{
			Schema: expiredCache.Schema,
			Metadata: CacheMetadata{
				FetchedAt:  time.Now().Add(-2 * time.Hour),
				TTLHours:   1,
				ExpiresAt:  time.Now().Add(-1 * time.Hour),
				SourceKind: "test",
			},
		}

		// Write expired cache manually
		expiredData, _ := json.Marshal(cached)
		//nolint:errcheck,gosec // test code
		os.WriteFile(cachePath, expiredData, 0600)
	}

	// Should not be fresh
	if cm2.IsFresh() {
		t.Error("Cache should not be fresh after expiry")
	}

	// Load should fail
	_, err := cm2.Load()
	if err == nil {
		t.Error("Load should fail for expired cache")
	}

	// Restore original data
	//nolint:errcheck,gosec // test code
	os.WriteFile(cachePath, data, 0600)
}

func TestCacheManagerClear(t *testing.T) {
	tmpDir := t.TempDir()

	cm := NewCacheManager(tmpDir, 24)

	// Save some data
	schema := NewPricingSchema()
	if err := cm.Save(schema, "test"); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify file exists
	cachePath := filepath.Join(tmpDir, "pricing.cache.json")
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		t.Fatal("Cache file should exist")
	}

	// Clear cache
	if err := cm.Clear(); err != nil {
		t.Fatalf("Clear failed: %v", err)
	}

	// Verify file is gone
	if _, err := os.Stat(cachePath); !os.IsNotExist(err) {
		t.Error("Cache file should be deleted")
	}

	// Clearing again should not error
	if err := cm.Clear(); err != nil {
		t.Errorf("Clear on non-existent cache should not error: %v", err)
	}
}

func TestCacheManagerGetMetadata(t *testing.T) {
	tmpDir := t.TempDir()

	cm := NewCacheManager(tmpDir, 24)

	schema := NewPricingSchema()
	if err := cm.Save(schema, "openrouter"); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	meta, err := cm.GetMetadata()
	if err != nil {
		t.Fatalf("GetMetadata failed: %v", err)
	}

	if meta.TTLHours != 24 {
		t.Errorf("TTLHours: got %d, want 24", meta.TTLHours)
	}

	if meta.SourceKind != "openrouter" {
		t.Errorf("SourceKind: got %s, want openrouter", meta.SourceKind)
	}

	if meta.FetchedAt.IsZero() {
		t.Error("FetchedAt should not be zero")
	}

	if meta.ExpiresAt.IsZero() {
		t.Error("ExpiresAt should not be zero")
	}

	// Verify expiry is approximately 24 hours from now
	expectedExpiry := time.Now().Add(24 * time.Hour)
	if meta.ExpiresAt.Before(expectedExpiry.Add(-1*time.Minute)) || meta.ExpiresAt.After(expectedExpiry.Add(1*time.Minute)) {
		t.Errorf("ExpiresAt is not approximately 24 hours from now: %v", meta.ExpiresAt)
	}
}
