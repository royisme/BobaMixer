package db

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
)

func TestBootstrapCreatesSchema(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	database := openTestDB(t)
	if err := Bootstrap(ctx, database); err != nil {
		t.Fatalf("Bootstrap() error = %v", err)
	}

	assertUserVersion(t, database, 2)
	ensureColumnExists(t, database, "usage_records", "estimate_level")
	ensureIndexExists(t, database, "usage_records", "idx_usage_ts")
}

func TestEnsureUpgradesAddsEstimateLevel(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	database := openTestDB(t)

	_, err := database.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS usage_records (
id INTEGER PRIMARY KEY AUTOINCREMENT,
session_id TEXT NOT NULL,
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
)`)
	if err != nil {
		t.Fatalf("create table error = %v", err)
	}
	if _, err := database.ExecContext(ctx, `PRAGMA user_version = 1`); err != nil {
		t.Fatalf("set version error = %v", err)
	}

	if err := EnsureUpgrades(ctx, database, 1, 2); err != nil {
		t.Fatalf("EnsureUpgrades() error = %v", err)
	}
	ensureColumnExists(t, database, "usage_records", "estimate_level")
	assertUserVersion(t, database, 2)

	if err := EnsureUpgrades(ctx, database, 2, 2); err != nil {
		t.Fatalf("EnsureUpgrades() second call error = %v", err)
	}
}

func TestWithPragmasAppendsParams(t *testing.T) {
	t.Parallel()

	got := WithPragmas("file:test.db")
	want := "file:test.db?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(ON)"
	if got != want {
		t.Fatalf("WithPragmas() = %q, want %q", got, want)
	}
}

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")
	database, err := Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := database.Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
	})
	return database
}

func ensureColumnExists(t *testing.T, database *sql.DB, table, column string) {
	t.Helper()

	ctx := context.Background()
	rows, err := database.QueryContext(ctx, `PRAGMA table_info(`+table+`)`)
	if err != nil {
		t.Fatalf("table_info error = %v", err)
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil {
			t.Fatalf("Close() error = %v", cerr)
		}
	}()

	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			t.Fatalf("scan error = %v", err)
		}
		if name == column {
			return
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows error = %v", err)
	}
	t.Fatalf("column %s not found in %s", column, table)
}

func ensureIndexExists(t *testing.T, database *sql.DB, table, index string) {
	t.Helper()

	ctx := context.Background()
	rows, err := database.QueryContext(ctx, `PRAGMA index_list(`+table+`)`)
	if err != nil {
		t.Fatalf("index_list error = %v", err)
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil {
			t.Fatalf("Close() error = %v", cerr)
		}
	}()

	for rows.Next() {
		var seq, unique int
		var name string
		var origin sql.NullString
		var partial sql.NullString
		if err := rows.Scan(&seq, &name, &unique, &origin, &partial); err != nil {
			t.Fatalf("scan error = %v", err)
		}
		if name == index {
			return
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows error = %v", err)
	}
	t.Fatalf("index %s not found on %s", index, table)
}

func assertUserVersion(t *testing.T, database *sql.DB, want int) {
	t.Helper()

	ctx := context.Background()
	row := database.QueryRowContext(ctx, `PRAGMA user_version`)
	var got int
	if err := row.Scan(&got); err != nil {
		t.Fatalf("scan error = %v", err)
	}
	if got != want {
		t.Fatalf("user_version = %d, want %d", got, want)
	}
}
