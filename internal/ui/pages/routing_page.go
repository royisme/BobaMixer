package pages

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/royisme/bobamixer/internal/ui/components"
	"github.com/royisme/bobamixer/internal/ui/layouts"
	"github.com/royisme/bobamixer/internal/ui/theme"
)

// RoutingPageProps holds the routing tester content.
type RoutingPageProps struct {
	Title           string
	TestTitle       string
	HowToTitle      string
	ExampleTitle    string
	ContextTitle    string
	TestDescription string
	HowToSteps      []string
	ExampleLines    []string
	ContextLines    []string
	NavigationHelp  string
	CommandHelpLine string
}

// RoutingPage composes the routing tester information screen.
type RoutingPage struct {
	title      components.TitleBar
	testDesc   components.Paragraph
	howTo      components.BulletList
	example    components.Paragraph
	context    components.BulletList
	help       components.HelpBar
	testTitle  string
	howToTitle string
	exTitle    string
	ctxTitle   string
}

// NewRoutingPage builds the routing page.
func NewRoutingPage(palette theme.Theme, props RoutingPageProps) RoutingPage {
	styles := theme.NewStyles(palette)
	helpText := strings.TrimSpace(props.NavigationHelp)
	if extra := strings.TrimSpace(props.CommandHelpLine); extra != "" {
		if helpText != "" {
			helpText += "\n"
		}
		helpText += extra
	}

	return RoutingPage{
		title:      components.NewTitleBar(props.Title, styles),
		testDesc:   components.NewParagraph(props.TestDescription, styles),
		howTo:      components.NewBulletList(props.HowToSteps, styles),
		example:    components.NewParagraph(strings.Join(props.ExampleLines, "\n"), styles),
		context:    components.NewBulletList(props.ContextLines, styles),
		help:       components.NewHelpBar(helpText, styles),
		testTitle:  props.TestTitle,
		howToTitle: props.HowToTitle,
		exTitle:    props.ExampleTitle,
		ctxTitle:   props.ContextTitle,
	}
}

// Init satisfies the Page interface.
func (p RoutingPage) Init() tea.Cmd {
	return nil
}

// Update satisfies the Page interface.
func (p RoutingPage) Update(msg tea.Msg) (Page, tea.Cmd) {
	_, cmd1 := p.title.Update(msg)
	_, cmd2 := p.testDesc.Update(msg)
	_, cmd3 := p.howTo.Update(msg)
	_, cmd4 := p.example.Update(msg)
	_, cmd5 := p.context.Update(msg)
	_, cmd6 := p.help.Update(msg)
	return p, tea.Batch(cmd1, cmd2, cmd3, cmd4, cmd5, cmd6)
}

// View assembles the routing tester view.
func (p RoutingPage) View() string {
	blocks := []string{
		layouts.Pad(2, p.title.View()),
		layouts.Gap(1),
		layouts.Section(p.testTitle, p.testDesc.View()),
		layouts.Gap(1),
		layouts.Section(p.howToTitle, p.howTo.View()),
		layouts.Gap(1),
		layouts.Section(p.exTitle, p.example.View()),
		layouts.Gap(1),
		layouts.Section(p.ctxTitle, p.context.View()),
		layouts.Gap(1),
		layouts.Pad(2, p.help.View()),
	}

	return layouts.Column(blocks...)
}
