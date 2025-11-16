package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	version "github.com/royisme/bobamixer/internal/domain/version"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		printUsage()
		return nil
	}

	switch args[0] {
	case "bump":
		return runBump(args[1:])
	case "release":
		return runRelease(args[1:])
	case "help", "--help", "-h":
		printUsage()
		return nil
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func printUsage() {
	fmt.Println("boba-maint - maintainer tooling for BobaMixer")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  boba-maint bump [patch|minor|major|auto] [--dry-run] [--notes text]")
	fmt.Println("  boba-maint release [--auto|--part patch] [--dry-run] [--push=false]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  bump     Update VERSION/CHANGELOG or preview the next version")
	fmt.Println("  release  Auto-bump, commit, tag, and optionally push a release")
}

func runBump(args []string) error {
	fs := flag.NewFlagSet("bump", flag.ContinueOnError)
	dryRun := fs.Bool("dry-run", false, "preview the bump without writing files")
	prerelease := fs.String("prerelease", "", "append prerelease suffix")
	notes := fs.String("notes", "", "changelog notes for this bump")
	fs.SetOutput(io.Discard)
	if err := fs.Parse(args); err != nil {
		return err
	}

	partArg := "auto"
	if fs.NArg() > 0 {
		partArg = fs.Arg(0)
	}

	repo, err := repoRoot()
	if err != nil {
		return err
	}
	mgr := version.NewManager(repo)
	current, err := mgr.Current()
	if err != nil {
		return fmt.Errorf("read current version: %w", err)
	}

	part, reason, err := resolvePart(partArg, repo)
	if err != nil {
		return err
	}

	next, err := mgr.Plan(part, *prerelease)
	if err != nil {
		return err
	}

	fmt.Printf("Current version: %s\n", current)
	if reason != "" {
		fmt.Printf("Bump type: %s (%s)\n", part, reason)
	} else {
		fmt.Printf("Bump type: %s\n", part)
	}
	fmt.Printf("Next version: %s\n", next)

	if *dryRun {
		return nil
	}

	bumped, err := mgr.Bump(part, *prerelease, *notes)
	if err != nil {
		return err
	}

	fmt.Printf("Updated VERSION and CHANGELOG for %s.\n", bumped)
	fmt.Println("Remember to review the changelog, then commit the changes before releasing.")
	return nil
}

func runRelease(args []string) error {
	fs := flag.NewFlagSet("release", flag.ContinueOnError)
	auto := fs.Bool("auto", false, "auto-detect bump from commits (default)")
	partFlag := fs.String("part", "", "explicit bump: patch, minor, or major")
	prerelease := fs.String("prerelease", "", "append prerelease suffix")
	notes := fs.String("notes", "", "changelog notes for this release")
	dryRun := fs.Bool("dry-run", false, "preview without modifying files")
	push := fs.Bool("push", true, "push the release commit and tag to origin")
	commitMsg := fs.String("message", "", "custom release commit message")
	fs.SetOutput(io.Discard)
	if err := fs.Parse(args); err != nil {
		return err
	}

	partArg := strings.TrimSpace(*partFlag)
	if partArg == "" {
		if *auto || fs.NArg() == 0 {
			partArg = "auto"
		} else if fs.NArg() > 0 {
			partArg = fs.Arg(0)
		}
	}

	repo, err := repoRoot()
	if err != nil {
		return err
	}
	mgr := version.NewManager(repo)
	current, err := mgr.Current()
	if err != nil {
		return fmt.Errorf("read current version: %w", err)
	}

	part, reason, err := resolvePart(partArg, repo)
	if err != nil {
		return err
	}

	next, err := mgr.Plan(part, *prerelease)
	if err != nil {
		return err
	}

	fmt.Printf("Current version: %s\n", current)
	if reason != "" {
		fmt.Printf("Planned bump: %s (%s)\n", part, reason)
	} else {
		fmt.Printf("Planned bump: %s\n", part)
	}
	fmt.Printf("Next version: %s\n", next)

	if *dryRun {
		return nil
	}

	if err := ensureClean(repo); err != nil {
		return err
	}

	bumped, err := mgr.Bump(part, *prerelease, *notes)
	if err != nil {
		return err
	}

	if err := gitAdd(repo, "VERSION", "CHANGELOG.md"); err != nil {
		return err
	}

	msg := strings.TrimSpace(*commitMsg)
	if msg == "" {
		msg = fmt.Sprintf("chore: release v%s", bumped)
	}
	if err := gitCommit(repo, msg); err != nil {
		return err
	}

	if err := gitTag(repo, bumped); err != nil {
		return err
	}

	pushedText := "!"
	if *push {
		if err := gitPush(repo); err != nil {
			return err
		}
		if err := gitPushTag(repo, bumped); err != nil {
			return err
		}
		pushedText = " and pushed!"
	}

	fmt.Printf("Creating release tag: v%s\n", bumped)
	fmt.Printf("✅ Release tag v%s created%s\n", bumped, pushedText)
	if *push {
		fmt.Println("🚀 GitHub Actions will now build and publish the release.")
	} else {
		fmt.Println("Tag created locally. Push it when you're ready:")
		fmt.Printf("  git push origin v%s\n", bumped)
	}
	return nil
}

func repoRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err == nil {
		root := strings.TrimSpace(string(out))
		if root != "" {
			return root, nil
		}
	}
	return os.Getwd()
}

