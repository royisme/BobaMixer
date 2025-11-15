// Package cli provides the command-line interface for BobaMixer.
package cli

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/royisme/bobamixer/internal/adapters"
	"github.com/royisme/bobamixer/internal/domain/budget"
	"github.com/royisme/bobamixer/internal/domain/hooks"
	"github.com/royisme/bobamixer/internal/domain/pricing"
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
	fmt.Println("  boba init                                     Initialize ~/.boba with defaults")
	fmt.Println("  boba edit <profiles|routes|pricing|secrets>  Edit configuration files")
	fmt.Println("  boba doctor                                   Run diagnostics")
	fmt.Println("  boba doctor --pricing                         Run pricing validation")
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

//nolint:unparam // Keeping error return for consistency with other command handlers
func runDoctor(home string, args []string) error {
	// Parse flags
	fs := flag.NewFlagSet("doctor", flag.ContinueOnError)
	pricingFlag := fs.Bool("pricing", false, "Run detailed pricing validation")
	verboseFlag := fs.Bool("v", false, "Verbose output")
	if err := fs.Parse(args); err != nil {
		return err
	}

	// If --pricing flag is set, run detailed pricing diagnostics
	if *pricingFlag {
		return runDoctorPricing(home, *verboseFlag)
	}

	fmt.Println("BobaMixer Doctor")
	fmt.Println("================")
	fmt.Println()

	fmt.Printf("%s Home directory: %s\n", statusOK, home)
	if info, err := os.Stat(home); err == nil {
		fmt.Printf("  Permissions: %04o\n", info.Mode().Perm())
	}

	fmt.Println()
	fmt.Println("Configuration Files:")

	profsPath := filepath.Join(home, "profiles.yaml")
	var profs config.Profiles
	if _, err := os.Stat(profsPath); err == nil {
		loaded, err := config.LoadProfiles(home)
		if err != nil {
			fmt.Printf("%s profiles.yaml: invalid (%v)\n", statusError, err)
		} else {
			profs = loaded
			fmt.Printf("%s profiles.yaml: %d profiles\n", statusOK, len(profs))
		}
	} else {
		fmt.Printf("%s profiles.yaml: not found\n", statusError)
	}

	secretsPath := filepath.Join(home, "secrets.yaml")
	if _, err := os.Stat(secretsPath); err == nil {
		if err := config.ValidateSecretsPermissions(home); err != nil {
			fmt.Printf("%s secrets.yaml: %v\n", statusError, err)
		} else if info, statErr := os.Stat(secretsPath); statErr == nil {
			fmt.Printf("%s secrets.yaml: permissions OK (%04o)\n", statusOK, info.Mode().Perm())
		}
	} else {
		fmt.Printf("%s secrets.yaml: not found (run 'boba edit secrets' to add API keys)\n", statusWarning)
	}

	routesPath := filepath.Join(home, "routes.yaml")
	if _, err := os.Stat(routesPath); err == nil {
		routes, err := config.LoadRoutes(home)
		if err != nil {
			fmt.Printf("%s routes.yaml: invalid (%v)\n", statusError, err)
		} else {
			fmt.Printf("%s routes.yaml: %d rules, %d sub-agents\n", statusOK, len(routes.Rules), len(routes.SubAgents))
		}
	} else {
		fmt.Printf("%s routes.yaml: not found (optional)\n", statusWarning)
	}

	pricingPath := filepath.Join(home, "pricing.yaml")
	if _, err := os.Stat(pricingPath); err == nil {
		pricingCfg, err := config.LoadPricing(home)
		if err != nil {
			fmt.Printf("%s pricing.yaml: invalid (%v)\n", statusError, err)
		} else {
			fmt.Printf("%s pricing.yaml: %d models configured\n", statusOK, len(pricingCfg.Models))
		}
	} else {
		fmt.Printf("%s pricing.yaml: not found (optional)\n", statusWarning)
	}

	fmt.Println()
	diagnoseDatabase(home)

	fmt.Println()
	diagnosePricingSources(home)

	fmt.Println()
	diagnoseNetworkAndKeys(home, profs)

	fmt.Println()
	fmt.Println("Diagnosis complete.")
	fmt.Println()
	fmt.Println("Summary:")
	fmt.Printf("  Review any %s/%s entries above for actionable fixes.\n", statusError, statusWarning)
	return nil
}

