package suggestions

import (
	"fmt"
	"sort"

	"github.com/royisme/bobamixer/internal/domain/stats"
	"github.com/royisme/bobamixer/internal/store/sqlite"
)

// SuggestionType represents the type of suggestion
type SuggestionType int

const (
	SuggestionCostOptimization SuggestionType = iota
	SuggestionProfileSwitch
	SuggestionBudgetAdjust
	SuggestionUsagePattern
	SuggestionAnomaly
)

// Suggestion represents a usage optimization suggestion
type Suggestion struct {
	Type        SuggestionType
	Title       string
	Description string
	Impact      string   // Expected impact (e.g., "Save $5/day")
	Priority    int      // 1-5, where 5 is highest priority
	ActionItems []string // Recommended actions
	Data        SuggestionData
}

// SuggestionData contains supporting data for suggestions
type SuggestionData struct {
	CurrentCost      float64
	EstimatedCost    float64
	Savings          float64
	CurrentProfile   string
	SuggestedProfile string
	AffectedDays     int
}

// Engine generates usage optimization suggestions
type Engine struct {
	db       *sqlite.DB
	analyzer *stats.Analyzer
}

// NewEngine creates a new suggestion engine
func NewEngine(db *sqlite.DB) *Engine {
	return &Engine{
		db:       db,
		analyzer: stats.NewAnalyzer(db),
	}
}

// GenerateSuggestions analyzes usage and generates suggestions
func (e *Engine) GenerateSuggestions(days int) ([]Suggestion, error) {
	var suggestions []Suggestion

	// Get usage trend
	trend, err := e.analyzer.GetTrend(days)
	if err != nil {
		return nil, err
	}

	// Get profile breakdown
	profileStats, err := e.analyzer.GetProfileStats(days)
	if err != nil {
		return nil, err
	}

	// Analyze cost trend
	costSugg := e.analyzeCostTrend(trend)
	if costSugg != nil {
		suggestions = append(suggestions, *costSugg)
	}

	// Analyze profile usage
	profileSugg := e.analyzeProfileUsage(profileStats, trend)
	if profileSugg != nil {
		suggestions = append(suggestions, *profileSugg)
	}

	// Check for anomalies
	anomalySugg := e.detectAnomalies(trend)
	if anomalySugg != nil {
		suggestions = append(suggestions, *anomalySugg)
	}

	// Budget optimization
	budgetSugg := e.suggestBudgetAdjustment(trend)
	if budgetSugg != nil {
		suggestions = append(suggestions, *budgetSugg)
	}

	// Sort by priority
	sort.Slice(suggestions, func(i, j int) bool {
		return suggestions[i].Priority > suggestions[j].Priority
	})

	return suggestions, nil
}

// analyzeCostTrend analyzes cost trend and suggests optimizations
func (e *Engine) analyzeCostTrend(trend *stats.Trend) *Suggestion {
	if trend == nil || len(trend.DataPoints) < 3 {
		return nil
	}

	trendDir := stats.DetectTrend(trend.DataPoints)

	if trendDir == "increasing" {
		// Calculate rate of increase
		avgCost := trend.Summary.AvgDailyCost
		recent := trend.DataPoints[len(trend.DataPoints)-1].Cost
		increase := ((recent - avgCost) / avgCost) * 100

		if increase > 20 {
			return &Suggestion{
				Type:        SuggestionCostOptimization,
				Title:       "Rising Cost Trend Detected",
				Description: fmt.Sprintf("Your daily costs have increased by %.0f%% recently. Consider optimizing usage or switching to more cost-effective models.", increase),
				Impact:      fmt.Sprintf("Could save $%.2f/day", recent-avgCost),
				Priority:    4,
				ActionItems: []string{
					"Review recent high-cost sessions",
					"Consider using smaller models for simple tasks",
					"Implement caching to reduce redundant API calls",
					"Set daily budget limits to control spending",
				},
				Data: SuggestionData{
					CurrentCost:   recent,
					EstimatedCost: avgCost,
					Savings:       recent - avgCost,
					AffectedDays:  len(trend.DataPoints),
				},
			}
		}
	}

	return nil
}

// analyzeProfileUsage analyzes profile usage patterns
func (e *Engine) analyzeProfileUsage(profiles []stats.ProfileStats, trend *stats.Trend) *Suggestion {
	if len(profiles) < 2 || trend == nil {
		return nil
	}

	// Find most expensive profile
	var mostExpensive stats.ProfileStats
	maxCost := 0.0
	for _, p := range profiles {
		if p.TotalCost > maxCost {
			maxCost = p.TotalCost
			mostExpensive = p
		}
	}

	// If one profile dominates costs (>60%), suggest alternatives
	if mostExpensive.CostPercent > 60 {
		var suggested string
		minCost := mostExpensive.TotalCost
		for _, p := range profiles {
			if p.ProfileName == mostExpensive.ProfileName {
				continue
			}
			if p.TotalCost < minCost {
				minCost = p.TotalCost
				suggested = p.ProfileName
			}
		}
		return &Suggestion{
			Type:        SuggestionProfileSwitch,
			Title:       "High Dependency on Expensive Profile",
			Description: fmt.Sprintf("Profile '%s' accounts for %.0f%% of your costs. Consider using cheaper alternatives for routine tasks.", mostExpensive.ProfileName, mostExpensive.CostPercent),
			Impact:      fmt.Sprintf("Could save $%.2f over %d days", mostExpensive.TotalCost*0.3, len(trend.DataPoints)),
			Priority:    3,
			ActionItems: []string{
				fmt.Sprintf("Use GPT-3.5 or Claude Haiku for simple queries instead of %s", mostExpensive.ProfileName),
				"Create separate profiles for different task complexities",
				"Review routing rules to optimize model selection",
			},
			Data: SuggestionData{
				CurrentProfile:   mostExpensive.ProfileName,
				SuggestedProfile: suggested,
				CurrentCost:      mostExpensive.TotalCost,
				EstimatedCost:    mostExpensive.TotalCost * 0.7,
				Savings:          mostExpensive.TotalCost * 0.3,
			},
		}
	}

	return nil
}

