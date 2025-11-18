package forms

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/royisme/bobamixer/internal/domain/core"
	"github.com/royisme/bobamixer/internal/ui/theme"
)

type providerField int

const (
	providerFieldID providerField = iota
	providerFieldKind
	providerFieldDisplayName
	providerFieldBaseURL
	providerFieldDefaultModel
	providerFieldAPIKeySource
	providerFieldAPIKeyEnv
)

var providerFieldSequence = []providerField{
	providerFieldID,
	providerFieldKind,
	providerFieldDisplayName,
	providerFieldBaseURL,
	providerFieldDefaultModel,
	providerFieldAPIKeySource,
	providerFieldAPIKeyEnv,
}

// ProviderForm manages the state machine for editing providers.
type ProviderForm struct {
	active    bool
	add       bool
	index     int
	fieldIdx  int
	input     textinput.Model
	provider  core.Provider
	message   string
	prompt    string
	providers *core.ProvidersConfig
}

// NewProviderForm creates a new provider form with the configured prompt prefix.
func NewProviderForm(prompt string) ProviderForm {
	input := textinput.New()
	input.CharLimit = 200
	input.Width = 50
	input.Prompt = prompt

	return ProviderForm{
		input:  input,
		prompt: prompt,
	}
}

// Active reports whether the form is currently collecting input.
func (f ProviderForm) Active() bool {
	return f.active
}

// Message returns the latest helper/error message.
func (f ProviderForm) Message() string {
	return f.message
}

// SetMessage overrides the helper message (useful for external validation feedback).
func (f *ProviderForm) SetMessage(msg string) {
	f.message = msg
}

// Provider returns the current provider being edited.
func (f ProviderForm) Provider() core.Provider {
	return f.provider
}

// Index returns the original provider index (for edit flows).
func (f ProviderForm) Index() int {
	return f.index
}

// AddMode indicates whether the form is adding a provider.
func (f ProviderForm) AddMode() bool {
	return f.add
}

// Start activates the form with either a blank or existing provider.
func (f *ProviderForm) Start(add bool, provider core.Provider, idx int, providers *core.ProvidersConfig) {
	f.add = add
	f.index = idx
	f.message = ""
	f.active = true
	f.fieldIdx = 0
	f.providers = providers

	if add {
		f.provider = core.Provider{
			Enabled: true,
			APIKey: core.APIKeyConfig{
				Source: core.APIKeySourceEnv,
			},
		}
		f.index = -1
	} else {
		f.provider = provider
	}

	f.skipDisabledFields()
	if f.fieldIdx >= len(providerFieldSequence) {
		f.active = false
		f.input.Blur()
		return
	}

	f.prepareField()
}

// Cancel stops the form and records the provided reason.
func (f *ProviderForm) Cancel(reason string) {
	f.active = false
	f.message = reason
	f.input.Blur()
	f.input.SetValue("")
}

// Update forwards messages to the text input when active.
func (f *ProviderForm) Update(msg tea.Msg) tea.Cmd {
	if !f.active {
		return nil
	}
	var cmd tea.Cmd
	f.input, cmd = f.input.Update(msg)
	return cmd
}

// Submit stores the current field value and advances the form.
// When it returns completed=true, the form is finished collecting data.
func (f *ProviderForm) Submit() (completed bool, err error) {
	if !f.active || f.fieldIdx >= len(providerFieldSequence) {
		return false, nil
	}

	field := providerFieldSequence[f.fieldIdx]
	value := strings.TrimSpace(f.input.Value())
	if err := f.setFieldValue(field, value); err != nil {
		f.message = err.Error()
		return false, err
	}

	f.message = ""
	f.input.SetValue("")
	f.fieldIdx++
	f.skipDisabledFields()

	if f.fieldIdx >= len(providerFieldSequence) {
		f.active = false
		f.input.Blur()
		return true, nil
	}

	f.prepareField()
	return false, nil
}

func (f *ProviderForm) prepareField() {
	if f.fieldIdx >= len(providerFieldSequence) {
		return
	}

	field := providerFieldSequence[f.fieldIdx]
	f.input.Placeholder = f.promptFor(field)
	f.input.SetValue(f.valueForField(field))
	f.input.CursorEnd()
	f.input.Focus()
}

func (f *ProviderForm) fieldEnabled(field providerField) bool {
	if !f.add && field == providerFieldID {
		return false
	}
	if field == providerFieldAPIKeyEnv {
		return strings.ToLower(string(f.provider.APIKey.Source)) == string(core.APIKeySourceEnv)
	}
	return true
}

