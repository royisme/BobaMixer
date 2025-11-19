// Package root provides the root UI model and orchestration for the BobaMixer TUI.
package root

import (
	"fmt"
	"strings"

	"github.com/royisme/bobamixer/internal/proxy"
	"github.com/royisme/bobamixer/internal/ui/components"
	dashboardsvc "github.com/royisme/bobamixer/internal/ui/features/dashboard"
	proxysvc "github.com/royisme/bobamixer/internal/ui/features/proxy"
	"github.com/royisme/bobamixer/internal/ui/features/secrets"
	statssvc "github.com/royisme/bobamixer/internal/ui/features/stats"
	"github.com/royisme/bobamixer/internal/ui/pages"
)

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
	proxyIcon := dashboardsvc.IconCircleEmpty
	proxyText := "Checking..."
	if m.proxyService != nil {
		viewData := m.proxyService.ViewData(m.proxyStatus)
		proxyIcon = viewData.StatusIcon
		proxyText = viewData.StatusText
	} else if m.proxyStatus == proxysvc.StatusStopped {
		proxyText = "Stopped"
	}

	page := pages.NewDashboardPage(m.theme, pages.DashboardPageProps{
		Title:          "BobaMixer - AI CLI Control Plane",
		Table:          m.table.View(),
		Message:        m.message,
		ProxyIcon:      proxyIcon,
		ProxyStatus:    proxyText,
		NavigationHelp: m.dashboardService.GetNavigationHelp(),
		ActionHelp:     m.dashboardService.GetActionHelp(),
	})

	return page.View()
}

// renderStatsView renders the usage statistics view
func (m DashboardModel) renderStatsView() string {
	viewData := m.statsService.ConvertToView(statssvc.StatsData{
		Today:        m.todayStats,
		Week:         m.weekStats,
		ProfileStats: m.profileStats,
	})

	props := pages.StatsPageProps{
		Title:           "BobaMixer - Usage Statistics",
		Loaded:          m.statsLoaded,
		Error:           m.statsError,
		LoadingMessage:  "Loading stats...",
		Today:           viewData.Today,
		Week:            viewData.Week,
		Profiles:        viewData.Profiles,
		NavigationHelp:  m.dashboardService.GetNavigationHelp(),
		LoadingHelp:     "[V] Back to Dashboard  [Q] Quit",
		ProfileSubtitle: "🎯 By Profile (7d)",
	}

	page := pages.NewStatsPage(m.theme, props)
	return page.View()
}

// renderProvidersView renders the AI providers management view
func (m DashboardModel) renderProvidersView() string {
	indexes := m.filteredProviderIndexes()
	if len(indexes) > 0 && m.selectedIndex >= len(indexes) {
		m.selectedIndex = len(indexes) - 1
	}

	var (
		providerRows    []components.ProviderRow
		providerDetails *components.ProviderDetails
		emptyState      string
	)
	if m.providerService != nil {
		providerRows = m.providerService.Rows(indexes)
		providerDetails = m.providerService.Details(indexes, m.selectedIndex)
		emptyState = m.providerService.EmptyStateMessage(len(indexes) == 0, m.viewHasSearch(viewProviders))
	}

	props := pages.ProvidersPageProps{
		Title:               "BobaMixer - AI Providers Management",
		SectionTitle:        "📡 Available Providers",
		DetailsTitle:        "Details",
		ProviderForm:        m.renderProviderForm(),
		ShowProviderForm:    m.providerForm.Active(),
		ProviderFormMessage: strings.TrimSpace(m.providerForm.Message()),
		SearchBar:           m.renderSearchBar(viewProviders),
		EmptyStateMessage:   emptyState,
		Providers:           providerRows,
		SelectedIndex:       m.selectedIndex,
		Details:             providerDetails,
		NavigationHelp:      m.dashboardService.GetNavigationHelp(),
		ActionHelp:          "[E] Edit provider  [A] Add provider",
		Icons: components.ProviderListIcons{
			Enabled:    dashboardsvc.IconCheckmark,
			Disabled:   dashboardsvc.IconCross,
			KeyPresent: "🔑",
			KeyMissing: "⚠",
		},
	}

	page := pages.NewProvidersPage(m.theme, props)
	return page.View()
}

