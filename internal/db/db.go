// Package db provides SQLite helpers for bootstrapping and upgrading the metrics store.
package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/royisme/bobamixer/internal/db/sqlitecli"

	"github.com/royisme/bobamixer/internal/bobaerrors"
)

const targetVersion = 2

// Open ensures the directory exists and opens a SQLite database with sane defaults.
func Open(path string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("create db directory: %w", err)
	}
	u := &url.URL{Scheme: "file", Path: path}
	dsn := WithPragmas(u.String())
	database, err := sql.Open("sqlitecli", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}
	database.SetMaxOpenConns(1)
	return database, nil
}

// Bootstrap ensures tables, indices, and the required user_version exist.
func Bootstrap(ctx context.Context, database *sql.DB) error {
	if database == nil {
		return fmt.Errorf("bootstrap nil db: %w", bobaerrors.ErrDB)
	}
	if err := applyBaseSchema(ctx, database); err != nil {
		return err
	}
	version, err := readUserVersion(ctx, database)
	if err != nil {
		return err
	}
	switch {
	case version == 0:
		if err := setUserVersion(ctx, database, targetVersion); err != nil {
			return err
		}
	case version > targetVersion:
		return fmt.Errorf("database version %d newer than supported %d: %w", version, targetVersion, bobaerrors.ErrDB)
	case version < targetVersion:
		if err := EnsureUpgrades(ctx, database, version, targetVersion); err != nil {
			return err
		}
	}
	return nil
}

// EnsureUpgrades migrates user_version forward, ensuring new columns exist.
func EnsureUpgrades(ctx context.Context, database *sql.DB, from, to int) error {
	if database == nil {
		return fmt.Errorf("upgrade nil db: %w", bobaerrors.ErrDB)
	}
	if from > to {
		return fmt.Errorf("invalid upgrade path %d -> %d: %w", from, to, bobaerrors.ErrDB)
	}
	current := from
	for current < to {
		next := current + 1
		switch next {
		case 2:
			if err := addEstimateLevelColumn(ctx, database); err != nil {
				return err
			}
			if err := ensureUsageIndex(ctx, database); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown upgrade %d -> %d: %w", current, next, bobaerrors.ErrDB)
		}
		current = next
	}
	return setUserVersion(ctx, database, to)
}

// WithPragmas appends recommended pragmas to the provided DSN.
func WithPragmas(dsn string) string {
	separator := "?"
	if strings.Contains(dsn, "?") {
		separator = "&"
	}
	return dsn + separator + "_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(ON)"
}

func applyBaseSchema(ctx context.Context, database *sql.DB) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS sessions (
                        id TEXT PRIMARY KEY,
                        source TEXT NOT NULL,
                        started_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
                        ended_at DATETIME,
                        success INTEGER NOT NULL DEFAULT 0,
                        notes TEXT
                )`,
		`CREATE TABLE IF NOT EXISTS usage_records (
                        id INTEGER PRIMARY KEY AUTOINCREMENT,
                        session_id TEXT NOT NULL,
                        profile TEXT,
                        provider TEXT,
                        model TEXT,
                        input_tokens INTEGER NOT NULL DEFAULT 0,
                        output_tokens INTEGER NOT NULL DEFAULT 0,
                        cost_usd REAL NOT NULL DEFAULT 0,
                        latency_ms INTEGER NOT NULL DEFAULT 0,
                        estimate_level TEXT NOT NULL DEFAULT 'exact',
                        created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
                        FOREIGN KEY (session_id) REFERENCES sessions(id)
                )`,
		`CREATE VIEW IF NOT EXISTS usage_daily_summary AS
                        SELECT date(created_at) AS usage_date,
                               SUM(input_tokens) AS total_input_tokens,
                               SUM(output_tokens) AS total_output_tokens,
                               SUM(cost_usd) AS total_cost_usd,
                               COUNT(DISTINCT session_id) AS sessions
                        FROM usage_records
                        GROUP BY date(created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_usage_ts ON usage_records(created_at)`,
	}
	for _, stmt := range statements {
		if _, err := database.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("apply schema: %w", err)
		}
	}
	return nil
}

func addEstimateLevelColumn(ctx context.Context, database *sql.DB) error {
	exists, err := columnExists(ctx, database, "usage_records", "estimate_level")
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	if _, err := database.ExecContext(ctx, `ALTER TABLE usage_records ADD COLUMN estimate_level TEXT NOT NULL DEFAULT 'exact'`); err != nil {
		return fmt.Errorf("add estimate_level: %w", err)
	}
	return nil
}

func ensureUsageIndex(ctx context.Context, database *sql.DB) error {
	_, err := database.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_usage_ts ON usage_records(created_at)`)
	if err != nil {
		return fmt.Errorf("ensure idx_usage_ts: %w", err)
	}
	return nil
}

func columnExists(ctx context.Context, database *sql.DB, table, column string) (_ bool, err error) {
	rows, err := database.QueryContext(ctx, `PRAGMA table_info(`+table+`)`)
	if err != nil {
		return false, fmt.Errorf("table_info %s: %w", table, err)
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil {
			err = errors.Join(err, fmt.Errorf("close table_info rows: %w", cerr))
		}
	}()
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return false, fmt.Errorf("scan table_info: %w", err)
		}
		if name == column {
			return true, nil
		}
	}
	if err := rows.Err(); err != nil {
		return false, fmt.Errorf("iterate table_info: %w", err)
	}
	return false, nil
}

func readUserVersion(ctx context.Context, database *sql.DB) (int, error) {
	row := database.QueryRowContext(ctx, `PRAGMA user_version`)
	var version int
	if err := row.Scan(&version); err != nil {
		return 0, fmt.Errorf("read user_version: %w", err)
	}
	return version, nil
}

func setUserVersion(ctx context.Context, database *sql.DB, version int) error {
	if _, err := database.ExecContext(ctx, fmt.Sprintf("PRAGMA user_version = %d", version)); err != nil {
		return fmt.Errorf("set user_version: %w", err)
	}
	return nil
}
