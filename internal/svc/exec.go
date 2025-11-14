// Package svc provides service layer for executing AI calls and managing sessions.
package svc

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/royisme/bobamixer/internal/adapters"
	httpadapter "github.com/royisme/bobamixer/internal/adapters/http"
	tooladapter "github.com/royisme/bobamixer/internal/adapters/tool"
	"github.com/royisme/bobamixer/internal/logging"
	"github.com/royisme/bobamixer/internal/store/config"
	"github.com/royisme/bobamixer/internal/store/sqlite"
)

// Executor handles the execution of AI calls with session and usage tracking
type Executor struct {
	db      *sqlite.DB
	home    string
	secrets config.Secrets
}

// NewExecutor creates a new executor
func NewExecutor(db *sqlite.DB, home string) (*Executor, error) {
	// Validate secrets permissions before loading
	if err := config.ValidateSecretsPermissions(home); err != nil {
		return nil, err
	}

	// Load secrets
	secrets, err := config.LoadSecrets(home)
	if err != nil {
		return nil, fmt.Errorf("load secrets: %w", err)
	}

	return &Executor{
		db:      db,
		home:    home,
		secrets: secrets,
	}, nil
}

// ExecuteRequest represents a request to execute an AI call
type ExecuteRequest struct {
	ProfileKey string
	Payload    []byte
	Project    string
	Branch     string
	TaskType   string
}

// ExecuteResult represents the result of an AI call execution
type ExecuteResult struct {
	SessionID string
	Success   bool
	Output    []byte
	Error     string
	Usage     adapters.Usage
}

// Execute runs an AI call with full session and usage tracking
func (e *Executor) Execute(ctx context.Context, req ExecuteRequest) (*ExecuteResult, error) {
	logging.Info("Executing AI call",
		logging.String("profile", req.ProfileKey),
		logging.String("project", req.Project),
		logging.String("branch", req.Branch),
		logging.String("task_type", req.TaskType))

	// Load profile
	profiles, err := config.LoadProfiles(e.home)
	if err != nil {
		logging.Error("Failed to load profiles", logging.Err(err))
		return nil, fmt.Errorf("load profiles: %w", err)
	}

	profile, ok := profiles[req.ProfileKey]
	if !ok {
		logging.Error("Profile not found", logging.String("profile", req.ProfileKey))
		return nil, fmt.Errorf("profile %s not found", req.ProfileKey)
	}

	// Create adapter
	adapter, err := e.createAdapter(profile)
	if err != nil {
		logging.Error("Failed to create adapter",
			logging.String("adapter", profile.Adapter),
			logging.Err(err))
		return nil, fmt.Errorf("create adapter: %w", err)
	}

	// Begin session
	sessionID := uuid.New().String()
	startTime := time.Now()
	logging.Info("Session started",
		logging.String("session_id", sessionID),
		logging.String("profile", req.ProfileKey),
		logging.String("adapter", profile.Adapter))

	if err := e.beginSession(sessionID, req, profile); err != nil {
		logging.Error("Failed to begin session", logging.Err(err))
		return nil, fmt.Errorf("begin session: %w", err)
	}

	// Execute adapter
	adapterReq := adapters.Request{
		Payload: req.Payload,
	}
	result, err := adapter.Execute(ctx, adapterReq)
	if err != nil {
		logging.Error("Adapter execution failed",
			logging.String("session_id", sessionID),
			logging.Err(err))
		// End session with error
		if endErr := e.endSession(sessionID, false, time.Since(startTime).Milliseconds(), err.Error()); endErr != nil {
			logging.Error("Failed to end session after adapter error", logging.Err(endErr))
		}
		return nil, fmt.Errorf("adapter execute: %w", err)
	}

	// Persist usage
	if err := e.persistUsage(sessionID, profile, result.Usage); err != nil {
		logging.Error("Failed to persist usage",
			logging.String("session_id", sessionID),
			logging.Err(err))
		return nil, fmt.Errorf("persist usage: %w", err)
	}

	// End session
	latency := time.Since(startTime).Milliseconds()
	if err := e.endSession(sessionID, result.Success, latency, result.Error); err != nil {
		logging.Error("Failed to end session",
			logging.String("session_id", sessionID),
			logging.Err(err))
		return nil, fmt.Errorf("end session: %w", err)
	}

	logging.Info("Session completed",
		logging.String("session_id", sessionID),
		logging.Bool("success", result.Success),
		logging.Int64("latency_ms", latency),
		logging.Int("input_tokens", result.Usage.InputTokens),
		logging.Int("output_tokens", result.Usage.OutputTokens),
		logging.String("estimate", string(result.Usage.Estimate)))

	return &ExecuteResult{
		SessionID: sessionID,
		Success:   result.Success,
		Output:    result.Output,
		Error:     result.Error,
		Usage:     result.Usage,
	}, nil
}

