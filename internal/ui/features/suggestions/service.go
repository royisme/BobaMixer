// Package suggestions provides the service layer for suggestions view data and logic.
package suggestions

import (
	"fmt"
	"path/filepath"

	"github.com/royisme/bobamixer/internal/domain/suggestions"
	"github.com/royisme/bobamixer/internal/store/sqlite"
	"github.com/royisme/bobamixer/internal/ui/components"
)

// Service manages suggestions data loading and conversion for the suggestions view.
type Service struct {
	home string
}

// NewService creates a new suggestions service.
func NewService(home string) *Service {
	return &Service{
		home: home,
	}
}

// LoadData loads optimization suggestions from the database.
func (s *Service) LoadData(days int) ([]suggestions.Suggestion, error) {
	dbPath := filepath.Join(s.home, "usage.db")
	db, err := sqlite.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	engine := suggestions.NewEngine(db)
	suggs, err := engine.GenerateSuggestions(days)
	if err != nil {
		return nil, fmt.Errorf("generate suggestions: %w", err)
	}

	return suggs, nil
}

// ConvertToView converts domain suggestions to UI components.
func (s *Service) ConvertToView(suggs []suggestions.Suggestion) []components.Suggestion {
	result := make([]components.Suggestion, len(suggs))
	for i, sugg := range suggs {
		result[i] = components.Suggestion{
			Title:       sugg.Title,
			Description: sugg.Description,
			Impact:      sugg.Impact,
			ActionItems: append([]string(nil), sugg.ActionItems...),
			Priority:    sugg.Priority,
			Type:        s.convertType(sugg.Type),
		}
	}
	return result
}

// convertType converts domain SuggestionType to string for UI.
func (s *Service) convertType(t suggestions.SuggestionType) string {
	switch t {
	case suggestions.SuggestionCostOptimization:
		return "cost"
	case suggestions.SuggestionProfileSwitch:
		return "profile"
	case suggestions.SuggestionBudgetAdjust:
		return "budget"
	case suggestions.SuggestionAnomaly:
		return "anomaly"
	case suggestions.SuggestionUsagePattern:
		return "usage"
	default:
		return "usage"
	}
}

// CommandHelp returns the CLI help text for suggestions.
func (s *Service) CommandHelp() string {
	return "Use CLI: boba action [--auto] to apply suggestions"
}
