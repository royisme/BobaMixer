package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/royisme/bobamixer/internal/ui/theme"
)

// HookInfo represents metadata about a git hook.
type HookInfo struct {
	Name   string
	Desc   string
	Active bool
}

// HookList renders hook entries with active/inactive indicators.
type HookList struct {
	hooks        []HookInfo
	activeIcon   string
	inactiveIcon string
	styles       theme.Styles
}

// NewHookList constructs the hook list component.
func NewHookList(hooks []HookInfo, activeIcon string, inactiveIcon string, styles theme.Styles) HookList {
	return HookList{
		hooks:        hooks,
		activeIcon:   activeIcon,
		inactiveIcon: inactiveIcon,
		styles:       styles,
	}
}

// Update satisfies the Bubble Tea component interface.
func (c HookList) Update(_ tea.Msg) (HookList, tea.Cmd) {
	return c, nil
}

// View renders the hook list.
func (c HookList) View() string {
	if len(c.hooks) == 0 {
		normalStyle := c.styles.Normal
		return normalStyle.PaddingLeft(2).Render("No hooks available.")
	}

	var b strings.Builder
	for _, hook := range c.hooks {
		statusStyle := c.styles.BudgetDanger
		icon := c.inactiveIcon
		if hook.Active {
			statusStyle = c.styles.BudgetOK
			icon = c.activeIcon
		}

		normalStyle := c.styles.Normal
		b.WriteString(normalStyle.PaddingLeft(2).Render(hook.Name))
		b.WriteString(statusStyle.Render("  " + icon))
		b.WriteString("\n")
		b.WriteString(normalStyle.PaddingLeft(6).Render("→ " + hook.Desc))
		b.WriteString("\n")
	}

	return strings.TrimRight(b.String(), "\n")
}
