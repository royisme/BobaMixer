// Package sqlitecli registers a database/sql driver backed by the sqlite3 binary.
package sqlitecli

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os/exec"
	"strings"
)

var (
	sqlitePath string
	lookupErr  error
)

func init() {
	sqlitePath, lookupErr = exec.LookPath("sqlite3")
	sql.Register("sqlitecli", &Driver{})
}

// Driver implements database/sql/driver.Driver using the sqlite3 CLI.
type Driver struct{}

// Open opens a CLI-backed connection.
func (d *Driver) Open(name string) (driver.Conn, error) {
	if lookupErr != nil {
		return nil, fmt.Errorf("locate sqlite3: %w", lookupErr)
	}
	path, pragmas, err := parseDSN(name)
	if err != nil {
		return nil, err
	}
	if path == "" {
		return nil, errors.New("empty sqlite path")
	}
	conn := &Conn{path: path}
	if err := conn.applyPragmas(pragmas); err != nil {
		return nil, err
	}
	return conn, nil
}

// Conn wraps sqlite3 operations.
type Conn struct {
	path string
}

var (
	_ driver.Conn              = (*Conn)(nil)
	_ driver.ExecerContext     = (*Conn)(nil)
	_ driver.QueryerContext    = (*Conn)(nil)
	_ driver.Pinger            = (*Conn)(nil)
	_ driver.SessionResetter   = (*Conn)(nil)
	_ driver.NamedValueChecker = (*Conn)(nil)
)

// Prepare implements driver.Conn.
func (c *Conn) Prepare(query string) (driver.Stmt, error) {
	return &Stmt{conn: c, query: query}, nil
}

// Close implements driver.Conn.
func (c *Conn) Close() error { return nil }

// Begin implements driver.Conn.
func (c *Conn) Begin() (driver.Tx, error) {
	return nil, errors.New("transactions not supported")
}

// ExecContext executes the SQL via sqlite3.
func (c *Conn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	if len(args) > 0 {
		return nil, fmt.Errorf("parameters not supported: %d", len(args))
	}
	if err := c.exec(ctx, query); err != nil {
		return nil, err
	}
	return driver.RowsAffected(0), nil
}

// QueryContext runs the SQL via sqlite3 and parses results.
func (c *Conn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	if len(args) > 0 {
		return nil, fmt.Errorf("parameters not supported: %d", len(args))
	}
	cols, data, err := c.query(ctx, query)
	if err != nil {
		return nil, err
	}
	return &Rows{columns: cols, data: data}, nil
}

// Ping verifies connectivity by reading user_version.
func (c *Conn) Ping(ctx context.Context) error {
	_, _, err := c.query(ctx, "PRAGMA user_version;")
	return err
}

// ResetSession is a no-op for CLI connections.
func (c *Conn) ResetSession(context.Context) error {
	return nil
}

// CheckNamedValue rejects positional parameters.
func (c *Conn) CheckNamedValue(value *driver.NamedValue) error {
	return fmt.Errorf("named values not supported: %v", value.Name)
}

// Stmt wraps a prepared SQL statement.
type Stmt struct {
	conn  *Conn
	query string
}

var (
	_ driver.Stmt             = (*Stmt)(nil)
	_ driver.StmtExecContext  = (*Stmt)(nil)
	_ driver.StmtQueryContext = (*Stmt)(nil)
)

// Close releases the statement.
func (s *Stmt) Close() error { return nil }

// NumInput indicates no positional parameters.
func (s *Stmt) NumInput() int { return 0 }

// ExecContext executes the statement.
func (s *Stmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	return s.conn.ExecContext(ctx, s.query, args)
}

// QueryContext queries the statement.
func (s *Stmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	return s.conn.QueryContext(ctx, s.query, args)
}

// Exec executes using background context.
func (s *Stmt) Exec(args []driver.Value) (driver.Result, error) {
	if len(args) > 0 {
		return nil, fmt.Errorf("parameters not supported: %d", len(args))
	}
	return s.ExecContext(context.Background(), nil)
}

// Query executes using background context.
func (s *Stmt) Query(args []driver.Value) (driver.Rows, error) {
	if len(args) > 0 {
		return nil, fmt.Errorf("parameters not supported: %d", len(args))
	}
	return s.QueryContext(context.Background(), nil)
}

// Rows provides iterator semantics for sqlite CLI output.
type Rows struct {
	columns []string
	data    [][]string
	idx     int
}

var _ driver.Rows = (*Rows)(nil)

// Columns returns the column names.
func (r *Rows) Columns() []string { return r.columns }

// Close releases rows.
func (r *Rows) Close() error { return nil }

// Next populates the destination slice.
func (r *Rows) Next(dest []driver.Value) error {
	if r.idx >= len(r.data) {
		return io.EOF
	}
	row := r.data[r.idx]
	r.idx++
	for i := range dest {
		if i < len(row) {
			dest[i] = row[i]
		} else {
			dest[i] = nil
		}
	}
	return nil
}

func (c *Conn) exec(ctx context.Context, query string) error {
	sql := normalizeSQL(query)
	if sql == "" {
		return nil
	}
	cmd := exec.CommandContext(ctx, sqlitePath, c.path, sql) //nolint:gosec // executes sqlite3 binary with validated path
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("sqlite3 exec: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func (c *Conn) query(ctx context.Context, query string) ([]string, [][]string, error) {
	sql := normalizeSQL(query)
	if sql == "" {
		return []string{}, [][]string{}, nil
	}
	cmd := exec.CommandContext(ctx, sqlitePath, "-header", "-separator", "|", c.path, sql) //nolint:gosec // executes sqlite3 binary with validated path
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, nil, fmt.Errorf("sqlite3 query: %w: %s", err, strings.TrimSpace(string(out)))
	}
	text := strings.TrimSpace(string(out))
	if text == "" {
		return []string{}, [][]string{}, nil
	}
	lines := strings.Split(text, "\n")
	columns := strings.Split(lines[0], "|")
	rows := make([][]string, 0, len(lines)-1)
	for _, line := range lines[1:] {
		if line == "" {
			continue
		}
		rows = append(rows, strings.Split(line, "|"))
	}
	return columns, rows, nil
}

func (c *Conn) applyPragmas(pragmas []string) error {
	for _, pragma := range pragmas {
		stmt := fmt.Sprintf("PRAGMA %s;", pragma)
		if err := c.exec(context.Background(), stmt); err != nil {
			return err
		}
	}
	return nil
}

func parseDSN(dsn string) (string, []string, error) {
	if strings.HasPrefix(dsn, "file:") {
		u, err := url.Parse(dsn)
		if err != nil {
			return "", nil, fmt.Errorf("parse sqlite dsn: %w", err)
		}
		return u.Path, u.Query()["_pragma"], nil
	}
	return dsn, nil, nil
}

func normalizeSQL(query string) string {
	trimmed := strings.TrimSpace(query)
	trimmed = strings.ReplaceAll(trimmed, "\n", " ")
	trimmed = strings.ReplaceAll(trimmed, "\t", " ")
	trimmed = strings.Join(strings.Fields(trimmed), " ")
	return trimmed
}
