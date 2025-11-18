package ui

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/royisme/bobamixer/internal/domain/core"
	"github.com/royisme/bobamixer/internal/domain/stats"
	"github.com/royisme/bobamixer/internal/domain/suggestions"
	"github.com/royisme/bobamixer/internal/proxy"
	"github.com/royisme/bobamixer/internal/store/sqlite"
	dashboardviews "github.com/royisme/bobamixer/internal/ui/dashboard/views"
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
	viewReports
	viewHooks
	viewConfig
	viewHelp
)

// UI constants for repeated strings
const (
	proxyStatusRunning    = "running"
	proxyStatusStopped    = "stopped"
	proxyStatusChecking   = "checking"
	proxyStateOn          = "ON"
	proxyStateOff         = "OFF"
	iconCircleFilled      = "●"
	iconCircleEmpty       = "○"
	iconCheckmark         = "✓"
	iconCross             = "✗"
	helpTextNavigation    = "[1-5] Switch Section  [Tab] Next Section  [[ / ]] Cycle Views  [/] Search  [?] Help  [Q] Quit"
	msgNoProviderSelected = "No provider selected"
	msgInvalidProvider    = "Invalid provider selection"
	promptPrefix          = "│ "
)

type providerFormField int

const (
	providerFieldID providerFormField = iota
	providerFieldKind
	providerFieldDisplayName
	providerFieldBaseURL
	providerFieldDefaultModel
	providerFieldAPIKeySource
	providerFieldAPIKeyEnv
)

var providerFieldSequence = []providerFormField{
	providerFieldID,
	providerFieldKind,
	providerFieldDisplayName,
	providerFieldBaseURL,
	providerFieldDefaultModel,
	providerFieldAPIKeySource,
	providerFieldAPIKeyEnv,
}

type bindingFormField int

const (
	bindingFieldToolID bindingFormField = iota
	bindingFieldProviderID
	bindingFieldModel
	bindingFieldUseProxy
)

var bindingFieldSequence = []bindingFormField{
	bindingFieldToolID,
	bindingFieldProviderID,
	bindingFieldModel,
	bindingFieldUseProxy,
}

type viewSection struct {
	name     string
	shortcut string
	views    []viewMode
}

func (m *DashboardModel) initSections() {
	m.sections = []viewSection{
		{
			name:     "Dashboard",
			shortcut: "1",
			views:    []viewMode{viewDashboard},
		},
		{
			name:     "Control Plane",
			shortcut: "2",
			views:    []viewMode{viewProviders, viewTools, viewBindings, viewSecrets, viewProxy},
		},
		{
			name:     "Usage",
			shortcut: "3",
			views:    []viewMode{viewStats, viewReports},
		},
		{
			name:     "Optimization",
			shortcut: "4",
			views:    []viewMode{viewSuggestions},
		},
		{
			name:     "DevOps",
			shortcut: "5",
			views:    []viewMode{viewRouting, viewHooks, viewConfig},
		},
	}
	m.currentSection = 0
	m.sectionViewIndex = 0
	m.updateViewFromSection()
}

func (m *DashboardModel) updateViewFromSection() {
	if len(m.sections) == 0 {
		m.currentView = viewDashboard
		return
	}

	if m.currentSection < 0 {
		m.currentSection = 0
	}
	if m.currentSection >= len(m.sections) {
		m.currentSection = 0
	}

	section := m.sections[m.currentSection]
	if len(section.views) == 0 {
		m.currentView = viewDashboard
		return
	}

	if m.sectionViewIndex < 0 {
		m.sectionViewIndex = 0
	}
	if m.sectionViewIndex >= len(section.views) {
		m.sectionViewIndex = 0
	}

	nextView := section.views[m.sectionViewIndex]
	if m.currentView != nextView {
		m.currentView = nextView
		m.selectedIndex = 0
	}
	if m.searchContextView != m.currentView {
		m.searchActive = false
		m.searchQuery = ""
		m.searchContextView = m.currentView
	}
}

func (m *DashboardModel) moveToSection(idx int) tea.Cmd {
	if idx < 0 || idx >= len(m.sections) {
		return nil
	}
	m.currentSection = idx
	m.sectionViewIndex = 0
	m.updateViewFromSection()
	return m.sectionEnterCmd()
}

func (m *DashboardModel) cycleSection(delta int) tea.Cmd {
	m.currentSection = (m.currentSection + delta + len(m.sections)) % len(m.sections)
	m.sectionViewIndex = 0
	m.updateViewFromSection()
	return m.sectionEnterCmd()
}

func (m *DashboardModel) cycleSubview(delta int) tea.Cmd {
	section := m.sections[m.currentSection]
	if len(section.views) == 0 {
		return nil
	}
	m.sectionViewIndex = (m.sectionViewIndex + delta + len(section.views)) % len(section.views)
	m.updateViewFromSection()
	return m.sectionEnterCmd()
}

func (m *DashboardModel) sectionEnterCmd() tea.Cmd {
	switch m.currentView {
	case viewStats:
		return m.loadStatsData
	case viewProxy:
		return checkProxyStatus
	case viewSuggestions:
		return m.loadSuggestions
	default:
		return nil
	}
}

func (m *DashboardModel) supportsSearch(view viewMode) bool {
	switch view {
	case viewProviders, viewTools, viewBindings, viewSecrets:
		return true
	default:
		return false
	}
}

func (m *DashboardModel) activateSearch() {
	m.searchActive = true
	m.searchInput.SetValue(m.searchQuery)
	m.searchInput.CursorEnd()
	m.searchContextView = m.currentView
}

func (m *DashboardModel) clearSearch() {
	m.searchActive = false
	m.searchQuery = ""
}

func (m *DashboardModel) startSecretInput() {
	indexes := m.filteredProviderIndexes()
	if len(indexes) == 0 || m.selectedIndex < 0 || m.selectedIndex >= len(indexes) {
		m.secretMessage = msgNoProviderSelected
		return
	}

	targetIdx := indexes[m.selectedIndex]
	if targetIdx < 0 || targetIdx >= len(m.providers.Providers) {
		m.secretMessage = msgInvalidProvider
		return
	}

	provider := m.providers.Providers[targetIdx]
	m.secretTargetIndex = targetIdx
	m.secretInput.SetValue("")
	m.secretInput.Placeholder = fmt.Sprintf("API key for %s", provider.DisplayName)
	m.secretInput.CursorEnd()
	m.secretInput.Focus()
	m.secretInputActive = true
	m.searchActive = false
	m.secretMessage = ""
}

func (m *DashboardModel) ensureSecretsConfig() {
	if m.secrets == nil {
		m.secrets = &core.SecretsConfig{
			Version: 1,
			Secrets: make(map[string]core.Secret),
		}
	}
	if m.secrets.Secrets == nil {
		m.secrets.Secrets = make(map[string]core.Secret)
	}
}

func (m *DashboardModel) saveSecretInput() {
	if m.secretTargetIndex < 0 || m.secretTargetIndex >= len(m.providers.Providers) {
		m.secretMessage = msgInvalidProvider
		m.secretInputActive = false
		return
	}

	apiKey := strings.TrimSpace(m.secretInput.Value())
	if apiKey == "" {
		m.secretMessage = "API key cannot be empty"
		return
	}

	provider := m.providers.Providers[m.secretTargetIndex]
	m.ensureSecretsConfig()
	m.secrets.Secrets[provider.ID] = core.Secret{
		ProviderID: provider.ID,
		APIKey:     apiKey,
	}

	if err := core.SaveSecrets(m.home, m.secrets); err != nil {
		m.secretMessage = fmt.Sprintf("Failed to save API key: %v", err)
	} else {
		m.secretMessage = fmt.Sprintf("API key saved for %s", provider.DisplayName)
	}

	m.secretInputActive = false
	m.secretInput.Blur()
	m.secretInput.SetValue("")
}

