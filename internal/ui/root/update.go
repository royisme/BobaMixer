// Package root provides the root UI model and orchestration for the BobaMixer TUI.
package root

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/royisme/bobamixer/internal/domain/core"
	"github.com/royisme/bobamixer/internal/settings"
	dashboardsvc "github.com/royisme/bobamixer/internal/ui/features/dashboard"
	proxysvc "github.com/royisme/bobamixer/internal/ui/features/proxy"
	"github.com/royisme/bobamixer/internal/ui/theme"
)

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
			m.proxyStatus = proxysvc.StatusRunning
		} else {
			m.proxyStatus = proxysvc.StatusStopped
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

		if m.providerForm.Active() {
			switch key {
			case keyEsc:
				m.providerForm.Cancel("Provider edit canceled")
				return m, nil
			case keyEnter:
				done, err := m.providerForm.Submit()
				if err != nil {
					return m, nil
				}
				if done {
					if m.providerService != nil {
						if err := m.providerService.Save(m.home); err != nil {
							m.providerForm.SetMessage(fmt.Sprintf("Error saving: %v", err))
						}
					}
				}
				return m, nil
			default:
				return m, m.providerForm.Update(msg)
			}
		}

		if m.bindingForm.Active() {
			switch key {
			case keyEsc:
				m.bindingForm.Cancel("Binding edit canceled")
				return m, nil
			case keyEnter:
				done, err := m.bindingForm.Submit()
				if err != nil {
					return m, nil
				}
				if done {
					if m.bindingService != nil {
						if err := m.bindingService.Save(m.home); err == nil {
							m.table.SetRows(m.dashboardService.BuildTableRows())
						}
					}
				}
				return m, nil
			default:
				return m, m.bindingForm.Update(msg)
			}
		}

		if m.secretForm.Active() {
			switch key {
			case keyEsc:
				m.secretForm.Cancel("Canceled secret input")
				return m, nil
			case keyEnter:
				value, err := m.secretForm.Submit()
				if err != nil {
					return m, nil
				}
				if m.secretService != nil {
					m.secretService.SaveValue(m.home, value)
				}
				return m, nil
			default:
				return m, m.secretForm.Update(msg)
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

		case "c":
			// Jump to Config view (in DevOps section, index 4)
			m.currentSection = 4
			m.sectionViewIndex = 2 // Config is the 3rd item in DevOps section
			m.updateViewFromSection()
			return m, m.sectionEnterCmd()

		case "r":
			if m.currentView == viewSecrets && m.secretService != nil {
				m.secretService.Remove(m.home, m.filteredProviderIndexes(), m.selectedIndex)
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
			if m.currentView == viewSecrets && m.secretService != nil {
				if m.secretService.StartForm(m.filteredProviderIndexes(), m.selectedIndex) {
					m.searchActive = false
				}
				return m, nil
			}
			// Check proxy status
			m.proxyStatus = proxysvc.StatusChecking
			return m, checkProxyStatus

		case "e":
			if m.currentView == viewProviders {
				indexes := m.filteredProviderIndexes()
				if m.providerService != nil && m.providerService.StartForm(false, indexes, m.selectedIndex) {
					m.searchActive = false
				}
				return m, nil
			}
			if m.currentView == viewBindings {
				indexes := m.filteredBindingIndexes()
				if m.bindingService != nil && m.bindingService.StartForm(false, indexes, m.selectedIndex) {
					m.searchActive = false
				}
				return m, nil
			}
			return m, nil

		case "n":
			if m.currentView == viewBindings {
				indexes := m.filteredBindingIndexes()
				if m.bindingService != nil && m.bindingService.StartForm(true, indexes, m.selectedIndex) {
					m.searchActive = false
				}
				return m, nil
			}
			return m, nil

		case "a":
			if m.currentView == viewProviders {
				indexes := m.filteredProviderIndexes()
				if m.providerService != nil && m.providerService.StartForm(true, indexes, m.selectedIndex) {
					m.searchActive = false
				}
				return m, nil
			}
			return m, nil

		case "t":
			if m.currentView == viewSecrets && m.secretService != nil {
				m.secretService.Test(m.filteredProviderIndexes(), m.selectedIndex)
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

		case "left", "h":
			if m.currentView == viewConfig {
				// Cycle theme left
				m.themeIndex--
				if m.themeIndex < 0 {
					m.themeIndex = len(m.themes) - 1
				}
				m.updateTheme()
				return m, nil
			}

		case "right", "l":
			if m.currentView == viewConfig {
				// Cycle theme right
				m.themeIndex++
				if m.themeIndex >= len(m.themes) {
					m.themeIndex = 0
				}
				m.updateTheme()
				return m, nil
			}

		case "up", "k":
			// Navigate up in list views
			if m.currentView != viewDashboard {
				if m.selectedIndex > 0 {
					m.selectedIndex--
				}
				return m, nil
			}
			// For dashboard, fall through to table.Update

		case "down", "j":
			// Navigate down in list views
			if m.currentView != viewDashboard {
				maxIndex := m.maxSelectableIndex()
				if m.selectedIndex < maxIndex {
					m.selectedIndex++
				}
				return m, nil
			}
			// For dashboard, fall through to table.Update
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

// updateTheme updates the current theme and saves it to settings
func (m *DashboardModel) updateTheme() {
	themeName := m.themes[m.themeIndex]
	m.theme = theme.GetTheme(themeName)
	m.styles = theme.NewStyles(m.theme)

	// Save to settings
	ctx := context.Background()
	if s, err := settings.Load(ctx, m.home); err == nil {
		s.Theme = themeName
		if err := settings.Save(ctx, m.home, s); err != nil {
			m.message = fmt.Sprintf("Failed to save theme: %v", err)
		}
	}
}

// maxSelectableIndex returns the maximum selectable index for the current view
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
		if m.reportsService != nil {
			return m.reportsService.OptionCount() - 1
		}
		return -1
	case viewConfig:
		return len(m.configService.GetConfigFiles()) - 1
	default:
		return 0
	}
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
	m.table.SetRows(m.dashboardService.BuildTableRows())

	// Set success message
	proxyState := dashboardsvc.ProxyStateOff
	if binding.UseProxy {
		proxyState = dashboardsvc.ProxyStateOn
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

	proxyState := dashboardsvc.ProxyStateOff
	if binding.UseProxy {
		proxyState = dashboardsvc.ProxyStateOn
	}

	// Update dashboard table rows to keep views consistent
	m.table.SetRows(m.dashboardService.BuildTableRows())
	m.message = fmt.Sprintf("Proxy %s for %s", proxyState, toolName)

	return m, nil
}
