package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/royisme/bobamixer/internal/ui/theme"
)

// ReportOption represents a report configuration entry.
type ReportOption struct {
	Label string
	Desc  string
}

// ReportOptionsList renders selectable report options.
type ReportOptionsList struct {
	options  []ReportOption
	selected int
	styles   theme.Styles
}

// NewReportOptionsList constructs the component.
func NewReportOptionsList(options []ReportOption, selected int, styles theme.Styles) ReportOptionsList {
	return ReportOptionsList{
		options:  options,
		selected: selected,
		styles:   styles,
	}
}

// Update satisfies the Bubble Tea component interface.
func (c ReportOptionsList) Update(_ tea.Msg) (ReportOptionsList, tea.Cmd) {
	return c, nil
}

// View renders the selectable options with descriptions.
func (c ReportOptionsList) View() string {
	if len(c.options) == 0 {
		normalStyle := c.styles.Normal
		return normalStyle.PaddingLeft(2).Render("No report templates configured.")
	}

	var b strings.Builder
	selected := c.selected
	if selected >= len(c.options) {
		selected = 0
	}

	for idx, option := range c.options {
		line := fmt.Sprintf("  %s", option.Label)
		if idx == selected {
			b.WriteString(c.styles.Selected.Render("▶ " + line))
		} else {
			b.WriteString(c.styles.Normal.Render("  " + line))
		}
		b.WriteString("\n")

		if idx == selected && strings.TrimSpace(option.Desc) != "" {
			normalStyle := c.styles.Normal
			desc := normalStyle.PaddingLeft(6).Render("→ " + option.Desc)
			b.WriteString(desc)
			b.WriteString("\n")
		}
	}

	return strings.TrimRight(b.String(), "\n")
}
