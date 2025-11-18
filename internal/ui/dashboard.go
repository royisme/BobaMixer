package ui

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/royisme/bobamixer/internal/domain/core"
	"github.com/royisme/bobamixer/internal/domain/stats"
	"github.com/royisme/bobamixer/internal/domain/suggestions"
	"github.com/royisme/bobamixer/internal/proxy"
	"github.com/royisme/bobamixer/internal/store/sqlite"
)

// viewMode represents the current view in the dashboard
type viewMode int

const (
	viewDashboard viewMode = iota
	viewProviders
	viewTools
	viewBindings
	viewSecrets
	viewStats
	viewProxy
	viewRouting
	viewSuggestions
)

// DashboardModel represents the control plane dashboard
type DashboardModel struct {
	home      string
	theme     Theme
	localizer *Localizer

	// Data
	providers *core.ProvidersConfig
	tools     *core.ToolsConfig
	bindings  *core.BindingsConfig
	secrets   *core.SecretsConfig

	// Stats data
	todayStats   stats.Summary
	weekStats    stats.Summary
	profileStats []stats.ProfileStats
	statsLoaded  bool
	statsError   string

	// Suggestions data
	suggestions      []suggestions.Suggestion
	suggestionsError string

	// Routing test input
	routingTestInput string

	// UI components
	table table.Model

	// State
	currentView   viewMode
	selectedIndex int    // Currently selected item in list views
	width         int
	height        int
	quitting      bool
	proxyStatus   string // "running", "stopped", "checking"
	message       string // Status message to display
}

// NewDashboard creates a new dashboard model
func NewDashboard(home string) (*DashboardModel, error) {
	// Load theme and localizer
	theme := loadTheme(home)
	localizer, err := NewLocalizer(GetUserLanguage())
	if err != nil {
		// Fallback to English if user language is not available
		localizer, err = NewLocalizer("en")
		if err != nil {
			// Should not happen with English, but handle it
			return nil, fmt.Errorf("failed to load localizer: %w", err)
		}
	}

	// Load all configurations
	providers, tools, bindings, secrets, err := core.LoadAll(home)
	if err != nil {
		return nil, fmt.Errorf("failed to load configurations: %w", err)
	}

	m := &DashboardModel{
		home:        home,
		theme:       theme,
		localizer:   localizer,
		providers:   providers,
		tools:       tools,
		bindings:    bindings,
		secrets:     secrets,
		proxyStatus: "checking",
		currentView: viewDashboard,
	}

	m.initializeTable()

	return m, nil
}

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
	dbPath := filepath.Join(m.home, "usage.db")
	db, err := sqlite.Open(dbPath)
	if err != nil {
		return statsLoadedMsg{err: err}
	}
	// Note: sqlite.DB uses CLI-based approach, no Close() needed

	ctx := context.Background()

	// Load today's stats
	today, err := stats.Today(ctx, db)
	if err != nil {
		return statsLoadedMsg{err: fmt.Errorf("load today stats: %w", err)}
	}

	// Load 7-day stats
	to := time.Now()
	from := to.AddDate(0, 0, -7)
	week, err := stats.Window(ctx, db, from, to)
	if err != nil {
		return statsLoadedMsg{err: fmt.Errorf("load week stats: %w", err)}
	}

	// Load profile stats
	analyzer := stats.NewAnalyzer(db)
	profileStats, err := analyzer.GetProfileStats(7)
	if err != nil {
		// Don't fail if profile stats can't be loaded
		profileStats = []stats.ProfileStats{}
	}

	return statsLoadedMsg{
		today:        today,
		week:         week,
		profileStats: profileStats,
	}
}

// initializeTable sets up the table with current data
func (m *DashboardModel) initializeTable() {
	columns := []table.Column{
		{Title: "Tool", Width: 12},
		{Title: "Provider", Width: 22},
		{Title: "Model", Width: 25},
		{Title: "Proxy", Width: 8},
		{Title: "Status", Width: 13},
	}

	rows := m.buildTableRows()

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	// Style the table
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(m.theme.Border).
		BorderBottom(true).
		Bold(true).
		Foreground(m.theme.Primary)

	s.Selected = s.Selected.
		Foreground(m.theme.Text).
		Background(m.theme.Primary).
		Bold(false)

	t.SetStyles(s)

	m.table = t
}

// buildTableRows creates table rows from current configuration
func (m *DashboardModel) buildTableRows() []table.Row {
	rows := make([]table.Row, 0)

	for _, tool := range m.tools.Tools {
		// Find binding for this tool
		binding, err := m.bindings.FindBinding(tool.ID)
		if err != nil {
			// No binding, show as not configured
			rows = append(rows, table.Row{
				tool.Name,
				"(not bound)",
				"-",
				"-",
				"⚠ Not configured",
			})
			continue
		}

		// Find provider
		provider, err := m.providers.FindProvider(binding.ProviderID)
		if err != nil {
			// Provider not found
			rows = append(rows, table.Row{
				tool.Name,
				fmt.Sprintf("(missing: %s)", binding.ProviderID),
				"-",
				"-",
				"❌ Error",
			})
			continue
		}

		// Check API key status
		keyStatus := "✓ Ready"
		if _, err := core.ResolveAPIKey(provider, m.secrets); err != nil {
			keyStatus = "⚠ No API key"
		}

		// Determine model
		model := provider.DefaultModel
		if binding.Options.Model != "" {
			model = binding.Options.Model
		}

		// Truncate if too long
		if len(model) > 23 {
			model = model[:20] + "..."
		}

		// Proxy status
		proxyStatus := "OFF"
		if binding.UseProxy {
			proxyStatus = "ON"
		}

		rows = append(rows, table.Row{
			tool.Name,
			provider.DisplayName,
			model,
			proxyStatus,
			keyStatus,
		})
	}

	if len(rows) == 0 {
		rows = append(rows, table.Row{
			"No tools configured",
			"-",
			"-",
			"-",
			"-",
		})
	}

	return rows
}

