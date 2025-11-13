package version

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBumpVersion(t *testing.T) {
	dir := t.TempDir()
	// #nosec G306 -- test file can have readable permissions
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
	// #nosec G304 -- test file in controlled temp directory
	data, _ := os.ReadFile(filepath.Join(dir, "VERSION")) //nolint:errcheck
	if string(data) != "1.3.0\n" {
		t.Fatalf("version file=%s", string(data))
	}
	// #nosec G304 -- test file in controlled temp directory
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
	// #nosec G306 -- test file can have readable permissions
	if err := os.WriteFile(filepath.Join(dir, "VERSION"), []byte("0.1.0\n"), 0o644); err != nil {
		t.Fatalf("failed to write VERSION file: %v", err)
	}
	mgr := NewManager(dir)
	next, err := mgr.Plan("patch", "rc.1")
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if next != "0.1.1-rc.1" {
		t.Fatalf("next=%s", next)
	}
}
