// Package pricing provides model pricing information and cost calculation.
package pricing

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/royisme/bobamixer/internal/store/config"
)

// ModelPrice represents the cost of a model per 1k tokens
type ModelPrice struct {
	InputPer1K  float64 `json:"input_per_1k" yaml:"input_per_1k"`
	OutputPer1K float64 `json:"output_per_1k" yaml:"output_per_1k"`
}

// Table contains pricing information for models
type Table struct {
	Models map[string]ModelPrice `json:"models" yaml:"models"`
}

// Load loads pricing table with fallback strategy:
// 1. If pricing.yaml has sources configured, use legacy loader (backward compatibility)
// 2. Otherwise, try new loader (OpenRouter + vendor JSON + cache)
// 3. Fallback to legacy sources (pricing.local.json, pricing.yaml models)
// 4. Fallback to profiles.yaml cost_per_1k
func Load(home string) (*Table, error) {
	pricingCfg, err := config.LoadPricing(home)
	if err != nil {
		pricingCfg = nil
	}

	// If user has configured sources in pricing.yaml, use legacy loader to respect their config
	// This ensures backward compatibility with existing setups
	if pricingCfg != nil && len(pricingCfg.Sources) > 0 {
		return loadLegacy(home)
	}

	// Otherwise, try new loader (OpenRouter + vendor JSON)
	table, err := LoadV2(home)
	if err == nil && table != nil && len(table.Models) > 0 {
		return table, nil
	}

	// Fallback to legacy loader
	return loadLegacy(home)
}

// LoadV2 uses the new pricing loader with OpenRouter and vendor JSON support
func LoadV2(home string) (*Table, error) {
	pricingCfg, err := config.LoadPricing(home)
	if err != nil {
		pricingCfg = nil
	}

	// Build loader config
	loaderCfg := DefaultLoaderConfig()
	if pricingCfg != nil {
		loaderCfg.RefreshOnStartup = pricingCfg.Refresh.OnStartup
		if pricingCfg.Refresh.IntervalHours > 0 {
			loaderCfg.CacheTTLHours = pricingCfg.Refresh.IntervalHours
		}
	}

	// Create loader
	loader := NewLoader(home, loaderCfg)

	// Load with fallback
	ctx := context.Background()
	schema, err := loader.LoadWithFallback(ctx)
	if err != nil {
		return nil, err
	}

	// Convert to legacy table format
	return schema.ToLegacyTable(), nil
}

// loadLegacy uses the legacy loading strategy for backward compatibility
func loadLegacy(home string) (*Table, error) {
	pricingCfg, err := config.LoadPricing(home)
	if err != nil {
		pricingCfg = nil
	}

	cache := filepath.Join(home, "pricing.cache.json")

	if hasRemoteSources(pricingCfg) {
		if t, err := fetchRemote(pricingCfg.Sources, home); err == nil && len(t.Models) > 0 {
			//nolint:errcheck,gosec // saveCache is best effort, failure doesn't affect functionality
			saveCache(cache, t)
			return t, nil
		}
	}

	if t, err := loadCache(cache); err == nil && len(t.Models) > 0 {
		return t, nil
	}

	localPath := filepath.Join(home, "pricing.local.json")
	if t, err := loadJSONFile(localPath); err == nil && len(t.Models) > 0 {
		return t, nil
	}

	// 4) Load from pricing.yaml
	if pricingCfg != nil && len(pricingCfg.Models) > 0 {
		table := &Table{Models: make(map[string]ModelPrice)}
		for name, price := range pricingCfg.Models {
			table.Models[name] = ModelPrice{
				InputPer1K:  price.InputPer1K,
				OutputPer1K: price.OutputPer1K,
			}
		}
		return table, nil
	}

	// 5) Final fallback: empty table (will use profiles.yaml cost_per_1k)
	return &Table{Models: make(map[string]ModelPrice)}, nil
}

func hasRemoteSources(table *config.PricingTable) bool {
	if table == nil {
		return false
	}
	return len(table.Sources) > 0
}

// loadCache loads pricing from cache if it's fresh (< 24h)
func loadCache(path string) (*Table, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	// Check if cache is fresh (< 24h)
	if time.Since(info.ModTime()) > 24*time.Hour {
		return nil, errors.New("cache expired")
	}

	return loadJSONFile(path)
}

// saveCache saves pricing table to cache
func saveCache(path string, table *Table) error {
	data, err := json.MarshalIndent(table, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// loadJSONFile loads pricing from a JSON file
func loadJSONFile(path string) (*Table, error) {
	// #nosec G304 -- path is constructed from safe directory structure
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var table Table
	if err := json.Unmarshal(data, &table); err != nil {
		return nil, err
	}

	return &table, nil
}

// fetchRemote fetches pricing from remote sources
func fetchRemote(sources []config.PricingSource, home string) (*Table, error) {
	if len(sources) == 0 {
		return nil, errors.New("no sources configured")
	}

	// Sort sources by priority (higher priority first)
	sorted := make([]config.PricingSource, len(sources))
	copy(sorted, sources)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Priority > sorted[j].Priority
	})

	// Try each source in priority order
	for _, source := range sorted {
		var table *Table
		var err error

		switch source.Type {
		case "http-json":
			table, err = fetchHTTP(source.URL)
		case "file":
			path := expandHome(source.Path, home)
			table, err = loadJSONFile(path)
		default:
			continue
		}

		if err == nil && len(table.Models) > 0 {
			return table, nil
		}
	}

	return nil, errors.New("all sources failed")
}

// fetchHTTP fetches pricing from HTTP endpoint
func fetchHTTP(url string) (*Table, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		//nolint:errcheck,gosec // Best effort cleanup, error irrelevant in defer
		resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("fetch failed with status: " + resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var table Table
	if err := json.Unmarshal(data, &table); err != nil {
		return nil, err
	}

	return &table, nil
}

// expandHome expands ~ in path to home directory
func expandHome(path, home string) string {
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(home, path[2:])
	}
	return path
}

// GetPrice returns the price for a model, with fallback to profile cost_per_1k
func (t *Table) GetPrice(modelName string, profileCost config.Cost) ModelPrice {
	if price, ok := t.Models[modelName]; ok {
		return price
	}

	// Fallback to profile cost
	return ModelPrice{
		InputPer1K:  profileCost.Input,
		OutputPer1K: profileCost.Output,
	}
}

// CalculateCost calculates the cost for given token usage
func (t *Table) CalculateCost(modelName string, profileCost config.Cost, inputTokens, outputTokens int) (inputCost, outputCost float64) {
	price := t.GetPrice(modelName, profileCost)

	inputCost = float64(inputTokens) / 1000.0 * price.InputPer1K
	outputCost = float64(outputTokens) / 1000.0 * price.OutputPer1K

	return inputCost, outputCost
}