func (m *DashboardModel) handleSecretRemove() {
	indexes := m.filteredProviderIndexes()
	if len(indexes) == 0 || m.selectedIndex < 0 || m.selectedIndex >= len(indexes) {
		m.secretMessage = msgNoProviderSelected
		return
	}
	targetIdx := indexes[m.selectedIndex]
	if targetIdx < 0 || targetIdx >= len(m.providers.Providers) {
		m.secretMessage = msgInvalidProvider
		return
	}
	provider := m.providers.Providers[targetIdx]

	m.ensureSecretsConfig()
	if _, ok := m.secrets.Secrets[provider.ID]; !ok {
		m.secretMessage = fmt.Sprintf("No API key found for %s", provider.DisplayName)
		return
	}

	delete(m.secrets.Secrets, provider.ID)
	if err := core.SaveSecrets(m.home, m.secrets); err != nil {
		m.secretMessage = fmt.Sprintf("Failed to remove API key: %v", err)
		return
	}
	m.secretMessage = fmt.Sprintf("Removed API key for %s", provider.DisplayName)
}

func (m *DashboardModel) handleSecretTest() {
	indexes := m.filteredProviderIndexes()
	if len(indexes) == 0 || m.selectedIndex < 0 || m.selectedIndex >= len(indexes) {
		m.secretMessage = msgNoProviderSelected
		return
	}
	targetIdx := indexes[m.selectedIndex]
	if targetIdx < 0 || targetIdx >= len(m.providers.Providers) {
		m.secretMessage = msgInvalidProvider
		return
	}
	provider := m.providers.Providers[targetIdx]

	if _, err := core.ResolveAPIKey(&provider, m.secrets); err != nil {
		m.secretMessage = fmt.Sprintf("API key missing: %v", err)
		return
	}
	m.secretMessage = fmt.Sprintf("API key available for %s", provider.DisplayName)
}

func (m *DashboardModel) startProviderForm(add bool) {
	indexes := m.filteredProviderIndexes()
	if !add {
		if len(indexes) == 0 || m.selectedIndex < 0 || m.selectedIndex >= len(indexes) {
			m.providerFormMessage = msgNoProviderSelected
			return
		}
		targetIdx := indexes[m.selectedIndex]
		if targetIdx < 0 || targetIdx >= len(m.providers.Providers) {
			m.providerFormMessage = msgInvalidProvider
			return
		}
		m.providerFormProvider = m.providers.Providers[targetIdx]
		m.providerFormIndex = targetIdx
	} else {
		m.providerFormProvider = core.Provider{
			Enabled: true,
			APIKey: core.APIKeyConfig{
				Source: core.APIKeySourceEnv,
			},
		}
		m.providerFormIndex = -1
	}

	m.providerFormAdd = add
	m.providerFormActive = true
	m.providerFormField = 0
	if !add {
		// Skip ID when editing existing provider
		m.providerFormField = 1
	}
	m.prepareProviderFormInput()
	m.providerFormInput.Focus()
	m.providerFormMessage = ""
	m.searchActive = false
}

func (m *DashboardModel) providerFieldEnabled(field providerFormField) bool {
	if !m.providerFormAdd && field == providerFieldID {
		return false
	}
	if field == providerFieldAPIKeyEnv {
		return strings.ToLower(string(m.providerFormProvider.APIKey.Source)) == string(core.APIKeySourceEnv)
	}
	return true
}

func (m *DashboardModel) prepareProviderFormInput() {
	if m.providerFormField >= len(providerFieldSequence) {
		return
	}
	field := providerFieldSequence[m.providerFormField]
	m.providerFormInput.Placeholder = m.providerFieldPrompt(field)
	switch field {
	case providerFieldID:
		m.providerFormInput.SetValue(m.providerFormProvider.ID)
	case providerFieldKind:
		if m.providerFormProvider.Kind != "" {
			m.providerFormInput.SetValue(string(m.providerFormProvider.Kind))
		} else {
			m.providerFormInput.SetValue("")
		}
	case providerFieldDisplayName:
		m.providerFormInput.SetValue(m.providerFormProvider.DisplayName)
	case providerFieldBaseURL:
		m.providerFormInput.SetValue(m.providerFormProvider.BaseURL)
	case providerFieldDefaultModel:
		m.providerFormInput.SetValue(m.providerFormProvider.DefaultModel)
	case providerFieldAPIKeySource:
		if m.providerFormProvider.APIKey.Source != "" {
			m.providerFormInput.SetValue(string(m.providerFormProvider.APIKey.Source))
		} else {
			m.providerFormInput.SetValue("")
		}
	case providerFieldAPIKeyEnv:
		m.providerFormInput.SetValue(m.providerFormProvider.APIKey.EnvVar)
	}
}

func (m *DashboardModel) providerFieldPrompt(field providerFormField) string {
	switch field {
	case providerFieldID:
		return "provider id (e.g. openai-official)"
	case providerFieldKind:
		return "provider kind (openai, anthropic, ...)"
	case providerFieldDisplayName:
		return "display name"
	case providerFieldBaseURL:
		return "base URL"
	case providerFieldDefaultModel:
		return "default model"
	case providerFieldAPIKeySource:
		return "api key source (env or secrets)"
	case providerFieldAPIKeyEnv:
		return "env var name (if source=env)"
	default:
		return ""
	}
}

func (m *DashboardModel) submitProviderFormValue() {
	if m.providerFormField >= len(providerFieldSequence) {
		return
	}
	field := providerFieldSequence[m.providerFormField]
	value := strings.TrimSpace(m.providerFormInput.Value())
	if err := m.setProviderFieldValue(field, value); err != nil {
		m.providerFormMessage = err.Error()
		return
	}
	m.providerFormMessage = ""
	m.providerFormInput.SetValue("")
	for {
		m.providerFormField++
		if m.providerFormField >= len(providerFieldSequence) {
			m.finishProviderForm()
			return
		}
		if m.providerFieldEnabled(providerFieldSequence[m.providerFormField]) {
			m.prepareProviderFormInput()
			return
		}
	}
}

func (m *DashboardModel) setProviderFieldValue(field providerFormField, value string) error {
	value = strings.TrimSpace(value)
	switch field {
	case providerFieldID:
		return m.setProviderID(value)
	case providerFieldKind:
		return m.setProviderKind(value)
	case providerFieldDisplayName:
		return m.setProviderDisplayName(value)
	case providerFieldBaseURL:
		return m.setProviderBaseURL(value)
	case providerFieldDefaultModel:
		return m.setProviderDefaultModel(value)
	case providerFieldAPIKeySource:
		return m.setProviderAPIKeySource(value)
	case providerFieldAPIKeyEnv:
		return m.setProviderAPIKeyEnv(value)
	default:
		return nil
	}
}

func (m *DashboardModel) setProviderID(value string) error {
	if value == "" {
		return fmt.Errorf("provider ID cannot be empty")
	}
	for i := range m.providers.Providers {
		if strings.EqualFold(m.providers.Providers[i].ID, value) {
			return fmt.Errorf("provider ID already exists")
		}
	}
	m.providerFormProvider.ID = value
	return nil
}

func (m *DashboardModel) setProviderKind(value string) error {
	if value == "" {
		return fmt.Errorf("provider kind cannot be empty")
	}
	m.providerFormProvider.Kind = core.ProviderKind(value)
	return nil
}

func (m *DashboardModel) setProviderDisplayName(value string) error {
	if value == "" {
		return fmt.Errorf("display name cannot be empty")
	}
	m.providerFormProvider.DisplayName = value
	return nil
}

func (m *DashboardModel) setProviderBaseURL(value string) error {
	if value == "" {
		return fmt.Errorf("base URL cannot be empty")
	}
	m.providerFormProvider.BaseURL = value
	return nil
}

func (m *DashboardModel) setProviderDefaultModel(value string) error {
	if value == "" {
		return fmt.Errorf("default model cannot be empty")
	}
	m.providerFormProvider.DefaultModel = value
	return nil
}

