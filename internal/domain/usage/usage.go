package usage

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/royisme/bobamixer/internal/adapters"
	"github.com/royisme/bobamixer/internal/domain/pricing"
	"github.com/royisme/bobamixer/internal/store/config"
	"github.com/royisme/bobamixer/internal/store/sqlite"
)

// Record represents a usage record
type Record struct {
	ID           string
	SessionID    string
	Timestamp    int64
	InputTokens  int
	OutputTokens int
	InputCost    float64
	OutputCost   float64
	Tool         string
	Model        string
}

// NewRecord creates a new usage record
func NewRecord(sessionID, tool, model string, result adapters.Result, pricingTable *pricing.Table, profileCost config.Cost) *Record {
	inputCost, outputCost := pricingTable.CalculateCost(model, profileCost, result.Usage.InputTokens, result.Usage.OutputTokens)

	return &Record{
		ID:           generateID(),
		SessionID:    sessionID,
		Timestamp:    time.Now().Unix(),
		InputTokens:  result.Usage.InputTokens,
		OutputTokens: result.Usage.OutputTokens,
		InputCost:    inputCost,
		OutputCost:   outputCost,
		Tool:         tool,
		Model:        model,
	}
}

// generateID generates a random record ID
func generateID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// rand.Read should never fail with crypto/rand, but handle it anyway
		panic(fmt.Sprintf("failed to generate random ID: %v", err))
	}
	return hex.EncodeToString(b)
}

// Save saves the usage record to database
func (r *Record) Save(db *sqlite.DB) error {
	query := fmt.Sprintf(`
		INSERT INTO usage_records
		(id, session_id, ts, input_tokens, output_tokens, input_cost, output_cost, tool, model)
		VALUES ('%s', '%s', %d, %d, %d, %f, %f, '%s', '%s');
	`,
		r.ID, r.SessionID, r.Timestamp,
		r.InputTokens, r.OutputTokens,
		r.InputCost, r.OutputCost,
		escape(r.Tool), escape(r.Model))

	return db.Exec(query)
}

// Stats represents usage statistics
type Stats struct {
	TotalTokens int
	TotalCost   float64
	Sessions    int
	AvgLatency  float64
}

// GetTodayStats returns today's usage statistics
func GetTodayStats(db *sqlite.DB) (*Stats, error) {
	query := `
		SELECT
			COALESCE(SUM(input_tokens + output_tokens), 0) as total_tokens,
			COALESCE(SUM(input_cost + output_cost), 0) as total_cost,
			COUNT(DISTINCT session_id) as sessions
		FROM usage_records
		WHERE date(ts, 'unixepoch') = date('now');
	`

	row, err := db.QueryRow(query)
	if err != nil {
		return &Stats{}, nil
	}

	// Parse row (simplified)
	stats := &Stats{}
	if _, err := fmt.Sscanf(row, "%d|%f|%d", &stats.TotalTokens, &stats.TotalCost, &stats.Sessions); err != nil {
		return &Stats{}, fmt.Errorf("failed to parse stats: %w", err)
	}

	return stats, nil
}

// GetPeriodStats returns statistics for a time period
func GetPeriodStats(db *sqlite.DB, days int) (*Stats, error) {
	query := fmt.Sprintf(`
		SELECT
			COALESCE(SUM(input_tokens + output_tokens), 0) as total_tokens,
			COALESCE(SUM(input_cost + output_cost), 0) as total_cost,
			COUNT(DISTINCT session_id) as sessions
		FROM usage_records
		WHERE date(ts, 'unixepoch') >= date('now', '-%d days');
	`, days)

	row, err := db.QueryRow(query)
	if err != nil {
		return &Stats{}, nil
	}

	stats := &Stats{}
	fmt.Sscanf(row, "%d|%f|%d", &stats.TotalTokens, &stats.TotalCost, &stats.Sessions)

	return stats, nil
}

func escape(s string) string {
	// Simple SQL escape - in production use parameterized queries
	return s
}
