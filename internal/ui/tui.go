// Package ui provides the terminal user interface for BobaMixer.
package ui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/royisme/bobamixer/internal/domain/budget"
	"github.com/royisme/bobamixer/internal/domain/session"
	"github.com/royisme/bobamixer/internal/domain/stats"
	"github.com/royisme/bobamixer/internal/domain/suggestions"
	"github.com/royisme/bobamixer/internal/notifications"
	"github.com/royisme/bobamixer/internal/store/config"
	"github.com/royisme/bobamixer/internal/store/sqlite"
)

// ViewMode represents different views in the TUI
type ViewMode int

const (
	ViewDashboard ViewMode = iota
	ViewProfiles
	ViewBudget
	ViewTrends
	ViewSessions
)

// Model represents the TUI state
type Model struct {
	home          string
	activeProfile string
	profiles      config.Profiles
	profileList   []string
	selectedIdx   int
	viewMode      ViewMode
	width         int
	height        int
	db            *sqlite.DB
	budgetTracker *budget.Tracker
	statsAnalyzer *stats.Analyzer
	todayStats    *stats.DataPoint
	trend7d       *stats.Trend
	budgetStatus  *budget.Status
	lastUpdate    time.Time
	sessionList   []*session.Session
	notifier      *notifications.Notifier
	notifications []notifications.Event
	flashMessage  string
	err           error
}

// Colors and styles
var (
	primaryColor = lipgloss.Color("#7C3AED")
	successColor = lipgloss.Color("#10B981")
	warningColor = lipgloss.Color("#F59E0B")
	dangerColor  = lipgloss.Color("#EF4444")
	mutedColor   = lipgloss.Color("#6B7280")

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			MarginBottom(1)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#E5E7EB")).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(0, 1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			PaddingLeft(2)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9CA3AF")).
			PaddingLeft(2)

	budgetOKStyle = lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true)

	budgetWarningStyle = lipgloss.NewStyle().
				Foreground(warningColor).
				Bold(true)

	budgetDangerStyle = lipgloss.NewStyle().
				Foreground(dangerColor).
				Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Italic(true)
)

// Init initializes the model
func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{m.loadData, tea.EnterAltScreen}
	if m.notifier != nil {
		cmds = append(cmds, m.watchNotifications())
	}
	return tea.Batch(cmds...)
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "tab":
			// Cycle through views
			m.viewMode = ViewMode((int(m.viewMode) + 1) % 4)
			return m, m.loadData

		case "up", "k":
			if m.viewMode == ViewProfiles && m.selectedIdx > 0 {
				m.selectedIdx--
			}

		case "down", "j":
			if m.viewMode == ViewProfiles && m.selectedIdx < len(m.profileList)-1 {
				m.selectedIdx++
			}

		case "enter":
			if m.viewMode == ViewProfiles && m.selectedIdx < len(m.profileList) {
				m.activeProfile = m.profileList[m.selectedIdx]
				return m, m.saveActiveProfile
			}

		case "r":
			// Refresh data
			return m, m.loadData
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case dataLoadedMsg:
		m.todayStats = msg.todayStats
		m.trend7d = msg.trend7d
		m.budgetStatus = msg.budgetStatus
		m.lastUpdate = time.Now()
		m.err = msg.err
		m.sessionList = msg.sessions

	case notificationMsg:
		if msg.err != nil {
			m.err = msg.err
		} else if len(msg.events) > 0 {
			m.notifications = append(m.notifications, msg.events...)
			latest := msg.events[len(msg.events)-1]
			m.flashMessage = latest.Title
		}
		return m, m.watchNotifications()

	case errMsg:
		m.err = msg.err
	}

	return m, nil
}

// View renders the UI
func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var content string

	// Header
	header := m.renderHeader()

	// Main content based on view mode
	switch m.viewMode {
	case ViewDashboard:
		content = m.renderDashboard()
	case ViewProfiles:
		content = m.renderProfiles()
	case ViewBudget:
		content = m.renderBudget()
	case ViewTrends:
		content = m.renderTrends()
	case ViewSessions:
		content = m.renderSessions()
	}

	// Footer
	footer := m.renderFooter()

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		content,
		"",
		footer,
	)
}