func (m *DashboardModel) setProviderAPIKeySource(value string) error {
	v := strings.ToLower(value)
	switch v {
	case "env":
		m.providerFormProvider.APIKey.Source = core.APIKeySourceEnv
	case "secrets":
		m.providerFormProvider.APIKey.Source = core.APIKeySourceSecrets
		m.providerFormProvider.APIKey.EnvVar = ""
	default:
		return fmt.Errorf("api key source must be 'env' or 'secrets'")
	}
	return nil
}

func (m *DashboardModel) setProviderAPIKeyEnv(value string) error {
	if m.providerFormProvider.APIKey.Source == core.APIKeySourceEnv && value == "" {
		return fmt.Errorf("env var is required when source=env")
	}
	m.providerFormProvider.APIKey.EnvVar = value
	return nil
}

func (m *DashboardModel) finishProviderForm() {
	if m.providerFormProvider.ID == "" {
		m.providerFormMessage = "provider ID is required"
		return
	}

	if m.providerFormAdd {
		m.providers.Providers = append(m.providers.Providers, m.providerFormProvider)
	} else if m.providerFormIndex >= 0 && m.providerFormIndex < len(m.providers.Providers) {
		m.providers.Providers[m.providerFormIndex] = m.providerFormProvider
	}

	if err := core.SaveProviders(m.home, m.providers); err != nil {
		m.providerFormMessage = fmt.Sprintf("failed to save provider: %v", err)
	} else if m.providerFormAdd {
		m.providerFormMessage = fmt.Sprintf("provider %s created", m.providerFormProvider.DisplayName)
	} else {
		m.providerFormMessage = fmt.Sprintf("provider %s updated", m.providerFormProvider.DisplayName)
	}

	m.providerFormActive = false
	m.providerFormAdd = false
	m.providerFormInput.Blur()
}

type reportOption struct {
	label string
	desc  string
}

var reportOptions = []reportOption{
	{"Last 7 Days Report", "Generate usage report for the past 7 days"},
	{"Last 30 Days Report", "Generate monthly usage report"},
	{"Custom Date Range", "Specify custom start and end dates"},
	{"JSON Format", "Export report as JSON (default)"},
	{"CSV Format", "Export report as CSV for spreadsheet tools"},
	{"HTML Format", "Generate visual HTML report with charts"},
}

type configFile struct {
	name string
	file string
	desc string
}

