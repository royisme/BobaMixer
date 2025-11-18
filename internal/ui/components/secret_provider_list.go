package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/royisme/bobamixer/internal/ui/theme"
)

// SecretProviderRow represents provider secret status for the secrets page.
type SecretProviderRow struct {
	DisplayName string
	HasKey      bool
	KeySource   string
}

// SecretProviderList renders rows with configured/missing key states.
type SecretProviderList struct {
	rows        []SecretProviderRow
	selected    int
	successIcon string
	failureIcon string
	emptyState  string
	styles      theme.Styles
}

// NewSecretProviderList constructs the list component.
func NewSecretProviderList(rows []SecretProviderRow, selected int, emptyState string, successIcon string, failureIcon string, styles theme.Styles) SecretProviderList {
	return SecretProviderList{
		rows:        rows,
		selected:    selected,
		successIcon: successIcon,
		failureIcon: failureIcon,
		emptyState:  strings.TrimSpace(emptyState),
		styles:      styles,
	}
}

// Update satisfies the Bubble Tea component interface.
func (c SecretProviderList) Update(_ tea.Msg) (SecretProviderList, tea.Cmd) {
	return c, nil
}

// View renders the providers or the empty state message.
func (c SecretProviderList) View() string {
	if len(c.rows) == 0 {
		if c.emptyState == "" {
			return ""
		}
		normalStyle := c.styles.Normal
		return normalStyle.PaddingLeft(2).Render(c.emptyState)
	}

	var b strings.Builder
	selected := c.selected
	if selected >= len(c.rows) {
		selected = len(c.rows) - 1
	}
	if selected < 0 {
		selected = 0
	}

	for idx, row := range c.rows {
		statusText := "Missing"
		statusIcon := c.failureIcon
		statusStyle := c.styles.BudgetDanger
		if row.HasKey {
			statusText = "Configured"
			statusIcon = c.successIcon
			statusStyle = c.styles.BudgetOK
		}

		namePart := fmt.Sprintf("  %-25s ", row.DisplayName)
		statusPart := fmt.Sprintf("%s %-15s [%s]", statusIcon, statusText, row.KeySource)

		if idx == selected {
			b.WriteString(c.styles.Selected.Render("▶ " + namePart + statusPart))
		} else {
			b.WriteString(c.styles.Normal.Render(namePart))
			b.WriteString(statusStyle.Render(statusPart))
		}
		b.WriteString("\n")
	}

	return strings.TrimRight(b.String(), "\n")
}
