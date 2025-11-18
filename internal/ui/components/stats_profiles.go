package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/royisme/bobamixer/internal/ui/theme"
)

// StatsProfile represents per-profile usage metrics.
type StatsProfile struct {
	Name       string
	Tokens     int
	Cost       float64
	Sessions   int
	AvgLatency float64
	UsagePct   float64
	CostPct    float64
}

// StatsProfilesList renders the list of profile stats.
type StatsProfilesList struct {
	profiles []StatsProfile
	styles   theme.Styles
}

// NewStatsProfilesList constructs the list component.
func NewStatsProfilesList(profiles []StatsProfile, styles theme.Styles) StatsProfilesList {
	return StatsProfilesList{
		profiles: profiles,
		styles:   styles,
	}
}

// Update satisfies the Bubble Tea component interface.
func (c StatsProfilesList) Update(_ tea.Msg) (StatsProfilesList, tea.Cmd) {
	return c, nil
}

// View renders the formatted profile rows.
func (c StatsProfilesList) View() string {
	if len(c.profiles) == 0 {
		return ""
	}

	var b strings.Builder
	style := c.styles.Normal
	style = style.PaddingLeft(2)

	for _, ps := range c.profiles {
		line := fmt.Sprintf("• %s: tokens=%d cost=$%.4f sessions=%d latency=%.0fms usage=%.1f%% cost=%.1f%%",
			ps.Name,
			ps.Tokens,
			ps.Cost,
			ps.Sessions,
			ps.AvgLatency,
			ps.UsagePct,
			ps.CostPct,
		)
		b.WriteString(style.Render(line))
		b.WriteString("\n")
	}

	return strings.TrimRight(b.String(), "\n")
}
