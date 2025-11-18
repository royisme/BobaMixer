package pages

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/royisme/bobamixer/internal/ui/components"
	"github.com/royisme/bobamixer/internal/ui/layouts"
	"github.com/royisme/bobamixer/internal/ui/theme"
)

// SecretsPageProps holds the content required to render the secrets screen.
type SecretsPageProps struct {
	Title             string
	StatusTitle       string
	SecurityTitle     string
	SecretForm        string
	ShowSecretForm    bool
	SearchBar         string
	EmptyStateMessage string
	Providers         []components.SecretProviderRow
	SelectedIndex     int
	SecretMessage     string
	NavigationHelp    string
	ActionHelp        string
	SuccessIcon       string
	FailureIcon       string
	SecurityTips      []string
}

// SecretsPage composes the API key management UI.
type SecretsPage struct {
	title       components.TitleBar
	list        components.SecretProviderList
	tips        components.BulletList
	message     components.InfoMessage
	help        components.HelpBar
	secretForm  string
	showForm    bool
	searchBar   string
	statusTitle string
	security    string
}

// NewSecretsPage builds a SecretsPage for the supplied props.
func NewSecretsPage(palette theme.Theme, props SecretsPageProps) SecretsPage {
	styles := theme.NewStyles(palette)

	helpText := strings.TrimSpace(props.NavigationHelp)
	if extra := strings.TrimSpace(props.ActionHelp); extra != "" {
		if helpText != "" {
			helpText += "  "
		}
		helpText += extra
	}

	return SecretsPage{
		title:       components.NewTitleBar(props.Title, styles),
		list:        components.NewSecretProviderList(props.Providers, props.SelectedIndex, props.EmptyStateMessage, props.SuccessIcon, props.FailureIcon, styles),
		tips:        components.NewBulletList(props.SecurityTips, styles),
		message:     components.NewInfoMessage(props.SecretMessage, styles),
		help:        components.NewHelpBar(helpText, styles),
		secretForm:  props.SecretForm,
		showForm:    props.ShowSecretForm,
		searchBar:   props.SearchBar,
		statusTitle: props.StatusTitle,
		security:    props.SecurityTitle,
	}
}

// Init satisfies the Page interface.
func (p SecretsPage) Init() tea.Cmd {
	return nil
}

// Update satisfies the Page interface.
func (p SecretsPage) Update(msg tea.Msg) (Page, tea.Cmd) {
	_, cmd1 := p.title.Update(msg)
	_, cmd2 := p.list.Update(msg)
	_, cmd3 := p.tips.Update(msg)
	_, cmd4 := p.message.Update(msg)
	_, cmd5 := p.help.Update(msg)
	return p, tea.Batch(cmd1, cmd2, cmd3, cmd4, cmd5)
}

// View assembles the secrets view.
func (p SecretsPage) View() string {
	body := []string{}
	if form := p.renderForm(); form != "" {
		body = append(body, form)
	}
	if bar := strings.TrimSpace(p.searchBar); bar != "" {
		body = append(body, bar)
	}
	if list := p.list.View(); list != "" {
		body = append(body, list)
	}

	if msg := strings.TrimSpace(p.message.View()); msg != "" {
		body = append(body, msg)
	}

	layoutBlocks := []string{
		layouts.Pad(2, p.title.View()),
		layouts.Gap(1),
		layouts.Section(p.statusTitle, layouts.Column(body...)),
	}

	if tips := p.tips.View(); tips != "" {
		layoutBlocks = append(layoutBlocks, layouts.Gap(1), layouts.Section(p.security, tips))
	}

	layoutBlocks = append(layoutBlocks, layouts.Gap(1), layouts.Pad(2, p.help.View()))
	return layouts.Column(layoutBlocks...)
}

func (p SecretsPage) renderForm() string {
	if p.showForm {
		return p.secretForm
	}
	return ""
}
