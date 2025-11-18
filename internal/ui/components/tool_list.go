package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/royisme/bobamixer/internal/ui/theme"
)

// ToolRow represents a CLI tool entry in the list.
type ToolRow struct {
	Name  string
	Exec  string
	Kind  string
	Bound bool
}

// ToolList renders tool rows along with their bound status.
type ToolList struct {
	rows        []ToolRow
	selected    int
	boundIcon   string
	unboundIcon string
	emptyState  string
	styles      theme.Styles
}

// NewToolList constructs the tool list component.
func NewToolList(rows []ToolRow, selected int, emptyState string, boundIcon string, unboundIcon string, styles theme.Styles) ToolList {
	return ToolList{
		rows:        rows,
		selected:    selected,
		boundIcon:   boundIcon,
		unboundIcon: unboundIcon,
		emptyState:  strings.TrimSpace(emptyState),
		styles:      styles,
	}
}

// Update satisfies the Bubble Tea component interface.
func (c ToolList) Update(_ tea.Msg) (ToolList, tea.Cmd) {
	return c, nil
}

// View renders the tools list or the empty state message.
func (c ToolList) View() string {
	if len(c.rows) == 0 {
		if c.emptyState == "" {
			return ""
		}
		normalStyle := c.styles.Normal
		return normalStyle.PaddingLeft(2).Render(c.emptyState)
	}

	var b strings.Builder
	selected := c.selected
	if selected >= len(c.rows) {
		selected = len(c.rows) - 1
	}

	for idx, row := range c.rows {
		icon := c.unboundIcon
		if row.Bound {
			icon = c.boundIcon
		}

		line := fmt.Sprintf("  %s %-15s %-30s %s",
			icon,
			row.Name,
			row.Exec,
			row.Kind,
		)

		if idx == selected {
			b.WriteString(c.styles.Selected.Render("▶ " + line))
		} else {
			b.WriteString(c.styles.Normal.Render("  " + line))
		}
		b.WriteString("\n")
	}

	return strings.TrimRight(b.String(), "\n")
}
