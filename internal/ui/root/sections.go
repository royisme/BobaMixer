// Package root provides the root UI model and orchestration for the BobaMixer TUI.
package root

import tea "github.com/charmbracelet/bubbletea"

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

// viewSection represents a logical grouping of related views
type viewSection struct {
	name     string
	shortcut string
	views    []viewMode
}

// initSections initializes the section navigation structure
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

// updateViewFromSection updates the current view based on section navigation state
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

// moveToSection jumps directly to a specific section by index
func (m *DashboardModel) moveToSection(idx int) tea.Cmd {
	if idx < 0 || idx >= len(m.sections) {
		return nil
	}
	m.currentSection = idx
	m.sectionViewIndex = 0
	m.updateViewFromSection()
	return m.sectionEnterCmd()
}

// cycleSection moves forward or backward through sections
func (m *DashboardModel) cycleSection(delta int) tea.Cmd {
	m.currentSection = (m.currentSection + delta + len(m.sections)) % len(m.sections)
	m.sectionViewIndex = 0
	m.updateViewFromSection()
	return m.sectionEnterCmd()
}

// cycleSubview moves through views within the current section
func (m *DashboardModel) cycleSubview(delta int) tea.Cmd {
	section := m.sections[m.currentSection]
	if len(section.views) == 0 {
		return nil
	}
	m.sectionViewIndex = (m.sectionViewIndex + delta + len(section.views)) % len(section.views)
	m.updateViewFromSection()
	return m.sectionEnterCmd()
}

// sectionEnterCmd returns any command needed when entering a view
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

// viewName returns the display name for a view mode
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
