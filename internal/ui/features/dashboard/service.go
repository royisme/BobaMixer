// Package dashboard provides the service layer for dashboard view data and logic.
package dashboard

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	"github.com/royisme/bobamixer/internal/domain/core"
)

const (
	// Icon constants for status display
	IconCircleFilled = "●"
	IconCircleEmpty  = "○"
	IconCheckmark    = "✓"
	IconCross        = "❌"
	IconWarning      = "⚠"

	// Proxy state indicators
	ProxyStateOn  = "🟢 ON"
	ProxyStateOff = "🔴 OFF"

	// Help text
	HelpTextNavigation = "[↑↓] Navigate  [Tab] Next Section  [V] Views  [C] Config  [Q] Quit"
	HelpTextActions    = "[R] Run Tool  [X] Toggle Proxy"

	// Messages
	MsgNoProviderSelected = "No provider selected"
	MsgInvalidProvider    = "Invalid provider configuration"
)

// Service manages dashboard table data construction.
type Service struct {
	tools     *core.ToolsConfig
	bindings  *core.BindingsConfig
	providers *core.ProvidersConfig
	secrets   *core.SecretsConfig
}

// NewService creates a new dashboard service.
func NewService(
	tools *core.ToolsConfig,
	bindings *core.BindingsConfig,
	providers *core.ProvidersConfig,
	secrets *core.SecretsConfig,
) *Service {
	return &Service{
		tools:     tools,
		bindings:  bindings,
		providers: providers,
		secrets:   secrets,
	}
}

// BuildTableRows creates table rows from current configuration.
func (s *Service) BuildTableRows() []table.Row {
	rows := make([]table.Row, 0)

	for _, tool := range s.tools.Tools {
		row := s.buildRowForTool(tool)
		rows = append(rows, row)
	}

	if len(rows) == 0 {
		rows = append(rows, table.Row{
			"No tools configured",
			"-",
			"-",
			"-",
			"-",
		})
	}

	return rows
}

// buildRowForTool builds a single table row for a given tool.
func (s *Service) buildRowForTool(tool core.Tool) table.Row {
	// Find binding for this tool
	binding, err := s.bindings.FindBinding(tool.ID)
	if err != nil {
		// No binding, show as not configured
		return table.Row{
			tool.Name,
			"(not bound)",
			"-",
			"-",
			IconWarning + " Not configured",
		}
	}

	// Find provider
	provider, err := s.providers.FindProvider(binding.ProviderID)
	if err != nil {
		// Provider not found
		return table.Row{
			tool.Name,
			fmt.Sprintf("(missing: %s)", binding.ProviderID),
			"-",
			"-",
			IconCross + " Error",
		}
	}

	// Check API key status
	keyStatus := IconCheckmark + " Ready"
	if _, err := core.ResolveAPIKey(provider, s.secrets); err != nil {
		keyStatus = IconWarning + " No API key"
	}

	// Determine model
	model := s.determineModel(provider, binding)

	// Proxy status
	proxyStatus := ProxyStateOff
	if binding.UseProxy {
		proxyStatus = ProxyStateOn
	}

	return table.Row{
		tool.Name,
		provider.DisplayName,
		model,
		proxyStatus,
		keyStatus,
	}
}

// determineModel determines the model to display for a binding.
func (s *Service) determineModel(provider *core.Provider, binding *core.Binding) string {
	model := provider.DefaultModel
	if binding.Options.Model != "" {
		model = binding.Options.Model
	}

	// Truncate if too long
	if len(model) > 23 {
		model = model[:20] + "..."
	}

	return model
}

// GetNavigationHelp returns the navigation help text.
func (s *Service) GetNavigationHelp() string {
	return HelpTextNavigation
}

// GetActionHelp returns the action help text.
func (s *Service) GetActionHelp() string {
	return HelpTextActions
}
