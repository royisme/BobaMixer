package stats

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/royisme/bobamixer/internal/store/sqlite"
)

func TestAnalyzerTodayAndComparison(t *testing.T) {
	dir := t.TempDir()
	db, err := sqlite.Open(filepath.Join(dir, "usage.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	for i := 0; i < 7; i++ {
		insertUsage(t, db, i, 100*(i+1), float64(i+1), fmt.Sprintf("p%d", i%2))
	}

	analyzer := NewAnalyzer(db)
	today, err := analyzer.GetTodayStats()
	if err != nil {
		t.Fatalf("GetTodayStats: %v", err)
	}
	if today == nil || today.Tokens == 0 {
		t.Fatalf("expected stats for today")
	}

	comparison, err := analyzer.ComparePeriods(3, 7)
	if err != nil {
		t.Fatalf("ComparePeriods: %v", err)
	}
	if comparison == nil || comparison.Period1 == nil || comparison.Period2 == nil {
		t.Fatalf("comparison missing periods: %+v", comparison)
	}
}

func TestSparklineAndFormatters(t *testing.T) {
	points := []DataPoint{
		{Date: "2024-01-01", Cost: 1.0},
		{Date: "2024-01-02", Cost: 5.0},
		{Date: "2024-01-03", Cost: 2.5},
	}
	spark := GetSparkline(points)
	if utf8.RuneCountInString(spark) != len(points) {
		t.Fatalf("sparkline length mismatch: %s", spark)
	}
	if trend := DetectTrend(points); trend == "stable" {
		t.Fatalf("expected non-stable trend")
	}
	if DetectTrend(points[:1]) != "stable" {
		t.Fatalf("single datapoint should be stable")
	}
	if got := FormatTokens(1500000); got != "1.5M" {
		t.Fatalf("FormatTokens: %s", got)
	}
	if got := FormatTokens(1500); got != "1.5K" {
		t.Fatalf("FormatTokens: %s", got)
	}
	if got := FormatTokens(50); got != "50" {
		t.Fatalf("FormatTokens: %s", got)
	}
	if got := FormatCurrency(12.3456); got != "$12.3456" {
		t.Fatalf("FormatCurrency: %s", got)
	}
	if parseInt("bad") != 0 || parseFloat("oops") != 0 {
		t.Fatalf("parsers should return zero on error")
	}
}

func insertUsage(t *testing.T, db *sqlite.DB, dayOffset int, tokens int, cost float64, profile string) {
	t.Helper()
	day := time.Now().AddDate(0, 0, -dayOffset)
	ts := day.Unix()
	sessionID := fmt.Sprintf("s-%s-%d", profile, dayOffset)
	stmtSession := fmt.Sprintf("INSERT INTO sessions (id, started_at, profile, adapter, success, latency_ms) VALUES ('%s', %d, '%s', 'http', 1, 100);", sessionID, ts, profile)
	if err := db.Exec(stmtSession); err != nil {
		t.Fatalf("insert session: %v", err)
	}
	stmtUsage := fmt.Sprintf("INSERT INTO usage_records (id, session_id, ts, input_tokens, output_tokens, input_cost, output_cost, model) VALUES ('u-%s-%d', '%s', %d, %d, %d, %.2f, %.2f, 'model');", profile, dayOffset, sessionID, ts, tokens/2, tokens/2, cost/2, cost/2)
	if err := db.Exec(stmtUsage); err != nil {
		t.Fatalf("insert usage: %v", err)
	}
}
