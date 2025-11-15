package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/royisme/bobamixer/internal/store/sqlite"
)

func TestRunStatsToday(t *testing.T) {
	home := t.TempDir()
	db := openUsageDB(t, home)
	seedUsageRecord(t, db, "s-today-1", "alpha", 100, 50, 0.01, 0, 150)
	seedUsageRecord(t, db, "s-today-2", "beta", 200, 75, 0.02, 0, 275)

	output := captureStdout(t, func() {
		if err := runStats(home, []string{"--today"}); err != nil {
			t.Fatalf("runStats today: %v", err)
		}
	})

	if !strings.Contains(output, "Today's Usage") {
		t.Fatalf("expected today's header, got %q", output)
	}
	if !strings.Contains(output, "Tokens:   425") {
		t.Fatalf("expected total tokens in output, got %q", output)
	}
	if !strings.Contains(output, "Sessions: 2") {
		t.Fatalf("expected session count, got %q", output)
	}
}

func TestRunStatsWindowByProfile(t *testing.T) {
	home := t.TempDir()
	db := openUsageDB(t, home)
	seedUsageRecord(t, db, "s-alpha", "alpha", 100, 40, 0.01, 1, 120)
	seedUsageRecord(t, db, "s-beta", "beta", 200, 100, 0.05, 2, 450)

	output := captureStdout(t, func() {
		if err := runStats(home, []string{"--7d", "--by-profile"}); err != nil {
			t.Fatalf("runStats 7d: %v", err)
		}
	})

	if !strings.Contains(output, "Last 7 Days Usage") {
		t.Fatalf("expected 7-day header, got %q", output)
	}
	if !strings.Contains(output, "By Profile:") {
		t.Fatalf("expected profile breakdown, got %q", output)
	}
	if !strings.Contains(output, "- alpha") || !strings.Contains(output, "- beta") {
		t.Fatalf("expected both profiles in output, got %q", output)
	}
	if !strings.Contains(output, "P95 Latency (ms)") {
		t.Fatalf("expected P95 latency output, got %q", output)
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w
	fn()
	w.Close()
	os.Stdout = orig
	data, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read pipe: %v", err)
	}
	return string(data)
}

func openUsageDB(t *testing.T, home string) *sqlite.DB {
	t.Helper()
	dbPath := filepath.Join(home, "usage.db")
	db, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	return db
}

func seedUsageRecord(t *testing.T, db *sqlite.DB, sessionID, profile string, inputTokens, outputTokens int, cost float64, daysAgo int, latencyMS int) {
	t.Helper()
	ts := time.Now().AddDate(0, 0, -daysAgo).Unix()
	sessionQuery := fmt.Sprintf(`INSERT INTO sessions (id, started_at, ended_at, profile, success, latency_ms)
        VALUES ('%s', %d, %d, '%s', 1, %d);`, sessionID, ts, ts+int64(latencyMS), profile, latencyMS)
	if err := db.Exec(sessionQuery); err != nil {
		t.Fatalf("insert session: %v", err)
	}
	usageQuery := fmt.Sprintf(`INSERT INTO usage_records (id, session_id, ts, input_tokens, output_tokens, input_cost, output_cost, model, estimate_level)
        VALUES ('usage-%s', '%s', %d, %d, %d, %.4f, %.4f, 'model', 'exact');`, sessionID, sessionID, ts, inputTokens, outputTokens, cost/2, cost/2)
	if err := db.Exec(usageQuery); err != nil {
		t.Fatalf("insert usage: %v", err)
	}
}