func (m Model) renderHeader() string {
	title := titleStyle.Render("🧋 BobaMixer")

	var profileInfo string
	if m.activeProfile != "" {
		prof, ok := m.profiles[m.activeProfile]
		if ok {
			profileInfo = fmt.Sprintf("Active: %s (%s)", prof.Name, prof.Model)
		}
	} else {
		profileInfo = "No active profile"
	}

	// View mode indicator
	viewNames := []string{"Dashboard", "Profiles", "Budget", "Trends", "Sessions"}
	viewIndicator := fmt.Sprintf("[%s]", viewNames[m.viewMode])

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		title,
		"  ",
		normalStyle.Render(profileInfo),
		"  ",
		helpStyle.Render(viewIndicator),
	)
}

func (m Model) renderDashboard() string {
	var sections []string

	// Today's Stats
	if m.todayStats != nil {
		statsBox := m.renderStatsBox("Today's Usage", []string{
			fmt.Sprintf("Tokens: %s", stats.FormatTokens(m.todayStats.Tokens)),
			fmt.Sprintf("Cost: %s", stats.FormatCurrency(m.todayStats.Cost)),
			fmt.Sprintf("Sessions: %d", m.todayStats.Count),
		})
		sections = append(sections, statsBox)
	}

	// Budget Status
	if m.budgetStatus != nil {
		sections = append(sections, m.renderBudgetStatusBox())
	}

	// 7-day Trend
	if m.trend7d != nil && len(m.trend7d.DataPoints) > 0 {
		sparkline := stats.GetSparkline(m.trend7d.DataPoints)
		trendBox := m.renderStatsBox("7-Day Trend", []string{
			sparkline,
			fmt.Sprintf("Total: %s", stats.FormatCurrency(m.trend7d.Summary.TotalCost)),
			fmt.Sprintf("Avg/day: %s", stats.FormatCurrency(m.trend7d.Summary.AvgDailyCost)),
		})
		sections = append(sections, trendBox)
	}

	if len(sections) == 0 {
		return helpStyle.Render("No data available. Use 'r' to refresh.")
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m Model) renderProfiles() string {
	var lines []string
	lines = append(lines, headerStyle.Render("Available Profiles"))
	lines = append(lines, "")

	for i, profileName := range m.profileList {
		prof := m.profiles[profileName]
		line := fmt.Sprintf("%s - %s", profileName, prof.Model)

		if profileName == m.activeProfile {
			line += " ✓"
		}

		if i == m.selectedIdx {
			lines = append(lines, selectedStyle.Render("▶ "+line))
		} else {
			lines = append(lines, normalStyle.Render("  "+line))
		}
	}

	lines = append(lines, "")
	lines = append(lines, helpStyle.Render("↑/↓: Navigate  Enter: Select  Tab: Switch view"))

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m Model) renderBudget() string {
	if m.budgetStatus == nil {
		return helpStyle.Render("No budget configured. Use 'boba budget' to set up.")
	}

	var lines []string
	lines = append(lines, headerStyle.Render("Budget Status"))
	lines = append(lines, "")

	// Daily limit
	dailyPercent := m.budgetStatus.DailyProgress
	dailyBar := m.renderProgressBar(dailyPercent, 30)
	dailyStyle := budgetOKStyle
	if m.budgetStatus.IsOverDaily {
		dailyStyle = budgetDangerStyle
	} else if dailyPercent > 80 {
		dailyStyle = budgetWarningStyle
	}

	lines = append(lines, fmt.Sprintf("Daily Limit: %s / %s",
		dailyStyle.Render(stats.FormatCurrency(m.budgetStatus.CurrentSpent)),
		stats.FormatCurrency(m.budgetStatus.DailyLimit),
	))
	lines = append(lines, dailyBar)
	lines = append(lines, "")

	// Hard cap
	totalPercent := m.budgetStatus.TotalProgress
	totalBar := m.renderProgressBar(totalPercent, 30)
	totalStyle := budgetOKStyle
	if m.budgetStatus.IsOverCap {
		totalStyle = budgetDangerStyle
	} else if totalPercent > 80 {
		totalStyle = budgetWarningStyle
	}

	// Calculate total spent
	totalSpent := m.budgetStatus.Budget.SpentUSD
	lines = append(lines, fmt.Sprintf("Hard Cap: %s / %s",
		totalStyle.Render(stats.FormatCurrency(totalSpent)),
		stats.FormatCurrency(m.budgetStatus.HardCap),
	))
	lines = append(lines, totalBar)
	lines = append(lines, "")

	// Warning level
	warningLevel := m.budgetStatus.GetWarningLevel()
	var warningMsg string
	switch warningLevel {
	case "critical":
		warningMsg = budgetDangerStyle.Render("⚠ CRITICAL: Budget limit exceeded!")
	case "warning":
		warningMsg = budgetWarningStyle.Render("⚡ WARNING: Approaching budget limit")
	default:
		warningMsg = budgetOKStyle.Render("✓ Budget healthy")
	}
	lines = append(lines, warningMsg)

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m Model) renderTrends() string {
	if m.trend7d == nil {
		return helpStyle.Render("No trend data available.")
	}

	var lines []string
	lines = append(lines, headerStyle.Render("Usage Trends (7 Days)"))
	lines = append(lines, "")

	// Sparkline
	if len(m.trend7d.DataPoints) > 0 {
		sparkline := stats.GetSparkline(m.trend7d.DataPoints)
		lines = append(lines, fmt.Sprintf("Cost Trend: %s", sparkline))
		lines = append(lines, "")
	}

	// Summary
	lines = append(lines, fmt.Sprintf("Total Cost: %s", stats.FormatCurrency(m.trend7d.Summary.TotalCost)))
	lines = append(lines, fmt.Sprintf("Total Tokens: %s", stats.FormatTokens(m.trend7d.Summary.TotalTokens)))
	lines = append(lines, fmt.Sprintf("Total Sessions: %d", m.trend7d.Summary.TotalSessions))
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("Avg Daily Cost: %s", stats.FormatCurrency(m.trend7d.Summary.AvgDailyCost)))
	lines = append(lines, fmt.Sprintf("Avg Daily Tokens: %.0f", m.trend7d.Summary.AvgDailyTokens))
	lines = append(lines, "")

	// Trend direction
	trendDir := stats.DetectTrend(m.trend7d.DataPoints)
	var trendMsg string
	switch trendDir {
	case "increasing":
		trendMsg = warningColor.Render("📈 Increasing")
	case "decreasing":
		trendMsg = successColor.Render("📉 Decreasing")
	default:
		trendMsg = mutedColor.Render("➡ Stable")
	}
	lines = append(lines, fmt.Sprintf("Trend: %s", trendMsg))

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m Model) renderSessions() string {
	if len(m.sessionList) == 0 {
		return helpStyle.Render("No sessions recorded yet.")
	}

	lines := []string{headerStyle.Render("Recent Sessions"), ""}
	for _, sess := range m.sessionList {
		started := time.Unix(sess.StartedAt, 0).Format("01-02 15:04")
		status := successColor.Render("✓")
		if !sess.Success {
			status = dangerColor.Render("✗")
		}
		dur := fmt.Sprintf("%dms", sess.LatencyMS)
		lines = append(lines,
			fmt.Sprintf("%s  %-12s %-10s %-8s %s", started, sess.Profile, sess.Adapter, dur, status),
		)
	}
	lines = append(lines, "")
	lines = append(lines, helpStyle.Render("Tab: Switch view  r: Refresh"))
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m Model) renderStatsBox(title string, lines []string) string {
	content := []string{
		headerStyle.Render(title),
		"",
	}
	content = append(content, lines...)
	return lipgloss.JoinVertical(lipgloss.Left, content...)
}

