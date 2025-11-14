package secrets

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidatePermissions(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "secrets.yaml")
	if err := os.WriteFile(path, []byte("key: value\n"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if err := ValidatePermissions(path); err != nil {
		t.Fatalf("ValidatePermissions() unexpected error = %v", err)
	}

	if err := os.Chmod(path, 0o644); err != nil {
		t.Fatalf("Chmod() error = %v", err)
	}

	if err := ValidatePermissions(path); err == nil {
		t.Fatalf("expected permissions error, got nil")
	}
}

func TestResolveEnv(t *testing.T) {
	t.Parallel()

	secrets := &Secrets{Values: map[string]string{"anthropic": "secret-value"}}
	env, missing := ResolveEnv(map[string]string{
		"ANTHROPIC_API_KEY": "secret://anthropic",
		"NODE_ENV":          "production",
	}, secrets)

	if len(missing) != 0 {
		t.Fatalf("expected no missing secrets, got %v", missing)
	}

	want := map[string]bool{
		"ANTHROPIC_API_KEY=secret-value": true,
		"NODE_ENV=production":            true,
	}
	for _, kv := range env {
		if !want[kv] {
			t.Fatalf("unexpected env entry %q", kv)
		}
	}
}

func TestResolveEnvMissing(t *testing.T) {
	t.Parallel()

	env, missing := ResolveEnv(map[string]string{
		"OPENAI_API_KEY": "secret://openai",
	}, &Secrets{Values: map[string]string{}})

	if len(env) != 0 {
		t.Fatalf("expected empty env, got %v", env)
	}
	if len(missing) != 1 || missing[0] != "openai" {
		t.Fatalf("expected missing openai, got %v", missing)
	}
}

func TestLoadSecretsFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "secrets.yaml")
	contents := "anthropic: demo\nopenai: foo\n"
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	s, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if got := s.Values["anthropic"]; got != "demo" {
		t.Fatalf("expected anthropic=demo, got %q", got)
	}
}
