package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/royisme/bobamixer/internal/ui/theme"
)

// BindingRow represents a binding entry in the list.
type BindingRow struct {
	ToolName     string
	ProviderName string
	UseProxy     bool
}

// BindingList renders tool-provider bindings with proxy status.
type BindingList struct {
	rows          []BindingRow
	selected      int
	proxyEnabled  string
	proxyDisabled string
	emptyState    string
	styles        theme.Styles
}

// NewBindingList constructs the bindings list component.
func NewBindingList(rows []BindingRow, selected int, emptyState string, proxyEnabled string, proxyDisabled string, styles theme.Styles) BindingList {
	return BindingList{
		rows:          rows,
		selected:      selected,
		proxyEnabled:  proxyEnabled,
		proxyDisabled: proxyDisabled,
		emptyState:    strings.TrimSpace(emptyState),
		styles:        styles,
	}
}

// Update satisfies the Bubble Tea component interface.
func (c BindingList) Update(_ tea.Msg) (BindingList, tea.Cmd) {
	return c, nil
}

// View renders the bindings list or the empty state message.
func (c BindingList) View() string {
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
	if selected < 0 {
		selected = 0
	}

	for idx, row := range c.rows {
		icon := c.proxyDisabled
		if row.UseProxy {
			icon = c.proxyEnabled
		}

		line := fmt.Sprintf("  %-15s → %-25s  Proxy: %s", row.ToolName, row.ProviderName, icon)
		if idx == selected {
			b.WriteString(c.styles.Selected.Render("▶ " + line))
		} else {
			b.WriteString(c.styles.Normal.Render("  " + line))
		}
		b.WriteString("\n")
	}

	return strings.TrimRight(b.String(), "\n")
}
