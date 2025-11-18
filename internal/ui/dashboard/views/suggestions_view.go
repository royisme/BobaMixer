package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Suggestion represents a view-friendly suggestion entry.
type Suggestion struct {
	Title       string
	Description string
	Impact      string
	ActionItems []string
	Priority    int
	Type        string
}

// SuggestionsViewProps carries data necessary to render suggestions.
type SuggestionsViewProps struct {
	Theme           ThemePalette
	Suggestions     []Suggestion
	SelectedIndex   int
	Error           string
	NavigationHelp  string
	CommandHelpLine string
}

// RenderSuggestionsView renders the optimization suggestions view.
func RenderSuggestionsView(props SuggestionsViewProps) string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(props.Theme.Primary).Padding(0, 2)
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(props.Theme.Success).Padding(1, 2)
	selectedStyle := lipgloss.NewStyle().Foreground(props.Theme.Text).Background(props.Theme.Primary).Bold(true).Padding(0, 1)
	normalStyle := lipgloss.NewStyle().Foreground(props.Theme.Text).Padding(0, 1)
	mutedStyle := lipgloss.NewStyle().Foreground(props.Theme.Muted).Padding(0, 1)
	warningStyle := lipgloss.NewStyle().Foreground(props.Theme.Warning).Padding(0, 1)
	dangerStyle := lipgloss.NewStyle().Foreground(props.Theme.Danger).Padding(0, 1)
	helpStyle := lipgloss.NewStyle().Foreground(props.Theme.Muted).Padding(1, 2)

	var content strings.Builder

	content.WriteString(titleStyle.Render("BobaMixer - Optimization Suggestions"))
	content.WriteString("\n\n")

	if props.Error != "" {
		content.WriteString(dangerStyle.Render(fmt.Sprintf("  Error: %s", props.Error)))
		content.WriteString("\n\n")
		content.WriteString(helpStyle.Render(props.NavigationHelp + "  [R] Retry"))
		return content.String()
	}

	content.WriteString(headerStyle.Render("💡 Recommendations (Last 7 Days)"))
	content.WriteString("\n\n")

	if len(props.Suggestions) == 0 {
		content.WriteString(mutedStyle.Render("  ✓ No suggestions - your usage is optimized!"))
		content.WriteString("\n\n")
	} else {
		index := props.SelectedIndex
		if index >= len(props.Suggestions) {
			index = 0
		}

		for i, sugg := range props.Suggestions {
			priorityStyle, priorityIcon := priorityPresentation(sugg.Priority, normalStyle, warningStyle, dangerStyle, mutedStyle)
			typeIcon := suggestionTypeIcon(sugg.Type)
			line := fmt.Sprintf("  %s %s [P%d] %s", priorityIcon, typeIcon, sugg.Priority, sugg.Title)
			if i == index {
				content.WriteString(selectedStyle.Render("▶ " + line))
			} else {
				content.WriteString(priorityStyle.Render(line))
			}
			content.WriteString("\n")
		}

		if len(props.Suggestions) > 0 {
			sugg := props.Suggestions[index]
			content.WriteString("\n")
			content.WriteString(headerStyle.Render("Details"))
			content.WriteString("\n")
			content.WriteString(normalStyle.Render(fmt.Sprintf("  %s", sugg.Description)))
			content.WriteString("\n")
			content.WriteString(normalStyle.Render(fmt.Sprintf("  Impact: %s", sugg.Impact)))
			content.WriteString("\n\n")

			if len(sugg.ActionItems) > 0 {
				content.WriteString(headerStyle.Render("Recommended Actions"))
				content.WriteString("\n")
				for idx, action := range sugg.ActionItems {
					content.WriteString(normalStyle.Render(fmt.Sprintf("  %d. %s", idx+1, action)))
					content.WriteString("\n")
				}
			}
		}
	}

	content.WriteString("\n")
	helpText := props.NavigationHelp
	if props.CommandHelpLine != "" {
		helpText += "\n" + props.CommandHelpLine
	}
	content.WriteString(helpStyle.Render(helpText))

	return content.String()
}

func priorityPresentation(priority int, normalStyle, warningStyle, dangerStyle, mutedStyle lipgloss.Style) (lipgloss.Style, string) {
	switch priority {
	case 5:
		return dangerStyle, "🔴"
	case 4:
		return warningStyle, "🟠"
	case 3:
		return normalStyle, "🟡"
	default:
		return mutedStyle, "🟢"
	}
}

func suggestionTypeIcon(t string) string {
	switch t {
	case "cost":
		return "💰"
	case "profile":
		return "🔄"
	case "budget":
		return "📊"
	case "anomaly":
		return "⚠️"
	case "usage":
		return "📈"
	default:
		return "📈"
	}
}
