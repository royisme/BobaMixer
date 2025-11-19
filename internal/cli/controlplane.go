package cli

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"

	"github.com/royisme/bobamixer/internal/domain/core"
	"github.com/royisme/bobamixer/internal/domain/pricing"
	"github.com/royisme/bobamixer/internal/logging"
	"github.com/royisme/bobamixer/internal/proxy"
	"github.com/royisme/bobamixer/internal/runner"
	"github.com/royisme/bobamixer/internal/store/config"
	"github.com/royisme/bobamixer/internal/ui/keys"
)

// runProviders lists all configured providers
func runProviders(home string, _ []string) error {
	logging.Info("Running providers command")

	providers, err := core.LoadProviders(home)
	if err != nil {
		return fmt.Errorf("failed to load providers: %w", err)
	}

	if len(providers.Providers) == 0 {
		fmt.Println("No providers configured.")
		fmt.Println("\nTo add a provider, edit ~/.boba/providers.yaml")
		return nil
	}

	secrets, err := core.LoadSecrets(home)
	if err != nil {
		return fmt.Errorf("failed to load secrets: %w", err)
	}

	// Define styles
	var (
		headerStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("99")).
				Padding(0, 1)

		cellStyle = lipgloss.NewStyle().
				Padding(0, 1)

		checkStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))  // Green
		crossStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196")) // Red
		envStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("208")) // Orange
	)

	var rows [][]string

	for _, provider := range providers.Providers {
		// Determine key status
		var keyStatus string
		if _, err := core.ResolveAPIKey(&provider, secrets); err == nil {
			if provider.APIKey.Source == core.APIKeySourceEnv {
				keyStatus = envStyle.Render("✓ env")
			} else {
				keyStatus = checkStyle.Render("✓ secrets")
			}
		} else {
			keyStatus = crossStyle.Render("✗")
		}

		// Enabled status
		var enabledStatus string
		if provider.Enabled {
			enabledStatus = checkStyle.Render("yes")
		} else {
			enabledStatus = crossStyle.Render("no")
		}

		// Truncate base URL if too long
		baseURL := provider.BaseURL
		if len(baseURL) > 35 {
			baseURL = baseURL[:32] + "..."
		}

		rows = append(rows, []string{
			provider.ID,
			string(provider.Kind),
			provider.DisplayName,
			baseURL,
			keyStatus,
			enabledStatus,
		})
	}

	t := table.New().
		Border(lipgloss.HiddenBorder()).
		Headers("ID", "TYPE", "NAME", "BASE URL", "KEY", "ENABLED").
		Rows(rows...).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == 0 {
				return headerStyle
			}
			return cellStyle
		})

	fmt.Println(t)
	fmt.Println()
	fmt.Println("✓ = Configured   ✗ = Missing   env = From environment   secrets = From secrets.yaml")

	return nil
}

// runTools lists all configured tools and their detection status
func runTools(home string, _ []string) error {
	logging.Info("Running tools command")

	tools, err := core.LoadTools(home)
	if err != nil {
		return fmt.Errorf("failed to load tools: %w", err)
	}

	if len(tools.Tools) == 0 {
		fmt.Println("No tools configured.")
		fmt.Println("\nTo add a tool, edit ~/.boba/tools.yaml")
		return nil
	}

	bindings, err := core.LoadBindings(home)
	if err != nil {
		return fmt.Errorf("failed to load bindings: %w", err)
	}

	// Define styles
	var (
		headerStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("99")).
				Padding(0, 1)

		cellStyle = lipgloss.NewStyle().
				Padding(0, 1)

		checkStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))  // Green
		crossStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196")) // Red
		faintStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240")) // Grey
	)

	var rows [][]string

	for _, tool := range tools.Tools {
		// Check if tool executable exists in PATH
		var status string
		if _, err := exec.LookPath(tool.Exec); err != nil {
			status = crossStyle.Render("✗ not found")
		} else {
			status = checkStyle.Render("✓ ready")
		}

		// Find binding
		boundTo := faintStyle.Render("(not bound)")
		if binding, err := bindings.FindBinding(tool.ID); err == nil {
			boundTo = binding.ProviderID
		}

		rows = append(rows, []string{
			tool.ID,
			tool.Name,
			tool.Exec,
			status,
			boundTo,
		})
	}

	t := table.New().
		Border(lipgloss.HiddenBorder()).
		Headers("ID", "NAME", "EXEC", "STATUS", "BOUND TO").
		Rows(rows...).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == 0 {
				return headerStyle
			}
			return cellStyle
		})

	fmt.Println(t)

	return nil
}