func ensureClean(repo string) error {
	out, err := gitOutput(repo, "status", "--porcelain")
	if err != nil {
		return err
	}
	if strings.TrimSpace(out) != "" {
		return errors.New("working tree is not clean; commit or stash changes before releasing")
	}
	return nil
}

func resolvePart(partArg, repo string) (string, string, error) {
	normalized := strings.ToLower(strings.TrimSpace(partArg))
	switch normalized {
	case "major", "minor", "patch":
		return normalized, "specified", nil
	case "", "auto":
		return detectAutoPart(repo)
	default:
		return "", "", fmt.Errorf("invalid bump type %q (use patch, minor, major, or auto)", partArg)
	}
}

func detectAutoPart(repo string) (string, string, error) {
	lastTag, _ := gitOutput(repo, "describe", "--tags", "--abbrev=0")
	args := []string{"log", "--pretty=format:%B%x00"}
	if strings.TrimSpace(lastTag) != "" {
		args = append(args, fmt.Sprintf("%s..HEAD", strings.TrimSpace(lastTag)))
	}
	raw, err := gitRaw(repo, args...)
	if err != nil {
		return "", "", err
	}
	chunks := bytes.Split(bytes.TrimSuffix(raw, []byte{0}), []byte{0})
	level := ""
	reason := ""
	found := false
	for _, chunk := range chunks {
		msg := strings.TrimSpace(string(chunk))
		if msg == "" {
			continue
		}
		found = true
		subject := firstLine(msg)
		if isMajorCommit(subject, msg) {
			return "major", fmt.Sprintf("found breaking change in '%s'", subject), nil
		}
		if strings.HasPrefix(strings.ToLower(subject), "feat") {
			level = "minor"
			reason = fmt.Sprintf("feature commit '%s'", subject)
			continue
		}
		if strings.HasPrefix(strings.ToLower(subject), "fix") || strings.HasPrefix(strings.ToLower(subject), "perf") {
			if level == "" {
				level = "patch"
				reason = fmt.Sprintf("patch commit '%s'", subject)
			}
		}
	}
	if !found {
		return "", "", errors.New("no commits found to analyze for auto bump")
	}
	if level == "" {
		level = "patch"
		reason = "defaulting to patch (no feat/fix commits detected)"
	}
	return level, reason, nil
}

func firstLine(msg string) string {
	if idx := strings.IndexByte(msg, '\n'); idx >= 0 {
		return strings.TrimSpace(msg[:idx])
	}
	return strings.TrimSpace(msg)
}

func isMajorCommit(subject, full string) bool {
	lower := strings.ToLower(subject)
	if strings.Contains(full, "BREAKING CHANGE") || strings.Contains(strings.ToLower(full), "breaking change") {
		return true
	}
	if idx := strings.Index(subject, ":"); idx >= 0 {
		header := subject[:idx]
		if strings.HasSuffix(header, "!") {
			return true
		}
	}
	return strings.Contains(lower, "feat!")
}

func gitAdd(repo string, files ...string) error {
	args := append([]string{"add"}, files...)
	_, err := gitOutput(repo, args...)
	return err
}

func gitCommit(repo, message string) error {
	if message == "" {
		return errors.New("commit message cannot be empty")
	}
	_, err := gitOutput(repo, "commit", "-m", message)
	return err
}

func gitTag(repo, version string) error {
	tag := fmt.Sprintf("v%s", version)
	existing, _ := gitOutput(repo, "tag", "--list", tag)
	if strings.TrimSpace(existing) == tag {
		return fmt.Errorf("tag %s already exists", tag)
	}
	_, err := gitOutput(repo, "tag", "-a", tag, "-m", fmt.Sprintf("Release %s", tag))
	return err
}

func gitPush(repo string) error {
	branch, err := gitOutput(repo, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return err
	}
	_, err = gitOutput(repo, "push", "origin", branch)
	return err
}

func gitPushTag(repo, version string) error {
	tag := fmt.Sprintf("v%s", version)
	_, err := gitOutput(repo, "push", "origin", tag)
	return err
}

func gitOutput(repo string, args ...string) (string, error) {
	data, err := gitRaw(repo, args...)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func gitRaw(repo string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = repo
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git %s: %w\n%s", strings.Join(args, " "), err, string(out))
	}
	return out, nil
}
