package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ConfigFile describes a single editable configuration file.
type ConfigFile struct {
	Name string
	File string
	Desc string
}

// ConfigViewProps carries the data required to render the configuration view.
type ConfigViewProps struct {
	Theme              ThemePalette
	SelectedIndex      int
	ConfigFiles        []ConfigFile
	Home               string
	HelpTextNavigation string
}

// RenderConfigView renders the configuration editor view.
func RenderConfigView(props ConfigViewProps) string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(props.Theme.Primary).Padding(0, 2)
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(props.Theme.Success).Padding(1, 2)
	normalStyle := lipgloss.NewStyle().Foreground(props.Theme.Text).Padding(0, 2)
	selectedStyle := lipgloss.NewStyle().Foreground(props.Theme.Text).Background(props.Theme.Primary).Bold(true).Padding(0, 1)
	mutedStyle := lipgloss.NewStyle().Foreground(props.Theme.Muted).Padding(0, 2)
	helpStyle := lipgloss.NewStyle().Foreground(props.Theme.Muted).Padding(1, 2)

	var content strings.Builder

	// Header
	content.WriteString(titleStyle.Render("⚙️  Configuration Editor"))
	content.WriteString("\n\n")

	content.WriteString(headerStyle.Render("Configuration Files"))
	content.WriteString("\n")

	selectedIndex := props.SelectedIndex
	if selectedIndex >= len(props.ConfigFiles) {
		selectedIndex = 0
	}

	for i, cfg := range props.ConfigFiles {
		line := fmt.Sprintf("  %s", cfg.Name)
		filePath := lipgloss.NewStyle().Foreground(props.Theme.Muted).Render(fmt.Sprintf(" (%s)", cfg.File))

		if i == selectedIndex {
			content.WriteString(selectedStyle.Render("▶ " + line))
			content.WriteString(filePath)
		} else {
			content.WriteString(normalStyle.Render("  " + line))
			content.WriteString(filePath)
		}
		content.WriteString("\n")

		// Show description for selected item
		if i == selectedIndex {
			content.WriteString(mutedStyle.Render(fmt.Sprintf("    %s", cfg.Desc)))
			content.WriteString("\n")
			content.WriteString(mutedStyle.Render(fmt.Sprintf("    Full path: %s/%s", props.Home, cfg.File)))
			content.WriteString("\n")
		}
	}

	content.WriteString("\n")
	content.WriteString(headerStyle.Render("Editor Settings"))
	content.WriteString("\n")

	editor := "vim" // Default, in real implementation check $EDITOR
	content.WriteString(normalStyle.Render(fmt.Sprintf("  Editor: $EDITOR (%s)", editor)))
	content.WriteString("\n")
	content.WriteString(mutedStyle.Render("  Tip: Set $EDITOR environment variable to use your preferred editor"))
	content.WriteString("\n\n")

	content.WriteString(headerStyle.Render("Safety Features"))
	content.WriteString("\n")
	content.WriteString(normalStyle.Render("  • Automatic backup before editing"))
	content.WriteString("\n")
	content.WriteString(normalStyle.Render("  • YAML syntax validation after save"))
	content.WriteString("\n")
	content.WriteString(normalStyle.Render("  • Rollback support if validation fails"))
	content.WriteString("\n\n")

	// Footer/Help
	helpText := props.HelpTextNavigation + "\n  Use CLI: boba edit <target> (to open in editor)"
	content.WriteString(helpStyle.Render(helpText))

	return content.String()
}