// runBind creates or updates a binding between a tool and a provider
//
//nolint:gocyclo // Command logic requires multiple validation steps
func runBind(home string, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: boba bind <tool> <provider> [--proxy=on|off]")
	}

	toolID := args[0]
	providerID := args[1]

	// Parse optional proxy flag
	useProxy := false
	if len(args) >= 3 {
		proxyArg := args[2]
		if strings.HasPrefix(proxyArg, "--proxy=") {
			proxyValue := strings.TrimPrefix(proxyArg, "--proxy=")
			if proxyValue == "on" || proxyValue == "true" {
				useProxy = true
			} else if proxyValue != "off" && proxyValue != "false" {
				return fmt.Errorf("invalid proxy value: %s (use on/off)", proxyValue)
			}
		}
	}

	logging.Info("Running bind command",
		logging.String("tool", toolID),
		logging.String("provider", providerID),
		logging.Bool("proxy", useProxy))

	// Load configs
	providers, err := core.LoadProviders(home)
	if err != nil {
		return fmt.Errorf("failed to load providers: %w", err)
	}

	tools, err := core.LoadTools(home)
	if err != nil {
		return fmt.Errorf("failed to load tools: %w", err)
	}

	bindings, err := core.LoadBindings(home)
	if err != nil {
		return fmt.Errorf("failed to load bindings: %w", err)
	}

	// Validate tool exists
	tool, err := tools.FindTool(toolID)
	if err != nil {
		return fmt.Errorf("tool not found: %s", toolID)
	}

	// Validate provider exists
	provider, err := providers.FindProvider(providerID)
	if err != nil {
		return fmt.Errorf("provider not found: %s", providerID)
	}

	// Check if binding already exists
	existingBinding, err := bindings.FindBinding(toolID)
	if err != nil {
		// Binding doesn't exist, will create new one below
		existingBinding = nil
	}
	if existingBinding != nil {
		// Update existing binding
		existingBinding.ProviderID = providerID
		existingBinding.UseProxy = useProxy
		fmt.Printf("Updated binding: %s → %s\n", tool.Name, provider.DisplayName)
	} else {
		// Create new binding
		newBinding := core.Binding{
			ToolID:     toolID,
			ProviderID: providerID,
			UseProxy:   useProxy,
			Options:    core.BindingOptions{},
		}
		bindings.Bindings = append(bindings.Bindings, newBinding)
		fmt.Printf("Created binding: %s → %s\n", tool.Name, provider.DisplayName)
	}

	// Save bindings
	if err := core.SaveBindings(home, bindings); err != nil {
		return fmt.Errorf("failed to save bindings: %w", err)
	}

	if useProxy {
		fmt.Println("Proxy: enabled")
	} else {
		fmt.Println("Proxy: disabled")
	}

	return nil
}

