package views

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// DashboardViewProps contains the information rendered in the dashboard view.
type DashboardViewProps struct {
	Theme          ThemePalette
	TableView      string
	Message        string
	ProxyIcon      string
	ProxyStatus    string
	NavigationHelp string
	HelpCommands   string
}

// RenderDashboardView renders the high-level dashboard summary.
func RenderDashboardView(props DashboardViewProps) string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(props.Theme.Primary).Padding(0, 2)
	proxyStyle := lipgloss.NewStyle().Foreground(props.Theme.Text).Padding(0, 2)
	messageStyle := lipgloss.NewStyle().Foreground(props.Theme.Success).Padding(0, 2)
	helpStyle := lipgloss.NewStyle().Foreground(props.Theme.Muted).Padding(1, 2)

	var content strings.Builder

	content.WriteString(titleStyle.Render("BobaMixer - AI CLI Control Plane"))
	content.WriteString("\n")

	statusLine := proxyStyle.Render("  Proxy: " + props.ProxyIcon + " " + props.ProxyStatus)
	content.WriteString(statusLine)
	content.WriteString("\n\n")

	content.WriteString(props.TableView)
	content.WriteString("\n")

	if msg := strings.TrimSpace(props.Message); msg != "" {
		content.WriteString(messageStyle.Render("  " + msg))
		content.WriteString("\n")
	}

	helpText := props.NavigationHelp
	if strings.TrimSpace(props.HelpCommands) != "" {
		helpText += "  " + props.HelpCommands
	}
	content.WriteString(helpStyle.Render(helpText))

	return content.String()
}
