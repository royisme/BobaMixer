// Package bindings provides the service layer for bindings view data and logic.
package bindings

import (
	"fmt"

	"github.com/royisme/bobamixer/internal/domain/core"
	"github.com/royisme/bobamixer/internal/ui/components"
	"github.com/royisme/bobamixer/internal/ui/forms"
)

// Service encapsulates bindings-related UI logic so the root model can focus on orchestration.
type Service struct {
	bindings       *core.BindingsConfig
	tools          *core.ToolsConfig
	providers      *core.ProvidersConfig
	form           *forms.BindingForm
	msgNoSelection string
	msgInvalid     string
}

// NewService returns a helper that wires configs, form, and shared messages for bindings.
func NewService(
	bindings *core.BindingsConfig,
	tools *core.ToolsConfig,
	providers *core.ProvidersConfig,
	form *forms.BindingForm,
	noSelectionMsg string,
	invalidMsg string,
) *Service {
	return &Service{
		bindings:       bindings,
		tools:          tools,
		providers:      providers,
		form:           form,
		msgNoSelection: noSelectionMsg,
		msgInvalid:     invalidMsg,
	}
}

// StartForm prepares the binding form for either creating or editing a binding.
func (s *Service) StartForm(add bool, indexes []int, selectedIndex int) bool {
	if s.form == nil {
		return false
	}

	var (
		binding core.Binding
		index   = -1
	)

	if add {
		binding = core.Binding{
			UseProxy: true,
			Options:  core.BindingOptions{},
		}
	} else {
		if len(indexes) == 0 || selectedIndex < 0 || selectedIndex >= len(indexes) {
			s.form.SetMessage(s.msgNoSelection)
			return false
		}
		if s.bindings == nil {
			s.form.SetMessage(s.msgInvalid)
			return false
		}
		targetIdx := indexes[selectedIndex]
		if targetIdx < 0 || targetIdx >= len(s.bindings.Bindings) {
			s.form.SetMessage(s.msgInvalid)
			return false
		}
		binding = s.bindings.Bindings[targetIdx]
		index = targetIdx
	}

	s.form.Start(add, binding, index, s.bindings, s.tools, s.providers)
	s.form.SetMessage("")
	return true
}

// Save persists the binding currently captured by the form.
func (s *Service) Save(home string) error {
	if s.form == nil || s.bindings == nil {
		return fmt.Errorf("binding service not initialized")
	}

	binding := s.form.Binding()
	index := s.form.Index()

	if binding.ToolID == "" {
		s.form.SetMessage("tool id is required")
		return fmt.Errorf("tool id is required")
	}
	if binding.ProviderID == "" {
		s.form.SetMessage("provider id is required")
		return fmt.Errorf("provider id is required")
	}

	if s.form.AddMode() {
		s.bindings.Bindings = append(s.bindings.Bindings, binding)
	} else if index >= 0 && index < len(s.bindings.Bindings) {
		s.bindings.Bindings[index] = binding
	} else {
		s.form.SetMessage(s.msgInvalid)
		return fmt.Errorf("invalid binding index")
	}

	if err := core.SaveBindings(home, s.bindings); err != nil {
		s.form.SetMessage(fmt.Sprintf("failed to save binding: %v", err))
		return err
	}

	if s.form.AddMode() {
		s.form.SetMessage(fmt.Sprintf("binding created for %s", binding.ToolID))
	} else {
		s.form.SetMessage(fmt.Sprintf("binding updated for %s", binding.ToolID))
	}

	return nil
}

// Rows converts filtered binding indexes into rows for the bindings page.
func (s *Service) Rows(indexes []int) []components.BindingRow {
	if s.bindings == nil || len(indexes) == 0 {
		return nil
	}

	result := make([]components.BindingRow, 0, len(indexes))
	for _, idx := range indexes {
		if idx < 0 || idx >= len(s.bindings.Bindings) {
			continue
		}

		b := s.bindings.Bindings[idx]
		toolName := b.ToolID
		if s.tools != nil {
			if tool, err := s.tools.FindTool(b.ToolID); err == nil && tool.Name != "" {
				toolName = tool.Name
			}
		}

		providerName := b.ProviderID
		if s.providers != nil {
			if provider, err := s.providers.FindProvider(b.ProviderID); err == nil && provider.DisplayName != "" {
				providerName = provider.DisplayName
			}
		}

		result = append(result, components.BindingRow{
			ToolName:     toolName,
			ProviderName: providerName,
			UseProxy:     b.UseProxy,
		})
	}

	return result
}

// Details returns the sidebar details for the selected binding.
func (s *Service) Details(indexes []int, selectedIndex int) *components.BindingDetails {
	if s.bindings == nil || len(indexes) == 0 || selectedIndex < 0 || selectedIndex >= len(indexes) {
		return nil
	}

	bindingIdx := indexes[selectedIndex]
	if bindingIdx < 0 || bindingIdx >= len(s.bindings.Bindings) {
		return nil
	}

	b := s.bindings.Bindings[bindingIdx]
	return &components.BindingDetails{
		ToolID:        b.ToolID,
		ProviderID:    b.ProviderID,
		UseProxy:      b.UseProxy,
		ModelOverride: b.Options.Model,
	}
}

// EmptyStateMessage builds the empty-state text for the bindings table.
func (s *Service) EmptyStateMessage(isEmpty bool, hasSearch bool) string {
	if !isEmpty {
		return ""
	}
	if hasSearch {
		return "No bindings match the current filter."
	}
	return "No bindings configured."
}
