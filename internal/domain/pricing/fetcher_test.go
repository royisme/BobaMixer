package pricing

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/royisme/bobamixer/internal/store/config"
)

func TestGetPrice(t *testing.T) {
	table := &Table{
		Models: map[string]ModelPrice{
			"claude-3-5-sonnet": {
				InputPer1K:  0.015,
				OutputPer1K: 0.075,
			},
			"deepseek-chat": {
				InputPer1K:  0.0005,
				OutputPer1K: 0.002,
			},
		},
	}

	profileCost := config.Cost{
		Input:  0.01,
		Output: 0.05,
	}

	tests := []struct {
		name           string
		modelName      string
		expectedInput  float64
		expectedOutput float64
	}{
		{
			name:           "existing model",
			modelName:      "claude-3-5-sonnet",
			expectedInput:  0.015,
			expectedOutput: 0.075,
		},
		{
			name:           "another existing model",
			modelName:      "deepseek-chat",
			expectedInput:  0.0005,
			expectedOutput: 0.002,
		},
		{
			name:           "fallback to profile cost",
			modelName:      "unknown-model",
			expectedInput:  0.01,
			expectedOutput: 0.05,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			price := table.GetPrice(tt.modelName, profileCost)

			if price.InputPer1K != tt.expectedInput {
				t.Errorf("InputPer1K: got %f, want %f", price.InputPer1K, tt.expectedInput)
			}
			if price.OutputPer1K != tt.expectedOutput {
				t.Errorf("OutputPer1K: got %f, want %f", price.OutputPer1K, tt.expectedOutput)
			}
		})
	}
}

func TestCalculateCost(t *testing.T) {
	table := &Table{
		Models: map[string]ModelPrice{
			"test-model": {
				InputPer1K:  0.01,
				OutputPer1K: 0.02,
			},
		},
	}

	profileCost := config.Cost{
		Input:  0.01,
		Output: 0.02,
	}

	tests := []struct {
		name               string
		modelName          string
		inputTokens        int
		outputTokens       int
		expectedInputCost  float64
		expectedOutputCost float64
	}{
		{
			name:               "1k tokens each",
			modelName:          "test-model",
			inputTokens:        1000,
			outputTokens:       1000,
			expectedInputCost:  0.01,
			expectedOutputCost: 0.02,
		},
		{
			name:               "500 tokens each",
			modelName:          "test-model",
			inputTokens:        500,
			outputTokens:       500,
			expectedInputCost:  0.005,
			expectedOutputCost: 0.01,
		},
		{
			name:               "zero tokens",
			modelName:          "test-model",
			inputTokens:        0,
			outputTokens:       0,
			expectedInputCost:  0.0,
			expectedOutputCost: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputCost, outputCost := table.CalculateCost(
				tt.modelName,
				profileCost,
				tt.inputTokens,
				tt.outputTokens,
			)

			if inputCost != tt.expectedInputCost {
				t.Errorf("inputCost: got %f, want %f", inputCost, tt.expectedInputCost)
			}
			if outputCost != tt.expectedOutputCost {
				t.Errorf("outputCost: got %f, want %f", outputCost, tt.expectedOutputCost)
			}
		})
	}
}

func TestExpandHome(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		home     string
		expected string
	}{
		{
			name:     "expand tilde",
			path:     "~/.boba/config.yaml",
			home:     "/home/user/.boba",
			expected: "/home/user/.boba/.boba/config.yaml",
		},
		{
			name:     "no tilde",
			path:     "/absolute/path",
			home:     "/home/user/.boba",
			expected: "/absolute/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandHome(tt.path, tt.home)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestLoadPrefersRemoteBeforeCache(t *testing.T) {
	home := t.TempDir()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte(`{"models":{"remote-model":{"input_per_1k":0.02,"output_per_1k":0.03}}}`)); err != nil {
			panic(err)
		}
	}))
	t.Cleanup(srv.Close)

	pricingYAML := fmt.Sprintf(`sources:
  - type: "http-json"
    url: %q
    priority: 10
refresh:
  on_startup: true
`, srv.URL)
	if err := os.WriteFile(filepath.Join(home, "pricing.yaml"), []byte(pricingYAML), 0600); err != nil {
		t.Fatalf("write pricing.yaml: %v", err)
	}

	if cfg, err := config.LoadPricing(home); err != nil {
		t.Fatalf("LoadPricing: %v", err)
	} else if len(cfg.Sources) == 0 {
		t.Fatalf("expected sources to be parsed")
	}

	cachePayload := []byte(`{"models":{"cache-model":{"input_per_1k":9.9,"output_per_1k":9.9}}}`)
	if err := os.WriteFile(filepath.Join(home, "pricing.cache.json"), cachePayload, 0600); err != nil {
		t.Fatalf("write cache: %v", err)
	}

	table, err := Load(home)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if _, ok := table.Models["remote-model"]; !ok {
		t.Fatalf("remote data not returned, got: %#v", table.Models)
	}
	if _, ok := table.Models["cache-model"]; ok {
		t.Fatalf("expected cache data to be ignored when remote succeeds")
	}
}

func TestLoadFallsBackToCacheWhenRemoteFails(t *testing.T) {
	home := t.TempDir()
	pricingYAML := `sources:
  - type: "http-json"
    url: "http://127.0.0.1:0"
    priority: 10
refresh:
  on_startup: true
`
	if err := os.WriteFile(filepath.Join(home, "pricing.yaml"), []byte(pricingYAML), 0600); err != nil {
		t.Fatalf("write pricing.yaml: %v", err)
	}

	cachePayload := []byte(`{"models":{"cache-model":{"input_per_1k":1.1,"output_per_1k":2.2}}}`)
	if err := os.WriteFile(filepath.Join(home, "pricing.cache.json"), cachePayload, 0600); err != nil {
		t.Fatalf("write cache: %v", err)
	}

	table, err := Load(home)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if _, ok := table.Models["cache-model"]; !ok {
		t.Fatalf("expected cache data to be used when remote fails")
	}
}
