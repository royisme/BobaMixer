package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadRoutesDefaults(t *testing.T) {
	dir := t.TempDir()
	cfg, err := LoadRoutes(dir)
	if err != nil {
		t.Fatalf("LoadRoutes: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected config")
	}
	if cfg.Explore.Rate != 0.03 || !cfg.Explore.Enabled {
		t.Fatalf("unexpected explore defaults: %+v", cfg.Explore)
	}
	if len(cfg.SubAgents) != 0 {
		t.Fatalf("expected no sub agents, got %d", len(cfg.SubAgents))
	}
}

func TestLoadRoutesParsesRules(t *testing.T) {
	dir := t.TempDir()
	contents := `sub_agents:
  summarizer:
    profile: "fast"
    triggers: ["summary"]
    conditions:
      project: "docs"
rules:
  - id: "1"
    if: "input.contains('summary')"
    use: "summarizer"
    fallback: "default"
    explain: "Prefer faster profile for summaries"
explore:
  enabled: false
  rate: 0.1
`
	if err := os.WriteFile(filepath.Join(dir, "routes.yaml"), []byte(contents), 0o600); err != nil {
		t.Fatalf("write routes: %v", err)
	}
	cfg, err := LoadRoutes(dir)
	if err != nil {
		t.Fatalf("LoadRoutes: %v", err)
	}
	agent, ok := cfg.SubAgents["summarizer"]
	if !ok || agent.Profile != "fast" || len(agent.Triggers) != 1 {
		t.Fatalf("unexpected sub agent: %+v", agent)
	}
	if len(cfg.Rules) != 1 || cfg.Rules[0].ID != "1" {
		t.Fatalf("unexpected rules: %#v", cfg.Rules)
	}
	if cfg.Explore.Enabled {
		t.Fatalf("explore should be disabled")
	}
	if cfg.Explore.Rate != 0.1 {
		t.Fatalf("explore rate mismatch: %v", cfg.Explore.Rate)
	}
}

func TestValidateSecretsPermissions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "secrets.yaml")
	if err := os.WriteFile(path, []byte("secrets: {}"), 0o600); err != nil {
		t.Fatalf("write secrets: %v", err)
	}
	if err := ValidateSecretsPermissions(dir); err != nil {
		t.Fatalf("expected valid permissions, got %v", err)
	}
	if err := os.Chmod(path, 0o644); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	err := ValidateSecretsPermissions(dir)
	if err == nil {
		t.Fatal("expected permission error")
	}
	if !strings.Contains(err.Error(), "chmod 600") {
		t.Fatalf("expected chmod hint, got %v", err)
	}
}
