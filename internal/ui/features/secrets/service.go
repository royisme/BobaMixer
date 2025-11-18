// Package secrets provides the service layer for secrets view data and logic.
package secrets

import (
	"fmt"
	"strings"

	"github.com/royisme/bobamixer/internal/domain/core"
	"github.com/royisme/bobamixer/internal/ui/components"
	"github.com/royisme/bobamixer/internal/ui/forms"
)

// Service manages secrets-related UI logic independently of the root model.
type Service struct {
	Providers      *core.ProvidersConfig
	secrets        **core.SecretsConfig
	form           *forms.SecretForm
	message        *string
	msgNoSelection string
	msgInvalid     string
}

// NewService wires the backing configs, form, and message target for secret management.
func NewService(
	providers *core.ProvidersConfig,
	secrets **core.SecretsConfig,
	form *forms.SecretForm,
	message *string,
	noSelectionMsg string,
	invalidMsg string,
) *Service {
	return &Service{
		Providers:      providers,
		secrets:        secrets,
		form:           form,
		message:        message,
		msgNoSelection: noSelectionMsg,
		msgInvalid:     invalidMsg,
	}
}

// StartForm prepares the secret form for the currently selected provider.
func (s *Service) StartForm(indexes []int, selectedIndex int) bool {
	if len(indexes) == 0 || selectedIndex < 0 || selectedIndex >= len(indexes) {
		s.setMessage(s.msgNoSelection)
		return false
	}

	targetIdx := indexes[selectedIndex]
	if targetIdx < 0 || targetIdx >= len(s.Providers.Providers) {
		s.setMessage(s.msgInvalid)
		return false
	}

	provider := s.Providers.Providers[targetIdx]
	s.form.Start(targetIdx, provider.DisplayName)
	s.form.SetMessage("")
	s.setMessage("")
	return true
}

// SaveValue persists the submitted secret for the current provider.
func (s *Service) SaveValue(home string, value string) {
	targetIdx := s.form.TargetIndex()
	if targetIdx < 0 || targetIdx >= len(s.Providers.Providers) {
		s.setMessage(s.msgInvalid)
		return
	}

	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		s.form.SetMessage("API key cannot be empty")
		s.setMessage("API key cannot be empty")
		return
	}

	provider := s.Providers.Providers[targetIdx]
	cfg := s.ensureConfig()
	cfg.Secrets[provider.ID] = core.Secret{
		ProviderID: provider.ID,
		APIKey:     trimmed,
	}

	if err := core.SaveSecrets(home, cfg); err != nil {
		msg := fmt.Sprintf("Failed to save API key: %v", err)
		s.form.SetMessage(msg)
		s.setMessage(msg)
		return
	}

	msg := fmt.Sprintf("API key saved for %s", provider.DisplayName)
	s.form.SetMessage(msg)
	s.setMessage(msg)
}

// Remove deletes the stored secret for the selected provider.
func (s *Service) Remove(home string, indexes []int, selectedIndex int) {
	if len(indexes) == 0 || selectedIndex < 0 || selectedIndex >= len(indexes) {
		s.setMessage(s.msgNoSelection)
		return
	}

	targetIdx := indexes[selectedIndex]
	if targetIdx < 0 || targetIdx >= len(s.Providers.Providers) {
		s.setMessage(s.msgInvalid)
		return
	}

	cfg := s.ensureConfig()
	provider := s.Providers.Providers[targetIdx]
	if _, ok := cfg.Secrets[provider.ID]; !ok {
		s.setMessage(fmt.Sprintf("No API key found for %s", provider.DisplayName))
		return
	}

	delete(cfg.Secrets, provider.ID)
	if err := core.SaveSecrets(home, cfg); err != nil {
		s.setMessage(fmt.Sprintf("Failed to remove API key: %v", err))
		return
	}

	s.setMessage(fmt.Sprintf("Removed API key for %s", provider.DisplayName))
}

// Test validates whether a secret exists for the selected provider.
func (s *Service) Test(indexes []int, selectedIndex int) {
	if len(indexes) == 0 || selectedIndex < 0 || selectedIndex >= len(indexes) {
		s.setMessage(s.msgNoSelection)
		return
	}

	targetIdx := indexes[selectedIndex]
	if targetIdx < 0 || targetIdx >= len(s.Providers.Providers) {
		s.setMessage(s.msgInvalid)
		return
	}

	provider := s.Providers.Providers[targetIdx]
	cfg := *s.secrets
	if cfg == nil {
		s.setMessage("API key missing: no secrets configured")
		return
	}

	if _, err := core.ResolveAPIKey(&provider, cfg); err != nil {
		s.setMessage(fmt.Sprintf("API key missing: %v", err))
		return
	}

	s.setMessage(fmt.Sprintf("API key available for %s", provider.DisplayName))
}

// Rows converts providers into UI rows annotated with secret status.
func (s *Service) Rows(indexes []int) []components.SecretProviderRow {
	cfg := *s.secrets
	if s.Providers == nil || cfg == nil || len(indexes) == 0 {
		return nil
	}

	result := make([]components.SecretProviderRow, 0, len(indexes))
	for _, idx := range indexes {
		if idx < 0 || idx >= len(s.Providers.Providers) {
			continue
		}

		provider := s.Providers.Providers[idx]
		hasKey := false
		keySource := "(not set)"
		if _, err := core.ResolveAPIKey(&provider, cfg); err == nil {
			hasKey = true
			keySource = string(provider.APIKey.Source)
		}

		result = append(result, components.SecretProviderRow{
			DisplayName: provider.DisplayName,
			HasKey:      hasKey,
			KeySource:   keySource,
		})
	}

	return result
}

// EmptyStateMessage describes secrets empty states with optional search context.
func EmptyStateMessage(isEmpty bool, hasSearch bool) string {
	if !isEmpty {
		return ""
	}
	if hasSearch {
		return "No providers match the current filter."
	}
	return "No providers configured."
}

func (s *Service) ensureConfig() *core.SecretsConfig {
	if *s.secrets == nil {
		*s.secrets = &core.SecretsConfig{
			Version: 1,
			Secrets: make(map[string]core.Secret),
		}
	}
	if (*s.secrets).Secrets == nil {
		(*s.secrets).Secrets = make(map[string]core.Secret)
	}
	return *s.secrets
}

func (s *Service) setMessage(msg string) {
	if s.message != nil {
		*s.message = msg
	}
}
