// Package settings provides configuration management for BobaMixer.
// It handles initialization, saving, and loading of user settings.
package settings

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

//go:embed templates/profiles.yaml.tmpl
var profilesTemplate string

//go:embed templates/secrets.yaml.tmpl
var secretsTemplate string

//go:embed templates/routes.yaml.tmpl
var routesTemplate string

//go:embed templates/pricing.yaml.tmpl
var pricingTemplate string

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

	// Initialize config files from embedded templates
	// Templates are embedded from configs/templates/*.tmpl files
	files := map[string]struct {
		content string
		mode    os.FileMode
	}{
		"profiles.yaml": {
			content: profilesTemplate,
			mode:    0644,
		},
		"routes.yaml": {
			content: routesTemplate,
			mode:    0644,
		},
		"pricing.yaml": {
			content: pricingTemplate,
			mode:    0644,
		},
		"secrets.yaml": {
			content: secretsTemplate,
			mode:    0600,
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
	if s.Explore.Rate < 0 || s.Explore.Rate > 1 {
		return fmt.Errorf("explore rate must be between 0 and 1, got %f", s.Explore.Rate)
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