func diagnoseDatabase(home string) {
	fmt.Println("Database Health:")
	dbPath := filepath.Join(home, "usage.db")
	if _, err := os.Stat(dbPath); err != nil {
		fmt.Printf("%s usage.db: will be created on first use\n", statusWarning)
		return
	}

	db, err := sqlite.Open(dbPath)
	if err != nil {
		fmt.Printf("%s usage.db: cannot open (%v)\n", statusError, err)
		fmt.Println("  Fix: remove corrupted usage.db or ensure sqlite3 is installed")
		return
	}

	version, err := db.QueryInt("PRAGMA user_version;")
	if err != nil {
		fmt.Printf("%s usage.db: cannot read schema version (%v)\n", statusError, err)
	} else {
		fmt.Printf("%s usage.db: schema v%d\n", statusOK, version)
	}

	walMode, err := db.QueryRow("PRAGMA journal_mode;")
	if err != nil {
		fmt.Printf("  %s Cannot check WAL mode: %v\n", statusWarning, err)
	} else if strings.EqualFold(strings.TrimSpace(walMode), "wal") {
		fmt.Printf("  %s WAL mode enabled\n", statusOK)
	} else {
		fmt.Printf("  %s WAL mode not enabled (current: %s)\n", statusWarning, strings.TrimSpace(walMode))
	}

	if _, err := db.QueryRow("SELECT COUNT(*) FROM sessions;"); err != nil {
		fmt.Printf("  %s Database read test failed: %v\n", statusWarning, err)
	} else {
		fmt.Printf("  %s Read test passed\n", statusOK)
	}

	if err := runDoctorWriteProbe(db); err != nil {
		fmt.Printf("  %s Write probe failed: %v\n", statusError, err)
		fmt.Println("    Fix: ensure ~/.boba is writable and sqlite3 supports WAL mode")
	} else {
		fmt.Printf("  %s Write probe passed\n", statusOK)
	}
}

func runDoctorWriteProbe(db *sqlite.DB) error {
	statements := []string{
		"CREATE TABLE IF NOT EXISTS __doctor_probe (id INTEGER PRIMARY KEY AUTOINCREMENT);",
		"INSERT INTO __doctor_probe DEFAULT VALUES;",
		"DELETE FROM __doctor_probe;",
		"DROP TABLE IF EXISTS __doctor_probe;",
	}
	for _, stmt := range statements {
		if err := db.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

func diagnosePricingSources(home string) {
	fmt.Println("Pricing Sources:")
	pricingCfg, err := config.LoadPricing(home)
	if err != nil {
		fmt.Printf("%s pricing.yaml: invalid (%v)\n", statusError, err)
		return
	}

	success := false
	for _, source := range pricingCfg.Sources {
		switch source.Type {
		case "http-json":
			count, err := doctorFetchPricingHTTP(source.URL)
			if err != nil {
				fmt.Printf("%s %s: %v\n", statusError, source.URL, err)
			} else {
				success = true
				fmt.Printf("%s %s reachable (%d models)\n", statusOK, source.URL, count)
			}
		case "file":
			path := doctorExpandHome(source.Path, home)
			count, err := doctorLoadPricingFile(path)
			if err != nil {
				fmt.Printf("%s %s: %v\n", statusError, path, err)
			} else {
				success = true
				fmt.Printf("%s %s loaded (%d models)\n", statusOK, path, count)
			}
		default:
			fmt.Printf("%s Unsupported pricing source type: %s\n", statusWarning, source.Type)
		}
	}

	cachePath := filepath.Join(home, "pricing.cache.json")
	if info, err := os.Stat(cachePath); err == nil {
		cacheAge := time.Since(info.ModTime())
		if cacheAge < 24*time.Hour {
			fmt.Printf("%s Pricing cache fresh (updated %s)\n", statusOK, info.ModTime().Format("2006-01-02 15:04"))
		} else {
			fmt.Printf("%s Pricing cache stale since %s (will refresh on next fetch)\n", statusWarning, info.ModTime().Format("2006-01-02 15:04"))
		}
		if !success {
			fmt.Println("  ↳ Falling back to cached pricing until remote fetch succeeds.")
		}
	} else {
		fmt.Printf("%s Pricing cache not found, will populate on next successful fetch\n", statusWarning)
	}

	if len(pricingCfg.Sources) == 0 {
		fmt.Printf("%s No remote pricing sources configured. Add an http-json source in pricing.yaml to enable auto refresh.\n", statusWarning)
	}
}

func doctorFetchPricingHTTP(url string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, fmt.Errorf("invalid url: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if isTimeoutErr(err) {
			return 0, fmt.Errorf("unreachable (timeout)")
		}
		return 0, fmt.Errorf("unreachable: %w", err)
	}
	defer func() {
		_ = resp.Body.Close() //nolint:errcheck,gosec
	}()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("HTTP %s", resp.Status)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return 0, err
	}

	var payload struct {
		Models map[string]interface{} `json:"models"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return 0, fmt.Errorf("format error: %w", err)
	}
	return len(payload.Models), nil
}

func doctorLoadPricingFile(path string) (int, error) {
	// #nosec G304 -- path is from pricing config file sources, validated by config loader
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	var payload struct {
		Models map[string]interface{} `json:"models"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return 0, err
	}
	return len(payload.Models), nil
}

