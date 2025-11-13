package stats

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/royisme/bobamixer/internal/store/sqlite"
)

func TestNewAnalyzer(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	analyzer := NewAnalyzer(db)
	if analyzer == nil {
		t.Error("NewAnalyzer returned nil")
	}
	if analyzer.db != db {
		t.Error("analyzer db not set correctly")
	}
}

func TestGetTrend(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	analyzer := NewAnalyzer(db)

	tests := []struct {
		name        string
		days        int
		expectError bool
	}{
		{
			name:        "7 day trend",
			days:        7,
			expectError: false,
		},
		{
			name:        "30 day trend",
			days:        30,
			expectError: false,
		},
		{
			name:        "1 day trend",
			days:        1,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trend, err := analyzer.GetTrend(tt.days)
			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if trend == nil {
				t.Fatal("trend is nil")
			}

			if trend.Period == "" {
				t.Error("Period should not be empty")
			}

			if trend.StartDate == "" {
				t.Error("StartDate should not be empty")
			}

			if trend.EndDate == "" {
				t.Error("EndDate should not be empty")
			}
		})
	}
}

func TestGetTodayStats(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	analyzer := NewAnalyzer(db)

	dataPoint, err := analyzer.GetTodayStats()
	if err != nil {
		t.Fatalf("GetTodayStats failed: %v", err)
	}

	if dataPoint == nil {
		t.Fatal("dataPoint is nil")
	}

	if dataPoint.Date == "" {
		t.Error("Date should not be empty")
	}
}

func TestGetProfileStats(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	analyzer := NewAnalyzer(db)

	stats, err := analyzer.GetProfileStats(7)
	if err != nil {
		t.Fatalf("GetProfileStats failed: %v", err)
	}

	// Should return empty slice with no data
	if stats == nil {
		t.Error("stats should not be nil (should be empty slice)")
	}
}

func TestComparePeriods(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	analyzer := NewAnalyzer(db)

	comparison, err := analyzer.ComparePeriods(7, 7)
	if err != nil {
		t.Fatalf("ComparePeriods failed: %v", err)
	}

	if comparison == nil {
		t.Fatal("comparison is nil")
	}

	if comparison.Period1 == nil {
		t.Error("Period1 should not be nil")
	}

	if comparison.Period2 == nil {
		t.Error("Period2 should not be nil")
	}
}

func TestGetSparkline(t *testing.T) {
	tests := []struct {
		name       string
		dataPoints []DataPoint
		wantEmpty  bool
	}{
		{
			name: "normal data",
			dataPoints: []DataPoint{
				{Date: "2024-01-01", Cost: 1.0},
				{Date: "2024-01-02", Cost: 2.0},
				{Date: "2024-01-03", Cost: 3.0},
				{Date: "2024-01-04", Cost: 2.5},
				{Date: "2024-01-05", Cost: 1.5},
			},
			wantEmpty: false,
		},
		{
			name:       "empty data",
			dataPoints: []DataPoint{},
			wantEmpty:  true,
		},
		{
			name: "single point",
			dataPoints: []DataPoint{
				{Date: "2024-01-01", Cost: 1.0},
			},
			wantEmpty: false,
		},
		{
			name: "all zeros",
			dataPoints: []DataPoint{
				{Date: "2024-01-01", Cost: 0.0},
				{Date: "2024-01-02", Cost: 0.0},
			},
			wantEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sparkline := GetSparkline(tt.dataPoints)

			if tt.wantEmpty {
				if sparkline != "" {
					t.Errorf("expected empty sparkline, got %s", sparkline)
				}
				return
			}

			// Count runes not bytes
			runeCount := len([]rune(sparkline))
			if runeCount != len(tt.dataPoints) {
				t.Errorf("sparkline length = %d runes, want %d", runeCount, len(tt.dataPoints))
			}

			// Verify sparkline contains valid characters
			validChars := "▁▂▃▄▅▆▇█"
			for _, r := range sparkline {
				if !strings.ContainsRune(validChars, r) {
					t.Errorf("sparkline contains invalid character: %c", r)
				}
			}
		})
	}
}

