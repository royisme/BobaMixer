package pages

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/royisme/bobamixer/internal/ui/components"
	"github.com/royisme/bobamixer/internal/ui/layouts"
	"github.com/royisme/bobamixer/internal/ui/theme"
)

// DashboardPageProps describes the data needed to render the dashboard summary.
type DashboardPageProps struct {
	Title          string
	Table          string
	Message        string
	ProxyIcon      string
	ProxyStatus    string
	NavigationHelp string
	ActionHelp     string
}

// DashboardPage composes the dashboard overview using the layout DSL.
type DashboardPage struct {
	title   components.TitleBar
	proxy   components.ProxyStatus
	message components.StatusMessage
	help    components.HelpBar
	table   string
}

// NewDashboardPage constructs a DashboardPage with the supplied palette.
func NewDashboardPage(palette theme.Theme, props DashboardPageProps) DashboardPage {
	styles := theme.NewStyles(palette)

	helpText := strings.TrimSpace(props.NavigationHelp)
	if extra := strings.TrimSpace(props.ActionHelp); extra != "" {
		if helpText != "" {
			helpText += "  "
		}
		helpText += extra
	}

	return DashboardPage{
		title:   components.NewTitleBar(props.Title, styles),
		proxy:   components.NewProxyStatus(props.ProxyIcon, props.ProxyStatus, styles),
		message: components.NewStatusMessage(props.Message, palette.Success),
		help:    components.NewHelpBar(helpText, styles),
		table:   props.Table,
	}
}

// Init satisfies the Page interface; no initialization commands are needed.
func (p DashboardPage) Init() tea.Cmd {
	return nil
}

// Update satisfies the Page interface and keeps the dashboard immutable for now.
func (p DashboardPage) Update(msg tea.Msg) (Page, tea.Cmd) {
	_, cmd1 := p.title.Update(msg)
	_, cmd2 := p.proxy.Update(msg)
	_, cmd3 := p.message.Update(msg)
	_, cmd4 := p.help.Update(msg)
	return p, tea.Batch(cmd1, cmd2, cmd3, cmd4)
}

// View composes the dashboard blocks using the layout helpers.
func (p DashboardPage) View() string {
	blocks := []string{
		layouts.Pad(2, p.title.View()),
		layouts.Gap(1),
		layouts.Pad(2, p.proxy.View()),
		layouts.Gap(1),
		p.table,
	}

	if msg := p.message.View(); msg != "" {
		blocks = append(blocks, layouts.Gap(1), layouts.Pad(2, msg))
	}

	blocks = append(blocks, layouts.Gap(1), layouts.Pad(2, p.help.View()))

	return layouts.Column(blocks...)
}
