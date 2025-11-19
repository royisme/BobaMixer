package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	// We can't easily access the theme styles here without passing them in.
	// For now, let's stick to simple colorization but maybe add a background if we could.
	// However, the plan was to use StatusBar style.
	// Since StatusMessage is a simple component, let's just enhance the styling locally or assume it's wrapped.
	// Actually, looking at the code, it just returns a string.
	// Let's make it look a bit better with a background if possible, but we only have adaptive color.
	// Let's just stick to the plan of using theme.Colorize but maybe add some padding/bold.
	return lipgloss.NewStyle().
		Foreground(c.color).
		Bold(true).
		Padding(0, 1).
		Render(c.message)
}
