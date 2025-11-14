// Package tool provides a unified interface for executing external CLI tools
// with usage tracking, session management, and error classification.
package tool

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"time"

	"github.com/royisme/bobamixer/internal/adapters"
	tooladapter "github.com/royisme/bobamixer/internal/adapters/tool"
)

// ExecSpec defines the specification for executing a CLI command.
type ExecSpec struct {
	SessionID  string        // Optional session ID; auto-generated if empty
	Bin        string        // Binary to execute (required)
	Args       []string      // Command arguments
	Env        []string      // Environment variables (KEY=VALUE format)
	Stdin      []byte        // Optional stdin input
	WorkingDir string        // Optional working directory
	Timeout    time.Duration // Optional timeout
	Apply      bool          // Must be true for commands with side effects (zero-coupling principle)
}

// Usage represents token usage information.
type Usage struct {
	InputTokens  int    // Number of input tokens
	OutputTokens int    // Number of output tokens
	LatencyMS    int64  // Execution latency in milliseconds
	Estimate     string // "exact" | "mapped" | "heuristic"
}

// ExecResult contains the result of command execution.
type ExecResult struct {
	Success     bool   // Whether command succeeded (exit code 0)
	ExitCode    int    // Process exit code
	StdoutSize  int64  // Size of stdout in bytes
	StderrSize  int64  // Size of stderr in bytes
	LatencyMS   int64  // Execution latency in milliseconds
	Usage       Usage  // Token usage (exact if available, heuristic otherwise)
	ErrorClass  string // Error classification: "tool_exit" | "timeout" | "io_error" | ""
}

// Run executes a CLI command according to the specification.
// It wraps the execution with session tracking, timeout handling, and usage parsing.
func Run(ctx context.Context, spec ExecSpec) (ExecResult, error) {
	// Validate required fields
	if spec.Bin == "" {
		return ExecResult{}, errors.New("bin is required")
	}

	// Apply timeout if specified
	execCtx := ctx
	if spec.Timeout > 0 {
		var cancel context.CancelFunc
		execCtx, cancel = context.WithTimeout(ctx, spec.Timeout)
		defer cancel()
	}

	// Build command
	// #nosec G204 -- bin and args are from trusted configuration, not direct user input
	cmd := exec.CommandContext(execCtx, spec.Bin, spec.Args...)

	// Set environment (merge with parent environment)
	if len(spec.Env) > 0 {
		cmd.Env = append(os.Environ(), spec.Env...)
	}

	// Set working directory
	if spec.WorkingDir != "" {
		cmd.Dir = spec.WorkingDir
	}

	// Set stdin
	if len(spec.Stdin) > 0 {
		cmd.Stdin = bytes.NewReader(spec.Stdin)
	}

	// Capture stdout and stderr separately
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute
	start := time.Now()
	err := cmd.Run()
	latencyMS := time.Since(start).Milliseconds()

	// Parse usage from output
	usage := parseUsage(stdout.Bytes(), stderr.Bytes())
	usage.LatencyMS = latencyMS

	// Determine exit code and success
	exitCode := 0
	success := err == nil
	errorClass := ""

	if err != nil {
		// Check for timeout
		if errors.Is(execCtx.Err(), context.DeadlineExceeded) {
			errorClass = "timeout"
			exitCode = -1
		} else if exitErr, ok := err.(*exec.ExitError); ok {
			// Non-zero exit code
			errorClass = "tool_exit"
			exitCode = exitErr.ExitCode()
		} else {
			// Other errors (IO, etc.)
			errorClass = "io_error"
			exitCode = -1
		}
		success = false
	}

	return ExecResult{
		Success:     success,
		ExitCode:    exitCode,
		StdoutSize:  int64(stdout.Len()),
		StderrSize:  int64(stderr.Len()),
		LatencyMS:   latencyMS,
		Usage:       usage,
		ErrorClass:  errorClass,
	}, nil
}

// parseUsage attempts to parse usage from tool output.
// It tries JSONL format first (exact), otherwise falls back to heuristic.
func parseUsage(stdout, stderr []byte) Usage {
	// Try parsing stdout
	if event, ok := parseJSONLinesUsage(stdout); ok {
		return Usage{
			InputTokens:  event.InputTokens,
			OutputTokens: event.OutputTokens,
			Estimate:     "exact",
		}
	}

	// Try parsing stderr
	if event, ok := parseJSONLinesUsage(stderr); ok {
		return Usage{
			InputTokens:  event.InputTokens,
			OutputTokens: event.OutputTokens,
			Estimate:     "exact",
		}
	}

	// No usage found, return heuristic placeholder
	return Usage{
		InputTokens:  0,
		OutputTokens: 0,
		Estimate:     "heuristic",
	}
}

// parseJSONLinesUsage parses JSONL usage events from output
func parseJSONLinesUsage(data []byte) (*tooladapter.UsageEvent, bool) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}

		var event tooladapter.UsageEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			continue
		}

		// Check if this is a usage event
		if event.Event == "usage" && (event.InputTokens > 0 || event.OutputTokens > 0) {
			return &event, true
		}
	}

	return nil, false
}

// ConvertAdapterUsage converts adapters.Usage to tool.Usage
func ConvertAdapterUsage(au adapters.Usage) Usage {
	estimate := "heuristic"
	switch au.Estimate {
	case adapters.EstimateExact:
		estimate = "exact"
	case adapters.EstimateMapped:
		estimate = "mapped"
	case adapters.EstimateHeuristic:
		estimate = "heuristic"
	}

	return Usage{
		InputTokens:  au.InputTokens,
		OutputTokens: au.OutputTokens,
		LatencyMS:    au.LatencyMS,
		Estimate:     estimate,
	}
}
