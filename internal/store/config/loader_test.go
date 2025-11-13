package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadProfiles(t *testing.T) {
	dir := t.TempDir()
	data := `profiles:
  quick:
    name: "Quick"
    adapter: "http"
    provider: "openrouter"
    endpoint: "https://example"
    model: "demo"
    max_tokens: 256
    temperature: 0.3
    tags: ["a","b"]
    cost_per_1k:
      input: 0.1
      output: 0.2
    env:
      API_KEY: "secret://demo"
`
	if err := os.WriteFile(filepath.Join(dir, "profiles.yaml"), []byte(data), 0600); err != nil {
		t.Fatalf("write profiles: %v", err)
	}
	profs, err := LoadProfiles(dir)
	if err != nil {
		t.Fatalf("LoadProfiles: %v", err)
	}
	quick, ok := profs["quick"]
	if !ok {
		t.Fatalf("missing quick profile")
	}
	if quick.Name != "Quick" || quick.MaxTokens != 256 || quick.CostPer1K.Input != 0.1 {
		t.Fatalf("unexpected profile: %#v", quick)
	}
}

func TestLoadSecrets(t *testing.T) {
	dir := t.TempDir()
	data := `secrets:
  demo: "value"
`
	if err := os.WriteFile(filepath.Join(dir, "secrets.yaml"), []byte(data), 0600); err != nil {
		t.Fatalf("write secrets: %v", err)
	}
	sec, err := LoadSecrets(dir)
	if err != nil {
		t.Fatalf("LoadSecrets: %v", err)
	}
	if sec["demo"] != "value" {
		t.Fatalf("unexpected secrets: %#v", sec)
	}
}