var configFiles = []configFile{
	{"Providers", "providers.yaml", "AI provider configurations and API endpoints"},
	{"Tools", "tools.yaml", "CLI tool detection and management"},
	{"Bindings", "bindings.yaml", "Tool-to-provider bindings and proxy settings"},
	{"Secrets", "secrets.yaml", "Encrypted API keys (edit with caution!)"},
	{"Routes", "routes.yaml", "Context-based routing rules"},
	{"Pricing", "pricing.yaml", "Token pricing for cost calculations"},
	{"Settings", "settings.yaml", "Global application settings"},
}

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

	// UI components
	table table.Model

	// State
	currentView          viewMode
	selectedIndex        int // Currently selected item in list views
	width                int
	height               int
	quitting             bool
	proxyStatus          string // proxyStatusRunning, proxyStatusStopped, proxyStatusChecking
	message              string // Status message to display
	sections             []viewSection
	currentSection       int
	sectionViewIndex     int
	showHelpOverlay      bool
	searchActive         bool
	searchInput          textinput.Model
	searchQuery          string
	searchContextView    viewMode
	secretInputActive    bool
	secretInput          textinput.Model
	secretTargetIndex    int
	secretMessage        string
	providerFormActive   bool
	providerFormAdd      bool
	providerFormIndex    int
	providerFormField    int
	providerFormInput    textinput.Model
	providerFormProvider core.Provider
	providerFormMessage  string
	bindingFormActive    bool
	bindingFormAdd       bool
	bindingFormIndex     int
	bindingFormField     int
	bindingFormInput     textinput.Model
	bindingFormBinding   core.Binding
	bindingFormMessage   string
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
		proxyStatus: proxyStatusChecking,
		currentView: viewDashboard,
	}

	m.initSections()
	searchInput := textinput.New()
	searchInput.Placeholder = "Search..."
	searchInput.CharLimit = 100
	searchInput.Width = 30
	m.searchInput = searchInput

	secretInput := textinput.New()
	secretInput.Placeholder = "Enter API key"
	secretInput.CharLimit = 200
	secretInput.Width = 40
	secretInput.Prompt = promptPrefix
	secretInput.EchoMode = textinput.EchoPassword
	secretInput.EchoCharacter = '•'
	m.secretInput = secretInput

	providerInput := textinput.New()
	providerInput.CharLimit = 200
	providerInput.Width = 50
	providerInput.Prompt = promptPrefix
	m.providerFormInput = providerInput

	bindingInput := textinput.New()
	bindingInput.CharLimit = 200
	bindingInput.Width = 40
	bindingInput.Prompt = promptPrefix
	m.bindingFormInput = bindingInput

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
		proxyStatus := proxyStateOff
		if binding.UseProxy {
			proxyStatus = proxyStateOn
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
//
//nolint:gocyclo // UI event handlers are inherently complex
func (m DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case proxyStatusMsg:
		// Update proxy status based on check
		if msg.running {
			m.proxyStatus = proxyStatusRunning
		} else {
			m.proxyStatus = proxyStatusStopped
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

	case suggestionsLoadedMsg:
		if msg.err != nil {
			m.suggestionsError = msg.err.Error()
			return m, nil
		}

		m.suggestions = msg.suggestions
		m.suggestionsError = ""
		return m, nil

	case tea.KeyMsg:
		key := msg.String()
		if key == "ctrl+c" || key == "q" {
			m.quitting = true
			return m, tea.Quit
		}

		if m.providerFormActive {
			switch key {
			case keyEsc:
				m.providerFormActive = false
				m.providerFormInput.Blur()
				m.providerFormMessage = "Provider edit canceled"
				return m, nil
			case keyEnter:
				m.submitProviderFormValue()
				return m, nil
			default:
				var cmd tea.Cmd
				m.providerFormInput, cmd = m.providerFormInput.Update(msg)
				return m, cmd
			}
		}

		if m.bindingFormActive {
			switch key {
			case keyEsc:
				m.bindingFormActive = false
				m.bindingFormInput.Blur()
				m.bindingFormMessage = "Binding edit canceled"
				return m, nil
			case keyEnter:
				m.submitBindingFormValue()
				return m, nil
			default:
				var cmd tea.Cmd
				m.bindingFormInput, cmd = m.bindingFormInput.Update(msg)
				return m, cmd
			}
		}

		if m.searchActive {
			switch key {
			case keyEsc:
				m.clearSearch()
				return m, nil
			case keyEnter:
				m.searchQuery = strings.TrimSpace(m.searchInput.Value())
				m.searchActive = false
				return m, nil
			default:
				var cmd tea.Cmd
				m.searchInput, cmd = m.searchInput.Update(msg)
				m.searchQuery = m.searchInput.Value()
				return m, cmd
			}
		}

		if m.secretInputActive {
			switch key {
			case keyEsc:
				m.secretInputActive = false
				m.secretInput.Blur()
				m.secretInput.SetValue("")
				return m, nil
			case keyEnter:
				m.saveSecretInput()
				return m, nil
			default:
				var cmd tea.Cmd
				m.secretInput, cmd = m.secretInput.Update(msg)
				return m, cmd
			}
		}

		switch key {
		case "tab":
			return m, m.cycleSection(1)
		case "shift+tab":
			return m, m.cycleSection(-1)
		case "[":
			return m, m.cycleSubview(-1)
		case "]":
			return m, m.cycleSubview(1)
		case "1", "2", "3", "4", "5":
			sectionIndex := int(key[0] - '1')
			return m, m.moveToSection(sectionIndex)

		case "r":
			if m.currentView == viewSecrets {
				m.handleSecretRemove()
				return m, nil
			}
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
			// Toggle proxy for selected tool or binding depending on view
			switch m.currentView {
			case viewDashboard:
				return m.handleToggleProxy()
			case viewBindings:
				return m.handleToggleBindingProxy()
			default:
				return m, nil
			}

		case "s":
			if m.currentView == viewSecrets {
				m.startSecretInput()
				return m, nil
			}
			// Check proxy status
			m.proxyStatus = proxyStatusChecking
			return m, checkProxyStatus

		case "e":
			if m.currentView == viewProviders {
				m.startProviderForm(false)
				return m, nil
			}
			if m.currentView == viewBindings {
				m.startBindingForm(false)
				return m, nil
			}
			return m, nil

		case "n":
			if m.currentView == viewBindings {
				m.startBindingForm(true)
				return m, nil
			}
			return m, nil

		case "a":
			if m.currentView == viewProviders {
				m.startProviderForm(true)
				return m, nil
			}
			return m, nil

		case "t":
			if m.currentView == viewSecrets {
				m.handleSecretTest()
				return m, nil
			}
			return m, nil

		case "/":
			if m.supportsSearch(m.currentView) {
				m.activateSearch()
			}
			return m, nil

		case "?":
			m.showHelpOverlay = !m.showHelpOverlay
			return m, nil

		case "esc":
			if m.showHelpOverlay {
				m.showHelpOverlay = false
				return m, nil
			}
			if m.viewHasSearch(m.currentView) {
				m.clearSearch()
				return m, nil
			}
			return m, nil

		case "up", "k":
			// Navigate up in list views
			if m.currentView != viewDashboard && m.selectedIndex > 0 {
				m.selectedIndex--
			}
			return m, nil

		case "down", "j":
			// Navigate down in list views
			maxIndex := m.maxSelectableIndex()
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

func (m DashboardModel) maxSelectableIndex() int {
	switch m.currentView {
	case viewProviders:
		return len(m.filteredProviderIndexes()) - 1
	case viewTools:
		return len(m.filteredToolIndexes()) - 1
	case viewBindings:
		return len(m.filteredBindingIndexes()) - 1
	case viewSecrets:
		return len(m.filteredProviderIndexes()) - 1 // Secrets are per-provider
	case viewSuggestions:
		return len(m.suggestions) - 1
	case viewReports:
		return len(reportOptions) - 1
	case viewConfig:
		return len(configFiles) - 1
	default:
		return 0
	}
}

func (m *DashboardModel) viewHasSearch(view viewMode) bool {
	return strings.TrimSpace(m.searchQuery) != "" && m.searchContextView == view
}

func (m *DashboardModel) filteredProviderIndexes() []int {
	if m.providers == nil {
		return nil
	}
	total := len(m.providers.Providers)
	indexes := make([]int, 0, total)
	query := strings.ToLower(strings.TrimSpace(m.searchQuery))
	for i, provider := range m.providers.Providers {
		if !m.viewHasSearch(viewProviders) ||
			strings.Contains(strings.ToLower(provider.DisplayName), query) ||
			strings.Contains(strings.ToLower(provider.ID), query) ||
			strings.Contains(strings.ToLower(provider.BaseURL), query) {
			indexes = append(indexes, i)
		}
	}
	return indexes
}

func (m *DashboardModel) filteredToolIndexes() []int {
	if m.tools == nil {
		return nil
	}
	total := len(m.tools.Tools)
	indexes := make([]int, 0, total)
	query := strings.ToLower(strings.TrimSpace(m.searchQuery))
	for i, tool := range m.tools.Tools {
		if !m.viewHasSearch(viewTools) ||
			strings.Contains(strings.ToLower(tool.Name), query) ||
			strings.Contains(strings.ToLower(tool.ID), query) ||
			strings.Contains(strings.ToLower(tool.Exec), query) {
			indexes = append(indexes, i)
		}
	}
	return indexes
}

func (m *DashboardModel) filteredBindingIndexes() []int {
	if m.bindings == nil {
		return nil
	}
	total := len(m.bindings.Bindings)
	indexes := make([]int, 0, total)
	query := strings.ToLower(strings.TrimSpace(m.searchQuery))
	for i, binding := range m.bindings.Bindings {
		if !m.viewHasSearch(viewBindings) ||
			strings.Contains(strings.ToLower(binding.ToolID), query) ||
			strings.Contains(strings.ToLower(binding.ProviderID), query) {
			indexes = append(indexes, i)
		}
	}
	return indexes
}

func (m DashboardModel) renderSearchBar(view viewMode) string {
	if !m.supportsSearch(view) {
		return ""
	}
	style := lipgloss.NewStyle().Foreground(m.theme.Muted).Padding(0, 2)
	switch {
	case m.searchActive && m.searchContextView == view:
		return style.Render("Search: " + m.searchInput.View())
	case m.viewHasSearch(view):
		return style.Render(fmt.Sprintf("Filter: %s  (Esc to clear)", m.searchQuery))
	default:
		return style.Render("Press '/' to search")
	}
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
		columns[0].Width = 15 // Tool
		columns[1].Width = 25 // Provider
		columns[2].Width = 28 // Model
		columns[3].Width = 10 // Proxy
		columns[4].Width = 15 // Status
	} else if m.width < 80 {
		columns[0].Width = 10 // Tool
		columns[1].Width = 18 // Provider
		columns[2].Width = 20 // Model
		columns[3].Width = 8  // Proxy
		columns[4].Width = 12 // Status
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
	proxyState := proxyStateOff
	if binding.UseProxy {
		proxyState = proxyStateOn
	}
	m.message = fmt.Sprintf("Proxy %s for %s", proxyState, tool.Name)

	return m, nil
}

// handleToggleBindingProxy toggles proxy usage for the selected binding in the bindings view
func (m DashboardModel) handleToggleBindingProxy() (tea.Model, tea.Cmd) {
	indexes := m.filteredBindingIndexes()

	if len(indexes) == 0 {
		m.message = "No bindings configured"
		return m, nil
	}

	if m.selectedIndex < 0 || m.selectedIndex >= len(indexes) {
		m.message = "No binding selected"
		return m, nil
	}

	binding := &m.bindings.Bindings[indexes[m.selectedIndex]]

	toolName := binding.ToolID
	if tool, err := m.tools.FindTool(binding.ToolID); err == nil {
		toolName = tool.Name
	}

	binding.UseProxy = !binding.UseProxy

	if err := core.SaveBindings(m.home, m.bindings); err != nil {
		m.message = fmt.Sprintf("Failed to save binding: %v", err)
		return m, nil
	}

	proxyState := proxyStateOff
	if binding.UseProxy {
		proxyState = proxyStateOn
	}

	// Update dashboard table rows to keep views consistent
	m.table.SetRows(m.buildTableRows())
	m.message = fmt.Sprintf("Proxy %s for %s", proxyState, toolName)

	return m, nil
}

// View renders the dashboard
func (m DashboardModel) View() string {
	if m.quitting {
		return ""
	}

	if m.showHelpOverlay {
		return m.renderHelpView()
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
	case viewReports:
		return m.renderReportsView()
	case viewHooks:
		return m.renderHooksView()
	case viewConfig:
		return m.renderConfigView()
	case viewHelp:
		return m.renderHelpView()
	default:
		return m.renderDashboardView()
	}
}

// renderDashboardView renders the main dashboard view
func (m DashboardModel) renderDashboardView() string {
	proxyIcon := iconCircleEmpty
	proxyText := "Checking..."
	switch m.proxyStatus {
	case proxyStatusRunning:
		proxyIcon = iconCircleFilled
		proxyText = "Running"
	case proxyStatusStopped:
		proxyIcon = iconCircleEmpty
		proxyText = "Stopped"
	}

	props := dashboardviews.DashboardViewProps{
		Theme:          newDashboardViewTheme(m.theme),
		TableView:      m.table.View(),
		Message:        m.message,
		ProxyIcon:      proxyIcon,
		ProxyStatus:    proxyText,
		NavigationHelp: helpTextNavigation,
		HelpCommands:   "[R] Run Tool  [X] Toggle Proxy",
	}

	return dashboardviews.RenderDashboardView(props)
}

// renderStatsView renders the usage statistics view
func (m DashboardModel) renderStatsView() string {
	props := dashboardviews.StatsViewProps{
		Theme:          newDashboardViewTheme(m.theme),
		Loaded:         m.statsLoaded,
		Error:          m.statsError,
		LoadingMessage: "Loading stats...",
		Today:          convertStatsSummaryToView("📅 Today's Usage", m.todayStats, false),
		Week:           convertStatsSummaryToView("📊 Last 7 Days", m.weekStats, true),
		Profiles:       convertProfileStatsToView(m.profileStats),
		NavigationHelp: "[V] Back to Dashboard  [S] Refresh  [Q] Quit",
		LoadingHelp:    "[V] Back to Dashboard  [Q] Quit",
		ProfileSubtitle: "🎯 By Profile (7d)",
	}

	return dashboardviews.RenderStatsView(props)
}

// renderProvidersView renders the AI providers management view
func (m DashboardModel) renderProvidersView() string {
	indexes := m.filteredProviderIndexes()
	if len(indexes) > 0 && m.selectedIndex >= len(indexes) {
		m.selectedIndex = len(indexes) - 1
	}

	props := dashboardviews.ProvidersViewProps{
		Theme:               newDashboardViewTheme(m.theme),
		ProviderForm:        m.renderProviderForm(),
		ShowProviderForm:    m.providerFormActive,
		ProviderFormMessage: strings.TrimSpace(m.providerFormMessage),
		SearchBar:           m.renderSearchBar(viewProviders),
		EmptyStateMessage:   providersEmptyStateMessage(len(indexes) == 0, m.viewHasSearch(viewProviders)),
		Providers:           convertProvidersToView(indexes, m.providers, m.secrets),
		SelectedIndex:       m.selectedIndex,
		Details:             convertProviderDetailsToView(indexes, m.selectedIndex, m.providers),
		NavigationHelp:      helpTextNavigation,
		HelpCommands:        "[E] Edit provider  [A] Add provider",
		EnabledIcon:         iconCheckmark,
		DisabledIcon:        iconCross,
		KeyPresentIcon:      "🔑",
		KeyMissingIcon:      "⚠",
	}

	return dashboardviews.RenderProvidersView(props)
}

func (m DashboardModel) renderProviderForm() string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.Primary).
		Padding(1, 2).
		Width(70)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.Primary)

	infoStyle := lipgloss.NewStyle().
		Foreground(m.theme.Muted)

	field := providerFieldSequence[m.providerFormField]
	title := "Edit Provider"
	if m.providerFormAdd {
		title = "Add Provider"
	}

	var currentName string
	if m.providerFormProvider.DisplayName != "" {
		currentName = fmt.Sprintf(" (%s)", m.providerFormProvider.DisplayName)
	}

	body := strings.Builder{}
	body.WriteString(titleStyle.Render(title + currentName))
	body.WriteString("\n\n")
	body.WriteString(infoStyle.Render(fmt.Sprintf("Field: %s", m.providerFieldPrompt(field))))
	body.WriteString("\n")
	body.WriteString(m.providerFormInput.View())
	body.WriteString("\n\n")
	body.WriteString(infoStyle.Render("Enter to confirm  •  Esc to cancel"))
	if strings.TrimSpace(m.providerFormMessage) != "" {
		body.WriteString("\n")
		body.WriteString(infoStyle.Render(m.providerFormMessage))
	}

	return boxStyle.Render(body.String())
}

func (m *DashboardModel) startBindingForm(add bool) {
	indexes := m.filteredBindingIndexes()
	if !add {
		if len(indexes) == 0 || m.selectedIndex < 0 || m.selectedIndex >= len(indexes) {
			m.bindingFormMessage = msgNoProviderSelected
			return
		}
		targetIdx := indexes[m.selectedIndex]
		if targetIdx < 0 || targetIdx >= len(m.bindings.Bindings) {
			m.bindingFormMessage = msgInvalidProvider
			return
		}
		m.bindingFormBinding = m.bindings.Bindings[targetIdx]
		m.bindingFormIndex = targetIdx
	} else {
		m.bindingFormBinding = core.Binding{
			Options: core.BindingOptions{},
		}
		m.bindingFormIndex = -1
	}

	m.bindingFormAdd = add
	m.bindingFormActive = true
	m.bindingFormField = 0
	if !add {
		m.bindingFormField = 1 // skip tool ID when editing
	}
	m.prepareBindingFormInput()
	m.bindingFormInput.Focus()
	m.bindingFormMessage = ""
	m.searchActive = false
}

func (m *DashboardModel) bindingFieldEnabled(field bindingFormField) bool {
	if !m.bindingFormAdd && field == bindingFieldToolID {
		return false
	}
	return true
}

func (m *DashboardModel) prepareBindingFormInput() {
	if m.bindingFormField >= len(bindingFieldSequence) {
		return
	}
	field := bindingFieldSequence[m.bindingFormField]
	m.bindingFormInput.Placeholder = m.bindingFieldPrompt(field)
	switch field {
	case bindingFieldToolID:
		m.bindingFormInput.SetValue(m.bindingFormBinding.ToolID)
	case bindingFieldProviderID:
		m.bindingFormInput.SetValue(m.bindingFormBinding.ProviderID)
	case bindingFieldModel:
		m.bindingFormInput.SetValue(m.bindingFormBinding.Options.Model)
	case bindingFieldUseProxy:
		if m.bindingFormBinding.UseProxy {
			m.bindingFormInput.SetValue("on")
		} else {
			m.bindingFormInput.SetValue("off")
		}
	}
}

func (m *DashboardModel) bindingFieldPrompt(field bindingFormField) string {
	switch field {
	case bindingFieldToolID:
		return "tool id (e.g. claude)"
	case bindingFieldProviderID:
		return "provider id (e.g. openai-official)"
	case bindingFieldModel:
		return "model override (optional)"
	case bindingFieldUseProxy:
		return "use proxy? (on/off)"
	default:
		return ""
	}
}

func (m *DashboardModel) submitBindingFormValue() {
	if m.bindingFormField >= len(bindingFieldSequence) {
		return
	}
	field := bindingFieldSequence[m.bindingFormField]
	value := strings.TrimSpace(m.bindingFormInput.Value())
	if err := m.setBindingFieldValue(field, value); err != nil {
		m.bindingFormMessage = err.Error()
		return
	}
	m.bindingFormMessage = ""
	m.bindingFormInput.SetValue("")
	for {
		m.bindingFormField++
		if m.bindingFormField >= len(bindingFieldSequence) {
			m.finishBindingForm()
			return
		}
		if m.bindingFieldEnabled(bindingFieldSequence[m.bindingFormField]) {
			m.prepareBindingFormInput()
			return
		}
	}
}

func (m *DashboardModel) setBindingFieldValue(field bindingFormField, value string) error {
	switch field {
	case bindingFieldToolID:
		if value == "" {
			return fmt.Errorf("tool id cannot be empty")
		}
		if _, err := m.tools.FindTool(value); err != nil {
			return fmt.Errorf("tool %s not found", value)
		}
		if _, err := m.bindings.FindBinding(value); err == nil {
			return fmt.Errorf("binding for %s already exists", value)
		}
		m.bindingFormBinding.ToolID = value
	case bindingFieldProviderID:
		if value == "" {
			return fmt.Errorf("provider id cannot be empty")
		}
		if _, err := m.providers.FindProvider(value); err != nil {
			return fmt.Errorf("provider %s not found", value)
		}
		m.bindingFormBinding.ProviderID = value
	case bindingFieldModel:
		m.bindingFormBinding.Options.Model = value
	case bindingFieldUseProxy:
		val := strings.ToLower(value)
		switch val {
		case "on", "true", "yes", "y":
			m.bindingFormBinding.UseProxy = true
		case "off", "false", "no", "n":
			m.bindingFormBinding.UseProxy = false
		default:
			return fmt.Errorf("use proxy must be on/off")
		}
	}
	return nil
}

func (m *DashboardModel) finishBindingForm() {
	if m.bindingFormBinding.ToolID == "" {
		m.bindingFormMessage = "tool id is required"
		return
	}
	if m.bindingFormBinding.ProviderID == "" {
		m.bindingFormMessage = "provider id is required"
		return
	}

	if m.bindingFormAdd {
		m.bindings.Bindings = append(m.bindings.Bindings, m.bindingFormBinding)
	} else if m.bindingFormIndex >= 0 && m.bindingFormIndex < len(m.bindings.Bindings) {
		m.bindings.Bindings[m.bindingFormIndex] = m.bindingFormBinding
	}

	if err := core.SaveBindings(m.home, m.bindings); err != nil {
		m.bindingFormMessage = fmt.Sprintf("failed to save binding: %v", err)
		return
	}

	m.table.SetRows(m.buildTableRows())
	if m.bindingFormAdd {
		m.bindingFormMessage = fmt.Sprintf("binding created for %s", m.bindingFormBinding.ToolID)
	} else {
		m.bindingFormMessage = fmt.Sprintf("binding updated for %s", m.bindingFormBinding.ToolID)
	}

	m.bindingFormActive = false
	m.bindingFormAdd = false
	m.bindingFormInput.Blur()
}

func (m DashboardModel) renderBindingForm() string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.Primary).
		Padding(1, 2).
		Width(70)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.Primary)

	infoStyle := lipgloss.NewStyle().
		Foreground(m.theme.Muted)

	field := bindingFieldSequence[m.bindingFormField]
	title := "Edit Binding"
	if m.bindingFormAdd {
		title = "Add Binding"
	}

	body := strings.Builder{}
	body.WriteString(titleStyle.Render(fmt.Sprintf("%s (%s)", title, m.bindingFormBinding.ToolID)))
	body.WriteString("\n\n")
	body.WriteString(infoStyle.Render(fmt.Sprintf("Field: %s", m.bindingFieldPrompt(field))))
	body.WriteString("\n")
	body.WriteString(m.bindingFormInput.View())
	body.WriteString("\n\n")
	body.WriteString(infoStyle.Render("Enter to confirm  •  Esc to cancel"))
	if strings.TrimSpace(m.bindingFormMessage) != "" {
		body.WriteString("\n")
		body.WriteString(infoStyle.Render(m.bindingFormMessage))
	}

	return boxStyle.Render(body.String())
}

