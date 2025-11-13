package suggestions

import (
	"path/filepath"
	"testing"

	"github.com/royisme/bobamixer/internal/domain/budget"
	"github.com/royisme/bobamixer/internal/store/config"
	"github.com/royisme/bobamixer/internal/store/sqlite"
)

func TestApplicatorProfileSwitch(t *testing.T) {
	dir := t.TempDir()
	profiles := config.Profiles{
		"work": {Key: "work"},
		"alt":  {Key: "alt"},
	}
	db, err := sqlite.Open(filepath.Join(dir, "usage.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	tracker := budget.NewTracker(db)
	app := NewApplicator(dir, tracker, profiles)
	summary, err := app.Apply(Suggestion{Type: SuggestionProfileSwitch, Data: SuggestionData{CurrentProfile: "work"}})
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if summary == "" {
		t.Fatalf("expected summary")
	}
	prof, err := config.LoadActiveProfile(dir)
	if err != nil {
		t.Fatalf("load active: %v", err)
	}
	if prof == "" {
		t.Fatalf("expected active profile written")
	}
}

func TestApplicatorBudgetAdjust(t *testing.T) {
	dir := t.TempDir()
	db, err := sqlite.Open(filepath.Join(dir, "usage.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	tracker := budget.NewTracker(db)
	app := NewApplicator(dir, tracker, config.Profiles{})
	summary, err := app.Apply(Suggestion{Type: SuggestionBudgetAdjust, Data: SuggestionData{EstimatedCost: 5}})
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if summary == "" {
		t.Fatalf("expected summary")
	}
}
