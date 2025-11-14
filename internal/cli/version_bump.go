package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// runBump implements the bump command for version management
func runBump() error {
	// Simple manual parsing to get the bump type
	var bumpType string
	var dryRun bool
	var auto bool

	args := os.Args[2:] // Skip "boba" command

	for _, arg := range args {
		switch arg {
		case "--dry-run":
			dryRun = true
		case "--auto":
			auto = true
		default:
			if strings.HasPrefix(arg, "-") {
				return fmt.Errorf("unknown flag: %s", arg)
			}
			if bumpType != "" {
				return fmt.Errorf("multiple bump types specified: %s and %s", bumpType, arg)
			}
			bumpType = arg
		}
	}

	if auto {
		bumpType = "auto"
	}

	if bumpType == "" {
		return fmt.Errorf("usage: boba bump [major|minor|patch|auto] [--dry-run]")
	}

	// Validate bump type
	if bumpType != "auto" && bumpType != "major" && bumpType != "minor" && bumpType != "patch" {
		return fmt.Errorf("invalid bump type: %s (must be major, minor, patch, or auto)", bumpType)
	}

	// Get current version and determine next version
	currentVersion, err := getCurrentVersion()
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	nextVersion, bumpReason, err := calculateNextVersion(currentVersion, bumpType)
	if err != nil {
		return fmt.Errorf("failed to calculate next version: %w", err)
	}

	// Show version bump information
	fmt.Printf("Current version: %s\n", currentVersion)
	fmt.Printf("Next version:    %s\n", nextVersion)
	fmt.Printf("Bump type:       %s\n", bumpReason)

	if dryRun {
		fmt.Println("\n[Dry run] No changes made. Use without --dry-run to apply changes.")
		return nil
	}

	// Update version in go.mod
	if err := updateVersionInGoMod(nextVersion); err != nil {
		return fmt.Errorf("failed to update go.mod: %w", err)
	}

	// Commit version update
	if err := commitVersionUpdate(nextVersion, bumpReason); err != nil {
		return fmt.Errorf("failed to commit version update: %w", err)
	}

	fmt.Printf("\n✅ Version bumped to %s\n", nextVersion)
	fmt.Printf("💡 To create a release tag, run: git tag v%s && git push origin v%s\n", nextVersion, nextVersion)

	return nil
}

// getCurrentVersion gets the current version from go.mod or git tags
func getCurrentVersion() (string, error) {
	// First try to get version from existing tags
	cmd := exec.Command("git", "describe", "--tags", "--abbrev=0", "--match=v*")
	output, err := cmd.Output()
	if err == nil {
		version := strings.TrimSpace(string(output))
		// Remove 'v' prefix if present
		if strings.HasPrefix(version, "v") {
			version = version[1:]
		}
		return version, nil
	}

	// Fallback: read from go.mod
	goModPath := filepath.Join(".", "go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		return "", fmt.Errorf("no git tags found and go.mod not present")
	}

	// Simple pattern matching for version in go.mod
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return "", err
	}

	// Look for version in go.mod (rare, but possible)
	re := regexp.MustCompile(`module.*github\.com/royisme/BobaMixer.*v([0-9]+\.[0-9]+\.[0-9]+)`)
	matches := re.FindStringSubmatch(string(content))
	if len(matches) > 1 {
		return matches[1], nil
	}

	return "1.0.0", nil // Default version
}

// calculateNextVersion determines the next version based on current version and bump type
func calculateNextVersion(current, bumpType string) (string, string, error) {
	// Parse current version
	parts := strings.Split(current, ".")
	if len(parts) != 3 {
		return "", "", fmt.Errorf("invalid version format: %s", current)
	}

	major, _ := strconv.Atoi(parts[0])
	minor, _ := strconv.Atoi(parts[1])
	patch, _ := strconv.Atoi(parts[2])

	var nextMajor, nextMinor, nextPatch int
	var reason string

	if bumpType == "auto" {
		// Analyze commits since last tag to determine bump type
		bumpType, reason = analyzeCommitsForBump()
	}

	switch bumpType {
	case "major":
		nextMajor, nextMinor, nextPatch = major+1, 0, 0
		reason = "major release (breaking changes)"
	case "minor":
		nextMajor, nextMinor, nextPatch = major, minor+1, 0
		reason = "minor release (new features)"
	case "patch":
		nextMajor, nextMinor, nextPatch = major, minor, patch+1
		reason = "patch release (bug fixes)"
	default:
		return "", "", fmt.Errorf("unknown bump type: %s", bumpType)
	}

	if bumpType == "auto" {
		reason = fmt.Sprintf("%s (auto-detected)", reason)
	}

	nextVersion := fmt.Sprintf("%d.%d.%d", nextMajor, nextMinor, nextPatch)
	return nextVersion, reason, nil
}

// analyzeCommitsForBump analyzes commits since last tag to determine bump type
func analyzeCommitsForBump() (string, string) {
	// Get commits since last tag
	cmd := exec.Command("sh", "-c", "git log --oneline $(git describe --tags --abbrev=0 2>/dev/null || echo v0.0.0)..HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "patch", "no commits found"
	}

	commits := strings.Split(string(output), "\n")

	// Look for breaking changes first
	for _, commit := range commits {
		if strings.Contains(strings.ToLower(commit), "break") ||
			strings.HasPrefix(commit, "feat") && strings.Contains(strings.ToLower(commit), "break") {
			return "major", "breaking changes detected"
		}
	}

	// Look for features
	for _, commit := range commits {
		if strings.HasPrefix(commit, "feat") {
			return "minor", "new features detected"
		}
	}

	// Look for fixes
	for _, commit := range commits {
		if strings.HasPrefix(commit, "fix") {
			return "patch", "bug fixes detected"
		}
	}

	return "patch", "no significant changes detected"
}

// updateVersionInGoMod updates the version in go.mod (if present)
func updateVersionInGoMod(newVersion string) error {
	// For now, we don't modify go.mod as version is managed by git tags
	// This function can be extended in the future if needed
	return nil
}

// commitVersionUpdate commits the version bump
func commitVersionUpdate(version, reason string) error {
	cmd := exec.Command("git", "add", "-A")
	if err := cmd.Run(); err != nil {
		return err
	}

	commitMsg := fmt.Sprintf("chore: bump version to %s (%s)", version, reason)
	cmd = exec.Command("git", "commit", "-m", commitMsg)
	return cmd.Run()
}
