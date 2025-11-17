// Package cli provides the command-line interface for BobaMixer.
package cli

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/royisme/bobamixer/internal/adapters"
	"github.com/royisme/bobamixer/internal/domain/budget"
	"github.com/royisme/bobamixer/internal/domain/core"
	"github.com/royisme/bobamixer/internal/domain/hooks"
	"github.com/royisme/bobamixer/internal/domain/routing"
	"github.com/royisme/bobamixer/internal/domain/stats"
	"github.com/royisme/bobamixer/internal/domain/suggestions"
	"github.com/royisme/bobamixer/internal/logging"
	"github.com/royisme/bobamixer/internal/settings"
	"github.com/royisme/bobamixer/internal/store/config"
	"github.com/royisme/bobamixer/internal/store/sqlite"
	"github.com/royisme/bobamixer/internal/svc"
	"github.com/royisme/bobamixer/internal/ui"
	"github.com/royisme/bobamixer/internal/version"
	"go.uber.org/zap"
)

const (
	scopeGlobal  = "global"
	scopeProject = "project"

	// Status symbols for output
	statusOK      = "[OK]"
	statusError   = "[ERROR]"
	statusWarning = "[WARN]"

	shellBash = "bash"
	shellZsh  = "zsh"
	shellFish = "fish"

	useSlowThreshold   = 2 * time.Second
	statsSlowThreshold = 3 * time.Second
)

var supportedShells = []string{shellBash, shellZsh, shellFish}

// Run executes the BobaMixer CLI with the given arguments and routes to appropriate subcommands.
//
//nolint:gocyclo // Complex CLI entry point with multiple subcommands
func Run(args []string) error {
	home, err := config.ResolveHome()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(home, 0o700); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(home, "logs"), 0o700); err != nil {
		return err
	}
	if err := settings.InitHome(home); err != nil {
		return err
	}

	// Initialize structured logging
	if err := logging.Init(home); err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer func() {
		if err := logging.Sync(); err != nil {
			fmt.Fprintf(os.Stderr, "failed to flush logs: %v\n", err)
		}
	}()

	logging.Info("BobaMixer CLI started")

	// Enforce secrets permissions early so every command respects the baseline security requirement
	if err := config.ValidateSecretsPermissions(home); err != nil {
		return err
	}

	// Handle help flag
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h" || args[0] == "help") {
		printUsage()
		return nil
	}

	// No arguments: launch TUI dashboard
	if len(args) == 0 {
		logging.Info("Launching TUI dashboard")
		return runTUI(home)
	}

	switch args[0] {
	// Control Plane Commands (Phase 1)
	case "providers":
		return runProviders(home, args[1:])
	case "tools":
		return runTools(home, args[1:])
	case "bind":
		return runBind(home, args[1:])
	case "run":
		return runRun(home, args[1:])
	case "secrets":
		return runSecrets(home, args[1:])
	case "proxy":
		return runProxy(home, args[1:])

	// Legacy Profile Commands
	case "ls":
		return runLS(home, args[1:])
	case "use":
		return runUse(home, args[1:])
	case "call":
		return runCall(home, args[1:])
	case "stats":
		return runStats(home, args[1:])
	case "edit":
		return runEdit(home, args[1:])
	case "doctor":
		return runDoctorV2(home, args[1:])
	case "budget":
		return runBudget(home, args[1:])
	case "hooks":
		return runHooks(home, args[1:])
	case "action":
		return runAction(home, args[1:])
	case "report":
		return runReport(home, args[1:])
	case "init":
		return runInit(home, args[1:])
	case "route":
		return runRoute(home, args[1:])
	case "completions":
		return runCompletions(args[1:])
	case "suggest":
		return runSuggest(args[1:])
	case "version":
		return runVersion()
	default:
		return fmt.Errorf("unknown command %s", args[0])
	}
}

func printUsage() {
	fmt.Println("BobaMixer - AI CLI Control Plane")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  boba                                          Launch TUI dashboard")
	fmt.Println("  boba --help                                   Show this help")
	fmt.Println()
	fmt.Println("Control Plane (Phase 1 & 2):")
	fmt.Println("  boba providers                                List AI providers")
	fmt.Println("  boba tools                                    List CLI tools")
	fmt.Println("  boba secrets [list|set|remove]                Manage API keys (no YAML editing!)")
	fmt.Println("  boba bind <tool> <provider> [--proxy=on|off]  Bind tool to provider")
	fmt.Println("  boba run <tool> [args...]                     Run CLI tool with injected config")
	fmt.Println("  boba proxy serve                              Start local proxy server")
	fmt.Println("  boba proxy status                             Check proxy server status")
	fmt.Println("  boba doctor                                   Run diagnostics")
	fmt.Println()
	fmt.Println("Profile Management (Legacy):")
	fmt.Println("  boba ls --profiles                            List available profiles")
	fmt.Println("  boba use <profile>                            Activate a profile")
	fmt.Println()
	fmt.Println("AI Calls:")
	fmt.Println("  boba call --profile <p> --data @file.json    Execute an AI call")
	fmt.Println()
	fmt.Println("Usage & Statistics:")
	fmt.Println("  boba stats [--today|--7d|--30d] [--by-profile]  Show usage statistics")
	fmt.Println("  boba report [--format json|csv] [--out file]   Generate usage report")
	fmt.Println()
	fmt.Println("Configuration:")
	fmt.Println("  boba init                                     Initialize ~/.boba with defaults")
	fmt.Println("  boba edit <profiles|routes|pricing|secrets>  Edit configuration files")
	fmt.Println()
	fmt.Println("Budget & Optimization:")
	fmt.Println("  boba budget [--status]                        Show budget status")
	fmt.Println("  boba action [--auto]                          View/apply suggestions")
	fmt.Println()
	fmt.Println("Routing:")
	fmt.Println("  boba route test <text|@file>                 Test routing rules")
	fmt.Println()
	fmt.Println("Advanced:")
	fmt.Println("  boba hooks install|remove|track              Manage git hooks")
	fmt.Println("  boba version                                  Show version info")
	fmt.Println()
	fmt.Println("For more information, visit: https://royisme.github.io/BobaMixer/")
}

