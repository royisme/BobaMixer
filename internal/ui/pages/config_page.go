package pages

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
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
	ConfigFiles     []components.ConfigFile
	SelectedIndex   int
	Home            string
	EditorName      string
	NavigationHelp  string
	CommandHelpLine string
}

// ConfigPage composes the configuration editor view.
type ConfigPage struct {
	title       components.TitleBar
	configList  components.ConfigFileList
	editorInfo  components.Paragraph
	safetyList  components.BulletList
	help        components.HelpBar
	configTitle string
	editorTitle string
	safetyTitle string
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

	return ConfigPage{
		title:       components.NewTitleBar(props.Title, styles),
		configList:  components.NewConfigFileList(props.ConfigFiles, props.SelectedIndex, props.Home, styles),
		editorInfo:  components.NewParagraph(editorText, styles),
		safetyList:  components.NewBulletList([]string{
			"Automatic backup before editing",
			"YAML syntax validation after save",
			"Rollback support if validation fails",
		}, styles),
		help:        components.NewHelpBar(helpText, styles),
		configTitle: props.ConfigTitle,
		editorTitle: props.EditorTitle,
		safetyTitle: props.SafetyTitle,
	}
}

// Init satisfies the Page interface.
func (p ConfigPage) Init() tea.Cmd {
	return nil
}

// Update satisfies the Page interface.
func (p ConfigPage) Update(msg tea.Msg) (Page, tea.Cmd) {
	_, cmd1 := p.title.Update(msg)
	_, cmd2 := p.configList.Update(msg)
	_, cmd3 := p.editorInfo.Update(msg)
	_, cmd4 := p.safetyList.Update(msg)
	_, cmd5 := p.help.Update(msg)
	return p, tea.Batch(cmd1, cmd2, cmd3, cmd4, cmd5)
}

// View assembles the configuration editor view.
func (p ConfigPage) View() string {
	blocks := []string{
		layouts.Pad(2, p.title.View()),
		layouts.Gap(1),
		layouts.Section(p.configTitle, p.configList.View()),
		layouts.Gap(1),
		layouts.Section(p.editorTitle, p.editorInfo.View()),
		layouts.Gap(1),
		layouts.Section(p.safetyTitle, p.safetyList.View()),
		layouts.Gap(1),
		layouts.Pad(2, p.help.View()),
	}

	return layouts.Column(blocks...)
}
