package components

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/royisme/bobamixer/internal/ui/theme"
)

// HelpBar renders the navigation help line.
type HelpBar struct {
	text   string
	styles theme.Styles
}

// NewHelpBar constructs a HelpBar component.
func NewHelpBar(text string, styles theme.Styles) HelpBar {
	return HelpBar{
		text:   text,
		styles: styles,
	}
}

// Update satisfies the Bubble Tea component contract (no mutations needed).
func (c HelpBar) Update(_ tea.Msg) (HelpBar, tea.Cmd) {
	return c, nil
}

// View renders the help text with the shared style.
func (c HelpBar) View() string {
	return c.styles.Help.Render(c.text)
}
