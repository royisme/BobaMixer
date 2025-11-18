// Package providers provides the service layer for providers view data and logic.
package providers

import (
	"fmt"

	"github.com/royisme/bobamixer/internal/domain/core"
	"github.com/royisme/bobamixer/internal/ui/components"
	"github.com/royisme/bobamixer/internal/ui/forms"
)

// Service encapsulates provider-related UI logic such as form handling and
// table data preparation so the root model can stay focused on orchestration.
type Service struct {
	providers      *core.ProvidersConfig
	secrets        *core.SecretsConfig
	form           *forms.ProviderForm
	msgNoSelection string
	msgInvalid     string
}

// NewService wires the provider config, secrets config, and backing form into
// a dedicated helper for the providers view.
func NewService(
	providers *core.ProvidersConfig,
	secrets *core.SecretsConfig,
	form *forms.ProviderForm,
	noSelectionMsg string,
	invalidMsg string,
) *Service {
	return &Service{
		providers:      providers,
		secrets:        secrets,
		form:           form,
		msgNoSelection: noSelectionMsg,
		msgInvalid:     invalidMsg,
	}
}

// StartForm starts the provider form either in add mode or edit mode depending
// on the provided flags and selection indexes.
func (s *Service) StartForm(add bool, indexes []int, selectedIndex int) bool {
	if s.form == nil {
		return false
	}

	if add {
		s.form.Start(true, core.Provider{}, -1, s.providers)
		s.form.SetMessage("")
		return true
	}

	if len(indexes) == 0 || selectedIndex < 0 || selectedIndex >= len(indexes) {
		s.form.SetMessage(s.msgNoSelection)
		return false
	}

	if s.providers == nil {
		s.form.SetMessage(s.msgInvalid)
		return false
	}

	targetIdx := indexes[selectedIndex]
	if targetIdx < 0 || targetIdx >= len(s.providers.Providers) {
		s.form.SetMessage(s.msgInvalid)
		return false
	}

	provider := s.providers.Providers[targetIdx]
	s.form.Start(false, provider, targetIdx, s.providers)
	s.form.SetMessage("")
	return true
}

// Save commits the provider currently captured by the form to disk.
func (s *Service) Save(home string) error {
	if s.form == nil || s.providers == nil {
		return fmt.Errorf("provider service not initialized")
	}

	provider := s.form.Provider()
	index := s.form.Index()

	if provider.ID == "" {
		s.form.SetMessage("provider ID is required")
		return fmt.Errorf("provider ID is required")
	}

	if s.form.AddMode() {
		s.providers.Providers = append(s.providers.Providers, provider)
	} else if index >= 0 && index < len(s.providers.Providers) {
		s.providers.Providers[index] = provider
	} else {
		s.form.SetMessage(s.msgInvalid)
		return fmt.Errorf("invalid provider index")
	}

	if err := core.SaveProviders(home, s.providers); err != nil {
		s.form.SetMessage(fmt.Sprintf("failed to save provider: %v", err))
		return err
	}

	if s.form.AddMode() {
		s.form.SetMessage(fmt.Sprintf("provider %s created", provider.DisplayName))
	} else {
		s.form.SetMessage(fmt.Sprintf("provider %s updated", provider.DisplayName))
	}
	return nil
}

// Rows converts the filtered provider indexes into table rows for the page.
func (s *Service) Rows(indexes []int) []components.ProviderRow {
	if s.providers == nil || len(indexes) == 0 {
		return nil
	}

	result := make([]components.ProviderRow, 0, len(indexes))
	for _, idx := range indexes {
		if idx < 0 || idx >= len(s.providers.Providers) {
			continue
		}

		provider := s.providers.Providers[idx]
		hasKey := false
		if s.secrets != nil {
			if _, err := core.ResolveAPIKey(&provider, s.secrets); err == nil {
				hasKey = true
			}
		}

		result = append(result, components.ProviderRow{
			DisplayName:  provider.DisplayName,
			BaseURL:      provider.BaseURL,
			DefaultModel: provider.DefaultModel,
			Enabled:      provider.Enabled,
			HasAPIKey:    hasKey,
		})
	}

	return result
}

// Details returns the sidebar details for the currently selected provider.
func (s *Service) Details(indexes []int, selectedIndex int) *components.ProviderDetails {
	if s.providers == nil || len(indexes) == 0 || selectedIndex < 0 || selectedIndex >= len(indexes) {
		return nil
	}

	idx := indexes[selectedIndex]
	if idx < 0 || idx >= len(s.providers.Providers) {
		return nil
	}

	provider := s.providers.Providers[idx]
	details := components.ProviderDetails{
		ID:           provider.ID,
		Kind:         string(provider.Kind),
		APIKeySource: string(provider.APIKey.Source),
	}

	if provider.APIKey.Source == core.APIKeySourceEnv && provider.APIKey.EnvVar != "" {
		details.EnvVar = provider.APIKey.EnvVar
		details.ShowEnvVar = true
	}

	return &details
}

// EmptyStateMessage builds the empty-state hint for the providers page.
func (s *Service) EmptyStateMessage(isEmpty bool, hasSearch bool) string {
	if !isEmpty {
		return ""
	}
	if hasSearch {
		return "No providers match the current filter."
	}
	return "No providers configured."
}