// Init initializes the dashboard
func (m DashboardModel) Init() tea.Cmd {
	// Check proxy status on startup and load stats
	return tea.Batch(
		checkProxyStatus,
		m.loadStatsData,
	)
}

// Update handles messages
//nolint:gocyclo // UI event handlers are inherently complex
func (m DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case proxyStatusMsg:
		// Update proxy status based on check
		if msg.running {
			m.proxyStatus = "running"
		} else {
			m.proxyStatus = "stopped"
		}
		return m, nil

	case statsLoadedMsg:
		// Update stats data
		if msg.err != nil {
			m.statsError = msg.err.Error()
		} else {
			m.todayStats = msg.today
			m.weekStats = msg.week
			m.profileStats = msg.profileStats
			m.statsLoaded = true
			m.statsError = ""
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "1":
			m.currentView = viewDashboard
			m.selectedIndex = 0
			return m, nil

		case "2":
			m.currentView = viewProviders
			m.selectedIndex = 0
			return m, nil

		case "3":
			m.currentView = viewTools
			m.selectedIndex = 0
			return m, nil

		case "4":
			m.currentView = viewBindings
			m.selectedIndex = 0
			return m, nil

		case "5":
			m.currentView = viewSecrets
			m.selectedIndex = 0
			return m, nil

		case "6", "v":
			// Stats view
			m.currentView = viewStats
			m.selectedIndex = 0
			// Reload stats when switching to stats view
			return m, m.loadStatsData

		case "7":
			m.currentView = viewProxy
			m.selectedIndex = 0
			return m, checkProxyStatus

		case "8":
			m.currentView = viewRouting
			m.selectedIndex = 0
			return m, nil

		case "9":
			m.currentView = viewSuggestions
			m.selectedIndex = 0
			return m, m.loadSuggestions

		case "tab":
			// Cycle through views
			m.currentView = (m.currentView + 1) % 9
			m.selectedIndex = 0
			if m.currentView == viewStats {
				return m, m.loadStatsData
			} else if m.currentView == viewProxy {
				return m, checkProxyStatus
			} else if m.currentView == viewSuggestions {
				return m, m.loadSuggestions
			}
			return m, nil

		case "r":
			// Run selected tool (only in dashboard view)
			if m.currentView == viewDashboard {
				return m.handleRun()
			}
			return m, nil

		case "b":
			// Change binding (placeholder for now)
			// In future, this would open a binding edit view
			return m, nil

		case "x":
			// Toggle proxy for selected tool (only in dashboard view)
			if m.currentView == viewDashboard {
				return m.handleToggleProxy()
			}
			return m, nil

		case "s":
			// Check proxy status
			m.proxyStatus = "checking"
			return m, checkProxyStatus

		case "p":
			// View providers (placeholder for now)
			// In future, this would open provider management view
			return m, nil

		case "?":
			// Show help (placeholder for now)
			return m, nil

		case "up", "k":
			// Navigate up in list views
			if m.currentView != viewDashboard && m.selectedIndex > 0 {
				m.selectedIndex--
			}
			return m, nil

		case "down", "j":
			// Navigate down in list views
			maxIndex := 0
			switch m.currentView {
			case viewProviders:
				maxIndex = len(m.providers.Providers) - 1
			case viewTools:
				maxIndex = len(m.tools.Tools) - 1
			case viewBindings:
				maxIndex = len(m.bindings.Bindings) - 1
			case viewSecrets:
				maxIndex = len(m.providers.Providers) - 1 // Secrets are per-provider
			case viewSuggestions:
				maxIndex = len(m.suggestions) - 1
			}
			if m.currentView != viewDashboard && m.selectedIndex < maxIndex {
				m.selectedIndex++
			}
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateTableSize()
		return m, nil
	}

	// Update table (only in dashboard view)
	if m.currentView == viewDashboard {
		m.table, cmd = m.table.Update(msg)
	}
	return m, cmd
}

// updateTableSize adjusts table dimensions based on window size
func (m *DashboardModel) updateTableSize() {
	// Calculate available height for table
	headerHeight := 3
	footerHeight := 2
	availableHeight := m.height - headerHeight - footerHeight

	if availableHeight < 5 {
		availableHeight = 5
	}

	// Update column widths based on width
	columns := m.table.Columns()
	if m.width > 100 {
		columns[0].Width = 15  // Tool
		columns[1].Width = 25  // Provider
		columns[2].Width = 28  // Model
		columns[3].Width = 10  // Proxy
		columns[4].Width = 15  // Status
	} else if m.width < 80 {
		columns[0].Width = 10  // Tool
		columns[1].Width = 18  // Provider
		columns[2].Width = 20  // Model
		columns[3].Width = 8   // Proxy
		columns[4].Width = 12  // Status
	}

	m.table.SetColumns(columns)
	m.table.SetHeight(availableHeight)
}

