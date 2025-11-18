package pages

import tea "github.com/charmbracelet/bubbletea"

// Page represents a composable UI unit following the Bubble Tea model contract.
type Page interface {
	Init() tea.Cmd
	Update(msg tea.Msg) (Page, tea.Cmd)
	View() string
}
