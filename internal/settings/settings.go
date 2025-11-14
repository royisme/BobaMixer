// Package settings provides configuration management for BobaMixer.
// It handles initialization, saving, and loading of user settings.
package settings

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Mode represents the operation mode of BobaMixer.
type Mode string

const (
	// ModeObserver only observes and records usage without making suggestions
	ModeObserver Mode = "observer"
	// ModeSuggest provides suggestions but requires explicit user confirmation
	ModeSuggest Mode = "suggest"
	// ModeApply automatically applies optimizations (requires Apply=true flag)
	ModeApply Mode = "apply"
)

// ExploreSettings configures the exploration behavior.
type ExploreSettings struct {
	Enabled bool    `yaml:"enabled"`
	Rate    float64 `yaml:"rate"` // epsilon for epsilon-greedy exploration
}

// Settings represents the user's configuration.
type Settings struct {
	Mode    Mode            `yaml:"mode"`
	Theme   string          `yaml:"theme,omitempty"`
	Explore ExploreSettings `yaml:"explore"`
}

const (
	settingsFilename = "settings.yaml"
	defaultTheme     = "auto"
)

// InitHome creates the ~/.boba directory and initializes placeholder config files.
// It is idempotent - safe to call multiple times.
func InitHome(home string) error {
	// Create home directory
	if err := os.MkdirAll(home, 0755); err != nil { //nolint:gosec // G301: 0755 is intentional for .boba directory
		return fmt.Errorf("create home directory: %w", err)
	}

	// Initialize four config files with placeholders
	files := map[string]struct {
		content string
		mode    os.FileMode
	}{
		"profiles.yaml": {
			content: `# BobaMixer Profiles Configuration
# Define your AI provider profiles here
# Example:
# work-heavy:
#   adapter: http
#   provider: anthropic
#   model: claude-sonnet-4
#   endpoint: https://api.anthropic.com/v1/messages
#   cost_per_1k:
#     input: 0.003
#     output: 0.015
#   env:
#     - ANTHROPIC_API_KEY=secret://anthropic_key
`,
			mode: 0644,
		},
		"routes.yaml": {
			content: `# BobaMixer Routing Rules
# Define context-aware routing rules
# Example:
# rules:
#   - id: quick-tasks
#     if: "ctx_chars<1000"
#     use: quick-tasks
#     explain: "Small context, use faster model"
#
# explore:
#   enabled: true
#   rate: 0.03
`,
			mode: 0644,
		},
		"pricing.yaml": {
			content: `# BobaMixer Pricing Configuration
# Local pricing fallback (used when remote pricing is unavailable)
# Example:
# models:
#   anthropic/claude-sonnet-4:
#     input_per_1k: 0.003
#     output_per_1k: 0.015
`,
			mode: 0644,
		},
		"secrets.yaml": {
			content: `# BobaMixer Secrets
# Store API keys and sensitive values here
# This file must have 0600 permissions
# Example:
# anthropic_key: sk-ant-...
# openai_key: sk-...
`,
			mode: 0600,
		},
	}

	for fname, fdata := range files {
		fpath := filepath.Join(home, fname)

		// Check if file already exists (idempotent)
		if _, err := os.Stat(fpath); err == nil {
			// File exists, skip creation
			continue
		} else if !errors.Is(err, os.ErrNotExist) {
			// Other error
			return fmt.Errorf("stat %s: %w", fname, err)
		}

		// Create file with content
		if err := os.WriteFile(fpath, []byte(fdata.content), fdata.mode); err != nil {
			return fmt.Errorf("write %s: %w", fname, err)
		}
	}

	return nil
}

// Save persists settings to the settings file with secure permissions.
func Save(ctx context.Context, home string, s Settings) error {
	// Validate mode
	if s.Mode != ModeObserver && s.Mode != ModeSuggest && s.Mode != ModeApply {
		return fmt.Errorf("invalid mode: %s (must be observer, suggest, or apply)", s.Mode)
	}

	// Set defaults
	if s.Theme == "" {
		s.Theme = defaultTheme
	}

	// Marshal to YAML
	data, err := yaml.Marshal(s)
	if err != nil {
		return fmt.Errorf("marshal settings: %w", err)
	}

	// Write with secure permissions (0600)
	settingsPath := filepath.Join(home, settingsFilename)
	if err := os.WriteFile(settingsPath, data, 0600); err != nil {
		return fmt.Errorf("write settings: %w", err)
	}

	return nil
}

// Load reads settings from the settings file.
// If the file does not exist, it returns default settings without error.
func Load(ctx context.Context, home string) (Settings, error) {
	settingsPath := filepath.Join(home, settingsFilename)

	// Check if file exists
	data, err := os.ReadFile(settingsPath) //nolint:gosec // G304: path is constructed from trusted home directory
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// Return default settings
			return DefaultSettings(), nil
		}
		return Settings{}, fmt.Errorf("read settings: %w", err)
	}

	// Parse YAML
	var s Settings
	if err := yaml.Unmarshal(data, &s); err != nil {
		return Settings{}, fmt.Errorf("unmarshal settings: %w", err)
	}

	// Validate and set defaults
	if s.Mode == "" {
		s.Mode = ModeObserver
	}
	if s.Theme == "" {
		s.Theme = defaultTheme
	}

	return s, nil
}

// DefaultSettings returns the default settings configuration.
func DefaultSettings() Settings {
	return Settings{
		Mode:  ModeObserver,
		Theme: defaultTheme,
		Explore: ExploreSettings{
			Enabled: true,
			Rate:    0.03,
		},
	}
}