func doctorExpandHome(path, home string) string {
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(home, path[2:])
	}
	return path
}

func diagnoseNetworkAndKeys(home string, profs config.Profiles) {
	fmt.Println("Network & API Keys:")
	if len(profs) == 0 {
		fmt.Printf("%s No profiles configured yet. Run 'boba edit profiles' to add one.\n", statusWarning)
		return
	}

	activeProfile := ""
	if ap, err := config.LoadActiveProfile(home); err == nil {
		activeProfile = ap
	}
	if activeProfile == "" {
		for key := range profs {
			activeProfile = key
			break
		}
	}

	prof, ok := profs[activeProfile]
	if !ok {
		fmt.Printf("%s Active profile not found in profiles.yaml\n", statusError)
		return
	}

	fmt.Printf("Testing profile: %s (%s)\n", prof.Key, prof.Provider)
	secrets, err := config.LoadSecrets(home)
	if err != nil {
		fmt.Printf("%s Cannot load secrets.yaml: %v\n", statusError, err)
		fmt.Println("  Fix: ensure secrets.yaml exists and is valid YAML")
		return
	}

	envVars := config.ResolveEnv(prof.Env, secrets)
	envMap := doctorEnvToMap(envVars)
	if !doctorHasAPIKey(envMap) {
		missing := doctorSecretPlaceholders(prof.Env)
		if len(missing) > 0 {
			fmt.Printf("%s API key missing. Expected secrets: %s\n", statusError, strings.Join(missing, ", "))
		} else {
			fmt.Printf("%s API key missing. Add provider credentials to secrets.yaml\n", statusError)
		}
		fmt.Println("  Fix: run 'boba edit secrets' and add the appropriate secret value.")
		return
	}
	fmt.Printf("%s API key detected\n", statusOK)

	if prof.Adapter != "http" {
		fmt.Printf("%s Adapter '%s' does not support HTTP diagnostics. Use 'boba call' to validate connectivity.\n", statusWarning, prof.Adapter)
		return
	}

	headers := doctorHeadersFromEnv(envVars)
	if headers == nil {
		fmt.Printf("%s Unable to derive HTTP headers from profile env. Ensure *_API_KEY entries are set.\n", statusError)
		return
	}

	status, detail := probeHTTPEndpoint(prof.Endpoint, headers, buildDoctorProbePayload(prof))
	fmt.Printf("%s Network probe: %s\n", status, detail)
	if status == statusError {
		fmt.Println("  Fix: verify the endpoint URL and credentials, then retry 'boba doctor'.")
	}
}

