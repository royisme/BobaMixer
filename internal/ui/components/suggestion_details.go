package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/royisme/bobamixer/internal/ui/theme"
)

// SuggestionDetails renders the details for the currently selected suggestion.
type SuggestionDetails struct {
	suggestion *Suggestion
	styles     theme.Styles
}

// NewSuggestionDetails constructs the details component.
func NewSuggestionDetails(suggestion *Suggestion, styles theme.Styles) SuggestionDetails {
	return SuggestionDetails{
		suggestion: suggestion,
		styles:     styles,
	}
}

// Update satisfies the Bubble Tea component interface.
func (c SuggestionDetails) Update(_ tea.Msg) (SuggestionDetails, tea.Cmd) {
	return c, nil
}

// View renders description, impact, and action items.
func (c SuggestionDetails) View() string {
	if c.suggestion == nil {
		return ""
	}

	var b strings.Builder
	normal := c.styles.Normal
	normal = normal.PaddingLeft(2)

	if desc := strings.TrimSpace(c.suggestion.Description); desc != "" {
		b.WriteString(normal.Render(desc))
		b.WriteString("\n")
	}

	if impact := strings.TrimSpace(c.suggestion.Impact); impact != "" {
		b.WriteString(normal.Render(fmt.Sprintf("Impact: %s", impact)))
		b.WriteString("\n")
	}

	if len(c.suggestion.ActionItems) > 0 {
		for idx, action := range c.suggestion.ActionItems {
			b.WriteString(normal.Render(fmt.Sprintf("%d. %s", idx+1, action)))
			b.WriteString("\n")
		}
	}

	return strings.TrimRight(b.String(), "\n")
}
