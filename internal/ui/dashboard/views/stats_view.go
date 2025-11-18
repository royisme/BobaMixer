package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// StatsSummary represents aggregate statistics for a period.
type StatsSummary struct {
	Title           string
	Tokens          int
	Cost            float64
	Sessions        int
	AvgDailyTokens  float64
	AvgDailyCost    float64
	ShowAverages    bool
	DisplayCurrency bool
}

// StatsProfile represents a per-profile stats entry.
type StatsProfile struct {
	Name        string
	Tokens      int
	Cost        float64
	Sessions    int
	AvgLatency  float64
	UsagePct    float64
	CostPct     float64
}

// StatsViewProps carries all data required for the stats screen.
type StatsViewProps struct {
	Theme           ThemePalette
	Loaded          bool
	Error           string
	LoadingMessage  string
	Today           StatsSummary
	Week            StatsSummary
	Profiles        []StatsProfile
	NavigationHelp  string
	LoadingHelp     string
	ProfileSubtitle string
}

// RenderStatsView renders the usage statistics page.
func RenderStatsView(props StatsViewProps) string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(props.Theme.Primary).Padding(0, 2)
	sectionStyle := lipgloss.NewStyle().Bold(true).Foreground(props.Theme.Success).Padding(1, 2)
	dataStyle := lipgloss.NewStyle().Foreground(props.Theme.Text).Padding(0, 2)
	helpStyle := lipgloss.NewStyle().Foreground(props.Theme.Muted).Padding(1, 2)
	errorStyle := lipgloss.NewStyle().Foreground(props.Theme.Danger).Padding(0, 2)

	var content strings.Builder
	content.WriteString(titleStyle.Render("BobaMixer - Usage Statistics"))
	content.WriteString("\n\n")

	if !props.Loaded {
		if strings.TrimSpace(props.Error) != "" {
			content.WriteString(errorStyle.Render(fmt.Sprintf("Error loading stats: %s", props.Error)))
		} else {
			content.WriteString(dataStyle.Render(props.LoadingMessage))
		}
		content.WriteString("\n\n")
		content.WriteString(helpStyle.Render(props.LoadingHelp))
		return content.String()
	}

	renderSummary(&content, sectionStyle, dataStyle, props.Today)
	content.WriteString("\n")
	renderSummary(&content, sectionStyle, dataStyle, props.Week)

	if len(props.Profiles) > 0 {
		content.WriteString("\n")
		title := props.ProfileSubtitle
		if strings.TrimSpace(title) == "" {
			title = "🎯 By Profile (7d)"
		}
		content.WriteString(sectionStyle.Render(title))
		content.WriteString("\n")
		for _, ps := range props.Profiles {
			line := fmt.Sprintf("  • %s: tokens=%d cost=$%.4f sessions=%d latency=%.0fms usage=%.1f%% cost=%.1f%%",
				ps.Name,
				ps.Tokens,
				ps.Cost,
				ps.Sessions,
				ps.AvgLatency,
				ps.UsagePct,
				ps.CostPct,
			)
			content.WriteString(dataStyle.Render(line))
			content.WriteString("\n")
		}
		content.WriteString("\n")
	}

	content.WriteString(helpStyle.Render(props.NavigationHelp))
	return content.String()
}

func renderSummary(content *strings.Builder, sectionStyle, dataStyle lipgloss.Style, summary StatsSummary) {
	if strings.TrimSpace(summary.Title) != "" {
		content.WriteString(sectionStyle.Render(summary.Title))
		content.WriteString("\n")
	}

	content.WriteString(dataStyle.Render(fmt.Sprintf("  Tokens:   %d", summary.Tokens)))
	content.WriteString("\n")
	content.WriteString(dataStyle.Render(fmt.Sprintf("  Cost:     $%.4f", summary.Cost)))
	content.WriteString("\n")
	content.WriteString(dataStyle.Render(fmt.Sprintf("  Sessions: %d", summary.Sessions)))
	content.WriteString("\n")

	if summary.ShowAverages {
		content.WriteString(dataStyle.Render(fmt.Sprintf("  Avg Daily Tokens: %.0f", summary.AvgDailyTokens)))
		content.WriteString("\n")
		content.WriteString(dataStyle.Render(fmt.Sprintf("  Avg Daily Cost:   $%.4f", summary.AvgDailyCost)))
		content.WriteString("\n")
	}
}