// detectAnomalies detects unusual spending patterns
func (e *Engine) detectAnomalies(trend *stats.Trend) *Suggestion {
	if trend == nil || len(trend.DataPoints) < 7 {
		return nil
	}

	// Calculate average and find outliers
	avgCost := trend.Summary.AvgDailyCost
	var outliers []stats.DataPoint

	for _, dp := range trend.DataPoints {
		if dp.Cost > avgCost*2 {
			outliers = append(outliers, dp)
		}
	}

	if len(outliers) > 0 {
		totalOutlierCost := 0.0
		for _, o := range outliers {
			totalOutlierCost += o.Cost
		}

		return &Suggestion{
			Type:        SuggestionAnomaly,
			Title:       "Unusual Spending Spikes Detected",
			Description: fmt.Sprintf("Found %d day(s) with costs more than 2x the average. Total excess spending: $%.2f", len(outliers), totalOutlierCost-avgCost*float64(len(outliers))),
			Impact:      "Prevent future spikes",
			Priority:    5,
			ActionItems: []string{
				"Review high-cost sessions for unusually long conversations",
				"Check for runaway processes or loops",
				"Consider implementing rate limiting",
				"Set up budget alerts to catch spikes early",
			},
			Data: SuggestionData{
				CurrentCost:   totalOutlierCost,
				EstimatedCost: avgCost * float64(len(outliers)),
				Savings:       totalOutlierCost - avgCost*float64(len(outliers)),
				AffectedDays:  len(outliers),
			},
		}
	}

	return nil
}

// suggestBudgetAdjustment suggests budget adjustments based on usage
func (e *Engine) suggestBudgetAdjustment(trend *stats.Trend) *Suggestion {
	if trend == nil || len(trend.DataPoints) < 7 {
		return nil
	}

	avgDaily := trend.Summary.AvgDailyCost
	peakDaily := trend.Summary.PeakDayCost

	// If peak is much higher than average, suggest buffer
	if peakDaily > avgDaily*1.5 {
		recommendedDaily := peakDaily * 1.2 // 20% buffer above peak

		return &Suggestion{
			Type:        SuggestionBudgetAdjust,
			Title:       "Budget Optimization Recommended",
			Description: fmt.Sprintf("Your peak daily cost ($%.2f) is %.0f%% higher than average ($%.2f). Consider adjusting your budget.", peakDaily, ((peakDaily-avgDaily)/avgDaily)*100, avgDaily),
			Impact:      "Avoid budget overruns",
			Priority:    2,
			ActionItems: []string{
				fmt.Sprintf("Set daily budget to $%.2f (20%% buffer above peak)", recommendedDaily),
				fmt.Sprintf("Set monthly cap to $%.2f", recommendedDaily*30),
				"Enable budget alerts at 80% threshold",
			},
			Data: SuggestionData{
				CurrentCost:   avgDaily,
				EstimatedCost: recommendedDaily,
				AffectedDays:  len(trend.DataPoints),
			},
		}
	}

	return nil
}

// FormatSuggestion formats a suggestion for display
func (s *Suggestion) FormatSuggestion() string {
	var typeStr string
	switch s.Type {
	case SuggestionCostOptimization:
		typeStr = "💰 Cost Optimization"
	case SuggestionProfileSwitch:
		typeStr = "🔄 Profile Recommendation"
	case SuggestionBudgetAdjust:
		typeStr = "📊 Budget Adjustment"
	case SuggestionUsagePattern:
		typeStr = "📈 Usage Pattern"
	case SuggestionAnomaly:
		typeStr = "⚠️  Anomaly Detection"
	}

	priority := ""
	for i := 0; i < s.Priority; i++ {
		priority += "★"
	}

	result := fmt.Sprintf("%s [%s]\n%s\n\n%s\nImpact: %s\n",
		typeStr, priority, s.Title, s.Description, s.Impact)

	if len(s.ActionItems) > 0 {
		result += "\nRecommended Actions:\n"
		for i, action := range s.ActionItems {
			result += fmt.Sprintf("  %d. %s\n", i+1, action)
		}
	}

	return result
}

// GetPriority returns the priority level as string
func (s *Suggestion) GetPriority() string {
	switch s.Priority {
	case 5:
		return "Critical"
	case 4:
		return "High"
	case 3:
		return "Medium"
	case 2:
		return "Low"
	default:
		return "Info"
	}
}

// TypeToString converts SuggestionType to string
func (t SuggestionType) String() string {
	switch t {
	case SuggestionCostOptimization:
		return "cost_optimization"
	case SuggestionProfileSwitch:
		return "profile_switch"
	case SuggestionBudgetAdjust:
		return "budget_adjust"
	case SuggestionUsagePattern:
		return "usage_pattern"
	case SuggestionAnomaly:
		return "anomaly"
	default:
		return "unknown"
	}
}