// renderToolsView renders the CLI tools management view
func (m DashboardModel) renderToolsView() string {
	indexes := m.filteredToolIndexes()
	if len(indexes) > 0 && m.selectedIndex >= len(indexes) {
		m.selectedIndex = len(indexes) - 1
	}

	props := dashboardviews.ToolsViewProps{
		Theme:             newDashboardViewTheme(m.theme),
		SearchBar:         m.renderSearchBar(viewTools),
		EmptyStateMessage: toolsEmptyStateMessage(len(indexes) == 0, m.viewHasSearch(viewTools)),
		Tools:             convertToolsToView(indexes, m.tools, m.bindings),
		SelectedIndex:     m.selectedIndex,
		Details:           convertToolDetailsToView(m.selectedIndex, m.tools),
		NavigationHelp:    helpTextNavigation,
		BoundIcon:         iconCircleFilled,
		UnboundIcon:       iconCircleEmpty,
	}

	return dashboardviews.RenderToolsView(props)
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

	content.WriteString(m.renderBindingFormSection(mutedStyle))

	if searchBar := m.renderSearchBar(viewBindings); searchBar != "" {
		content.WriteString(searchBar)
		content.WriteString("\n\n")
	}

	indexes := m.filteredBindingIndexes()
	content.WriteString(m.renderBindingList(indexes, selectedStyle, normalStyle, mutedStyle))
	content.WriteString("\n")
	content.WriteString(m.renderBindingDetails(indexes, headerStyle, normalStyle))

	// Footer/Help
	helpText := helpTextNavigation + "  [E] Edit binding  [N] New binding  [X] Toggle Proxy"
	content.WriteString(helpStyle.Render(helpText))

	return content.String()
}

