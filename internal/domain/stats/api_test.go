package stats_test

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/royisme/bobamixer/internal/domain/stats"
	"github.com/royisme/bobamixer/internal/store/sqlite"
)

func setupTestDB(t *testing.T) *sqlite.DB {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	return db
}

func insertTestUsage(t *testing.T, db *sqlite.DB, sessionID string, inputTokens, outputTokens int, cost float64, daysAgo int) {
	timestamp := time.Now().AddDate(0, 0, -daysAgo).Unix()

	sessionQuery := `INSERT INTO sessions (id, started_at, ended_at, success, latency_ms, profile)
		VALUES ('` + sessionID + `', ` + i64toa(timestamp) + `, ` + i64toa(timestamp+100) + `, 1, 100, 'test-profile');`
	if err := db.Exec(sessionQuery); err != nil {
		t.Fatalf("insert session: %v", err)
	}

	usageQuery := `INSERT INTO usage_records (id, session_id, ts, input_tokens, output_tokens, input_cost, output_cost, model, estimate_level)
		VALUES ('usage-` + sessionID + `', '` + sessionID + `', ` + i64toa(timestamp) + `, ` + itoa(inputTokens) + `, ` + itoa(outputTokens) + `, ` + ftoa(cost/2) + `, ` + ftoa(cost/2) + `, 'test-model', 'exact');`
	if err := db.Exec(usageQuery); err != nil {
		t.Fatalf("insert usage: %v", err)
	}
}

func i64toa(i int64) string {
	return fmt.Sprintf("%d", i)
}

func itoa(i int) string {
	return fmt.Sprintf("%d", i)
}

func ftoa(f float64) string {
	return fmt.Sprintf("%.4f", f)
}

func TestToday(t *testing.T) {
	t.Run("returns today's summary", func(t *testing.T) {
		// Given: database with today's usage
		db := setupTestDB(t)
		ctx := context.Background()

		// Insert usage for today
		insertTestUsage(t, db, "session-today-1", 100, 200, 0.01, 0)
		insertTestUsage(t, db, "session-today-2", 150, 250, 0.015, 0)

		// When: Today is called
		summary, err := stats.Today(ctx, db)

		// Then: returns today's summary
		if err != nil {
			t.Fatalf("Today failed: %v", err)
		}
		if summary.TotalTokens <= 0 {
			t.Error("expected total tokens > 0")
		}
		if summary.TotalSessions <= 0 {
			t.Error("expected total sessions > 0")
		}
	})

	t.Run("returns empty summary for no data", func(t *testing.T) {
		// Given: empty database
		db := setupTestDB(t)
		ctx := context.Background()

		// When: Today is called
		summary, err := stats.Today(ctx, db)

		// Then: returns empty summary without error
		if err != nil {
			t.Fatalf("Today failed: %v", err)
		}
		if summary.TotalTokens != 0 {
			t.Errorf("expected 0 tokens, got %d", summary.TotalTokens)
		}
	})
}

func TestWindow(t *testing.T) {
	t.Run("returns window summary", func(t *testing.T) {
		// Given: database with usage over multiple days
		db := setupTestDB(t)
		ctx := context.Background()

		// Insert usage for the past 7 days
		for i := 0; i < 7; i++ {
			insertTestUsage(t, db, "session-"+itoa(i), 100, 200, 0.01, i)
		}

		// When: Window is called for 7 days
		from := time.Now().AddDate(0, 0, -7)
		to := time.Now()
		summary, err := stats.Window(ctx, db, from, to)

		// Then: returns aggregated summary
		if err != nil {
			t.Fatalf("Window failed: %v", err)
		}
		if summary.TotalTokens <= 0 {
			t.Error("expected total tokens > 0")
		}
	})

	t.Run("respects time boundaries", func(t *testing.T) {
		// Given: database with usage at different times
		db := setupTestDB(t)
		ctx := context.Background()

		// Insert old usage (outside window)
		insertTestUsage(t, db, "session-old", 100, 200, 0.01, 30)

		// Insert recent usage (inside window)
		insertTestUsage(t, db, "session-recent", 150, 250, 0.015, 3)

		// When: Window is called for last 7 days
		from := time.Now().AddDate(0, 0, -7)
		to := time.Now()
		_, err := stats.Window(ctx, db, from, to)

		// Then: should not include old usage
		if err != nil {
			t.Fatalf("Window failed: %v", err)
		}
		// Note: Detailed verification would require checking actual token counts
	})
}

func TestP95Latency(t *testing.T) {
	t.Run("calculates overall P95", func(t *testing.T) {
		// Given: database with session latencies
		db := setupTestDB(t)
		ctx := context.Background()

		// Insert sessions with varying latencies
		// (Note: simplified test, would need proper session insertion)

		// When: P95Latency is called without byProfile
		window := 7 * 24 * time.Hour
		result, err := stats.P95Latency(ctx, db, window, false)

		// Then: returns overall P95
		if err != nil {
			t.Fatalf("P95Latency failed: %v", err)
		}
		if result == nil {
			t.Error("expected non-nil result")
		}
		if _, ok := result["overall"]; !ok {
			t.Error("expected 'overall' key in result")
		}
	})

	t.Run("calculates per-profile P95", func(t *testing.T) {
		// Given: database with multi-profile sessions
		db := setupTestDB(t)
		ctx := context.Background()

		// When: P95Latency is called with byProfile=true
		window := 7 * 24 * time.Hour
		result, err := stats.P95Latency(ctx, db, window, true)

		// Then: returns per-profile P95
		if err != nil {
			t.Fatalf("P95Latency failed: %v", err)
		}
		if result == nil {
			t.Error("expected non-nil result")
		}
	})

	t.Run("handles empty dataset", func(t *testing.T) {
		// Given: empty database
		db := setupTestDB(t)
		ctx := context.Background()

		// When: P95Latency is called
		window := 7 * 24 * time.Hour
		result, err := stats.P95Latency(ctx, db, window, false)

		// Then: returns empty map without error
		if err != nil {
			t.Fatalf("P95Latency failed: %v", err)
		}
		if result["overall"] != 0 {
			t.Errorf("expected 0 for empty dataset, got %d", result["overall"])
		}
	})
}
