package cli

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/royisme/bobamixer/internal/settings"
)

func TestRunInitWritesSettingsFile(t *testing.T) {
	t.Parallel()

	home := filepath.Join(t.TempDir(), ".boba")
	if err := os.MkdirAll(home, 0o700); err != nil {
		t.Fatalf("mkdir home: %v", err)
	}
	if err := runInit(home, []string{"--mode", string(settings.ModeApply), "--theme", "dark", "--explore-rate", "0.2"}); err != nil {
		t.Fatalf("runInit error: %v", err)
	}
	ctx := context.Background()
	loaded, err := settings.Load(ctx, home)
	if err != nil {
		t.Fatalf("Load settings: %v", err)
	}
	if loaded.Mode != settings.ModeApply {
		t.Fatalf("Mode = %s, want apply", loaded.Mode)
	}
	if loaded.Theme != "dark" {
		t.Fatalf("Theme = %s, want dark", loaded.Theme)
	}
	if !loaded.Explore.Enabled || loaded.Explore.Rate != 0.2 {
		t.Fatalf("Explore not configured correctly: %+v", loaded.Explore)
	}
}
