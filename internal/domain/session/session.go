// Package session manages execution sessions and their lifecycle.
package session

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/royisme/bobamixer/internal/store/sqlite"
)

// Session represents a single execution session
type Session struct {
	ID        string
	StartedAt int64
	EndedAt   int64
	Project   string
	Branch    string
	Profile   string
	Adapter   string
	TaskType  string
	Success   bool
	LatencyMS int64
	Notes     string
}

// NewSession creates a new session
func NewSession(profile, adapter string) *Session {
	return &Session{
		ID:        generateID(),
		StartedAt: time.Now().Unix(),
		Profile:   profile,
		Adapter:   adapter,
	}
}

// generateID generates a random session ID
func generateID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// rand.Read should never fail with crypto/rand, but handle it anyway
		panic(fmt.Sprintf("failed to generate random session ID: %v", err))
	}
	return hex.EncodeToString(b)
}

// End marks the session as ended
func (s *Session) End(success bool, latencyMS int64) {
	s.EndedAt = time.Now().Unix()
	s.Success = success
	s.LatencyMS = latencyMS
}

// Save saves the session to database
func (s *Session) Save(db *sqlite.DB) error {
	query := fmt.Sprintf(`
		INSERT OR REPLACE INTO sessions
		(id, started_at, ended_at, project, branch, profile, adapter, task_type, success, latency_ms, notes)
		VALUES ('%s', %d, %d, '%s', '%s', '%s', '%s', '%s', %d, %d, '%s');
	`,
		s.ID, s.StartedAt, s.EndedAt,
		escape(s.Project), escape(s.Branch), escape(s.Profile), escape(s.Adapter), escape(s.TaskType),
		boolToInt(s.Success), s.LatencyMS, escape(s.Notes))

	return db.Exec(query)
}

// GetSession retrieves a session by ID
func GetSession(db *sqlite.DB, id string) (*Session, error) {
	query := fmt.Sprintf("SELECT id, started_at, ended_at, project, branch, profile, adapter, task_type, success, latency_ms, notes FROM sessions WHERE id='%s';", escape(id))
	row, err := db.QueryRow(query)
	if err != nil {
		return nil, err
	}
	if row == "" {
		return nil, fmt.Errorf("session %s not found", id)
	}
	parts := strings.Split(row, "|")
	if len(parts) < 11 {
		return nil, fmt.Errorf("unexpected session row: %s", row)
	}
	sess := &Session{ID: parts[0]}
	sess.StartedAt = parseInt64(parts[1])
	sess.EndedAt = parseInt64(parts[2])
	sess.Project = parts[3]
	sess.Branch = parts[4]
	sess.Profile = parts[5]
	sess.Adapter = parts[6]
	sess.TaskType = parts[7]
	sess.Success = parts[8] == "1"
	sess.LatencyMS = parseInt64(parts[9])
	sess.Notes = parts[10]
	return sess, nil
}

// ListRecentSessions returns recent sessions
func ListRecentSessions(db *sqlite.DB, limit int) ([]*Session, error) {
	query := fmt.Sprintf("SELECT id, started_at, ended_at, profile, adapter, success, latency_ms, task_type FROM sessions ORDER BY started_at DESC LIMIT %d;", limit)
	rows, err := db.QueryRows(query)
	if err != nil {
		return nil, err
	}
	sessions := make([]*Session, 0, len(rows))
	for _, row := range rows {
		if row == "" {
			continue
		}
		parts := strings.Split(row, "|")
		if len(parts) < 8 {
			continue
		}
		sess := &Session{ID: parts[0]}
		sess.StartedAt = parseInt64(parts[1])
		sess.EndedAt = parseInt64(parts[2])
		sess.Profile = parts[3]
		sess.Adapter = parts[4]
		sess.Success = parts[5] == "1"
		sess.LatencyMS = parseInt64(parts[6])
		sess.TaskType = parts[7]
		sessions = append(sessions, sess)
	}
	return sessions, nil
}

func escape(s string) string {
	// Simple SQL escape - in production use parameterized queries
	return s
}

func parseInt64(raw string) int64 {
	v, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil {
		return 0
	}
	return v
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
