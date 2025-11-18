// Package forms provides interactive forms for managing providers, bindings, and secrets in the BobaMixer TUI.
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

type bindingField int

const (
	bindingFieldToolID bindingField = iota
	bindingFieldProviderID
	bindingFieldModel
	bindingFieldUseProxy
)

var bindingFieldSequence = []bindingField{
	bindingFieldToolID,
	bindingFieldProviderID,
	bindingFieldModel,
	bindingFieldUseProxy,
}

// BindingForm manages tool-provider binding input flow.
type BindingForm struct {
	active    bool
	add       bool
	index     int
	fieldIdx  int
	input     textinput.Model
	binding   core.Binding
	message   string
	prompt    string
	bindings  *core.BindingsConfig
	tools     *core.ToolsConfig
	providers *core.ProvidersConfig
}

// NewBindingForm creates a binding form with the provided prompt.
func NewBindingForm(prompt string) BindingForm {
	input := textinput.New()
	input.CharLimit = 200
	input.Width = 40
	input.Prompt = prompt

	return BindingForm{
		input:  input,
		prompt: prompt,
	}
}

func (f BindingForm) Active() bool {
	return f.active
}

func (f BindingForm) Message() string {
	return f.message
}

func (f *BindingForm) SetMessage(msg string) {
	f.message = msg
}

func (f BindingForm) Binding() core.Binding {
	return f.binding
}

func (f BindingForm) Index() int {
	return f.index
}

func (f BindingForm) AddMode() bool {
	return f.add
}

func (f *BindingForm) Start(
	add bool,
	binding core.Binding,
	idx int,
	bindings *core.BindingsConfig,
	tools *core.ToolsConfig,
	providers *core.ProvidersConfig,
) {
	f.add = add
	f.index = idx
	f.fieldIdx = 0
	f.active = true
	f.message = ""
	f.bindings = bindings
	f.tools = tools
	f.providers = providers

	if add {
		f.binding = core.Binding{
			UseProxy: true,
			Options:  core.BindingOptions{},
		}
		f.index = -1
	} else {
		f.binding = binding
	}

	f.skipDisabledFields()
	if f.fieldIdx >= len(bindingFieldSequence) {
		f.active = false
		f.input.Blur()
		return
	}

	f.prepareField()
}

func (f *BindingForm) Cancel(reason string) {
	f.active = false
	f.message = reason
	f.input.Blur()
	f.input.SetValue("")
}

func (f *BindingForm) Update(msg tea.Msg) tea.Cmd {
	if !f.active {
		return nil
	}
	var cmd tea.Cmd
	f.input, cmd = f.input.Update(msg)
	return cmd
}

func (f *BindingForm) Submit() (bool, error) {
	if !f.active || f.fieldIdx >= len(bindingFieldSequence) {
		return false, nil
	}

	field := bindingFieldSequence[f.fieldIdx]
	value := strings.TrimSpace(f.input.Value())
	if err := f.setFieldValue(field, value); err != nil {
		f.message = err.Error()
		return false, err
	}

	f.message = ""
	f.input.SetValue("")
	f.fieldIdx++
	f.skipDisabledFields()

	if f.fieldIdx >= len(bindingFieldSequence) {
		f.active = false
		f.input.Blur()
		return true, nil
	}

	f.prepareField()
	return false, nil
}

func (f *BindingForm) fieldEnabled(field bindingField) bool {
	return field != bindingFieldToolID || f.add
}

func (f *BindingForm) skipDisabledFields() {
	for f.fieldIdx < len(bindingFieldSequence) && !f.fieldEnabled(bindingFieldSequence[f.fieldIdx]) {
		f.fieldIdx++
	}
}

func (f *BindingForm) prepareField() {
	if f.fieldIdx >= len(bindingFieldSequence) {
		return
	}

	field := bindingFieldSequence[f.fieldIdx]
	f.input.Placeholder = f.promptFor(field)
	f.input.SetValue(f.valueForField(field))
	f.input.CursorEnd()
	f.input.Focus()
}

