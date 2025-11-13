package version

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBumpVersion(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "VERSION"), []byte("1.2.3\n"), 0o644); err != nil {
		t.Fatalf("write version: %v", err)
	}
	mgr := NewManager(dir)
	next, err := mgr.Bump("minor", "", "add feature")
	if err != nil {
		t.Fatalf("Bump: %v", err)
	}
	if next != "1.3.0" {
		t.Fatalf("next=%s", next)
	}
	data, _ := os.ReadFile(filepath.Join(dir, "VERSION"))
	if string(data) != "1.3.0\n" {
		t.Fatalf("version file=%s", string(data))
	}
	changelog, err := os.ReadFile(filepath.Join(dir, "CHANGELOG.md"))
	if err != nil {
		t.Fatalf("changelog: %v", err)
	}
	if len(changelog) == 0 {
		t.Fatal("expected changelog entry")
	}
}

func TestPlanPrerelease(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "VERSION"), []byte("0.1.0\n"), 0o644)
	mgr := NewManager(dir)
	next, err := mgr.Plan("patch", "rc.1")
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if next != "0.1.1-rc.1" {
		t.Fatalf("next=%s", next)
	}
}