// handleRun attempts to run the selected tool
func (m DashboardModel) handleRun() (tea.Model, tea.Cmd) {
	// Get selected row index
	selectedIdx := m.table.Cursor()

	if selectedIdx < 0 || selectedIdx >= len(m.tools.Tools) {
		return m, nil
	}

	tool := m.tools.Tools[selectedIdx]

	// Exit TUI and run the command
	// We'll quit and let the shell run `boba run <tool>`
	m.quitting = true

	// Print command hint
	fmt.Printf("\nRun: boba run %s\n", tool.ID)

	return m, tea.Quit
}

// handleToggleProxy toggles the proxy setting for the selected tool
func (m DashboardModel) handleToggleProxy() (tea.Model, tea.Cmd) {
	selectedIdx := m.table.Cursor()

	if selectedIdx < 0 || selectedIdx >= len(m.tools.Tools) {
		return m, nil
	}

	tool := m.tools.Tools[selectedIdx]

	// Find and toggle the binding
	binding, err := m.bindings.FindBinding(tool.ID)
	if err != nil {
		m.message = fmt.Sprintf("Tool %s is not bound to any provider", tool.Name)
		return m, nil
	}

	// Toggle proxy setting
	binding.UseProxy = !binding.UseProxy

	// Save the bindings
	if err := core.SaveBindings(m.home, m.bindings); err != nil {
		m.message = fmt.Sprintf("Failed to save binding: %v", err)
		return m, nil
	}

	// Update table rows to reflect the change
	m.table.SetRows(m.buildTableRows())

	// Set success message
	proxyState := "OFF"
	if binding.UseProxy {
		proxyState = "ON"
	}
	m.message = fmt.Sprintf("Proxy %s for %s", proxyState, tool.Name)

	return m, nil
}

// View renders the dashboard
func (m DashboardModel) View() string {
	if m.quitting {
		return ""
	}

	switch m.currentView {
	case viewProviders:
		return m.renderProvidersView()
	case viewTools:
		return m.renderToolsView()
	case viewBindings:
		return m.renderBindingsView()
	case viewSecrets:
		return m.renderSecretsView()
	case viewStats:
		return m.renderStatsView()
	case viewProxy:
		return m.renderProxyView()
	case viewRouting:
		return m.renderRoutingView()
	case viewSuggestions:
		return m.renderSuggestionsView()
	default:
		return m.renderDashboardView()
	}
}

// renderDashboardView renders the main dashboard view
func (m DashboardModel) renderDashboardView() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.Primary).
		Padding(0, 2)

	proxyStyle := lipgloss.NewStyle().
		Foreground(m.theme.Text).
		Padding(0, 2)

	helpStyle := lipgloss.NewStyle().
		Foreground(m.theme.Muted).
		Padding(1, 2)

	messageStyle := lipgloss.NewStyle().
		Foreground(m.theme.Success).
		Padding(0, 2)

	var content strings.Builder

	// Header
	title := "BobaMixer - AI CLI Control Plane"
	content.WriteString(titleStyle.Render(title))

	// Proxy status
	proxyStatusIcon := "○"
	proxyStatusText := "Checking..."
	switch m.proxyStatus {
	case "running":
		proxyStatusIcon = "●"
		proxyStatusText = "Running"
	case "stopped":
		proxyStatusIcon = "○"
		proxyStatusText = "Stopped"
	}
	proxyInfo := fmt.Sprintf("  Proxy: %s %s", proxyStatusIcon, proxyStatusText)
	content.WriteString(proxyStyle.Render(proxyInfo))
	content.WriteString("\n\n")

	// Table
	content.WriteString(m.table.View())
	content.WriteString("\n")

	// Message
	if m.message != "" {
		content.WriteString(messageStyle.Render("  "+m.message))
		content.WriteString("\n")
	}

	// Footer/Help
	helpText := "[1-9] Views: Dashboard|Providers|Tools|Bindings|Secrets|Stats|Proxy|Routing|Suggestions  [R] Run  [X] Proxy  [Tab] Next  [Q] Quit"
	content.WriteString(helpStyle.Render(helpText))

	return content.String()
}

