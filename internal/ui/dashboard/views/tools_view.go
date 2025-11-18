package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ToolRow represents a compact tool entry.
type ToolRow struct {
	Name  string
	Exec  string
	Kind  string
	Bound bool
}

// ToolDetails contains metadata for the selected tool.
type ToolDetails struct {
	ID          string
	ConfigType  string
	ConfigPath  string
	Description string
}

// ToolsViewProps carries all data required to render the tools view.
type ToolsViewProps struct {
	Theme             ThemePalette
	SearchBar         string
	EmptyStateMessage string
	Tools             []ToolRow
	SelectedIndex     int
	Details           *ToolDetails
	NavigationHelp    string
	HelpCommands      string
	BoundIcon         string
	UnboundIcon       string
}

// RenderToolsView renders the CLI tools management UI.
func RenderToolsView(props ToolsViewProps) string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(props.Theme.Primary).Padding(0, 2)
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(props.Theme.Success).Padding(1, 2)
	selectedStyle := lipgloss.NewStyle().Foreground(props.Theme.Text).Background(props.Theme.Primary).Bold(true).Padding(0, 1)
	normalStyle := lipgloss.NewStyle().Foreground(props.Theme.Text).Padding(0, 1)
	mutedStyle := lipgloss.NewStyle().Foreground(props.Theme.Muted).Padding(0, 1)
	helpStyle := lipgloss.NewStyle().Foreground(props.Theme.Muted).Padding(1, 2)

	var content strings.Builder

	content.WriteString(titleStyle.Render("BobaMixer - CLI Tools Management"))
	content.WriteString("\n\n")

	content.WriteString(headerStyle.Render("🛠 Detected Tools"))
	content.WriteString("\n\n")

	if bar := strings.TrimSpace(props.SearchBar); bar != "" {
		content.WriteString(bar)
		content.WriteString("\n\n")
	}

	rows := props.Tools
	if len(rows) == 0 {
		if msg := strings.TrimSpace(props.EmptyStateMessage); msg != "" {
			content.WriteString(mutedStyle.Render("  " + msg))
			content.WriteString("\n")
		}
	} else {
		selectedIndex := props.SelectedIndex
		if selectedIndex >= len(rows) {
			selectedIndex = len(rows) - 1
		}

		for idx, row := range rows {
			icon := props.UnboundIcon
			if row.Bound {
				icon = props.BoundIcon
			}

			line := fmt.Sprintf("  %s %-15s %-30s %s",
				icon,
				row.Name,
				row.Exec,
				row.Kind,
			)

			if idx == selectedIndex {
				content.WriteString(selectedStyle.Render("▶ " + line))
			} else {
				content.WriteString(normalStyle.Render("  " + line))
			}
			content.WriteString("\n")
		}
	}

	content.WriteString("\n")
	if props.Details != nil {
		content.WriteString(headerStyle.Render("Details"))
		content.WriteString("\n")
		content.WriteString(normalStyle.Render(fmt.Sprintf("  ID: %s", props.Details.ID)))
		content.WriteString("\n")
		content.WriteString(normalStyle.Render(fmt.Sprintf("  Config Type: %s", props.Details.ConfigType)))
		content.WriteString("\n")
		content.WriteString(normalStyle.Render(fmt.Sprintf("  Config Path: %s", props.Details.ConfigPath)))
		content.WriteString("\n")
		if strings.TrimSpace(props.Details.Description) != "" {
			content.WriteString(normalStyle.Render(fmt.Sprintf("  Description: %s", props.Details.Description)))
			content.WriteString("\n")
		}
		content.WriteString("\n")
	}

	helpText := props.NavigationHelp
	if cmds := strings.TrimSpace(props.HelpCommands); cmds != "" {
		helpText += "  " + cmds
	}
	content.WriteString(helpStyle.Render(helpText))

	return content.String()
}
