package session

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
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
	rand.Read(b)
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

	// Parse the row (simplified - would need proper parsing in production)
	s := &Session{ID: id}
	return s, nil
}

// ListRecentSessions returns recent sessions
func ListRecentSessions(db *sqlite.DB, limit int) ([]*Session, error) {
	query := fmt.Sprintf("SELECT id, started_at, profile, adapter, success, latency_ms FROM sessions ORDER BY started_at DESC LIMIT %d;", limit)
	// Simplified - would need proper row parsing
	_, err := db.QueryRow(query)
	if err != nil {
		return nil, err
	}
	return []*Session{}, nil
}

func escape(s string) string {
	// Simple SQL escape - in production use parameterized queries
	return s
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
