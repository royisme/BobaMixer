package logging

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewLoggerWritesJSONL(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "boba.jsonl")
	log, err := New(Config{Path: path})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	child := log.With(String("session_id", "demo"))
	child.Info("session", String("api_key", "sk-secret"))

	if err := child.Sync(); err != nil {
		t.Fatalf("Sync() error = %v", err)
	}

	data, err := os.ReadFile(path) //nolint:gosec // test reads file from controlled temp dir
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "session") {
		t.Fatalf("expected log message to contain 'session', got %q", content)
	}
	if strings.Contains(content, "sk-secret") {
		t.Fatalf("expected secret to be redacted, got %q", content)
	}
	if !strings.Contains(content, "***REDACTED***") {
		t.Fatalf("expected redaction marker, got %q", content)
	}
}

func TestSanitizeRedactsSensitiveValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{name: "api key", input: "x-api-key: sk-demo", contains: "***REDACTED***"},
		{name: "bearer", input: "Authorization: Bearer token", contains: "***REDACTED***"},
		{name: "body", input: `"messages": "` + strings.Repeat("a", 60) + `"`, contains: "***TRUNCATED***"},
		{name: "safe", input: "ok", contains: "ok"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := Sanitize(tt.input)
			if !strings.Contains(got, tt.contains) {
				t.Fatalf("expected %q to contain %q", got, tt.contains)
			}
		})
	}
}
