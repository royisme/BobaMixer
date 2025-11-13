package notifications

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/royisme/bobamixer/internal/domain/budget"
	"github.com/royisme/bobamixer/internal/domain/session"
	"github.com/royisme/bobamixer/internal/domain/suggestions"
	"github.com/royisme/bobamixer/internal/store/sqlite"
)

func TestNotifierPollBudgetAlert(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "usage.db")
	db, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	tracker := budget.NewTracker(db)
	if _, err := tracker.CreateBudget("global", "", 1.0, 5.0); err != nil {
		t.Fatalf("CreateBudget: %v", err)
	}

	// insert session + usage to exceed daily budget
	sess := &session.Session{ID: "s1", StartedAt: time.Now().Unix(), Profile: "work", Adapter: "http", Success: true}
	if err := sess.Save(db); err != nil {
		t.Fatalf("Save session: %v", err)
	}
	stmt := fmt.Sprintf("INSERT INTO usage_records (id, session_id, ts, input_cost, output_cost) VALUES ('u1', '%s', %d, 2.0, 0.5);", sess.ID, time.Now().Unix())
	if err := db.Exec(stmt); err != nil {
		t.Fatalf("insert usage: %v", err)
	}

	notifier := NewNotifier(tracker, suggestions.NewEngine(db), nil)
	events, err := notifier.Poll()
	if err != nil {
		t.Fatalf("Poll: %v", err)
	}
	if len(events) == 0 {
		t.Fatalf("expected events, got none")
	}

	// subsequent poll should be deduped
	events, err = notifier.Poll()
	if err != nil {
		t.Fatalf("Poll 2: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("expected no new events, got %d", len(events))
	}
}
