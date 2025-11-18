package components

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/royisme/bobamixer/internal/ui/theme"
)

// HelpFooter renders the closing hint for the help overlay.
type HelpFooter struct {
	message string
	styles  theme.Styles
}

// NewHelpFooter constructs the footer component.
func NewHelpFooter(message string, styles theme.Styles) HelpFooter {
	return HelpFooter{
		message: message,
		styles:  styles,
	}
}

// Update keeps the component immutable because the footer has no state.
func (c HelpFooter) Update(_ tea.Msg) (HelpFooter, tea.Cmd) {
	return c, nil
}

// View renders the footer hint message.
func (c HelpFooter) View() string {
	return c.styles.Help.Render(c.message)
}
