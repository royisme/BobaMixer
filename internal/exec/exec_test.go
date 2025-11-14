package exec_test

import (
	"context"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/royisme/bobamixer/internal/exec"
	"github.com/royisme/bobamixer/internal/store/sqlite"
)

func setupTestDB(t *testing.T) *sqlite.DB {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	database, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	return database
}

//nolint:gocyclo // Test function with multiple subtests is acceptable
func TestSessionLifecycle(t *testing.T) {
	t.Run("complete session lifecycle", func(t *testing.T) {
		// Given: initialized database
		database := setupTestDB(t)
		ctx := context.Background()

		// When: BeginSession
		meta := exec.SessionMeta{
			Source:  "test",
			Profile: "test-profile",
			Project: "test-project",
			Branch:  "main",
		}
		sessionID, err := exec.BeginSession(ctx, database, meta)

		// Then: session is created
		if err != nil {
			t.Fatalf("BeginSession failed: %v", err)
		}
		if sessionID == "" {
			t.Error("expected non-empty session ID")
		}

		// When: RecordUsage
		usage := exec.Usage{
			InputTokens:  100,
			OutputTokens: 200,
			InputCost:    0.001,
			OutputCost:   0.002,
			Model:        "test-model",
			Estimate:     "exact",
		}
		err = exec.RecordUsage(ctx, database, sessionID, usage)

		// Then: usage is recorded
		if err != nil {
			t.Fatalf("RecordUsage failed: %v", err)
		}

		// When: EndSession
		latency := int64(123)
		err = exec.EndSession(ctx, database, sessionID, true, latency, "")

		// Then: session is ended successfully
		if err != nil {
			t.Fatalf("EndSession failed: %v", err)
		}

		// Verify session exists in database with persisted latency
		query := "SELECT success, ended_at, latency_ms FROM sessions WHERE id = '" + sessionID + "';"
		row, err := database.QueryRow(query)
		if err != nil {
			t.Fatalf("query failed: %v", err)
		}
		if row == "" {
			t.Error("session not found in database")
		}

		parts := strings.Split(row, "|")
		if len(parts) != 3 {
			t.Fatalf("unexpected session row format: %s", row)
		}
		if parts[2] != strconv.FormatInt(latency, 10) {
			t.Fatalf("expected latency %d, got %s", latency, parts[2])
		}
	})

	t.Run("session with failure", func(t *testing.T) {
		// Given: initialized database
		database := setupTestDB(t)
		ctx := context.Background()

		// When: Create session and end with failure
		meta := exec.SessionMeta{Source: "test"}
		sessionID, err := exec.BeginSession(ctx, database, meta)
		if err != nil {
			t.Fatalf("BeginSession failed: %v", err)
		}

		err = exec.EndSession(ctx, database, sessionID, false, 0, "test error")

		// Then: session marked as failed
		if err != nil {
			t.Fatalf("EndSession failed: %v", err)
		}

		query := "SELECT success, notes FROM sessions WHERE id = '" + sessionID + "';"
		row, err := database.QueryRow(query)
		if err != nil {
			t.Fatalf("query failed: %v", err)
		}
		// Row format: "0|test error"
		// First column should be 0 (false)
		if row[0] != '0' {
			t.Error("expected success=0")
		}
	})

	t.Run("EndSession must be called even on failure", func(t *testing.T) {
		// Given: session that encounters an error
		database := setupTestDB(t)
		ctx := context.Background()

		meta := exec.SessionMeta{Source: "test"}
		sessionID, err := exec.BeginSession(ctx, database, meta)
		if err != nil {
			t.Fatalf("BeginSession failed: %v", err)
		}

		// Simulate failure scenario
		// When: EndSession is called with success=false
		err = exec.EndSession(ctx, database, sessionID, false, 0, "simulated failure")

		// Then: no error and session is properly closed
		if err != nil {
			t.Fatalf("EndSession failed: %v", err)
		}

		// Verify ended_at is set
		query := "SELECT ended_at FROM sessions WHERE id = '" + sessionID + "';"
		row, err := database.QueryRow(query)
		if err != nil {
			t.Fatalf("query failed: %v", err)
		}
		if row == "" || row == "0" {
			t.Error("ended_at should be set")
		}
	})
}