// runDoctorV2 runs diagnostics for the control plane configuration
//
//nolint:gocyclo // Comprehensive diagnostics require checking multiple components
func runDoctorV2(home string, args []string) error {
	logging.Info("Running doctor (control plane)")

	// Parse flags
	checkPricing := false
	for _, arg := range args {
		if arg == "--pricing" {
			checkPricing = true
			break
		}
	}

	// If --pricing flag is set, only check pricing
	if checkPricing {
		return runDoctorPricing(home)
	}

	fmt.Println("BobaMixer Control Plane Diagnostics")
	fmt.Println("═══════════════════════════════════")
	fmt.Println()

	hasErrors := false
	hasWarnings := false

	// Check providers
	fmt.Println("📋 Providers")
	fmt.Println("───────────")
	providers, err := core.LoadProviders(home)
	if err != nil {
		fmt.Printf("  %s Failed to load providers.yaml: %v\n", statusError, err)
		hasErrors = true
	} else if len(providers.Providers) == 0 {
		fmt.Printf("  %s No providers configured\n", statusWarning)
		hasWarnings = true
	} else {
		fmt.Printf("  %s Found %d provider(s)\n", statusOK, len(providers.Providers))

		// Check each provider's API key
		secrets, err := core.LoadSecrets(home)
		if err != nil {
			logging.Warn("Failed to load secrets", logging.String("error", err.Error()))
			secrets = &core.SecretsConfig{} // Use empty secrets
		}
		for _, provider := range providers.Providers {
			if !provider.Enabled {
				continue
			}

			if _, err := core.ResolveAPIKey(&provider, secrets); err != nil {
				fmt.Printf("  %s %s: %v\n", statusWarning, provider.DisplayName, err)
				hasWarnings = true
			} else {
				fmt.Printf("  %s %s: API key configured\n", statusOK, provider.DisplayName)
			}
		}
	}
	fmt.Println()

	// Check tools
	fmt.Println("🔧 Tools")
	fmt.Println("────────")
	tools, err := core.LoadTools(home)
	if err != nil {
		fmt.Printf("  %s Failed to load tools.yaml: %v\n", statusError, err)
		hasErrors = true
	} else if len(tools.Tools) == 0 {
		fmt.Printf("  %s No tools configured\n", statusWarning)
		hasWarnings = true
	} else {
		fmt.Printf("  %s Found %d tool(s)\n", statusOK, len(tools.Tools))

		// Check if tools are in PATH
		for _, tool := range tools.Tools {
			if _, err := exec.LookPath(tool.Exec); err != nil {
				fmt.Printf("  %s %s (%s): not found in PATH\n", statusWarning, tool.Name, tool.Exec)
				hasWarnings = true
			} else {
				fmt.Printf("  %s %s (%s): found in PATH\n", statusOK, tool.Name, tool.Exec)
			}
		}
	}
	fmt.Println()

	// Check bindings
	fmt.Println("🔗 Bindings")
	fmt.Println("───────────")
	bindings, err := core.LoadBindings(home)
	if err != nil {
		fmt.Printf("  %s Failed to load bindings.yaml: %v\n", statusError, err)
		hasErrors = true
	} else if len(bindings.Bindings) == 0 {
		fmt.Printf("  %s No bindings configured\n", statusWarning)
		hasWarnings = true
	} else {
		fmt.Printf("  %s Found %d binding(s)\n", statusOK, len(bindings.Bindings))

		// Validate each binding
		if providers != nil && tools != nil {
			for _, binding := range bindings.Bindings {
				tool, toolErr := tools.FindTool(binding.ToolID)
				provider, provErr := providers.FindProvider(binding.ProviderID)

				if toolErr != nil {
					fmt.Printf("  %s Binding references unknown tool: %s\n", statusError, binding.ToolID)
					hasErrors = true
				} else if provErr != nil {
					fmt.Printf("  %s Binding references unknown provider: %s\n", statusError, binding.ProviderID)
					hasErrors = true
				} else {
					proxyStatus := ""
					if binding.UseProxy {
						proxyStatus = " (via proxy)"
					}
					fmt.Printf("  %s %s → %s%s\n", statusOK, tool.Name, provider.DisplayName, proxyStatus)
				}
			}
		}
	}
	fmt.Println()

	// Summary
	fmt.Println("Summary")
	fmt.Println("───────")
	if hasErrors {
		fmt.Printf("%s Configuration has errors. Please fix the issues above.\n", statusError)
		return fmt.Errorf("configuration errors detected")
	} else if hasWarnings {
		fmt.Printf("%s Configuration is valid but has warnings.\n", statusWarning)
	} else {
		fmt.Printf("%s All checks passed! Your configuration is healthy.\n", statusOK)
	}

	return nil
}

