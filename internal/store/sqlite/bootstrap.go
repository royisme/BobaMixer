// Package sqlite provides SQLite database connection and schema management.
package sqlite

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

const schemaVersion = 2

type DB struct {
	Path string
}

func Open(path string) (*DB, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	db := &DB{Path: abs}
	if err := db.ensureFile(); err != nil {
		return nil, err
	}
	if err := db.bootstrap(); err != nil {
		return nil, err
	}
	return db, nil
}

func (db *DB) ensureFile() error {
	// #nosec G204 -- db.Path is from safe home directory structure
	cmd := exec.CommandContext(context.Background(), "sqlite3", db.Path, "PRAGMA journal_mode=WAL;")
	return cmd.Run()
}

func (db *DB) Exec(query string) error {
	// #nosec G204 -- db.Path is from safe home directory structure
	cmd := exec.CommandContext(context.Background(), "sqlite3", db.Path, query)
	return cmd.Run()
}

func (db *DB) QueryRow(query string) (string, error) {
	rows, err := db.QueryRows(query)
	if err != nil {
		return "", err
	}
	if len(rows) == 0 {
		return "", nil
	}
	return rows[0], nil
}

// QueryRows executes a query and returns each row as a raw pipe-delimited string.
func (db *DB) QueryRows(query string) ([]string, error) {
	// #nosec G204 -- db.Path is from safe home directory structure
	cmd := exec.CommandContext(context.Background(), "sqlite3", db.Path, query)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	trimmed := strings.TrimSpace(string(out))
	if trimmed == "" {
		return []string{}, nil
	}
	parts := strings.Split(trimmed, "\n")
	return parts, nil
}

func (db *DB) QueryInt(query string) (int, error) {
	out, err := db.QueryRow(query)
	if err != nil {
		return 0, err
	}
	var v int
	_, err = fmt.Sscanf(out, "%d", &v)
	return v, err
}

func (db *DB) bootstrap() error {
	version, err := db.QueryInt("PRAGMA user_version;")
	if err != nil {
		return err
	}
	if version >= schemaVersion {
		return nil
	}

	// Version 0 -> 1: Initial schema
	if version == 0 {
		if err := db.migrateToV1(); err != nil {
			return fmt.Errorf("migrate to v1: %w", err)
		}
		version = 1
	}

	// Version 1 -> 2: Add estimate_level to usage_records
	if version == 1 {
		if err := db.migrateToV2(); err != nil {
			return fmt.Errorf("migrate to v2: %w", err)
		}
		version = 2
	}

	return nil
}

func (db *DB) migrateToV1() error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS sessions (
            id TEXT PRIMARY KEY,
            started_at INTEGER NOT NULL,
            ended_at INTEGER,
            project TEXT,
            branch TEXT,
            profile TEXT,
            adapter TEXT,
            task_type TEXT,
            success INTEGER,
            latency_ms INTEGER,
            notes TEXT
        );`,
		`CREATE TABLE IF NOT EXISTS usage_records (
            id TEXT PRIMARY KEY,
            session_id TEXT NOT NULL,
            ts INTEGER NOT NULL,
            input_tokens INTEGER DEFAULT 0,
            output_tokens INTEGER DEFAULT 0,
            input_cost REAL DEFAULT 0,
            output_cost REAL DEFAULT 0,
            tool TEXT,
            model TEXT,
            estimate_level TEXT NOT NULL DEFAULT 'heuristic' CHECK(estimate_level IN ('exact','mapped','heuristic')),
            FOREIGN KEY(session_id) REFERENCES sessions(id) ON DELETE CASCADE
        );`,
		`CREATE TABLE IF NOT EXISTS budgets (
            id TEXT PRIMARY KEY,
            scope TEXT NOT NULL,
            target TEXT,
            daily_usd REAL,
            hard_cap REAL,
            period_start INTEGER,
            period_end INTEGER,
            spent_usd REAL DEFAULT 0
        );`,
		`CREATE VIEW IF NOT EXISTS v_daily_summary AS
            SELECT date(ts, 'unixepoch') AS date,
                   SUM(input_tokens + output_tokens) AS total_tokens,
                   SUM(input_cost + output_cost) AS total_cost
            FROM usage_records GROUP BY date;`,
		"PRAGMA user_version = 1;",
	}
	for _, stmt := range statements {
		if err := db.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

func (db *DB) migrateToV2() error {
	// Add estimate_level column to existing usage_records table
	statements := []string{
		`ALTER TABLE usage_records ADD COLUMN estimate_level TEXT NOT NULL DEFAULT 'heuristic' CHECK(estimate_level IN ('exact','mapped','heuristic'));`,
		"PRAGMA user_version = 2;",
	}
	for _, stmt := range statements {
		if err := db.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}
