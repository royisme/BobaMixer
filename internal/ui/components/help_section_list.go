package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/royisme/bobamixer/internal/ui/theme"
)

// HelpSection describes a high-level dashboard grouping.
type HelpSection struct {
	Name     string
	Shortcut string
	Views    []string
}

// HelpSectionList renders the list of sections and their view shortcuts.
type HelpSectionList struct {
	sections []HelpSection
	styles   theme.Styles
}

// NewHelpSectionList constructs a new HelpSectionList component.
func NewHelpSectionList(sections []HelpSection, styles theme.Styles) HelpSectionList {
	return HelpSectionList{
		sections: sections,
		styles:   styles,
	}
}

// Update does not mutate state because section metadata is static.
func (c HelpSectionList) Update(_ tea.Msg) (HelpSectionList, tea.Cmd) {
	return c, nil
}

// View renders each section, its shortcut, and the associated views.
func (c HelpSectionList) View() string {
	var b strings.Builder
	shortcutStyle := c.styles.Selected
	shortcutStyle = shortcutStyle.PaddingLeft(0).PaddingRight(1)
	textStyle := c.styles.Normal
	textStyle = textStyle.PaddingLeft(0)

	for _, section := range c.sections {
		if section.Shortcut == "" {
			continue
		}

		b.WriteString(textStyle.Render("  "))
		b.WriteString(shortcutStyle.Render(fmt.Sprintf("[%s]", section.Shortcut)))
		joinedViews := strings.Join(section.Views, ", ")
		b.WriteString(textStyle.Render(fmt.Sprintf("  %s → %s", section.Name, joinedViews)))
		b.WriteString("\n")
	}

	// Append help overlay toggle instruction.
	b.WriteString(textStyle.Render("  "))
	b.WriteString(shortcutStyle.Render("[?]"))
	b.WriteString(textStyle.Render("  Toggle this help overlay"))

	return strings.TrimRight(b.String(), "\n")
}
