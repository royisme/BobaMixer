package suggestions

import (
	"path/filepath"
	"testing"

	"github.com/royisme/bobamixer/internal/domain/stats"
	"github.com/royisme/bobamixer/internal/store/sqlite"
)

func TestNewEngine(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	engine := NewEngine(db)

	if engine == nil {
		t.Fatal("NewEngine returned nil")
	}

	if engine.db != db {
		t.Error("db not set correctly")
	}

	if engine.analyzer == nil {
		t.Error("analyzer should not be nil")
	}
}

func TestAnalyzeCostTrend(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	engine := NewEngine(db)

	tests := []struct {
		name          string
		trend         *stats.Trend
		shouldSuggest bool
	}{
		{
			name: "increasing trend with significant increase",
			trend: &stats.Trend{
				DataPoints: []stats.DataPoint{
					{Date: "2024-01-01", Cost: 1.0},
					{Date: "2024-01-02", Cost: 1.5},
					{Date: "2024-01-03", Cost: 2.5},
				},
				Summary: stats.Summary{
					AvgDailyCost: 1.5,
				},
			},
			shouldSuggest: true,
		},
		{
			name: "stable trend",
			trend: &stats.Trend{
				DataPoints: []stats.DataPoint{
					{Date: "2024-01-01", Cost: 1.0},
					{Date: "2024-01-02", Cost: 1.0},
					{Date: "2024-01-03", Cost: 1.0},
				},
				Summary: stats.Summary{
					AvgDailyCost: 1.0,
				},
			},
			shouldSuggest: false,
		},
		{
			name:          "nil trend",
			trend:         nil,
			shouldSuggest: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestion := engine.analyzeCostTrend(tt.trend)

			if tt.shouldSuggest && suggestion == nil {
				t.Error("expected suggestion but got nil")
			}

			if !tt.shouldSuggest && suggestion != nil {
				t.Error("expected no suggestion but got one")
			}

			if suggestion != nil {
				if suggestion.Type != SuggestionCostOptimization {
					t.Errorf("Type = %v, want SuggestionCostOptimization", suggestion.Type)
				}

				if suggestion.Priority == 0 {
					t.Error("Priority should be set")
				}

				if len(suggestion.ActionItems) == 0 {
					t.Error("ActionItems should not be empty")
				}
			}
		})
	}
}

func TestAnalyzeProfileUsage(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	engine := NewEngine(db)

	trend := &stats.Trend{
		DataPoints: []stats.DataPoint{{Date: "2024-01-01", Cost: 10.0}},
	}

	tests := []struct {
		name          string
		profiles      []stats.ProfileStats
		shouldSuggest bool
	}{
		{
			name: "high dependency on expensive profile",
			profiles: []stats.ProfileStats{
				{ProfileName: "gpt-4", TotalCost: 100.0, CostPercent: 80.0},
				{ProfileName: "gpt-3.5", TotalCost: 25.0, CostPercent: 20.0},
			},
			shouldSuggest: true,
		},
		{
			name: "balanced usage",
			profiles: []stats.ProfileStats{
				{ProfileName: "gpt-4", TotalCost: 50.0, CostPercent: 50.0},
				{ProfileName: "gpt-3.5", TotalCost: 50.0, CostPercent: 50.0},
			},
			shouldSuggest: false,
		},
		{
			name:          "no profiles",
			profiles:      []stats.ProfileStats{},
			shouldSuggest: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestion := engine.analyzeProfileUsage(tt.profiles, trend)

			if tt.shouldSuggest && suggestion == nil {
				t.Error("expected suggestion but got nil")
			}

			if !tt.shouldSuggest && suggestion != nil {
				t.Error("expected no suggestion but got one")
			}

			if suggestion != nil {
				if suggestion.Type != SuggestionProfileSwitch {
					t.Errorf("Type = %v, want SuggestionProfileSwitch", suggestion.Type)
				}
			}
		})
	}
}

