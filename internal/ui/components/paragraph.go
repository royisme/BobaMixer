package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/royisme/bobamixer/internal/ui/theme"
)

// Paragraph renders a block of text with the normal style.
type Paragraph struct {
	text   string
	styles theme.Styles
}

// NewParagraph constructs a paragraph component.
func NewParagraph(text string, styles theme.Styles) Paragraph {
	return Paragraph{
		text:   strings.TrimSpace(text),
		styles: styles,
	}
}

// Update satisfies the Bubble Tea component interface.
func (c Paragraph) Update(_ tea.Msg) (Paragraph, tea.Cmd) {
	return c, nil
}

// View renders the paragraph content if non-empty.
func (c Paragraph) View() string {
	if c.text == "" {
		return ""
	}
	lines := strings.Split(c.text, "\n")
	var b strings.Builder
	style := c.styles.Normal
	style = style.PaddingLeft(2)
	for _, line := range lines {
		b.WriteString(style.Render(line))
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n")
}
