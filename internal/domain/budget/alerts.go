// Package budget provides budget tracking and alerting functionality for API usage costs.
package budget

import (
	"fmt"
	"time"
)

// AlertLevel represents the severity of a budget alert
type AlertLevel int

const (
	AlertLevelNone AlertLevel = iota
	AlertLevelInfo
	AlertLevelWarning
	AlertLevelCritical
)

const (
	scopeGlobal  = "global"
	scopeProfile = "profile"
	scopeProject = "project"
)

// Alert represents a budget alert/notification
type Alert struct {
	Timestamp  time.Time
	CurrentUSD float64 // current spending
	LimitUSD   float64 // budget limit that was exceeded
	Percent    float64 // percentage of budget used
	Title      string
	Message    string
	Scope      string // "global", "project", "profile"
	Target     string // project name or profile name
	Level      AlertLevel
}

// AlertConfig represents alert configuration
type AlertConfig struct {
	EnableDaily     bool
	EnableCap       bool
	WarningPercent  float64 // Percentage to trigger warning (e.g., 80)
	CriticalPercent float64 // Percentage to trigger critical (e.g., 100)
}

// DefaultAlertConfig returns default alert configuration
func DefaultAlertConfig() *AlertConfig {
	return &AlertConfig{
		EnableDaily:     true,
		EnableCap:       true,
		WarningPercent:  80.0,
		CriticalPercent: 100.0,
	}
}

// AlertManager manages budget alerts
type AlertManager struct {
	config  *AlertConfig
	tracker *Tracker
	history []Alert
}

// NewAlertManager creates a new alert manager
func NewAlertManager(tracker *Tracker, config *AlertConfig) *AlertManager {
	if config == nil {
		config = DefaultAlertConfig()
	}

	return &AlertManager{
		config:  config,
		tracker: tracker,
		history: make([]Alert, 0),
	}
}

// CheckBudgetAlerts checks budget status and generates alerts if needed
func (am *AlertManager) CheckBudgetAlerts(scope, target string) []Alert {
	status, err := am.tracker.GetStatus(scope, target)
	if err != nil {
		// No budget configured, no alerts
		return nil
	}

	var alerts []Alert

	// Check daily limit
	if am.config.EnableDaily && status.DailyLimit > 0 {
		dailyAlert := am.checkThreshold(
			status.DailyProgress,
			status.CurrentSpent,
			status.DailyLimit,
			scope,
			target,
			"daily",
		)
		if dailyAlert != nil {
			alerts = append(alerts, *dailyAlert)
		}
	}

	// Check hard cap
	if am.config.EnableCap && status.HardCap > 0 {
		totalSpent := status.Budget.SpentUSD
		capAlert := am.checkThreshold(
			status.TotalProgress,
			totalSpent,
			status.HardCap,
			scope,
			target,
			"cap",
		)
		if capAlert != nil {
			alerts = append(alerts, *capAlert)
		}
	}

	// Add to history
	am.history = append(am.history, alerts...)

	return alerts
}

// checkThreshold checks if a threshold has been exceeded
func (am *AlertManager) checkThreshold(
	percent float64,
	current float64,
	limit float64,
	scope string,
	target string,
	limitType string,
) *Alert {
	var alert *Alert

	if percent >= am.config.CriticalPercent {
		alert = &Alert{
			Level:      AlertLevelCritical,
			Timestamp:  time.Now(),
			Scope:      scope,
			Target:     target,
			CurrentUSD: current,
			LimitUSD:   limit,
			Percent:    percent,
		}

		if limitType == "daily" {
			alert.Title = "Daily Budget Exceeded"
			alert.Message = fmt.Sprintf(
				"Daily spending ($%.2f) has exceeded the limit ($%.2f) by %.1f%%",
				current, limit, percent-100,
			)
		} else {
			alert.Title = "Budget Cap Exceeded"
			alert.Message = fmt.Sprintf(
				"Total spending ($%.2f) has exceeded the hard cap ($%.2f)",
				current, limit,
			)
		}
	} else if percent >= am.config.WarningPercent {
		alert = &Alert{
			Level:      AlertLevelWarning,
			Timestamp:  time.Now(),
			Scope:      scope,
			Target:     target,
			CurrentUSD: current,
			LimitUSD:   limit,
			Percent:    percent,
		}

		if limitType == "daily" {
			alert.Title = "Approaching Daily Budget Limit"
			alert.Message = fmt.Sprintf(
				"Daily spending is at %.0f%% of the limit ($%.2f / $%.2f)",
				percent, current, limit,
			)
		} else {
			alert.Title = "Approaching Budget Cap"
			alert.Message = fmt.Sprintf(
				"Total spending is at %.0f%% of the hard cap ($%.2f / $%.2f)",
				percent, current, limit,
			)
		}
	}

	return alert
}

// GetRecentAlerts returns recent alerts (last N)
func (am *AlertManager) GetRecentAlerts(count int) []Alert {
	if count <= 0 || len(am.history) == 0 {
		return nil
	}

	start := len(am.history) - count
	if start < 0 {
		start = 0
	}

	return am.history[start:]
}

// GetAlertsByLevel returns alerts filtered by level
func (am *AlertManager) GetAlertsByLevel(level AlertLevel) []Alert {
	var filtered []Alert
	for _, alert := range am.history {
		if alert.Level == level {
			filtered = append(filtered, alert)
		}
	}
	return filtered
}

// ClearHistory clears alert history
func (am *AlertManager) ClearHistory() {
	am.history = make([]Alert, 0)
}

// FormatAlert formats an alert for display
func (alert *Alert) FormatAlert() string {
	var level string
	switch alert.Level {
	case AlertLevelCritical:
		level = "🔴 CRITICAL"
	case AlertLevelWarning:
		level = "🟡 WARNING"
	case AlertLevelInfo:
		level = "🔵 INFO"
	default:
		level = "ℹ️  NOTICE"
	}

	var scopeInfo string
	switch alert.Scope {
	case scopeProfile:
		scopeInfo = fmt.Sprintf("Profile: %s", alert.Target)
	case scopeProject:
		scopeInfo = fmt.Sprintf("Project: %s", alert.Target)
	default:
		scopeInfo = "Global Budget"
	}

	return fmt.Sprintf(
		"%s - %s\n%s\n%s\n%.1f%% of budget used ($%.2f / $%.2f)\nTime: %s",
		level,
		alert.Title,
		scopeInfo,
		alert.Message,
		alert.Percent,
		alert.CurrentUSD,
		alert.LimitUSD,
		alert.Timestamp.Format("2006-01-02 15:04:05"),
	)
}

// ShouldBlock determines if spending should be blocked based on alert level
func (alert *Alert) ShouldBlock() bool {
	// Block spending if critical alert and budget is exceeded
	return alert.Level == AlertLevelCritical && alert.Percent >= 100
}

// GetSuggestion returns a suggestion for how to proceed
func (alert *Alert) GetSuggestion() string {
	switch alert.Level {
	case AlertLevelCritical:
		if alert.Percent >= 100 {
			return "Consider pausing usage or increasing your budget limit to continue."
		}
		return "You've exceeded your budget. Review your spending and consider adjusting limits."

	case AlertLevelWarning:
		return "You're approaching your budget limit. Monitor your usage closely."

	default:
		return "Keep track of your spending to stay within budget."
	}
}

// String converts AlertLevel to string
func (level AlertLevel) String() string {
	switch level {
	case AlertLevelCritical:
		return "critical"
	case AlertLevelWarning:
		return "warning"
	case AlertLevelInfo:
		return "info"
	default:
		return "none"
	}
}
