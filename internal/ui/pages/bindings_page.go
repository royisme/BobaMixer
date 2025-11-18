// Package pages provides page-level UI components for the BobaMixer TUI.
package pages

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/royisme/bobamixer/internal/ui/components"
	"github.com/royisme/bobamixer/internal/ui/layouts"
	"github.com/royisme/bobamixer/internal/ui/theme"
)

// BindingsPageProps holds data to render the bindings screen.
type BindingsPageProps struct {
	Title              string
	SectionTitle       string
	DetailsTitle       string
	BindingForm        string
	ShowBindingForm    bool
	BindingFormMessage string
	SearchBar          string
	EmptyStateMessage  string
	Bindings           []components.BindingRow
	SelectedIndex      int
	Details            *components.BindingDetails
	NavigationHelp     string
	ActionHelp         string
	ProxyEnabledIcon   string
	ProxyDisabledIcon  string
}

// BindingsPage composes the tool-provider binding view.
type BindingsPage struct {
	title       components.TitleBar
	formMessage components.InfoMessage
	list        components.BindingList
	details     components.BindingDetailsPanel
	help        components.HelpBar
	searchBar   string
	section     string
	detailTitle string
	form        string
	showForm    bool
}

// NewBindingsPage builds the bindings page.
func NewBindingsPage(palette theme.Theme, props BindingsPageProps) BindingsPage {
	styles := theme.NewStyles(palette)

	helpText := strings.TrimSpace(props.NavigationHelp)
	if extra := strings.TrimSpace(props.ActionHelp); extra != "" {
		if helpText != "" {
			helpText += "  "
		}
		helpText += extra
	}

	return BindingsPage{
		title:       components.NewTitleBar(props.Title, styles),
		formMessage: components.NewInfoMessage(props.BindingFormMessage, styles),
		list:        components.NewBindingList(props.Bindings, props.SelectedIndex, props.EmptyStateMessage, props.ProxyEnabledIcon, props.ProxyDisabledIcon, styles),
		details:     components.NewBindingDetailsPanel(props.Details, styles),
		help:        components.NewHelpBar(helpText, styles),
		searchBar:   props.SearchBar,
		section:     props.SectionTitle,
		detailTitle: props.DetailsTitle,
		form:        props.BindingForm,
		showForm:    props.ShowBindingForm,
	}
}

// Init satisfies the Page interface.
func (p BindingsPage) Init() tea.Cmd {
	return nil
}

// Update satisfies the Page interface (bindings page is static).
func (p BindingsPage) Update(msg tea.Msg) (Page, tea.Cmd) {
	_, cmd1 := p.title.Update(msg)
	_, cmd2 := p.formMessage.Update(msg)
	_, cmd3 := p.list.Update(msg)
	_, cmd4 := p.details.Update(msg)
	_, cmd5 := p.help.Update(msg)
	return p, tea.Batch(cmd1, cmd2, cmd3, cmd4, cmd5)
}

// View assembles the bindings view using the layout DSL.
func (p BindingsPage) View() string {
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

	detailsView := p.details.View()

	layoutBlocks := []string{
		layouts.Pad(2, p.title.View()),
		layouts.Gap(1),
		layouts.Section(p.section, layouts.Column(body...)),
	}

	if detailsView != "" {
		layoutBlocks = append(layoutBlocks, layouts.Gap(1), layouts.Section(p.detailTitle, detailsView))
	}

	layoutBlocks = append(layoutBlocks, layouts.Gap(1), layouts.Pad(2, p.help.View()))

	return layouts.Column(layoutBlocks...)
}

func (p BindingsPage) renderForm() string {
	if p.showForm {
		if strings.TrimSpace(p.form) != "" {
			return p.form
		}
		return ""
	}
	return p.formMessage.View()
}
