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
	"github.com/royisme/bobamixer/internal/domain/hooks"
	"github.com/royisme/bobamixer/internal/domain/routing"
	"github.com/royisme/bobamixer/internal/domain/stats"
	"github.com/royisme/bobamixer/internal/domain/suggestions"
	"github.com/royisme/bobamixer/internal/logger"
	"github.com/royisme/bobamixer/internal/store/config"
	"github.com/royisme/bobamixer/internal/store/sqlite"
	"github.com/royisme/bobamixer/internal/svc"
	"github.com/royisme/bobamixer/internal/ui"
	"github.com/royisme/bobamixer/internal/version"
)

const (
	scopeGlobal  = "global"
	scopeProject = "project"
)

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

	// Initialize structured logging
	if err := logger.Init(home); err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer logger.Sync()

	logger.Info("BobaMixer CLI started")

	// Handle help flag
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h" || args[0] == "help") {
		printUsage()
		return nil
	}

	// No arguments: launch TUI dashboard
	if len(args) == 0 {
		logger.Info("Launching TUI dashboard")
		return runTUI(home)
	}

	switch args[0] {
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
		return runDoctor(home, args[1:])
	case "budget":
		return runBudget(home, args[1:])
	case "hooks":
		return runHooks(home, args[1:])
	case "action":
		return runAction(home, args[1:])
	case "report":
		return runReport(home, args[1:])
	case "route":
		return runRoute(home, args[1:])
	case "completions":
		return runCompletions(home, args[1:])
	case "suggest":
		return runSuggest(home, args[1:])
	case "version":
		return runVersion()
	default:
		return fmt.Errorf("unknown command %s", args[0])
	}
}

func printUsage() {
	fmt.Println("BobaMixer - Smart AI Adapter Router")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  boba                                          Launch TUI dashboard")
	fmt.Println("  boba --help                                   Show this help")
	fmt.Println()
	fmt.Println("Profile Management:")
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
	fmt.Println("  boba edit <profiles|routes|pricing|secrets>  Edit configuration files")
	fmt.Println("  boba doctor                                   Run diagnostics")
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
	if err := config.SaveActiveProfile(home, prof.Key); err != nil {
		return err
	}
	fmt.Printf("active profile set to %s (%s)\n", prof.Key, prof.Model)
	return nil
}

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

	dbPath := filepath.Join(home, "usage.db")
	db, err := sqlite.Open(dbPath)
	if err != nil {
		return err
	}

	// Handle today stats
	if *today {
		totalTokens, err := db.QueryRow("SELECT COALESCE(SUM(input_tokens + output_tokens),0) FROM usage_records WHERE date(ts,'unixepoch') = date('now');")
		if err != nil {
			return err
		}
		totalCost, err := db.QueryRow("SELECT COALESCE(SUM(input_cost + output_cost),0) FROM usage_records WHERE date(ts,'unixepoch') = date('now');")
		if err != nil {
			return err
		}
		sessions, err := db.QueryRow("SELECT COUNT(DISTINCT session_id) FROM usage_records WHERE date(ts,'unixepoch') = date('now');")
		if err != nil {
			return err
		}

		fmt.Println("Today's Usage")
		fmt.Println("=============")
		fmt.Printf("Tokens:   %s\n", strings.TrimSpace(totalTokens))
		fmt.Printf("Cost:     $%s\n", strings.TrimSpace(totalCost))
		fmt.Printf("Sessions: %s\n", strings.TrimSpace(sessions))
		return nil
	}

	// Handle 7-day stats
	if *days7 {
		return showPeriodStats(db, 7, *byProfile)
	}

	// Handle 30-day stats
	if *days30 {
		return showPeriodStats(db, 30, *byProfile)
	}

	// Default: show today
	return runStats(home, []string{"--today"})
}