func (m DashboardModel) renderProviderForm() string {
	return m.providerForm.View(m.theme, m.styles)
}

func (m DashboardModel) renderBindingForm() string {
	return m.bindingForm.View(m.theme, m.styles)
}

// renderToolsView renders the CLI tools management view
func (m DashboardModel) renderToolsView() string {
	indexes := m.filteredToolIndexes()
	if len(indexes) > 0 && m.selectedIndex >= len(indexes) {
		m.selectedIndex = len(indexes) - 1
	}

	var (
		toolRows    []components.ToolRow
		toolDetails *components.ToolDetails
		emptyState  string
	)
	if m.toolsService != nil {
		toolRows = m.toolsService.Rows(indexes)
		toolDetails = m.toolsService.Details(indexes, m.selectedIndex)
		emptyState = m.toolsService.EmptyStateMessage(len(indexes) == 0, m.viewHasSearch(viewTools))
	}

	props := pages.ToolsPageProps{
		Title:             "BobaMixer - CLI Tools Management",
		SectionTitle:      "🛠 Detected Tools",
		DetailsTitle:      "Details",
		SearchBar:         m.renderSearchBar(viewTools),
		EmptyStateMessage: emptyState,
		Tools:             toolRows,
		SelectedIndex:     m.selectedIndex,
		Details:           toolDetails,
		NavigationHelp:    m.dashboardService.GetNavigationHelp(),
		ActionHelp:        "[B] Bind tool  [R] Refresh tools",
		BoundIcon:         dashboardsvc.IconCircleFilled,
		UnboundIcon:       dashboardsvc.IconCircleEmpty,
	}

	page := pages.NewToolsPage(m.theme, props)
	return page.View()
}

// renderBindingsView renders the tool-to-provider bindings view
func (m DashboardModel) renderBindingsView() string {
	indexes := m.filteredBindingIndexes()
	if len(indexes) > 0 && m.selectedIndex >= len(indexes) {
		m.selectedIndex = len(indexes) - 1
	}

	var (
		bindingRows    []components.BindingRow
		bindingDetails *components.BindingDetails
		emptyState     string
	)
	if m.bindingService != nil {
		bindingRows = m.bindingService.Rows(indexes)
		bindingDetails = m.bindingService.Details(indexes, m.selectedIndex)
		emptyState = m.bindingService.EmptyStateMessage(len(indexes) == 0, m.viewHasSearch(viewBindings))
	}

	props := pages.BindingsPageProps{
		Title:              "BobaMixer - Tool ↔ Provider Bindings",
		SectionTitle:       "🔗 Active Bindings",
		DetailsTitle:       "Details",
		BindingForm:        m.renderBindingForm(),
		ShowBindingForm:    m.bindingForm.Active(),
		BindingFormMessage: strings.TrimSpace(m.bindingForm.Message()),
		SearchBar:          m.renderSearchBar(viewBindings),
		EmptyStateMessage:  emptyState,
		Bindings:           bindingRows,
		SelectedIndex:      m.selectedIndex,
		Details:            bindingDetails,
		NavigationHelp:     m.dashboardService.GetNavigationHelp(),
		ActionHelp:         "[E] Edit binding  [N] New binding  [X] Toggle Proxy",
		ProxyEnabledIcon:   dashboardsvc.IconCircleFilled,
		ProxyDisabledIcon:  dashboardsvc.IconCircleEmpty,
	}

	page := pages.NewBindingsPage(m.theme, props)
	return page.View()
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

	var secretRows []components.SecretProviderRow
	if m.secretService != nil {
		secretRows = m.secretService.Rows(indexes)
	}

	props := pages.SecretsPageProps{
		Title:             "BobaMixer - Secrets Management (API Keys)",
		StatusTitle:       "🔒 API Key Status",
		SecurityTitle:     "🔐 Security",
		SecretForm:        m.renderSecretForm(),
		ShowSecretForm:    m.secretForm.Active(),
		SearchBar:         searchBar,
		EmptyStateMessage: secrets.EmptyStateMessage(len(indexes) == 0, m.viewHasSearch(viewSecrets)),
		Providers:         secretRows,
		SelectedIndex:     m.selectedIndex,
		SecretMessage:     strings.TrimSpace(m.secretMessage),
		NavigationHelp:    m.dashboardService.GetNavigationHelp(),
		ActionHelp:        "[S] Set  [R] Remove  [T] Test",
		SuccessIcon:       dashboardsvc.IconCheckmark,
		FailureIcon:       dashboardsvc.IconCross,
		SecurityTips: []string{
			"API keys are stored encrypted in ~/.boba/secrets.yaml",
			"Keys can also be loaded from environment variables",
			"Use 'boba edit secrets' to manage keys manually",
		},
	}

	page := pages.NewSecretsPage(m.theme, props)
	return page.View()
}

