package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestConfigMergerMerge(t *testing.T) {
	dir := t.TempDir()
	if err := SaveActiveProfile(dir, "global-default"); err != nil {
		t.Fatalf("SaveActiveProfile: %v", err)
	}
	routesYAML := `routes:
  - id: r1
    if: "true"
    use: "global"
`
	if err := os.WriteFile(filepath.Join(dir, "routes.yaml"), []byte(routesYAML), 0o600); err != nil {
		t.Fatalf("write routes: %v", err)
	}
	merger := NewConfigMerger(dir)
	overrides := map[string]interface{}{"profile": "session-profile"}
	merged, err := merger.Merge("projectA", "branchB", overrides)
	if err != nil {
		t.Fatalf("Merge: %v", err)
	}
	if merged.ActiveProfile != "session-profile" {
		t.Fatalf("expected session override, got %s", merged.ActiveProfile)
	}
	if merged.Routes == nil || len(merged.Overrides) != 4 {
		t.Fatalf("unexpected merged result: %#v", merged)
	}
}

func TestGetEffectiveProfile(t *testing.T) {
	dir := t.TempDir()
	if err := SaveActiveProfile(dir, "base-profile"); err != nil {
		t.Fatalf("SaveActiveProfile: %v", err)
	}
	merger := NewConfigMerger(dir)
	profile, order := merger.GetEffectiveProfile("proj", "branch", "session")
	if profile != "session" {
		t.Fatalf("expected session profile, got %s", profile)
	}
	expected := []string{"global:base-profile", "project:proj", "branch:branch", "session:session"}
	if !reflect.DeepEqual(order, expected) {
		t.Fatalf("unexpected overrides: %#v", order)
	}
}

func TestResolveConfigOrder(t *testing.T) {
	order := ResolveConfigOrder()
	if len(order) != 4 {
		t.Fatalf("expected four order entries, got %d", len(order))
	}
}
