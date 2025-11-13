package budget

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/royisme/bobamixer/internal/store/sqlite"
)

func TestNewAlertManager(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	tracker := NewTracker(db)
	am := NewAlertManager(tracker, nil)

	if am == nil {
		t.Fatal("NewAlertManager returned nil")
	}

	if am.config == nil {
		t.Error("config should not be nil")
	}

	if am.tracker != tracker {
		t.Error("tracker not set correctly")
	}
}

func TestDefaultAlertConfig(t *testing.T) {
	config := DefaultAlertConfig()

	if !config.EnableDaily {
		t.Error("EnableDaily should be true by default")
	}

	if !config.EnableCap {
		t.Error("EnableCap should be true by default")
	}

	if config.WarningPercent != 80.0 {
		t.Errorf("WarningPercent = %f, want 80.0", config.WarningPercent)
	}

	if config.CriticalPercent != 100.0 {
		t.Errorf("CriticalPercent = %f, want 100.0", config.CriticalPercent)
	}
}

func TestCheckBudgetAlerts(t *testing.T) {
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

	am := NewAlertManager(tracker, DefaultAlertConfig())

	// Check alerts (should be none initially with no spending)
	alerts := am.CheckBudgetAlerts("project", "test-project")

	// With no spending, there should be no alerts
	if len(alerts) > 0 {
		t.Errorf("expected no alerts with no spending, got %d", len(alerts))
	}
}

func TestCheckThreshold(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	tracker := NewTracker(db)
	am := NewAlertManager(tracker, DefaultAlertConfig())

	tests := []struct {
		name          string
		percent       float64
		current       float64
		limit         float64
		expectedLevel AlertLevel
		shouldAlert   bool
	}{
		{
			name:          "under threshold",
			percent:       50.0,
			current:       5.0,
			limit:         10.0,
			expectedLevel: AlertLevelNone,
			shouldAlert:   false,
		},
		{
			name:          "warning threshold",
			percent:       85.0,
			current:       8.5,
			limit:         10.0,
			expectedLevel: AlertLevelWarning,
			shouldAlert:   true,
		},
		{
			name:          "critical threshold",
			percent:       105.0,
			current:       10.5,
			limit:         10.0,
			expectedLevel: AlertLevelCritical,
			shouldAlert:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alert := am.checkThreshold(tt.percent, tt.current, tt.limit, "project", "test", "daily")

			if tt.shouldAlert {
				if alert == nil {
					t.Error("expected alert but got nil")
					return
				}

				if alert.Level != tt.expectedLevel {
					t.Errorf("Level = %v, want %v", alert.Level, tt.expectedLevel)
				}

				if alert.CurrentUSD != tt.current {
					t.Errorf("CurrentUSD = %f, want %f", alert.CurrentUSD, tt.current)
				}

				if alert.LimitUSD != tt.limit {
					t.Errorf("LimitUSD = %f, want %f", alert.LimitUSD, tt.limit)
				}
			} else {
				if alert != nil {
					t.Error("expected no alert but got one")
				}
			}
		})
	}
}

func TestGetRecentAlerts(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	tracker := NewTracker(db)
	am := NewAlertManager(tracker, DefaultAlertConfig())

	// Add some alerts to history
	for i := 0; i < 5; i++ {
		alert := Alert{
			Level:      AlertLevelWarning,
			Title:      "Test Alert",
			Timestamp:  time.Now(),
			Scope:      "project",
			Target:     "test",
			CurrentUSD: float64(i),
			LimitUSD:   10.0,
			Percent:    float64(i * 10),
		}
		am.history = append(am.history, alert)
	}

	// Get recent 3 alerts
	recent := am.GetRecentAlerts(3)
	if len(recent) != 3 {
		t.Errorf("got %d alerts, want 3", len(recent))
	}

	// Get more than available
	recent = am.GetRecentAlerts(10)
	if len(recent) != 5 {
		t.Errorf("got %d alerts, want 5", len(recent))
	}

	// Get 0 alerts
	recent = am.GetRecentAlerts(0)
	if recent != nil {
		t.Error("expected nil for count 0")
	}
}

