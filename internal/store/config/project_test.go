package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindProjectConfig(t *testing.T) {
	dir := t.TempDir()
	nested := filepath.Join(dir, "a", "b")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	cfgPath := filepath.Join(dir, ".boba-project.yaml")
	yaml := `project:
  name: sample
budget:
  daily_usd: 2.5
  hard_cap: 25
`
	if err := os.WriteFile(cfgPath, []byte(yaml), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	cfg, path, err := FindProjectConfig(nested)
	if err != nil {
		t.Fatalf("FindProjectConfig: %v", err)
	}
	if path != cfgPath {
		t.Fatalf("path=%s", path)
	}
	if cfg.Budget == nil || cfg.Budget.DailyUSD != 2.5 {
		t.Fatalf("budget not parsed")
	}
}
