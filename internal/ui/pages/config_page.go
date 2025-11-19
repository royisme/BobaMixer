package pages

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/royisme/bobamixer/internal/ui/components"
	"github.com/royisme/bobamixer/internal/ui/layouts"
	"github.com/royisme/bobamixer/internal/ui/theme"
)

// ConfigPageProps holds the configuration editor data.
type ConfigPageProps struct {
	Title           string
	ConfigTitle     string
	EditorTitle     string
	SafetyTitle     string
	ThemeTitle      string
	ConfigFiles     []components.ConfigFile
	SelectedIndex   int
	Home            string
	EditorName      string
	NavigationHelp  string
	CommandHelpLine string
	Themes          []string
	CurrentTheme    string
	ActiveTab       int
}

// ConfigPage composes the configuration editor view.
type ConfigPage struct {
	title       components.TitleBar
	configList  components.ConfigFileList
	editorInfo  components.Paragraph
	safetyList  components.BulletList
	themeInfo   components.Paragraph
	help        components.HelpBar
	styles      theme.Styles
	configTitle string
	editorTitle string
	safetyTitle string
	themeTitle  string
	activeTab   int // 0: Files, 1: Appearance, 2: System
}

// NewConfigPage builds the config page.
func NewConfigPage(palette theme.Theme, props ConfigPageProps) ConfigPage {
	styles := theme.NewStyles(palette)
	editorText := "Editor: $EDITOR (" + props.EditorName + ")\nTip: Set $EDITOR to use your preferred editor"
	helpText := strings.TrimSpace(props.NavigationHelp)
	if extra := strings.TrimSpace(props.CommandHelpLine); extra != "" {
		if helpText != "" {
			helpText += "\n"
		}
		helpText += extra
	}

	var themeBuilder strings.Builder
	for _, t := range props.Themes {
		if t == props.CurrentTheme {
			themeBuilder.WriteString(styles.Selected.Render(" "+t) + "\n")
		} else {
			themeBuilder.WriteString(styles.Normal.Render(" "+t) + "\n")
		}
	}
	themeText := themeBuilder.String()

	return ConfigPage{
		title:      components.NewTitleBar(props.Title, styles),
		configList: components.NewConfigFileList(props.ConfigFiles, props.SelectedIndex, props.Home, styles),
		editorInfo: components.NewParagraph(editorText, styles),
		safetyList: components.NewBulletList([]string{
			"Automatic backup before editing",
			"YAML syntax validation after save",
			"Rollback support if validation fails",
		}, styles),
		themeInfo:   components.NewParagraph(themeText, styles),
		help:        components.NewHelpBar(helpText, styles),
		styles:      styles, // Initialize styles
		configTitle: props.ConfigTitle,
		editorTitle: props.EditorTitle,
		safetyTitle: props.SafetyTitle,
		themeTitle:  props.ThemeTitle,
		activeTab:   props.ActiveTab,
	}
}

// Init satisfies the Page interface.
func (p ConfigPage) Init() tea.Cmd {
	return nil
}

// Update satisfies the Page interface.
func (p ConfigPage) Update(msg tea.Msg) (Page, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h":
			if p.activeTab > 0 {
				p.activeTab--
			}
			return p, nil
		case "right", "l":
			if p.activeTab < 2 {
				p.activeTab++
			}
			return p, nil
		}
	}

	// Forward messages to active component
	switch p.activeTab {
	case 0: // Files
		_, cmd = p.configList.Update(msg)
	case 1: // Appearance
		// Theme selection is handled by parent (DashboardModel) via left/right
		// But since we use left/right for tabs now, we need a different way or
		// we need to let the parent handle theme switching when this tab is active.
		// For now, let's just update the info view.
		_, cmd = p.themeInfo.Update(msg)
	case 2: // System
		// These are currently static/informational
		_, cmd = p.editorInfo.Update(msg)
	}

	return p, cmd
}

// View assembles the configuration editor view.
func (p ConfigPage) View() string {
	// Tabs
	tabs := []string{"Files", "Appearance", "System"}
	var tabViews []string
	for i, t := range tabs {
		if i == p.activeTab {
			tabViews = append(tabViews, p.styles.ActiveTab.Render(t))
		} else {
			tabViews = append(tabViews, p.styles.Normal.Render(t))
		}
	}
	tabBar := lipgloss.JoinHorizontal(lipgloss.Top, tabViews...)
	tabBar = lipgloss.NewStyle().MarginBottom(1).Render(tabBar)

	// Content
	var content string
	switch p.activeTab {
	case 0: // Files
		content = components.NewCard(p.styles).
			WithWidth(60).
			Render(layouts.Column(
				p.styles.Header.Render(p.configTitle),
				p.configList.View(),
			))
	case 1: // Appearance
		content = components.NewCard(p.styles).
			WithWidth(60).
			Render(layouts.Column(
				p.styles.Header.Render(p.themeTitle),
				p.themeInfo.View(),
			))
	case 2: // System
		editorCard := components.NewCard(p.styles).
			WithWidth(60).
			Render(layouts.Column(
				p.styles.Header.Render(p.editorTitle),
				p.editorInfo.View(),
			))

		safetyCard := components.NewCard(p.styles).
			WithWidth(60).
			Render(layouts.Column(
				p.styles.Header.Render(p.safetyTitle),
				p.safetyList.View(),
			))

		content = layouts.Column(editorCard, layouts.Gap(1), safetyCard)
	}

	blocks := []string{
		layouts.Pad(2, p.title.View()),
		layouts.Gap(1),
		layouts.Pad(2, tabBar),
		layouts.Pad(2, content),
		layouts.Gap(1),
		layouts.Pad(2, p.help.View()),
	}

	return layouts.Column(blocks...)
}
