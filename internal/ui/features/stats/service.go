// Package stats provides the service layer for stats view data and logic.
package stats

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/royisme/bobamixer/internal/domain/stats"
	"github.com/royisme/bobamixer/internal/store/sqlite"
	"github.com/royisme/bobamixer/internal/ui/components"
)

// Service manages stats data loading and conversion for the stats view.
type Service struct {
	home string
}

// NewService creates a new stats service.
func NewService(home string) *Service {
	return &Service{
		home: home,
	}
}

// LoadData loads all stats data (today, week, profiles) from the database.
func (s *Service) LoadData() (StatsData, error) {
	dbPath := filepath.Join(s.home, "usage.db")
	db, err := sqlite.Open(dbPath)
	if err != nil {
		return StatsData{}, fmt.Errorf("open database: %w", err)
	}

	ctx := context.Background()

	// Load today's stats
	today, err := stats.Today(ctx, db)
	if err != nil {
		return StatsData{}, fmt.Errorf("load today stats: %w", err)
	}

	// Load 7-day stats
	to := time.Now()
	from := to.AddDate(0, 0, -7)
	week, err := stats.Window(ctx, db, from, to)
	if err != nil {
		return StatsData{}, fmt.Errorf("load week stats: %w", err)
	}

	// Load profile stats
	analyzer := stats.NewAnalyzer(db)
	profileStats, err := analyzer.GetProfileStats(7)
	if err != nil {
		// Don't fail if profile stats can't be loaded
		profileStats = []stats.ProfileStats{}
	}

	return StatsData{
		Today:        today,
		Week:         week,
		ProfileStats: profileStats,
	}, nil
}

// StatsData holds the raw domain data loaded from the database.
type StatsData struct {
	Today        stats.Summary
	Week         stats.Summary
	ProfileStats []stats.ProfileStats
}

// ConvertToView converts domain stats data to UI components.
func (s *Service) ConvertToView(data StatsData) ViewData {
	return ViewData{
		Today:    s.convertSummary("📅 Today's Usage", data.Today, false),
		Week:     s.convertSummary("📊 Last 7 Days", data.Week, true),
		Profiles: s.convertProfiles(data.ProfileStats),
	}
}

// ViewData holds the UI-ready data for rendering.
type ViewData struct {
	Today    components.StatsSummary
	Week     components.StatsSummary
	Profiles []components.StatsProfile
}

// convertSummary converts domain Summary to component StatsSummary.
func (s *Service) convertSummary(title string, summary stats.Summary, includeAverages bool) components.StatsSummary {
	return components.StatsSummary{
		Title:          title,
		Tokens:         summary.TotalTokens,
		Cost:           summary.TotalCost,
		Sessions:       summary.TotalSessions,
		AvgDailyTokens: summary.AvgDailyTokens,
		AvgDailyCost:   summary.AvgDailyCost,
		ShowAverages:   includeAverages,
	}
}

// convertProfiles converts domain ProfileStats to component StatsProfile.
func (s *Service) convertProfiles(statsList []stats.ProfileStats) []components.StatsProfile {
	if len(statsList) == 0 {
		return nil
	}

	result := make([]components.StatsProfile, 0, len(statsList))
	for _, ps := range statsList {
		result = append(result, components.StatsProfile{
			Name:       ps.ProfileName,
			Tokens:     ps.TotalTokens,
			Cost:       ps.TotalCost,
			Sessions:   ps.SessionCount,
			AvgLatency: ps.AvgLatencyMS,
			UsagePct:   ps.UsagePercent,
			CostPct:    ps.CostPercent,
		})
	}
	return result
}