// renderStatsView renders the usage statistics view
func (m DashboardModel) renderStatsView() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.Primary).
		Padding(0, 2)

	sectionStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.Success).
		Padding(1, 2)

	dataStyle := lipgloss.NewStyle().
		Foreground(m.theme.Text).
		Padding(0, 2)

	helpStyle := lipgloss.NewStyle().
		Foreground(m.theme.Muted).
		Padding(1, 2)

	errorStyle := lipgloss.NewStyle().
		Foreground(m.theme.Danger).
		Padding(0, 2)

	var content strings.Builder

	// Header
	title := "BobaMixer - Usage Statistics"
	content.WriteString(titleStyle.Render(title))
	content.WriteString("\n\n")

	// Check if stats are loaded
	if !m.statsLoaded {
		if m.statsError != "" {
			content.WriteString(errorStyle.Render(fmt.Sprintf("Error loading stats: %s", m.statsError)))
		} else {
			content.WriteString(dataStyle.Render("Loading stats..."))
		}
		content.WriteString("\n\n")
		helpText := "[V] Back to Dashboard  [Q] Quit"
		content.WriteString(helpStyle.Render(helpText))
		return content.String()
	}

	// Today's Stats
	content.WriteString(sectionStyle.Render("📅 Today's Usage"))
	content.WriteString("\n")
	content.WriteString(dataStyle.Render(fmt.Sprintf("  Tokens:   %d", m.todayStats.TotalTokens)))
	content.WriteString("\n")
	content.WriteString(dataStyle.Render(fmt.Sprintf("  Cost:     $%.4f", m.todayStats.TotalCost)))
	content.WriteString("\n")
	content.WriteString(dataStyle.Render(fmt.Sprintf("  Sessions: %d", m.todayStats.TotalSessions)))
	content.WriteString("\n\n")

	// Last 7 Days Stats
	content.WriteString(sectionStyle.Render("📊 Last 7 Days"))
	content.WriteString("\n")
	content.WriteString(dataStyle.Render(fmt.Sprintf("  Total Tokens:   %d", m.weekStats.TotalTokens)))
	content.WriteString("\n")
	content.WriteString(dataStyle.Render(fmt.Sprintf("  Total Cost:     $%.4f", m.weekStats.TotalCost)))
	content.WriteString("\n")
	content.WriteString(dataStyle.Render(fmt.Sprintf("  Total Sessions: %d", m.weekStats.TotalSessions)))
	content.WriteString("\n")
	content.WriteString(dataStyle.Render(fmt.Sprintf("  Avg Daily Tokens: %.0f", m.weekStats.AvgDailyTokens)))
	content.WriteString("\n")
	content.WriteString(dataStyle.Render(fmt.Sprintf("  Avg Daily Cost:   $%.4f", m.weekStats.AvgDailyCost)))
	content.WriteString("\n\n")

	// Profile Breakdown
	if len(m.profileStats) > 0 {
		content.WriteString(sectionStyle.Render("🎯 By Profile (7d)"))
		content.WriteString("\n")
		for _, ps := range m.profileStats {
			line := fmt.Sprintf("  • %s: tokens=%d cost=$%.4f sessions=%d latency=%.0fms usage=%.1f%% cost=%.1f%%",
				ps.ProfileName,
				ps.TotalTokens,
				ps.TotalCost,
				ps.SessionCount,
				ps.AvgLatencyMS,
				ps.UsagePercent,
				ps.CostPercent,
			)
			content.WriteString(dataStyle.Render(line))
			content.WriteString("\n")
		}
		content.WriteString("\n")
	}

	// Footer/Help
	helpText := "[V] Back to Dashboard  [S] Refresh  [Q] Quit"
	content.WriteString(helpStyle.Render(helpText))

	return content.String()
}

// renderProvidersView renders the AI providers management view
func (m DashboardModel) renderProvidersView() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.Primary).
		Padding(0, 2)

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.Success).
		Padding(1, 2)

	selectedStyle := lipgloss.NewStyle().
		Foreground(m.theme.Text).
		Background(m.theme.Primary).
		Bold(true).
		Padding(0, 1)

	normalStyle := lipgloss.NewStyle().
		Foreground(m.theme.Text).
		Padding(0, 1)

	mutedStyle := lipgloss.NewStyle().
		Foreground(m.theme.Muted).
		Padding(0, 1)

	helpStyle := lipgloss.NewStyle().
		Foreground(m.theme.Muted).
		Padding(1, 2)

	var content strings.Builder

	// Header
	title := "BobaMixer - AI Providers Management"
	content.WriteString(titleStyle.Render(title))
	content.WriteString("\n\n")

	// Section header
	content.WriteString(headerStyle.Render("📡 Available Providers"))
	content.WriteString("\n\n")

	// Provider list
	if len(m.providers.Providers) == 0 {
		content.WriteString(mutedStyle.Render("  No providers configured."))
		content.WriteString("\n")
	} else {
		for i, provider := range m.providers.Providers {
			// Status indicators
			enabledIcon := "✓"
			if !provider.Enabled {
				enabledIcon = "✗"
			}

			// Check if API key is configured
			keyStatus := "⚠"
			if _, err := core.ResolveAPIKey(&provider, m.secrets); err == nil {
				keyStatus = "🔑"
			}

			line := fmt.Sprintf("  %s %s %-25s %-35s %s",
				enabledIcon,
				keyStatus,
				provider.DisplayName,
				provider.BaseURL,
				provider.DefaultModel,
			)

			if i == m.selectedIndex {
				content.WriteString(selectedStyle.Render("▶ "+line))
			} else {
				content.WriteString(normalStyle.Render("  "+line))
			}
			content.WriteString("\n")
		}
	}

	content.WriteString("\n")

	// Selected provider details
	if m.selectedIndex < len(m.providers.Providers) {
		provider := m.providers.Providers[m.selectedIndex]
		content.WriteString(headerStyle.Render("Details"))
		content.WriteString("\n")
		content.WriteString(normalStyle.Render(fmt.Sprintf("  ID: %s", provider.ID)))
		content.WriteString("\n")
		content.WriteString(normalStyle.Render(fmt.Sprintf("  Kind: %s", provider.Kind)))
		content.WriteString("\n")
		content.WriteString(normalStyle.Render(fmt.Sprintf("  API Key Source: %s", provider.APIKey.Source)))
		content.WriteString("\n")
		if provider.APIKey.Source == core.APIKeySourceEnv {
			content.WriteString(normalStyle.Render(fmt.Sprintf("  Env Var: %s", provider.APIKey.EnvVar)))
			content.WriteString("\n")
		}
		content.WriteString("\n")
	}

	// Footer/Help
	helpText := "[1-6] Switch View  [↑/↓] Navigate  [Tab] Next View  [Q] Quit"
	content.WriteString(helpStyle.Render(helpText))

	return content.String()
}

