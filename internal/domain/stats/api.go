// Package stats provides TDD-spec aligned public APIs for statistics queries.
package stats

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/royisme/bobamixer/internal/store/sqlite"
)

// Today returns today's usage summary.
// This is the unified API for both CLI --today and TUI dashboard.
func Today(ctx context.Context, db *sqlite.DB) (Summary, error) {
	analyzer := NewAnalyzer(db)
	dataPoint, err := analyzer.GetTodayStats()
	if err != nil {
		return Summary{}, fmt.Errorf("get today stats: %w", err)
	}

	// Convert DataPoint to Summary
	return Summary{
		TotalTokens:   dataPoint.Tokens,
		TotalCost:     dataPoint.Cost,
		TotalSessions: dataPoint.Count,
	}, nil
}

// Window returns usage summary for a time window.
func Window(ctx context.Context, db *sqlite.DB, from, to time.Time) (Summary, error) {
	// Calculate days between from and to
	days := int(to.Sub(from).Hours() / 24)
	if days <= 0 {
		days = 1
	}

	analyzer := NewAnalyzer(db)

	// Query usage records within the window
	query := fmt.Sprintf(`
		SELECT
			SUM(input_tokens + output_tokens) as tokens,
			SUM(input_cost + output_cost) as cost,
			COUNT(DISTINCT session_id) as sessions
		FROM usage_records
		WHERE date(ts, 'unixepoch') >= '%s'
		  AND date(ts, 'unixepoch') <= '%s';
	`, from.Format("2006-01-02"), to.Format("2006-01-02"))

	row, err := analyzer.db.QueryRow(query)
	if err != nil {
		return Summary{}, fmt.Errorf("query window: %w", err)
	}

	if row == "" {
		return Summary{}, nil
	}

	parts := strings.Split(row, "|")
	if len(parts) < 3 {
		return Summary{}, fmt.Errorf("unexpected row format: %s", row)
	}

	totalTokens := parseInt(parts[0])
	totalCost := parseFloat(parts[1])
	totalSessions := parseInt(parts[2])

	return Summary{
		TotalTokens:    totalTokens,
		TotalCost:      totalCost,
		TotalSessions:  totalSessions,
		AvgDailyTokens: float64(totalTokens) / float64(days),
		AvgDailyCost:   totalCost / float64(days),
	}, nil
}

// P95Latency returns the 95th percentile latency for a given time window.
// If byProfile is true, returns per-profile latencies; otherwise returns overall P95.
var ErrSchemaTooOld = errors.New("stats schema version too old")

func P95Latency(ctx context.Context, db *sqlite.DB, window time.Duration, byProfile bool) (map[string]int64, error) {
	if err := requireSchemaVersion(db, 3); err != nil {
		return nil, err
	}
	analyzer := NewAnalyzer(db)
	startDate := time.Now().Add(-window).Format("2006-01-02")

	result := make(map[string]int64)

	if byProfile {
		// Get all profiles
		profilesQuery := fmt.Sprintf(`
			SELECT DISTINCT COALESCE(s.profile, 'unknown') as profile
			FROM sessions s
			WHERE date(s.started_at, 'unixepoch') >= '%s';
		`, startDate)

		rows, err := analyzer.db.QueryRows(profilesQuery)
		if err != nil {
			return nil, fmt.Errorf("query profiles: %w", err)
		}

		// Calculate P95 for each profile
		for _, row := range rows {
			profile := strings.TrimSpace(row)
			if profile == "" {
				profile = "unknown"
			}

			p95, err := calculateP95ForProfile(analyzer.db, profile, startDate)
			if err != nil {
				return nil, fmt.Errorf("calculate P95 for %s: %w", profile, err)
			}
			result[profile] = p95
		}
	} else {
		// Overall P95
		p95, err := calculateP95Overall(analyzer.db, startDate)
		if err != nil {
			return nil, fmt.Errorf("calculate overall P95: %w", err)
		}
		result["overall"] = p95
	}

	return result, nil
}

// calculateP95ForProfile calculates P95 latency for a specific profile.
func calculateP95ForProfile(db *sqlite.DB, profile, startDate string) (int64, error) {
	query := fmt.Sprintf(`
		SELECT s.latency_ms
		FROM sessions s
		WHERE COALESCE(s.profile, 'unknown') = '%s'
		  AND date(s.started_at, 'unixepoch') >= '%s'
		  AND s.latency_ms > 0
		ORDER BY s.latency_ms;
	`, sqlEscape(profile), startDate)

	rows, err := db.QueryRows(query)
	if err != nil {
		return 0, err
	}

	if len(rows) == 0 {
		return 0, nil
	}

	// Parse latencies
	latencies := make([]int64, 0, len(rows))
	for _, row := range rows {
		latency := int64(parseInt(row))
		if latency > 0 {
			latencies = append(latencies, latency)
		}
	}

	if len(latencies) == 0 {
		return 0, nil
	}

	return percentile(latencies, 95), nil
}

// calculateP95Overall calculates overall P95 latency.
func calculateP95Overall(db *sqlite.DB, startDate string) (int64, error) {
	query := fmt.Sprintf(`
		SELECT s.latency_ms
		FROM sessions s
		WHERE date(s.started_at, 'unixepoch') >= '%s'
		  AND s.latency_ms > 0
		ORDER BY s.latency_ms;
	`, startDate)

	rows, err := db.QueryRows(query)
	if err != nil {
		return 0, err
	}

	if len(rows) == 0 {
		return 0, nil
	}

	// Parse latencies
	latencies := make([]int64, 0, len(rows))
	for _, row := range rows {
		latency := int64(parseInt(row))
		if latency > 0 {
			latencies = append(latencies, latency)
		}
	}

	if len(latencies) == 0 {
		return 0, nil
	}

	return percentile(latencies, 95), nil
}

// percentile calculates the Pth percentile of a sorted slice.
func percentile(sorted []int64, p int) int64 {
	if len(sorted) == 0 {
		return 0
	}

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})

	index := int(math.Ceil(float64(len(sorted)) * float64(p) / 100.0))
	if index >= len(sorted) {
		index = len(sorted) - 1
	}
	if index < 0 {
		index = 0
	}

	return sorted[index]
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

func requireSchemaVersion(db *sqlite.DB, minVersion int) error {
	version, err := db.QueryInt("PRAGMA user_version;")
	if err != nil {
		return fmt.Errorf("read schema version: %w", err)
	}
	if version < minVersion {
		return fmt.Errorf("schema version %d is below required %d: %w", version, minVersion, ErrSchemaTooOld)
	}
	return nil
}
