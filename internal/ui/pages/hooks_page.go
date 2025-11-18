package pages

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/royisme/bobamixer/internal/ui/components"
	"github.com/royisme/bobamixer/internal/ui/layouts"
	"github.com/royisme/bobamixer/internal/ui/theme"
)

// HooksPageProps holds the git hooks screen data.
type HooksPageProps struct {
	Title           string
	RepoTitle       string
	HooksTitle      string
	BenefitsTitle   string
	ActivityTitle   string
	RepoPath        string
	HooksInstalled  bool
	Hooks           []components.HookInfo
	NavigationHelp  string
	CommandHelpLine string
	ActiveIcon      string
	InactiveIcon    string
}

// HooksPage composes the git hooks management view.
type HooksPage struct {
	title      components.TitleBar
	repoInfo   components.Paragraph
	hookList   components.HookList
	benefits   components.BulletList
	activity   components.Paragraph
	help       components.HelpBar
	repoTitle  string
	hooksTitle string
	benefitsT  string
	activityT  string
}

// NewHooksPage builds the hooks page.
func NewHooksPage(palette theme.Theme, props HooksPageProps) HooksPage {
	styles := theme.NewStyles(palette)
	status := "✗ Hooks Not Installed"
	if props.HooksInstalled {
		status = "✓ Hooks Installed"
	}
	repoText := fmt.Sprintf("Path: %s\nStatus: %s", props.RepoPath, status)

	helpText := strings.TrimSpace(props.NavigationHelp)
	if extra := strings.TrimSpace(props.CommandHelpLine); extra != "" {
		if helpText != "" {
			helpText += "\n"
		}
		helpText += extra
	}

	return HooksPage{
		title:      components.NewTitleBar(props.Title, styles),
		repoInfo:   components.NewParagraph(repoText, styles),
		hookList:   components.NewHookList(props.Hooks, props.ActiveIcon, props.InactiveIcon, styles),
		benefits:   components.NewBulletList([]string{
			"Automatic profile suggestions based on branch/project",
			"Track repository events for better usage analytics",
			"Context-aware AI model selection",
			"Zero-overhead tracking (async logging)",
		}, styles),
		activity:   components.NewParagraph("No recent activity recorded", styles),
		help:       components.NewHelpBar(helpText, styles),
		repoTitle:  props.RepoTitle,
		hooksTitle: props.HooksTitle,
		benefitsT:  props.BenefitsTitle,
		activityT:  props.ActivityTitle,
	}
}

// Init satisfies the Page interface.
func (p HooksPage) Init() tea.Cmd {
	return nil
}

// Update satisfies the Page interface.
func (p HooksPage) Update(msg tea.Msg) (Page, tea.Cmd) {
	_, cmd1 := p.title.Update(msg)
	_, cmd2 := p.repoInfo.Update(msg)
	_, cmd3 := p.hookList.Update(msg)
	_, cmd4 := p.benefits.Update(msg)
	_, cmd5 := p.activity.Update(msg)
	_, cmd6 := p.help.Update(msg)
	return p, tea.Batch(cmd1, cmd2, cmd3, cmd4, cmd5, cmd6)
}

// View assembles the hooks management view.
func (p HooksPage) View() string {
	blocks := []string{
		layouts.Pad(2, p.title.View()),
		layouts.Gap(1),
		layouts.Section(p.repoTitle, p.repoInfo.View()),
		layouts.Gap(1),
		layouts.Section(p.hooksTitle, p.hookList.View()),
		layouts.Gap(1),
		layouts.Section(p.benefitsT, p.benefits.View()),
		layouts.Gap(1),
		layouts.Section(p.activityT, p.activity.View()),
		layouts.Gap(1),
		layouts.Pad(2, p.help.View()),
	}

	return layouts.Column(blocks...)
}
