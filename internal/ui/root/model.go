// Package root provides the root UI model and orchestration for the BobaMixer TUI.
package root

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/royisme/bobamixer/internal/domain/core"
	"github.com/royisme/bobamixer/internal/domain/stats"
	"github.com/royisme/bobamixer/internal/domain/suggestions"
	"github.com/royisme/bobamixer/internal/proxy"
	"github.com/royisme/bobamixer/internal/settings"
	bindingsvc "github.com/royisme/bobamixer/internal/ui/features/bindings"
	configsvc "github.com/royisme/bobamixer/internal/ui/features/config"
	dashboardsvc "github.com/royisme/bobamixer/internal/ui/features/dashboard"
	helpsvc "github.com/royisme/bobamixer/internal/ui/features/help"
	hookssvc "github.com/royisme/bobamixer/internal/ui/features/hooks"
	providersvc "github.com/royisme/bobamixer/internal/ui/features/providers"
	proxysvc "github.com/royisme/bobamixer/internal/ui/features/proxy"
	reportsvc "github.com/royisme/bobamixer/internal/ui/features/reports"
	routingsvc "github.com/royisme/bobamixer/internal/ui/features/routing"
	"github.com/royisme/bobamixer/internal/ui/features/secrets"
	statssvc "github.com/royisme/bobamixer/internal/ui/features/stats"
	suggestionssvc "github.com/royisme/bobamixer/internal/ui/features/suggestions"
	toolsvc "github.com/royisme/bobamixer/internal/ui/features/tools"
	"github.com/royisme/bobamixer/internal/ui/forms"
	"github.com/royisme/bobamixer/internal/ui/i18n"
	"github.com/royisme/bobamixer/internal/ui/theme"
)

const (
	keyCtrlC = "ctrl+c"
	keyEsc   = "esc"
	keyEnter = "enter"
)

// UI constants for repeated strings
const (
	promptPrefix = "│ "
)

// DashboardModel represents the control plane dashboard
type DashboardModel struct {
	home      string
	theme     theme.Theme
	styles    theme.Styles
	localizer *i18n.Localizer

	// Data
	providers  *core.ProvidersConfig
	tools      *core.ToolsConfig
	bindings   *core.BindingsConfig
	secrets    *core.SecretsConfig
	themes     []string
	themeIndex int

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
	currentView        viewMode
	selectedIndex      int // Currently selected item in list views
	width              int
	height             int
	quitting           bool
	proxyStatus        string // proxysvc.StatusRunning, StatusStopped, StatusChecking
	message            string // Status message to display
	sections           []viewSection
	currentSection     int
	sectionViewIndex   int
	showHelpOverlay    bool
	searchActive       bool
	searchInput        textinput.Model
	searchQuery        string
	searchContextView  viewMode
	configActiveTab    int // 0: Files, 1: Appearance, 2: System
	secretMessage      string
	providerForm       *forms.ProviderForm
	bindingForm        *forms.BindingForm
	secretForm         *forms.SecretForm
	toolsService       *toolsvc.Service
	reportsService     *reportsvc.Service
	proxyService       *proxysvc.Service
	bindingService     *bindingsvc.Service
	providerService    *providersvc.Service
	secretService      *secrets.Service
	statsService       *statssvc.Service
	suggestionsService *suggestionssvc.Service
	dashboardService   *dashboardsvc.Service
	routingService     *routingsvc.Service
	configService      *configsvc.Service
	hooksService       *hookssvc.Service
	helpService        *helpsvc.Service
}

// NewDashboard creates a new dashboard model
func NewDashboard(home string) (*DashboardModel, error) {
	// Load theme and localizer
	palette := loadTheme(home)
	localizer, err := i18n.NewLocalizer(i18n.GetUserLanguage())
	if err != nil {
		// Fallback to English if user language is not available
		localizer, err = i18n.NewLocalizer("en")
		if err != nil {
			// Should not happen with English, but handle it
			return nil, fmt.Errorf("failed to load localizer: %w", err)
		}
	}

	// Load all configurations
	providers, tools, bindings, secretsConfig, err := core.LoadAll(home)
	if err != nil {
		return nil, fmt.Errorf("failed to load configurations: %w", err)
	}

	m := &DashboardModel{
		home:        home,
		theme:       palette,
		styles:      theme.NewStyles(palette),
		localizer:   localizer,
		providers:   providers,
		tools:       tools,
		bindings:    bindings,
		secrets:     secretsConfig,
		proxyStatus: proxysvc.StatusChecking,
		currentView: viewDashboard,
		themes:      []string{"auto", "catppuccin", "dracula"},
	}

	// Set initial theme index
	currentTheme := settings.DefaultSettings().Theme
	if userSettings, err := settings.Load(context.Background(), home); err == nil {
		currentTheme = userSettings.Theme
	}
	for i, t := range m.themes {
		if t == currentTheme {
			m.themeIndex = i
			break
		}
	}

	m.initSections()
	searchInput := textinput.New()
	searchInput.Placeholder = "Search..."
	searchInput.CharLimit = 100
	searchInput.Width = 30
	m.searchInput = searchInput
	providerForm := forms.NewProviderForm(promptPrefix)
	m.providerForm = &providerForm
	bindingForm := forms.NewBindingForm(promptPrefix)
	m.bindingForm = &bindingForm
	secretForm := forms.NewSecretForm(promptPrefix)
	m.secretForm = &secretForm

	m.toolsService = toolsvc.NewService(m.tools, m.bindings)
	m.reportsService = reportsvc.NewService()
	m.proxyService = proxysvc.NewService(proxy.DefaultAddr)
	m.bindingService = bindingsvc.NewService(
		m.bindings,
		m.tools,
		m.providers,
		m.bindingForm,
		dashboardsvc.MsgNoProviderSelected,
		dashboardsvc.MsgInvalidProvider,
	)
	m.providerService = providersvc.NewService(
		m.providers,
		m.secrets,
		m.providerForm,
		dashboardsvc.MsgNoProviderSelected,
		dashboardsvc.MsgInvalidProvider,
	)
	m.secretService = secrets.NewService(
		m.providers,
		&m.secrets,
		m.secretForm,
		&m.secretMessage,
		dashboardsvc.MsgNoProviderSelected,
		dashboardsvc.MsgInvalidProvider,
	)
	m.statsService = statssvc.NewService(home)
	m.suggestionsService = suggestionssvc.NewService(home)
	m.dashboardService = dashboardsvc.NewService(m.tools, m.bindings, m.providers, m.secrets)
	m.routingService = routingsvc.NewService()
	m.configService = configsvc.NewService()
	m.hooksService = hookssvc.NewService()
	m.helpService = helpsvc.NewService()

	m.initializeTable()

	return m, nil
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

func loadTheme(home string) theme.Theme {
	ctx := context.Background()

	userSettings, err := settings.Load(ctx, home)
	if err != nil {
		return theme.GetTheme(settings.DefaultSettings().Theme)
	}

	themeName := userSettings.Theme
	if themeName == "" {
		themeName = settings.DefaultSettings().Theme
	}

	return theme.GetTheme(themeName)
}