func TestDetectTrend(t *testing.T) {
	tests := []struct {
		name         string
		dataPoints   []DataPoint
		expectedTrend string
	}{
		{
			name: "increasing trend",
			dataPoints: []DataPoint{
				{Date: "2024-01-01", Cost: 1.0},
				{Date: "2024-01-02", Cost: 2.0},
				{Date: "2024-01-03", Cost: 3.0},
				{Date: "2024-01-04", Cost: 4.0},
			},
			expectedTrend: "increasing",
		},
		{
			name: "decreasing trend",
			dataPoints: []DataPoint{
				{Date: "2024-01-01", Cost: 4.0},
				{Date: "2024-01-02", Cost: 3.0},
				{Date: "2024-01-03", Cost: 2.0},
				{Date: "2024-01-04", Cost: 1.0},
			},
			expectedTrend: "decreasing",
		},
		{
			name: "stable trend",
			dataPoints: []DataPoint{
				{Date: "2024-01-01", Cost: 2.0},
				{Date: "2024-01-02", Cost: 2.1},
				{Date: "2024-01-03", Cost: 2.0},
				{Date: "2024-01-04", Cost: 2.1},
			},
			expectedTrend: "stable",
		},
		{
			name:          "single point",
			dataPoints:    []DataPoint{{Date: "2024-01-01", Cost: 1.0}},
			expectedTrend: "stable",
		},
		{
			name:          "empty data",
			dataPoints:    []DataPoint{},
			expectedTrend: "stable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trend := DetectTrend(tt.dataPoints)
			if trend != tt.expectedTrend {
				t.Errorf("DetectTrend() = %s, want %s", trend, tt.expectedTrend)
			}
		})
	}
}

func TestFormatCurrency(t *testing.T) {
	tests := []struct {
		name   string
		amount float64
		want   string
	}{
		{"zero", 0.0, "$0.0000"},
		{"small", 0.0123, "$0.0123"},
		{"medium", 1.234, "$1.2340"},
		{"large", 123.45, "$123.4500"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatCurrency(tt.amount)
			if got != tt.want {
				t.Errorf("FormatCurrency(%f) = %s, want %s", tt.amount, got, tt.want)
			}
		})
	}
}

func TestFormatTokens(t *testing.T) {
	tests := []struct {
		name   string
		tokens int
		want   string
	}{
		{"zero", 0, "0"},
		{"small", 123, "123"},
		{"thousands", 12345, "12.3K"},
		{"millions", 1234567, "1.2M"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatTokens(tt.tokens)
			if got != tt.want {
				t.Errorf("FormatTokens(%d) = %s, want %s", tt.tokens, got, tt.want)
			}
		})
	}
}

func TestDataPointCreation(t *testing.T) {
	dp := DataPoint{
		Date:   "2024-01-01",
		Tokens: 1000,
		Cost:   0.50,
		Count:  5,
	}

	if dp.Date != "2024-01-01" {
		t.Errorf("Date = %s, want 2024-01-01", dp.Date)
	}
	if dp.Tokens != 1000 {
		t.Errorf("Tokens = %d, want 1000", dp.Tokens)
	}
	if dp.Cost != 0.50 {
		t.Errorf("Cost = %f, want 0.50", dp.Cost)
	}
	if dp.Count != 5 {
		t.Errorf("Count = %d, want 5", dp.Count)
	}
}

func TestTrendStructure(t *testing.T) {
	trend := &Trend{
		Period:    "7d",
		StartDate: "2024-01-01",
		EndDate:   "2024-01-07",
		DataPoints: []DataPoint{
			{Date: "2024-01-01", Cost: 1.0, Tokens: 1000},
		},
		Summary: Summary{
			TotalTokens:    1000,
			TotalCost:      1.0,
			TotalSessions:  5,
			AvgDailyTokens: 142.86,
			AvgDailyCost:   0.14,
		},
	}

	if trend.Period != "7d" {
		t.Errorf("Period = %s, want 7d", trend.Period)
	}
	if len(trend.DataPoints) != 1 {
		t.Errorf("DataPoints length = %d, want 1", len(trend.DataPoints))
	}
	if trend.Summary.TotalTokens != 1000 {
		t.Errorf("TotalTokens = %d, want 1000", trend.Summary.TotalTokens)
	}
}

func TestProfileStatsStructure(t *testing.T) {
	ps := ProfileStats{
		ProfileName:   "gpt-4",
		TotalTokens:   10000,
		TotalCost:     5.00,
		SessionCount:  10,
		AvgLatencyMS:  150.5,
		UsagePercent:  25.0,
		CostPercent:   30.0,
	}

	if ps.ProfileName != "gpt-4" {
		t.Errorf("ProfileName = %s, want gpt-4", ps.ProfileName)
	}
	if ps.TotalTokens != 10000 {
		t.Errorf("TotalTokens = %d, want 10000", ps.TotalTokens)
	}
	if ps.TotalCost != 5.00 {
		t.Errorf("TotalCost = %f, want 5.00", ps.TotalCost)
	}
}

func TestComparisonStructure(t *testing.T) {
	comparison := &Comparison{
		Period1: &Trend{
			Period: "7d",
			Summary: Summary{
				TotalCost:   10.00,
				TotalTokens: 10000,
			},
		},
		Period2: &Trend{
			Period: "7d",
			Summary: Summary{
				TotalCost:   8.00,
				TotalTokens: 8000,
			},
		},
		CostChange:  25.0,
		TokenChange: 25.0,
	}

	if comparison.CostChange != 25.0 {
		t.Errorf("CostChange = %f, want 25.0", comparison.CostChange)
	}
	if comparison.TokenChange != 25.0 {
		t.Errorf("TokenChange = %f, want 25.0", comparison.TokenChange)
	}
}
