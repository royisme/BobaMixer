package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/royisme/bobamixer/internal/ui/theme"
)

// InfoMessage renders muted informational text (e.g., form hints).
type InfoMessage struct {
	text   string
	styles theme.Styles
}

// NewInfoMessage constructs a muted info message component.
func NewInfoMessage(text string, styles theme.Styles) InfoMessage {
	return InfoMessage{
		text:   strings.TrimSpace(text),
		styles: styles,
	}
}

// Update satisfies the Bubble Tea component interface.
func (c InfoMessage) Update(_ tea.Msg) (InfoMessage, tea.Cmd) {
	return c, nil
}

// View renders the info text if provided.
func (c InfoMessage) View() string {
	if c.text == "" {
		return ""
	}
	normalStyle := c.styles.Normal
	return normalStyle.PaddingLeft(2).Render(c.text)
}
