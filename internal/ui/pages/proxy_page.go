package pages

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/royisme/bobamixer/internal/ui/components"
	"github.com/royisme/bobamixer/internal/ui/layouts"
	"github.com/royisme/bobamixer/internal/ui/theme"
)

// ProxyPageProps holds the proxy screen data.
type ProxyPageProps struct {
	Title           string
	StatusTitle     string
	InfoTitle       string
	ConfigTitle     string
	StatusState     string
	StatusText      string
	StatusIcon      string
	Address         string
	ShowConfig      bool
	NavigationHelp  string
	CommandHelpLine string
	AdditionalNote  string
	InfoLines       []string
	ConfigLines     []string
}

// ProxyPage composes the proxy server control UI.
type ProxyPage struct {
	title   components.TitleBar
	status  components.ProxyStatusPanel
	info    components.BulletList
	config  components.BulletList
	help    components.HelpBar
	section string
	infoHdr string
	cfgHdr  string
	showCfg bool
}

// NewProxyPage builds a ProxyPage using the shared palette.
func NewProxyPage(palette theme.Theme, props ProxyPageProps) ProxyPage {
	styles := theme.NewStyles(palette)

	helpLines := []string{}
	if strings.TrimSpace(props.NavigationHelp) != "" {
		helpLines = append(helpLines, strings.TrimSpace(props.NavigationHelp))
	}
	if strings.TrimSpace(props.CommandHelpLine) != "" {
		helpLines = append(helpLines, strings.TrimSpace(props.CommandHelpLine))
	}
	if strings.TrimSpace(props.AdditionalNote) != "" {
		helpLines = append(helpLines, strings.TrimSpace(props.AdditionalNote))
	}

	helpText := strings.Join(helpLines, "\n")

	return ProxyPage{
		title:   components.NewTitleBar(props.Title, styles),
		status:  components.NewProxyStatusPanel(props.StatusState, props.StatusText, props.StatusIcon, props.Address, styles),
		info:    components.NewBulletList(props.InfoLines, styles),
		config:  components.NewBulletList(props.ConfigLines, styles),
		help:    components.NewHelpBar(helpText, styles),
		section: props.StatusTitle,
		infoHdr: props.InfoTitle,
		cfgHdr:  props.ConfigTitle,
		showCfg: props.ShowConfig,
	}
}

// Init satisfies the Page interface.
func (p ProxyPage) Init() tea.Cmd {
	return nil
}

// Update satisfies the Page interface.
func (p ProxyPage) Update(msg tea.Msg) (Page, tea.Cmd) {
	_, cmd1 := p.title.Update(msg)
	_, cmd2 := p.status.Update(msg)
	_, cmd3 := p.info.Update(msg)
	_, cmd4 := p.config.Update(msg)
	_, cmd5 := p.help.Update(msg)
	return p, tea.Batch(cmd1, cmd2, cmd3, cmd4, cmd5)
}

// View assembles the proxy view.
func (p ProxyPage) View() string {
	blocks := []string{
		layouts.Pad(2, p.title.View()),
		layouts.Gap(1),
		layouts.Section(p.section, p.status.View()),
		layouts.Gap(1),
		layouts.Section(p.infoHdr, p.info.View()),
	}

	if p.showCfg {
		blocks = append(blocks, layouts.Gap(1), layouts.Section(p.cfgHdr, p.config.View()))
	}

	blocks = append(blocks, layouts.Gap(1), layouts.Pad(2, p.help.View()))
	return layouts.Column(blocks...)
}