func doctorEnvToMap(env []string) map[string]string {
	result := make(map[string]string, len(env))
	for _, entry := range env {
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}
	return result
}

func doctorHasAPIKey(env map[string]string) bool {
	for key, val := range env {
		if val == "" {
			continue
		}
		lower := strings.ToLower(key)
		if strings.Contains(lower, "api_key") || strings.Contains(lower, "token") || strings.Contains(lower, "secret") {
			return true
		}
	}
	return false
}

func doctorSecretPlaceholders(env map[string]string) []string {
	var names []string
	for _, val := range env {
		if strings.HasPrefix(val, "secret://") {
			names = append(names, strings.TrimPrefix(val, "secret://"))
		}
	}
	sort.Strings(names)
	return names
}

func doctorHeadersFromEnv(env []string) map[string]string {
	headers := map[string]string{"Content-Type": "application/json"}
	for _, entry := range env {
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) != 2 || parts[1] == "" {
			continue
		}
		switch parts[0] {
		case "ANTHROPIC_API_KEY":
			headers["x-api-key"] = parts[1]
			headers["anthropic-version"] = "2023-06-01"
		case "OPENAI_API_KEY", "OPENROUTER_API_KEY", "AZURE_OPENAI_API_KEY":
			headers["Authorization"] = "Bearer " + parts[1]
		}
	}
	if len(headers) == 1 {
		return nil
	}
	return headers
}

func probeHTTPEndpoint(endpoint string, headers map[string]string, payload []byte) (string, string) {
	if endpoint == "" {
		return statusError, "endpoint not configured"
	}
	status, detail, retry := sendHTTPProbe(endpoint, headers, http.MethodHead, nil)
	if retry {
		status, detail, _ = sendHTTPProbe(endpoint, headers, http.MethodPost, payload)
		return status, detail
	}
	return status, detail
}

func sendHTTPProbe(endpoint string, headers map[string]string, method string, body []byte) (string, string, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	var reader io.Reader
	if len(body) > 0 {
		reader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, reader)
	if err != nil {
		return statusError, fmt.Sprintf("invalid endpoint: %v", err), false
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if isTimeoutErr(err) {
			return statusError, "request timed out", false
		}
		return statusError, fmt.Sprintf("network error: %v", err), false
	}
	defer func() {
		_ = resp.Body.Close() //nolint:errcheck,gosec
	}()

	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		return statusOK, fmt.Sprintf("reachable (%s)", resp.Status), false
	}
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return statusError, "401 unauthorized - update API key in secrets.yaml", false
	case http.StatusForbidden:
		return statusError, "403 forbidden - check provider access/billing", false
	case http.StatusTooManyRequests:
		return statusWarning, "rate limited (429) - wait before retrying", false
	case http.StatusBadRequest:
		if method == http.MethodPost {
			return statusOK, "authentication OK (probe payload rejected as expected)", false
		}
	case http.StatusNotFound, http.StatusMethodNotAllowed:
		return statusWarning, fmt.Sprintf("%s - retrying with POST", resp.Status), true
	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return statusWarning, fmt.Sprintf("provider error (%s)", resp.Status), false
	}
	return statusWarning, fmt.Sprintf("unexpected response (%s)", resp.Status), false
}

func buildDoctorProbePayload(profile config.Profile) []byte {
	payload := map[string]interface{}{
		"model":       profile.Model,
		"max_tokens":  1,
		"temperature": 0,
		"messages": []map[string]string{
			{"role": "user", "content": "ping"},
		},
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return []byte(`{"ping":"boba-doctor"}`)
	}
	return data
}

func isTimeoutErr(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}
	return false
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
	fmt.Println("Settings initialized in", filepath.Join(home, "settings.yaml"))
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

