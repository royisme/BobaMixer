package components

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/royisme/bobamixer/internal/ui/theme"
)

// ProxyStatus renders the proxy indicator line.
type ProxyStatus struct {
	icon   string
	status string
	styles theme.Styles
}

// NewProxyStatus constructs a ProxyStatus component.
func NewProxyStatus(icon string, status string, styles theme.Styles) ProxyStatus {
	return ProxyStatus{
		icon:   icon,
		status: status,
		styles: styles,
	}
}

// Update satisfies the Bubble Tea component contract (no mutations needed).
func (c ProxyStatus) Update(_ tea.Msg) (ProxyStatus, tea.Cmd) {
	return c, nil
}

// View renders the proxy indicator using the shared typography.
func (c ProxyStatus) View() string {
	text := fmt.Sprintf("Proxy: %s %s", c.icon, c.status)
	normalStyle := c.styles.Normal
	return normalStyle.PaddingLeft(0).Render(text)
}
