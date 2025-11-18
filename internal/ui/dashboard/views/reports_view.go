package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ReportOption describes a report configuration entry.
type ReportOption struct {
	Label string
	Desc  string
}

// ReportsViewProps carries the data required to render the reports view.
type ReportsViewProps struct {
	Theme           ThemePalette
	Options         []ReportOption
	SelectedIndex   int
	Home            string
	NavigationHelp  string
	CommandHelpLine string
}

// RenderReportsView renders the usage reports view.
func RenderReportsView(props ReportsViewProps) string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(props.Theme.Primary).Padding(0, 2)
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(props.Theme.Success).Padding(1, 2)
	normalStyle := lipgloss.NewStyle().Foreground(props.Theme.Text).Padding(0, 2)
	selectedStyle := lipgloss.NewStyle().Foreground(props.Theme.Text).Background(props.Theme.Primary).Bold(true).Padding(0, 1)
	helpStyle := lipgloss.NewStyle().Foreground(props.Theme.Muted).Padding(1, 2)

	var content strings.Builder

	// Header
	content.WriteString(titleStyle.Render("📊 Generate Usage Report"))
	content.WriteString("\n\n")

	selectedIndex := props.SelectedIndex
	if len(props.Options) > 0 && selectedIndex >= len(props.Options) {
		selectedIndex = 0
	}

	content.WriteString(headerStyle.Render("Report Options"))
	content.WriteString("\n")

	for i, opt := range props.Options {
		line := fmt.Sprintf("  %s", opt.Label)
		if i == selectedIndex {
			content.WriteString(selectedStyle.Render("▶ " + line))
		} else {
			content.WriteString(normalStyle.Render("  " + line))
		}
		content.WriteString("\n")

		if i == selectedIndex {
			content.WriteString(lipgloss.NewStyle().Foreground(props.Theme.Muted).Padding(0, 4).Render("  → " + opt.Desc))
			content.WriteString("\n")
		}
	}

	content.WriteString("\n")
	content.WriteString(headerStyle.Render("Output Configuration"))
	content.WriteString("\n")
	content.WriteString(normalStyle.Render(fmt.Sprintf("  Default path: %s/reports/", props.Home)))
	content.WriteString("\n")
	content.WriteString(normalStyle.Render("  Filename: bobamixer-<date>.<format>"))
	content.WriteString("\n\n")

	content.WriteString(headerStyle.Render("Report Contents"))
	content.WriteString("\n")
	content.WriteString(normalStyle.Render("  ✓ Summary statistics (tokens, costs, sessions)"))
	content.WriteString("\n")
	content.WriteString(normalStyle.Render("  ✓ Daily trends and usage patterns"))
	content.WriteString("\n")
	content.WriteString(normalStyle.Render("  ✓ Profile breakdown and comparison"))
	content.WriteString("\n")
	content.WriteString(normalStyle.Render("  ✓ Cost analysis and optimization opportunities"))
	content.WriteString("\n")
	content.WriteString(normalStyle.Render("  ✓ Peak usage times and anomalies"))
	content.WriteString("\n\n")

	helpText := props.NavigationHelp
	if props.CommandHelpLine != "" {
		helpText += "\n" + props.CommandHelpLine
	}
	content.WriteString(helpStyle.Render(helpText))

	return content.String()
}
