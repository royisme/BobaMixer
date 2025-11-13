package budget

import (
	"path/filepath"
	"testing"

	"github.com/royisme/bobamixer/internal/store/sqlite"
)

func TestNewTracker(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	tracker := NewTracker(db)
	if tracker == nil {
		t.Fatal("NewTracker returned nil")
	}
	if tracker.db != db {
		t.Error("tracker db not set correctly")
	}
}

func TestCreateBudget(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	tracker := NewTracker(db)

	tests := []struct {
		name       string
		scope      string
		target     string
		dailyUSD   float64
		hardCapUSD float64
	}{
		{
			name:       "create project budget",
			scope:      "project",
			target:     "test-project",
			dailyUSD:   10.00,
			hardCapUSD: 100.00,
		},
		{
			name:       "create profile budget",
			scope:      "profile",
			target:     "gpt-4",
			dailyUSD:   5.00,
			hardCapUSD: 50.00,
		},
		{
			name:       "create global budget",
			scope:      "global",
			target:     "",
			dailyUSD:   20.00,
			hardCapUSD: 200.00,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			budget, err := tracker.CreateBudget(tt.scope, tt.target, tt.dailyUSD, tt.hardCapUSD)
			if err != nil {
				t.Fatalf("CreateBudget failed: %v", err)
			}

			if budget == nil {
				t.Fatal("budget is nil")
			}

			if budget.Scope != tt.scope {
				t.Errorf("Scope = %s, want %s", budget.Scope, tt.scope)
			}

			if budget.Target != tt.target {
				t.Errorf("Target = %s, want %s", budget.Target, tt.target)
			}

			if budget.DailyUSD != tt.dailyUSD {
				t.Errorf("DailyUSD = %f, want %f", budget.DailyUSD, tt.dailyUSD)
			}

			if budget.HardCapUSD != tt.hardCapUSD {
				t.Errorf("HardCapUSD = %f, want %f", budget.HardCapUSD, tt.hardCapUSD)
			}

			if budget.ID == "" {
				t.Error("budget ID is empty")
			}
		})
	}
}

func TestGetBudget(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	tracker := NewTracker(db)

	// Create a budget first
	created, err := tracker.CreateBudget("project", "test-project", 10.00, 100.00)
	if err != nil {
		t.Fatalf("CreateBudget failed: %v", err)
	}

	// Try to get it
	retrieved, err := tracker.GetBudget("project", "test-project")
	if err != nil {
		t.Fatalf("GetBudget failed: %v", err)
	}

	if retrieved == nil {
		t.Fatal("retrieved budget is nil")
	}

	if retrieved.Scope != created.Scope {
		t.Errorf("Scope = %s, want %s", retrieved.Scope, created.Scope)
	}

	if retrieved.Target != created.Target {
		t.Errorf("Target = %s, want %s", retrieved.Target, created.Target)
	}
}

func TestGetGlobalBudget(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	tracker := NewTracker(db)

	// Create global budget
	_, err = tracker.CreateBudget("global", "", 50.00, 500.00)
	if err != nil {
		t.Fatalf("CreateBudget failed: %v", err)
	}

	// Get global budget
	budget, err := tracker.GetGlobalBudget()
	if err != nil {
		t.Fatalf("GetGlobalBudget failed: %v", err)
	}

	if budget.Scope != scopeGlobal {
		t.Errorf("Scope = %s, want global", budget.Scope)
	}
}

func TestUpdateLimits(t *testing.T) {
	tempDir := t.TempDir()
	db, err := sqlite.Open(filepath.Join(tempDir, "test.db"))
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	tracker := NewTracker(db)
	budget, err := tracker.CreateBudget("global", "", 10, 100)
	if err != nil {
		t.Fatalf("CreateBudget: %v", err)
	}

	if err := tracker.UpdateLimits(budget.ID, 20, 200); err != nil {
		t.Fatalf("UpdateLimits: %v", err)
	}

	updated, err := tracker.GetGlobalBudget()
	if err != nil {
		t.Fatalf("GetGlobalBudget: %v", err)
	}
	if updated.DailyUSD != 20 || updated.HardCapUSD != 200 {
		t.Fatalf("limits not updated: %+v", updated)
	}
}

