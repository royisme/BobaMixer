// Package ui provides the terminal user interface for BobaMixer.
package ui

import (
	"context"
	"fmt"
	"os"
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
	"github.com/royisme/bobamixer/internal/settings"
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

const viewCount = int(ViewSessions) + 1

// Model represents the TUI state
type Model struct {
	lastUpdate    time.Time
	profiles      config.Profiles
	profileList   []string
	sessionList   []*session.Session
	notifications []notifications.Event
	db            *sqlite.DB
	budgetTracker *budget.Tracker
	statsAnalyzer *stats.Analyzer
	todayStats    *stats.DataPoint
	trend7d       *stats.Trend
	budgetStatus  *budget.Status
	notifier      *notifications.Notifier
	theme         Theme
	localizer     *Localizer
	home          string
	activeProfile string
	flashMessage  string
	viewMode      ViewMode
	selectedIdx   int
	width         int
	height        int
	err           error
}

// Style helper methods using adaptive theme
func (m Model) titleStyle() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(m.theme.Primary).MarginBottom(1)
}

func (m Model) headerStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.Text).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.Border).
		Padding(0, 1)
}

func (m Model) selectedStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(m.theme.Primary).Bold(true).PaddingLeft(2)
}

func (m Model) normalStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(m.theme.Muted).PaddingLeft(2)
}

func (m Model) budgetOKStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(m.theme.Success).Bold(true)
}

func (m Model) budgetWarningStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(m.theme.Warning).Bold(true)
}

func (m Model) budgetDangerStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(m.theme.Danger).Bold(true)
}

func (m Model) helpStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(m.theme.Muted).Italic(true)
}

func (m Model) colorize(color lipgloss.AdaptiveColor, text string) string {
	return lipgloss.NewStyle().Foreground(color).Render(text)
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
        cmds := []tea.Cmd{m.loadData, tea.EnterAltScreen}
        if m.notifier != nil {
                cmds = append(cmds, m.watchNotifications())
        }
        return tea.Batch(cmds...)
}

// Update handles messages
//nolint:gocyclo // Complex TUI event handling with multiple message types and view modes
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
        switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "tab":
			// Cycle through views
			m.viewMode = ViewMode((int(m.viewMode) + 1) % viewCount)
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
	title := m.titleStyle().Render("🧋 BobaMixer")

	var profileInfo string
	if m.activeProfile != "" {
		prof, ok := m.profiles[m.activeProfile]
		if ok {
			profileInfo = m.localizer.TP("tui.active_profile", map[string]interface{}{
				"Name":  prof.Name,
				"Model": prof.Model,
			})
		}
	} else {
		profileInfo = m.localizer.T("tui.no_active_profile")
	}

	// View mode indicator
	viewNames := []string{
		m.localizer.T("tui.dashboard"),
		m.localizer.T("tui.profiles"),
		m.localizer.T("tui.budget"),
		m.localizer.T("tui.trends"),
		m.localizer.T("tui.sessions"),
	}
	viewIndicator := fmt.Sprintf("[%s]", viewNames[m.viewMode])

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		title,
		"  ",
		m.normalStyle().Render(profileInfo),
		"  ",
		m.helpStyle().Render(viewIndicator),
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
		return m.helpStyle().Render("No data available. Use 'r' to refresh.")
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m Model) renderProfiles() string {
	var lines []string
	lines = append(lines, m.headerStyle().Render("Available Profiles"))
	lines = append(lines, "")

	for i, profileName := range m.profileList {
		prof := m.profiles[profileName]
		line := fmt.Sprintf("%s - %s", profileName, prof.Model)

		if profileName == m.activeProfile {
			line += " ✓"
		}

		if i == m.selectedIdx {
			lines = append(lines, m.selectedStyle().Render("▶ "+line))
		} else {
			lines = append(lines, m.normalStyle().Render("  "+line))
		}
	}

	lines = append(lines, "")
	lines = append(lines, m.helpStyle().Render("↑/↓: Navigate  Enter: Select  Tab: Switch view"))

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m Model) renderBudget() string {
	if m.budgetStatus == nil {
		return m.helpStyle().Render("No budget configured. Use 'boba budget' to set up.")
	}

	var lines []string
	lines = append(lines, m.headerStyle().Render("Budget Status"))
	lines = append(lines, "")

	// Daily limit
	dailyPercent := m.budgetStatus.DailyProgress
	dailyBar := m.renderProgressBar(dailyPercent, 30)
	dailyStyle := m.budgetOKStyle()
	if m.budgetStatus.IsOverDaily {
		dailyStyle = m.budgetDangerStyle()
	} else if dailyPercent > 80 {
		dailyStyle = m.budgetWarningStyle()
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
	totalStyle := m.budgetOKStyle()
	if m.budgetStatus.IsOverCap {
		totalStyle = m.budgetDangerStyle()
	} else if totalPercent > 80 {
		totalStyle = m.budgetWarningStyle()
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
		warningMsg = m.budgetDangerStyle().Render("⚠ CRITICAL: Budget limit exceeded!")
	case "warning":
		warningMsg = m.budgetWarningStyle().Render("⚡ WARNING: Approaching budget limit")
	default:
		warningMsg = m.budgetOKStyle().Render("✓ Budget healthy")
	}
	lines = append(lines, warningMsg)

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m Model) renderTrends() string {
	if m.trend7d == nil {
		return m.helpStyle().Render("No trend data available.")
	}

	var lines []string
	lines = append(lines, m.headerStyle().Render("Usage Trends (7 Days)"))
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
		trendMsg = m.colorize(m.theme.Warning, "📈 Increasing")
	case "decreasing":
		trendMsg = m.colorize(m.theme.Success, "📉 Decreasing")
	default:
		trendMsg = m.colorize(m.theme.Muted, "➡ Stable")
	}
	lines = append(lines, fmt.Sprintf("Trend: %s", trendMsg))

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m Model) renderSessions() string {
	if len(m.sessionList) == 0 {
		return m.helpStyle().Render("No sessions recorded yet.")
	}

	lines := []string{m.headerStyle().Render("Recent Sessions"), ""}
	for _, sess := range m.sessionList {
		started := time.Unix(sess.StartedAt, 0).Format("01-02 15:04")
		status := m.colorize(m.theme.Success, "✓")
		if !sess.Success {
			status = m.colorize(m.theme.Danger, "✗")
		}
		dur := fmt.Sprintf("%dms", sess.LatencyMS)
		lines = append(lines,
			fmt.Sprintf("%s  %-12s %-10s %-8s %s", started, sess.Profile, sess.Adapter, dur, status),
		)
	}
	lines = append(lines, "")
	lines = append(lines, m.helpStyle().Render("Tab: Switch view  r: Refresh"))
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m Model) renderStatsBox(title string, lines []string) string {
	content := []string{
		m.headerStyle().Render(title),
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
		statusStyle = m.budgetDangerStyle()
		statusIcon = "🔴"
	case "warning":
		statusStyle = m.budgetWarningStyle()
		statusIcon = "🟡"
	default:
		statusStyle = m.budgetOKStyle()
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
		style = m.budgetDangerStyle()
	} else if percent > 80 {
		style = m.budgetWarningStyle()
	} else {
		style = m.budgetOKStyle()
	}

	return style.Render(bar) + fmt.Sprintf(" %.1f%%", percent)
}

