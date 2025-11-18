package views

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// RoutingViewProps contains static text fragments for the routing view.
type RoutingViewProps struct {
	Theme           ThemePalette
	NavigationHelp  string
	CommandHelpLine string
}

// RenderRoutingView renders the routing tester information view.
func RenderRoutingView(props RoutingViewProps) string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(props.Theme.Primary).Padding(0, 2)
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(props.Theme.Success).Padding(1, 2)
	normalStyle := lipgloss.NewStyle().Foreground(props.Theme.Text).Padding(0, 1)
	mutedStyle := lipgloss.NewStyle().Foreground(props.Theme.Muted).Padding(0, 1)
	helpStyle := lipgloss.NewStyle().Foreground(props.Theme.Muted).Padding(1, 2)

	var content strings.Builder

	content.WriteString(titleStyle.Render("BobaMixer - Routing Rules Tester"))
	content.WriteString("\n\n")

	content.WriteString(headerStyle.Render("🧪 Test Routing Rules"))
	content.WriteString("\n")
	content.WriteString(mutedStyle.Render("  Test how routing rules would apply to different queries."))
	content.WriteString("\n\n")

	content.WriteString(headerStyle.Render("💡 How to Use"))
	content.WriteString("\n")
	content.WriteString(normalStyle.Render("  1. Prepare a test query (text or file)"))
	content.WriteString("\n")
	content.WriteString(normalStyle.Render("  2. Run: boba route test \"your query text\""))
	content.WriteString("\n")
	content.WriteString(normalStyle.Render("  3. Or: boba route test @path/to/file.txt"))
	content.WriteString("\n\n")

	content.WriteString(headerStyle.Render("📋 Example"))
	content.WriteString("\n")
	content.WriteString(normalStyle.Render("  $ boba route test \"Write a Python function\""))
	content.WriteString("\n")
	content.WriteString(mutedStyle.Render("  → Profile: claude-sonnet-3.5"))
	content.WriteString("\n")
	content.WriteString(mutedStyle.Render("  → Rule: short-query-fast-model"))
	content.WriteString("\n")
	content.WriteString(mutedStyle.Render("  → Reason: Query < 100 chars"))
	content.WriteString("\n\n")

	content.WriteString(headerStyle.Render("ℹ️  Context Detection"))
	content.WriteString("\n")
	content.WriteString(normalStyle.Render("  Routing considers:"))
	content.WriteString("\n")
	content.WriteString(normalStyle.Render("  • Query length and complexity"))
	content.WriteString("\n")
	content.WriteString(normalStyle.Render("  • Current project and branch"))
	content.WriteString("\n")
	content.WriteString(normalStyle.Render("  • Time of day (day/evening/night)"))
	content.WriteString("\n")
	content.WriteString(normalStyle.Render("  • Project type (go, web, etc.)"))
	content.WriteString("\n\n")

	helpText := props.NavigationHelp
	if props.CommandHelpLine != "" {
		helpText += "\n" + props.CommandHelpLine
	}
	content.WriteString(helpStyle.Render(helpText))

	return content.String()
}