func (m Model) renderBudgetStatusBox() string {
	if m.budgetStatus == nil {
		return ""
	}

	warningLevel := m.budgetStatus.GetWarningLevel()
	var statusStyle lipgloss.Style
	var statusIcon string

	switch warningLevel {
	case "critical":
		statusStyle = budgetDangerStyle
		statusIcon = "🔴"
	case "warning":
		statusStyle = budgetWarningStyle
		statusIcon = "🟡"
	default:
		statusStyle = budgetOKStyle
		statusIcon = "🟢"
	}

	dailyPercent := m.budgetStatus.DailyProgress
	bar := m.renderProgressBar(dailyPercent, 20)

	lines := []string{
		statusStyle.Render(fmt.Sprintf("%s Budget: %.0f%%", statusIcon, dailyPercent)),
		bar,
		fmt.Sprintf("%s / %s daily",
			stats.FormatCurrency(m.budgetStatus.CurrentSpent),
			stats.FormatCurrency(m.budgetStatus.DailyLimit),
		),
	}

	return m.renderStatsBox("Budget Status", lines)
}

func (m Model) renderProgressBar(percent float64, width int) string {
	if percent < 0 {
		percent = 0
	}
	filled := int(percent * float64(width) / 100)
	if filled > width {
		filled = width
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)

	var style lipgloss.Style
	if percent > 100 {
		style = budgetDangerStyle
	} else if percent > 80 {
		style = budgetWarningStyle
	} else {
		style = budgetOKStyle
	}

	return style.Render(bar) + fmt.Sprintf(" %.1f%%", percent)
}