func (m DashboardModel) renderSecretForm() string {
	return m.secretForm.View(m.styles, m.theme)
}

// renderProxyView renders the proxy server control panel
func (m DashboardModel) renderProxyView() string {
	viewData := proxysvc.ViewData{
		StatusIcon: "⋯",
		StatusText: "Checking...",
		InfoLines: []string{
			"The proxy server intercepts AI API requests from CLI tools",
			"and routes them through BobaMixer for tracking and control.",
		},
		ConfigLines: []string{
			fmt.Sprintf("Tools with proxy enabled automatically use HTTP_PROXY=%s", proxy.DefaultAddr),
			fmt.Sprintf("and HTTPS_PROXY=%s", proxy.DefaultAddr),
		},
		CommandHelp: "[S] Refresh Status",
		Address:     proxy.DefaultAddr,
	}
	if m.proxyService != nil {
		viewData = m.proxyService.ViewData(m.proxyStatus)
	}

	props := pages.ProxyPageProps{
		Title:           "BobaMixer - Proxy Server Control",
		StatusTitle:     "🌐 Proxy Status",
		InfoTitle:       "ℹ️  Information",
		ConfigTitle:     "📝 Configuration",
		StatusState:     m.proxyStatus,
		Address:         viewData.Address,
		NavigationHelp:  m.dashboardService.GetNavigationHelp(),
		CommandHelpLine: viewData.CommandHelp,
		InfoLines:       viewData.InfoLines,
		ConfigLines:     viewData.ConfigLines,
		StatusIcon:      viewData.StatusIcon,
		StatusText:      viewData.StatusText,
		AdditionalNote:  viewData.AdditionalNote,
		ShowConfig:      viewData.ShowConfig,
	}

	page := pages.NewProxyPage(m.theme, props)
	return page.View()
}

// renderRoutingView renders the routing rules tester
func (m DashboardModel) renderRoutingView() string {
	data := m.routingService.ViewData()
	props := pages.RoutingPageProps{
		Title:           data.Title,
		TestTitle:       data.TestTitle,
		HowToTitle:      data.HowToTitle,
		ExampleTitle:    data.ExampleTitle,
		ContextTitle:    data.ContextTitle,
		TestDescription: data.TestDescription,
		HowToSteps:      data.HowToSteps,
		ExampleLines:    data.ExampleLines,
		ContextLines:    data.ContextLines,
		NavigationHelp:  m.dashboardService.GetNavigationHelp(),
		CommandHelpLine: data.CommandHelpLine,
	}

	page := pages.NewRoutingPage(m.theme, props)
	return page.View()
}

// renderSuggestionsView renders the optimization suggestions view
func (m DashboardModel) renderSuggestionsView() string {
	selectedIndex := m.selectedIndex
	if len(m.suggestions) > 0 && selectedIndex >= len(m.suggestions) {
		selectedIndex = 0
	}

	props := pages.SuggestionsPageProps{
		Title:           "BobaMixer - Optimization Suggestions",
		SectionTitle:    "💡 Recommendations (Last 7 Days)",
		DetailsTitle:    "Details",
		Suggestions:     m.suggestionsService.ConvertToView(m.suggestions),
		SelectedIndex:   selectedIndex,
		Error:           m.suggestionsError,
		NavigationHelp:  m.dashboardService.GetNavigationHelp(),
		CommandHelpLine: m.suggestionsService.CommandHelp(),
	}

	page := pages.NewSuggestionsPage(m.theme, props)
	return page.View()
}