func (e *Executor) createAdapter(profile config.Profile) (adapters.Adapter, error) {
	switch profile.Adapter {
	case "http":
		// Resolve environment variables with secrets
		env := config.ResolveEnv(profile.Env, e.secrets)

		// Build headers from environment
		headers := make(map[string]string)
		headers["Content-Type"] = "application/json"

		// Extract API key from env
		for _, envVar := range env {
			if len(envVar) > 0 {
				// Parse KEY=VALUE
				parts := splitEnv(envVar)
				if len(parts) == 2 {
					key, value := parts[0], parts[1]
					// Map common env var names to headers
					switch key {
					case "ANTHROPIC_API_KEY":
						headers["x-api-key"] = value
						headers["anthropic-version"] = "2023-06-01"
					case "OPENAI_API_KEY":
						headers["Authorization"] = "Bearer " + value
					case "OPENROUTER_API_KEY":
						headers["Authorization"] = "Bearer " + value
					}
				}
			}
		}

		return httpadapter.NewWithProvider(
			profile.Key,
			profile.Provider,
			profile.Endpoint,
			headers,
		), nil

	case "tool":
		env := config.ResolveEnv(profile.Env, e.secrets)
		return tooladapter.New(profile.Key, profile.Model, env), nil

	default:
		return nil, fmt.Errorf("unsupported adapter type: %s", profile.Adapter)
	}
}

func (e *Executor) beginSession(sessionID string, req ExecuteRequest, profile config.Profile) error {
	query := fmt.Sprintf(
		`INSERT INTO sessions (id, started_at, project, branch, profile, adapter, task_type)
		 VALUES ('%s', %d, '%s', '%s', '%s', '%s', '%s');`,
		sessionID,
		time.Now().Unix(),
		sqlEscape(req.Project),
		sqlEscape(req.Branch),
		sqlEscape(profile.Key),
		sqlEscape(profile.Adapter),
		sqlEscape(req.TaskType),
	)
	return e.db.Exec(query)
}

func (e *Executor) endSession(sessionID string, success bool, latencyMS int64, notes string) error {
	successInt := 0
	if success {
		successInt = 1
	}
	query := fmt.Sprintf(
		`UPDATE sessions SET ended_at = %d, success = %d, latency_ms = %d, notes = '%s' WHERE id = '%s';`,
		time.Now().Unix(),
		successInt,
		latencyMS,
		sqlEscape(notes),
		sessionID,
	)
	return e.db.Exec(query)
}

func (e *Executor) persistUsage(sessionID string, profile config.Profile, usage adapters.Usage) error {
	usageID := uuid.New().String()

	// Calculate costs
	inputCost := float64(usage.InputTokens) * profile.CostPer1K.Input / 1000.0
	outputCost := float64(usage.OutputTokens) * profile.CostPer1K.Output / 1000.0

	// Map estimate level to string
	estimateLevel := "heuristic"
	switch usage.Estimate {
	case adapters.EstimateExact:
		estimateLevel = "exact"
	case adapters.EstimateMapped:
		estimateLevel = "mapped"
	case adapters.EstimateHeuristic:
		estimateLevel = "heuristic"
	}

	query := fmt.Sprintf(
		`INSERT INTO usage_records
		 (id, session_id, ts, input_tokens, output_tokens, input_cost, output_cost, tool, model, estimate_level)
		 VALUES ('%s', '%s', %d, %d, %d, %f, %f, '%s', '%s', '%s');`,
		usageID,
		sessionID,
		time.Now().Unix(),
		usage.InputTokens,
		usage.OutputTokens,
		inputCost,
		outputCost,
		sqlEscape(profile.Adapter),
		sqlEscape(profile.Model),
		estimateLevel,
	)
	return e.db.Exec(query)
}

// splitEnv splits "KEY=VALUE" into ["KEY", "VALUE"]
func splitEnv(s string) []string {
	idx := -1
	for i, c := range s {
		if c == '=' {
			idx = i
			break
		}
	}
	if idx == -1 {
		return []string{s}
	}
	return []string{s[:idx], s[idx+1:]}
}

// sqlEscape escapes single quotes for SQLite
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
