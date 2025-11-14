package tool_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/royisme/bobamixer/internal/tool"
)

//nolint:gocyclo // Test function with multiple subtests is acceptable
func TestRun(t *testing.T) {
	t.Run("successful command execution", func(t *testing.T) {
		// Given: simple echo command
		ctx := context.Background()
		spec := tool.ExecSpec{
			Bin:  "echo",
			Args: []string{"hello", "world"},
		}

		// When: Run is called
		result, err := tool.Run(ctx, spec)

		// Then: command succeeds
		if err != nil {
			t.Fatalf("Run failed: %v", err)
		}
		if !result.Success {
			t.Error("expected success=true")
		}
		if result.ExitCode != 0 {
			t.Errorf("exit code = %d, want 0", result.ExitCode)
		}
		if result.LatencyMS <= 0 {
			t.Errorf("latency = %d, want > 0", result.LatencyMS)
		}
	})

	t.Run("command with non-zero exit code", func(t *testing.T) {
		// Given: command that exits with error
		ctx := context.Background()
		spec := tool.ExecSpec{
			Bin:  "sh",
			Args: []string{"-c", "exit 42"},
		}

		// When: Run is called
		result, err := tool.Run(ctx, spec)

		// Then: result indicates failure
		if err != nil {
			t.Fatalf("Run failed: %v", err)
		}
		if result.Success {
			t.Error("expected success=false for non-zero exit")
		}
		if result.ExitCode != 42 {
			t.Errorf("exit code = %d, want 42", result.ExitCode)
		}
		if result.ErrorClass != "tool_exit" {
			t.Errorf("error class = %s, want tool_exit", result.ErrorClass)
		}
	})

	t.Run("timeout handling", func(t *testing.T) {
		// Given: long-running command with short timeout
		ctx := context.Background()
		spec := tool.ExecSpec{
			Bin:     "sleep",
			Args:    []string{"10"},
			Timeout: 100 * time.Millisecond,
		}

		// When: Run is called
		result, err := tool.Run(ctx, spec)

		// Then: times out with appropriate error class
		if err != nil {
			t.Fatalf("Run failed: %v", err)
		}
		if result.Success {
			t.Error("expected success=false for timeout")
		}
		if result.ErrorClass != "timeout" {
			t.Errorf("error class = %s, want timeout", result.ErrorClass)
		}
	})

	t.Run("stdin input", func(t *testing.T) {
		// Given: command that reads stdin
		ctx := context.Background()
		spec := tool.ExecSpec{
			Bin:   "cat",
			Stdin: []byte("test input"),
		}

		// When: Run is called
		result, err := tool.Run(ctx, spec)

		// Then: succeeds and captures output size
		if err != nil {
			t.Fatalf("Run failed: %v", err)
		}
		if !result.Success {
			t.Error("expected success=true")
		}
		if result.StdoutSize == 0 {
			t.Error("expected stdout size > 0")
		}
	})

	t.Run("working directory", func(t *testing.T) {
		// Given: command with specific working directory
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.txt")
		if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil { //nolint:gosec // G306: test file permissions
			t.Fatalf("failed to create test file: %v", err)
		}

		ctx := context.Background()
		spec := tool.ExecSpec{
			Bin:        "ls",
			Args:       []string{"-1"},
			WorkingDir: tmpDir,
		}

		// When: Run is called
		result, err := tool.Run(ctx, spec)

		// Then: command runs in specified directory
		if err != nil {
			t.Fatalf("Run failed: %v", err)
		}
		if !result.Success {
			t.Error("expected success=true")
		}
		// Output should contain test.txt (captured in stdout)
		if result.StdoutSize == 0 {
			t.Error("expected stdout size > 0")
		}
	})

	t.Run("environment variables", func(t *testing.T) {
		// Given: command that uses environment variable
		ctx := context.Background()
		spec := tool.ExecSpec{
			Bin:  "sh",
			Args: []string{"-c", "echo $TEST_VAR"},
			Env:  []string{"TEST_VAR=test_value"},
		}

		// When: Run is called
		result, err := tool.Run(ctx, spec)

		// Then: succeeds with environment applied
		if err != nil {
			t.Fatalf("Run failed: %v", err)
		}
		if !result.Success {
			t.Error("expected success=true")
		}
	})

	t.Run("large output handling", func(t *testing.T) {
		// Given: command with large output
		ctx := context.Background()
		spec := tool.ExecSpec{
			Bin:  "sh",
			Args: []string{"-c", "seq 1 10000"},
		}

		// When: Run is called
		result, err := tool.Run(ctx, spec)

		// Then: handles large output without crash
		if err != nil {
			t.Fatalf("Run failed: %v", err)
		}
		if !result.Success {
			t.Error("expected success=true")
		}
		if result.StdoutSize == 0 {
			t.Error("expected large stdout size")
		}
	})

	t.Run("session ID auto-generation", func(t *testing.T) {
		// Given: spec without session ID
		ctx := context.Background()
		spec := tool.ExecSpec{
			Bin:  "echo",
			Args: []string{"test"},
		}

		// When: Run is called
		result, err := tool.Run(ctx, spec)

		// Then: session ID is auto-generated
		if err != nil {
			t.Fatalf("Run failed: %v", err)
		}
		// Session ID should be in result metadata (not visible in ExecResult struct yet)
		// This is a placeholder for future session tracking
		_ = result
	})
}

