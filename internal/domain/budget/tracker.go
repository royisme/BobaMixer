package budget

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/royisme/bobamixer/internal/store/sqlite"
)

// Budget represents a budget configuration
type Budget struct {
	ID          string
	Scope       string  // "global", "project", "profile"
	Target      string  // project name or profile name
	DailyUSD    float64 // daily spending limit
	HardCapUSD  float64 // absolute maximum
	PeriodStart int64   // unix timestamp
	PeriodEnd   int64   // unix timestamp
	SpentUSD    float64 // current spending
}

// Status represents budget status
type Status struct {
	Budget        *Budget
	CurrentSpent  float64
	DailyLimit    float64
	HardCap       float64
	DailyProgress float64 // percentage of daily limit used
	TotalProgress float64 // percentage of hard cap used
	IsOverDaily   bool
	IsOverCap     bool
	DaysRemaining int
}

// Tracker manages budget tracking
type Tracker struct {
	db *sqlite.DB
}

// NewTracker creates a new budget tracker
func NewTracker(db *sqlite.DB) *Tracker {
	return &Tracker{db: db}
}

// CreateBudget creates a new budget
func (t *Tracker) CreateBudget(scope, target string, dailyUSD, hardCapUSD float64) (*Budget, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfMonth := time.Date(now.Year(), now.Month()+1, 0, 23, 59, 59, 0, now.Location())

	budget := &Budget{
		ID:          generateID(),
		Scope:       scope,
		Target:      target,
		DailyUSD:    dailyUSD,
		HardCapUSD:  hardCapUSD,
		PeriodStart: startOfDay.Unix(),
		PeriodEnd:   endOfMonth.Unix(),
		SpentUSD:    0,
	}

	query := fmt.Sprintf(`
		INSERT INTO budgets (id, scope, target, daily_usd, hard_cap, period_start, period_end, spent_usd)
		VALUES ('%s', '%s', '%s', %f, %f, %d, %d, %f);
	`, budget.ID, budget.Scope, escape(budget.Target), budget.DailyUSD, budget.HardCapUSD,
		budget.PeriodStart, budget.PeriodEnd, budget.SpentUSD)

	if err := t.db.Exec(query); err != nil {
		return nil, err
	}

	return budget, nil
}

// GetBudget retrieves a budget by scope and target
func (t *Tracker) GetBudget(scope, target string) (*Budget, error) {
	query := fmt.Sprintf(`
                SELECT id, scope, target, daily_usd, hard_cap, period_start, period_end, spent_usd
                FROM budgets WHERE scope='%s' AND target='%s' LIMIT 1;
        `, scope, escape(target))

	row, err := t.db.QueryRow(query)
	if err != nil {
		return nil, err
	}
	if row == "" {
		return nil, fmt.Errorf("budget not found")
	}
	parts := strings.Split(row, "|")
	if len(parts) < 8 {
		return nil, fmt.Errorf("invalid budget row: %s", row)
	}
	daily, _ := strconv.ParseFloat(parts[3], 64)
	hard, _ := strconv.ParseFloat(parts[4], 64)
	periodStart, _ := strconv.ParseInt(parts[5], 10, 64)
	periodEnd, _ := strconv.ParseInt(parts[6], 10, 64)
	spent, _ := strconv.ParseFloat(parts[7], 64)
	budget := &Budget{
		ID:          parts[0],
		Scope:       parts[1],
		Target:      parts[2],
		DailyUSD:    daily,
		HardCapUSD:  hard,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		SpentUSD:    spent,
	}

	return budget, nil
}

// GetGlobalBudget retrieves the global budget
func (t *Tracker) GetGlobalBudget() (*Budget, error) {
	return t.GetBudget("global", "")
}

// UpdateSpending updates the spent amount for a budget
func (t *Tracker) UpdateSpending(budgetID string, amount float64) error {
	query := fmt.Sprintf(`
                UPDATE budgets SET spent_usd = spent_usd + %f WHERE id='%s';
        `, amount, budgetID)

	return t.db.Exec(query)
}

// UpdateLimits updates the daily and hard cap limits for a budget id.
func (t *Tracker) UpdateLimits(budgetID string, daily, hard float64) error {
	query := fmt.Sprintf(`
            UPDATE budgets SET daily_usd = %f, hard_cap = %f WHERE id='%s';
    `, daily, hard, budgetID)
	return t.db.Exec(query)
}

