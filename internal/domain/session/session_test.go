package session

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/royisme/bobamixer/internal/store/sqlite"
)

func TestListRecentSessions(t *testing.T) {
	dir := t.TempDir()
	db, err := sqlite.Open(filepath.Join(dir, "usage.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	for i := 0; i < 3; i++ {
		sess := &Session{
			ID:        fmt.Sprintf("s%d", i),
			StartedAt: time.Now().Add(-time.Duration(i) * time.Hour).Unix(),
			Profile:   "work",
			Adapter:   "http",
			Success:   i%2 == 0,
			LatencyMS: int64(100 + i),
			TaskType:  "test",
		}
		if err := sess.Save(db); err != nil {
			t.Fatalf("Save session: %v", err)
		}
	}

	sessions, err := ListRecentSessions(db, 2)
	if err != nil {
		t.Fatalf("ListRecentSessions: %v", err)
	}
	if len(sessions) != 2 {
		t.Fatalf("expected 2, got %d", len(sessions))
	}
	if sessions[0].ID == sessions[1].ID {
		t.Fatalf("expected distinct sessions")
	}
}