// renderToolsView renders the CLI tools management view
func (m DashboardModel) renderToolsView() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.Primary).
		Padding(0, 2)

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.Success).
		Padding(1, 2)

	selectedStyle := lipgloss.NewStyle().
		Foreground(m.theme.Text).
		Background(m.theme.Primary).
		Bold(true).
		Padding(0, 1)

	normalStyle := lipgloss.NewStyle().
		Foreground(m.theme.Text).
		Padding(0, 1)

	mutedStyle := lipgloss.NewStyle().
		Foreground(m.theme.Muted).
		Padding(0, 1)

	helpStyle := lipgloss.NewStyle().
		Foreground(m.theme.Muted).
		Padding(1, 2)

	var content strings.Builder

	// Header
	title := "BobaMixer - CLI Tools Management"
	content.WriteString(titleStyle.Render(title))
	content.WriteString("\n\n")

	// Section header
	content.WriteString(headerStyle.Render("🛠 Detected Tools"))
	content.WriteString("\n\n")

	// Tools list
	if len(m.tools.Tools) == 0 {
		content.WriteString(mutedStyle.Render("  No tools configured."))
		content.WriteString("\n")
	} else {
		for i, tool := range m.tools.Tools {
			// Check if tool has a binding
			boundIcon := "○"
			if _, err := m.bindings.FindBinding(tool.ID); err == nil {
				boundIcon = "●"
			}

			line := fmt.Sprintf("  %s %-15s %-30s %s",
				boundIcon,
				tool.Name,
				tool.Exec,
				tool.Kind,
			)

			if i == m.selectedIndex {
				content.WriteString(selectedStyle.Render("▶ "+line))
			} else {
				content.WriteString(normalStyle.Render("  "+line))
			}
			content.WriteString("\n")
		}
	}

	content.WriteString("\n")

	// Selected tool details
	if m.selectedIndex < len(m.tools.Tools) {
		tool := m.tools.Tools[m.selectedIndex]
		content.WriteString(headerStyle.Render("Details"))
		content.WriteString("\n")
		content.WriteString(normalStyle.Render(fmt.Sprintf("  ID: %s", tool.ID)))
		content.WriteString("\n")
		content.WriteString(normalStyle.Render(fmt.Sprintf("  Config Type: %s", tool.ConfigType)))
		content.WriteString("\n")
		content.WriteString(normalStyle.Render(fmt.Sprintf("  Config Path: %s", tool.ConfigPath)))
		content.WriteString("\n")
		if tool.Description != "" {
			content.WriteString(normalStyle.Render(fmt.Sprintf("  Description: %s", tool.Description)))
			content.WriteString("\n")
		}
		content.WriteString("\n")
	}

	// Footer/Help
	helpText := "[1-6] Switch View  [↑/↓] Navigate  [Tab] Next View  [Q] Quit"
	content.WriteString(helpStyle.Render(helpText))

	return content.String()
}

// renderBindingsView renders the tool-to-provider bindings view
func (m DashboardModel) renderBindingsView() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.Primary).
		Padding(0, 2)

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.Success).
		Padding(1, 2)

	selectedStyle := lipgloss.NewStyle().
		Foreground(m.theme.Text).
		Background(m.theme.Primary).
		Bold(true).
		Padding(0, 1)

	normalStyle := lipgloss.NewStyle().
		Foreground(m.theme.Text).
		Padding(0, 1)

	mutedStyle := lipgloss.NewStyle().
		Foreground(m.theme.Muted).
		Padding(0, 1)

	helpStyle := lipgloss.NewStyle().
		Foreground(m.theme.Muted).
		Padding(1, 2)

	var content strings.Builder

	// Header
	title := "BobaMixer - Tool ↔ Provider Bindings"
	content.WriteString(titleStyle.Render(title))
	content.WriteString("\n\n")

	// Section header
	content.WriteString(headerStyle.Render("🔗 Active Bindings"))
	content.WriteString("\n\n")

	// Bindings list
	if len(m.bindings.Bindings) == 0 {
		content.WriteString(mutedStyle.Render("  No bindings configured."))
		content.WriteString("\n")
	} else {
		for i, binding := range m.bindings.Bindings {
			// Get tool name
			toolName := binding.ToolID
			if tool, err := m.tools.FindTool(binding.ToolID); err == nil {
				toolName = tool.Name
			}

			// Get provider name
			providerName := binding.ProviderID
			if provider, err := m.providers.FindProvider(binding.ProviderID); err == nil {
				providerName = provider.DisplayName
			}

			// Proxy status
			proxyIcon := "○"
			if binding.UseProxy {
				proxyIcon = "●"
			}

			line := fmt.Sprintf("  %-15s → %-25s  Proxy: %s",
				toolName,
				providerName,
				proxyIcon,
			)

			if i == m.selectedIndex {
				content.WriteString(selectedStyle.Render("▶ "+line))
			} else {
				content.WriteString(normalStyle.Render("  "+line))
			}
			content.WriteString("\n")
		}
	}

	content.WriteString("\n")

	// Selected binding details
	if m.selectedIndex < len(m.bindings.Bindings) {
		binding := m.bindings.Bindings[m.selectedIndex]
		content.WriteString(headerStyle.Render("Details"))
		content.WriteString("\n")
		content.WriteString(normalStyle.Render(fmt.Sprintf("  Tool ID: %s", binding.ToolID)))
		content.WriteString("\n")
		content.WriteString(normalStyle.Render(fmt.Sprintf("  Provider ID: %s", binding.ProviderID)))
		content.WriteString("\n")
		content.WriteString(normalStyle.Render(fmt.Sprintf("  Use Proxy: %t", binding.UseProxy)))
		content.WriteString("\n")
		if binding.Options.Model != "" {
			content.WriteString(normalStyle.Render(fmt.Sprintf("  Model Override: %s", binding.Options.Model)))
			content.WriteString("\n")
		}
		content.WriteString("\n")
	}

	// Footer/Help
	helpText := "[1-6] Switch View  [↑/↓] Navigate  [X] Toggle Proxy  [Tab] Next View  [Q] Quit"
	content.WriteString(helpStyle.Render(helpText))

	return content.String()
}