// GetStatus calculates the current budget status
func (t *Tracker) GetStatus(scope, target string) (*Status, error) {
	budget, err := t.GetBudget(scope, target)
	if err != nil {
		return nil, err
	}

	// Calculate today's spending
	todaySpent, err := t.getTodaySpending(scope, target)
	if err != nil {
		return nil, err
	}

	// Calculate total period spending
	totalSpent, err := t.getPeriodSpending(scope, target, budget.PeriodStart, budget.PeriodEnd)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	periodEnd := time.Unix(budget.PeriodEnd, 0)
	daysRemaining := int(periodEnd.Sub(now).Hours() / 24)

	status := &Status{
		Budget:        budget,
		CurrentSpent:  todaySpent,
		DailyLimit:    budget.DailyUSD,
		HardCap:       budget.HardCapUSD,
		DaysRemaining: daysRemaining,
	}

	// Calculate progress percentages
	if budget.DailyUSD > 0 {
		status.DailyProgress = (todaySpent / budget.DailyUSD) * 100
		status.IsOverDaily = todaySpent > budget.DailyUSD
	}

	if budget.HardCapUSD > 0 {
		status.TotalProgress = (totalSpent / budget.HardCapUSD) * 100
		status.IsOverCap = totalSpent > budget.HardCapUSD
	}

	return status, nil
}

// getTodaySpending calculates spending for today
func (t *Tracker) getTodaySpending(scope, target string) (float64, error) {
	var whereClause string
	if scope == "global" {
		whereClause = ""
	} else if scope == "profile" {
		whereClause = fmt.Sprintf(" AND profile='%s'", escape(target))
	} else if scope == "project" {
		whereClause = fmt.Sprintf(" AND project='%s'", escape(target))
	}

	query := fmt.Sprintf(`
		SELECT COALESCE(SUM(input_cost + output_cost), 0)
		FROM usage_records
		WHERE date(ts, 'unixepoch') = date('now')%s;
	`, whereClause)

	row, err := t.db.QueryRow(query)
	if err != nil {
		return 0, err
	}

	var spent float64
	fmt.Sscanf(row, "%f", &spent)
	return spent, nil
}

// getPeriodSpending calculates spending for a time period
func (t *Tracker) getPeriodSpending(scope, target string, start, end int64) (float64, error) {
	var whereClause string
	if scope == "global" {
		whereClause = ""
	} else if scope == "profile" {
		whereClause = fmt.Sprintf(" AND profile='%s'", escape(target))
	} else if scope == "project" {
		whereClause = fmt.Sprintf(" AND project='%s'", escape(target))
	}

	query := fmt.Sprintf(`
		SELECT COALESCE(SUM(input_cost + output_cost), 0)
		FROM usage_records
		WHERE ts >= %d AND ts <= %d%s;
	`, start, end, whereClause)

	row, err := t.db.QueryRow(query)
	if err != nil {
		return 0, err
	}

	var spent float64
	fmt.Sscanf(row, "%f", &spent)
	return spent, nil
}

// CheckBudget checks if a planned spending would exceed budget
func (t *Tracker) CheckBudget(scope, target string, plannedAmount float64) (bool, string, error) {
	status, err := t.GetStatus(scope, target)
	if err != nil {
		// If budget not found, allow spending
		return true, "", nil
	}

	// Check daily limit
	if status.DailyLimit > 0 {
		projectedDaily := status.CurrentSpent + plannedAmount
		if projectedDaily > status.DailyLimit {
			msg := fmt.Sprintf("Would exceed daily budget: $%.4f / $%.2f (%.1f%%)",
				projectedDaily, status.DailyLimit, (projectedDaily/status.DailyLimit)*100)
			return false, msg, nil
		}
	}

	// Check hard cap
	if status.HardCap > 0 {
		totalSpent, _ := t.getPeriodSpending(scope, target, status.Budget.PeriodStart, status.Budget.PeriodEnd)
		projectedTotal := totalSpent + plannedAmount
		if projectedTotal > status.HardCap {
			msg := fmt.Sprintf("Would exceed hard cap: $%.4f / $%.2f",
				projectedTotal, status.HardCap)
			return false, msg, nil
		}
	}

	return true, "", nil
}

// GetWarningLevel returns the warning level (none, warning, critical)
func (s *Status) GetWarningLevel() string {
	if s.IsOverCap || s.IsOverDaily {
		return "critical"
	}

	if s.DailyProgress > 80 || s.TotalProgress > 80 {
		return "warning"
	}

	return "none"
}

// FormatStatus returns a human-readable status string
func (s *Status) FormatStatus() string {
	return fmt.Sprintf(
		"Daily: $%.4f / $%.2f (%.1f%%) | Total: $%.4f / $%.2f (%.1f%%) | %d days remaining",
		s.CurrentSpent, s.DailyLimit, s.DailyProgress,
		s.Budget.SpentUSD, s.HardCap, s.TotalProgress,
		s.DaysRemaining,
	)
}

func generateID() string {
	return fmt.Sprintf("budget_%d", time.Now().UnixNano())
}

func escape(s string) string {
	// Simple SQL escape - in production use parameterized queries
	return s
}
