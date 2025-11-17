// Package core defines the fundamental domain types for BobaMixer's control plane.
// These types represent the core concepts: Provider, Tool, Binding, and Secrets.
package core

import (
	"errors"
	"fmt"
)

// ProviderKind represents the type of AI provider
type ProviderKind string

// AI provider kinds
const (
	ProviderKindOpenAI              ProviderKind = "openai"
	ProviderKindAnthropic           ProviderKind = "anthropic"
	ProviderKindGemini              ProviderKind = "gemini"
	ProviderKindOpenAICompatible    ProviderKind = "openai-compatible"
	ProviderKindAnthropicCompatible ProviderKind = "anthropic-compatible"
)

// APIKeySource indicates where the API key should be retrieved from
type APIKeySource string

// API key source types
const (
	APIKeySourceEnv     APIKeySource = "env"     // From environment variable
	APIKeySourceSecrets APIKeySource = "secrets" // From secrets.yaml
)

// APIKeyConfig describes how to retrieve an API key for a provider
type APIKeyConfig struct {
	Source APIKeySource `yaml:"source"`            // Where to get the key from
	EnvVar string       `yaml:"env_var,omitempty"` // Environment variable name (if source=env)
}

// Provider represents an AI service provider (e.g., OpenAI, Anthropic, Z.AI)
type Provider struct {
	ID           string         `yaml:"id"`                 // Unique identifier (e.g., "claude-anthropic-official")
	Kind         ProviderKind   `yaml:"kind"`               // Provider type
	DisplayName  string         `yaml:"display_name"`       // Human-readable name
	BaseURL      string         `yaml:"base_url"`           // API endpoint
	APIKey       APIKeyConfig   `yaml:"api_key"`            // How to get the API key
	DefaultModel string         `yaml:"default_model"`      // Default model to use
	Enabled      bool           `yaml:"enabled"`            // Whether this provider is active
	Metadata     map[string]any `yaml:"metadata,omitempty"` // Additional provider-specific metadata
}

// ToolKind represents the type of CLI tool
type ToolKind string

// CLI tool kinds
const (
	ToolKindClaude ToolKind = "claude"
	ToolKindCodex  ToolKind = "codex"
	ToolKindGemini ToolKind = "gemini"
)

// ConfigType indicates how the tool's configuration should be managed
type ConfigType string

// Tool configuration types
const (
	ConfigTypeClaudeSettingsJSON ConfigType = "claude-settings-json" // ~/.claude/settings.json
	ConfigTypeCodexConfigTOML    ConfigType = "codex-config-toml"    // ~/.codex/config.toml
	ConfigTypeGeminiSettingsJSON ConfigType = "gemini-settings-json" // ~/.gemini/settings.json
)

// Tool represents a local CLI tool (e.g., claude, codex, gemini)
type Tool struct {
	ID          string     `yaml:"id"`                    // Unique identifier (e.g., "claude")
	Name        string     `yaml:"name"`                  // Human-readable name
	Exec        string     `yaml:"exec"`                  // Command to execute (e.g., "claude")
	Kind        ToolKind   `yaml:"kind"`                  // Tool type
	ConfigType  ConfigType `yaml:"config_type"`           // How to manage config
	ConfigPath  string     `yaml:"config_path"`           // Path to config file
	Description string     `yaml:"description,omitempty"` // Optional description
}

// BindingOptions contains tool-specific configuration options for a binding
type BindingOptions struct {
	// Model mapping for Claude (e.g., opus -> glm-4.6)
	ModelMapping map[string]string `yaml:"model_mapping,omitempty"`

	// Explicit model override
	Model string `yaml:"model,omitempty"`

	// Additional custom options
	Custom map[string]any `yaml:"custom,omitempty"`
}

// Binding represents the connection between a Tool and a Provider
type Binding struct {
	ToolID     string         `yaml:"tool_id"`           // Tool this binding applies to
	ProviderID string         `yaml:"provider_id"`       // Provider to use for this tool
	UseProxy   bool           `yaml:"use_proxy"`         // Whether to route through local proxy
	Options    BindingOptions `yaml:"options,omitempty"` // Tool-specific options
}

// Secret represents an API key or other sensitive credential
type Secret struct {
	ProviderID string            `yaml:"-"`                  // Provider this secret belongs to (not in YAML)
	APIKey     string            `yaml:"api_key"`            // The actual API key
	Metadata   map[string]string `yaml:"metadata,omitempty"` // Additional metadata
}

// ProvidersConfig is the root structure for providers.yaml
type ProvidersConfig struct {
	Version   int        `yaml:"version"`
	Providers []Provider `yaml:"providers"`
}

// ToolsConfig is the root structure for tools.yaml
type ToolsConfig struct {
	Version int    `yaml:"version"`
	Tools   []Tool `yaml:"tools"`
}

// BindingsConfig is the root structure for bindings.yaml
type BindingsConfig struct {
	Version  int       `yaml:"version"`
	Bindings []Binding `yaml:"bindings"`
}