func TestCheckBudget(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	tracker := NewTracker(db)

	// Create a budget
	_, err = tracker.CreateBudget("project", "test-project", 10.00, 100.00)
	if err != nil {
		t.Fatalf("CreateBudget failed: %v", err)
	}

	// Test checking with no spending yet
	allowed, warning, err := tracker.CheckBudget("project", "test-project", 5.00)
	if err != nil {
		t.Fatalf("CheckBudget failed: %v", err)
	}

	if !allowed {
		t.Error("should allow spending under limit")
	}

	if warning != "" {
		t.Errorf("unexpected warning: %s", warning)
	}

	// Test checking budget that doesn't exist (should allow by default)
	allowed, warning, err = tracker.CheckBudget("profile", "non-existent", 100.00)
	if err != nil {
		t.Errorf("CheckBudget with non-existent budget should not error: %v", err)
	}

	if !allowed {
		t.Error("should allow spending when budget doesn't exist")
	}

	if warning != "" {
		t.Errorf("unexpected warning for non-existent budget: %s", warning)
	}
}

func TestStatusGetWarningLevel(t *testing.T) {
	tests := []struct {
		name          string
		status        *Status
		expectedLevel string
	}{
		{
			name: "no warning",
			status: &Status{
				DailyProgress: 50,
				TotalProgress: 50,
				IsOverDaily:   false,
				IsOverCap:     false,
			},
			expectedLevel: "none",
		},
		{
			name: "warning level",
			status: &Status{
				DailyProgress: 85,
				TotalProgress: 70,
				IsOverDaily:   false,
				IsOverCap:     false,
			},
			expectedLevel: "warning",
		},
		{
			name: "critical level - over daily",
			status: &Status{
				DailyProgress: 120,
				TotalProgress: 70,
				IsOverDaily:   true,
				IsOverCap:     false,
			},
			expectedLevel: "critical",
		},
		{
			name: "critical level - over cap",
			status: &Status{
				DailyProgress: 50,
				TotalProgress: 110,
				IsOverDaily:   false,
				IsOverCap:     true,
			},
			expectedLevel: "critical",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level := tt.status.GetWarningLevel()
			if level != tt.expectedLevel {
				t.Errorf("GetWarningLevel() = %s, want %s", level, tt.expectedLevel)
			}
		})
	}
}

func TestStatusFormatStatus(t *testing.T) {
	budget := &Budget{
		DailyUSD:   10.00,
		HardCapUSD: 100.00,
		SpentUSD:   50.00,
	}

	status := &Status{
		Budget:        budget,
		CurrentSpent:  5.00,
		DailyLimit:    10.00,
		HardCap:       100.00,
		DailyProgress: 50.0,
		TotalProgress: 50.0,
		DaysRemaining: 15,
	}

	formatted := status.FormatStatus()
	if formatted == "" {
		t.Error("FormatStatus returned empty string")
	}

	// Check that it contains key information
	if !contains(formatted, "5.00") {
		t.Error("formatted status should contain current spent")
	}
	if !contains(formatted, "10.00") {
		t.Error("formatted status should contain daily limit")
	}
	if !contains(formatted, "100.00") {
		t.Error("formatted status should contain hard cap")
	}
	if !contains(formatted, "15") {
		t.Error("formatted status should contain days remaining")
	}
}

func TestGenerateID(t *testing.T) {
	id1 := generateID()
	id2 := generateID()

	if id1 == "" {
		t.Error("generateID returned empty string")
	}

	if id1 == id2 {
		t.Error("generateID should return unique IDs")
	}

	if !contains(id1, "budget_") {
		t.Error("generateID should contain 'budget_' prefix")
	}
}

func TestEscape(t *testing.T) {
	// Simple test for escape function
	input := "test-string"
	output := escape(input)

	if output != input {
		t.Errorf("escape(%s) = %s, want %s", input, output, input)
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
