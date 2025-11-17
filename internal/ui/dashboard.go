package ui

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/royisme/bobamixer/internal/domain/core"
	"github.com/royisme/bobamixer/internal/proxy"
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

	// UI components
	table table.Model

	// State
	width       int
	height      int
	quitting    bool
	proxyStatus string // "running", "stopped", "checking"
	message     string // Status message to display
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
	}

	m.initializeTable()

	return m, nil
}

// proxyStatusMsg is sent when proxy status is checked
type proxyStatusMsg struct {
	running bool
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
	// Check proxy status on startup
	return checkProxyStatus
}

// Update handles messages
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

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "r":
			// Run selected tool
			return m.handleRun()

		case "b":
			// Change binding (placeholder for now)
			// In future, this would open a binding edit view
			return m, nil

		case "x":
			// Toggle proxy for selected tool
			return m.handleToggleProxy()

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
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateTableSize()
		return m, nil
	}

	// Update table
	m.table, cmd = m.table.Update(msg)
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
	helpText := "[R] Run  [X] Toggle Proxy  [S] Check Proxy  [P] Providers  [?] Help  [Q] Quit"
	content.WriteString(helpStyle.Render(helpText))

	return content.String()
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
