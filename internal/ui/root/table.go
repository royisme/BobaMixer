// Package root provides the root UI model and orchestration for the BobaMixer TUI.
package root

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

// initializeTable creates and configures the dashboard table
func (m *DashboardModel) initializeTable() {
	columns := []table.Column{
		{Title: "Tool", Width: 15},
		{Title: "Provider", Width: 20},
		{Title: "Model", Width: 25},
		{Title: "Proxy", Width: 10},
		{Title: "Status", Width: 20},
	}

	rows := m.dashboardService.BuildTableRows()

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(false),
		table.WithHeight(7),
	)

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

// updateTableSize adjusts table dimensions based on terminal size
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
