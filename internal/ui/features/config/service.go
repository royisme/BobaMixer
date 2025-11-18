// Package config provides the service layer for configuration view data and logic.
package config

import "github.com/royisme/bobamixer/internal/ui/components"

// Service manages configuration view data and logic.
type Service struct{}

// NewService creates a new config service.
func NewService() *Service {
	return &Service{}
}

// ViewData returns all static data for the config view.
func (s *Service) ViewData(home string) ViewData {
	return ViewData{
		Title:           "⚙️  Configuration Editor",
		ConfigTitle:     "Configuration Files",
		EditorTitle:     "Editor Settings",
		SafetyTitle:     "Safety Features",
		ConfigFiles:     s.GetConfigFiles(),
		Home:            home,
		EditorName:      "vim",
		CommandHelpLine: "Use CLI: boba edit <target> (to open in editor)",
	}
}

// GetConfigFiles returns the list of configuration files.
func (s *Service) GetConfigFiles() []ConfigFile {
	return []ConfigFile{
		{Name: "Providers", File: "providers.yaml", Desc: "AI provider configurations and API endpoints"},
		{Name: "Tools", File: "tools.yaml", Desc: "CLI tool detection and management"},
		{Name: "Bindings", File: "bindings.yaml", Desc: "Tool-to-provider bindings and proxy settings"},
		{Name: "Secrets", File: "secrets.yaml", Desc: "Encrypted API keys (edit with caution!)"},
		{Name: "Routes", File: "routes.yaml", Desc: "Context-based routing rules"},
		{Name: "Pricing", File: "pricing.yaml", Desc: "Token pricing for cost calculations"},
		{Name: "Settings", File: "settings.yaml", Desc: "Global application settings"},
	}
}

// ConvertToComponents converts config files to component format.
func (s *Service) ConvertToComponents() []components.ConfigFile {
	files := s.GetConfigFiles()
	result := make([]components.ConfigFile, len(files))
	for i, cfg := range files {
		result[i] = components.ConfigFile{
			Name: cfg.Name,
			File: cfg.File,
			Desc: cfg.Desc,
		}
	}
	return result
}

// ConfigFile represents a configuration file entry.
type ConfigFile struct {
	Name string
	File string
	Desc string
}

// ViewData holds all data needed to render the config view.
type ViewData struct {
	Title           string
	ConfigTitle     string
	EditorTitle     string
	SafetyTitle     string
	ConfigFiles     []ConfigFile
	Home            string
	EditorName      string
	CommandHelpLine string
}
