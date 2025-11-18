package pages

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/royisme/bobamixer/internal/ui/components"
	"github.com/royisme/bobamixer/internal/ui/layouts"
	"github.com/royisme/bobamixer/internal/ui/theme"
)

// SuggestionsPageProps describes the suggestions screen data.
type SuggestionsPageProps struct {
	Title           string
	SectionTitle    string
	DetailsTitle    string
	Suggestions     []components.Suggestion
	SelectedIndex   int
	Error           string
	NavigationHelp  string
	CommandHelpLine string
}

// SuggestionsPage composes the optimization suggestions UI.
type SuggestionsPage struct {
	title    components.TitleBar
	list     components.SuggestionList
	details  components.SuggestionDetails
	errorMsg components.StatusMessage
	help     components.HelpBar
	section  string
	detailsT string
}

// NewSuggestionsPage builds the suggestions page.
func NewSuggestionsPage(palette theme.Theme, props SuggestionsPageProps) SuggestionsPage {
	styles := theme.NewStyles(palette)
	helpText := strings.TrimSpace(props.NavigationHelp)
	if extra := strings.TrimSpace(props.CommandHelpLine); extra != "" {
		if helpText != "" {
			helpText += "\n"
		}
		helpText += extra
	}

	var selectedSuggestion *components.Suggestion
	if len(props.Suggestions) > 0 {
		index := props.SelectedIndex
		if index < 0 || index >= len(props.Suggestions) {
			index = 0
		}
		selectedSuggestion = &props.Suggestions[index]
	}

	return SuggestionsPage{
		title:    components.NewTitleBar(props.Title, styles),
		list:     components.NewSuggestionList(props.Suggestions, props.SelectedIndex, styles),
		details:  components.NewSuggestionDetails(selectedSuggestion, styles),
		errorMsg: components.NewStatusMessage(strings.TrimSpace(props.Error), palette.Danger),
		help:     components.NewHelpBar(helpText, styles),
		section:  props.SectionTitle,
		detailsT: props.DetailsTitle,
	}
}

// Init satisfies the Page interface.
func (p SuggestionsPage) Init() tea.Cmd {
	return nil
}

// Update satisfies the Page interface.
func (p SuggestionsPage) Update(msg tea.Msg) (Page, tea.Cmd) {
	_, cmd1 := p.title.Update(msg)
	_, cmd2 := p.list.Update(msg)
	_, cmd3 := p.details.Update(msg)
	_, cmd4 := p.errorMsg.Update(msg)
	_, cmd5 := p.help.Update(msg)
	return p, tea.Batch(cmd1, cmd2, cmd3, cmd4, cmd5)
}

// View assembles the suggestions view.
func (p SuggestionsPage) View() string {
	blocks := []string{
		layouts.Pad(2, p.title.View()),
		layouts.Gap(1),
	}

	if errView := strings.TrimSpace(p.errorMsg.View()); errView != "" {
		blocks = append(blocks, layouts.Pad(2, errView), layouts.Gap(1), layouts.Pad(2, p.help.View()))
		return layouts.Column(blocks...)
	}

	blocks = append(blocks, layouts.Section(p.section, p.list.View()))

	if details := p.details.View(); details != "" {
		blocks = append(blocks, layouts.Gap(1), layouts.Section(p.detailsT, details))
	}

	blocks = append(blocks, layouts.Gap(1), layouts.Pad(2, p.help.View()))
	return layouts.Column(blocks...)
}
