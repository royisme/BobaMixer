package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/royisme/bobamixer/internal/ui/theme"
)

// BulletList renders bullet points using the muted text style.
type BulletList struct {
	items  []string
	styles theme.Styles
}

// NewBulletList constructs a bullet list component.
func NewBulletList(items []string, styles theme.Styles) BulletList {
	return BulletList{
		items:  items,
		styles: styles,
	}
}

// Update satisfies the Bubble Tea component interface.
func (c BulletList) Update(_ tea.Msg) (BulletList, tea.Cmd) {
	return c, nil
}

// View renders each bullet with muted styling.
func (c BulletList) View() string {
	if len(c.items) == 0 {
		return ""
	}

	var b strings.Builder
	style := c.styles.Normal
	style = style.PaddingLeft(2)
	for _, item := range c.items {
		text := strings.TrimSpace(item)
		if text == "" {
			continue
		}
		b.WriteString(style.Render("• " + text))
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n")
}
