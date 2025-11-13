package stats

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/royisme/bobamixer/internal/store/sqlite"
)

func TestAnalyzerTrendAndProfiles(t *testing.T) {
	dir := t.TempDir()
	db, err := sqlite.Open(filepath.Join(dir, "usage.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	base := time.Now().AddDate(0, 0, -2)
	for i := 0; i < 3; i++ {
		day := base.AddDate(0, 0, i)
		ts := day.Unix()
		sessionID := fmt.Sprintf("s%d", i)
		if err := db.Exec(fmt.Sprintf("INSERT INTO sessions (id, started_at, profile, adapter, success, latency_ms) VALUES ('%s', %d, 'p%d', 'http', 1, %d);", sessionID, ts, i%2, 100+i)); err != nil {
			t.Fatalf("insert session: %v", err)
		}
		cost := float64(1 + i)
		stmt := fmt.Sprintf("INSERT INTO usage_records (id, session_id, ts, input_tokens, output_tokens, input_cost, output_cost, model) VALUES ('u%d', '%s', %d, %d, %d, %.2f, %.2f, 'm');", i, sessionID, ts, 100*(i+1), 50*(i+1), cost, cost)
		if err := db.Exec(stmt); err != nil {
			t.Fatalf("insert usage: %v", err)
		}
	}

	analyzer := NewAnalyzer(db)
	trend, err := analyzer.GetTrend(3)
	if err != nil {
		t.Fatalf("GetTrend: %v", err)
	}
	if len(trend.DataPoints) == 0 {
		t.Fatalf("expected datapoints")
	}
	if trend.Summary.TotalTokens == 0 {
		t.Fatalf("summary missing totals")
	}

	profiles, err := analyzer.GetProfileStats(3)
	if err != nil {
		t.Fatalf("GetProfileStats: %v", err)
	}
	if len(profiles) == 0 {
		t.Fatalf("expected profile stats")
	}
}