func (m Model) renderFooter() string {
	var parts []string

	if m.err != nil {
		parts = append(parts, dangerColor.Render("Error: "+m.err.Error()))
	}

	if m.flashMessage != "" {
		parts = append(parts, successColor.Render(m.flashMessage))
	}

	if !m.lastUpdate.IsZero() {
		parts = append(parts, helpStyle.Render(
			fmt.Sprintf("Last updated: %s", m.lastUpdate.Format("15:04:05")),
		))
	}

	parts = append(parts, helpStyle.Render("Tab: Switch view • r: Refresh • q: Quit"))

	return strings.Join(parts, " | ")
}

// Messages
type dataLoadedMsg struct {
	todayStats   *stats.DataPoint
	trend7d      *stats.Trend
	budgetStatus *budget.Status
	sessions     []*session.Session
	err          error
}

type notificationMsg struct {
	events []notifications.Event
	err    error
}

type errMsg struct {
	err error
}

func (m Model) loadData() tea.Msg {
	var msg dataLoadedMsg

	if m.statsAnalyzer != nil {
		// Load today's stats
		todayStats, err := m.statsAnalyzer.GetTodayStats()
		if err == nil {
			msg.todayStats = todayStats
		} else {
			msg.err = err
		}

		// Load 7-day trend
		trend7d, err := m.statsAnalyzer.GetTrend(7)
		if err == nil {
			msg.trend7d = trend7d
		}
	}

	if m.budgetTracker != nil && m.activeProfile != "" {
		// Load budget status for active profile
		status, err := m.budgetTracker.GetStatus("profile", m.activeProfile)
		if err == nil {
			msg.budgetStatus = status
		}
	}

	if sessions, err := session.ListRecentSessions(m.db, 10); err == nil {
		msg.sessions = sessions
	}

	return msg
}

func (m Model) saveActiveProfile() tea.Msg {
	if err := config.SaveActiveProfile(m.home, m.activeProfile); err != nil {
		return errMsg{err: err}
	}
	return nil
}

// Run starts the TUI
func Run(home string) error {
	// Open database
	dbPath := filepath.Join(home, "usage.db")
	db, err := sqlite.Open(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Load profiles
	profiles, err := config.LoadProfiles(home)
	if err != nil {
		return fmt.Errorf("failed to load profiles: %w", err)
	}

	// Build profile list
	profileList := make([]string, 0, len(profiles))
	for name := range profiles {
		profileList = append(profileList, name)
	}

	// Get active profile
	activeProfile := getActiveProfile(home)
	selectedIdx := 0
	if activeProfile != "" {
		for i, name := range profileList {
			if name == activeProfile {
				selectedIdx = i
				break
			}
		}
	}

	// Initialize model
	tracker := budget.NewTracker(db)
	suggEngine := suggestions.NewEngine(db)

	m := Model{
		home:          home,
		activeProfile: activeProfile,
		profiles:      profiles,
		profileList:   profileList,
		selectedIdx:   selectedIdx,
		viewMode:      ViewDashboard,
		db:            db,
		budgetTracker: tracker,
		statsAnalyzer: stats.NewAnalyzer(db),
		notifier:      notifications.NewNotifier(tracker, suggEngine, nil),
	}

	// Run the program
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err = p.Run()
	return err
}

func getActiveProfile(home string) string {
	prof, err := config.LoadActiveProfile(home)
	if err != nil {
		return ""
	}
	return prof
}

func (m Model) watchNotifications() tea.Cmd {
	if m.notifier == nil {
		return nil
	}
	return tea.Tick(5*time.Second, func(time.Time) tea.Msg {
		events, err := m.notifier.Poll()
		return notificationMsg{events: events, err: err}
	})
}