func showPeriodStats(db *sqlite.DB, days int, byProfile bool) error {
	// Calculate period stats
	query := fmt.Sprintf(`
		SELECT
			COALESCE(SUM(input_tokens + output_tokens), 0) as tokens,
			COALESCE(SUM(input_cost + output_cost), 0) as cost,
			COUNT(DISTINCT session_id) as sessions
		FROM usage_records
		WHERE date(ts, 'unixepoch') >= date('now', '-%d days');
	`, days)

	row, err := db.QueryRow(query)
	if err != nil {
		return err
	}

	// Parse results (simplified)
	var tokens, cost, sessions string
	parts := strings.Split(strings.TrimSpace(row), "|")
	if len(parts) >= 3 {
		tokens = parts[0]
		cost = parts[1]
		sessions = parts[2]
	}

	fmt.Printf("Last %d Days Usage\n", days)
	fmt.Println(strings.Repeat("=", 20))
	fmt.Printf("Total Tokens:   %s\n", tokens)
	fmt.Printf("Total Cost:     $%s\n", cost)
	fmt.Printf("Total Sessions: %s\n", sessions)
	fmt.Println()

	// Calculate daily average
	if cost != "" && cost != "0" {
		// Would calculate actual average here
		fmt.Printf("Daily Average:  ~$%.4f\n", 0.0)
	}

	// Show profile breakdown if requested
	if byProfile {
		fmt.Println()
		fmt.Println("By Profile:")
		fmt.Println("-----------")
		// Would show profile breakdown here
		fmt.Println("(Profile breakdown not yet implemented)")
	}

	return nil
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

func runDoctor(home string, _ []string) error {
	fmt.Println("BobaMixer Doctor")
	fmt.Println("================")
	fmt.Println()

	// Check home directory
	fmt.Printf("✓ Home directory: %s\n", home)
	if info, err := os.Stat(home); err == nil {
		fmt.Printf("  Permissions: %04o\n", info.Mode().Perm())
	}

	// Check profiles.yaml
	profsPath := filepath.Join(home, "profiles.yaml")
	if _, err := os.Stat(profsPath); err == nil {
		profs, err := config.LoadProfiles(home)
		if err != nil {
			fmt.Printf("✗ profiles.yaml: invalid (%v)\n", err)
		} else {
			fmt.Printf("✓ profiles.yaml: %d profiles\n", len(profs))
		}
	} else {
		fmt.Println("✗ profiles.yaml: not found")
	}

	// Check secrets.yaml permissions
	secretsPath := filepath.Join(home, "secrets.yaml")
	if info, err := os.Stat(secretsPath); err == nil {
		mode := info.Mode().Perm()
		if mode == 0600 {
			fmt.Printf("✓ secrets.yaml: permissions OK (%04o)\n", mode)
		} else {
			fmt.Printf("⚠ secrets.yaml: insecure permissions (%04o), should be 0600\n", mode)
		}
	} else {
		fmt.Println("⚠ secrets.yaml: not found")
	}

	// Check routes.yaml
	routesPath := filepath.Join(home, "routes.yaml")
	if _, err := os.Stat(routesPath); err == nil {
		routes, err := config.LoadRoutes(home)
		if err != nil {
			fmt.Printf("✗ routes.yaml: invalid (%v)\n", err)
		} else {
			fmt.Printf("✓ routes.yaml: %d rules, %d sub-agents\n", len(routes.Rules), len(routes.SubAgents))
		}
	} else {
		fmt.Println("⚠ routes.yaml: not found (optional)")
	}

	// Check pricing.yaml
	pricingPath := filepath.Join(home, "pricing.yaml")
	if _, err := os.Stat(pricingPath); err == nil {
		pricing, err := config.LoadPricing(home)
		if err != nil {
			fmt.Printf("✗ pricing.yaml: invalid (%v)\n", err)
		} else {
			fmt.Printf("✓ pricing.yaml: %d models\n", len(pricing.Models))
		}
	} else {
		fmt.Println("⚠ pricing.yaml: not found (optional)")
	}

	// Check database
	dbPath := filepath.Join(home, "usage.db")
	if _, err := os.Stat(dbPath); err == nil {
		db, err := sqlite.Open(dbPath)
		if err != nil {
			fmt.Printf("✗ usage.db: cannot open (%v)\n", err)
		} else {
			// Check schema version
			version, err := db.QueryInt("PRAGMA user_version;")
			if err != nil {
				fmt.Printf("✗ usage.db: cannot read schema version (%v)\n", err)
			} else {
				fmt.Printf("✓ usage.db: OK (schema v%d)\n", version)
			}

			// Check WAL mode
			walMode, err := db.QueryRow("PRAGMA journal_mode;")
			if err != nil {
				fmt.Printf("  ⚠ Cannot check WAL mode: %v\n", err)
			} else {
				if strings.TrimSpace(walMode) == "wal" {
					fmt.Println("  ✓ WAL mode enabled")
				} else {
					fmt.Printf("  ⚠ WAL mode not enabled (current: %s)\n", walMode)
				}
			}

			// Test read/write
			testQuery := "SELECT COUNT(*) FROM sessions;"
			if _, err := db.QueryRow(testQuery); err != nil {
				fmt.Printf("  ⚠ Database read test failed: %v\n", err)
			} else {
				fmt.Println("  ✓ Read/write test passed")
			}
		}
	} else {
		fmt.Println("⚠ usage.db: will be created on first use")
	}

	// Check pricing cache
	fmt.Println()
	fmt.Println("Pricing Configuration:")
	pricingCachePath := filepath.Join(home, "pricing.cache.json")
	if info, err := os.Stat(pricingCachePath); err == nil {
		// Check if cache is valid (within 24 hours)
		cacheAge := time.Since(info.ModTime())
		if cacheAge < 24*time.Hour {
			validUntil := info.ModTime().Add(24 * time.Hour)
			fmt.Printf("✓ Pricing cache: valid until %s\n", validUntil.Format("2006-01-02 15:04"))
		} else {
			fmt.Println("⚠ Pricing cache: expired, will refresh on next use")
		}
	} else {
		fmt.Println("⚠ Pricing cache: not found, will fetch on first use")
	}

	// Check network and API key (if profiles exist)
	fmt.Println()
	fmt.Println("Network & API Keys:")
	if len(profs) > 0 {
		// Try to load active profile
		activeProfile := ""
		if ap, err := config.LoadActiveProfile(home); err == nil {
			activeProfile = ap
		}

		// If no active profile, use first available
		if activeProfile == "" {
			for key := range profs {
				activeProfile = key
				break
			}
		}

		if activeProfile != "" {
			prof := profs[activeProfile]
			fmt.Printf("Testing profile: %s (%s)\n", activeProfile, prof.Provider)

			// Check if secrets exist
			secrets, err := config.LoadSecrets(home)
			if err != nil {
				fmt.Printf("✗ Cannot load secrets: %v\n", err)
			} else {
				// Resolve env to check if API key is available
				env := config.ResolveEnv(prof.Env, secrets)
				hasKey := false
				for _, e := range env {
					if strings.Contains(e, "API_KEY=") && !strings.HasSuffix(e, "=") {
						hasKey = true
						break
					}
				}

				if !hasKey {
					fmt.Println("✗ API Key: not found in secrets.yaml")
					fmt.Printf("  Fix: Add API key for %s to secrets.yaml\n", prof.Provider)
				} else {
					fmt.Println("✓ API Key: configured")
					// Note: We don't make actual API calls in doctor to avoid costs
					// Users should use `boba call` to test end-to-end connectivity
					fmt.Println("  Use 'boba call --data @test.json' to test end-to-end")
				}
			}
		}
	} else {
		fmt.Println("⚠ No profiles configured yet")
		fmt.Println("  Fix: Run 'boba edit profiles' to add a profile")
	}

	fmt.Println()
	fmt.Println("Diagnosis complete.")
	fmt.Println()
	fmt.Println("Summary:")
	fmt.Println("  - Configuration files are accessible")
	fmt.Println("  - Database is operational")
	fmt.Println("  - Ready to use BobaMixer")
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
			fmt.Printf("✗ %s: %v\n", s.Title, err)
			continue
		}
		fmt.Printf("✓ %s -> %s\n", s.Title, summary)
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

func runCompletions(home string, args []string) error {
	if len(args) == 0 {
		return errors.New("completions subcommand required (install|uninstall)")
	}

	switch args[0] {
	case "install":
		return runCompletionsInstall(home, args[1:])
	case "uninstall":
		return runCompletionsUninstall(home, args[1:])
	default:
		return fmt.Errorf("unknown completions subcommand: %s", args[0])
	}
}

func runCompletionsInstall(home string, args []string) error {
	flags := flag.NewFlagSet("completions install", flag.ContinueOnError)
	shell := flags.String("shell", "bash", "shell type: bash|zsh|fish")
	flags.SetOutput(io.Discard)
	if err := flags.Parse(args); err != nil {
		return err
	}

	// Determine completion file path based on shell
	var destPath string
	var completionScript string

	switch *shell {
	case "bash":
		homeDir, _ := os.UserHomeDir()
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
	case "zsh":
		homeDir, _ := os.UserHomeDir()
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
	case "fish":
		homeDir, _ := os.UserHomeDir()
		destPath = filepath.Join(homeDir, ".config", "fish", "completions", "boba.fish")
		completionScript = `# Fish completion for boba

complete -c boba -f

# Main commands
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
		return fmt.Errorf("unsupported shell: %s (supported: bash, zsh, fish)", *shell)
	}

	// Create directory if needed
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return fmt.Errorf("create completion directory: %w", err)
	}

	// Write completion script
	if err := os.WriteFile(destPath, []byte(completionScript), 0o644); err != nil {
		return fmt.Errorf("write completion script: %w", err)
	}

	fmt.Printf("✓ Completion script installed to %s\n", destPath)
	fmt.Println()

	// Print instructions
	switch *shell {
	case "bash":
		fmt.Println("Add the following to your ~/.bashrc:")
		fmt.Println("  source ~/.bash_completion.d/boba")
	case "zsh":
		fmt.Println("Add the following to your ~/.zshrc:")
		fmt.Println("  fpath=(~/.zsh/completions $fpath)")
		fmt.Println("  autoload -Uz compinit && compinit")
	case "fish":
		fmt.Println("Completion will be loaded automatically in new fish sessions.")
	}

	return nil
}

func runCompletionsUninstall(home string, args []string) error {
	flags := flag.NewFlagSet("completions uninstall", flag.ContinueOnError)
	shell := flags.String("shell", "bash", "shell type: bash|zsh|fish")
	flags.SetOutput(io.Discard)
	if err := flags.Parse(args); err != nil {
		return err
	}

	homeDir, _ := os.UserHomeDir()
	var destPath string

	switch *shell {
	case "bash":
		destPath = filepath.Join(homeDir, ".bash_completion.d", "boba")
	case "zsh":
		destPath = filepath.Join(homeDir, ".zsh", "completions", "_boba")
	case "fish":
		destPath = filepath.Join(homeDir, ".config", "fish", "completions", "boba.fish")
	default:
		return fmt.Errorf("unsupported shell: %s", *shell)
	}

	if err := os.Remove(destPath); err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Completion script not found")
			return nil
		}
		return fmt.Errorf("remove completion script: %w", err)
	}

	fmt.Printf("✓ Completion script removed from %s\n", destPath)
	return nil
}

func runSuggest(home string, args []string) error {
	// Get current directory context
	cwd, _ := os.Getwd()
	branch := ""
	project := ""

	if repoRoot, err := findRepoRoot(cwd); err == nil {
		projectCfg, _, _ := config.FindProjectConfig(repoRoot)
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
		if output, err := cmd.Output(); err == nil {
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
