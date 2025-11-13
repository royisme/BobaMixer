package cli

import (
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

	"github.com/royisme/bobamixer/internal/domain/budget"
	"github.com/royisme/bobamixer/internal/domain/stats"
	"github.com/royisme/bobamixer/internal/domain/suggestions"
	"github.com/royisme/bobamixer/internal/store/config"
	"github.com/royisme/bobamixer/internal/store/sqlite"
)

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
	if len(args) == 0 {
		printUsage()
		return nil
	}
	switch args[0] {
	case "ls":
		return runLS(home, args[1:])
	case "use":
		return runUse(home, args[1:])
	case "stats":
		return runStats(home, args[1:])
	case "edit":
		return runEdit(home, args[1:])
	case "doctor":
		return runDoctor(home, args[1:])
	case "budget":
		return runBudget(home, args[1:])
	case "action":
		return runAction(home, args[1:])
	case "report":
		return runReport(home, args[1:])
	default:
		return fmt.Errorf("unknown command %s", args[0])
	}
}

func printUsage() {
	fmt.Println("BobaMixer CLI")
	fmt.Println("Usage:")
	fmt.Println("  boba ls --profiles")
	fmt.Println("  boba use <profile>")
	fmt.Println("  boba stats --today")
	fmt.Println("  boba edit <profiles|routes|pricing|secrets>")
	fmt.Println("  boba doctor")
	fmt.Println("  boba budget [--status]")
	fmt.Println("  boba action [--auto]")
	fmt.Println("  boba report [--format json|csv]")
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
		sessions, _ := db.QueryRow("SELECT COUNT(DISTINCT session_id) FROM usage_records WHERE date(ts,'unixepoch') = date('now');")

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
		cmd := exec.Command(editor, path)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		return cmd.Run()
	}
	fmt.Println(path)
	return nil
}

func runDoctor(home string, args []string) error {
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
			fmt.Println("✓ usage.db: OK")
			// Try to query version
			if version, err := db.QueryInt("PRAGMA user_version;"); err == nil {
				fmt.Printf("  Schema version: %d\n", version)
			}
		}
	} else {
		fmt.Println("⚠ usage.db: will be created on first use")
	}

	fmt.Println()
	fmt.Println("Diagnosis complete.")
	return nil
}

func runBudget(home string, args []string) error {
	flags := flag.NewFlagSet("budget", flag.ContinueOnError)
	status := flags.Bool("status", false, "show budget status")
	flags.SetOutput(io.Discard)
	if err := flags.Parse(args); err != nil {
		return err
	}

	if *status {
		// TODO: Implement budget tracking
		fmt.Println("Budget Status")
		fmt.Println("=============")
		fmt.Println()
		fmt.Println("Daily budget: Not configured")
		fmt.Println("Hard cap: Not configured")
		fmt.Println()
		fmt.Println("To configure budgets, edit .boba-project.yaml in your project root")
		return nil
	}

	return errors.New("budget: specify --status")
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
	suggs, _ := suggestions.NewEngine(db).GenerateSuggestions(*days)

	fileName := *output
	if fileName == "" {
		fileName = fmt.Sprintf("bobamixer-report-%s.%s", time.Now().Format("20060102-1504"), *format)
	}
	if !filepath.IsAbs(fileName) {
		fileName = filepath.Join(home, fileName)
	}
	if err := os.MkdirAll(filepath.Dir(fileName), 0o755); err != nil {
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
		f, err := os.Create(fileName)
		if err != nil {
			return err
		}
		defer f.Close()
		writer := csv.NewWriter(f)
		defer writer.Flush()
		writer.Write([]string{"date", "tokens", "cost", "sessions"})
		for _, dp := range trend.DataPoints {
			writer.Write([]string{
				dp.Date,
				fmt.Sprintf("%d", dp.Tokens),
				fmt.Sprintf("%.4f", dp.Cost),
				fmt.Sprintf("%d", dp.Count),
			})
		}
		if err := writer.Error(); err != nil {
			return err
		}
	}

	fmt.Printf("Report exported to %s\n", fileName)
	return nil
}
