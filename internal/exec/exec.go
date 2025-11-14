// Package exec provides execution services for managing sessions and usage tracking.
// It exposes public APIs for session lifecycle management and tool/HTTP execution.
package exec

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/royisme/bobamixer/internal/httpx"
	"github.com/royisme/bobamixer/internal/store/config"
	"github.com/royisme/bobamixer/internal/store/sqlite"
	"github.com/royisme/bobamixer/internal/tool"
)

// SessionMeta contains metadata for starting a session.
type SessionMeta struct {
	Source  string // "wrap" | "call" | "http" | "tool" | etc.
	Profile string // profile key
	Project string // project name
	Branch  string // git branch
}

// Usage represents token usage and cost information.
type Usage struct {
	InputTokens  int     // number of input tokens
	OutputTokens int     // number of output tokens
	InputCost    float64 // cost for input tokens
	OutputCost   float64 // cost for output tokens
	Model        string  // model name
	Estimate     string  // "exact" | "mapped" | "heuristic"
}

// ToolExecSpec specifies how to execute a tool command.
type ToolExecSpec struct {
	Bin        string   // binary to execute
	Args       []string // arguments
	Env        []string // environment variables
	Stdin      []byte   // stdin input
	WorkingDir string   // working directory
	Timeout    time.Duration
	Profile    string // profile for usage tracking
}

// ToolExecResult contains the result of tool execution.
type ToolExecResult struct {
	SessionID  string
	Success    bool
	ExitCode   int
	StdoutSize int64
	StderrSize int64
	LatencyMS  int64
	ErrorClass string
}

// BeginSession creates a new session in the database and returns the session ID.
// This must be called before any usage recording.
func BeginSession(ctx context.Context, db *sqlite.DB, meta SessionMeta) (string, error) {
	sessionID := uuid.New().String()
	now := time.Now().Unix()

	query := fmt.Sprintf(
		`INSERT INTO sessions (id, started_at, project, branch, profile, adapter, task_type)
		 VALUES ('%s', %d, '%s', '%s', '%s', '%s', '%s');`,
		sessionID,
		now,
		sqlEscape(meta.Project),
		sqlEscape(meta.Branch),
		sqlEscape(meta.Profile),
		sqlEscape(meta.Source), // use source as adapter for now
		"",                     // task_type can be empty
	)

	if err := db.Exec(query); err != nil {
		return "", fmt.Errorf("insert session: %w", err)
	}

	return sessionID, nil
}

// RecordUsage records a usage entry for a given session.
// Can be called multiple times per session.
func RecordUsage(ctx context.Context, db *sqlite.DB, sessionID string, usage Usage) error {
	usageID := uuid.New().String()
	now := time.Now().Unix()

	query := fmt.Sprintf(
		`INSERT INTO usage_records
		 (id, session_id, ts, input_tokens, output_tokens, input_cost, output_cost, tool, model, estimate_level)
		 VALUES ('%s', '%s', %d, %d, %d, %f, %f, '%s', '%s', '%s');`,
		usageID,
		sessionID,
		now,
		usage.InputTokens,
		usage.OutputTokens,
		usage.InputCost,
		usage.OutputCost,
		"", // tool can be empty for now
		sqlEscape(usage.Model),
		sqlEscape(usage.Estimate),
	)

	if err := db.Exec(query); err != nil {
		return fmt.Errorf("insert usage record: %w", err)
	}

	return nil
}

// EndSession marks a session as complete with success status, latency and optional notes.
// This must be called exactly once per session, even if the session failed.
func EndSession(ctx context.Context, db *sqlite.DB, sessionID string, success bool, latencyMS int64, notes string) error {
	successInt := 0
	if success {
		successInt = 1
	}
	now := time.Now().Unix()

	query := fmt.Sprintf(
		`UPDATE sessions SET ended_at = %d, success = %d, latency_ms = %d, notes = '%s' WHERE id = '%s';`,
		now,
		successInt,
		latencyMS,
		sqlEscape(notes),
		sessionID,
	)

	if err := db.Exec(query); err != nil {
		return fmt.Errorf("update session: %w", err)
	}

	return nil
}