func (m Model) renderFooter() string {
	var parts []string

	if m.err != nil {
		parts = append(parts, m.colorize(m.theme.Danger, "Error: "+m.err.Error()))
	}

	if m.flashMessage != "" {
		parts = append(parts, m.colorize(m.theme.Success, m.flashMessage))
	}

	if !m.lastUpdate.IsZero() {
		parts = append(parts, m.helpStyle().Render(
			fmt.Sprintf("Last updated: %s", m.lastUpdate.Format("15:04:05")),
		))
	}

	parts = append(parts, m.helpStyle().Render(m.localizer.T("tui.quit")))

	return strings.Join(parts, " | ")
}

// Messages
type dataLoadedMsg struct {
	sessions     []*session.Session
	todayStats   *stats.DataPoint
	trend7d      *stats.Trend
	budgetStatus *budget.Status
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
//nolint:gocyclo // Entry point handles multiple modes and fallback logic
func Run(home string) error {
	// Check if we should use new control plane or legacy profile system
	// New system uses tools.yaml and bindings.yaml
	toolsPath := filepath.Join(home, "tools.yaml")
	bindingsPath := filepath.Join(home, "bindings.yaml")

	useControlPlane := false
	if _, err := os.Stat(toolsPath); err == nil {
		if _, err := os.Stat(bindingsPath); err == nil {
			useControlPlane = true
		}
	}

	if useControlPlane {
		// Use new control plane dashboard
		return RunDashboard(home)
	}

	// Check if first-run (no configuration at all)
	providersPath := filepath.Join(home, "providers.yaml")
	if _, err := os.Stat(providersPath); os.IsNotExist(err) {
		// First-run: launch interactive onboarding
		shouldContinue, onboardErr := RunOnboarding(home)
		if onboardErr != nil {
			return fmt.Errorf("onboarding failed: %w", onboardErr)
		}
		if !shouldContinue {
			// User canceled onboarding
			return nil
		}

		// Onboarding completed, launch dashboard
		return RunDashboard(home)
	}

	// Legacy: Load profiles (gracefully handle missing/invalid config)
	profiles, err := config.LoadProfiles(home)
	if err != nil || len(profiles) == 0 {
		// Try onboarding for legacy users
		shouldContinue, wizardErr := RunOnboarding(home)
		if wizardErr != nil {
			return fmt.Errorf("setup wizard failed: %w", wizardErr)
		}
		if !shouldContinue {
			// User canceled wizard
			return nil
		}

		// Launch dashboard after onboarding
		return RunDashboard(home)
	}

	// Open database
	dbPath := filepath.Join(home, "usage.db")
	db, err := sqlite.Open(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
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

	// Initialize theme and i18n
	theme := loadTheme(home)
	localizer, err := NewLocalizer(GetUserLanguage())
	if err != nil {
		// Fallback to English - this should never fail with embedded locales
		var fallbackErr error
		localizer, fallbackErr = NewLocalizer("en")
		if fallbackErr != nil {
			return fmt.Errorf("failed to initialize localizer: %w (fallback also failed: %w)", err, fallbackErr)
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
		theme:         theme,
		localizer:     localizer,
	}

	// Run the program
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err = p.Run()
	return err
}

func loadTheme(home string) Theme {
	ctx := context.Background()

	userSettings, err := settings.Load(ctx, home)
	if err != nil {
		return GetTheme(settings.DefaultSettings().Theme)
	}

	themeName := userSettings.Theme
	if themeName == "" {
		themeName = settings.DefaultSettings().Theme
	}

	return GetTheme(themeName)
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
