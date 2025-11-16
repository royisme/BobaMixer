// Package runner provides the execution pipeline for running CLI tools with injected configuration
package runner

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/royisme/bobamixer/internal/domain/core"
)

// RunContext contains all information needed to run a tool
type RunContext struct {
	Home     string
	Tool     *core.Tool
	Binding  *core.Binding
	Provider *core.Provider
	Secrets  *core.SecretsConfig
	Env      map[string]string // Environment variables to inject
	Args     []string          // Arguments to pass to the tool
}

// Runner is the interface for tool-specific runners
type Runner interface {
	// Prepare prepares the environment and configuration for running the tool
	Prepare(ctx *RunContext) error

	// Exec executes the tool with the prepared configuration
	Exec(ctx *RunContext) error
}

// Registry maintains a mapping of tool kinds to their runners
var registry = make(map[core.ToolKind]Runner)

// Register registers a runner for a specific tool kind
func Register(kind core.ToolKind, runner Runner) {
	registry[kind] = runner
}

// Get retrieves the runner for a specific tool kind
func Get(kind core.ToolKind) (Runner, error) {
	runner, ok := registry[kind]
	if !ok {
		return nil, fmt.Errorf("no runner registered for tool kind: %s", kind)
	}
	return runner, nil
}

// Run is a convenience function that prepares and executes a tool
func Run(ctx *RunContext) error {
	runner, err := Get(ctx.Tool.Kind)
	if err != nil {
		return err
	}

	// Prepare environment
	if err := runner.Prepare(ctx); err != nil {
		return fmt.Errorf("failed to prepare: %w", err)
	}

	// Execute
	if err := runner.Exec(ctx); err != nil {
		return fmt.Errorf("failed to execute: %w", err)
	}

	return nil
}

// BaseRunner provides common functionality for all runners
type BaseRunner struct{}

// Exec implements the default execution logic
func (b *BaseRunner) Exec(ctx *RunContext) error {
	//nolint:gosec // Executing configured CLI tools is the intended behavior
	cmd := exec.Command(ctx.Tool.Exec, ctx.Args...)

	// Merge environment variables
	cmd.Env = os.Environ()
	for key, value := range ctx.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	// Connect stdio
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command
	return cmd.Run()
}

// ResolveAPIKey is a helper to get the API key for a provider
func ResolveAPIKey(provider *core.Provider, secrets *core.SecretsConfig) (string, error) {
	return core.ResolveAPIKey(provider, secrets)
}