// renderSecretsView renders the API keys/secrets management view
func (m DashboardModel) renderSecretsView() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.Primary).
		Padding(0, 2)

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.Success).
		Padding(1, 2)

	selectedStyle := lipgloss.NewStyle().
		Foreground(m.theme.Text).
		Background(m.theme.Primary).
		Bold(true).
		Padding(0, 1)

	normalStyle := lipgloss.NewStyle().
		Foreground(m.theme.Text).
		Padding(0, 1)

	mutedStyle := lipgloss.NewStyle().
		Foreground(m.theme.Muted).
		Padding(0, 1)

	dangerStyle := lipgloss.NewStyle().
		Foreground(m.theme.Danger).
		Padding(0, 1)

	successStyle := lipgloss.NewStyle().
		Foreground(m.theme.Success).
		Padding(0, 1)

	helpStyle := lipgloss.NewStyle().
		Foreground(m.theme.Muted).
		Padding(1, 2)

	var content strings.Builder

	// Header
	title := "BobaMixer - Secrets Management (API Keys)"
	content.WriteString(titleStyle.Render(title))
	content.WriteString("\n\n")

	// Section header
	content.WriteString(headerStyle.Render("🔒 API Key Status"))
	content.WriteString("\n\n")

	// Provider secrets list
	if len(m.providers.Providers) == 0 {
		content.WriteString(mutedStyle.Render("  No providers configured."))
		content.WriteString("\n")
	} else {
		for i, provider := range m.providers.Providers {
			// Check if API key is configured
			hasKey := false
			keySource := "(not set)"
			if _, err := core.ResolveAPIKey(&provider, m.secrets); err == nil {
				hasKey = true
				keySource = string(provider.APIKey.Source)
			}

			var statusIcon, statusText string
			var keyStatusStyle lipgloss.Style
			if hasKey {
				statusIcon = "✓"
				statusText = "Configured"
				keyStatusStyle = successStyle
			} else {
				statusIcon = "✗"
				statusText = "Missing"
				keyStatusStyle = dangerStyle
			}

			line := fmt.Sprintf("  %-25s %s %-15s [%s]",
				provider.DisplayName,
				statusIcon,
				statusText,
				keySource,
			)

			var fullLine string
			if i == m.selectedIndex {
				fullLine = selectedStyle.Render("▶ " + line)
			} else {
				fullLine = normalStyle.Render("  "+line[:len("  ")+len(provider.DisplayName)+1]) +
					keyStatusStyle.Render(line[len("  ")+len(provider.DisplayName)+1:])
			}
			content.WriteString(fullLine)
			content.WriteString("\n")
		}
	}

	content.WriteString("\n")

	// Security notice
	content.WriteString(headerStyle.Render("🔐 Security"))
	content.WriteString("\n")
	content.WriteString(mutedStyle.Render("  • API keys are stored encrypted in ~/.boba/secrets.yaml"))
	content.WriteString("\n")
	content.WriteString(mutedStyle.Render("  • Keys can also be loaded from environment variables"))
	content.WriteString("\n")
	content.WriteString(mutedStyle.Render("  • Use 'boba edit secrets' to manage keys manually"))
	content.WriteString("\n\n")

	// Footer/Help
	helpText := "[1-6] Switch View  [↑/↓] Navigate  [Tab] Next View  [Q] Quit"
	content.WriteString(helpStyle.Render(helpText))

	return content.String()
}

