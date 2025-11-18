// Package root provides the root UI model and orchestration for the BobaMixer TUI.
package root

import (
	"fmt"
	"strings"
)

// supportsSearch returns whether a view has search functionality
func (m *DashboardModel) supportsSearch(view viewMode) bool {
	switch view {
	case viewProviders, viewTools, viewBindings, viewSecrets:
		return true
	default:
		return false
	}
}

// activateSearch enters search mode for the current view
func (m *DashboardModel) activateSearch() {
	m.searchActive = true
	m.searchInput.SetValue(m.searchQuery)
	m.searchInput.CursorEnd()
	m.searchContextView = m.currentView
}

// clearSearch exits search mode and clears the search query
func (m *DashboardModel) clearSearch() {
	m.searchActive = false
	m.searchQuery = ""
}

// viewHasSearch returns whether the view has an active search filter
func (m *DashboardModel) viewHasSearch(view viewMode) bool {
	return strings.TrimSpace(m.searchQuery) != "" && m.searchContextView == view
}

// filteredProviderIndexes returns indexes of providers matching the search query
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

// filteredToolIndexes returns indexes of tools matching the search query
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

// filteredBindingIndexes returns indexes of bindings matching the search query
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

// renderSearchBar renders the search bar UI for views that support search
func (m DashboardModel) renderSearchBar(view viewMode) string {
	if !m.supportsSearch(view) {
		return ""
	}
	style := m.styles.Normal
	style = style.Padding(0, 2)
	switch {
	case m.searchActive && m.searchContextView == view:
		return style.Render("Search: " + m.searchInput.View())
	case m.viewHasSearch(view):
		return style.Render(fmt.Sprintf("Filter: %s  (Esc to clear)", m.searchQuery))
	default:
		return style.Render("Press '/' to search")
	}
}
