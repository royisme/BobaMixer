package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/royisme/bobamixer/internal/ui/theme"
)

// StatusMessage renders a highlighted status line.
type StatusMessage struct {
	message string
	color   lipgloss.AdaptiveColor
}

// NewStatusMessage constructs a StatusMessage component.
func NewStatusMessage(message string, color lipgloss.AdaptiveColor) StatusMessage {
	return StatusMessage{
		message: strings.TrimSpace(message),
		color:   color,
	}
}

// Update satisfies the Bubble Tea component contract (no mutations needed).
func (c StatusMessage) Update(_ tea.Msg) (StatusMessage, tea.Cmd) {
	return c, nil
}

// View renders the status message in the configured color.
func (c StatusMessage) View() string {
	if c.message == "" {
		return ""
	}
	return theme.Colorize(c.color, c.message)
}
