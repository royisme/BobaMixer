package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/royisme/bobamixer/internal/ui/theme"
)

// ProxyStatusPanel renders proxy server status details.
type ProxyStatusPanel struct {
	state      string
	statusText string
	statusIcon string
	address    string
	styles     theme.Styles
}

// NewProxyStatusPanel constructs the panel with the provided state metadata.
func NewProxyStatusPanel(state string, statusText string, statusIcon string, address string, styles theme.Styles) ProxyStatusPanel {
	return ProxyStatusPanel{
		state:      state,
		statusText: statusText,
		statusIcon: statusIcon,
		address:    address,
		styles:     styles,
	}
}

// Update satisfies the Bubble Tea component interface.
func (c ProxyStatusPanel) Update(_ tea.Msg) (ProxyStatusPanel, tea.Cmd) {
	return c, nil
}

// View renders the status and address lines.
func (c ProxyStatusPanel) View() string {
	var statusStyle string
	switch strings.ToLower(c.state) {
	case "running":
		statusStyle = c.styles.BudgetOK.Render(c.statusIcon + " " + c.statusText)
	case "stopped":
		statusStyle = c.styles.BudgetDanger.Render(c.statusIcon + " " + c.statusText)
	default:
		statusStyle = c.styles.Normal.Render(c.statusIcon + " " + c.statusText)
	}

	lines := []string{
		fmt.Sprintf("Status:   %s", statusStyle),
		fmt.Sprintf("Address:  %s", c.address),
	}

	var b strings.Builder
	normalStyle := c.styles.Normal
	for _, line := range lines {
		b.WriteString(normalStyle.PaddingLeft(2).Render(line))
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n")
}