func runLS(home string, args []string) error {
	flags := flag.NewFlagSet("ls", flag.ContinueOnError)
	showProfiles := flags.Bool("profiles", false, "list profiles")
	flags.SetOutput(io.Discard)
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *showProfiles {
		profs, err := config.LoadProfiles(home)
		if err != nil {
			return err
		}
		if len(profs) == 0 {
			fmt.Println("no profiles defined")
			return nil
		}
		keys := make([]string, 0, len(profs))
		for k := range profs {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, key := range keys {
			prof := profs[key]
			fmt.Printf("- %s (%s -> %s)\n", prof.Key, prof.Adapter, prof.Model)
		}
		return nil
	}
	return errors.New("ls: specify --profiles")
}

func runUse(home string, args []string) error {
	start := time.Now()
	var profileKey string
	defer func() {
		logCommandDuration("use", start, useSlowThreshold, logging.String("profile", profileKey))
	}()

	if len(args) != 1 {
		return errors.New("use requires profile name")
	}
	profs, err := config.LoadProfiles(home)
	if err != nil {
		return err
	}
	prof, ok := profs[args[0]]
	if !ok {
		return fmt.Errorf("profile %s not found", args[0])
	}
	profileKey = prof.Key
	if err := config.SaveActiveProfile(home, prof.Key); err != nil {
		return err
	}
	fmt.Printf("active profile set to %s (%s)\n", prof.Key, prof.Model)
	return nil
}

//nolint:unparam // Keeping error return for consistency with other command handlers
func runVersion() error {
	v := version.GetVersionInfo()
	fmt.Println(v.String())
	return nil
}

func runStats(home string, args []string) error {
	flags := flag.NewFlagSet("stats", flag.ContinueOnError)
	today := flags.Bool("today", false, "show today's totals")
	days7 := flags.Bool("7d", false, "show last 7 days")
	days30 := flags.Bool("30d", false, "show last 30 days")
	byProfile := flags.Bool("by-profile", false, "breakdown by profile")
	flags.SetOutput(io.Discard)
	if err := flags.Parse(args); err != nil {
		return err
	}

	ctx := context.Background()
	dbPath := filepath.Join(home, "usage.db")
	db, err := sqlite.Open(dbPath)
	if err != nil {
		return err
	}

	if *today {
		summary, err := stats.Today(ctx, db)
		if err != nil {
			return err
		}
		printTodaySummary(summary)
		return nil
	}

	if *days7 {
		start := time.Now()
		err := showWindowStats(ctx, db, 7, *byProfile)
		logCommandDuration("stats", start, statsSlowThreshold,
			logging.String("window", "7d"),
			logging.Bool("by_profile", *byProfile),
		)
		return err
	}

	if *days30 {
		return showWindowStats(ctx, db, 30, *byProfile)
	}

	return runStats(home, []string{"--today"})
}

func logCommandDuration(command string, start time.Time, threshold time.Duration, extra ...zap.Field) {
	duration := time.Since(start)
	fields := []zap.Field{
		logging.String("command", command),
		logging.Int64("duration_ms", duration.Milliseconds()),
	}
	if len(extra) > 0 {
		fields = append(fields, extra...)
	}
	logging.Info("command_duration", fields...)
	if threshold > 0 && duration > threshold {
		slowFields := append([]zap.Field{}, fields...)
		slowFields = append(slowFields, logging.Int64("slow_threshold_ms", threshold.Milliseconds()))
		logging.Warn("command_duration_slow", slowFields...)
	}
}

func showWindowStats(ctx context.Context, db *sqlite.DB, days int, byProfile bool) error {
	to := time.Now()
	from := to.AddDate(0, 0, -days)
	summary, err := stats.Window(ctx, db, from, to)
	if err != nil {
		return err
	}
	printWindowSummary(days, summary)

	latencies, err := stats.P95Latency(ctx, db, time.Duration(days)*24*time.Hour, byProfile)
	if err != nil {
		if errors.Is(err, stats.ErrSchemaTooOld) {
			fmt.Println()
			fmt.Println("P95 latency requires database schema v3 or newer. Run 'boba doctor --db' to upgrade.")
		} else {
			return err
		}
	} else {
		printP95Latency(latencies)
	}

	if !byProfile {
		return nil
	}

	profiles, err := stats.NewAnalyzer(db).GetProfileStats(days)
	if err != nil {
		return err
	}
	printProfileBreakdown(profiles)
	return nil
}

func printTodaySummary(summary stats.Summary) {
	title := "Today's Usage"
	fmt.Println(title)
	fmt.Println(strings.Repeat("=", len(title)))
	fmt.Printf("Tokens:   %d\n", summary.TotalTokens)
	fmt.Printf("Cost:     $%.4f\n", summary.TotalCost)
	fmt.Printf("Sessions: %d\n", summary.TotalSessions)
}

func printWindowSummary(days int, summary stats.Summary) {
	fmt.Printf("Last %d Days Usage\n", days)
	fmt.Println(strings.Repeat("=", 20))
	fmt.Printf("Total Tokens:   %d\n", summary.TotalTokens)
	fmt.Printf("Total Cost:     $%.4f\n", summary.TotalCost)
	fmt.Printf("Total Sessions: %d\n", summary.TotalSessions)
	fmt.Printf("Avg Daily Tokens: %.2f\n", summary.AvgDailyTokens)
	fmt.Printf("Avg Daily Cost:   $%.4f\n", summary.AvgDailyCost)
}

func printP95Latency(latencies map[string]int64) {
	if len(latencies) == 0 {
		return
	}
	fmt.Println()
	fmt.Println("P95 Latency (ms):")
	fmt.Println("-----------------")
	keys := make([]string, 0, len(latencies))
	for k := range latencies {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		fmt.Printf("- %s: %dms\n", key, latencies[key])
	}
}

func printProfileBreakdown(statsByProfile []stats.ProfileStats) {
	if len(statsByProfile) == 0 {
		fmt.Println()
		fmt.Println("By Profile: (no data)")
		return
	}
	fmt.Println()
	fmt.Println("By Profile:")
	fmt.Println("-----------")
	for _, ps := range statsByProfile {
		fmt.Printf("- %s: tokens=%d cost=$%.4f sessions=%d avg_latency=%.0fms usage=%.1f%% cost=%.1f%%\n",
			ps.ProfileName,
			ps.TotalTokens,
			ps.TotalCost,
			ps.SessionCount,
			ps.AvgLatencyMS,
			ps.UsagePercent,
			ps.CostPercent,
		)
	}
}

func runEdit(home string, args []string) error {
	if len(args) != 1 {
		return errors.New("edit requires target")
	}
	name := args[0]
	allowed := map[string]string{
		"profiles": filepath.Join(home, "profiles.yaml"),
		"routes":   filepath.Join(home, "routes.yaml"),
		"pricing":  filepath.Join(home, "pricing.yaml"),
		"secrets":  filepath.Join(home, "secrets.yaml"),
	}
	path, ok := allowed[name]
	if !ok {
		return fmt.Errorf("unknown edit target %s", name)
	}
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		if err := os.WriteFile(path, []byte{}, 0o600); err != nil {
			return err
		}
	}
	if editor := os.Getenv("EDITOR"); editor != "" {
		// #nosec G204 -- EDITOR is intentionally user-configurable environment variable
		cmd := exec.CommandContext(context.Background(), editor, path)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		return cmd.Run()
	}
	fmt.Println(path)
	return nil
}

