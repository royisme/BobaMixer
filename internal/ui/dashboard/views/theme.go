package views

import "github.com/charmbracelet/lipgloss"

// ThemePalette contains the minimal colors required by dashboard views.
type ThemePalette struct {
	Primary lipgloss.AdaptiveColor
	Success lipgloss.AdaptiveColor
	Danger  lipgloss.AdaptiveColor
	Warning lipgloss.AdaptiveColor
	Text    lipgloss.AdaptiveColor
	Muted   lipgloss.AdaptiveColor
}
