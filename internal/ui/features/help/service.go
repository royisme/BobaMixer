// Package help provides the service layer for help view data and logic.
package help

import "github.com/royisme/bobamixer/internal/ui/components"

// Service manages help view data and logic.
type Service struct{}

// NewService creates a new help service.
func NewService() *Service {
	return &Service{}
}

// ViewData returns all static data for the help view.
func (s *Service) ViewData() ViewData {
	return ViewData{
		Title:          "❓ BobaMixer Help & Shortcuts",
		Subtitle:       "",
		NavigationHint: "Press Esc to close this overlay",
	}
}

// GetDefaultTips returns the default help tips.
func (s *Service) GetDefaultTips() []string {
	return []string{
		"Use number keys (1-5) to jump between sections",
		"All interactive features live in the TUI",
		"CLI commands remain available for automation",
		"Press ? anytime to toggle this help overlay",
	}
}

// GetDefaultLinks returns the default help links.
func (s *Service) GetDefaultLinks() []HelpLink {
	return []HelpLink{
		{Label: "Full docs", URL: "https://royisme.github.io/BobaMixer/"},
		{Label: "GitHub", URL: "https://github.com/royisme/BobaMixer"},
	}
}

// ConvertLinksToComponents converts help links to component format.
func (s *Service) ConvertLinksToComponents(links []HelpLink) []components.HelpLink {
	result := make([]components.HelpLink, len(links))
	for i, link := range links {
		result[i] = components.HelpLink{
			Label: link.Label,
			URL:   link.URL,
		}
	}
	return result
}

// HelpLink represents a help documentation link.
type HelpLink struct {
	Label string
	URL   string
}

// ViewData holds all data needed to render the help view.
type ViewData struct {
	Title          string
	Subtitle       string
	NavigationHint string
}