//nolint:gocyclo // Doctor command logic is complex but necessary for comprehensive diagnostics
func runDoctorPricing(home string, verbose bool) error {
	fmt.Println("BobaMixer Doctor - Pricing Validation")
	fmt.Println("=====================================")
	fmt.Println()

	// Load pricing configuration
	pricingCfg, err := config.LoadPricing(home)
	if err != nil {
		fmt.Printf("%s Failed to load pricing configuration: %v\n", statusError, err)
		return nil // Don't fail, just report
	}

	// Create loader
	loaderCfg := pricing.DefaultLoaderConfig()
	if pricingCfg != nil {
		loaderCfg.RefreshOnStartup = pricingCfg.Refresh.OnStartup
		if pricingCfg.Refresh.IntervalHours > 0 {
			loaderCfg.CacheTTLHours = pricingCfg.Refresh.IntervalHours
		}
	}

	loader := pricing.NewLoader(home, loaderCfg)

	// Check cache status
	fmt.Println("Cache Status:")
	isFresh, meta, err := loader.GetCacheStatus()
	if err != nil {
		fmt.Printf("%s Cache not found or invalid: %v\n", statusWarning, err)
	} else {
		if isFresh {
			fmt.Printf("%s Cache is fresh\n", statusOK)
		} else {
			fmt.Printf("%s Cache is expired\n", statusWarning)
		}
		fmt.Printf("  Source: %s\n", meta.SourceKind)
		fmt.Printf("  Fetched at: %s\n", meta.FetchedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("  Expires at: %s\n", meta.ExpiresAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("  TTL: %d hours\n", meta.TTLHours)
	}

	fmt.Println()

	// Load pricing schema
	fmt.Println("Loading Pricing Data:")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	schema, err := loader.LoadWithFallback(ctx)
	if err != nil {
		fmt.Printf("%s Failed to load pricing: %v\n", statusError, err)
		return nil
	}

	fmt.Printf("%s Successfully loaded %d models\n", statusOK, len(schema.Models))

	if len(schema.Models) == 0 {
		fmt.Printf("%s No pricing models found\n", statusWarning)
		fmt.Println("  Tip: Configure OpenRouter API or add pricing.vendor.json")
		return nil
	}

	// Group by provider
	providerCount := make(map[string]int)
	for _, model := range schema.Models {
		providerCount[model.Provider]++
	}

	fmt.Println("\nModels by Provider:")
	providers := make([]string, 0, len(providerCount))
	for provider := range providerCount {
		providers = append(providers, provider)
	}
	sort.Strings(providers)

	for _, provider := range providers {
		fmt.Printf("  %s: %d models\n", provider, providerCount[provider])
	}

	// Validate pricing
	fmt.Println("\nValidating Pricing:")
	validator := pricing.NewPricingValidator()
	warnings := validator.ValidateAgainstRefs(schema)

	if len(warnings) == 0 {
		fmt.Printf("%s No validation warnings\n", statusOK)
	} else {
		fmt.Printf("%s Found %d validation warnings\n", statusWarning, len(warnings))
		fmt.Println()

		// Group warnings by severity
		errors := 0
		for _, w := range warnings {
			if strings.Contains(w.Message, "zero or negative") || strings.Contains(w.Message, "missing") {
				errors++
			}
		}

		if errors > 0 {
			fmt.Printf("  Critical issues: %d\n", errors)
		}
		fmt.Printf("  Total warnings: %d\n", len(warnings))

		// Show warnings
		if verbose || errors > 0 {
			fmt.Println("\nDetailed Warnings:")
			fmt.Println(pricing.FormatWarnings(warnings))
		} else {
			fmt.Println("\nSample Warnings (first 5):")
			sampleWarnings := warnings
			if len(warnings) > 5 {
				sampleWarnings = warnings[:5]
			}
			fmt.Println(pricing.FormatWarnings(sampleWarnings))
			fmt.Printf("\n  Use 'boba doctor --pricing -v' to see all %d warnings\n", len(warnings))
		}
	}

	// Show official references
	fmt.Println("\nOfficial Pricing References:")
	fmt.Println(validator.GetReferenceList())

	fmt.Println("Pricing validation complete.")
	return nil
}
