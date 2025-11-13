package stats

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/royisme/bobamixer/internal/store/sqlite"
)

// DataPoint represents a single data point in a time series
type DataPoint struct {
	Date   string  // YYYY-MM-DD
	Tokens int     // total tokens
	Cost   float64 // total cost in USD
	Count  int     // number of sessions
}

// Trend represents usage trend analysis
type Trend struct {
	Period     string      // "7d", "30d", "all"
	StartDate  string      // YYYY-MM-DD
	EndDate    string      // YYYY-MM-DD
	DataPoints []DataPoint // daily data points
	Summary    Summary     // aggregate summary
}

// Summary provides aggregate statistics
type Summary struct {
	TotalTokens    int
	TotalCost      float64
	TotalSessions  int
	AvgDailyTokens float64
	AvgDailyCost   float64
	PeakDayCost    float64
	PeakDayDate    string
	Trend          string // "increasing", "decreasing", "stable"
}

// ProfileStats represents statistics for a specific profile
type ProfileStats struct {
	ProfileName  string
	TotalTokens  int
	TotalCost    float64
	SessionCount int
	AvgLatencyMS float64
	UsagePercent float64 // percentage of total usage
	CostPercent  float64 // percentage of total cost
}

// Analyzer provides statistical analysis
type Analyzer struct {
	db *sqlite.DB
}

// NewAnalyzer creates a new stats analyzer
func NewAnalyzer(db *sqlite.DB) *Analyzer {
	return &Analyzer{db: db}
}

// GetTrend retrieves usage trend for the specified period
func (a *Analyzer) GetTrend(days int) (*Trend, error) {
	now := time.Now()
	startDate := now.AddDate(0, 0, -days+1)
	startDateStr := startDate.Format("2006-01-02")
	endDateStr := now.Format("2006-01-02")

	query := fmt.Sprintf(`
		SELECT
			date(ts, 'unixepoch') as date,
			SUM(input_tokens + output_tokens) as tokens,
			SUM(input_cost + output_cost) as cost,
			COUNT(DISTINCT session_id) as sessions
		FROM usage_records
		WHERE date(ts, 'unixepoch') >= '%s'
		GROUP BY date
		ORDER BY date;
	`, startDateStr)

	trend := &Trend{
		Period:     fmt.Sprintf("%dd", days),
		StartDate:  startDateStr,
		EndDate:    endDateStr,
		DataPoints: []DataPoint{},
		Summary:    Summary{},
	}

	rows, err := a.db.QueryRows(query)
	if err != nil {
		return trend, err
	}
	var totalTokens int
	var totalCost float64
	var totalSessions int
	var peakCost float64
	var peakDate string
	for _, row := range rows {
		parts := strings.Split(row, "|")
		if len(parts) < 4 {
			continue
		}
		tokens := parseInt(parts[1])
		cost := parseFloat(parts[2])
		sessions := parseInt(parts[3])
		trend.DataPoints = append(trend.DataPoints, DataPoint{
			Date:   parts[0],
			Tokens: tokens,
			Cost:   cost,
			Count:  sessions,
		})
		totalTokens += tokens
		totalCost += cost
		totalSessions += sessions
		if cost > peakCost {
			peakCost = cost
			peakDate = parts[0]
		}
	}
	if len(trend.DataPoints) > 0 {
		daysCount := float64(len(trend.DataPoints))
		trend.Summary = Summary{
			TotalTokens:    totalTokens,
			TotalCost:      totalCost,
			TotalSessions:  totalSessions,
			AvgDailyTokens: float64(totalTokens) / daysCount,
			AvgDailyCost:   totalCost / daysCount,
			PeakDayCost:    peakCost,
			PeakDayDate:    peakDate,
			Trend:          DetectTrend(trend.DataPoints),
		}
	}

	return trend, nil
}

// GetTodayStats retrieves today's statistics
func (a *Analyzer) GetTodayStats() (*DataPoint, error) {
	query := `
		SELECT
			date('now') as date,
			COALESCE(SUM(input_tokens + output_tokens), 0) as tokens,
			COALESCE(SUM(input_cost + output_cost), 0) as cost,
			COUNT(DISTINCT session_id) as sessions
		FROM usage_records
		WHERE date(ts, 'unixepoch') = date('now');
	`

	row, err := a.db.QueryRow(query)
	if err != nil {
		return nil, err
	}
	parts := strings.Split(row, "|")
	if len(parts) < 4 {
		return nil, fmt.Errorf("unexpected row: %s", row)
	}
	dp := &DataPoint{
		Date:   parts[0],
		Tokens: parseInt(parts[1]),
		Cost:   parseFloat(parts[2]),
		Count:  parseInt(parts[3]),
	}
	return dp, nil
}