// runRun executes a CLI tool with injected configuration
func runRun(home string, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: boba run <tool> [args...]")
	}

	toolID := args[0]
	toolArgs := args[1:]

	logging.Info("Running tool", logging.String("tool", toolID))

	// Load configurations
	providers, tools, bindings, secrets, err := core.LoadAll(home)
	if err != nil {
		return fmt.Errorf("failed to load configurations: %w", err)
	}

	// Find the tool
	tool, err := tools.FindTool(toolID)
	if err != nil {
		return fmt.Errorf("tool not found: %s\nRun 'boba tools' to list available tools", toolID)
	}

	// Find the binding
	binding, err := bindings.FindBinding(toolID)
	if err != nil {
		return fmt.Errorf("tool %s is not bound to any provider\nRun 'boba bind %s <provider>' to create a binding", toolID, toolID)
	}

	// Find the provider
	provider, err := providers.FindProvider(binding.ProviderID)
	if err != nil {
		return fmt.Errorf("provider %s not found\nRun 'boba providers' to list available providers", binding.ProviderID)
	}

	// Create run context
	ctx := &runner.RunContext{
		Home:     home,
		Tool:     tool,
		Binding:  binding,
		Provider: provider,
		Secrets:  secrets,
		Args:     toolArgs,
	}

	// Run the tool
	if err := runner.Run(ctx); err != nil {
		return fmt.Errorf("failed to run %s: %w", tool.Name, err)
	}

	return nil
}

// runProxy handles proxy subcommands
func runProxy(home string, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("proxy subcommand required: serve, status, stop")
	}

	switch args[0] {
	case "serve":
		return runProxyServe(home, args[1:])
	case "status":
		return runProxyStatus(home, args[1:])
	case "stop":
		return runProxyStop(home, args[1:])
	default:
		return fmt.Errorf("unknown proxy subcommand: %s", args[0])
	}
}

// runProxyServe starts the proxy server
func runProxyServe(home string, _ []string) error {
	logging.Info("Starting proxy server")

	dbPath := filepath.Join(home, "usage.db")
	server, err := proxy.NewServer(proxy.DefaultAddr, dbPath)
	if err != nil {
		return fmt.Errorf("failed to create proxy server: %w", err)
	}

	if err := server.Start(); err != nil {
		return fmt.Errorf("failed to start proxy server: %w", err)
	}

	fmt.Printf("✓ Proxy server started on %s\n", server.Addr())
	fmt.Printf("\nPress %s to stop...\n", keys.CtrlC)

	// Wait for interrupt signal
	select {}
}

// runProxyStatus shows the proxy server status
func runProxyStatus(_ string, _ []string) error {
	logging.Info("Checking proxy status")

	// Try to connect to the proxy to check if it's running
	addr := proxy.DefaultAddr
	ctx := context.Background()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+addr+"/health", nil)
	if err != nil {
		fmt.Println("Proxy Status: ❌ Not running")
		fmt.Printf("Address: %s\n", addr)
		fmt.Println("\nTo start: boba proxy serve")
		return nil
	}

	client := &http.Client{
		Timeout: 1 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Proxy Status: ❌ Not running")
		fmt.Printf("Address: %s\n", addr)
		fmt.Println("\nTo start: boba proxy serve")
		return nil
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			logging.Warn("failed to close response body", logging.Err(cerr))
		}
	}()

	fmt.Println("Proxy Status: ✅ Running")
	fmt.Printf("Address: %s\n", addr)
	fmt.Println("\nEndpoints:")
	fmt.Println("  - http://127.0.0.1:7777/openai/v1/*")
	fmt.Println("  - http://127.0.0.1:7777/anthropic/v1/*")

	return nil
}

// runProxyStop stops the proxy server
func runProxyStop(_ string, _ []string) error {
	logging.Info("Stopping proxy server")

	// For now, just inform the user
	// In a production implementation, we'd use a PID file or similar
	fmt.Printf("To stop the proxy server, press %s in the terminal where it's running\n", keys.CtrlC)
	fmt.Println("Or use: killall -SIGTERM boba")

	return nil
}