// renderProxyView renders the proxy server control panel
func (m DashboardModel) renderProxyView() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.Primary).
		Padding(0, 2)

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.Success).
		Padding(1, 2)

	normalStyle := lipgloss.NewStyle().
		Foreground(m.theme.Text).
		Padding(0, 1)

	successStyle := lipgloss.NewStyle().
		Foreground(m.theme.Success).
		Padding(0, 1)

	dangerStyle := lipgloss.NewStyle().
		Foreground(m.theme.Danger).
		Padding(0, 1)

	helpStyle := lipgloss.NewStyle().
		Foreground(m.theme.Muted).
		Padding(1, 2)

	var content strings.Builder

	// Header
	title := "BobaMixer - Proxy Server Control"
	content.WriteString(titleStyle.Render(title))
	content.WriteString("\n\n")

	// Proxy status section
	content.WriteString(headerStyle.Render("🌐 Proxy Status"))
	content.WriteString("\n\n")

	var statusStyle lipgloss.Style
	var statusIcon, statusText string

	switch m.proxyStatus {
	case "running":
		statusIcon = "●"
		statusText = "Running"
		statusStyle = successStyle
	case "stopped":
		statusIcon = "○"
		statusText = "Stopped"
		statusStyle = dangerStyle
	default:
		statusIcon = "⋯"
		statusText = "Checking..."
		statusStyle = normalStyle
	}

	content.WriteString(normalStyle.Render(fmt.Sprintf("  Status:   %s", statusStyle.Render(statusIcon+" "+statusText))))
	content.WriteString("\n")
	content.WriteString(normalStyle.Render(fmt.Sprintf("  Address:  %s", proxy.DefaultAddr)))
	content.WriteString("\n\n")

	// Information section
	content.WriteString(headerStyle.Render("ℹ️  Information"))
	content.WriteString("\n")
	content.WriteString(normalStyle.Render("  The proxy server intercepts AI API requests from CLI tools"))
	content.WriteString("\n")
	content.WriteString(normalStyle.Render("  and routes them through BobaMixer for tracking and control."))
	content.WriteString("\n\n")

	// Usage
	if m.proxyStatus == "running" {
		content.WriteString(headerStyle.Render("📝 Configuration"))
		content.WriteString("\n")
		content.WriteString(normalStyle.Render("  Tools with proxy enabled will automatically use:"))
		content.WriteString("\n")
		content.WriteString(normalStyle.Render(fmt.Sprintf("  • HTTP_PROXY=%s", proxy.DefaultAddr)))
		content.WriteString("\n")
		content.WriteString(normalStyle.Render(fmt.Sprintf("  • HTTPS_PROXY=%s", proxy.DefaultAddr)))
		content.WriteString("\n\n")
	}

	// Footer/Help
	var helpText string
	if m.proxyStatus == "running" {
		helpText = "[1-9] Switch View  [S] Refresh Status  [Tab] Next View  [Q] Quit"
	} else {
		helpText = "[1-9] Switch View  [S] Refresh Status  [Tab] Next View  [Q] Quit\n  Note: Use 'boba proxy serve' in terminal to start the proxy server"
	}
	content.WriteString(helpStyle.Render(helpText))

	return content.String()
}

// renderRoutingView renders the routing rules tester
func (m DashboardModel) renderRoutingView() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.Primary).
		Padding(0, 2)

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.Success).
		Padding(1, 2)

	normalStyle := lipgloss.NewStyle().
		Foreground(m.theme.Text).
		Padding(0, 1)

	mutedStyle := lipgloss.NewStyle().
		Foreground(m.theme.Muted).
		Padding(0, 1)

	helpStyle := lipgloss.NewStyle().
		Foreground(m.theme.Muted).
		Padding(1, 2)

	var content strings.Builder

	// Header
	title := "BobaMixer - Routing Rules Tester"
	content.WriteString(titleStyle.Render(title))
	content.WriteString("\n\n")

	// Description
	content.WriteString(headerStyle.Render("🧪 Test Routing Rules"))
	content.WriteString("\n")
	content.WriteString(mutedStyle.Render("  Test how routing rules would apply to different queries."))
	content.WriteString("\n\n")

	// Example usage
	content.WriteString(headerStyle.Render("💡 How to Use"))
	content.WriteString("\n")
	content.WriteString(normalStyle.Render("  1. Prepare a test query (text or file)"))
	content.WriteString("\n")
	content.WriteString(normalStyle.Render("  2. Run: boba route test \"your query text\""))
	content.WriteString("\n")
	content.WriteString(normalStyle.Render("  3. Or: boba route test @path/to/file.txt"))
	content.WriteString("\n\n")

	// Example
	content.WriteString(headerStyle.Render("📋 Example"))
	content.WriteString("\n")
	content.WriteString(normalStyle.Render("  $ boba route test \"Write a Python function\""))
	content.WriteString("\n")
	content.WriteString(mutedStyle.Render("  → Profile: claude-sonnet-3.5"))
	content.WriteString("\n")
	content.WriteString(mutedStyle.Render("  → Rule: short-query-fast-model"))
	content.WriteString("\n")
	content.WriteString(mutedStyle.Render("  → Reason: Query < 100 chars"))
	content.WriteString("\n\n")

	// Info
	content.WriteString(headerStyle.Render("ℹ️  Context Detection"))
	content.WriteString("\n")
	content.WriteString(normalStyle.Render("  Routing considers:"))
	content.WriteString("\n")
	content.WriteString(normalStyle.Render("  • Query length and complexity"))
	content.WriteString("\n")
	content.WriteString(normalStyle.Render("  • Current project and branch"))
	content.WriteString("\n")
	content.WriteString(normalStyle.Render("  • Time of day (day/evening/night)"))
	content.WriteString("\n")
	content.WriteString(normalStyle.Render("  • Project type (go, web, etc.)"))
	content.WriteString("\n\n")

	// Footer/Help
	helpText := "[1-9] Switch View  [Tab] Next View  [Q] Quit\n  Use CLI: boba route test <text|@file>"
	content.WriteString(helpStyle.Render(helpText))

	return content.String()
}