func runBudget(home string, args []string) error {
	flags := flag.NewFlagSet("budget", flag.ContinueOnError)
	status := flags.Bool("status", true, "show budget status summary")
	daily := flags.Float64("daily", 0, "set daily budget limit (USD)")
	cap := flags.Float64("cap", 0, "set hard cap (USD)")
	scopeFlag := flags.String("scope", "auto", "scope: auto|global|project|profile")
	targetFlag := flags.String("target", "", "scope target (profile or project name)")
	flags.SetOutput(io.Discard)
	if err := flags.Parse(args); err != nil {
		return err
	}

	dbPath := filepath.Join(home, "usage.db")
	db, err := sqlite.Open(dbPath)
	if err != nil {
		return err
	}
	tracker := budget.NewTracker(db)
	scope, target, projectCfg, err := resolveBudgetScope(*scopeFlag, *targetFlag)
	if err != nil {
		return err
	}
	if projectCfg != nil && projectCfg.Budget != nil && projectCfg.Budget.DailyUSD > 0 {
		if err := ensureBudget(tracker, "project", target, projectCfg.Budget); err != nil {
			return err
		}
	}
	if *daily > 0 || *cap > 0 {
		if err := applyBudgetLimits(tracker, scope, target, *daily, *cap); err != nil {
			return err
		}
		fmt.Println("Budget limits updated.")
	}
	if !*status {
		return nil
	}
	statusInfo, err := tracker.GetStatus(scope, target)
	if err != nil {
		return err
	}
	printBudgetStatus(scope, target, statusInfo)
	alerts := budget.NewAlertManager(tracker, nil).CheckBudgetAlerts(scope, target)
	if len(alerts) > 0 {
		fmt.Println()
		fmt.Println("Alerts:")
		for _, alert := range alerts {
			fmt.Printf("[%s] %s - %s\n", alert.Level, alert.Title, alert.Message)
		}
	}
	return nil
}

