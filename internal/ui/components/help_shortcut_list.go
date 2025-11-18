package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/royisme/bobamixer/internal/ui/theme"
)

// Shortcut describes a global keybinding.
type Shortcut struct {
	Key         string
	Description string
}

// ShortcutList renders a list of shortcuts.
type ShortcutList struct {
	shortcuts []Shortcut
	styles    theme.Styles
}

// NewShortcutList constructs a shortcut list component.
func NewShortcutList(shortcuts []Shortcut, styles theme.Styles) ShortcutList {
	return ShortcutList{
		shortcuts: shortcuts,
		styles:    styles,
	}
}

// Update keeps the component immutable because shortcut hints are static.
func (c ShortcutList) Update(_ tea.Msg) (ShortcutList, tea.Cmd) {
	return c, nil
}

// View renders the shortcut table-like content.
func (c ShortcutList) View() string {
	var b strings.Builder
	shortcutStyle := c.styles.Selected
	shortcutStyle = shortcutStyle.PaddingLeft(0).PaddingRight(1)
	textStyle := c.styles.Normal
	textStyle = textStyle.PaddingLeft(0)

	for _, sc := range c.shortcuts {
		if strings.TrimSpace(sc.Key) == "" {
			continue
		}
		b.WriteString(textStyle.Render("  "))
		b.WriteString(shortcutStyle.Render(fmt.Sprintf("[%s]", sc.Key)))
		b.WriteString(textStyle.Render("  " + sc.Description))
		b.WriteString("\n")
	}

	return strings.TrimRight(b.String(), "\n")
}
