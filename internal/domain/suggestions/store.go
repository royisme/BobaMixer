// Package suggestions provides storage for managing suggestion status
package suggestions

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/royisme/bobamixer/internal/store/sqlite"
)

// Status represents the status of a suggestion
type Status string

const (
	StatusNew     Status = "new"
	StatusApplied Status = "applied"
	StatusIgnored Status = "ignored"
	StatusSnoozed Status = "snoozed"
)

// StoredSuggestion represents a suggestion stored in the database
type StoredSuggestion struct {
	ID              string
	CreatedAt       time.Time
	SuggestionType  string
	Title           string
	Description     string
	ActionCmd       string
	Status          Status
	UntilTS         *time.Time // For snoozed suggestions
	Context         string     // JSON or text context
}

// Store manages suggestion persistence
type Store struct {
	db *sqlite.DB
}

// NewStore creates a new suggestion store
func NewStore(db *sqlite.DB) *Store {
	return &Store{db: db}
}

// Save saves a suggestion to the database
func (s *Store) Save(sugg *StoredSuggestion) error {
	if sugg.ID == "" {
		sugg.ID = fmt.Sprintf("sugg_%d", time.Now().UnixNano())
	}

	var untilTS int64
	if sugg.UntilTS != nil {
		untilTS = sugg.UntilTS.Unix()
	}

	query := fmt.Sprintf(`INSERT OR REPLACE INTO suggestions
		(id, created_at, suggestion_type, title, description, action_cmd, status, until_ts, context)
		VALUES ('%s', %d, '%s', '%s', '%s', '%s', '%s', %d, '%s');`,
		sugg.ID,
		sugg.CreatedAt.Unix(),
		sugg.SuggestionType,
		escapeSQLString(sugg.Title),
		escapeSQLString(sugg.Description),
		escapeSQLString(sugg.ActionCmd),
		string(sugg.Status),
		untilTS,
		escapeSQLString(sugg.Context),
	)

	return s.db.Exec(query)
}

// GetActive returns all active (non-ignored, non-snoozed) suggestions
func (s *Store) GetActive() ([]*StoredSuggestion, error) {
	now := time.Now().Unix()
	query := fmt.Sprintf(`SELECT id, created_at, suggestion_type, title, description, action_cmd, status, until_ts, context
		FROM suggestions
		WHERE status IN ('new', 'snoozed')
		AND (until_ts IS NULL OR until_ts = 0 OR until_ts < %d)
		ORDER BY created_at DESC;`, now)

	rows, err := s.db.QueryRows(query)
	if err != nil {
		return nil, err
	}

	return s.parseRows(rows), nil
}

// UpdateStatus updates the status of a suggestion
func (s *Store) UpdateStatus(id string, status Status, untilTS *time.Time) error {
	var untilVal int64
	if untilTS != nil {
		untilVal = untilTS.Unix()
	}

	query := fmt.Sprintf(`UPDATE suggestions
		SET status = '%s', until_ts = %d
		WHERE id = '%s';`,
		string(status),
		untilVal,
		id,
	)

	return s.db.Exec(query)
}

// Apply marks a suggestion as applied
func (s *Store) Apply(id string) error {
	return s.UpdateStatus(id, StatusApplied, nil)
}

// Ignore marks a suggestion as ignored
func (s *Store) Ignore(id string) error {
	return s.UpdateStatus(id, StatusIgnored, nil)
}

// Snooze snoozes a suggestion until a specific time
func (s *Store) Snooze(id string, duration time.Duration) error {
	until := time.Now().Add(duration)
	return s.UpdateStatus(id, StatusSnoozed, &until)
}

// parseRows parses database rows into StoredSuggestion objects
func (s *Store) parseRows(rows []string) []*StoredSuggestion {
	var result []*StoredSuggestion
	for _, row := range rows {
		// Parse pipe-delimited row
		// Format: id|created_at|type|title|description|action_cmd|status|until_ts|context
		parts := strings.Split(row, "|")
		if len(parts) < 9 {
			continue // Skip invalid rows
		}

		// Parse timestamps (Unix timestamp integers)
		var createdAt time.Time
		if ts, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
			createdAt = time.Unix(ts, 0)
		}

		var untilTS *time.Time
		if parts[7] != "" && parts[7] != "0" {
			if ts, err := strconv.ParseInt(parts[7], 10, 64); err == nil {
				t := time.Unix(ts, 0)
				untilTS = &t
			}
		}

		result = append(result, &StoredSuggestion{
			ID:             parts[0],
			CreatedAt:      createdAt,
			SuggestionType: parts[2],
			Title:          parts[3],
			Description:    parts[4],
			ActionCmd:      parts[5],
			Status:         Status(parts[6]),
			UntilTS:        untilTS,
			Context:        parts[8],
		})
	}
	return result
}

// escapeSQLString escapes single quotes in SQL strings
func escapeSQLString(s string) string {
	// Simple escape - replace ' with ''
	result := ""
	for _, ch := range s {
		if ch == '\'' {
			result += "''"
		} else {
			result += string(ch)
		}
	}
	return result
}