func TestGetAlertsByLevel(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	tracker := NewTracker(db)
	am := NewAlertManager(tracker, DefaultAlertConfig())

	// Add alerts with different levels
	levels := []AlertLevel{AlertLevelWarning, AlertLevelCritical, AlertLevelWarning, AlertLevelInfo}
	for _, level := range levels {
		alert := Alert{
			Level:     level,
			Title:     "Test",
			Timestamp: time.Now(),
		}
		am.history = append(am.history, alert)
	}

	// Get warning alerts
	warnings := am.GetAlertsByLevel(AlertLevelWarning)
	if len(warnings) != 2 {
		t.Errorf("got %d warning alerts, want 2", len(warnings))
	}

	// Get critical alerts
	critical := am.GetAlertsByLevel(AlertLevelCritical)
	if len(critical) != 1 {
		t.Errorf("got %d critical alerts, want 1", len(critical))
	}
}

func TestClearHistory(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	tracker := NewTracker(db)
	am := NewAlertManager(tracker, DefaultAlertConfig())

	// Add some alerts
	am.history = append(am.history, Alert{Level: AlertLevelWarning})
	am.history = append(am.history, Alert{Level: AlertLevelCritical})

	if len(am.history) != 2 {
		t.Fatalf("setup failed: expected 2 alerts, got %d", len(am.history))
	}

	// Clear history
	am.ClearHistory()

	if len(am.history) != 0 {
		t.Errorf("after clear, got %d alerts, want 0", len(am.history))
	}
}

func TestAlertFormatAlert(t *testing.T) {
	alert := Alert{
		Level:      AlertLevelCritical,
		Title:      "Test Alert",
		Message:    "Test message",
		Timestamp:  time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		Scope:      "profile",
		Target:     "gpt-4",
		CurrentUSD: 15.0,
		LimitUSD:   10.0,
		Percent:    150.0,
	}

	formatted := alert.FormatAlert()

	if formatted == "" {
		t.Error("formatted alert is empty")
	}

	// Check that it contains key information
	if !contains(formatted, "CRITICAL") {
		t.Error("should contain CRITICAL")
	}

	if !contains(formatted, "Test Alert") {
		t.Error("should contain title")
	}

	if !contains(formatted, "gpt-4") {
		t.Error("should contain target")
	}
}

func TestAlertShouldBlock(t *testing.T) {
	tests := []struct {
		name        string
		level       AlertLevel
		percent     float64
		shouldBlock bool
	}{
		{
			name:        "critical over 100%",
			level:       AlertLevelCritical,
			percent:     105.0,
			shouldBlock: true,
		},
		{
			name:        "critical at 100%",
			level:       AlertLevelCritical,
			percent:     100.0,
			shouldBlock: true,
		},
		{
			name:        "critical under 100%",
			level:       AlertLevelCritical,
			percent:     95.0,
			shouldBlock: false,
		},
		{
			name:        "warning",
			level:       AlertLevelWarning,
			percent:     105.0,
			shouldBlock: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alert := Alert{
				Level:   tt.level,
				Percent: tt.percent,
			}

			if alert.ShouldBlock() != tt.shouldBlock {
				t.Errorf("ShouldBlock() = %v, want %v", alert.ShouldBlock(), tt.shouldBlock)
			}
		})
	}
}

func TestAlertGetSuggestion(t *testing.T) {
	tests := []struct {
		name    string
		level   AlertLevel
		percent float64
	}{
		{"critical exceeded", AlertLevelCritical, 105.0},
		{"critical at limit", AlertLevelCritical, 99.0},
		{"warning", AlertLevelWarning, 85.0},
		{"info", AlertLevelInfo, 50.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alert := Alert{
				Level:   tt.level,
				Percent: tt.percent,
			}

			suggestion := alert.GetSuggestion()
			if suggestion == "" {
				t.Error("suggestion should not be empty")
			}
		})
	}
}

func TestAlertLevelString(t *testing.T) {
	tests := []struct {
		level AlertLevel
		want  string
	}{
		{AlertLevelCritical, "critical"},
		{AlertLevelWarning, "warning"},
		{AlertLevelInfo, "info"},
		{AlertLevelNone, "none"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if tt.level.String() != tt.want {
				t.Errorf("String() = %s, want %s", tt.level.String(), tt.want)
			}
		})
	}
}