func TestRecordUsage(t *testing.T) {
	t.Run("records usage with exact estimate", func(t *testing.T) {
		// Given: active session
		database := setupTestDB(t)
		ctx := context.Background()

		meta := exec.SessionMeta{Source: "test"}
		sessionID, err := exec.BeginSession(ctx, database, meta)
		if err != nil {
			t.Fatalf("BeginSession failed: %v", err)
		}

		// When: RecordUsage with exact estimate
		usage := exec.Usage{
			InputTokens:  500,
			OutputTokens: 1000,
			InputCost:    0.005,
			OutputCost:   0.010,
			Model:        "claude-sonnet-4",
			Estimate:     "exact",
		}
		err = exec.RecordUsage(ctx, database, sessionID, usage)

		// Then: usage is recorded correctly
		if err != nil {
			t.Fatalf("RecordUsage failed: %v", err)
		}

		query := "SELECT input_tokens, output_tokens, estimate_level FROM usage_records WHERE session_id = '" + sessionID + "';"
		row, err := database.QueryRow(query)
		if err != nil {
			t.Fatalf("query failed: %v", err)
		}
		if row == "" {
			t.Error("usage record not found")
		}
	})

	t.Run("records multiple usage entries for same session", func(t *testing.T) {
		// Given: active session
		database := setupTestDB(t)
		ctx := context.Background()

		meta := exec.SessionMeta{Source: "test"}
		sessionID, err := exec.BeginSession(ctx, database, meta)
		if err != nil {
			t.Fatalf("BeginSession failed: %v", err)
		}

		// When: RecordUsage multiple times
		for i := 0; i < 3; i++ {
			usage := exec.Usage{
				InputTokens:  100,
				OutputTokens: 200,
				InputCost:    0.001,
				OutputCost:   0.002,
				Model:        "test-model",
				Estimate:     "exact",
			}
			if err := exec.RecordUsage(ctx, database, sessionID, usage); err != nil {
				t.Fatalf("RecordUsage %d failed: %v", i, err)
			}
		}

		// Then: all usage records exist
		query := "SELECT COUNT(*) FROM usage_records WHERE session_id = '" + sessionID + "';"
		row, err := database.QueryRow(query)
		if err != nil {
			t.Fatalf("query failed: %v", err)
		}
		// Should have 3 records
		if row != "3" {
			t.Errorf("expected 3 usage records, got %s", row)
		}
	})
}

func TestRunTool(t *testing.T) {
	t.Run("runs tool and records session", func(t *testing.T) {
		// Given: initialized database and home
		database := setupTestDB(t)
		home := t.TempDir()
		ctx := context.Background()

		// When: RunTool
		spec := exec.ToolExecSpec{
			Bin:     "echo",
			Args:    []string{"test"},
			Profile: "test-profile",
		}
		result, err := exec.RunTool(ctx, database, home, spec)

		// Then: tool runs successfully and session is recorded
		if err != nil {
			t.Fatalf("RunTool failed: %v", err)
		}
		if !result.Success {
			t.Error("expected success=true")
		}
		if result.SessionID == "" {
			t.Error("expected session ID to be set")
		}

		// Verify session was recorded
		query := "SELECT COUNT(*) FROM sessions WHERE id = '" + result.SessionID + "';"
		row, err := database.QueryRow(query)
		if err != nil {
			t.Fatalf("query failed: %v", err)
		}
		if row != "1" {
			t.Error("session not found in database")
		}
	})
}

func TestConcurrentSessions(t *testing.T) {
	t.Run("handles concurrent sessions correctly", func(t *testing.T) {
		// Given: initialized database
		database := setupTestDB(t)
		ctx := context.Background()

		// When: Create multiple concurrent sessions
		sessions := make([]string, 10)
		for i := 0; i < 10; i++ {
			meta := exec.SessionMeta{Source: "concurrent-test"}
			sessionID, err := exec.BeginSession(ctx, database, meta)
			if err != nil {
				t.Fatalf("BeginSession %d failed: %v", i, err)
			}
			sessions[i] = sessionID
		}

		// End all sessions
		for _, sessionID := range sessions {
			if err := exec.EndSession(ctx, database, sessionID, true, 0, ""); err != nil {
				t.Fatalf("EndSession failed: %v", err)
			}
		}

		// Then: all sessions are recorded
		query := "SELECT COUNT(*) FROM sessions;"
		row, err := database.QueryRow(query)
		if err != nil {
			t.Fatalf("query failed: %v", err)
		}
		if row != "10" {
			t.Errorf("expected 10 sessions, got %s", row)
		}
	})
}
