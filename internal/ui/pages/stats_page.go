package pages

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/royisme/bobamixer/internal/ui/components"
	"github.com/royisme/bobamixer/internal/ui/layouts"
	"github.com/royisme/bobamixer/internal/ui/theme"
)

// StatsPageProps holds the information required to render the usage stats screen.
type StatsPageProps struct {
	Title           string
	Loaded          bool
	Error           string
	LoadingMessage  string
	Today           components.StatsSummary
	Week            components.StatsSummary
	Profiles        []components.StatsProfile
	NavigationHelp  string
	LoadingHelp     string
	ProfileSubtitle string
}

// StatsPage composes the usage statistics UI.
type StatsPage struct {
	title      components.TitleBar
	today      components.StatsSummaryPanel
	week       components.StatsSummaryPanel
	profiles   components.StatsProfilesList
	errorMsg   components.StatusMessage
	loadingMsg components.InfoMessage
	help       components.HelpBar
	profileSub string
	loaded     bool
	todayTitle string
	weekTitle  string
}

// NewStatsPage builds a StatsPage using the provided palette and props.
func NewStatsPage(palette theme.Theme, props StatsPageProps) StatsPage {
	styles := theme.NewStyles(palette)
	errorText := ""
	if !props.Loaded {
		errorText = props.Error
	}

	return StatsPage{
		title:      components.NewTitleBar(props.Title, styles),
		today:      components.NewStatsSummaryPanel(props.Today, styles),
		week:       components.NewStatsSummaryPanel(props.Week, styles),
		profiles:   components.NewStatsProfilesList(props.Profiles, styles),
		errorMsg:   components.NewStatusMessage(strings.TrimSpace(errorText), palette.Danger),
		loadingMsg: components.NewInfoMessage(props.LoadingMessage, styles),
		help:       components.NewHelpBar(props.NavigationHelp, styles),
		profileSub: props.ProfileSubtitle,
		loaded:     props.Loaded,
		todayTitle: props.Today.Title,
		weekTitle:  props.Week.Title,
	}
}

// Init satisfies the Page interface.
func (p StatsPage) Init() tea.Cmd {
	return nil
}

// Update satisfies the Page interface.
func (p StatsPage) Update(msg tea.Msg) (Page, tea.Cmd) {
	_, cmd1 := p.title.Update(msg)
	_, cmd2 := p.today.Update(msg)
	_, cmd3 := p.week.Update(msg)
	_, cmd4 := p.profiles.Update(msg)
	_, cmd5 := p.errorMsg.Update(msg)
	_, cmd6 := p.loadingMsg.Update(msg)
	_, cmd7 := p.help.Update(msg)
	return p, tea.Batch(cmd1, cmd2, cmd3, cmd4, cmd5, cmd6, cmd7)
}

// View assembles the stats screen with the layout DSL.
func (p StatsPage) View() string {
	blocks := []string{
		layouts.Pad(2, p.title.View()),
		layouts.Gap(1),
	}

	if !p.loaded {
		if errView := strings.TrimSpace(p.errorMsg.View()); errView != "" {
			blocks = append(blocks, layouts.Pad(2, errView), layouts.Gap(1))
		} else {
			blocks = append(blocks, layouts.Pad(2, p.loadingMsg.View()), layouts.Gap(1))
		}
		blocks = append(blocks, layouts.Pad(2, p.help.View()))
		return layouts.Column(blocks...)
	}

	if today := p.today.View(); today != "" {
		blocks = append(blocks, layouts.Section(strings.TrimSpace(p.todayTitle), today), layouts.Gap(1))
	}
	if week := p.week.View(); week != "" {
		blocks = append(blocks, layouts.Section(strings.TrimSpace(p.weekTitle), week), layouts.Gap(1))
	}

	if profiles := p.profiles.View(); profiles != "" {
		title := strings.TrimSpace(p.profileSub)
		if title == "" {
			title = "🎯 By Profile (7d)"
		}
		blocks = append(blocks, layouts.Section(title, profiles), layouts.Gap(1))
	}

	blocks = append(blocks, layouts.Pad(2, p.help.View()))
	return layouts.Column(blocks...)
}
