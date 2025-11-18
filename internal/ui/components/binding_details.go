// Package components provides reusable UI components for the BobaMixer TUI.
package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/royisme/bobamixer/internal/ui/theme"
)

// BindingDetails represents the selected binding metadata.
type BindingDetails struct {
	ToolID        string
	ProviderID    string
	UseProxy      bool
	ModelOverride string
}

// BindingDetailsPanel renders the binding metadata.
type BindingDetailsPanel struct {
	details *BindingDetails
	styles  theme.Styles
}

// NewBindingDetailsPanel constructs the panel.
func NewBindingDetailsPanel(details *BindingDetails, styles theme.Styles) BindingDetailsPanel {
	return BindingDetailsPanel{
		details: details,
		styles:  styles,
	}
}

// Update satisfies the Bubble Tea component interface.
func (c BindingDetailsPanel) Update(_ tea.Msg) (BindingDetailsPanel, tea.Cmd) {
	return c, nil
}

// View renders the binding details if available.
func (c BindingDetailsPanel) View() string {
	if c.details == nil {
		return ""
	}

	lines := []string{
		fmt.Sprintf("Tool ID: %s", c.details.ToolID),
		fmt.Sprintf("Provider ID: %s", c.details.ProviderID),
		fmt.Sprintf("Use Proxy: %t", c.details.UseProxy),
	}

	if model := strings.TrimSpace(c.details.ModelOverride); model != "" {
		lines = append(lines, fmt.Sprintf("Model Override: %s", model))
	}

	var b strings.Builder
	for _, line := range lines {
		normalStyle := c.styles.Normal
		b.WriteString(normalStyle.PaddingLeft(2).Render(line))
		b.WriteString("\n")
	}

	return strings.TrimRight(b.String(), "\n")
}