// runDoctorPricing runs diagnostics specifically for pricing configuration
//
//nolint:gocyclo // Comprehensive pricing diagnostics require checking multiple aspects
func runDoctorPricing(home string) error {
	logging.Info("Running doctor --pricing")

	fmt.Println("BobaMixer Pricing Diagnostics")
	fmt.Println("══════════════════════════════")
	fmt.Println()

	hasErrors := false
	hasWarnings := false

	// Check pricing.yaml configuration
	fmt.Println("💰 Pricing Configuration")
	fmt.Println("────────────────────────")

	pricingCfg, err := config.LoadPricing(home)
	if err != nil {
		fmt.Printf("  %s Failed to load pricing.yaml: %v\n", statusError, err)
		hasErrors = true
	} else {
		fmt.Printf("  %s pricing.yaml loaded successfully\n", statusOK)

		// Check if models are configured
		if len(pricingCfg.Models) == 0 {
			fmt.Printf("  %s No models configured in pricing.yaml\n", statusWarning)
			hasWarnings = true
		} else {
			fmt.Printf("  %s Found %d model(s) in pricing.yaml\n", statusOK, len(pricingCfg.Models))
		}

		// Check sources configuration
		if len(pricingCfg.Sources) == 0 {
			fmt.Printf("  %s No remote sources configured (using automatic OpenRouter fetch with cache)\n", statusWarning)
			hasWarnings = true
		} else {
			fmt.Printf("  %s Found %d pricing source(s)\n", statusOK, len(pricingCfg.Sources))

			// Verify each source
			for i, source := range pricingCfg.Sources {
				switch source.Type {
				case "http-json":
					if source.URL == "" {
						fmt.Printf("  %s Source #%d: missing URL\n", statusError, i+1)
						hasErrors = true
					} else {
						fmt.Printf("  %s Source #%d: %s (priority: %d)\n", statusOK, i+1, source.URL, source.Priority)
					}
				case "file":
					if source.Path == "" {
						fmt.Printf("  %s Source #%d: missing file path\n", statusError, i+1)
						hasErrors = true
					} else {
						fmt.Printf("  %s Source #%d: file %s (priority: %d)\n", statusOK, i+1, source.Path, source.Priority)
					}
				default:
					fmt.Printf("  %s Source #%d: unknown type '%s'\n", statusWarning, i+1, source.Type)
					hasWarnings = true
				}
			}
		}

		// Check refresh settings
		if pricingCfg.Refresh.IntervalHours > 0 {
			fmt.Printf("  %s Refresh interval: %d hours\n", statusOK, pricingCfg.Refresh.IntervalHours)
		} else {
			fmt.Printf("  %s Refresh interval: not configured (default 24h cache)\n", statusWarning)
			hasWarnings = true
		}

		if pricingCfg.Refresh.OnStartup {
			fmt.Printf("  %s Refresh on startup: enabled\n", statusOK)
		} else {
			fmt.Printf("  %s Refresh on startup: disabled (will reuse cache if fresh)\n", statusWarning)
		}
	}
	fmt.Println()

	loaderCfg := pricing.DefaultLoaderConfig()
	if pricingCfg != nil {
		loaderCfg.RefreshOnStartup = pricingCfg.Refresh.OnStartup
		if pricingCfg.Refresh.IntervalHours > 0 {
			loaderCfg.CacheTTLHours = pricingCfg.Refresh.IntervalHours
		}
	}
	loader := pricing.NewLoader(home, loaderCfg)

	// Try to load pricing data
	fmt.Println("🔍 Pricing Data Validation")
	fmt.Println("──────────────────────────")

	useLegacy := pricingCfg != nil && len(pricingCfg.Sources) > 0

	if useLegacy {
		fmt.Println("  Using pricing.yaml sources (legacy pipeline)")

		table, err := pricing.Load(home)
		if err != nil {
			fmt.Printf("  %s Failed to load pricing table: %v\n", statusError, err)
			hasErrors = true
		} else if len(table.Models) == 0 {
			fmt.Printf("  %s No pricing data available\n", statusWarning)
			fmt.Println("  ℹ️  Tip: Verify remote sources and pricing.yaml entries")
			hasWarnings = true
		} else {
			fmt.Printf("  %s Successfully loaded pricing for %d model(s)\n", statusOK, len(table.Models))
			printPricingSamples(table.Models)
		}
	} else {
		schema, err := loader.LoadWithFallback(context.Background())
		if err != nil {
			fmt.Printf("  %s Failed to load pricing table: %v\n", statusError, err)
			hasErrors = true
		} else {
			table := schema.ToLegacyTable()
			if len(table.Models) == 0 {
				fmt.Printf("  %s No pricing data available\n", statusWarning)
				fmt.Println("  ℹ️  Tip: Configure pricing.yaml or ensure OpenRouter access is available")
				hasWarnings = true
			} else {
				fmt.Printf("  %s Successfully loaded pricing for %d model(s)\n", statusOK, len(table.Models))
				printPricingSamples(table.Models)

				// Summarize source kinds from schema
				sourceCounts := make(map[string]int)
				for _, model := range schema.Models {
					sourceCounts[model.Source.Kind]++
				}
				if len(sourceCounts) > 0 {
					fmt.Println("  Sources:")
					for kind, count := range sourceCounts {
						fmt.Printf("    - %s: %d model(s)\n", kind, count)
					}
				}
			}
		}
	}
	fmt.Println()

	// Pricing cache status
	fmt.Println("📦 Pricing Cache")
	fmt.Println("────────────────")

	cacheFresh, cacheMeta, cacheErr := loader.GetCacheStatus()
	if cacheErr != nil {
		fmt.Printf("  %s Cache metadata unavailable: %v\n", statusWarning, cacheErr)
		hasWarnings = true
	} else {
		cacheAge := time.Since(cacheMeta.FetchedAt)
		timeLeft := time.Until(cacheMeta.ExpiresAt)

		if cacheFresh {
			fmt.Printf("  %s Cache is fresh (Source: %s)\n", statusOK, cacheMeta.SourceKind)
		} else {
			fmt.Printf("  %s Cache is stale (Source: %s)\n", statusWarning, cacheMeta.SourceKind)
			hasWarnings = true
		}

		fmt.Printf("  %s Fetched: %s ago\n", statusOK, formatDuration(cacheAge))
		fmt.Printf("  %s Expires in: %s\n", statusOK, formatDuration(timeLeft))
		fmt.Printf("  %s TTL (hours): %d\n", statusOK, cacheMeta.TTLHours)
	}
	fmt.Println()

	// Summary
	fmt.Println("Summary")
	fmt.Println("───────")
	if hasErrors {
		fmt.Printf("%s Pricing configuration has errors. Please fix the issues above.\n", statusError)
		return fmt.Errorf("pricing configuration errors detected")
	} else if hasWarnings {
		fmt.Printf("%s Pricing configuration is functional but has warnings.\n", statusWarning)
		fmt.Println("\nRecommendations:")
		fmt.Println("  1. Configure remote pricing sources in pricing.yaml")
		fmt.Println("  2. Enable automatic refresh (refresh.on_startup and refresh.interval_hours)")
		fmt.Println("  3. Run 'boba init' to regenerate default pricing configuration")
	} else {
		fmt.Printf("%s Pricing configuration is healthy!\n", statusOK)
	}

	return nil
}

// formatDuration renders a duration in a concise human-readable form
func formatDuration(d time.Duration) string {
	if d < 0 {
		return "expired"
	}

	switch {
	case d >= time.Hour:
		hours := int(d.Hours())
		minutes := int((d % time.Hour) / time.Minute)
		return fmt.Sprintf("%dh%dm", hours, minutes)
	case d >= time.Minute:
		minutes := int(d / time.Minute)
		seconds := int((d % time.Minute) / time.Second)
		return fmt.Sprintf("%dm%ds", minutes, seconds)
	default:
		return fmt.Sprintf("%ds", int(d/time.Second))
	}
}

func printPricingSamples(models map[string]pricing.ModelPrice) {
	sampleCount := 0
	for modelName, price := range models {
		if sampleCount >= 5 {
			fmt.Printf("  ... and %d more models\n", len(models)-5)
			break
		}
		fmt.Printf("    - %s: $%.4f/$%.4f per 1K tokens\n", modelName, price.InputPer1K, price.OutputPer1K)
		sampleCount++
	}
}
