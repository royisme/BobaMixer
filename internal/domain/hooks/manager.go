// Package hooks manages git hooks for tracking repository events.
package hooks

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Manager installs git hooks and stores recorded events under ~/.boba.
type Manager struct {
	home string
}

// NewManager creates a hook manager.
func NewManager(home string) *Manager {
	return &Manager{home: home}
}

// Install writes helper scripts into repo/.git/hooks.
func (m *Manager) Install(repo string) error {
	repoPath, err := filepath.Abs(repo)
	if err != nil {
		return err
	}
	if _, err := os.Stat(filepath.Join(repoPath, ".git")); err != nil {
		return fmt.Errorf("not a git repository: %w", err)
	}
	hooksDir := filepath.Join(repoPath, ".git", "hooks")
	if err := os.MkdirAll(hooksDir, 0o750); err != nil {
		return err
	}
	helper := filepath.Join(hooksDir, "boba-hook")
	// #nosec G306 -- git hook script needs executable permissions
	if err := os.WriteFile(helper, []byte(m.helperScript(repoPath)), 0o750); err != nil {
		return err
	}
	for _, name := range []string{"post-checkout", "post-merge", "post-commit"} {
		path := filepath.Join(hooksDir, name)
		if existsAndNotOwned(path) {
			return fmt.Errorf("hook %s already exists", name)
		}
		content := fmt.Sprintf("#!/bin/sh\nexec \"%s\" %s \"$@\"\n", helper, name)
		// #nosec G306 -- git hook script needs executable permissions
		if err := os.WriteFile(path, []byte(content), 0o750); err != nil {
			return err
		}
	}
	return nil
}

// Remove deletes helper hooks if installed.
func (m *Manager) Remove(repo string) error {
	repoPath, err := filepath.Abs(repo)
	if err != nil {
		return err
	}
	hooksDir := filepath.Join(repoPath, ".git", "hooks")
	helper := filepath.Join(hooksDir, "boba-hook")
	//nolint:errcheck // Best effort cleanup
	// #nosec G104 -- best effort cleanup, error can be ignored
	os.Remove(helper)
	for _, name := range []string{"post-checkout", "post-merge", "post-commit"} {
		path := filepath.Join(hooksDir, name)
		// #nosec G304 -- path is constructed from git hooks directory and fixed hook names
		data, err := os.ReadFile(path)
		if err == nil && strings.Contains(string(data), "boba-hook") {
			//nolint:errcheck // Best effort cleanup
			// #nosec G104 -- best effort cleanup, error can be ignored
			os.Remove(path)
		}
	}
	return nil
}

// Record persists a git hook event.
func (m *Manager) Record(event, repo, branch string) error {
	if event == "" {
		return errors.New("event required")
	}
	if repo == "" {
		return errors.New("repo required")
	}
	payload := map[string]string{
		"event":  event,
		"repo":   repo,
		"branch": branch,
		"ts":     time.Now().Format(time.RFC3339),
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	destDir := filepath.Join(m.home, "git-hooks")
	if err := os.MkdirAll(destDir, 0o700); err != nil {
		return err
	}
	file := filepath.Join(destDir, fmt.Sprintf("%s.jsonl", slugify(repo)))
	// #nosec G304 -- file path is constructed from safe directory and slugified repo name
	f, err := os.OpenFile(file, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer func() {
		//nolint:errcheck // Best effort cleanup
		// #nosec G104 -- best effort cleanup, error can be ignored
		f.Close()
	}()
	if _, err := f.Write(append(data, '\n')); err != nil {
		return err
	}
	return nil
}

func (m *Manager) helperScript(repo string) string {
	return fmt.Sprintf(`#!/bin/sh
set -e
repo="%s"
branch=$(git -C "$repo" rev-parse --abbrev-ref HEAD 2>/dev/null || true)
if ! command -v boba >/dev/null 2>&1; then
  exit 0
fi
boba hooks track --event "$1" --repo "$repo" --branch "$branch" >/dev/null 2>&1 || true

# Show profile suggestion after checkout
if [ "$1" = "post-checkout" ]; then
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo "📍 Branch changed to: $branch"
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  boba suggest 2>/dev/null || true
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
fi
`, repo)
}

func existsAndNotOwned(path string) bool {
	// #nosec G304 -- path is git hook path being validated
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return !strings.Contains(string(data), "boba-hook")
}

func slugify(repo string) string {
	base := filepath.Base(repo)
	cleaned := make([]rune, 0, len(base))
	for _, r := range base {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			cleaned = append(cleaned, r)
		} else {
			cleaned = append(cleaned, '-')
		}
	}
	return strings.Trim(strings.ToLower(string(cleaned)), "-")
}
