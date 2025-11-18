package pages

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/royisme/bobamixer/internal/ui/components"
	"github.com/royisme/bobamixer/internal/ui/layouts"
	"github.com/royisme/bobamixer/internal/ui/theme"
)

// HelpPageProps represents the required data to render the help overlay.
type HelpPageProps struct {
	Title          string
	Subtitle       string
	Sections       []components.HelpSection
	Shortcuts      []components.Shortcut
	Tips           []string
	Links          []components.HelpLink
	NavigationHint string
}

// HelpPage composes the help overlay via the new components + layout architecture.
type HelpPage struct {
	header    components.HelpHeader
	sections  components.HelpSectionList
	shortcuts components.ShortcutList
	tips      components.HelpTips
	links     components.HelpLinks
	footer    components.HelpFooter
}

// NewHelpPage constructs a HelpPage with the shared palette.
func NewHelpPage(palette theme.Theme, props HelpPageProps) HelpPage {
	styles := theme.NewStyles(palette)

	shortcuts := props.Shortcuts
	if len(shortcuts) == 0 {
		shortcuts = defaultHelpShortcuts()
	}

	return HelpPage{
		header:    components.NewHelpHeader(props.Title, props.Subtitle, styles),
		sections:  components.NewHelpSectionList(props.Sections, styles),
		shortcuts: components.NewShortcutList(shortcuts, styles),
		tips:      components.NewHelpTips(props.Tips, styles),
		links:     components.NewHelpLinks(props.Links, styles),
		footer:    components.NewHelpFooter(props.NavigationHint, styles),
	}
}

// Init satisfies the Page interface; the help page has no startup commands.
func (p HelpPage) Init() tea.Cmd {
	return nil
}

// Update keeps the page immutable because its content is static.
func (p HelpPage) Update(msg tea.Msg) (Page, tea.Cmd) {
	_, cmd := p.header.Update(msg)
	_, cmd2 := p.sections.Update(msg)
	_, cmd3 := p.shortcuts.Update(msg)
	_, cmd4 := p.tips.Update(msg)
	_, cmd5 := p.links.Update(msg)
	_, cmd6 := p.footer.Update(msg)
	return p, tea.Batch(cmd, cmd2, cmd3, cmd4, cmd5, cmd6)
}

// View composes the help overlay using the layout DSL.
func (p HelpPage) View() string {
	return layouts.Column(
		p.header.View(),
		layouts.Gap(1),
		layouts.Section("Section Navigation", p.sections.View()),
		layouts.Gap(1),
		layouts.Section("Global Shortcuts", p.shortcuts.View()),
		layouts.Gap(1),
		layouts.Section("Quick Tips", p.tips.View()),
		layouts.Gap(1),
		layouts.Section("Documentation", p.links.View()),
		layouts.Gap(1),
		p.footer.View(),
	)
}

func defaultHelpShortcuts() []components.Shortcut {
	return []components.Shortcut{
		{Key: "Tab / Shift+Tab", Description: "Cycle sections"},
		{Key: "[ / ]", Description: "Cycle views within a section"},
		{Key: "↑/↓ or k/j", Description: "Navigate in lists"},
		{Key: "/", Description: "Search within supported lists"},
		{Key: "Esc", Description: "Clear search / close dialogs"},
		{Key: "R", Description: "Run selected tool (Dashboard view)"},
		{Key: "X", Description: "Toggle proxy (Dashboard view)"},
		{Key: "S", Description: "Refresh proxy status (Proxy view)"},
		{Key: "Q or Ctrl+C", Description: "Quit BobaMixer"},
	}
}
