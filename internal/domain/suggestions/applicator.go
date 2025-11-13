package suggestions

import (
	"errors"
	"fmt"
	"sort"

	"github.com/royisme/bobamixer/internal/domain/budget"
	"github.com/royisme/bobamixer/internal/store/config"
)

// Applicator applies actionable suggestions to the local environment automatically.
type Applicator struct {
	home     string
	tracker  *budget.Tracker
	profiles config.Profiles
}

// NewApplicator creates an applicator bound to the filesystem home + budget tracker.
func NewApplicator(home string, tracker *budget.Tracker, profiles config.Profiles) *Applicator {
	return &Applicator{home: home, tracker: tracker, profiles: profiles}
}

// Apply executes automatic actions for a suggestion and returns a summary.
func (a *Applicator) Apply(s Suggestion) (string, error) {
	switch s.Type {
	case SuggestionProfileSwitch:
		return a.applyProfileSwitch(&s)
	case SuggestionBudgetAdjust:
		return a.applyBudgetAdjustment(&s)
	default:
		return "no automatic action available", nil
	}
}

func (a *Applicator) applyProfileSwitch(s *Suggestion) (string, error) {
	if len(a.profiles) == 0 {
		return "", errors.New("no profiles loaded")
	}
	target := s.Data.SuggestedProfile
	if target == "" {
		target = a.findAlternateProfile(s.Data.CurrentProfile)
	}
	if target == "" {
		return "", errors.New("no alternate profile available")
	}
	if err := config.SaveActiveProfile(a.home, target); err != nil {
		return "", err
	}
	return fmt.Sprintf("active profile switched to %s", target), nil
}

func (a *Applicator) applyBudgetAdjustment(s *Suggestion) (string, error) {
	if a.tracker == nil {
		return "", errors.New("budget tracker missing")
	}
	recommended := s.Data.EstimatedCost
	if recommended <= 0 {
		recommended = s.Data.CurrentCost
	}
	if recommended <= 0 {
		recommended = 1
	}
	hardCap := recommended * 30
	budget, err := a.tracker.GetGlobalBudget()
	if err != nil {
		if _, createErr := a.tracker.CreateBudget("global", "", recommended, hardCap); createErr != nil {
			return "", createErr
		}
		return fmt.Sprintf("created global budget %.2f / %.2f", recommended, hardCap), nil
	}
	if err := a.tracker.UpdateLimits(budget.ID, recommended, hardCap); err != nil {
		return "", err
	}
	return fmt.Sprintf("updated global budget %.2f / %.2f", recommended, hardCap), nil
}

func (a *Applicator) findAlternateProfile(current string) string {
	if len(a.profiles) == 0 {
		return ""
	}
	keys := make([]string, 0, len(a.profiles))
	for k := range a.profiles {
		if k == current {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	if len(keys) == 0 {
		return ""
	}
	return keys[0]
}
