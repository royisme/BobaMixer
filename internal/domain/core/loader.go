package core

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// LoadProviders loads and validates providers from providers.yaml
func LoadProviders(home string) (*ProvidersConfig, error) {
	path := filepath.Join(home, "providers.yaml")

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty config if file doesn't exist
			return &ProvidersConfig{Version: 1, Providers: []Provider{}}, nil
		}
		return nil, fmt.Errorf("failed to read providers.yaml: %w", err)
	}

	var config ProvidersConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse providers.yaml: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid providers.yaml: %w", err)
	}

	return &config, nil
}

// SaveProviders saves providers to providers.yaml
func SaveProviders(home string, config *ProvidersConfig) error {
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	path := filepath.Join(home, "providers.yaml")

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal providers: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write providers.yaml: %w", err)
	}

	return nil
}

// LoadTools loads and validates tools from tools.yaml
func LoadTools(home string) (*ToolsConfig, error) {
	path := filepath.Join(home, "tools.yaml")

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty config if file doesn't exist
			return &ToolsConfig{Version: 1, Tools: []Tool{}}, nil
		}
		return nil, fmt.Errorf("failed to read tools.yaml: %w", err)
	}

	var config ToolsConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse tools.yaml: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid tools.yaml: %w", err)
	}

	return &config, nil
}

// SaveTools saves tools to tools.yaml
func SaveTools(home string, config *ToolsConfig) error {
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	path := filepath.Join(home, "tools.yaml")

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal tools: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write tools.yaml: %w", err)
	}

	return nil
}

// LoadBindings loads and validates bindings from bindings.yaml
func LoadBindings(home string) (*BindingsConfig, error) {
	path := filepath.Join(home, "bindings.yaml")

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty config if file doesn't exist
			return &BindingsConfig{Version: 1, Bindings: []Binding{}}, nil
		}
		return nil, fmt.Errorf("failed to read bindings.yaml: %w", err)
	}

	var config BindingsConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse bindings.yaml: %w", err)
	}

	return &config, nil
}

// SaveBindings saves bindings to bindings.yaml
func SaveBindings(home string, config *BindingsConfig) error {
	path := filepath.Join(home, "bindings.yaml")

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal bindings: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write bindings.yaml: %w", err)
	}

	return nil
}

// LoadSecrets loads secrets from secrets.yaml
func LoadSecrets(home string) (*SecretsConfig, error) {
	path := filepath.Join(home, "secrets.yaml")

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty config if file doesn't exist
			return &SecretsConfig{Version: 1, Secrets: make(map[string]Secret)}, nil
		}
		return nil, fmt.Errorf("failed to read secrets.yaml: %w", err)
	}

	var config SecretsConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse secrets.yaml: %w", err)
	}

	// Set ProviderID for each secret based on the map key
	for providerID, secret := range config.Secrets {
		secret.ProviderID = providerID
		config.Secrets[providerID] = secret
	}

	return &config, nil
}

// SaveSecrets saves secrets to secrets.yaml with proper permissions
func SaveSecrets(home string, config *SecretsConfig) error {
	path := filepath.Join(home, "secrets.yaml")

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal secrets: %w", err)
	}

	// Use 0600 permissions for secrets file
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write secrets.yaml: %w", err)
	}

	return nil
}

// ResolveAPIKey retrieves the API key for a provider from either environment or secrets
func ResolveAPIKey(provider *Provider, secrets *SecretsConfig) (string, error) {
	switch provider.APIKey.Source {
	case APIKeySourceEnv:
		// Try to get from environment
		key := os.Getenv(provider.APIKey.EnvVar)
		if key == "" {
			return "", fmt.Errorf("%w: environment variable %s not set for provider %s",
				ErrMissingAPIKey, provider.APIKey.EnvVar, provider.ID)
		}
		return key, nil

	case APIKeySourceSecrets:
		// Get from secrets.yaml
		secret, ok := secrets.Secrets[provider.ID]
		if !ok || secret.APIKey == "" {
			return "", fmt.Errorf("%w: no secret found for provider %s",
				ErrMissingAPIKey, provider.ID)
		}
		return secret.APIKey, nil

	default:
		return "", fmt.Errorf("unknown API key source: %s", provider.APIKey.Source)
	}
}

// LoadAll loads all configuration files and returns them
func LoadAll(home string) (*ProvidersConfig, *ToolsConfig, *BindingsConfig, *SecretsConfig, error) {
	providers, err := LoadProviders(home)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to load providers: %w", err)
	}

	tools, err := LoadTools(home)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to load tools: %w", err)
	}

	bindings, err := LoadBindings(home)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to load bindings: %w", err)
	}

	// Validate bindings reference valid tools and providers
	if err := bindings.Validate(providers, tools); err != nil {
		return nil, nil, nil, nil, fmt.Errorf("invalid bindings: %w", err)
	}

	secrets, err := LoadSecrets(home)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to load secrets: %w", err)
	}

	return providers, tools, bindings, secrets, nil
}

// InitDefaultConfigs creates default configuration files if they don't exist
func InitDefaultConfigs(home string) error {
	// Ensure .boba directory exists
	if err := os.MkdirAll(home, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create default providers.yaml if it doesn't exist
	providersPath := filepath.Join(home, "providers.yaml")
	if _, err := os.Stat(providersPath); os.IsNotExist(err) {
		defaultProviders := &ProvidersConfig{
			Version: 1,
			Providers: []Provider{
				{
					ID:          "claude-anthropic-official",
					Kind:        ProviderKindAnthropic,
					DisplayName: "Anthropic (Official)",
					BaseURL:     "https://api.anthropic.com",
					APIKey: APIKeyConfig{
						Source: APIKeySourceEnv,
						EnvVar: "ANTHROPIC_API_KEY",
					},
					DefaultModel: "claude-3-5-sonnet-20241022",
					Enabled:      true,
				},
			},
		}
		if err := SaveProviders(home, defaultProviders); err != nil {
			return err
		}
	}

	// Create default tools.yaml if it doesn't exist
	toolsPath := filepath.Join(home, "tools.yaml")
	if _, err := os.Stat(toolsPath); os.IsNotExist(err) {
		defaultTools := &ToolsConfig{
			Version: 1,
			Tools: []Tool{
				{
					ID:          "claude",
					Name:        "Claude Code CLI",
					Exec:        "claude",
					Kind:        ToolKindClaude,
					ConfigType:  ConfigTypeClaudeSettingsJSON,
					ConfigPath:  "~/.claude/settings.json",
					Description: "Claude Code CLI for AI-assisted coding",
				},
			},
		}
		if err := SaveTools(home, defaultTools); err != nil {
			return err
		}
	}

	// Create default bindings.yaml if it doesn't exist
	bindingsPath := filepath.Join(home, "bindings.yaml")
	if _, err := os.Stat(bindingsPath); os.IsNotExist(err) {
		defaultBindings := &BindingsConfig{
			Version: 1,
			Bindings: []Binding{
				{
					ToolID:     "claude",
					ProviderID: "claude-anthropic-official",
					UseProxy:   false,
					Options:    BindingOptions{},
				},
			},
		}
		if err := SaveBindings(home, defaultBindings); err != nil {
			return err
		}
	}

	// Create empty secrets.yaml if it doesn't exist
	secretsPath := filepath.Join(home, "secrets.yaml")
	if _, err := os.Stat(secretsPath); os.IsNotExist(err) {
		defaultSecrets := &SecretsConfig{
			Version: 1,
			Secrets: make(map[string]Secret),
		}
		if err := SaveSecrets(home, defaultSecrets); err != nil {
			return err
		}
	}

	return nil
}
