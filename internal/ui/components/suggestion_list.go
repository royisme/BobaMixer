package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/royisme/bobamixer/internal/ui/theme"
)

// Suggestion represents a recommendation entry.
type Suggestion struct {
	Title       string
	Description string
	Impact      string
	ActionItems []string
	Priority    int
	Type        string
}

// SuggestionList renders the suggestion overview entries.
type SuggestionList struct {
	items    []Suggestion
	selected int
	styles   theme.Styles
}

// NewSuggestionList constructs the list component.
func NewSuggestionList(items []Suggestion, selected int, styles theme.Styles) SuggestionList {
	return SuggestionList{
		items:    items,
		selected: selected,
		styles:   styles,
	}
}

// Update satisfies the Bubble Tea component interface.
func (c SuggestionList) Update(_ tea.Msg) (SuggestionList, tea.Cmd) {
	return c, nil
}

// View renders the list of suggestions with priority highlighting.
func (c SuggestionList) View() string {
	if len(c.items) == 0 {
		normalStyle := c.styles.Normal
		return normalStyle.PaddingLeft(2).Render("✓ No suggestions - your usage is optimized!")
	}

	var b strings.Builder
	selected := c.selected
	if selected >= len(c.items) {
		selected = 0
	}

	for idx, sugg := range c.items {
		style, icon := c.priorityStyle(sugg.Priority)
		typeIcon := suggestionTypeIcon(sugg.Type)
		line := fmt.Sprintf("  %s %s [P%d] %s", icon, typeIcon, sugg.Priority, sugg.Title)
		if idx == selected {
			b.WriteString(c.styles.Selected.Render("▶ " + line))
		} else {
			b.WriteString(style.Render(line))
		}
		b.WriteString("\n")
	}

	return strings.TrimRight(b.String(), "\n")
}

func (c SuggestionList) priorityStyle(priority int) (lipgloss.Style, string) {
	switch priority {
	case 5:
		return c.styles.BudgetDanger, "🔴"
	case 4:
		return c.styles.BudgetWarn, "🟠"
	case 3:
		return c.styles.Normal, "🟡"
	default:
		return c.styles.Normal, "🟢"
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
