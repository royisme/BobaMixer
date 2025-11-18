// Package hooks provides the service layer for hooks view data and logic.
package hooks

import "github.com/royisme/bobamixer/internal/ui/components"

// Service manages hooks view data and logic.
type Service struct{}

// NewService creates a new hooks service.
func NewService() *Service {
	return &Service{}
}

// ViewData returns all static data for the hooks view.
func (s *Service) ViewData() ViewData {
	return ViewData{
		Title:           "🪝 Git Hooks Management",
		RepoTitle:       "Current Repository",
		HooksTitle:      "Available Hooks",
		BenefitsTitle:   "Benefits",
		ActivityTitle:   "Recent Hook Activity",
		CommandHelpLine: "Use CLI: boba hooks install (install) | boba hooks remove (uninstall)",
	}
}

// GetAvailableHooks returns the list of available git hooks.
func (s *Service) GetAvailableHooks(installed bool) []HookInfo {
	return []HookInfo{
		{
			Name:   "post-checkout",
			Desc:   "Track branch switches and suggest optimal profiles",
			Active: installed,
		},
		{
			Name:   "post-commit",
			Desc:   "Record commit events for usage tracking",
			Active: installed,
		},
		{
			Name:   "post-merge",
			Desc:   "Track merge events and repository changes",
			Active: installed,
		},
	}
}

// ConvertToComponents converts hooks to component format.
func (s *Service) ConvertToComponents(hooks []HookInfo) []components.HookInfo {
	result := make([]components.HookInfo, len(hooks))
	for i, hook := range hooks {
		result[i] = components.HookInfo{
			Name:   hook.Name,
			Desc:   hook.Desc,
			Active: hook.Active,
		}
	}
	return result
}

// HookInfo represents a git hook entry.
type HookInfo struct {
	Name   string
	Desc   string
	Active bool
}

// ViewData holds all data needed to render the hooks view.
type ViewData struct {
	Title           string
	RepoTitle       string
	HooksTitle      string
	BenefitsTitle   string
	ActivityTitle   string
	CommandHelpLine string
}