func resolveBudgetScope(scopeOpt, target string) (string, string, *config.ProjectConfig, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", "", nil, err
	}
	if scopeOpt == "auto" {
		cfg, path, err := config.FindProjectConfig(cwd)
		if err != nil {
			return "", "", nil, err
		}
		if cfg != nil {
			targetName := cfg.Project.Name
			if targetName == "" {
				targetName = filepath.Base(filepath.Dir(path))
			}
			return scopeProject, targetName, cfg, nil
		}
		return scopeGlobal, "", nil, nil
	}
	switch scopeOpt {
	case scopeGlobal:
		return scopeGlobal, "", nil, nil
	case "project":
		if target == "" {
			return "", "", nil, errors.New("--target required for project scope")
		}
		return "project", target, nil, nil
	case "profile":
		if target == "" {
			return "", "", nil, errors.New("--target required for profile scope")
		}
		return "profile", target, nil, nil
	default:
		return "", "", nil, fmt.Errorf("unknown scope %s", scopeOpt)
	}
}

func ensureBudget(tracker *budget.Tracker, scope, target string, cfg *config.BudgetSettings) error {
	if cfg == nil {
		return nil
	}
	entry, err := tracker.GetBudget(scope, target)
	if err != nil {
		_, createErr := tracker.CreateBudget(scope, target, cfg.DailyUSD, cfg.HardCap)
		return createErr
	}
	return tracker.UpdateLimits(entry.ID, cfg.DailyUSD, cfg.HardCap)
}

func applyBudgetLimits(tracker *budget.Tracker, scope, target string, daily, cap float64) error {
	budgetEntry, err := tracker.GetBudget(scope, target)
	if err != nil {
		_, err = tracker.CreateBudget(scope, target, daily, cap)
		return err
	}
	if daily == 0 {
		daily = budgetEntry.DailyUSD
	}
	if cap == 0 {
		cap = budgetEntry.HardCapUSD
	}
	return tracker.UpdateLimits(budgetEntry.ID, daily, cap)
}

func printBudgetStatus(scope, target string, status *budget.Status) {
	fmt.Printf("Budget Scope: %s (%s)\n", scope, target)
	fmt.Println(strings.Repeat("=", 40))
	fmt.Printf("Today:  $%.4f of $%.2f (%.1f%%)\n", status.CurrentSpent, status.DailyLimit, status.DailyProgress)
	fmt.Printf("Period: $%.4f of $%.2f (%.1f%%)\n", status.Budget.SpentUSD, status.HardCap, status.TotalProgress)
	fmt.Printf("Days Remaining: %d\n", status.DaysRemaining)
	if level := status.GetWarningLevel(); level != "none" {
		fmt.Printf("Warning Level: %s\n", strings.ToUpper(level))
	}
}

func runHooks(home string, args []string) error {
	if len(args) == 0 {
		return errors.New("hooks requires subcommand")
	}
	manager := hooks.NewManager(home)
	switch args[0] {
	case "install":
		repo, err := findRepoRootFromArgs(args[1:])
		if err != nil {
			return err
		}
		return manager.Install(repo)
	case "remove":
		repo, err := findRepoRootFromArgs(args[1:])
		if err != nil {
			return err
		}
		return manager.Remove(repo)
	case "track":
		flags := flag.NewFlagSet("track", flag.ContinueOnError)
		event := flags.String("event", "", "git hook event")
		repo := flags.String("repo", "", "repo path")
		branch := flags.String("branch", "", "branch name")
		flags.SetOutput(io.Discard)
		if err := flags.Parse(args[1:]); err != nil {
			return err
		}
		return manager.Record(*event, *repo, *branch)
	default:
		return fmt.Errorf("unknown hooks subcommand %s", args[0])
	}
}

func findRepoRootFromArgs(args []string) (string, error) {
	var start string
	if len(args) > 0 {
		start = args[0]
	}
	if start == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		start = cwd
	}
	return findRepoRoot(start)
}

func runTUI(home string) error {
	return ui.Run(home)
}

func runAction(home string, args []string) error {
	flags := flag.NewFlagSet("action", flag.ContinueOnError)
	auto := flags.Bool("auto", false, "automatically apply actionable suggestions")
	days := flags.Int("days", 7, "analysis window in days")
	flags.SetOutput(io.Discard)
	if err := flags.Parse(args); err != nil {
		return err
	}

	dbPath := filepath.Join(home, "usage.db")
	db, err := sqlite.Open(dbPath)
	if err != nil {
		return err
	}

	engine := suggestions.NewEngine(db)
	suggs, err := engine.GenerateSuggestions(*days)
	if err != nil {
		return err
	}
	if len(suggs) == 0 {
		fmt.Println("No suggestions available.")
		return nil
	}

	if !*auto {
		for _, s := range suggs {
			fmt.Println(s.FormatSuggestion())
		}
		return nil
	}

	tracker := budget.NewTracker(db)
	profiles, err := config.LoadProfiles(home)
	if err != nil {
		return err
	}
	app := suggestions.NewApplicator(home, tracker, profiles)
	applied := 0
	for _, s := range suggs {
		if s.Priority < 3 {
			continue
		}
		summary, err := app.Apply(s)
		if err != nil {
			fmt.Printf("%s %s: %v\n", statusError, s.Title, err)
			continue
		}
		fmt.Printf("%s %s -> %s\n", statusOK, s.Title, summary)
		applied++
	}
	if applied == 0 {
		fmt.Println("No suggestions were applicable.")
	}
	return nil
}

