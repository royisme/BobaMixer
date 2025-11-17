// Package config provides configuration merging capabilities
package config

import "fmt"

const defaultProfileKey = "default"

// MergedConfig represents the final configuration after applying all overrides
// Priority order: Session > Branch > Project > Global
type MergedConfig struct {
	ActiveProfile string
	Routes        *RoutesConfig
	Budget        *BudgetConfig
	Overrides     []string // List of sources that provided overrides
}

// BudgetConfig represents budget configuration at any level
type BudgetConfig struct {
	DailyUSD   float64
	HardCapUSD float64
	Source     string // "global", "project", "branch", or "session"
}

// ConfigMerger handles configuration merging from multiple sources.
//
//nolint:revive // ConfigMerger is the established API name
type ConfigMerger struct {
	home string
}

// NewConfigMerger creates a new configuration merger
func NewConfigMerger(home string) *ConfigMerger {
	return &ConfigMerger{home: home}
}

// Merge merges configurations with the following priority (highest to lowest):
// 1. Session-specific config (env vars, CLI flags)
// 2. Branch-specific config (.boba-project.yaml with branch overrides)
// 3. Project config (.boba-project.yaml)
// 4. Global config (~/.boba/)
func (m *ConfigMerger) Merge(project, branch string, sessionOverrides map[string]interface{}) (*MergedConfig, error) {
	merged := &MergedConfig{
		Overrides: []string{},
	}

	// 1. Load global configuration (base layer)
	activeProfile, err := LoadActiveProfile(m.home)
	if err != nil {
		return nil, fmt.Errorf("load active profile: %w", err)
	}
	if activeProfile == "" {
		activeProfile = defaultProfileKey
	}
	merged.ActiveProfile = activeProfile
	merged.Overrides = append(merged.Overrides, "global")

	routes, err := LoadRoutes(m.home)
	if err == nil {
		merged.Routes = routes
	}

	// 2. Load project configuration (overrides global)
	if project != "" {
		// In a real implementation, we would load project-specific config
		// For now, we just record that project config was considered
		merged.Overrides = append(merged.Overrides, "project:"+project)
	}

	// 3. Load branch-specific configuration (overrides project)
	if branch != "" {
		// In a real implementation, we would load branch-specific config
		// For now, we just record that branch config was considered
		merged.Overrides = append(merged.Overrides, "branch:"+branch)
	}

	// 4. Apply session overrides (highest priority)
	if len(sessionOverrides) > 0 {
		if profile, ok := sessionOverrides["profile"].(string); ok {
			merged.ActiveProfile = profile
		}
		merged.Overrides = append(merged.Overrides, "session")
	}

	return merged, nil
}

// GetEffectiveProfile returns the effective profile after applying all overrides
func (m *ConfigMerger) GetEffectiveProfile(project, branch string, sessionProfile string) (string, []string) {
	overrides := []string{}

	// Start with global
	activeProfile, err := LoadActiveProfile(m.home)
	if err != nil {
		activeProfile = defaultProfileKey
	}
	if activeProfile == "" {
		activeProfile = defaultProfileKey
	}
	overrides = append(overrides, "global:"+activeProfile)

	// Project override (if project-specific profile is configured)
	if project != "" {
		// In real implementation, check .boba-project.yaml for preferred_profiles
		overrides = append(overrides, "project:"+project)
	}

	// Branch override (if branch-specific profile is configured)
	if branch != "" {
		// In real implementation, check for branch-specific config
		overrides = append(overrides, "branch:"+branch)
	}

	// Session override (highest priority)
	if sessionProfile != "" {
		activeProfile = sessionProfile
		overrides = append(overrides, "session:"+sessionProfile)
	}

	return activeProfile, overrides
}

// ResolveConfigOrder describes the configuration resolution order
func ResolveConfigOrder() []string {
	return []string{
		"1. Global (~/.boba/) - Base configuration",
		"2. Project (.boba-project.yaml) - Project-specific overrides",
		"3. Branch (branch config in .boba-project.yaml) - Branch-specific overrides",
		"4. Session (env vars, CLI flags) - Runtime overrides (highest priority)",
	}
}