func TestDetectAnomalies(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	engine := NewEngine(db)

	tests := []struct {
		name         string
		trend        *stats.Trend
		shouldDetect bool
	}{
		{
			name: "with spike",
			trend: &stats.Trend{
				DataPoints: []stats.DataPoint{
					{Date: "2024-01-01", Cost: 1.0},
					{Date: "2024-01-02", Cost: 1.0},
					{Date: "2024-01-03", Cost: 1.0},
					{Date: "2024-01-04", Cost: 1.0},
					{Date: "2024-01-05", Cost: 5.0}, // Spike
					{Date: "2024-01-06", Cost: 1.0},
					{Date: "2024-01-07", Cost: 1.0},
				},
				Summary: stats.Summary{
					AvgDailyCost: 1.57,
				},
			},
			shouldDetect: true,
		},
		{
			name: "no spike",
			trend: &stats.Trend{
				DataPoints: []stats.DataPoint{
					{Date: "2024-01-01", Cost: 1.0},
					{Date: "2024-01-02", Cost: 1.1},
					{Date: "2024-01-03", Cost: 0.9},
					{Date: "2024-01-04", Cost: 1.0},
					{Date: "2024-01-05", Cost: 1.2},
					{Date: "2024-01-06", Cost: 0.8},
					{Date: "2024-01-07", Cost: 1.0},
				},
				Summary: stats.Summary{
					AvgDailyCost: 1.0,
				},
			},
			shouldDetect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestion := engine.detectAnomalies(tt.trend)

			if tt.shouldDetect && suggestion == nil {
				t.Error("expected anomaly detection but got nil")
			}

			if !tt.shouldDetect && suggestion != nil {
				t.Error("expected no anomaly detection but got one")
			}

			if suggestion != nil {
				if suggestion.Type != SuggestionAnomaly {
					t.Errorf("Type = %v, want SuggestionAnomaly", suggestion.Type)
				}

				if suggestion.Priority != 5 {
					t.Errorf("Priority = %d, want 5 for anomalies", suggestion.Priority)
				}
			}
		})
	}
}

func TestSuggestBudgetAdjustment(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	engine := NewEngine(db)

	tests := []struct {
		name          string
		trend         *stats.Trend
		shouldSuggest bool
	}{
		{
			name: "peak much higher than average",
			trend: &stats.Trend{
				DataPoints: []stats.DataPoint{
					{Date: "2024-01-01", Cost: 1.0},
					{Date: "2024-01-02", Cost: 1.0},
					{Date: "2024-01-03", Cost: 3.0}, // Peak
					{Date: "2024-01-04", Cost: 1.0},
					{Date: "2024-01-05", Cost: 1.0},
					{Date: "2024-01-06", Cost: 1.0},
					{Date: "2024-01-07", Cost: 1.0},
				},
				Summary: stats.Summary{
					AvgDailyCost: 1.29,
					PeakDayCost:  3.0,
				},
			},
			shouldSuggest: true,
		},
		{
			name: "consistent usage",
			trend: &stats.Trend{
				DataPoints: []stats.DataPoint{
					{Date: "2024-01-01", Cost: 1.0},
					{Date: "2024-01-02", Cost: 1.1},
					{Date: "2024-01-03", Cost: 1.2},
					{Date: "2024-01-04", Cost: 1.0},
					{Date: "2024-01-05", Cost: 1.1},
					{Date: "2024-01-06", Cost: 1.0},
					{Date: "2024-01-07", Cost: 1.1},
				},
				Summary: stats.Summary{
					AvgDailyCost: 1.07,
					PeakDayCost:  1.2,
				},
			},
			shouldSuggest: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestion := engine.suggestBudgetAdjustment(tt.trend)

			if tt.shouldSuggest && suggestion == nil {
				t.Error("expected budget suggestion but got nil")
			}

			if !tt.shouldSuggest && suggestion != nil {
				t.Error("expected no budget suggestion but got one")
			}

			if suggestion != nil {
				if suggestion.Type != SuggestionBudgetAdjust {
					t.Errorf("Type = %v, want SuggestionBudgetAdjust", suggestion.Type)
				}
			}
		})
	}
}

func TestSuggestionFormatSuggestion(t *testing.T) {
	suggestion := Suggestion{
		Type:        SuggestionCostOptimization,
		Title:       "Test Suggestion",
		Description: "This is a test",
		Impact:      "Save $5/day",
		Priority:    4,
		ActionItems: []string{
			"Action 1",
			"Action 2",
		},
	}

	formatted := suggestion.FormatSuggestion()

	if formatted == "" {
		t.Error("formatted suggestion is empty")
	}

	// Check that it contains key elements
	if !contains(formatted, "Test Suggestion") {
		t.Error("should contain title")
	}

	if !contains(formatted, "Action 1") {
		t.Error("should contain action items")
	}

	if !contains(formatted, "★★★★") {
		t.Error("should contain priority stars")
	}
}

func TestSuggestionGetPriority(t *testing.T) {
	tests := []struct {
		priority int
		want     string
	}{
		{5, "Critical"},
		{4, "High"},
		{3, "Medium"},
		{2, "Low"},
		{1, "Info"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			s := Suggestion{Priority: tt.priority}
			if s.GetPriority() != tt.want {
				t.Errorf("GetPriority() = %s, want %s", s.GetPriority(), tt.want)
			}
		})
	}
}

func TestSuggestionTypeString(t *testing.T) {
	tests := []struct {
		suggType SuggestionType
		want     string
	}{
		{SuggestionCostOptimization, "cost_optimization"},
		{SuggestionProfileSwitch, "profile_switch"},
		{SuggestionBudgetAdjust, "budget_adjust"},
		{SuggestionUsagePattern, "usage_pattern"},
		{SuggestionAnomaly, "anomaly"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if tt.suggType.String() != tt.want {
				t.Errorf("String() = %s, want %s", tt.suggType.String(), tt.want)
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
