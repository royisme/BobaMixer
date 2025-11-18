package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/royisme/bobamixer/internal/ui/theme"
)

// StatsSummary represents aggregate statistics for a time range.
type StatsSummary struct {
	Title          string
	Tokens         int
	Cost           float64
	Sessions       int
	AvgDailyTokens float64
	AvgDailyCost   float64
	ShowAverages   bool
}

// StatsSummaryPanel renders usage metrics in a simple block.
type StatsSummaryPanel struct {
	summary StatsSummary
	styles  theme.Styles
}

// NewStatsSummaryPanel builds a panel for the provided summary data.
func NewStatsSummaryPanel(summary StatsSummary, styles theme.Styles) StatsSummaryPanel {
	return StatsSummaryPanel{
		summary: summary,
		styles:  styles,
	}
}

// Update satisfies the Bubble Tea component interface.
func (c StatsSummaryPanel) Update(_ tea.Msg) (StatsSummaryPanel, tea.Cmd) {
	return c, nil
}

// View renders the formatted metrics list.
func (c StatsSummaryPanel) View() string {
	lines := []string{
		fmt.Sprintf("Tokens:   %d", c.summary.Tokens),
		fmt.Sprintf("Cost:     $%.4f", c.summary.Cost),
		fmt.Sprintf("Sessions: %d", c.summary.Sessions),
	}

	if c.summary.ShowAverages {
		lines = append(lines,
			fmt.Sprintf("Avg Daily Tokens: %.0f", c.summary.AvgDailyTokens),
			fmt.Sprintf("Avg Daily Cost:   $%.4f", c.summary.AvgDailyCost),
		)
	}

	var b strings.Builder
	style := c.styles.Normal
	style = style.PaddingLeft(2)
	for _, line := range lines {
		b.WriteString(style.Render(line))
		b.WriteString("\n")
	}

	return strings.TrimRight(b.String(), "\n")
}
