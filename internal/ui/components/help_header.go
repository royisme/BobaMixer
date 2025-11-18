package components

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/royisme/bobamixer/internal/ui/theme"
)

// HelpHeader renders the main title for the help overlay.
type HelpHeader struct {
	title   string
	subtext string
	styles  theme.Styles
}

// NewHelpHeader builds a header component with the provided title and optional subtext.
func NewHelpHeader(title string, subtext string, styles theme.Styles) HelpHeader {
	return HelpHeader{
		title:   title,
		subtext: subtext,
		styles:  styles,
	}
}

// Update satisfies the component contract but the header has no runtime state to change.
func (c HelpHeader) Update(_ tea.Msg) (HelpHeader, tea.Cmd) {
	return c, nil
}

// View renders the header and subtext using the shared theme styles.
func (c HelpHeader) View() string {
	content := c.styles.Title.Render(c.title)
	if c.subtext != "" {
		content += "\n" + c.styles.Help.Render(c.subtext)
	}
	return content
}
