// Package version provides version information for BobaMixer.
package version

import (
	"fmt"
	"runtime"
)

// Version information
const defaultVersion = "dev"

var (
	Version = defaultVersion // Version is set during build
	Commit  = "unknown"      // Commit hash is set during build
	Date    = "unknown"      // Build date is set during build
	BuiltBy = "unknown"      // Builder information
)

// VersionInfo holds all version information
type VersionInfo struct {
	Version string
	Commit  string
	Date    string
	BuiltBy string
	GoOS    string
	GoArch  string
}

// GetVersionInfo returns complete version information
func GetVersionInfo() VersionInfo {
	return VersionInfo{
		Version: Version,
		Commit:  Commit,
		Date:    Date,
		BuiltBy: BuiltBy,
		GoOS:    runtime.GOOS,
		GoArch:  runtime.GOARCH,
	}
}

// String returns a formatted version string
func (v VersionInfo) String() string {
	version := v.Version
	if version == "" {
		version = defaultVersion
	}
	commit := v.Commit
	if commit == "" {
		commit = "unknown"
	}
	if len(commit) > 7 {
		commit = commit[:7]
	}
	build := v.Date
	if build == "" {
		build = "unknown"
	}
	meta := "release"
	if version == defaultVersion {
		meta = "development"
	}
	return fmt.Sprintf("BobaMixer %s %s (commit: %s, built: %s, os/arch: %s/%s)",
		meta, version, commit, build, v.GoOS, v.GoArch)
}

// FullString returns a detailed version string
func (v VersionInfo) FullString() string {
	return fmt.Sprintf(`BobaMixer Version: %s
Commit:      %s
Build Date:  %s
Built By:    %s
Platform:    %s/%s
Go Version:  %s`,
		v.Version, v.Commit, v.Date, v.BuiltBy, v.GoOS, v.GoArch, runtime.Version())
}

// IsDev returns true if this is a development version
func (v VersionInfo) IsDev() bool {
	return v.Version == defaultVersion || v.Version == ""
}

// IsRelease returns true if this is a release version (not dev)
func (v VersionInfo) IsRelease() bool {
	return !v.IsDev()
}
