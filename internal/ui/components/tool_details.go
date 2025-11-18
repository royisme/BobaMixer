package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/royisme/bobamixer/internal/ui/theme"
)

// ToolDetails represents metadata of the selected CLI tool.
type ToolDetails struct {
	ID          string
	ConfigType  string
	ConfigPath  string
	Description string
}

// ToolDetailsPanel renders the details block for a tool.
type ToolDetailsPanel struct {
	details *ToolDetails
	styles  theme.Styles
}

// NewToolDetailsPanel constructs the panel.
func NewToolDetailsPanel(details *ToolDetails, styles theme.Styles) ToolDetailsPanel {
	return ToolDetailsPanel{
		details: details,
		styles:  styles,
	}
}

// Update satisfies the Bubble Tea component interface.
func (c ToolDetailsPanel) Update(_ tea.Msg) (ToolDetailsPanel, tea.Cmd) {
	return c, nil
}

// View renders the tool details if available.
func (c ToolDetailsPanel) View() string {
	if c.details == nil {
		return ""
	}

	lines := []string{
		fmt.Sprintf("ID: %s", c.details.ID),
		fmt.Sprintf("Config Type: %s", c.details.ConfigType),
		fmt.Sprintf("Config Path: %s", c.details.ConfigPath),
	}

	if desc := strings.TrimSpace(c.details.Description); desc != "" {
		lines = append(lines, fmt.Sprintf("Description: %s", desc))
	}

	var b strings.Builder
	normalStyle := c.styles.Normal
	for _, line := range lines {
		b.WriteString(normalStyle.PaddingLeft(2).Render(line))
		b.WriteString("\n")
	}

	return strings.TrimRight(b.String(), "\n")
}
