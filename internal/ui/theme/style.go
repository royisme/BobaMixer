package theme

import "github.com/charmbracelet/lipgloss"

// Styles bundles the lipgloss styles derived from a theme palette.
type Styles struct {
	Title        lipgloss.Style
	Header       lipgloss.Style
	Selected     lipgloss.Style
	Normal       lipgloss.Style
	BudgetOK     lipgloss.Style
	BudgetWarn   lipgloss.Style
	BudgetDanger lipgloss.Style
	Help         lipgloss.Style

	// Component Styles
	Card      lipgloss.Style
	StatusBar lipgloss.Style
	Tab       lipgloss.Style
	ActiveTab lipgloss.Style
	Input     lipgloss.Style
	List      lipgloss.Style
}

// NewStyles builds the default style set for the provided theme palette.
func NewStyles(palette Theme) Styles {
	return Styles{
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(palette.Primary).
			MarginBottom(1),
		Header: lipgloss.NewStyle().
			Bold(true).
			Foreground(palette.Text).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(palette.Border).
			Padding(0, 1),
		Selected: lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#FFFFFF"}).
			Background(palette.Primary).
			Bold(true).
			Padding(0, 1).
			MarginLeft(1),
		Normal: lipgloss.NewStyle().
			Foreground(palette.Muted).
			PaddingLeft(2),
		BudgetOK: lipgloss.NewStyle().
			Foreground(palette.Success).
			Bold(true),
		BudgetWarn: lipgloss.NewStyle().
			Foreground(palette.Warning).
			Bold(true),
		BudgetDanger: lipgloss.NewStyle().
			Foreground(palette.Danger).
			Bold(true),
		Help: lipgloss.NewStyle().
			Foreground(palette.Muted).
			Italic(true),

		// Component Styles
		Card: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(palette.Border).
			Padding(1, 2).
			Margin(1),
		StatusBar: lipgloss.NewStyle().
			Foreground(palette.Text).
			Background(palette.Border). // Using Border color as subtle background
			Padding(0, 1),
		Tab: lipgloss.NewStyle().
			Foreground(palette.Muted).
			Padding(0, 1),
		ActiveTab: lipgloss.NewStyle().
			Foreground(palette.Primary).
			Bold(true).
			BorderBottom(true).
			BorderStyle(lipgloss.ThickBorder()).
			BorderForeground(palette.Primary).
			Padding(0, 1),
		Input: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(palette.Primary).
			Padding(0, 1),
		List: lipgloss.NewStyle().
			Margin(1),
	}
}

// Colorize renders the text with the supplied adaptive color.
func Colorize(color lipgloss.AdaptiveColor, text string) string {
	return lipgloss.NewStyle().Foreground(color).Render(text)
}
