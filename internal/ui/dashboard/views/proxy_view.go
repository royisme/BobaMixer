package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ProxyViewProps describes the proxy status view.
type ProxyViewProps struct {
	Theme           ThemePalette
	StatusState     string
	StatusText      string
	StatusIcon      string
	Address         string
	ShowConfig      bool
	NavigationHelp  string
	AdditionalNote  string
	CommandHelpLine string
}

// RenderProxyView renders the proxy server control view.
func RenderProxyView(props ProxyViewProps) string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(props.Theme.Primary).Padding(0, 2)
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(props.Theme.Success).Padding(1, 2)
	normalStyle := lipgloss.NewStyle().Foreground(props.Theme.Text).Padding(0, 1)
	successStyle := lipgloss.NewStyle().Foreground(props.Theme.Success).Padding(0, 1)
	dangerStyle := lipgloss.NewStyle().Foreground(props.Theme.Danger).Padding(0, 1)
	helpStyle := lipgloss.NewStyle().Foreground(props.Theme.Muted).Padding(1, 2)

	var content strings.Builder

	content.WriteString(titleStyle.Render("BobaMixer - Proxy Server Control"))
	content.WriteString("\n\n")

	content.WriteString(headerStyle.Render("🌐 Proxy Status"))
	content.WriteString("\n\n")

	var statusStyle lipgloss.Style
	switch props.StatusState {
	case "running":
		statusStyle = successStyle
	case "stopped":
		statusStyle = dangerStyle
	default:
		statusStyle = normalStyle
	}

	statusLine := fmt.Sprintf("  Status:   %s", statusStyle.Render(props.StatusIcon+" "+props.StatusText))
	content.WriteString(normalStyle.Render(statusLine))
	content.WriteString("\n")
	content.WriteString(normalStyle.Render(fmt.Sprintf("  Address:  %s", props.Address)))
	content.WriteString("\n\n")

	content.WriteString(headerStyle.Render("ℹ️  Information"))
	content.WriteString("\n")
	content.WriteString(normalStyle.Render("  The proxy server intercepts AI API requests from CLI tools"))
	content.WriteString("\n")
	content.WriteString(normalStyle.Render("  and routes them through BobaMixer for tracking and control."))
	content.WriteString("\n\n")

	if props.ShowConfig {
		content.WriteString(headerStyle.Render("📝 Configuration"))
		content.WriteString("\n")
		content.WriteString(normalStyle.Render("  Tools with proxy enabled will automatically use:"))
		content.WriteString("\n")
		content.WriteString(normalStyle.Render(fmt.Sprintf("  • HTTP_PROXY=%s", props.Address)))
		content.WriteString("\n")
		content.WriteString(normalStyle.Render(fmt.Sprintf("  • HTTPS_PROXY=%s", props.Address)))
		content.WriteString("\n\n")
	}

	helpLines := []string{props.NavigationHelp}
	if props.CommandHelpLine != "" {
		helpLines = append(helpLines, props.CommandHelpLine)
	}
	if props.AdditionalNote != "" {
		helpLines = append(helpLines, props.AdditionalNote)
	}

	content.WriteString(helpStyle.Render(strings.Join(helpLines, "\n")))

	return content.String()
}
