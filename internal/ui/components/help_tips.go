package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/royisme/bobamixer/internal/ui/theme"
)

// HelpTips renders a bullet list of quick tips for the user.
type HelpTips struct {
	tips   []string
	styles theme.Styles
}

// NewHelpTips constructs the tips component.
func NewHelpTips(tips []string, styles theme.Styles) HelpTips {
	return HelpTips{
		tips:   tips,
		styles: styles,
	}
}

// Update keeps the component immutable because tips are static.
func (c HelpTips) Update(_ tea.Msg) (HelpTips, tea.Cmd) {
	return c, nil
}

// View renders each tip with the shared normal style.
func (c HelpTips) View() string {
	if len(c.tips) == 0 {
		return ""
	}

	var b strings.Builder
	textStyle := c.styles.Normal
	textStyle = textStyle.PaddingLeft(0)
	for _, tip := range c.tips {
		b.WriteString(textStyle.Render("  • " + strings.TrimSpace(tip)))
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n")
}