func (m DashboardModel) renderBindingFormSection(mutedStyle lipgloss.Style) string {
	var b strings.Builder
	if m.bindingFormActive {
		b.WriteString(m.renderBindingForm())
		b.WriteString("\n\n")
		return b.String()
	}
	if msg := strings.TrimSpace(m.bindingFormMessage); msg != "" {
		b.WriteString(mutedStyle.Render("  " + msg))
		b.WriteString("\n\n")
	}
	return b.String()
}

func (m *DashboardModel) renderBindingList(indexes []int, selectedStyle, normalStyle, mutedStyle lipgloss.Style) string {
	var b strings.Builder
	if len(indexes) == 0 {
		if m.viewHasSearch(viewBindings) {
			b.WriteString(mutedStyle.Render("  No bindings match the current filter."))
		} else {
			b.WriteString(mutedStyle.Render("  No bindings configured."))
		}
		b.WriteString("\n")
		return b.String()
	}

	if m.selectedIndex >= len(indexes) {
		m.selectedIndex = len(indexes) - 1
	}

	for displayIdx, bindingIdx := range indexes {
		binding := m.bindings.Bindings[bindingIdx]
		toolName := binding.ToolID
		if tool, err := m.tools.FindTool(binding.ToolID); err == nil {
			toolName = tool.Name
		}

		providerName := binding.ProviderID
		if provider, err := m.providers.FindProvider(binding.ProviderID); err == nil {
			providerName = provider.DisplayName
		}

		proxyIcon := iconCircleEmpty
		if binding.UseProxy {
			proxyIcon = iconCircleFilled
		}

		line := fmt.Sprintf("  %-15s → %-25s  Proxy: %s", toolName, providerName, proxyIcon)

		if displayIdx == m.selectedIndex {
			b.WriteString(selectedStyle.Render("▶ " + line))
		} else {
			b.WriteString(normalStyle.Render("  " + line))
		}
		b.WriteString("\n")
	}
	return b.String()
}

func (m DashboardModel) renderBindingDetails(indexes []int, headerStyle, normalStyle lipgloss.Style) string {
	if len(indexes) == 0 || m.selectedIndex < 0 || m.selectedIndex >= len(indexes) {
		return ""
	}

	var b strings.Builder
	binding := m.bindings.Bindings[indexes[m.selectedIndex]]
	b.WriteString(headerStyle.Render("Details"))
	b.WriteString("\n")
	b.WriteString(normalStyle.Render(fmt.Sprintf("  Tool ID: %s", binding.ToolID)))
	b.WriteString("\n")
	b.WriteString(normalStyle.Render(fmt.Sprintf("  Provider ID: %s", binding.ProviderID)))
	b.WriteString("\n")
	b.WriteString(normalStyle.Render(fmt.Sprintf("  Use Proxy: %t", binding.UseProxy)))
	b.WriteString("\n")
	if binding.Options.Model != "" {
		b.WriteString(normalStyle.Render(fmt.Sprintf("  Model Override: %s", binding.Options.Model)))
		b.WriteString("\n")
	}
	b.WriteString("\n")
	return b.String()
}