// SecretsConfig is the root structure for secrets.yaml
type SecretsConfig struct {
	Version int               `yaml:"version"`
	Secrets map[string]Secret `yaml:"secrets"` // provider_id -> Secret
}

// Validation errors
var (
	ErrProviderNotFound = errors.New("provider not found")
	ErrToolNotFound     = errors.New("tool not found")
	ErrBindingNotFound  = errors.New("binding not found")
	ErrInvalidKind      = errors.New("invalid provider/tool kind")
	ErrMissingAPIKey    = errors.New("API key not configured")
	ErrDuplicateID      = errors.New("duplicate ID")
)

// Helper methods

// IsValid checks if a Provider has all required fields
func (p *Provider) IsValid() error {
	if p.ID == "" {
		return fmt.Errorf("provider ID is required")
	}
	if p.Kind == "" {
		return fmt.Errorf("provider kind is required for %s", p.ID)
	}
	if p.BaseURL == "" {
		return fmt.Errorf("base_url is required for provider %s", p.ID)
	}
	if p.APIKey.Source == "" {
		return fmt.Errorf("api_key.source is required for provider %s", p.ID)
	}
	if p.APIKey.Source == APIKeySourceEnv && p.APIKey.EnvVar == "" {
		return fmt.Errorf("api_key.env_var is required when source=env for provider %s", p.ID)
	}
	return nil
}

// IsValid checks if a Tool has all required fields
func (t *Tool) IsValid() error {
	if t.ID == "" {
		return fmt.Errorf("tool ID is required")
	}
	if t.Exec == "" {
		return fmt.Errorf("exec is required for tool %s", t.ID)
	}
	if t.Kind == "" {
		return fmt.Errorf("kind is required for tool %s", t.ID)
	}
	if t.ConfigType == "" {
		return fmt.Errorf("config_type is required for tool %s", t.ID)
	}
	return nil
}

// IsValid checks if a Binding has all required fields
func (b *Binding) IsValid() error {
	if b.ToolID == "" {
		return fmt.Errorf("tool_id is required")
	}
	if b.ProviderID == "" {
		return fmt.Errorf("provider_id is required")
	}
	return nil
}

// FindProvider searches for a provider by ID
func (pc *ProvidersConfig) FindProvider(id string) (*Provider, error) {
	for i := range pc.Providers {
		if pc.Providers[i].ID == id {
			return &pc.Providers[i], nil
		}
	}
	return nil, fmt.Errorf("%w: %s", ErrProviderNotFound, id)
}

// FindTool searches for a tool by ID
func (tc *ToolsConfig) FindTool(id string) (*Tool, error) {
	for i := range tc.Tools {
		if tc.Tools[i].ID == id {
			return &tc.Tools[i], nil
		}
	}
	return nil, fmt.Errorf("%w: %s", ErrToolNotFound, id)
}

// FindBinding searches for a binding by tool ID
func (bc *BindingsConfig) FindBinding(toolID string) (*Binding, error) {
	for i := range bc.Bindings {
		if bc.Bindings[i].ToolID == toolID {
			return &bc.Bindings[i], nil
		}
	}
	return nil, fmt.Errorf("%w for tool: %s", ErrBindingNotFound, toolID)
}

// Validate checks if all providers in the config are valid
func (pc *ProvidersConfig) Validate() error {
	seen := make(map[string]bool)
	for i, provider := range pc.Providers {
		if err := provider.IsValid(); err != nil {
			return fmt.Errorf("provider[%d]: %w", i, err)
		}
		if seen[provider.ID] {
			return fmt.Errorf("%w: provider ID %s", ErrDuplicateID, provider.ID)
		}
		seen[provider.ID] = true
	}
	return nil
}

// Validate checks if all tools in the config are valid
func (tc *ToolsConfig) Validate() error {
	seen := make(map[string]bool)
	for i, tool := range tc.Tools {
		if err := tool.IsValid(); err != nil {
			return fmt.Errorf("tool[%d]: %w", i, err)
		}
		if seen[tool.ID] {
			return fmt.Errorf("%w: tool ID %s", ErrDuplicateID, tool.ID)
		}
		seen[tool.ID] = true
	}
	return nil
}

// Validate checks if all bindings reference valid tools and providers
func (bc *BindingsConfig) Validate(providers *ProvidersConfig, tools *ToolsConfig) error {
	for i, binding := range bc.Bindings {
		if err := binding.IsValid(); err != nil {
			return fmt.Errorf("binding[%d]: %w", i, err)
		}

		// Check if tool exists
		if _, err := tools.FindTool(binding.ToolID); err != nil {
			return fmt.Errorf("binding[%d]: %w", i, err)
		}

		// Check if provider exists
		if _, err := providers.FindProvider(binding.ProviderID); err != nil {
			return fmt.Errorf("binding[%d]: %w", i, err)
		}
	}
	return nil
}