func (f *BindingForm) promptFor(field bindingField) string {
	switch field {
	case bindingFieldToolID:
		return "tool id"
	case bindingFieldProviderID:
		return "provider id"
	case bindingFieldModel:
		return "model override"
	case bindingFieldUseProxy:
		return "use proxy (on/off)"
	default:
		return "value"
	}
}

func (f *BindingForm) valueForField(field bindingField) string {
	switch field {
	case bindingFieldToolID:
		return f.binding.ToolID
	case bindingFieldProviderID:
		return f.binding.ProviderID
	case bindingFieldModel:
		return f.binding.Options.Model
	case bindingFieldUseProxy:
		if f.binding.UseProxy {
			return "on"
		}
		return "off"
	default:
		return ""
	}
}

func (f *BindingForm) setFieldValue(field bindingField, value string) error {
	switch field {
	case bindingFieldToolID:
		return f.setToolID(value)
	case bindingFieldProviderID:
		return f.setProviderID(value)
	case bindingFieldModel:
		f.binding.Options.Model = value
		return nil
	case bindingFieldUseProxy:
		return f.setUseProxy(value)
	default:
		return fmt.Errorf("unknown field")
	}
}

func (f *BindingForm) setToolID(value string) error {
	if value == "" {
		return fmt.Errorf("tool id is required")
	}
	if f.tools != nil {
		if _, err := f.tools.FindTool(value); err != nil {
			return fmt.Errorf("tool %s not found", value)
		}
	}
	if f.bindings != nil {
		for idx := range f.bindings.Bindings {
			if strings.EqualFold(f.bindings.Bindings[idx].ToolID, value) {
				if f.add || idx != f.index {
					return fmt.Errorf("binding for %s already exists", value)
				}
			}
		}
	}
	f.binding.ToolID = value
	return nil
}

func (f *BindingForm) setProviderID(value string) error {
	if value == "" {
		return fmt.Errorf("provider id is required")
	}
	if f.providers != nil {
		if _, err := f.providers.FindProvider(value); err != nil {
			return fmt.Errorf("provider %s not found", value)
		}
	}
	f.binding.ProviderID = value
	return nil
}

func (f *BindingForm) setUseProxy(value string) error {
	lower := strings.ToLower(value)
	switch lower {
	case "on", "true", "yes":
		f.binding.UseProxy = true
	case "off", "false", "no":
		f.binding.UseProxy = false
	default:
		return fmt.Errorf("proxy value must be on/off")
	}
	return nil
}

func (f BindingForm) View(palette theme.Theme, styles theme.Styles) string {
	if !f.active {
		return ""
	}

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(palette.Primary).
		Padding(1, 2).
		Width(70)

	title := "Edit Binding"
	if f.add {
		title = "Add Binding"
	}

	body := strings.Builder{}
	titleStyle := styles.Title
	body.WriteString(titleStyle.MarginBottom(0).Render(fmt.Sprintf("%s (%s)", title, f.binding.ToolID)))
	body.WriteString("\n\n")
	if f.fieldIdx < len(bindingFieldSequence) {
		helpStyle := styles.Help
		body.WriteString(helpStyle.Italic(false).Render(
			fmt.Sprintf("Field: %s", f.promptFor(bindingFieldSequence[f.fieldIdx])),
		))
		body.WriteString("\n")
	}
	body.WriteString(f.input.View())
	body.WriteString("\n\n")
	helpStyle := styles.Help
	body.WriteString(helpStyle.Italic(false).Render("Enter to confirm  •  Esc to cancel"))
	if strings.TrimSpace(f.message) != "" {
		body.WriteString("\n")
		body.WriteString(helpStyle.Italic(false).Render(f.message))
	}

	return boxStyle.Render(body.String())
}