// renderSecretsView renders the API keys/secrets management view
func (m DashboardModel) renderSecretsView() string {
	searchBar := m.renderSearchBar(viewSecrets)
	indexes := m.filteredProviderIndexes()

	if len(indexes) == 0 {
		m.selectedIndex = 0
	} else if m.selectedIndex >= len(indexes) {
		m.selectedIndex = len(indexes) - 1
	}

	secretForm := ""
	if m.secretInputActive {
		secretForm = m.renderSecretInputForm()
	}

	props := dashboardviews.SecretsViewProps{
		Theme:             newDashboardViewTheme(m.theme),
		SecretForm:        secretForm,
		ShowSecretForm:    m.secretInputActive,
		SearchBar:         searchBar,
		EmptyStateMessage: secretsEmptyStateMessage(len(indexes) == 0, m.viewHasSearch(viewSecrets)),
		Providers:         convertSecretProvidersToView(indexes, m.providers, m.secrets),
		SelectedIndex:     m.selectedIndex,
		SecretMessage:     strings.TrimSpace(m.secretMessage),
		NavigationHelp:    helpTextNavigation,
		HelpCommands:      "[S] Set  [R] Remove  [T] Test",
		SuccessIcon:       iconCheckmark,
		FailureIcon:       iconCross,
	}

	return dashboardviews.RenderSecretsView(props)
}

func (m DashboardModel) renderSecretInputForm() string {
	if !m.secretInputActive || m.secretTargetIndex < 0 || m.secretTargetIndex >= len(m.providers.Providers) {
		return ""
	}

	provider := m.providers.Providers[m.secretTargetIndex]
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.Primary).
		Padding(1, 2).
		Width(60)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.Primary)

	infoStyle := lipgloss.NewStyle().
		Foreground(m.theme.Muted)

	body := strings.Builder{}
	body.WriteString(titleStyle.Render(fmt.Sprintf("Set API key for %s", provider.DisplayName)))
	body.WriteString("\n\n")
	body.WriteString(m.secretInput.View())
	body.WriteString("\n\n")
	body.WriteString(infoStyle.Render("Enter to save  •  Esc to cancel"))

	return boxStyle.Render(body.String())
}

// renderProxyView renders the proxy server control panel
func (m DashboardModel) renderProxyView() string {
	props := dashboardviews.ProxyViewProps{
		Theme:           newDashboardViewTheme(m.theme),
		StatusState:     m.proxyStatus,
		Address:         proxy.DefaultAddr,
		NavigationHelp:  helpTextNavigation,
		CommandHelpLine: "  [S] Refresh Status",
	}

	switch m.proxyStatus {
	case proxyStatusRunning:
		props.StatusIcon = iconCircleFilled
		props.StatusText = "Running"
		props.ShowConfig = true
	case proxyStatusStopped:
		props.StatusIcon = iconCircleEmpty
		props.StatusText = "Stopped"
		props.AdditionalNote = "  Note: Use 'boba proxy serve' in terminal to start the proxy server"
	default:
		props.StatusIcon = "⋯"
		props.StatusText = "Checking..."
	}

	return dashboardviews.RenderProxyView(props)
}

// renderRoutingView renders the routing rules tester
func (m DashboardModel) renderRoutingView() string {
	props := dashboardviews.RoutingViewProps{
		Theme:           newDashboardViewTheme(m.theme),
		NavigationHelp:  helpTextNavigation,
		CommandHelpLine: "  Use CLI: boba route test <text|@file>",
	}

	return dashboardviews.RenderRoutingView(props)
}

// renderSuggestionsView renders the optimization suggestions view
func (m DashboardModel) renderSuggestionsView() string {
	selectedIndex := m.selectedIndex
	if len(m.suggestions) > 0 && selectedIndex >= len(m.suggestions) {
		selectedIndex = 0
	}

	props := dashboardviews.SuggestionsViewProps{
		Theme:           newDashboardViewTheme(m.theme),
		Suggestions:     convertSuggestionsToView(m.suggestions),
		SelectedIndex:   selectedIndex,
		Error:           m.suggestionsError,
		NavigationHelp:  helpTextNavigation,
		CommandHelpLine: "  Use CLI: boba action [--auto] to apply suggestions",
	}

	return dashboardviews.RenderSuggestionsView(props)
}

// loadSuggestions loads optimization suggestions
func (m *DashboardModel) loadSuggestions() tea.Msg {
	dbPath := filepath.Join(m.home, "usage.db")
	db, err := sqlite.Open(dbPath)
	if err != nil {
		return suggestionsLoadedMsg{err: err}
	}

	engine := suggestions.NewEngine(db)
	suggs, err := engine.GenerateSuggestions(7)
	if err != nil {
		return suggestionsLoadedMsg{err: err}
	}

	return suggestionsLoadedMsg{suggestions: suggs}
}

// renderReportsView renders the report generation interface
func (m DashboardModel) renderReportsView() string {
	if m.selectedIndex >= len(reportOptions) {
		m.selectedIndex = 0
	}

	props := dashboardviews.ReportsViewProps{
		Theme:           newDashboardViewTheme(m.theme),
		Options:         convertReportOptionsToView(reportOptions),
		SelectedIndex:   m.selectedIndex,
		Home:            m.home,
		NavigationHelp:  helpTextNavigation,
		CommandHelpLine: "  Use CLI: boba report --format <json|csv|html> --days <N> --out <file>",
	}

	return dashboardviews.RenderReportsView(props)
}

// renderHooksView renders the Git hooks management interface
func (m DashboardModel) renderHooksView() string {
	repoPath := "(Not in a git repository)"
	hooksInstalled := false

	props := dashboardviews.HooksViewProps{
		Theme:          newDashboardViewTheme(m.theme),
		RepoPath:       repoPath,
		HooksInstalled: hooksInstalled,
		Hooks: []dashboardviews.HookInfo{
			{Name: "post-checkout", Desc: "Track branch switches and suggest optimal profiles", Active: hooksInstalled},
			{Name: "post-commit", Desc: "Record commit events for usage tracking", Active: hooksInstalled},
			{Name: "post-merge", Desc: "Track merge events and repository changes", Active: hooksInstalled},
		},
		NavigationHelp:  helpTextNavigation,
		CommandHelpLine: "  Use CLI: boba hooks install (to install hooks)  |  boba hooks remove (to uninstall)",
		ActiveIcon:      iconCheckmark,
		InactiveIcon:    iconCross,
	}

	return dashboardviews.RenderHooksView(props)
}

func (m DashboardModel) renderConfigView() string {
	props := dashboardviews.ConfigViewProps{
		Theme:              newDashboardViewTheme(m.theme),
		SelectedIndex:      m.selectedIndex,
		ConfigFiles:        convertConfigFilesToView(configFiles),
		Home:               m.home,
		HelpTextNavigation: helpTextNavigation,
	}

	return dashboardviews.RenderConfigView(props)
}

func (m DashboardModel) renderHelpView() string {
	props := dashboardviews.HelpViewProps{
		Theme:          newDashboardViewTheme(m.theme),
		Sections:       convertSectionsToView(m.sections),
		NavigationHelp: helpTextNavigation,
	}

	return dashboardviews.RenderHelpView(props)
}

func newDashboardViewTheme(theme Theme) dashboardviews.ThemePalette {
	return dashboardviews.ThemePalette{
		Primary: theme.Primary,
		Success: theme.Success,
		Danger:  theme.Danger,
		Warning: theme.Warning,
		Text:    theme.Text,
		Muted:   theme.Muted,
	}
}

func convertSuggestionsToView(suggs []suggestions.Suggestion) []dashboardviews.Suggestion {
	result := make([]dashboardviews.Suggestion, len(suggs))
	for i, sugg := range suggs {
		result[i] = dashboardviews.Suggestion{
			Title:       sugg.Title,
			Description: sugg.Description,
			Impact:      sugg.Impact,
			ActionItems: append([]string(nil), sugg.ActionItems...),
			Priority:    sugg.Priority,
			Type:        suggestionTypeToView(sugg.Type),
		}
	}
	return result
}

