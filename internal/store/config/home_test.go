package config

import (
	"path/filepath"
	"testing"
)

func TestResolveHomeCustom(t *testing.T) {
	t.Setenv("BOBA_HOME", "/tmp/boba-home")
	dir, err := ResolveHome()
	if err != nil {
		t.Fatalf("ResolveHome custom: %v", err)
	}
	if dir != "/tmp/boba-home" {
		t.Fatalf("expected custom dir, got %s", dir)
	}
}

func TestResolveHomeDefault(t *testing.T) {
	t.Setenv("BOBA_HOME", "")
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	dir, err := ResolveHome()
	if err != nil {
		t.Fatalf("ResolveHome default: %v", err)
	}
	if dir != filepath.Join(tmp, ".boba") {
		t.Fatalf("unexpected default dir: %s", dir)
	}
}
