// Package tooladapter provides an adapter for executing external command-line tools.
package tooladapter

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/royisme/bobamixer/internal/adapters"
)

// Runner is an adapter that executes external command-line tools.
type Runner struct {
	name string
	bin  string
	args []string
	env  []string
}

// UsageEvent represents a JSON Lines usage event from tool output
type UsageEvent struct {
	Event        string `json:"event"`
	InputTokens  int    `json:"input_tokens"`
	OutputTokens int    `json:"output_tokens"`
	LatencyMS    int64  `json:"latency_ms"`
}

// New creates a new tool runner with the given name, binary path, and environment variables.
func New(name, bin string, env []string) *Runner {
	return &Runner{name: name, bin: bin, env: env, args: []string{}}
}

// NewWithArgs creates a new tool runner with additional command-line arguments.
func NewWithArgs(name, bin string, args []string, env []string) *Runner {
	return &Runner{name: name, bin: bin, args: args, env: env}
}

// Name returns the name of this tool runner.
func (r *Runner) Name() string { return r.name }

// Execute runs the external tool and returns the result.
func (r *Runner) Execute(ctx context.Context, req adapters.Request) (adapters.Result, error) {
	// Build command with args
	// #nosec G204 -- bin and args are from tool configuration, not direct user input
	cmd := exec.CommandContext(ctx, r.bin, r.args...)

	// Set environment
	cmd.Env = append(os.Environ(), r.env...)

	// Set stdin if payload provided
	if len(req.Payload) > 0 {
		cmd.Stdin = strings.NewReader(string(req.Payload))
	}

	// Capture stdout and stderr separately
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	err := cmd.Run()
	latencyMS := time.Since(start).Milliseconds()

	// Try to parse usage from output
	usage := r.parseUsage(stdout.Bytes(), stderr.Bytes())
	usage.LatencyMS = latencyMS

	// Combine output
	output := append(stdout.Bytes(), stderr.Bytes()...)

	result := adapters.Result{
		Success: err == nil,
		Output:  output,
		Usage:   usage,
	}

	if err != nil {
		result.Error = err.Error()
	}

	return result, nil
}

// parseUsage attempts to parse usage from JSON Lines events in output
func (r *Runner) parseUsage(stdout, stderr []byte) adapters.Usage {
	usage := adapters.Usage{
		InputTokens:  0,
		OutputTokens: 0,
		Estimate:     adapters.EstimateHeuristic,
	}

	// Try parsing stdout first
	if parsed, ok := r.tryParseJSONLines(stdout); ok {
		usage.InputTokens = parsed.InputTokens
		usage.OutputTokens = parsed.OutputTokens
		usage.Estimate = adapters.EstimateExact
		return usage
	}

	// Try parsing stderr
	if parsed, ok := r.tryParseJSONLines(stderr); ok {
		usage.InputTokens = parsed.InputTokens
		usage.OutputTokens = parsed.OutputTokens
		usage.Estimate = adapters.EstimateExact
		return usage
	}

	// If no usage found, will use tokenizer estimation (EstimateHeuristic)
	return usage
}

// tryParseJSONLines attempts to parse JSON Lines format usage events
func (r *Runner) tryParseJSONLines(data []byte) (*UsageEvent, bool) {
	scanner := bufio.NewScanner(bytes.NewReader(data))

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}

		var event UsageEvent
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

// ExecuteStreaming executes the tool with streaming output
func (r *Runner) ExecuteStreaming(ctx context.Context, req adapters.Request, outWriter, errWriter io.Writer) (adapters.Result, error) {
	// #nosec G204 -- bin and args are from tool configuration, not direct user input
	cmd := exec.CommandContext(ctx, r.bin, r.args...)
	cmd.Env = append(os.Environ(), r.env...)

	if len(req.Payload) > 0 {
		cmd.Stdin = strings.NewReader(string(req.Payload))
	}

	// Capture while streaming
	var stdout, stderr bytes.Buffer
	cmd.Stdout = io.MultiWriter(&stdout, outWriter)
	cmd.Stderr = io.MultiWriter(&stderr, errWriter)

	start := time.Now()
	err := cmd.Run()
	latencyMS := time.Since(start).Milliseconds()

	usage := r.parseUsage(stdout.Bytes(), stderr.Bytes())
	usage.LatencyMS = latencyMS

	result := adapters.Result{
		Success: err == nil,
		Output:  append(stdout.Bytes(), stderr.Bytes()...),
		Usage:   usage,
	}

	if err != nil {
		result.Error = err.Error()
	}

	return result, nil
}
