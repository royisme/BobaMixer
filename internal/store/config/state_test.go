package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestActiveProfileState(t *testing.T) {
	dir := t.TempDir()
	// Loading before file exists should return empty
	prof, err := LoadActiveProfile(dir)
	if err != nil {
		t.Fatalf("LoadActiveProfile: %v", err)
	}
	if prof != "" {
		t.Fatalf("expected empty, got %q", prof)
	}

	if err := SaveActiveProfile(dir, "work-heavy"); err != nil {
		t.Fatalf("SaveActiveProfile: %v", err)
	}

	prof, err = LoadActiveProfile(dir)
	if err != nil {
		t.Fatalf("LoadActiveProfile 2: %v", err)
	}
	if prof != "work-heavy" {
		t.Fatalf("expected work-heavy, got %s", prof)
	}

	path := filepath.Join(dir, "active_profile")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("expected 600 perms, got %o", info.Mode().Perm())
	}
}