// renderSuggestionsView renders the optimization suggestions view
func (m DashboardModel) renderSuggestionsView() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.Primary).
		Padding(0, 2)

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.Success).
		Padding(1, 2)

	selectedStyle := lipgloss.NewStyle().
		Foreground(m.theme.Text).
		Background(m.theme.Primary).
		Bold(true).
		Padding(0, 1)

	normalStyle := lipgloss.NewStyle().
		Foreground(m.theme.Text).
		Padding(0, 1)

	mutedStyle := lipgloss.NewStyle().
		Foreground(m.theme.Muted).
		Padding(0, 1)

	warningStyle := lipgloss.NewStyle().
		Foreground(m.theme.Warning).
		Padding(0, 1)

	dangerStyle := lipgloss.NewStyle().
		Foreground(m.theme.Danger).
		Padding(0, 1)

	helpStyle := lipgloss.NewStyle().
		Foreground(m.theme.Muted).
		Padding(1, 2)

	var content strings.Builder

	// Header
	title := "BobaMixer - Optimization Suggestions"
	content.WriteString(titleStyle.Render(title))
	content.WriteString("\n\n")

	// Check for errors
	if m.suggestionsError != "" {
		content.WriteString(dangerStyle.Render(fmt.Sprintf("  Error: %s", m.suggestionsError)))
		content.WriteString("\n\n")
		helpText := "[1-9] Switch View  [R] Retry  [Tab] Next View  [Q] Quit"
		content.WriteString(helpStyle.Render(helpText))
		return content.String()
	}

	// Section header
	content.WriteString(headerStyle.Render("💡 Recommendations (Last 7 Days)"))
	content.WriteString("\n\n")

	// Suggestions list
	if len(m.suggestions) == 0 {
		content.WriteString(mutedStyle.Render("  ✓ No suggestions - your usage is optimized!"))
		content.WriteString("\n")
	} else {
		for i, sugg := range m.suggestions {
			// Priority indicator
			var priorityStyle lipgloss.Style
			var priorityIcon string
			switch sugg.Priority {
			case 5:
				priorityStyle = dangerStyle
				priorityIcon = "🔴"
			case 4:
				priorityStyle = warningStyle
				priorityIcon = "🟠"
			case 3:
				priorityStyle = normalStyle
				priorityIcon = "🟡"
			default:
				priorityStyle = mutedStyle
				priorityIcon = "🟢"
			}

			// Type icon
			var typeIcon string
			switch sugg.Type {
			case suggestions.SuggestionCostOptimization:
				typeIcon = "💰"
			case suggestions.SuggestionProfileSwitch:
				typeIcon = "🔄"
			case suggestions.SuggestionBudgetAdjust:
				typeIcon = "📊"
			case suggestions.SuggestionAnomaly:
				typeIcon = "⚠️ "
			default:
				typeIcon = "📈"
			}

			line := fmt.Sprintf("  %s %s [P%d] %s",
				priorityIcon,
				typeIcon,
				sugg.Priority,
				sugg.Title,
			)

			if i == m.selectedIndex {
				content.WriteString(selectedStyle.Render("▶ "+line))
			} else {
				content.WriteString(priorityStyle.Render(line))
			}
			content.WriteString("\n")
		}

		// Selected suggestion details
		if m.selectedIndex < len(m.suggestions) {
			sugg := m.suggestions[m.selectedIndex]
			content.WriteString("\n")
			content.WriteString(headerStyle.Render("Details"))
			content.WriteString("\n")
			content.WriteString(normalStyle.Render(fmt.Sprintf("  %s", sugg.Description)))
			content.WriteString("\n")
			content.WriteString(normalStyle.Render(fmt.Sprintf("  Impact: %s", sugg.Impact)))
			content.WriteString("\n\n")

			if len(sugg.ActionItems) > 0 {
				content.WriteString(headerStyle.Render("Recommended Actions"))
				content.WriteString("\n")
				for idx, action := range sugg.ActionItems {
					content.WriteString(normalStyle.Render(fmt.Sprintf("  %d. %s", idx+1, action)))
					content.WriteString("\n")
				}
			}
		}
	}

	content.WriteString("\n")

	// Footer/Help
	helpText := "[1-9] Switch View  [↑/↓] Navigate  [Tab] Next View  [Q] Quit\n  Use CLI: boba action [--auto] to apply suggestions"
	content.WriteString(helpStyle.Render(helpText))

	return content.String()
}

// loadSuggestions loads optimization suggestions
func (m *DashboardModel) loadSuggestions() tea.Msg {
	dbPath := filepath.Join(m.home, "usage.db")
	db, err := sqlite.Open(dbPath)
	if err != nil {
		m.suggestionsError = err.Error()
		return nil
	}

	engine := suggestions.NewEngine(db)
	suggs, err := engine.GenerateSuggestions(7)
	if err != nil {
		m.suggestionsError = err.Error()
		return nil
	}

	m.suggestions = suggs
	m.suggestionsError = ""
	return nil
}

// RunDashboard starts the dashboard TUI
func RunDashboard(home string) error {
	dashboard, err := NewDashboard(home)
	if err != nil {
		return fmt.Errorf("failed to create dashboard: %w", err)
	}

	p := tea.NewProgram(dashboard, tea.WithAltScreen())
	_, err = p.Run()
	return err
}
