// Package mcp provides an adapter for Model Context Protocol (MCP) integrations.
package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"time"

	"github.com/royisme/bobamixer/internal/adapters"
)

// Transport abstracts how MCP messages are delivered.
type Transport interface {
	Call(ctx context.Context, payload []byte) ([]byte, error)
}

// Adapter routes requests through an MCP transport.
type Adapter struct {
	transport   Transport
	defaultTool string
}

// New creates an MCP adapter with a transport and optional default tool name.
func New(transport Transport, defaultTool string) *Adapter {
	return &Adapter{transport: transport, defaultTool: defaultTool}
}

// Execute sends the payload to the MCP transport and parses the response.
func (a *Adapter) Execute(ctx context.Context, req adapters.Request) (adapters.Result, error) {
	if a.transport == nil {
		return adapters.Result{}, errors.New("mcp transport not configured")
	}
	payload, err := a.buildPayload(req)
	if err != nil {
		return adapters.Result{}, err
	}
	start := time.Now()
	raw, err := a.transport.Call(ctx, payload)
	latency := time.Since(start)
	if err != nil {
		return adapters.Result{}, err
	}
	var resp responseEnvelope
	if err := json.Unmarshal(raw, &resp); err != nil {
		return adapters.Result{}, fmt.Errorf("decode MCP response: %w", err)
	}
	if resp.Error != "" {
		return adapters.Result{}, errors.New(resp.Error)
	}
	return adapters.Result{
		Success: true,
		Output:  []byte(resp.Output),
		Usage: adapters.Usage{
			InputTokens:  resp.InputTokens,
			OutputTokens: resp.OutputTokens,
			Estimate:     adapters.EstimateExact,
			LatencyMS:    latency.Milliseconds(),
		},
	}, nil
}

func (a *Adapter) Name() string {
	return "mcp"
}

func (a *Adapter) buildPayload(req adapters.Request) ([]byte, error) {
	if len(req.Payload) > 0 {
		return req.Payload, nil
	}
	tool := req.Tool
	if tool == "" {
		tool = a.defaultTool
	}
	msg := map[string]interface{}{
		"tool":     tool,
		"profile":  req.Profile,
		"session":  req.SessionID,
		"metadata": req.Metadata,
		"model":    req.Model,
	}
	return json.Marshal(msg)
}

// StdIOTransport launches an MCP server command per request.
type StdIOTransport struct {
	Command string
	Args    []string
}

// Call executes the command and streams payload via stdin/stdout.
func (t *StdIOTransport) Call(ctx context.Context, payload []byte) ([]byte, error) {
	if t == nil || t.Command == "" {
		return nil, errors.New("command not configured")
	}
	// #nosec G204 -- Command and Args are from MCP configuration, not direct user input
	cmd := exec.CommandContext(ctx, t.Command, t.Args...)
	cmd.Stdin = bytes.NewReader(payload)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("mcp command: %w", err)
	}
	return out.Bytes(), nil
}

type responseEnvelope struct {
	Output       string `json:"output"`
	InputTokens  int    `json:"input_tokens"`
	OutputTokens int    `json:"output_tokens"`
	Error        string `json:"error"`
}
