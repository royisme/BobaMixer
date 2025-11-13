package version

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Manager coordinates semantic version management plus changelog updates.
type Manager struct {
	versionPath   string
	changelogPath string
}

// NewManager creates a manager rooted at repoRoot (expects VERSION file there).
func NewManager(repoRoot string) *Manager {
	return &Manager{
		versionPath:   filepath.Join(repoRoot, "VERSION"),
		changelogPath: filepath.Join(repoRoot, "CHANGELOG.md"),
	}
}

// Current returns the current semantic version string stored on disk.
func (m *Manager) Current() (string, error) {
	data, err := os.ReadFile(m.versionPath)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

// Plan calculates the bumped version without writing it back to disk.
func (m *Manager) Plan(part, prerelease string) (string, error) {
	current, err := m.Current()
	if err != nil {
		return "", err
	}
	return bumpVersion(current, part, prerelease)
}

// Bump updates the VERSION file and changelog entry.
func (m *Manager) Bump(part, prerelease, notes string) (string, error) {
	next, err := m.Plan(part, prerelease)
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(m.versionPath, []byte(next+"\n"), 0o644); err != nil {
		return "", err
	}
	if err := m.appendChangelog(next, notes); err != nil {
		return "", err
	}
	return next, nil
}

func (m *Manager) appendChangelog(version, notes string) error {
	entry := fmt.Sprintf("## %s - %s\n", version, time.Now().Format("2006-01-02"))
	if strings.TrimSpace(notes) == "" {
		notes = "- (notes pending)\n"
	} else {
		notes = strings.TrimSpace(notes)
		if !strings.HasPrefix(notes, "-") {
			notes = "- " + notes
		}
		notes = notes + "\n"
	}
	var header string = "# Changelog\n\n"
	var tail string
	if data, err := os.ReadFile(m.changelogPath); err == nil {
		existing := strings.TrimSpace(string(data))
		if existing != "" {
			if strings.HasPrefix(existing, "# Changelog") {
				parts := strings.SplitN(existing, "\n\n", 2)
				header = parts[0] + "\n\n"
				if len(parts) > 1 {
					tail = strings.TrimSpace(parts[1])
				}
			} else {
				tail = existing
			}
		}
	}
	var builder strings.Builder
	builder.WriteString(header)
	builder.WriteString(entry)
	builder.WriteString(notes)
	if tail != "" {
		builder.WriteString("\n")
		builder.WriteString(tail)
		builder.WriteString("\n")
	}
	return os.WriteFile(m.changelogPath, []byte(builder.String()), 0o644)
}

func bumpVersion(current, part, prerelease string) (string, error) {
	var major, minor, patch int
	if _, err := fmt.Sscanf(current, "%d.%d.%d", &major, &minor, &patch); err != nil {
		return "", fmt.Errorf("invalid version %q: %w", current, err)
	}
	switch part {
	case "major":
		major++
		minor = 0
		patch = 0
	case "minor":
		minor++
		patch = 0
	case "patch", "":
		patch++
	default:
		return "", errors.New("bump must be major, minor, or patch")
	}
	next := fmt.Sprintf("%d.%d.%d", major, minor, patch)
	if prerelease = strings.TrimSpace(prerelease); prerelease != "" {
		next = fmt.Sprintf("%s-%s", next, sanitize(prerelease))
	}
	return next, nil
}

func sanitize(raw string) string {
	filtered := make([]rune, 0, len(raw))
	for _, r := range raw {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '.' {
			filtered = append(filtered, r)
		}
	}
	if len(filtered) == 0 {
		return "rc"
	}
	return string(filtered)
}