//nolint:gocyclo // Complex report generation with multiple output formats and filters
func runReport(home string, args []string) error {
	flags := flag.NewFlagSet("report", flag.ContinueOnError)
	days := flags.Int("days", 7, "number of days to include")
	format := flags.String("format", "json", "output format: json or csv")
	output := flags.String("out", "", "output file path")
	flags.SetOutput(io.Discard)
	if err := flags.Parse(args); err != nil {
		return err
	}

	dbPath := filepath.Join(home, "usage.db")
	db, err := sqlite.Open(dbPath)
	if err != nil {
		return err
	}

	analyzer := stats.NewAnalyzer(db)
	trend, err := analyzer.GetTrend(*days)
	if err != nil {
		return err
	}
	profiles, err := analyzer.GetProfileStats(*days)
	if err != nil {
		return err
	}
	suggs, err := suggestions.NewEngine(db).GenerateSuggestions(*days)
	if err != nil {
		return err
	}

	fileName := *output
	if fileName == "" {
		fileName = fmt.Sprintf("bobamixer-report-%s.%s", time.Now().Format("20060102-1504"), *format)
	}
	if !filepath.IsAbs(fileName) {
		fileName = filepath.Join(home, fileName)
	}
	if err := os.MkdirAll(filepath.Dir(fileName), 0o750); err != nil {
		return err
	}

	switch strings.ToLower(*format) {
	case "json":
		payload := map[string]interface{}{
			"generated_at": time.Now().Format(time.RFC3339),
			"days":         *days,
			"summary":      trend.Summary,
			"trend":        trend.DataPoints,
			"profiles":     profiles,
			"suggestions":  suggs,
		}
		data, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(fileName, data, 0o600); err != nil {
			return err
		}
	default:
		// #nosec G304 -- fileName is constructed from validated user input and home directory
		f, err := os.Create(fileName)
		if err != nil {
			return err
		}
		defer func() {
			//nolint:errcheck,gosec // Best effort cleanup, error irrelevant in defer
			f.Close()
		}()
		writer := csv.NewWriter(f)
		defer writer.Flush()
		if err := writer.Write([]string{"date", "tokens", "cost", "sessions"}); err != nil {
			return err
		}
		for _, dp := range trend.DataPoints {
			if err := writer.Write([]string{
				dp.Date,
				fmt.Sprintf("%d", dp.Tokens),
				fmt.Sprintf("%.4f", dp.Cost),
				fmt.Sprintf("%d", dp.Count),
			}); err != nil {
				return err
			}
		}
		if err := writer.Error(); err != nil {
			return err
		}
	}

	fmt.Printf("Report exported to %s\n", fileName)
	return nil
}

func runInit(home string, args []string) error {
	flags := flag.NewFlagSet("init", flag.ContinueOnError)
	mode := flags.String("mode", "", "operation mode: observer|suggest|apply")
	theme := flags.String("theme", "", "ui theme: auto|dark|light")
	disableExplore := flags.Bool("disable-explore", false, "disable exploration suggestions")
	exploreRate := flags.Float64("explore-rate", -1, "exploration rate between 0 and 1")
	flags.SetOutput(io.Discard)
	if err := flags.Parse(args); err != nil {
		return err
	}

	// Initialize default control plane configs
	logging.Info("Initializing BobaMixer configuration", logging.String("home", home))
	if err := core.InitDefaultConfigs(home); err != nil {
		return fmt.Errorf("failed to initialize configs: %w", err)
	}

	// Initialize settings
	if err := settings.InitHome(home); err != nil {
		return err
	}
	ctx := context.Background()
	current, err := settings.Load(ctx, home)
	if err != nil {
		return err
	}
	if *mode != "" {
		current.Mode = settings.Mode(*mode)
	} else if current.Mode == "" {
		current.Mode = settings.ModeObserver
	}
	if *theme != "" {
		current.Theme = *theme
	}
	if *disableExplore {
		current.Explore.Enabled = false
	}
	if *exploreRate >= 0 {
		current.Explore.Rate = *exploreRate
		current.Explore.Enabled = true
	}
	if err := settings.Save(ctx, home, current); err != nil {
		return err
	}

	fmt.Println("✅ BobaMixer initialized successfully")
	fmt.Println()
	fmt.Println("Configuration directory:", home)
	fmt.Println()
	fmt.Println("Created files:")
	fmt.Println("  - providers.yaml  (AI service providers)")
	fmt.Println("  - tools.yaml      (Local CLI tools)")
	fmt.Println("  - bindings.yaml   (Tool ↔ Provider bindings)")
	fmt.Println("  - secrets.yaml    (API keys)")
	fmt.Println("  - settings.yaml   (UI preferences)")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Add your API keys to environment variables or secrets.yaml")
	fmt.Println("  2. Run 'boba tools' to see detected CLI tools")
	fmt.Println("  3. Run 'boba providers' to see available providers")
	fmt.Println("  4. Run 'boba bind <tool> <provider>' to create bindings")
	fmt.Println("  5. Run 'boba doctor' to verify your setup")
	fmt.Println()

	return nil
}