func (f *ProviderForm) skipDisabledFields() {
	for f.fieldIdx < len(providerFieldSequence) && !f.fieldEnabled(providerFieldSequence[f.fieldIdx]) {
		f.fieldIdx++
	}
}

func (f *ProviderForm) promptFor(field providerField) string {
	switch field {
	case providerFieldID:
		return "provider id"
	case providerFieldKind:
		return "provider kind (openai, anthropic...)"
	case providerFieldDisplayName:
		return "display name"
	case providerFieldBaseURL:
		return "base url"
	case providerFieldDefaultModel:
		return "default model"
	case providerFieldAPIKeySource:
		return "api key source (env or secrets)"
	case providerFieldAPIKeyEnv:
		return "env var name"
	default:
		return "value"
	}
}

func (f *ProviderForm) valueForField(field providerField) string {
	switch field {
	case providerFieldID:
		return f.provider.ID
	case providerFieldKind:
		return string(f.provider.Kind)
	case providerFieldDisplayName:
		return f.provider.DisplayName
	case providerFieldBaseURL:
		return f.provider.BaseURL
	case providerFieldDefaultModel:
		return f.provider.DefaultModel
	case providerFieldAPIKeySource:
		if f.provider.APIKey.Source != "" {
			return string(f.provider.APIKey.Source)
		}
		return ""
	case providerFieldAPIKeyEnv:
		return f.provider.APIKey.EnvVar
	default:
		return ""
	}
}

func (f *ProviderForm) setFieldValue(field providerField, value string) error {
	switch field {
	case providerFieldID:
		if value == "" {
			return fmt.Errorf("provider id is required")
		}
		if f.providers != nil {
			for idx := range f.providers.Providers {
				if strings.EqualFold(f.providers.Providers[idx].ID, value) {
					if f.add || idx != f.index {
						return fmt.Errorf("provider id already exists")
					}
				}
			}
		}
		f.provider.ID = value
	case providerFieldKind:
		f.provider.Kind = core.ProviderKind(value)
	case providerFieldDisplayName:
		f.provider.DisplayName = value
	case providerFieldBaseURL:
		f.provider.BaseURL = value
	case providerFieldDefaultModel:
		f.provider.DefaultModel = value
	case providerFieldAPIKeySource:
		return f.setAPIKeySource(value)
	case providerFieldAPIKeyEnv:
		return f.setAPIKeyEnv(value)
	default:
		return fmt.Errorf("unknown field")
	}
	return nil
}

func (f *ProviderForm) setAPIKeySource(value string) error {
	switch strings.ToLower(value) {
	case "env":
		f.provider.APIKey.Source = core.APIKeySourceEnv
	case "secrets":
		f.provider.APIKey.Source = core.APIKeySourceSecrets
		f.provider.APIKey.EnvVar = ""
	default:
		return fmt.Errorf("api key source must be 'env' or 'secrets'")
	}
	return nil
}

func (f *ProviderForm) setAPIKeyEnv(value string) error {
	if f.provider.APIKey.Source == core.APIKeySourceEnv && value == "" {
		return fmt.Errorf("env var is required when source=env")
	}
	f.provider.APIKey.EnvVar = value
	return nil
}

// View renders the form UI.
func (f ProviderForm) View(palette theme.Theme, styles theme.Styles) string {
	if !f.active {
		return ""
	}

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(palette.Primary).
		Padding(1, 2).
		Width(70)

	title := "Edit Provider"
	if f.add {
		title = "Add Provider"
	}

	var currentName string
	if f.provider.DisplayName != "" {
		currentName = fmt.Sprintf(" (%s)", f.provider.DisplayName)
	}

	var body strings.Builder
	titleStyle := styles.Title
	body.WriteString(titleStyle.MarginBottom(0).Render(title + currentName))
	body.WriteString("\n\n")
	if f.fieldIdx < len(providerFieldSequence) {
		helpStyle := styles.Help
		body.WriteString(helpStyle.Italic(false).Render(
			fmt.Sprintf("Field: %s", f.promptFor(providerFieldSequence[f.fieldIdx])),
		))
	}
	body.WriteString("\n")
	body.WriteString(f.input.View())
	body.WriteString("\n\n")
	helpStyle2 := styles.Help
	body.WriteString(helpStyle2.Italic(false).Render("Enter to confirm  •  Esc to cancel"))
	if strings.TrimSpace(f.message) != "" {
		body.WriteString("\n")
		body.WriteString(helpStyle2.Italic(false).Render(f.message))
	}

	return boxStyle.Render(body.String())
}
