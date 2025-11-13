package hooks

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRecordEvent(t *testing.T) {
	dir := t.TempDir()
	mgr := NewManager(dir)
	if err := mgr.Record("post-commit", "/tmp/repo", "main"); err != nil {
		t.Fatalf("Record: %v", err)
	}
	files, err := os.ReadDir(filepath.Join(dir, "git-hooks"))
	if err != nil || len(files) == 0 {
		t.Fatalf("expected log file")
	}
}
