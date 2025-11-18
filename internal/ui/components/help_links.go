package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/royisme/bobamixer/internal/ui/theme"
)

// HelpLink represents a documentation reference.
type HelpLink struct {
	Label string
	URL   string
}

// HelpLinks renders a list of documentation references.
type HelpLinks struct {
	links  []HelpLink
	styles theme.Styles
}

// NewHelpLinks constructs the documentation links component.
func NewHelpLinks(links []HelpLink, styles theme.Styles) HelpLinks {
	return HelpLinks{
		links:  links,
		styles: styles,
	}
}

// Update keeps the component immutable because documentation links are static.
func (c HelpLinks) Update(_ tea.Msg) (HelpLinks, tea.Cmd) {
	return c, nil
}

// View renders each documentation link on its own line.
func (c HelpLinks) View() string {
	if len(c.links) == 0 {
		return ""
	}
	var b strings.Builder
	textStyle := c.styles.Normal
	textStyle = textStyle.PaddingLeft(0)

	for _, link := range c.links {
		label := strings.TrimSpace(link.Label)
		url := strings.TrimSpace(link.URL)
		if label == "" && url == "" {
			continue
		}
		var line string
		switch {
		case label == "":
			line = url
		case url == "":
			line = label
		default:
			line = label + ": " + url
		}
		b.WriteString(textStyle.Render("  " + line))
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n")
}
