package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ProviderRow represents a compact provider entry for the list.
type ProviderRow struct {
	DisplayName  string
	BaseURL      string
	DefaultModel string
	Enabled      bool
	HasAPIKey    bool
}

// ProviderDetails captures the selected provider metadata.
type ProviderDetails struct {
	ID           string
	Kind         string
	APIKeySource string
	EnvVar       string
	ShowEnvVar   bool
}

// ProvidersViewProps carries all data required for rendering the providers view.
type ProvidersViewProps struct {
	Theme               ThemePalette
	ProviderForm        string
	ShowProviderForm    bool
	ProviderFormMessage string
	SearchBar           string
	EmptyStateMessage   string
	Providers           []ProviderRow
	SelectedIndex       int
	Details             *ProviderDetails
	NavigationHelp      string
	HelpCommands        string
	EnabledIcon         string
	DisabledIcon        string
	KeyPresentIcon      string
	KeyMissingIcon      string
}

// RenderProvidersView renders the providers management UI.
func RenderProvidersView(props ProvidersViewProps) string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(props.Theme.Primary).Padding(0, 2)
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(props.Theme.Success).Padding(1, 2)
	selectedStyle := lipgloss.NewStyle().Foreground(props.Theme.Text).Background(props.Theme.Primary).Bold(true).Padding(0, 1)
	normalStyle := lipgloss.NewStyle().Foreground(props.Theme.Text).Padding(0, 1)
	mutedStyle := lipgloss.NewStyle().Foreground(props.Theme.Muted).Padding(0, 1)
	helpStyle := lipgloss.NewStyle().Foreground(props.Theme.Muted).Padding(1, 2)

	var content strings.Builder
	content.WriteString(titleStyle.Render("BobaMixer - AI Providers Management"))
	content.WriteString("\n\n")

	content.WriteString(headerStyle.Render("📡 Available Providers"))
	content.WriteString("\n\n")

	if props.ShowProviderForm && strings.TrimSpace(props.ProviderForm) != "" {
		content.WriteString(props.ProviderForm)
		content.WriteString("\n\n")
	} else if msg := strings.TrimSpace(props.ProviderFormMessage); msg != "" {
		content.WriteString(mutedStyle.Render("  " + msg))
		content.WriteString("\n\n")
	}

	if bar := strings.TrimSpace(props.SearchBar); bar != "" {
		content.WriteString(bar)
		content.WriteString("\n\n")
	}

	rows := props.Providers
	if len(rows) == 0 {
		if msg := strings.TrimSpace(props.EmptyStateMessage); msg != "" {
			content.WriteString(mutedStyle.Render("  " + msg))
			content.WriteString("\n")
		}
	} else {
		selectedIndex := props.SelectedIndex
		if selectedIndex >= len(rows) {
			selectedIndex = len(rows) - 1
		}

		for idx, row := range rows {
			enabledIcon := props.EnabledIcon
			if !row.Enabled {
				enabledIcon = props.DisabledIcon
			}

			keyIcon := props.KeyMissingIcon
			if row.HasAPIKey {
				keyIcon = props.KeyPresentIcon
			}

			line := fmt.Sprintf("  %s %s %-25s %-35s %s",
				enabledIcon,
				keyIcon,
				row.DisplayName,
				row.BaseURL,
				row.DefaultModel,
			)

			if idx == selectedIndex {
				content.WriteString(selectedStyle.Render("▶ " + line))
			} else {
				content.WriteString(normalStyle.Render("  " + line))
			}
			content.WriteString("\n")
		}
	}

	content.WriteString("\n")
	if props.Details != nil {
		content.WriteString(headerStyle.Render("Details"))
		content.WriteString("\n")
		content.WriteString(normalStyle.Render(fmt.Sprintf("  ID: %s", props.Details.ID)))
		content.WriteString("\n")
		content.WriteString(normalStyle.Render(fmt.Sprintf("  Kind: %s", props.Details.Kind)))
		content.WriteString("\n")
		content.WriteString(normalStyle.Render(fmt.Sprintf("  API Key Source: %s", props.Details.APIKeySource)))
		content.WriteString("\n")
		if props.Details.ShowEnvVar {
			content.WriteString(normalStyle.Render(fmt.Sprintf("  Env Var: %s", props.Details.EnvVar)))
			content.WriteString("\n")
		}
		content.WriteString("\n")
	}

	helpText := props.NavigationHelp
	if cmds := strings.TrimSpace(props.HelpCommands); cmds != "" {
		helpText += "  " + cmds
	}
	content.WriteString(helpStyle.Render(helpText))

	return content.String()
}