// GetProfileStats retrieves statistics grouped by profile
func (a *Analyzer) GetProfileStats(days int) ([]ProfileStats, error) {
	startDate := time.Now().AddDate(0, 0, -days+1).Format("2006-01-02")

	query := fmt.Sprintf(`
		SELECT
			COALESCE(s.profile, 'unknown') as profile,
			SUM(u.input_tokens + u.output_tokens) as tokens,
			SUM(u.input_cost + u.output_cost) as cost,
			COUNT(DISTINCT u.session_id) as sessions,
			AVG(s.latency_ms) as avg_latency
		FROM usage_records u
		LEFT JOIN sessions s ON u.session_id = s.id
		WHERE date(u.ts, 'unixepoch') >= '%s'
		GROUP BY profile
		ORDER BY cost DESC;
	`, startDate)

	rows, err := a.db.QueryRows(query)
	if err != nil {
		return nil, err
	}
	stats := make([]ProfileStats, 0, len(rows))
	var totalTokens float64
	var totalCost float64
	for _, row := range rows {
		parts := strings.Split(row, "|")
		if len(parts) < 5 {
			continue
		}
		tokens := parseInt(parts[1])
		cost := parseFloat(parts[2])
		sessions := parseInt(parts[3])
		avgLatency := parseFloat(parts[4])
		stats = append(stats, ProfileStats{
			ProfileName:  parts[0],
			TotalTokens:  tokens,
			TotalCost:    cost,
			SessionCount: sessions,
			AvgLatencyMS: avgLatency,
		})
		totalTokens += float64(tokens)
		totalCost += cost
	}
	for i := range stats {
		if totalTokens > 0 {
			stats[i].UsagePercent = float64(stats[i].TotalTokens) / totalTokens * 100
		}
		if totalCost > 0 {
			stats[i].CostPercent = stats[i].TotalCost / totalCost * 100
		}
	}
	return stats, nil
}

// ComparePeriods compares two time periods
func (a *Analyzer) ComparePeriods(days1, days2 int) (*Comparison, error) {
	trend1, err1 := a.GetTrend(days1)
	if err1 != nil {
		return nil, err1
	}

	trend2, err2 := a.GetTrend(days2)
	if err2 != nil {
		return nil, err2
	}

	comparison := &Comparison{
		Period1: trend1,
		Period2: trend2,
	}

	// Calculate change percentages
	if trend2.Summary.TotalCost > 0 {
		comparison.CostChange = ((trend1.Summary.TotalCost - trend2.Summary.TotalCost) / trend2.Summary.TotalCost) * 100
	}

	if trend2.Summary.TotalTokens > 0 {
		comparison.TokenChange = float64((trend1.Summary.TotalTokens - trend2.Summary.TotalTokens)) / float64(trend2.Summary.TotalTokens) * 100
	}

	return comparison, nil
}

// Comparison represents a comparison between two periods
type Comparison struct {
	Period1     *Trend
	Period2     *Trend
	CostChange  float64 // percentage change
	TokenChange float64 // percentage change
}

// GetSparkline generates a simple ASCII sparkline for visualization
func GetSparkline(dataPoints []DataPoint) string {
	if len(dataPoints) == 0 {
		return ""
	}

	// Find max value for scaling
	maxCost := 0.0
	for _, dp := range dataPoints {
		if dp.Cost > maxCost {
			maxCost = dp.Cost
		}
	}

	if maxCost == 0 {
		// Return lowest character for each data point
		result := ""
		for range dataPoints {
			result += "▁"
		}
		return result
	}

	// Sparkline characters from low to high
	chars := []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

	sparkline := ""
	for _, dp := range dataPoints {
		ratio := dp.Cost / maxCost
		index := int(ratio * float64(len(chars)-1))
		if index >= len(chars) {
			index = len(chars) - 1
		}
		sparkline += string(chars[index])
	}

	return sparkline
}

// DetectTrend analyzes data points and detects trend direction
func DetectTrend(dataPoints []DataPoint) string {
	if len(dataPoints) < 2 {
		return "stable"
	}

	// Simple linear trend detection
	firstHalf := dataPoints[:len(dataPoints)/2]
	secondHalf := dataPoints[len(dataPoints)/2:]

	firstAvg := 0.0
	for _, dp := range firstHalf {
		firstAvg += dp.Cost
	}
	firstAvg /= float64(len(firstHalf))

	secondAvg := 0.0
	for _, dp := range secondHalf {
		secondAvg += dp.Cost
	}
	secondAvg /= float64(len(secondHalf))

	change := ((secondAvg - firstAvg) / firstAvg) * 100

	if change > 10 {
		return "increasing"
	} else if change < -10 {
		return "decreasing"
	}

	return "stable"
}

// FormatCurrency formats a float as USD currency
func FormatCurrency(amount float64) string {
	return fmt.Sprintf("$%.4f", amount)
}

// FormatTokens formats token count with K/M suffix
func FormatTokens(tokens int) string {
	if tokens >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(tokens)/1000000)
	} else if tokens >= 1000 {
		return fmt.Sprintf("%.1fK", float64(tokens)/1000)
	}
	return fmt.Sprintf("%d", tokens)
}

func parseInt(raw string) int {
	v, _ := strconv.Atoi(strings.TrimSpace(raw))
	return v
}

func parseFloat(raw string) float64 {
	f, _ := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	return f
}