// renderReportsView renders the report generation interface
func (m DashboardModel) renderReportsView() string {
	optionCount := 0
	commandHelp := ""
	var options []components.ReportOption
	if m.reportsService != nil {
		optionCount = m.reportsService.OptionCount()
		options = m.reportsService.Options()
		commandHelp = m.reportsService.CommandHelp()
	}

	if optionCount > 0 && m.selectedIndex >= optionCount {
		m.selectedIndex = 0
	}

	props := pages.ReportsPageProps{
		Title:           "📊 Generate Usage Report",
		OptionsTitle:    "Report Options",
		OutputTitle:     "Output Configuration",
		ContentsTitle:   "Report Contents",
		Options:         options,
		SelectedIndex:   m.selectedIndex,
		Home:            m.home,
		NavigationHelp:  m.dashboardService.GetNavigationHelp(),
		CommandHelpLine: commandHelp,
	}

	page := pages.NewReportsPage(m.theme, props)
	return page.View()
}

// renderHooksView renders the Git hooks management interface
func (m DashboardModel) renderHooksView() string {
	repoPath := "(Not in a git repository)"
	hooksInstalled := false

	data := m.hooksService.ViewData()
	hooks := m.hooksService.GetAvailableHooks(hooksInstalled)
	props := pages.HooksPageProps{
		Title:           data.Title,
		RepoTitle:       data.RepoTitle,
		HooksTitle:      data.HooksTitle,
		BenefitsTitle:   data.BenefitsTitle,
		ActivityTitle:   data.ActivityTitle,
		RepoPath:        repoPath,
		HooksInstalled:  hooksInstalled,
		Hooks:           m.hooksService.ConvertToComponents(hooks),
		NavigationHelp:  m.dashboardService.GetNavigationHelp(),
		CommandHelpLine: data.CommandHelpLine,
		ActiveIcon:      dashboardsvc.IconCheckmark,
		InactiveIcon:    dashboardsvc.IconCross,
	}

	page := pages.NewHooksPage(m.theme, props)
	return page.View()
}

func (m DashboardModel) renderConfigView() string {
	currentTheme := m.themes[m.themeIndex]
	data := m.configService.ViewData(m.home, currentTheme)
	props := pages.ConfigPageProps{
		Title:           data.Title,
		ConfigTitle:     data.ConfigTitle,
		EditorTitle:     data.EditorTitle,
		SafetyTitle:     data.SafetyTitle,
		ThemeTitle:      data.ThemeTitle,
		ConfigFiles:     m.configService.ConvertToComponents(),
		SelectedIndex:   m.selectedIndex,
		Home:            data.Home,
		EditorName:      data.EditorName,
		NavigationHelp:  m.dashboardService.GetNavigationHelp(),
		CommandHelpLine: data.CommandHelpLine,
		Themes:          data.Themes,
		CurrentTheme:    data.CurrentTheme,
	}

	page := pages.NewConfigPage(m.theme, props)
	return page.View()
}

func (m DashboardModel) renderHelpView() string {
	page := pages.NewHelpPage(m.theme, m.helpPageProps())
	return page.View()
}

func (m DashboardModel) helpPageProps() pages.HelpPageProps {
	data := m.helpService.ViewData()
	tips := m.helpService.GetDefaultTips()
	links := m.helpService.GetDefaultLinks()

	return pages.HelpPageProps{
		Title:          data.Title,
		Subtitle:       data.Subtitle,
		Sections:       convertSectionsToComponents(m.sections),
		Shortcuts:      nil,
		Tips:           tips,
		Links:          m.helpService.ConvertLinksToComponents(links),
		NavigationHint: data.NavigationHint + "  |  " + m.dashboardService.GetNavigationHelp(),
	}
}

func convertSectionsToComponents(sections []viewSection) []components.HelpSection {
	result := make([]components.HelpSection, 0, len(sections))
	for _, section := range sections {
		viewNames := make([]string, 0, len(section.views))
		for _, v := range section.views {
			viewNames = append(viewNames, viewName(v))
		}
		result = append(result, components.HelpSection{
			Name:     section.name,
			Shortcut: section.shortcut,
			Views:    viewNames,
		})
	}
	return result
}
