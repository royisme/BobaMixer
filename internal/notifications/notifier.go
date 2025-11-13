// Package notifications manages real-time notifications and events for the user.
package notifications

import (
	"fmt"
	"sync"
	"time"

	"github.com/royisme/bobamixer/internal/domain/budget"
	"github.com/royisme/bobamixer/internal/domain/suggestions"
)

// Event represents a realtime notification that can be surfaced to the user.
type Event struct {
	Type      string
	Title     string
	Message   string
	Timestamp time.Time
	Metadata  map[string]string
}

// Notifier polls multiple subsystems to produce realtime events.
type Notifier struct {
	alerts     *budget.AlertManager
	suggEngine *suggestions.Engine
	seen       map[string]struct{}
	mu         sync.Mutex
}

// NewNotifier constructs a notifier bound to budget alerts and suggestion engine.
func NewNotifier(tracker *budget.Tracker, engine *suggestions.Engine, cfg *budget.AlertConfig) *Notifier {
	if cfg == nil {
		cfg = budget.DefaultAlertConfig()
	}
	return &Notifier{
		alerts:     budget.NewAlertManager(tracker, cfg),
		suggEngine: engine,
		seen:       map[string]struct{}{},
	}
}

// Poll inspects subsystems and returns newly detected events.
func (n *Notifier) Poll() ([]Event, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	var events []Event

	if n.alerts != nil {
		// Currently we only monitor the global scope; per-profile/project can be added later.
		alerts := n.alerts.CheckBudgetAlerts("global", "")
		for _, alert := range alerts {
			key := fmt.Sprintf("alert:%s:%s:%s", alert.Scope, alert.Target, alert.Title)
			if _, ok := n.seen[key]; ok {
				continue
			}
			n.seen[key] = struct{}{}
			events = append(events, Event{
				Type:      "budget_alert",
				Title:     alert.Title,
				Message:   alert.FormatAlert(),
				Timestamp: alert.Timestamp,
				Metadata: map[string]string{
					"scope":  alert.Scope,
					"target": alert.Target,
					"level":  alert.Level.String(),
				},
			})
		}
	}

	if n.suggEngine != nil {
		if suggs, err := n.suggEngine.GenerateSuggestions(3); err == nil {
			for _, sugg := range suggs {
				if sugg.Priority < 4 {
					continue
				}
				key := fmt.Sprintf("suggestion:%s:%s", sugg.Type.String(), sugg.Title)
				if _, ok := n.seen[key]; ok {
					continue
				}
				n.seen[key] = struct{}{}
				events = append(events, Event{
					Type:      "suggestion",
					Title:     sugg.Title,
					Message:   sugg.FormatSuggestion(),
					Timestamp: time.Now(),
					Metadata: map[string]string{
						"type":     sugg.Type.String(),
						"priority": sugg.GetPriority(),
					},
				})
			}
		}
	}

	return events, nil
}

// Clear resets the deduplication cache (primarily for tests).
func (n *Notifier) Clear() {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.seen = map[string]struct{}{}
}
