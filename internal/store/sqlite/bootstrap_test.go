package sqlite

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBootstrapCreatesTables(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "usage.db")
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if _, err := os.Stat(db.Path); err != nil {
		t.Fatalf("stat db: %v", err)
	}
	version, err := db.QueryInt("PRAGMA user_version;")
	if err != nil {
		t.Fatalf("QueryInt: %v", err)
	}
	if version != schemaVersion {
		t.Fatalf("version=%d", version)
	}
}
