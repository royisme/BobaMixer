package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// HookInfo represents a supported git hook entry.
type HookInfo struct {
	Name   string
	Desc   string
	Active bool
}

// HooksViewProps contains the data necessary to render the git hooks view.
type HooksViewProps struct {
	Theme           ThemePalette
	RepoPath        string
	HooksInstalled  bool
	Hooks           []HookInfo
	NavigationHelp  string
	CommandHelpLine string
	ActiveIcon      string
	InactiveIcon    string
}

// RenderHooksView renders the git hooks management view.
func RenderHooksView(props HooksViewProps) string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(props.Theme.Primary).Padding(0, 2)
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(props.Theme.Success).Padding(1, 2)
	normalStyle := lipgloss.NewStyle().Foreground(props.Theme.Text).Padding(0, 2)
	successStyle := lipgloss.NewStyle().Foreground(props.Theme.Success).Padding(0, 2)
	dangerStyle := lipgloss.NewStyle().Foreground(props.Theme.Danger).Padding(0, 2)
	helpStyle := lipgloss.NewStyle().Foreground(props.Theme.Muted).Padding(1, 2)

	var content strings.Builder

	// Header
	content.WriteString(titleStyle.Render("🪝 Git Hooks Management"))
	content.WriteString("\n\n")

	// Repository detection
	content.WriteString(headerStyle.Render("Current Repository"))
	content.WriteString("\n")
	content.WriteString(normalStyle.Render(fmt.Sprintf("  Path: %s", props.RepoPath)))
	content.WriteString("\n")

	if props.HooksInstalled {
		content.WriteString(successStyle.Render("  Status: ✓ Hooks Installed"))
	} else {
		content.WriteString(dangerStyle.Render("  Status: ✗ Hooks Not Installed"))
	}
	content.WriteString("\n\n")

	content.WriteString(headerStyle.Render("Available Hooks"))
	content.WriteString("\n")

	for _, hook := range props.Hooks {
		statusStyle := dangerStyle
		statusIcon := props.InactiveIcon
		if hook.Active {
			statusStyle = successStyle
			statusIcon = props.ActiveIcon
		}

		content.WriteString(normalStyle.Render(fmt.Sprintf("  %s", hook.Name)))
		content.WriteString(statusStyle.Render(fmt.Sprintf("  %s", statusIcon)))
		content.WriteString("\n")
		content.WriteString(lipgloss.NewStyle().Foreground(props.Theme.Muted).Padding(0, 4).Render(fmt.Sprintf("  → %s", hook.Desc)))
		content.WriteString("\n")
	}

	content.WriteString("\n")
	content.WriteString(headerStyle.Render("Benefits"))
	content.WriteString("\n")
	content.WriteString(normalStyle.Render("  • Automatic profile suggestions based on branch/project"))
	content.WriteString("\n")
	content.WriteString(normalStyle.Render("  • Track repository events for better usage analytics"))
	content.WriteString("\n")
	content.WriteString(normalStyle.Render("  • Context-aware AI model selection"))
	content.WriteString("\n")
	content.WriteString(normalStyle.Render("  • Zero-overhead tracking (async logging)"))
	content.WriteString("\n\n")

	content.WriteString(headerStyle.Render("Recent Hook Activity"))
	content.WriteString("\n")
	content.WriteString(lipgloss.NewStyle().Foreground(props.Theme.Muted).Padding(0, 2).Render("  No recent activity recorded"))
	content.WriteString("\n\n")

	helpText := props.NavigationHelp
	if props.CommandHelpLine != "" {
		helpText += "\n" + props.CommandHelpLine
	}
	content.WriteString(helpStyle.Render(helpText))

	return content.String()
}
