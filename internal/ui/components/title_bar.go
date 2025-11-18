package components

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/royisme/bobamixer/internal/ui/theme"
)

// TitleBar renders the primary page title.
type TitleBar struct {
	text   string
	styles theme.Styles
}

// NewTitleBar constructs a TitleBar component.
func NewTitleBar(text string, styles theme.Styles) TitleBar {
	return TitleBar{
		text:   text,
		styles: styles,
	}
}

// Update satisfies the Bubble Tea component contract (no state changes needed).
func (c TitleBar) Update(_ tea.Msg) (TitleBar, tea.Cmd) {
	return c, nil
}

// View renders the title using the shared style definitions.
func (c TitleBar) View() string {
	return c.styles.Title.Render(c.text)
}