func TestJSONLUsageParsing(t *testing.T) {
	t.Run("parses JSONL usage from stdout", func(t *testing.T) {
		// Given: command that outputs JSONL usage
		ctx := context.Background()
		tmpDir := t.TempDir()
		scriptPath := filepath.Join(tmpDir, "test_cli.sh")
		script := `#!/bin/sh
echo 'Some output'
echo '{"event":"usage","input_tokens":100,"output_tokens":200,"latency_ms":50}'
`
		if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil { //nolint:gosec // G306: test script needs execute permission
			t.Fatalf("failed to create script: %v", err)
		}

		spec := tool.ExecSpec{
			Bin: "sh",
			Args: []string{scriptPath},
		}

		// When: Run is called
		result, err := tool.Run(ctx, spec)

		// Then: usage is parsed as exact
		if err != nil {
			t.Fatalf("Run failed: %v", err)
		}
		if !result.Success {
			t.Error("expected success=true")
		}
		if result.Usage.Estimate != "exact" {
			t.Errorf("usage estimate = %s, want exact", result.Usage.Estimate)
		}
		if result.Usage.InputTokens != 100 {
			t.Errorf("input tokens = %d, want 100", result.Usage.InputTokens)
		}
		if result.Usage.OutputTokens != 200 {
			t.Errorf("output tokens = %d, want 200", result.Usage.OutputTokens)
		}
	})

	t.Run("falls back to heuristic when no usage", func(t *testing.T) {
		// Given: command with no JSONL usage
		ctx := context.Background()
		spec := tool.ExecSpec{
			Bin:  "echo",
			Args: []string{"plain text"},
		}

		// When: Run is called
		result, err := tool.Run(ctx, spec)

		// Then: usage estimate is heuristic
		if err != nil {
			t.Fatalf("Run failed: %v", err)
		}
		if result.Usage.Estimate != "heuristic" {
			t.Errorf("usage estimate = %s, want heuristic", result.Usage.Estimate)
		}
	})
}

func TestErrorClassification(t *testing.T) {
	tests := []struct {
		name          string
		spec          tool.ExecSpec
		wantErrorClass string
	}{
		{
			name: "tool_exit for non-zero exit",
			spec: tool.ExecSpec{
				Bin:  "sh",
				Args: []string{"-c", "exit 1"},
			},
			wantErrorClass: "tool_exit",
		},
		{
			name: "timeout for command timeout",
			spec: tool.ExecSpec{
				Bin:     "sleep",
				Args:    []string{"5"},
				Timeout: 50 * time.Millisecond,
			},
			wantErrorClass: "timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := tool.Run(ctx, tt.spec)
			if err != nil {
				t.Fatalf("Run failed: %v", err)
			}
			if result.Success {
				t.Error("expected success=false")
			}
			if !strings.Contains(result.ErrorClass, tt.wantErrorClass) {
				t.Errorf("error class = %s, want %s", result.ErrorClass, tt.wantErrorClass)
			}
		})
	}
}
