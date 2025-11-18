package forms

import (
	"errors"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/royisme/bobamixer/internal/ui/theme"
)

// SecretForm manages API key input per provider.
type SecretForm struct {
	active       bool
	targetIndex  int
	providerName string
	input        textinput.Model
	message      string
	prompt       string
}

// NewSecretForm creates a secret form configured with prompt prefix.
func NewSecretForm(prompt string) SecretForm {
	input := textinput.New()
	input.Placeholder = "Enter API key"
	input.CharLimit = 200
	input.Width = 40
	input.Prompt = prompt
	input.EchoMode = textinput.EchoPassword
	input.EchoCharacter = '•'

	return SecretForm{
		input:  input,
		prompt: prompt,
	}
}

func (f SecretForm) Active() bool {
	return f.active
}

func (f SecretForm) Message() string {
	return f.message
}

func (f *SecretForm) SetMessage(msg string) {
	f.message = msg
}

func (f SecretForm) TargetIndex() int {
	return f.targetIndex
}

// Start activates the form for the given provider.
func (f *SecretForm) Start(targetIdx int, providerName string) {
	f.targetIndex = targetIdx
	f.providerName = providerName
	f.message = ""
	f.input.Placeholder = "API key for " + providerName
	f.input.SetValue("")
	f.input.CursorEnd()
	f.input.Focus()
	f.active = true
}

// Cancel terminates the form.
func (f *SecretForm) Cancel(reason string) {
	f.message = reason
	f.active = false
	f.input.Blur()
	f.input.SetValue("")
}

// Update forwards events to the text input.
func (f *SecretForm) Update(msg tea.Msg) tea.Cmd {
	if !f.active {
		return nil
	}
	var cmd tea.Cmd
	f.input, cmd = f.input.Update(msg)
	return cmd
}

// Submit returns the API key and deactivates the form.
func (f *SecretForm) Submit() (string, error) {
	value := strings.TrimSpace(f.input.Value())
	if value == "" {
		f.message = "API key cannot be empty"
		return "", ErrEmptySecret
	}
	f.active = false
	f.input.Blur()
	f.input.SetValue("")
	return value, nil
}

// View renders the secret form.
func (f SecretForm) View(styles theme.Styles, palette theme.Theme) string {
	if !f.active {
		return ""
	}

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(palette.Primary).
		Padding(1, 2).
		Width(60)

	body := strings.Builder{}
	titleStyle := styles.Title
	body.WriteString(titleStyle.MarginBottom(0).Render("Set API key for " + f.providerName))
	body.WriteString("\n\n")
	body.WriteString(f.input.View())
	body.WriteString("\n\n")
	helpStyle := styles.Help
	body.WriteString(helpStyle.Italic(false).Render("Enter to save  •  Esc to cancel"))
	if strings.TrimSpace(f.message) != "" {
		body.WriteString("\n")
		body.WriteString(helpStyle.Italic(false).Render(f.message))
	}

	return boxStyle.Render(body.String())
}

// ErrEmptySecret is returned when submitting an empty secret.
var ErrEmptySecret = errors.New("api key cannot be empty")
