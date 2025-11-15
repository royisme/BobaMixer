package logging

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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

func TestNewLoggerCreatesDefaultPathWithPermissions(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("permissions assertions unreliable on windows")
	}

	home := t.TempDir()
	log, err := New(Config{Home: home})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	log.Info("hello", String("api_key", "sk-secret"))
	if err := log.Sync(); err != nil {
		t.Fatalf("Sync() error = %v", err)
	}

	expected := filepath.Join(home, "logs", "boba-"+time.Now().Format("20060102")+".jsonl")
	info, err := os.Stat(expected)
	if err != nil {
		t.Fatalf("stat log file: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("expected permissions 0600, got %o", info.Mode().Perm())
	}
}

func TestNewLoggerFailsWhenDirectoryInvalid(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	blocked := filepath.Join(dir, "blocked")
	if err := os.WriteFile(blocked, []byte("x"), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}
	path := filepath.Join(blocked, "boba.jsonl")
	if _, err := New(Config{Path: path}); err == nil {
		t.Fatal("expected New to fail when parent path is a file")
	}
}

func TestSanitizeFieldsRedactsComplexTypes(t *testing.T) {
	t.Parallel()

	fields := []Field{
		zap.Any("payload", map[string]string{"prompt": "Authorization: Bearer tok"}),
		{Key: "safe", Type: zapcore.StringType, String: "ok"},
	}
	sanitized := sanitizeFields(fields)
	if sanitized[0].Type != zapcore.StringType {
		// zap.Any is converted into sanitized string field
		t.Fatalf("expected first field to become string, got %v", sanitized[0].Type)
	}
	if !strings.Contains(sanitized[0].String, "***REDACTED***") {
		t.Fatalf("expected redaction marker, got %q", sanitized[0].String)
	}
	if sanitized[1].String != "ok" {
		t.Fatalf("unexpected change to safe field: %q", sanitized[1].String)
	}
}
