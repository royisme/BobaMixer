package cli

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/royisme/bobamixer/internal/domain/pricing"
)

func TestRunDoctorPricingFetchesOpenRouterData(t *testing.T) {
	home := t.TempDir()

	// Ensure pricing.yaml exists so loader picks up refresh config
	pricingYAML := "refresh:\n  interval_hours: 1\n  on_startup: true\n"
	if err := os.WriteFile(filepath.Join(home, "pricing.yaml"), []byte(pricingYAML), 0o600); err != nil {
		t.Fatalf("failed to write pricing.yaml: %v", err)
	}

	// Mock OpenRouter endpoint
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := fmt.Fprint(w, `{"data":[{"id":"mock/provider-model","name":"Mock Model","context_length":4096,"pricing":{"prompt":"0.00001","completion":"0.00002"}}]}`); err != nil {
			t.Fatalf("failed to write mock response: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	t.Setenv("BOBA_PRICING_OPENROUTER_API", server.URL)

	output, err := captureDoctorPricingOutput(home)
	if err != nil {
		t.Fatalf("runDoctorPricing returned error: %v", err)
	}

	if !strings.Contains(output, "Successfully loaded pricing for 1 model") {
		t.Fatalf("expected pricing load success in output, got: %s", output)
	}

	if !strings.Contains(output, "mock/provider-model") {
		t.Fatalf("expected model ID to appear in output, got: %s", output)
	}

	if !strings.Contains(output, "Cache is fresh") {
		t.Fatalf("expected cache status in output, got: %s", output)
	}
}

func TestRunDoctorPricingUsesCacheMetadata(t *testing.T) {
	home := t.TempDir()

	schema := pricing.NewPricingSchema()
	schema.Models = append(schema.Models, pricing.ModelPricing{
		Provider: "cached",
		ID:       "cached/model",
		Pricing: pricing.PricingTiers{
			Token: &pricing.TokenPricing{Input: 10, Output: 20},
		},
		Source: pricing.SourceMeta{Kind: "test"},
	})

	cache := pricing.NewCacheManager(home, pricing.DefaultCacheTTL)
	if err := cache.Save(schema, "test"); err != nil {
		t.Fatalf("failed to seed cache: %v", err)
	}

	output, err := captureDoctorPricingOutput(home)
	if err != nil {
		t.Fatalf("runDoctorPricing returned error: %v", err)
	}

	if !strings.Contains(output, "Cache is fresh") {
		t.Fatalf("expected fresh cache status, got: %s", output)
	}

	if !strings.Contains(output, "Source: test") {
		t.Fatalf("expected cache source in output, got: %s", output)
	}

	if !strings.Contains(output, "cached/model") {
		t.Fatalf("expected cached model to appear in output, got: %s", output)
	}
}

func captureDoctorPricingOutput(home string) (string, error) {
	var buf bytes.Buffer

	originalStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		return "", err
	}
	os.Stdout = w

	runErr := runDoctorPricing(home)

	if err := w.Close(); err != nil {
		return "", err
	}
	os.Stdout = originalStdout

	if _, err := buf.ReadFrom(r); err != nil {
		return "", err
	}

	if err := r.Close(); err != nil {
		return "", err
	}

	return buf.String(), runErr
}
