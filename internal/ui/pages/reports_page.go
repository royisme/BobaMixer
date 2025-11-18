package pages

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/royisme/bobamixer/internal/ui/components"
	"github.com/royisme/bobamixer/internal/ui/layouts"
	"github.com/royisme/bobamixer/internal/ui/theme"
)

// ReportsPageProps holds the reports screen content.
type ReportsPageProps struct {
	Title           string
	OptionsTitle    string
	OutputTitle     string
	ContentsTitle   string
	Options         []components.ReportOption
	SelectedIndex   int
	Home            string
	NavigationHelp  string
	CommandHelpLine string
}

// ReportsPage composes the usage reports view.
type ReportsPage struct {
	title     components.TitleBar
	options   components.ReportOptionsList
	output    components.Paragraph
	contents  components.BulletList
	help      components.HelpBar
	optionsT  string
	outputT   string
	contentsT string
}

// NewReportsPage builds the reports page.
func NewReportsPage(palette theme.Theme, props ReportsPageProps) ReportsPage {
	styles := theme.NewStyles(palette)
	outputText := fmt.Sprintf("Default path: %s/reports/\nFilename: bobamixer-<date>.<format>", strings.TrimSpace(props.Home))
	helpText := strings.TrimSpace(props.NavigationHelp)
	if extra := strings.TrimSpace(props.CommandHelpLine); extra != "" {
		if helpText != "" {
			helpText += "\n"
		}
		helpText += extra
	}

	return ReportsPage{
		title:     components.NewTitleBar(props.Title, styles),
		options:   components.NewReportOptionsList(props.Options, props.SelectedIndex, styles),
		output:    components.NewParagraph(outputText, styles),
		contents:  components.NewBulletList([]string{
			"Summary statistics (tokens, costs, sessions)",
			"Daily trends and usage patterns",
			"Profile breakdown and comparison",
			"Cost analysis and optimization opportunities",
			"Peak usage times and anomalies",
		}, styles),
		help:      components.NewHelpBar(helpText, styles),
		optionsT:  props.OptionsTitle,
		outputT:   props.OutputTitle,
		contentsT: props.ContentsTitle,
	}
}

// Init satisfies the Page interface.
func (p ReportsPage) Init() tea.Cmd {
	return nil
}

// Update satisfies the Page interface.
func (p ReportsPage) Update(msg tea.Msg) (Page, tea.Cmd) {
	_, cmd1 := p.title.Update(msg)
	_, cmd2 := p.options.Update(msg)
	_, cmd3 := p.output.Update(msg)
	_, cmd4 := p.contents.Update(msg)
	_, cmd5 := p.help.Update(msg)
	return p, tea.Batch(cmd1, cmd2, cmd3, cmd4, cmd5)
}

// View assembles the reports view.
func (p ReportsPage) View() string {
	blocks := []string{
		layouts.Pad(2, p.title.View()),
		layouts.Gap(1),
		layouts.Section(p.optionsT, p.options.View()),
		layouts.Gap(1),
		layouts.Section(p.outputT, p.output.View()),
		layouts.Gap(1),
		layouts.Section(p.contentsT, p.contents.View()),
		layouts.Gap(1),
		layouts.Pad(2, p.help.View()),
	}

	return layouts.Column(blocks...)
}
