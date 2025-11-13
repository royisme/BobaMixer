package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/vantagecraft-dev/bobamixer/internal/store/config"
	"github.com/vantagecraft-dev/bobamixer/internal/store/sqlite"
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
	path := filepath.Join(home, "active_profile")
	if err := os.WriteFile(path, []byte(prof.Key), 0o600); err != nil {
		return err
	}
	fmt.Printf("active profile set to %s (%s)\n", prof.Key, prof.Model)
	return nil
}

func runStats(home string, args []string) error {
	flags := flag.NewFlagSet("stats", flag.ContinueOnError)
	today := flags.Bool("today", false, "show today's totals")
	flags.SetOutput(io.Discard)
	if err := flags.Parse(args); err != nil {
		return err
	}
	if !*today {
		return errors.New("stats currently supports --today only")
	}
	dbPath := filepath.Join(home, "usage.db")
	db, err := sqlite.Open(dbPath)
	if err != nil {
		return err
	}
	totalTokens, err := db.QueryRow("SELECT COALESCE(SUM(input_tokens + output_tokens),0) FROM usage_records WHERE date(ts,'unixepoch') = date('now');")
	if err != nil {
		return err
	}
	totalCost, err := db.QueryRow("SELECT COALESCE(SUM(input_cost + output_cost),0) FROM usage_records WHERE date(ts,'unixepoch') = date('now');")
	if err != nil {
		return err
	}
	fmt.Printf("Today tokens: %s | cost: $%s\n", strings.TrimSpace(totalTokens), strings.TrimSpace(totalCost))
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
