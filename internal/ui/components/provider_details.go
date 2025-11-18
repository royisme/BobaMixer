package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/royisme/bobamixer/internal/ui/theme"
)

// ProviderDetails represents metadata for the selected provider row.
type ProviderDetails struct {
	ID           string
	Kind         string
	APIKeySource string
	EnvVar       string
	ShowEnvVar   bool
}

// ProviderDetailsPanel renders a detail block for the selected provider.
type ProviderDetailsPanel struct {
	details *ProviderDetails
	styles  theme.Styles
}

// NewProviderDetailsPanel constructs the panel with the given details.
func NewProviderDetailsPanel(details *ProviderDetails, styles theme.Styles) ProviderDetailsPanel {
	return ProviderDetailsPanel{
		details: details,
		styles:  styles,
	}
}

// Update satisfies the Bubble Tea component interface.
func (c ProviderDetailsPanel) Update(_ tea.Msg) (ProviderDetailsPanel, tea.Cmd) {
	return c, nil
}

// View renders the details block if data is available.
func (c ProviderDetailsPanel) View() string {
	if c.details == nil {
		return ""
	}

	lines := []string{
		fmt.Sprintf("ID: %s", c.details.ID),
		fmt.Sprintf("Kind: %s", c.details.Kind),
		fmt.Sprintf("API Key Source: %s", c.details.APIKeySource),
	}

	if c.details.ShowEnvVar && strings.TrimSpace(c.details.EnvVar) != "" {
		lines = append(lines, fmt.Sprintf("Env Var: %s", c.details.EnvVar))
	}

	var b strings.Builder
	normalStyle := c.styles.Normal
	for _, line := range lines {
		b.WriteString(normalStyle.PaddingLeft(2).Render(line))
		b.WriteString("\n")
	}

	return strings.TrimRight(b.String(), "\n")
}
