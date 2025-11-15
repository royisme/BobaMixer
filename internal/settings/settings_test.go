package settings_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/royisme/bobamixer/internal/settings"
)

func TestInitHome(t *testing.T) {
	t.Run("creates home directory and four config files", func(t *testing.T) {
		// Given: empty temp directory
		tmpDir := t.TempDir()
		home := filepath.Join(tmpDir, ".boba")

		// When: InitHome is called
		err := settings.InitHome(home)

		// Then: no error and files exist with correct permissions
		if err != nil {
			t.Fatalf("InitHome failed: %v", err)
		}

		// Check directory exists
		info, err := os.Stat(home)
		if err != nil {
			t.Fatalf("home directory not created: %v", err)
		}
		if !info.IsDir() {
			t.Fatalf("home is not a directory")
		}

		// Check four config files exist
		expectedFiles := []string{
			"profiles.yaml",
			"routes.yaml",
			"pricing.yaml",
			"secrets.yaml",
		}

		for _, fname := range expectedFiles {
			fpath := filepath.Join(home, fname)
			info, err := os.Stat(fpath)
			if err != nil {
				t.Errorf("file %s not created: %v", fname, err)
				continue
			}
			if info.IsDir() {
				t.Errorf("%s is a directory, expected file", fname)
			}

			// Check secrets.yaml has 0600 permissions
			if fname == "secrets.yaml" {
				mode := info.Mode().Perm()
				if mode != 0600 {
					t.Errorf("secrets.yaml has mode %o, want 0600", mode)
				}
			}
		}
	})

	t.Run("idempotent - does not fail if already initialized", func(t *testing.T) {
		// Given: already initialized directory
		tmpDir := t.TempDir()
		home := filepath.Join(tmpDir, ".boba")
		if err := settings.InitHome(home); err != nil {
			t.Fatalf("first InitHome failed: %v", err)
		}

		// When: InitHome is called again
		err := settings.InitHome(home)

		// Then: no error
		if err != nil {
			t.Fatalf("second InitHome failed: %v", err)
		}
	})
}

//nolint:gocyclo // Test function with multiple subtests is acceptable
func TestSaveAndLoad(t *testing.T) {
	t.Run("saves and loads settings correctly", func(t *testing.T) {
		// Given: initialized home directory
		tmpDir := t.TempDir()
		home := filepath.Join(tmpDir, ".boba")
		if err := settings.InitHome(home); err != nil {
			t.Fatalf("InitHome failed: %v", err)
		}

		ctx := context.Background()

		// When: Save settings with Mode=observer
		s := settings.Settings{
			Mode:  settings.ModeObserver,
			Theme: "dark",
			Explore: settings.ExploreSettings{
				Enabled: true,
				Rate:    0.03,
			},
		}
		err := settings.Save(ctx, home, s)

		// Then: no error
		if err != nil {
			t.Fatalf("Save failed: %v", err)
		}

		// When: Load settings
		loaded, err := settings.Load(ctx, home)

		// Then: loaded settings match saved settings
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		if loaded.Mode != settings.ModeObserver {
			t.Errorf("Mode = %s, want %s", loaded.Mode, settings.ModeObserver)
		}
		if loaded.Theme != "dark" {
			t.Errorf("Theme = %s, want dark", loaded.Theme)
		}
		if !loaded.Explore.Enabled {
			t.Errorf("Explore.Enabled = false, want true")
		}
		if loaded.Explore.Rate != 0.03 {
			t.Errorf("Explore.Rate = %f, want 0.03", loaded.Explore.Rate)
		}
	})

	t.Run("returns default settings when file does not exist", func(t *testing.T) {
		// Given: home directory without settings file
		tmpDir := t.TempDir()
		home := filepath.Join(tmpDir, ".boba")
		if err := os.MkdirAll(home, 0755); err != nil { //nolint:gosec // G301: test file permissions
			t.Fatalf("failed to create home: %v", err)
		}

		ctx := context.Background()

		// When: Load settings
		loaded, err := settings.Load(ctx, home)

		// Then: returns default settings without error
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		if loaded.Mode != settings.ModeObserver {
			t.Errorf("default Mode = %s, want %s", loaded.Mode, settings.ModeObserver)
		}
	})

	t.Run("rejects invalid explore rate", func(t *testing.T) {
		tmpDir := t.TempDir()
		home := filepath.Join(tmpDir, ".boba")
		if err := settings.InitHome(home); err != nil {
			t.Fatalf("InitHome failed: %v", err)
		}
		ctx := context.Background()
		err := settings.Save(ctx, home, settings.Settings{
			Mode: settings.ModeObserver,
			Explore: settings.ExploreSettings{
				Enabled: true,
				Rate:    1.5,
			},
		})
		if err == nil {
			t.Fatal("expected error for invalid explore rate")
		}
	})

	t.Run("supports all three modes", func(t *testing.T) {
		tmpDir := t.TempDir()
		home := filepath.Join(tmpDir, ".boba")
		if err := settings.InitHome(home); err != nil {
			t.Fatalf("InitHome failed: %v", err)
		}

		ctx := context.Background()

		modes := []settings.Mode{
			settings.ModeObserver,
			settings.ModeSuggest,
			settings.ModeApply,
		}

		for _, mode := range modes {
			s := settings.Settings{Mode: mode}
			if err := settings.Save(ctx, home, s); err != nil {
				t.Errorf("Save mode %s failed: %v", mode, err)
			}

			loaded, err := settings.Load(ctx, home)
			if err != nil {
				t.Errorf("Load mode %s failed: %v", mode, err)
			}
			if loaded.Mode != mode {
				t.Errorf("Mode = %s, want %s", loaded.Mode, mode)
			}
		}
	})
}

func TestSettingsPermissions(t *testing.T) {
	t.Run("settings file has secure permissions", func(t *testing.T) {
		// Given: initialized home
		tmpDir := t.TempDir()
		home := filepath.Join(tmpDir, ".boba")
		if err := settings.InitHome(home); err != nil {
			t.Fatalf("InitHome failed: %v", err)
		}

		ctx := context.Background()

		// When: Save settings
		s := settings.Settings{Mode: settings.ModeObserver}
		if err := settings.Save(ctx, home, s); err != nil {
			t.Fatalf("Save failed: %v", err)
		}

		// Then: settings file has 0600 permissions
		settingsPath := filepath.Join(home, "settings.yaml")
		info, err := os.Stat(settingsPath)
		if err != nil {
			t.Fatalf("settings file not found: %v", err)
		}

		mode := info.Mode().Perm()
		if mode != 0600 {
			t.Errorf("settings.yaml has mode %o, want 0600", mode)
		}
	})
}
