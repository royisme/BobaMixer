package pages

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/royisme/bobamixer/internal/ui/components"
	"github.com/royisme/bobamixer/internal/ui/layouts"
	"github.com/royisme/bobamixer/internal/ui/theme"
)

// ToolsPageProps holds the data required to render the tools screen.
type ToolsPageProps struct {
	Title             string
	SectionTitle      string
	DetailsTitle      string
	SearchBar         string
	EmptyStateMessage string
	Tools             []components.ToolRow
	SelectedIndex     int
	Details           *components.ToolDetails
	NavigationHelp    string
	ActionHelp        string
	BoundIcon         string
	UnboundIcon       string
}

// ToolsPage composes the CLI tools view.
type ToolsPage struct {
	title       components.TitleBar
	list        components.ToolList
	details     components.ToolDetailsPanel
	help        components.HelpBar
	searchBar   string
	section     string
	detailTitle string
}

// NewToolsPage builds the tools page using the shared theme palette.
func NewToolsPage(palette theme.Theme, props ToolsPageProps) ToolsPage {
	styles := theme.NewStyles(palette)

	helpText := strings.TrimSpace(props.NavigationHelp)
	if extra := strings.TrimSpace(props.ActionHelp); extra != "" {
		if helpText != "" {
			helpText += "  "
		}
		helpText += extra
	}

	return ToolsPage{
		title:       components.NewTitleBar(props.Title, styles),
		list:        components.NewToolList(props.Tools, props.SelectedIndex, props.EmptyStateMessage, props.BoundIcon, props.UnboundIcon, styles),
		details:     components.NewToolDetailsPanel(props.Details, styles),
		help:        components.NewHelpBar(helpText, styles),
		searchBar:   props.SearchBar,
		section:     props.SectionTitle,
		detailTitle: props.DetailsTitle,
	}
}

// Init satisfies the Page interface.
func (p ToolsPage) Init() tea.Cmd {
	return nil
}

// Update satisfies the Page interface (tools page is static).
func (p ToolsPage) Update(msg tea.Msg) (Page, tea.Cmd) {
	_, cmd1 := p.title.Update(msg)
	_, cmd2 := p.list.Update(msg)
	_, cmd3 := p.details.Update(msg)
	_, cmd4 := p.help.Update(msg)
	return p, tea.Batch(cmd1, cmd2, cmd3, cmd4)
}

// View assembles the tools view with the layout DSL.
func (p ToolsPage) View() string {
	body := []string{}
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
