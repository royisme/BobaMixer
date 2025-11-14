package suggestions

import (
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/royisme/bobamixer/internal/store/sqlite"
)

func TestStoreLifecycle(t *testing.T) {
	dir := t.TempDir()
	db, err := sqlite.Open(filepath.Join(dir, "usage.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	store := NewStore(db)
	sugg := &StoredSuggestion{
		CreatedAt:      time.Now(),
		SuggestionType: "cost",
		Title:          "Use 'cheaper' model",
		Description:    "Switch to smaller tier",
		ActionCmd:      "boba use fast",
		Status:         StatusNew,
		Context:        `{"profile":"fast"}`,
	}
	if err := store.Save(sugg); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if sugg.ID == "" {
		t.Fatalf("save should assign id")
	}
	actives, err := store.GetActive()
	if err != nil {
		t.Fatalf("GetActive: %v", err)
	}
	if len(actives) != 1 || actives[0].Title != "Use 'cheaper' model" {
		t.Fatalf("unexpected active suggestions: %#v", actives)
	}
	if err := store.Apply(sugg.ID); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if err := store.Ignore(sugg.ID); err != nil {
		t.Fatalf("Ignore: %v", err)
	}
	if err := store.Snooze(sugg.ID, time.Hour); err != nil {
		t.Fatalf("Snooze: %v", err)
	}
}

func TestStoreParseRowsAndEscape(t *testing.T) {
	dir := t.TempDir()
	db, err := sqlite.Open(filepath.Join(dir, "usage.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	store := NewStore(db)
	now := time.Now()
	until := now.Add(time.Hour)
	row := []string{"s1|" + strconv.FormatInt(now.Unix(), 10) + "|cost|title|desc|cmd|new|" + strconv.FormatInt(until.Unix(), 10) + "|ctx"}
	parsed := store.parseRows(row)
	if len(parsed) != 1 || parsed[0].UntilTS == nil {
		t.Fatalf("parseRows failed: %#v", parsed)
	}
	escaped := escapeSQLString("it's fine")
	if escaped != "it''s fine" {
		t.Fatalf("escapeSQLString got %s", escaped)
	}
}
