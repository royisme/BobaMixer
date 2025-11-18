// Package root provides the root UI model and orchestration for the BobaMixer TUI.
package root

import (
	"context"
	"net/http"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/royisme/bobamixer/internal/domain/stats"
	"github.com/royisme/bobamixer/internal/domain/suggestions"
	"github.com/royisme/bobamixer/internal/proxy"
)

// proxyStatusMsg is sent when proxy status is checked
type proxyStatusMsg struct {
	running bool
}

// statsLoadedMsg is sent when stats are loaded
type statsLoadedMsg struct {
	today        stats.Summary
	week         stats.Summary
	profileStats []stats.ProfileStats
	err          error
}

// suggestionsLoadedMsg is sent when suggestions are loaded
type suggestionsLoadedMsg struct {
	suggestions []suggestions.Suggestion
	err         error
}

// checkProxyStatus checks if the proxy server is running
func checkProxyStatus() tea.Msg {
	addr := proxy.DefaultAddr
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+addr+"/health", nil)
	if err != nil {
		return proxyStatusMsg{running: false}
	}

	client := &http.Client{Timeout: 500 * time.Millisecond}
	resp, err := client.Do(req)
	if err != nil {
		return proxyStatusMsg{running: false}
	}
	defer func() {
		// Close response body; error ignored as it doesn't affect proxy status check
		//nolint:errcheck,gosec // Error on close is not critical for status check
		resp.Body.Close()
	}()

	return proxyStatusMsg{running: resp.StatusCode == http.StatusOK}
}

// loadStatsData loads usage statistics from the database
func (m *DashboardModel) loadStatsData() tea.Msg {
	data, err := m.statsService.LoadData()
	if err != nil {
		return statsLoadedMsg{err: err}
	}

	return statsLoadedMsg{
		today:        data.Today,
		week:         data.Week,
		profileStats: data.ProfileStats,
	}
}

// loadSuggestions loads optimization suggestions
func (m *DashboardModel) loadSuggestions() tea.Msg {
	suggs, err := m.suggestionsService.LoadData(7)
	if err != nil {
		return suggestionsLoadedMsg{err: err}
	}

	return suggestionsLoadedMsg{suggestions: suggs}
}
