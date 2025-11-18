package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// HelpSection describes a high-level dashboard section.
type HelpSection struct {
	Name     string
	Shortcut string
	Views    []string
}

// HelpViewProps contains the data necessary to render the help view.
type HelpViewProps struct {
	Theme             ThemePalette
	Sections          []HelpSection
	NavigationHelp    string
	ShowcaseShortcuts []Shortcut
}

// Shortcut describes a keybinding and its behavior.
type Shortcut struct {
	Key  string
	Desc string
}

// RenderHelpView renders the help overlay content.
func RenderHelpView(props HelpViewProps) string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(props.Theme.Primary).Padding(0, 2)
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(props.Theme.Success).Padding(1, 2)
	normalStyle := lipgloss.NewStyle().Foreground(props.Theme.Text).Padding(0, 2)
	keyStyle := lipgloss.NewStyle().Foreground(props.Theme.Primary).Bold(true)
	helpStyle := lipgloss.NewStyle().Foreground(props.Theme.Muted).Padding(1, 2)

	var content strings.Builder

	// Header
	content.WriteString(titleStyle.Render("❓ BobaMixer Help & Shortcuts"))
	content.WriteString("\n\n")

	// Navigation
	content.WriteString(headerStyle.Render("Section Navigation"))
	content.WriteString("\n")
	for _, section := range props.Sections {
		content.WriteString(normalStyle.Render("  "))
		content.WriteString(keyStyle.Render(fmt.Sprintf("[%s]", section.Shortcut)))
		content.WriteString(normalStyle.Render(fmt.Sprintf("  %s → %s", section.Name, strings.Join(section.Views, ", "))))
		content.WriteString("\n")
	}
	content.WriteString(normalStyle.Render("  "))
	content.WriteString(keyStyle.Render("[?]"))
	content.WriteString(normalStyle.Render("  Toggle this help overlay"))
	content.WriteString("\n")

	content.WriteString("\n")
	content.WriteString(headerStyle.Render("Global Shortcuts"))
	content.WriteString("\n")

	shortcuts := defaultShortcuts()
	if len(props.ShowcaseShortcuts) > 0 {
		shortcuts = props.ShowcaseShortcuts
	}

	for _, sc := range shortcuts {
		content.WriteString(normalStyle.Render("  "))
		content.WriteString(keyStyle.Render(fmt.Sprintf("[%s]", sc.Key)))
		content.WriteString(normalStyle.Render(fmt.Sprintf("  %s", sc.Desc)))
		content.WriteString("\n")
	}

	content.WriteString("\n")
	content.WriteString(headerStyle.Render("Quick Tips"))
	content.WriteString("\n")
	content.WriteString(normalStyle.Render("  • Use number keys (1-5) to jump between sections"))
	content.WriteString("\n")
	content.WriteString(normalStyle.Render("  • All interactive features are in the TUI"))
	content.WriteString("\n")
	content.WriteString(normalStyle.Render("  • CLI commands available for automation"))
	content.WriteString("\n")
	content.WriteString(normalStyle.Render("  • Press ? anytime to toggle this help overlay"))
	content.WriteString("\n\n")

	content.WriteString(headerStyle.Render("Documentation"))
	content.WriteString("\n")
	content.WriteString(normalStyle.Render("  Full docs: https://royisme.github.io/BobaMixer/"))
	content.WriteString("\n")
	content.WriteString(normalStyle.Render("  GitHub: https://github.com/royisme/BobaMixer"))
	content.WriteString("\n\n")

	// Footer/Help
	helpText := "Press Esc to close this overlay  |  " + props.NavigationHelp
	content.WriteString(helpStyle.Render(helpText))

	return content.String()
}

func defaultShortcuts() []Shortcut {
	return []Shortcut{
		{"Tab / Shift+Tab", "Cycle sections"},
		{"[ / ]", "Cycle views within a section"},
		{"↑/↓ or k/j", "Navigate in lists"},
		{"/", "Search within supported lists"},
		{"Esc", "Clear search / close dialogs"},
		{"R", "Run selected tool (Dashboard view)"},
		{"X", "Toggle proxy (Dashboard view)"},
		{"S", "Refresh proxy status (Proxy view)"},
		{"Q or Ctrl+C", "Quit BobaMixer"},
	}
}