func findRepoRoot(start string) (string, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("repo root not found")
		}
		dir = parent
	}
}

func runRoute(home string, args []string) error {
	if len(args) == 0 {
		return errors.New("route subcommand required (test)")
	}

	switch args[0] {
	case "test":
		return runRouteTest(home, args[1:])
	default:
		return fmt.Errorf("unknown route subcommand: %s", args[0])
	}
}

//nolint:gocyclo // Complex route testing with multiple output formats and conditions
func runRouteTest(home string, args []string) error {
	flags := flag.NewFlagSet("route test", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	if err := flags.Parse(args); err != nil {
		return err
	}

	if flags.NArg() == 0 {
		return errors.New("route test requires text or @file argument")
	}

	// Load configurations
	routes, err := config.LoadRoutes(home)
	if err != nil {
		return fmt.Errorf("load routes: %w", err)
	}

	activeProfile, err := config.LoadActiveProfile(home)
	if err != nil {
		activeProfile = "default"
	}

	// Get text input
	input := flags.Arg(0)
	var text string
	if strings.HasPrefix(input, "@") {
		// Read from file
		filePath := strings.TrimPrefix(input, "@")
		// #nosec G304 -- user-provided file path for route testing
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("read file: %w", err)
		}
		text = string(data)
	} else {
		text = input
	}

	// Detect project context
	cwd, _ := os.Getwd() //nolint:errcheck
	project := ""
	branch := ""
	projectTypes := []string{}

	if repoRoot, err := findRepoRoot(cwd); err == nil {
		projectCfg, _, _ := config.FindProjectConfig(repoRoot) //nolint:errcheck
		if projectCfg != nil {
			project = projectCfg.Project.Name
			projectTypes = projectCfg.Project.Type
		}

		// Get git branch
		cmd := exec.CommandContext(context.Background(), "git", "rev-parse", "--abbrev-ref", "HEAD")
		cmd.Dir = repoRoot
		if output, err := cmd.Output(); err == nil {
			branch = strings.TrimSpace(string(output))
		}
	}

	// Determine time of day
	hour := time.Now().Hour()
	timeOfDay := "day"
	if hour < 6 || hour >= 22 {
		timeOfDay = "night"
	} else if hour >= 18 {
		timeOfDay = "evening"
	}

	// Build routing context
	ctx := routing.Context{
		Text:        text,
		CtxChars:    len(text),
		Project:     project,
		Branch:      branch,
		ProjectType: projectTypes,
		TimeOfDay:   timeOfDay,
	}

	// Route decision
	router := routing.NewRouter(routes)
	decision := router.Route(ctx, activeProfile)

	// Display results
	fmt.Println("=== Route Test Results ===")
	fmt.Printf("Text length: %d chars\n", ctx.CtxChars)
	if ctx.Project != "" {
		fmt.Printf("Project: %s (types: %v)\n", ctx.Project, ctx.ProjectType)
	}
	if ctx.Branch != "" {
		fmt.Printf("Branch: %s\n", ctx.Branch)
	}
	fmt.Printf("Time of day: %s\n", ctx.TimeOfDay)
	fmt.Println()

	fmt.Println("=== Routing Decision ===")
	fmt.Printf("Profile: %s\n", decision.ProfileKey)
	if decision.RuleID != "" {
		fmt.Printf("Rule ID: %s\n", decision.RuleID)
	}
	if decision.Explain != "" {
		fmt.Printf("Explanation: %s\n", decision.Explain)
	}
	if decision.Explore {
		fmt.Println("(exploration mode)")
	}
	if decision.Fallback != "" {
		fmt.Printf("Fallback: %s\n", decision.Fallback)
	}

	return nil
}

