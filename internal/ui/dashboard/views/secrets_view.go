package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// SecretProviderRow represents a provider and its key status.
type SecretProviderRow struct {
	DisplayName string
	HasKey      bool
	KeySource   string
}

// SecretsViewProps carries the data required for the secrets view.
type SecretsViewProps struct {
	Theme             ThemePalette
	SecretForm        string
	ShowSecretForm    bool
	SearchBar         string
	EmptyStateMessage string
	Providers         []SecretProviderRow
	SelectedIndex     int
	SecretMessage     string
	NavigationHelp    string
	HelpCommands      string
	SuccessIcon       string
	FailureIcon       string
}

// RenderSecretsView renders the secrets management view.
func RenderSecretsView(props SecretsViewProps) string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(props.Theme.Primary).Padding(0, 2)
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(props.Theme.Success).Padding(1, 2)
	selectedStyle := lipgloss.NewStyle().Foreground(props.Theme.Text).Background(props.Theme.Primary).Bold(true).Padding(0, 1)
	normalStyle := lipgloss.NewStyle().Foreground(props.Theme.Text).Padding(0, 1)
	mutedStyle := lipgloss.NewStyle().Foreground(props.Theme.Muted).Padding(0, 1)
	successStyle := lipgloss.NewStyle().Foreground(props.Theme.Success).Padding(0, 1)
	dangerStyle := lipgloss.NewStyle().Foreground(props.Theme.Danger).Padding(0, 1)
	helpStyle := lipgloss.NewStyle().Foreground(props.Theme.Muted).Padding(1, 2)

	var content strings.Builder

	content.WriteString(titleStyle.Render("BobaMixer - Secrets Management (API Keys)"))
	content.WriteString("\n\n")
	content.WriteString(headerStyle.Render("🔒 API Key Status"))
	content.WriteString("\n\n")

	if props.ShowSecretForm && props.SecretForm != "" {
		content.WriteString(props.SecretForm)
		content.WriteString("\n\n")
	}

	if props.SearchBar != "" {
		content.WriteString(props.SearchBar)
		content.WriteString("\n\n")
	}

	if len(props.Providers) == 0 {
		content.WriteString(mutedStyle.Render("  " + props.EmptyStateMessage))
		content.WriteString("\n\n")
	} else {
		index := props.SelectedIndex
		if index >= len(props.Providers) {
			index = len(props.Providers) - 1
		} else if index < 0 {
			index = 0
		}

		for i, row := range props.Providers {
			statusText := "Missing"
			statusIcon := props.FailureIcon
			statusStyle := dangerStyle
			if row.HasKey {
				statusText = "Configured"
				statusIcon = props.SuccessIcon
				statusStyle = successStyle
			}

			namePart := fmt.Sprintf("  %-25s ", row.DisplayName)
			statusPart := fmt.Sprintf("%s %-15s [%s]", statusIcon, statusText, row.KeySource)
			if i == index {
				line := fmt.Sprintf("%s%s", namePart, statusPart)
				content.WriteString(selectedStyle.Render("▶ " + line))
			} else {
				content.WriteString(normalStyle.Render(namePart))
				content.WriteString(statusStyle.Render(statusPart))
			}
			content.WriteString("\n")
		}
		content.WriteString("\n")
	}

	content.WriteString(headerStyle.Render("🔐 Security"))
	content.WriteString("\n")
	content.WriteString(mutedStyle.Render("  • API keys are stored encrypted in ~/.boba/secrets.yaml"))
	content.WriteString("\n")
	content.WriteString(mutedStyle.Render("  • Keys can also be loaded from environment variables"))
	content.WriteString("\n")
	content.WriteString(mutedStyle.Render("  • Use 'boba edit secrets' to manage keys manually"))
	content.WriteString("\n\n")

	if msg := strings.TrimSpace(props.SecretMessage); msg != "" {
		content.WriteString(normalStyle.Render("  " + msg))
		content.WriteString("\n\n")
	}

	helpLines := []string{props.NavigationHelp}
	if props.HelpCommands != "" {
		helpLines = append(helpLines, props.HelpCommands)
	}

	content.WriteString(helpStyle.Render(strings.Join(helpLines, "  ")))

	return content.String()
}