func convertSecretProvidersToView(indexes []int, providers *core.ProvidersConfig, secretStore *core.SecretsConfig) []dashboardviews.SecretProviderRow {
	if providers == nil || secretStore == nil || len(indexes) == 0 {
		return nil
	}

	result := make([]dashboardviews.SecretProviderRow, 0, len(indexes))
	for _, idx := range indexes {
		if idx < 0 || idx >= len(providers.Providers) {
			continue
		}

		provider := providers.Providers[idx]

		hasKey := false
		keySource := "(not set)"
		if _, err := core.ResolveAPIKey(&provider, secretStore); err == nil {
			hasKey = true
			keySource = string(provider.APIKey.Source)
		}

		result = append(result, dashboardviews.SecretProviderRow{
			DisplayName: provider.DisplayName,
			HasKey:      hasKey,
			KeySource:   keySource,
		})
	}

	return result
}

func secretsEmptyStateMessage(isEmpty bool, hasSearch bool) string {
	if !isEmpty {
		return ""
	}
	if hasSearch {
		return "No providers match the current filter."
	}
	return "No providers configured."
}

func convertStatsSummaryToView(title string, summary stats.Summary, includeAverages bool) dashboardviews.StatsSummary {
	return dashboardviews.StatsSummary{
		Title:          title,
		Tokens:         summary.TotalTokens,
		Cost:           summary.TotalCost,
		Sessions:       summary.TotalSessions,
		AvgDailyTokens: summary.AvgDailyTokens,
		AvgDailyCost:   summary.AvgDailyCost,
		ShowAverages:   includeAverages,
	}
}

func convertProfileStatsToView(statsList []stats.ProfileStats) []dashboardviews.StatsProfile {
	if len(statsList) == 0 {
		return nil
	}

	result := make([]dashboardviews.StatsProfile, 0, len(statsList))
	for _, ps := range statsList {
		result = append(result, dashboardviews.StatsProfile{
			Name:       ps.ProfileName,
			Tokens:     ps.TotalTokens,
			Cost:       ps.TotalCost,
			Sessions:   ps.SessionCount,
			AvgLatency: ps.AvgLatencyMS,
			UsagePct:   ps.UsagePercent,
			CostPct:    ps.CostPercent,
		})
	}
	return result
}

func providersEmptyStateMessage(isEmpty bool, hasSearch bool) string {
	if !isEmpty {
		return ""
	}
	if hasSearch {
		return "No providers match the current filter."
	}
	return "No providers configured."
}

func toolsEmptyStateMessage(isEmpty bool, hasSearch bool) string {
	if !isEmpty {
		return ""
	}
	if hasSearch {
		return "No tools match the current filter."
	}
	return "No tools configured."
}

func suggestionTypeToView(t suggestions.SuggestionType) string {
	switch t {
	case suggestions.SuggestionCostOptimization:
		return "cost"
	case suggestions.SuggestionProfileSwitch:
		return "profile"
	case suggestions.SuggestionBudgetAdjust:
		return "budget"
	case suggestions.SuggestionAnomaly:
		return "anomaly"
	case suggestions.SuggestionUsagePattern:
		return "usage"
	default:
		return "usage"
	}
}

func convertReportOptionsToView(options []reportOption) []dashboardviews.ReportOption {
	result := make([]dashboardviews.ReportOption, len(options))
	for i, opt := range options {
		result[i] = dashboardviews.ReportOption{
			Label: opt.label,
			Desc:  opt.desc,
		}
	}
	return result
}

func convertConfigFilesToView(files []configFile) []dashboardviews.ConfigFile {
	result := make([]dashboardviews.ConfigFile, len(files))
	for i, cfg := range files {
		result[i] = dashboardviews.ConfigFile{
			Name: cfg.name,
			File: cfg.file,
			Desc: cfg.desc,
		}
	}
	return result
}

func convertSectionsToView(sections []viewSection) []dashboardviews.HelpSection {
	result := make([]dashboardviews.HelpSection, 0, len(sections))
	for _, section := range sections {
		viewNames := make([]string, 0, len(section.views))
		for _, v := range section.views {
			viewNames = append(viewNames, viewName(v))
		}
		result = append(result, dashboardviews.HelpSection{
			Name:     section.name,
			Shortcut: section.shortcut,
			Views:    viewNames,
		})
	}
	return result
}

func convertProvidersToView(indexes []int, providers *core.ProvidersConfig, secrets *core.SecretsConfig) []dashboardviews.ProviderRow {
	if providers == nil || len(indexes) == 0 {
		return nil
	}

	result := make([]dashboardviews.ProviderRow, 0, len(indexes))
	for _, idx := range indexes {
		if idx < 0 || idx >= len(providers.Providers) {
			continue
		}

		provider := providers.Providers[idx]
		hasKey := false
		if secrets != nil {
			if _, err := core.ResolveAPIKey(&provider, secrets); err == nil {
				hasKey = true
			}
		}

		result = append(result, dashboardviews.ProviderRow{
			DisplayName:  provider.DisplayName,
			BaseURL:      provider.BaseURL,
			DefaultModel: provider.DefaultModel,
			Enabled:      provider.Enabled,
			HasAPIKey:    hasKey,
		})
	}

	return result
}

func convertProviderDetailsToView(indexes []int, selectedIndex int, providers *core.ProvidersConfig) *dashboardviews.ProviderDetails {
	if providers == nil || len(indexes) == 0 || selectedIndex < 0 || selectedIndex >= len(indexes) {
		return nil
	}

	idx := indexes[selectedIndex]
	if idx < 0 || idx >= len(providers.Providers) {
		return nil
	}

	provider := providers.Providers[idx]
	details := dashboardviews.ProviderDetails{
		ID:           provider.ID,
		Kind:         string(provider.Kind),
		APIKeySource: string(provider.APIKey.Source),
	}

	if provider.APIKey.Source == core.APIKeySourceEnv && provider.APIKey.EnvVar != "" {
		details.EnvVar = provider.APIKey.EnvVar
		details.ShowEnvVar = true
	}

	return &details
}

func convertToolsToView(indexes []int, tools *core.ToolsConfig, bindings *core.BindingsConfig) []dashboardviews.ToolRow {
	if tools == nil || len(indexes) == 0 {
		return nil
	}

	result := make([]dashboardviews.ToolRow, 0, len(indexes))
	for _, idx := range indexes {
		if idx < 0 || idx >= len(tools.Tools) {
			continue
		}

		tool := tools.Tools[idx]
		bound := false
		if bindings != nil {
			if _, err := bindings.FindBinding(tool.ID); err == nil {
				bound = true
			}
		}

		result = append(result, dashboardviews.ToolRow{
			Name:  tool.Name,
			Exec:  tool.Exec,
			Kind:  string(tool.Kind),
			Bound: bound,
		})
	}

	return result
}

func convertToolDetailsToView(selectedIndex int, tools *core.ToolsConfig) *dashboardviews.ToolDetails {
	if tools == nil || selectedIndex < 0 || selectedIndex >= len(tools.Tools) {
		return nil
	}

	tool := tools.Tools[selectedIndex]
	return &dashboardviews.ToolDetails{
		ID:          tool.ID,
		ConfigType:  string(tool.ConfigType),
		ConfigPath:  tool.ConfigPath,
		Description: tool.Description,
	}
}

func viewName(view viewMode) string {
	switch view {
	case viewDashboard:
		return "Dashboard"
	case viewProviders:
		return "Providers"
	case viewTools:
		return "Tools"
	case viewBindings:
		return "Bindings"
	case viewSecrets:
		return "Secrets"
	case viewStats:
		return "Usage Stats"
	case viewProxy:
		return "Proxy"
	case viewRouting:
		return "Routing Tester"
	case viewSuggestions:
		return "Suggestions"
	case viewReports:
		return "Reports"
	case viewHooks:
		return "Hooks"
	case viewConfig:
		return "Config Editor"
	case viewHelp:
		return "Help"
	default:
		return "View"
	}
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
