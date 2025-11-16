package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/tabwriter"

	"github.com/royisme/bobamixer/internal/domain/core"
	"github.com/royisme/bobamixer/internal/logging"
	"github.com/royisme/bobamixer/internal/runner"
)

// runProviders lists all configured providers
func runProviders(home string, args []string) error {
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

	// Print providers in a table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tTYPE\tNAME\tBASE URL\tKEY\tENABLED")
	fmt.Fprintln(w, "──────────────────────────────────────────────────────────────────────────────")

	for _, provider := range providers.Providers {
		// Determine key status
		keyStatus := "✗"
		if _, err := core.ResolveAPIKey(&provider, secrets); err == nil {
			if provider.APIKey.Source == core.APIKeySourceEnv {
				keyStatus = "✓ env"
			} else {
				keyStatus = "✓ secrets"
			}
		}

		// Enabled status
		enabledStatus := "yes"
		if !provider.Enabled {
			enabledStatus = "no"
		}

		// Truncate base URL if too long
		baseURL := provider.BaseURL
		if len(baseURL) > 35 {
			baseURL = baseURL[:32] + "..."
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			provider.ID,
			provider.Kind,
			provider.DisplayName,
			baseURL,
			keyStatus,
			enabledStatus,
		)
	}
	w.Flush()

	fmt.Println()
	fmt.Println("✓ = Configured   ✗ = Missing   env = From environment   secrets = From secrets.yaml")

	return nil
}

// runTools lists all configured tools and their detection status
func runTools(home string, args []string) error {
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

	// Print tools in a table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tEXEC\tSTATUS\tBOUND TO")
	fmt.Fprintln(w, "────────────────────────────────────────────────────────────────")

	for _, tool := range tools.Tools {
		// Check if tool executable exists in PATH
		status := "✓ ready"
		if _, err := exec.LookPath(tool.Exec); err != nil {
			status = "✗ not found"
		}

		// Find binding
		boundTo := "(not bound)"
		if binding, err := bindings.FindBinding(tool.ID); err == nil {
			boundTo = binding.ProviderID
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			tool.ID,
			tool.Name,
			tool.Exec,
			status,
			boundTo,
		)
	}
	w.Flush()

	return nil
}

// runBind creates or updates a binding between a tool and a provider
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
	existingBinding, _ := bindings.FindBinding(toolID)
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
func runDoctorV2(home string, args []string) error {
	logging.Info("Running doctor (control plane)")

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
		secrets, _ := core.LoadSecrets(home)
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