func runCall(home string, args []string) error {
	flags := flag.NewFlagSet("call", flag.ContinueOnError)
	profileFlag := flags.String("profile", "", "profile to use")
	dataFlag := flags.String("data", "", "data file (use @file.json syntax)")
	flags.SetOutput(io.Discard)
	if err := flags.Parse(args); err != nil {
		return err
	}

	// Load active profile if not specified
	profileKey := *profileFlag
	if profileKey == "" {
		active, err := config.LoadActiveProfile(home)
		if err != nil {
			return fmt.Errorf("no active profile, use --profile or run 'boba use <profile>'")
		}
		profileKey = active
	}

	// Read data payload
	var payload []byte
	var err error
	if *dataFlag == "" {
		return errors.New("--data is required (use @file.json syntax)")
	}

	if strings.HasPrefix(*dataFlag, "@") {
		// Read from file
		filePath := strings.TrimPrefix(*dataFlag, "@")
		// #nosec G304 -- user-provided file path for call data
		payload, err = os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("read data file: %w", err)
		}
	} else {
		// Use as inline JSON
		payload = []byte(*dataFlag)
	}

	// Open database
	dbPath := filepath.Join(home, "usage.db")
	db, err := sqlite.Open(dbPath)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}

	// Create executor
	executor, err := svc.NewExecutor(db, home)
	if err != nil {
		return fmt.Errorf("create executor: %w", err)
	}

	// Get current directory context
	cwd, _ := os.Getwd() //nolint:errcheck
	project := ""
	branch := ""

	if repoRoot, err := findRepoRoot(cwd); err == nil {
		projectCfg, _, _ := config.FindProjectConfig(repoRoot) //nolint:errcheck
		if projectCfg != nil {
			project = projectCfg.Project.Name
		}

		// Get git branch
		cmd := exec.CommandContext(context.Background(), "git", "rev-parse", "--abbrev-ref", "HEAD")
		cmd.Dir = repoRoot
		if output, err := cmd.Output(); err == nil {
			branch = strings.TrimSpace(string(output))
		}
	}

	// Execute call
	req := svc.ExecuteRequest{
		ProfileKey: profileKey,
		Payload:    payload,
		Project:    project,
		Branch:     branch,
		TaskType:   "api-call",
	}

	fmt.Printf("Calling %s...\n", profileKey)
	result, err := executor.Execute(context.Background(), req)
	if err != nil {
		return fmt.Errorf("execute: %w", err)
	}

	// Display result
	fmt.Println()
	if !result.Success {
		fmt.Printf("❌ Call failed: %s\n", result.Error)
		return nil
	}

	fmt.Println("✅ Call succeeded")
	fmt.Println()
	fmt.Println("Response:")
	fmt.Println(string(result.Output))
	fmt.Println()
	fmt.Printf("Session ID: %s\n", result.SessionID)
	fmt.Printf("Tokens: %d in + %d out = %d total\n",
		result.Usage.InputTokens,
		result.Usage.OutputTokens,
		result.Usage.InputTokens+result.Usage.OutputTokens)
	fmt.Printf("Estimate level: %s\n", getEstimateLevel(result.Usage))
	fmt.Printf("Latency: %dms\n", result.Usage.LatencyMS)

	return nil
}

func getEstimateLevel(usage adapters.Usage) string {
	switch usage.Estimate {
	case adapters.EstimateExact:
		return "exact"
	case adapters.EstimateMapped:
		return "mapped"
	case adapters.EstimateHeuristic:
		return "heuristic"
	default:
		return "unknown"
	}
}

func runCompletions(args []string) error {
	if len(args) == 0 {
		return errors.New("completions subcommand required (install|uninstall)")
	}

	switch args[0] {
	case "install":
		return runCompletionsInstall(args[1:])
	case "uninstall":
		return runCompletionsUninstall(args[1:])
	default:
		return fmt.Errorf("unknown completions subcommand: %s", args[0])
	}
}

func runCompletionsInstall(args []string) error {
	flags := flag.NewFlagSet("completions install", flag.ContinueOnError)
	shell := flags.String("shell", shellBash, "shell type: bash|zsh|fish")
	flags.SetOutput(io.Discard)
	if err := flags.Parse(args); err != nil {
		return err
	}

	// Determine completion file path based on shell
	var destPath string
	var completionScript string
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("determine home directory: %w", err)
	}

	switch *shell {
	case shellBash:
		destPath = filepath.Join(homeDir, ".bash_completion.d", "boba")
		completionScript = `# Bash completion for boba
_boba_completion() {
    local cur prev opts
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"

    opts="ls use call stats edit doctor budget hooks action report route completions suggest version"

    if [[ ${COMP_CWORD} -eq 1 ]]; then
        COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) )
        return 0
    fi

    case "${prev}" in
        edit)
            COMPREPLY=( $(compgen -W "profiles routes pricing secrets" -- ${cur}) )
            ;;
        hooks)
            COMPREPLY=( $(compgen -W "install remove track" -- ${cur}) )
            ;;
        completions)
            COMPREPLY=( $(compgen -W "install uninstall" -- ${cur}) )
            ;;
        *)
            ;;
    esac
}

complete -F _boba_completion boba
`
	case shellZsh:
		destPath = filepath.Join(homeDir, ".zsh", "completions", "_boba")
		completionScript = `#compdef boba

_boba() {
    local -a commands
    commands=(
        'ls:List profiles'
        'use:Activate a profile'
        'call:Execute an AI call'
        'stats:Show usage statistics'
        'edit:Edit configuration files'
        'doctor:Run diagnostics'
        'budget:Show budget status'
        'hooks:Manage git hooks'
        'action:View/apply suggestions'
        'report:Generate usage report'
        'route:Test routing rules'
        'completions:Manage shell completions'
        'suggest:Get profile suggestions'
        'version:Show version info'
    )

    _arguments '1: :->command' '*:: :->args'

    case $state in
        command)
            _describe 'command' commands
            ;;
        args)
            case $words[1] in
                edit)
                    compadd profiles routes pricing secrets
                    ;;
                hooks)
                    compadd install remove track
                    ;;
                completions)
                    compadd install uninstall
                    ;;
            esac
            ;;
    esac
}

_boba
`
	case shellFish:
		destPath = filepath.Join(homeDir, ".config", "fish", "completions", "boba.fish")
		completionScript = `# Fish completion for boba
function __boba_subcommands
    set -l commands budget hooks action report route completions suggest version
    printf "%s\n" $commands
end

complete -c boba -n "__fish_use_subcommand" -a "ls" -d "List profiles"
complete -c boba -n "__fish_use_subcommand" -a "use" -d "Activate a profile"
complete -c boba -n "__fish_use_subcommand" -a "call" -d "Execute an AI call"
complete -c boba -n "__fish_use_subcommand" -a "stats" -d "Show usage statistics"
complete -c boba -n "__fish_use_subcommand" -a "edit" -d "Edit configuration files"
complete -c boba -n "__fish_use_subcommand" -a "doctor" -d "Run diagnostics"
complete -c boba -n "__fish_use_subcommand" -a "budget" -d "Show budget status"
complete -c boba -n "__fish_use_subcommand" -a "hooks" -d "Manage git hooks"
complete -c boba -n "__fish_use_subcommand" -a "action" -d "View/apply suggestions"
complete -c boba -n "__fish_use_subcommand" -a "report" -d "Generate usage report"
complete -c boba -n "__fish_use_subcommand" -a "route" -d "Test routing rules"
complete -c boba -n "__fish_use_subcommand" -a "completions" -d "Manage shell completions"
complete -c boba -n "__fish_use_subcommand" -a "suggest" -d "Get profile suggestions"
complete -c boba -n "__fish_use_subcommand" -a "version" -d "Show version info"

# Subcommands
complete -c boba -n "__fish_seen_subcommand_from edit" -a "profiles routes pricing secrets"
complete -c boba -n "__fish_seen_subcommand_from hooks" -a "install remove track"
complete -c boba -n "__fish_seen_subcommand_from completions" -a "install uninstall"
`
	default:
		return fmt.Errorf("unsupported shell: %s (supported: %s)", *shell, strings.Join(supportedShells, ", "))
	}

	// Create directory if needed
	if err := os.MkdirAll(filepath.Dir(destPath), 0o750); err != nil {
		return fmt.Errorf("create completion directory: %w", err)
	}

	// Write completion script
	if err := os.WriteFile(destPath, []byte(completionScript), 0o600); err != nil {
		return fmt.Errorf("write completion script: %w", err)
	}

	fmt.Printf("%s Completion script installed to %s\n", statusOK, destPath)
	fmt.Println()

	// Print instructions
	switch *shell {
	case shellBash:
		fmt.Println("Add the following to your ~/.bashrc:")
		fmt.Println("  source ~/.bash_completion.d/boba")
	case shellZsh:
		fmt.Println("Add the following to your ~/.zshrc:")
		fmt.Println("  fpath=(~/.zsh/completions $fpath)")
		fmt.Println("  autoload -Uz compinit && compinit")
	case shellFish:
		fmt.Println("Completion will be loaded automatically in new fish sessions.")
	}

	return nil
}

