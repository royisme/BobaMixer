package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
)

type stringList []string

func (s *stringList) String() string {
	return strings.Join(*s, ",")
}

func (s *stringList) Set(value string) error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return errors.New("core package cannot be empty")
	}
	*s = append(*s, trimmed)
	return nil
}

type summary struct {
	statements float64
	covered    float64
}

func main() {
	var (
		profile    = flag.String("profile", "coverage.out", "path to coverage profile")
		modulePath = flag.String("module", "", "module import path, e.g. github.com/example/project")
		minTotal   = flag.Float64("min-total", 0, "minimum overall coverage percentage")
		minCore    = flag.Float64("min-core", 0, "minimum per-core-package coverage percentage")
		cores      stringList
	)
	flag.Var(&cores, "core", "core package requiring stricter threshold (relative path)")
	flag.Parse()

	if err := runCovCheck(*profile, *modulePath, *minTotal, *minCore, cores); err != nil {
		fatal(err)
	}
}

func runCovCheck(profilePath, modulePath string, minTotal, minCore float64, cores []string) (retErr error) {
	if modulePath == "" {
		return errors.New("module path is required")
	}

	// #nosec G304 -- coverage profile path is provided via CLI flag intentionally
	file, err := os.Open(profilePath)
	if err != nil {
		return fmt.Errorf("open coverage profile: %w", err)
	}
	defer func() {
		if cerr := file.Close(); cerr != nil && retErr == nil {
			retErr = fmt.Errorf("close coverage profile: %w", cerr)
		}
	}()

	summaries, total, err := summarizeCoverage(file, modulePath)
	if err != nil {
		return err
	}

	totalPct := percentage(total)
	if totalPct < minTotal {
		return fmt.Errorf("overall coverage %.2f%% below required %.2f%%", totalPct, minTotal)
	}

	if err := verifyCorePackages(summaries, cores, minCore); err != nil {
		return err
	}

	fmt.Printf("overall coverage: %.2f%%\n", totalPct)
	return nil
}

func summarizeCoverage(reader io.Reader, modulePath string) (map[string]*summary, *summary, error) {
	summaries := map[string]*summary{}
	total := &summary{}
	modulePrefix := strings.TrimSuffix(modulePath, "/") + "/"

	scanner := bufio.NewScanner(reader)
	lineNum := 0
	for scanner.Scan() {
		line := scanner.Text()
		lineNum++
		if lineNum == 1 && strings.HasPrefix(line, "mode:") {
			continue
		}
		pkg, stmts, covered := parseProfileLine(line, modulePrefix)
		if stmts == 0 {
			continue
		}
		if _, ok := summaries[pkg]; !ok {
			summaries[pkg] = &summary{}
		}
		summaries[pkg].statements += stmts
		if covered {
			summaries[pkg].covered += stmts
		}
		total.statements += stmts
		if covered {
			total.covered += stmts
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, nil, fmt.Errorf("read coverage profile: %w", err)
	}

	return summaries, total, nil
}

func verifyCorePackages(summaries map[string]*summary, cores []string, minCore float64) error {
	for _, corePkg := range cores {
		pkg := strings.Trim(corePkg, "/")
		sum, ok := summaries[pkg]
		if !ok {
			return fmt.Errorf("no coverage data for core package %s", pkg)
		}
		pct := percentage(sum)
		if pct < minCore {
			return fmt.Errorf("core package %s coverage %.2f%% below required %.2f%%", pkg, pct, minCore)
		}
		fmt.Printf("core package %s coverage: %.2f%%\n", pkg, pct)
	}
	return nil
}

func parseProfileLine(line, modulePrefix string) (string, float64, bool) {
	fields := strings.Fields(line)
	if len(fields) < 3 {
		return "", 0, false
	}

	filePath := fields[0]
	stmts, err := strconv.ParseFloat(fields[1], 64)
	if err != nil {
		return "", 0, false
	}
	count, err := strconv.Atoi(fields[2])
	if err != nil {
		return "", 0, false
	}

	pkg := derivePackage(filePath, modulePrefix)
	return pkg, stmts, count > 0
}

func derivePackage(filePath, modulePrefix string) string {
	pathWithoutModule := strings.TrimPrefix(filePath, modulePrefix)
	if pathWithoutModule == filePath {
		// not under module, fall back to path directory
		return path.Dir(filePath)
	}
	pkg := path.Dir(pathWithoutModule)
	if pkg == "." || pkg == "" {
		return pathWithoutModule
	}
	return pkg
}

func percentage(sum *summary) float64 {
	if sum.statements == 0 {
		return 100
	}
	return (sum.covered / sum.statements) * 100
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
