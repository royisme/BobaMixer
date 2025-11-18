package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/royisme/bobamixer/internal/ui/theme"
)

// ProviderRow represents a compact provider entry for the list component.
type ProviderRow struct {
	DisplayName  string
	BaseURL      string
	DefaultModel string
	Enabled      bool
	HasAPIKey    bool
}

// ProviderList renders provider rows with selection state and health/status icons.
type ProviderList struct {
	rows           []ProviderRow
	selectedIndex  int
	enabledIcon    string
	disabledIcon   string
	keyPresentIcon string
	keyMissingIcon string
	emptyState     string
	styles         theme.Styles
}

// NewProviderList creates a ProviderList component.
func NewProviderList(rows []ProviderRow, selected int, emptyState string, icons ProviderListIcons, styles theme.Styles) ProviderList {
	return ProviderList{
		rows:           rows,
		selectedIndex:  selected,
		enabledIcon:    icons.Enabled,
		disabledIcon:   icons.Disabled,
		keyPresentIcon: icons.KeyPresent,
		keyMissingIcon: icons.KeyMissing,
		emptyState:     strings.TrimSpace(emptyState),
		styles:         styles,
	}
}

// ProviderListIcons groups the glyphs used in the provider list.
type ProviderListIcons struct {
	Enabled    string
	Disabled   string
	KeyPresent string
	KeyMissing string
}

// Update satisfies the Bubble Tea component interface.
func (c ProviderList) Update(_ tea.Msg) (ProviderList, tea.Cmd) {
	return c, nil
}

// View renders either the empty state or the list of providers.
func (c ProviderList) View() string {
	if len(c.rows) == 0 {
		if c.emptyState == "" {
			return ""
		}
		normalStyle := c.styles.Normal
		return normalStyle.PaddingLeft(2).Render(c.emptyState)
	}

	var b strings.Builder
	selected := c.selectedIndex
	if selected >= len(c.rows) {
		selected = len(c.rows) - 1
	}

	for idx, row := range c.rows {
		enabledIcon := c.enabledIcon
		if !row.Enabled {
			enabledIcon = c.disabledIcon
		}

		keyIcon := c.keyMissingIcon
		if row.HasAPIKey {
			keyIcon = c.keyPresentIcon
		}

		line := fmt.Sprintf("  %s %s %-25s %-35s %s",
			enabledIcon,
			keyIcon,
			row.DisplayName,
			row.BaseURL,
			row.DefaultModel,
		)

		if idx == selected {
			b.WriteString(c.styles.Selected.Render("▶ " + line))
		} else {
			b.WriteString(c.styles.Normal.Render("  " + line))
		}
		b.WriteString("\n")
	}

	return strings.TrimRight(b.String(), "\n")
}