// RunTool executes a tool command with full session and usage tracking.
// It automatically creates a session, executes the tool, records usage, and ends the session.
func RunTool(ctx context.Context, db *sqlite.DB, home string, spec ToolExecSpec) (*ToolExecResult, error) {
	// Begin session
	meta := SessionMeta{
		Source:  "tool",
		Profile: spec.Profile,
	}
	sessionID, err := BeginSession(ctx, db, meta)
	if err != nil {
		return nil, fmt.Errorf("begin session: %w", err)
	}

	// Execute tool
	toolSpec := tool.ExecSpec{
		SessionID:  sessionID,
		Bin:        spec.Bin,
		Args:       spec.Args,
		Env:        spec.Env,
		Stdin:      spec.Stdin,
		WorkingDir: spec.WorkingDir,
		Timeout:    spec.Timeout,
	}

	toolResult, err := tool.Run(ctx, toolSpec)
	if err != nil {
		// End session with error
		if endErr := EndSession(ctx, db, sessionID, false, 0, err.Error()); endErr != nil {
			// Log but don't override original error
			fmt.Printf("Warning: failed to end session: %v\n", endErr)
		}
		return nil, fmt.Errorf("tool execution: %w", err)
	}

	// Record usage if available
	if toolResult.Usage.InputTokens > 0 || toolResult.Usage.OutputTokens > 0 {
		usage := Usage{
			InputTokens:  toolResult.Usage.InputTokens,
			OutputTokens: toolResult.Usage.OutputTokens,
			Model:        "unknown", // tool execution doesn't specify model
			Estimate:     toolResult.Usage.Estimate,
		}
		if err := RecordUsage(ctx, db, sessionID, usage); err != nil {
			// Log error but don't fail the execution
			fmt.Printf("Warning: failed to record usage: %v\n", err)
		}
	}

	// End session
	success := toolResult.Success
	notes := ""
	if !success {
		notes = fmt.Sprintf("exit code %d: %s", toolResult.ExitCode, toolResult.ErrorClass)
	}
	if err := EndSession(ctx, db, sessionID, success, toolResult.LatencyMS, notes); err != nil {
		return nil, fmt.Errorf("end session: %w", err)
	}

	return &ToolExecResult{
		SessionID:  sessionID,
		Success:    toolResult.Success,
		ExitCode:   toolResult.ExitCode,
		StdoutSize: toolResult.StdoutSize,
		StderrSize: toolResult.StderrSize,
		LatencyMS:  toolResult.LatencyMS,
		ErrorClass: toolResult.ErrorClass,
	}, nil
}

// RunHTTP executes an HTTP request with full session and usage tracking.
// It automatically creates a session, executes the HTTP request, records usage, and ends the session.
func RunHTTP(ctx context.Context, db *sqlite.DB, home string, req HTTPRequest) (*HTTPResult, error) {
	// Begin session
	meta := SessionMeta{
		Source:  "http",
		Profile: req.Profile,
	}
	sessionID := req.SessionID
	if sessionID == "" {
		var err error
		sessionID, err = BeginSession(ctx, db, meta)
		if err != nil {
			return nil, fmt.Errorf("begin session: %w", err)
		}
	}

	// Execute HTTP request
	httpxReq := httpx.HTTPRequest{
		SessionID: sessionID,
		Endpoint:  req.Endpoint,
		Headers:   req.Headers,
		Payload:   req.Payload,
		Timeout:   req.Timeout,
		Retries:   req.Retries,
	}

	httpxResult, err := httpx.Execute(ctx, httpxReq)
	if err != nil {
		// End session with error
		if endErr := EndSession(ctx, db, sessionID, false, 0, err.Error()); endErr != nil {
			// Log but don't override original error
			fmt.Printf("Warning: failed to end session: %v\n", endErr)
		}
		return nil, fmt.Errorf("http execution: %w", err)
	}

	// Record usage if available
	if httpxResult.Usage.InputTokens > 0 || httpxResult.Usage.OutputTokens > 0 {
		usage := Usage{
			InputTokens:  httpxResult.Usage.InputTokens,
			OutputTokens: httpxResult.Usage.OutputTokens,
			Model:        req.Profile, // use profile as model for now
			Estimate:     httpxResult.Usage.Estimate,
		}
		if err := RecordUsage(ctx, db, sessionID, usage); err != nil {
			// Log error but don't fail the execution
			fmt.Printf("Warning: failed to record usage: %v\n", err)
		}
	}

	// End session
	success := httpxResult.Success
	notes := ""
	if !success {
		notes = fmt.Sprintf("status %d: %s", httpxResult.StatusCode, httpxResult.ErrorClass)
	}
	if err := EndSession(ctx, db, sessionID, success, httpxResult.Usage.LatencyMS, notes); err != nil {
		return nil, fmt.Errorf("end session: %w", err)
	}

	return &HTTPResult{
		SessionID:  sessionID,
		Success:    httpxResult.Success,
		StatusCode: httpxResult.StatusCode,
		Body:       httpxResult.Body,
		LatencyMS:  httpxResult.Usage.LatencyMS,
		ErrorClass: httpxResult.ErrorClass,
	}, nil
}

// HTTPRequest represents an HTTP request specification.
type HTTPRequest struct {
	SessionID string
	Endpoint  string
	Headers   map[string]string
	Payload   []byte
	Timeout   time.Duration
	Retries   int
	Profile   string
}

// HTTPResult represents the result of an HTTP request.
type HTTPResult struct {
	SessionID  string
	Success    bool
	StatusCode int
	Body       []byte
	LatencyMS  int64
	ErrorClass string
}

// sqlEscape escapes single quotes for SQLite.
func sqlEscape(s string) string {
	result := ""
	for _, c := range s {
		if c == '\'' {
			result += "''"
		} else {
			result += string(c)
		}
	}
	return result
}

// LoadSecrets loads secrets from the home directory for environment resolution.
func LoadSecrets(home string) (config.Secrets, error) {
	if err := config.ValidateSecretsPermissions(home); err != nil {
		return nil, fmt.Errorf("validate secrets permissions: %w", err)
	}

	secrets, err := config.LoadSecrets(home)
	if err != nil {
		return nil, fmt.Errorf("load secrets: %w", err)
	}

	return secrets, nil
}
