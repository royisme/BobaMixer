package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/royisme/bobamixer/internal/domain/core"
	"golang.org/x/term"
)

// runSecrets handles the secrets subcommand
func runSecrets(home string, args []string) error {
	if len(args) == 0 {
		return runSecretsList(home, args)
	}

	switch args[0] {
	case "list", "ls":
		return runSecretsList(home, args[1:])
	case "set", "add":
		return runSecretsSet(home, args[1:])
	case "remove", "rm", "delete":
		return runSecretsRemove(home, args[1:])
	default:
		return fmt.Errorf("unknown secrets subcommand: %s\n\nUsage:\n  boba secrets list          List configured secrets\n  boba secrets set <provider>   Set API key for a provider\n  boba secrets remove <provider> Remove API key for a provider", args[0])
	}
}

// runSecretsList lists all configured secrets and their status
func runSecretsList(home string, _ []string) error {
	// Load providers to get the list of available providers
	providers, err := core.LoadProviders(home)
	if err != nil {
		return fmt.Errorf("failed to load providers: %w", err)
	}

	if len(providers.Providers) == 0 {
		fmt.Println("No providers configured.")
		fmt.Println("\nRun 'boba init' to initialize BobaMixer")
		return nil
	}

	// Load secrets
	secrets, err := core.LoadSecrets(home)
	if err != nil {
		return fmt.Errorf("failed to load secrets: %w", err)
	}

	fmt.Println("Configured Secrets")
	fmt.Println("==================")
	fmt.Println()

	// Print table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	if _, err := fmt.Fprintln(w, "PROVIDER\tSTATUS\tSOURCE"); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}
	if _, err := fmt.Fprintln(w, "────────────────────────────────────────────────────"); err != nil {
		return fmt.Errorf("failed to write separator: %w", err)
	}

	for _, provider := range providers.Providers {
		status := "✗ Missing"
		source := "-"

		// Check if key exists in environment
		if provider.APIKey.Source == core.APIKeySourceEnv && provider.APIKey.EnvVar != "" {
			if os.Getenv(provider.APIKey.EnvVar) != "" {
				status = "✓ Set"
				source = fmt.Sprintf("env (%s)", provider.APIKey.EnvVar)
			}
		}

		// Check if key exists in secrets.yaml
		if _, ok := secrets.Secrets[provider.ID]; ok {
			status = "✓ Set"
			source = "secrets.yaml"
		}

		if _, err := fmt.Fprintf(w, "%s\t%s\t%s\n", provider.ID, status, source); err != nil {
			return fmt.Errorf("failed to write row: %w", err)
		}
	}

	if err := w.Flush(); err != nil {
		return fmt.Errorf("failed to flush output: %w", err)
	}

	fmt.Println()
	fmt.Println("Legend:")
	fmt.Println("  ✓ Set     - API key configured")
	fmt.Println("  ✗ Missing - API key not found")
	fmt.Println()
	fmt.Println("Tip: Use 'boba secrets set <provider>' to add a missing key")

	return nil
}

// runSecretsSet sets an API key for a provider
func runSecretsSet(home string, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: boba secrets set <provider-id>\n\nExample:\n  boba secrets set claude-anthropic-official")
	}

	providerID := args[0]

	// Check if --key flag is provided (non-interactive mode)
	var apiKey string
	for i, arg := range args {
		if arg == "--key" && i+1 < len(args) {
			apiKey = args[i+1]
			break
		}
	}

	// Validate provider exists
	providers, err := core.LoadProviders(home)
	if err != nil {
		return fmt.Errorf("failed to load providers: %w", err)
	}

	var provider *core.Provider
	for i := range providers.Providers {
		if providers.Providers[i].ID == providerID {
			provider = &providers.Providers[i]
			break
		}
	}

	if provider == nil {
		fmt.Printf("Error: Provider '%s' not found\n\n", providerID)
		fmt.Println("Available providers:")

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		for _, p := range providers.Providers {
			if _, err := fmt.Fprintf(w, "  %s\t%s\n", p.ID, p.DisplayName); err != nil {
				return fmt.Errorf("failed to write provider: %w", err)
			}
		}
		if err := w.Flush(); err != nil {
			return fmt.Errorf("failed to flush output: %w", err)
		}

		fmt.Println("\nRun 'boba providers' to see more details")
		return fmt.Errorf("provider not found: %s", providerID)
	}

	// Interactive mode: prompt for API key
	if apiKey == "" {
		fmt.Printf("Enter API key for %s: ", provider.DisplayName)

		// Use terminal.ReadPassword for secure input
		keyBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			return fmt.Errorf("failed to read password: %w", err)
		}
		fmt.Println() // Newline after password input

		apiKey = string(keyBytes)
	}

	if apiKey == "" {
		return fmt.Errorf("API key cannot be empty")
	}

	// Load existing secrets
	secrets, err := core.LoadSecrets(home)
	if err != nil {
		return fmt.Errorf("failed to load secrets: %w", err)
	}

	if secrets.Secrets == nil {
		secrets.Secrets = make(map[string]core.Secret)
	}

	// Save the new secret
	secrets.Secrets[providerID] = core.Secret{
		APIKey: apiKey,
	}

	if err := core.SaveSecrets(home, secrets); err != nil {
		return fmt.Errorf("failed to save secrets: %w", err)
	}

	fmt.Println("✓ API key saved")
	fmt.Printf("  Provider: %s\n", provider.DisplayName)
	fmt.Printf("  Location: ~/.boba/secrets.yaml\n")
	fmt.Printf("  Permissions: 0600 (secure)\n")

	// Check if provider also uses env var
	if provider.APIKey.Source == core.APIKeySourceEnv && provider.APIKey.EnvVar != "" {
		if os.Getenv(provider.APIKey.EnvVar) == "" {
			fmt.Println()
			fmt.Printf("Note: This provider also supports environment variable %s\n", provider.APIKey.EnvVar)
			fmt.Printf("      BobaMixer will use secrets.yaml since env var is not set\n")
		}
	}

	return nil
}

// runSecretsRemove removes an API key for a provider
func runSecretsRemove(home string, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: boba secrets remove <provider-id>")
	}

	providerID := args[0]

	// Load secrets
	secrets, err := core.LoadSecrets(home)
	if err != nil {
		return fmt.Errorf("failed to load secrets: %w", err)
	}

	// Check if secret exists
	if _, ok := secrets.Secrets[providerID]; !ok {
		return fmt.Errorf("no secret found for provider: %s", providerID)
	}

	// Remove the secret
	delete(secrets.Secrets, providerID)

	// Save updated secrets
	if err := core.SaveSecrets(home, secrets); err != nil {
		return fmt.Errorf("failed to save secrets: %w", err)
	}

	fmt.Printf("✓ Removed API key for provider: %s\n", providerID)

	return nil
}