func runCompletionsUninstall(args []string) error {
	flags := flag.NewFlagSet("completions uninstall", flag.ContinueOnError)
	shell := flags.String("shell", shellBash, "shell type: bash|zsh|fish")
	flags.SetOutput(io.Discard)
	if err := flags.Parse(args); err != nil {
		return err
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("determine home directory: %w", err)
	}
	var destPath string

	switch *shell {
	case shellBash:
		destPath = filepath.Join(homeDir, ".bash_completion.d", "boba")
	case shellZsh:
		destPath = filepath.Join(homeDir, ".zsh", "completions", "_boba")
	case shellFish:
		destPath = filepath.Join(homeDir, ".config", "fish", "completions", "boba.fish")
	default:
		return fmt.Errorf("unsupported shell: %s (supported: %s)", *shell, strings.Join(supportedShells, ", "))
	}

	if err := os.Remove(destPath); err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Completion script not found")
			return nil
		}
		return fmt.Errorf("remove completion script: %w", err)
	}

	fmt.Printf("%s Completion script removed from %s\n", statusOK, destPath)
	return nil
}

func runSuggest(_ []string) error {
	// Get current directory context
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("determine working directory: %w", err)
	}
	branch := ""
	project := ""

	if repoRoot, err := findRepoRoot(cwd); err == nil {
		projectCfg, _, cfgErr := config.FindProjectConfig(repoRoot)
		if cfgErr != nil {
			return fmt.Errorf("find project config: %w", cfgErr)
		}
		if projectCfg != nil {
			project = projectCfg.Project.Name
			// Get preferred profiles from project config
			if len(projectCfg.Project.PreferredProfiles) > 0 {
				fmt.Println("=== Recommended Profiles for", project, "===")
				for _, prof := range projectCfg.Project.PreferredProfiles {
					fmt.Printf("  • %s\n", prof)
				}
				fmt.Println()
				fmt.Printf("Tip: Use 'boba use %s' to switch\n", projectCfg.Project.PreferredProfiles[0])
				return nil
			}
		}

		// Get git branch
		cmd := exec.CommandContext(context.Background(), "git", "rev-parse", "--abbrev-ref", "HEAD")
		cmd.Dir = repoRoot
		if output, cmdErr := cmd.Output(); cmdErr == nil {
			branch = strings.TrimSpace(string(output))
		}
	}

	fmt.Println("=== Profile Suggestion ===")
	if project != "" {
		fmt.Printf("Project: %s\n", project)
	}
	if branch != "" {
		fmt.Printf("Branch: %s\n", branch)
	}
	fmt.Println()
	fmt.Println("No specific profile recommendation configured for this project.")
	fmt.Println("Tip: Add preferred_profiles to .boba-project.yaml")

	return nil
}
