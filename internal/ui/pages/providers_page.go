package pages

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/royisme/bobamixer/internal/ui/components"
	"github.com/royisme/bobamixer/internal/ui/layouts"
	"github.com/royisme/bobamixer/internal/ui/theme"
)

// ProvidersPageProps holds the information required to render the providers screen.
type ProvidersPageProps struct {
	Title               string
	SectionTitle        string
	DetailsTitle        string
	ProviderForm        string
	ShowProviderForm    bool
	ProviderFormMessage string
	SearchBar           string
	EmptyStateMessage   string
	Providers           []components.ProviderRow
	SelectedIndex       int
	Details             *components.ProviderDetails
	NavigationHelp      string
	ActionHelp          string
	Icons               components.ProviderListIcons
}

// ProvidersPage composes the providers management UI.
type ProvidersPage struct {
	title        components.TitleBar
	formMessage  components.InfoMessage
	list         components.ProviderList
	details      components.ProviderDetailsPanel
	help         components.HelpBar
	providerForm string
	showForm     bool
	searchBar    string
	sectionTitle string
	detailsTitle string
}

// NewProvidersPage creates a ProvidersPage for the supplied props.
func NewProvidersPage(palette theme.Theme, props ProvidersPageProps) ProvidersPage {
	styles := theme.NewStyles(palette)

	helpText := strings.TrimSpace(props.NavigationHelp)
	if extra := strings.TrimSpace(props.ActionHelp); extra != "" {
		if helpText != "" {
			helpText += "  "
		}
		helpText += extra
	}

	return ProvidersPage{
		title:        components.NewTitleBar(props.Title, styles),
		formMessage:  components.NewInfoMessage(props.ProviderFormMessage, styles),
		list:         components.NewProviderList(props.Providers, props.SelectedIndex, props.EmptyStateMessage, props.Icons, styles),
		details:      components.NewProviderDetailsPanel(props.Details, styles),
		help:         components.NewHelpBar(helpText, styles),
		providerForm: props.ProviderForm,
		showForm:     props.ShowProviderForm,
		searchBar:    props.SearchBar,
		sectionTitle: props.SectionTitle,
		detailsTitle: props.DetailsTitle,
	}
}

// Init satisfies the Page interface.
func (p ProvidersPage) Init() tea.Cmd {
	return nil
}

// Update satisfies the Page interface (providers page is static for now).
func (p ProvidersPage) Update(msg tea.Msg) (Page, tea.Cmd) {
	_, cmd1 := p.title.Update(msg)
	_, cmd2 := p.formMessage.Update(msg)
	_, cmd3 := p.list.Update(msg)
	_, cmd4 := p.details.Update(msg)
	_, cmd5 := p.help.Update(msg)
	return p, tea.Batch(cmd1, cmd2, cmd3, cmd4, cmd5)
}

// View assembles the providers view using the layout DSL.
func (p ProvidersPage) View() string {
	bodyBlocks := []string{}

	if form := p.renderForm(); form != "" {
		bodyBlocks = append(bodyBlocks, form)
	}
	if bar := strings.TrimSpace(p.searchBar); bar != "" {
		bodyBlocks = append(bodyBlocks, bar)
	}
	if list := p.list.View(); list != "" {
		bodyBlocks = append(bodyBlocks, list)
	}

	detailsView := p.details.View()

	layoutBlocks := []string{
		layouts.Pad(2, p.title.View()),
		layouts.Gap(1),
		layouts.Section(p.sectionTitle, layouts.Column(bodyBlocks...)),
	}

	if detailsView != "" {
		layoutBlocks = append(layoutBlocks, layouts.Gap(1), layouts.Section(p.detailsTitle, detailsView))
	}

	layoutBlocks = append(layoutBlocks, layouts.Gap(1), layouts.Pad(2, p.help.View()))

	return layouts.Column(layoutBlocks...)
}

func (p ProvidersPage) renderForm() string {
	if p.showForm {
		if strings.TrimSpace(p.providerForm) != "" {
			return p.providerForm
		}
		return ""
	}
	return p.formMessage.View()
}
