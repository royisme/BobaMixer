package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/royisme/bobamixer/internal/store/config"
	"gopkg.in/yaml.v3"
)

// WizardStage represents the current stage in the setup wizard
type WizardStage int

const (
	StageWelcome WizardStage = iota
	StageProviderSelect
	StageConfigForm
	StageSaving
	StageComplete
)

// Provider represents an AI provider configuration template
type Provider struct {
	Name         string
	DisplayName  string
	BaseURL      string
	DefaultModel string
	KeyEnvVar    string
	Description  string
}

var providers = []Provider{
	{
		Name:         "anthropic",
		DisplayName:  "Anthropic (Claude)",
		BaseURL:      "https://api.anthropic.com",
		DefaultModel: "claude-3-5-sonnet-20241022",
		KeyEnvVar:    "ANTHROPIC_API_KEY",
		Description:  "Claude 3.5 Sonnet - Most intelligent model",
	},
	{
		Name:         "openai",
		DisplayName:  "OpenAI (GPT)",
		BaseURL:      "https://api.openai.com/v1",
		DefaultModel: "gpt-4o",
		KeyEnvVar:    "OPENAI_API_KEY",
		Description:  "GPT-4o - Fast and capable",
	},
	{
		Name:         "gemini",
		DisplayName:  "Google (Gemini)",
		BaseURL:      "https://generativelanguage.googleapis.com",
		DefaultModel: "gemini-2.0-flash-exp",
		KeyEnvVar:    "GOOGLE_API_KEY",
		Description:  "Gemini 2.0 Flash - Experimental features",
	},
}

// WizardModel is the Bubble Tea model for the setup wizard
type WizardModel struct {
	stage           WizardStage
	home            string
	theme           Theme
	localizer       *Localizer

	// Provider selection
	providerList    list.Model
	selectedProvider *Provider

	// Configuration form
	inputs          []textinput.Model
	focusedInput    int

	// State
	width           int
	height          int
	err             error
	quitting        bool
}

// Form field indices
const (
	fieldProfileName = iota
	fieldAPIKey
	fieldBaseURL
	fieldModel
	fieldCount
)

// providerItem implements list.Item for the provider list
type providerItem struct {
	provider Provider
}

func (p providerItem) FilterValue() string {
	return p.provider.DisplayName
}

func (p providerItem) Title() string {
	return p.provider.DisplayName
}

func (p providerItem) Description() string {
	return p.provider.Description
}

// NewWizard creates a new setup wizard
func NewWizard(home string) (*WizardModel, error) {
	// Load theme and localizer
	theme := loadTheme(home)
	localizer, err := NewLocalizer(GetUserLanguage())
	if err != nil {
		localizer, _ = NewLocalizer("en")
	}

	// Create provider list
	items := make([]list.Item, len(providers))
	for i, p := range providers {
		items[i] = providerItem{provider: p}
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().
		Foreground(theme.Primary).
		BorderLeft(true).
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(theme.Primary).
		Padding(0, 0, 0, 1)

	delegate.Styles.SelectedDesc = lipgloss.NewStyle().
		Foreground(theme.Muted).
		BorderLeft(true).
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(theme.Primary).
		Padding(0, 0, 0, 1)

	providerList := list.New(items, delegate, 0, 0)
	providerList.Title = "Choose Your AI Provider"
	providerList.SetShowStatusBar(false)
	providerList.SetFilteringEnabled(false)
	providerList.SetShowHelp(false)

	return &WizardModel{
		stage:        StageWelcome,
		home:         home,
		theme:        theme,
		localizer:    localizer,
		providerList: providerList,
	}, nil
}

// Init initializes the wizard
func (m WizardModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the wizard state
func (m WizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			if m.stage == StageWelcome {
				m.quitting = true
				return m, tea.Quit
			}
			// Allow going back to previous stage
			if m.stage > StageWelcome {
				m.stage--
			}
			return m, nil

		case "enter":
			return m.handleEnter()

		case "tab", "shift+tab", "up", "down":
			if m.stage == StageConfigForm {
				return m.handleFormNavigation(msg.String())
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.providerList.SetSize(msg.Width-4, msg.Height-8)
		return m, nil
	}

	// Update active component based on stage
	switch m.stage {
	case StageProviderSelect:
		m.providerList, cmd = m.providerList.Update(msg)
		cmds = append(cmds, cmd)

	case StageConfigForm:
		cmds = m.updateFormInputs(msg)
	}

	return m, tea.Batch(cmds...)
}

// handleEnter processes Enter key based on current stage
func (m WizardModel) handleEnter() (tea.Model, tea.Cmd) {
	switch m.stage {
	case StageWelcome:
		m.stage = StageProviderSelect
		return m, nil

	case StageProviderSelect:
		// Get selected provider
		if item, ok := m.providerList.SelectedItem().(providerItem); ok {
			m.selectedProvider = &item.provider
			m.initializeForm()
			m.stage = StageConfigForm
		}
		return m, nil

	case StageConfigForm:
		// Validate and move to next field or save
		if m.validateCurrentField() {
			if m.focusedInput < fieldCount-1 {
				m.focusedInput++
				m.updateFormFocus()
			} else {
				// All fields filled, save configuration
				return m.saveConfiguration()
			}
		}
		return m, nil

	case StageComplete:
		m.quitting = true
		return m, tea.Quit
	}

	return m, nil
}

// handleFormNavigation handles Tab/Shift+Tab/Arrow keys in the form
func (m WizardModel) handleFormNavigation(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "tab", "down":
		m.focusedInput = (m.focusedInput + 1) % fieldCount
	case "shift+tab", "up":
		m.focusedInput--
		if m.focusedInput < 0 {
			m.focusedInput = fieldCount - 1
		}
	}
	m.updateFormFocus()
	return m, nil
}

// initializeForm creates and configures the form inputs
func (m *WizardModel) initializeForm() {
	m.inputs = make([]textinput.Model, fieldCount)

	// Profile Name
	m.inputs[fieldProfileName] = textinput.New()
	m.inputs[fieldProfileName].Placeholder = "e.g., default, work, personal"
	m.inputs[fieldProfileName].CharLimit = 50
	m.inputs[fieldProfileName].Width = 50
	m.inputs[fieldProfileName].Prompt = "│ "
	m.inputs[fieldProfileName].Focus()

	// API Key
	m.inputs[fieldAPIKey] = textinput.New()
	m.inputs[fieldAPIKey].Placeholder = fmt.Sprintf("Your %s API key", m.selectedProvider.DisplayName)
	m.inputs[fieldAPIKey].CharLimit = 200
	m.inputs[fieldAPIKey].Width = 50
	m.inputs[fieldAPIKey].Prompt = "│ "
	m.inputs[fieldAPIKey].EchoMode = textinput.EchoPassword
	m.inputs[fieldAPIKey].EchoCharacter = '•'

	// Base URL (optional, pre-filled with default)
	m.inputs[fieldBaseURL] = textinput.New()
	m.inputs[fieldBaseURL].Placeholder = "Optional (press Enter to use default)"
	m.inputs[fieldBaseURL].SetValue(m.selectedProvider.BaseURL)
	m.inputs[fieldBaseURL].CharLimit = 200
	m.inputs[fieldBaseURL].Width = 50
	m.inputs[fieldBaseURL].Prompt = "│ "

	// Model (optional, pre-filled with default)
	m.inputs[fieldModel] = textinput.New()
	m.inputs[fieldModel].Placeholder = "Optional (press Enter to use default)"
	m.inputs[fieldModel].SetValue(m.selectedProvider.DefaultModel)
	m.inputs[fieldModel].CharLimit = 100
	m.inputs[fieldModel].Width = 50
	m.inputs[fieldModel].Prompt = "│ "

	m.focusedInput = 0
}

// updateFormFocus updates which input has focus
func (m *WizardModel) updateFormFocus() {
	for i := range m.inputs {
		if i == m.focusedInput {
			m.inputs[i].Focus()
		} else {
			m.inputs[i].Blur()
		}
	}
}

// updateFormInputs updates all form inputs and returns their commands
func (m *WizardModel) updateFormInputs(msg tea.Msg) []tea.Cmd {
	cmds := make([]tea.Cmd, fieldCount)
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}
	return cmds
}

// validateCurrentField validates the currently focused field
func (m *WizardModel) validateCurrentField() bool {
	switch m.focusedInput {
	case fieldProfileName:
		return strings.TrimSpace(m.inputs[fieldProfileName].Value()) != ""
	case fieldAPIKey:
		return strings.TrimSpace(m.inputs[fieldAPIKey].Value()) != ""
	case fieldBaseURL, fieldModel:
		// These are optional
		return true
	}
	return false
}

// saveConfiguration saves the configuration and transitions to complete stage
func (m *WizardModel) saveConfiguration() (tea.Model, tea.Cmd) {
	m.stage = StageSaving

	profileName := strings.TrimSpace(m.inputs[fieldProfileName].Value())
	apiKey := strings.TrimSpace(m.inputs[fieldAPIKey].Value())
	baseURL := strings.TrimSpace(m.inputs[fieldBaseURL].Value())
	model := strings.TrimSpace(m.inputs[fieldModel].Value())

	// Use defaults if not provided
	if baseURL == "" {
		baseURL = m.selectedProvider.BaseURL
	}
	if model == "" {
		model = m.selectedProvider.DefaultModel
	}

	// Create profile configuration
	profile := config.Profile{
		Name:     profileName,
		Provider: m.selectedProvider.Name,
		Model:    model,
		Endpoint: baseURL,
	}

	// Save profile to profiles.yaml
	profilesPath := filepath.Join(m.home, "profiles.yaml")
	profilesData := map[string]interface{}{
		"profiles": map[string]interface{}{
			profileName: map[string]interface{}{
				"provider": profile.Provider,
				"model":    profile.Model,
				"endpoint": profile.Endpoint,
			},
		},
	}

	data, err := yaml.Marshal(profilesData)
	if err != nil {
		m.err = fmt.Errorf("failed to marshal profiles: %w", err)
		return m, nil
	}

	if err := os.WriteFile(profilesPath, data, 0644); err != nil {
		m.err = fmt.Errorf("failed to write profiles.yaml: %w", err)
		return m, nil
	}

	// Save API key to secrets.yaml
	secretsPath := filepath.Join(m.home, "secrets.yaml")
	secretsData := map[string]interface{}{
		m.selectedProvider.Name: apiKey,
	}

	data, err = yaml.Marshal(secretsData)
	if err != nil {
		m.err = fmt.Errorf("failed to marshal secrets: %w", err)
		return m, nil
	}

	if err := os.WriteFile(secretsPath, data, 0600); err != nil {
		m.err = fmt.Errorf("failed to write secrets.yaml: %w", err)
		return m, nil
	}

	// Set as active profile
	if err := config.SaveActiveProfile(m.home, profileName); err != nil {
		m.err = fmt.Errorf("failed to set active profile: %w", err)
		return m, nil
	}

	m.stage = StageComplete
	return m, nil
}

// View renders the wizard UI
func (m WizardModel) View() string {
	if m.quitting {
		return ""
	}

	switch m.stage {
	case StageWelcome:
		return m.viewWelcome()
	case StageProviderSelect:
		return m.viewProviderSelect()
	case StageConfigForm:
		return m.viewConfigForm()
	case StageSaving:
		return m.viewSaving()
	case StageComplete:
		return m.viewComplete()
	}

	return ""
}

// viewWelcome renders the welcome screen
func (m WizardModel) viewWelcome() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.Primary).
		Padding(1, 0).
		Align(lipgloss.Center)

	subtitleStyle := lipgloss.NewStyle().
		Foreground(m.theme.Muted).
		Italic(true).
		Align(lipgloss.Center)

	promptStyle := lipgloss.NewStyle().
		Foreground(m.theme.Text).
		Padding(1, 0).
		Align(lipgloss.Center)

	helpStyle := lipgloss.NewStyle().
		Foreground(m.theme.Muted).
		Padding(2, 0).
		Align(lipgloss.Center)

	var content strings.Builder
	content.WriteString("\n\n\n")
	content.WriteString(titleStyle.Render("🧋 Welcome to BobaMixer!"))
	content.WriteString("\n\n")
	content.WriteString(subtitleStyle.Render("Smart AI Adapter Router"))
	content.WriteString("\n\n")
	content.WriteString(promptStyle.Render("Let's get you set up in just a few steps"))
	content.WriteString("\n\n")
	content.WriteString(helpStyle.Render("Press Enter to continue • Esc to exit"))

	return content.String()
}

// viewProviderSelect renders the provider selection screen
func (m WizardModel) viewProviderSelect() string {
	helpStyle := lipgloss.NewStyle().
		Foreground(m.theme.Muted).
		Padding(1, 2)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		"\n",
		m.providerList.View(),
		helpStyle.Render("↑/↓: Navigate • Enter: Select • Esc: Back"),
	)
}

// viewConfigForm renders the configuration form
func (m WizardModel) viewConfigForm() string {
	labelStyle := lipgloss.NewStyle().
		Foreground(m.theme.Primary).
		Bold(true).
		Width(20)

	helpTextStyle := lipgloss.NewStyle().
		Foreground(m.theme.Muted).
		Italic(true).
		MarginLeft(2)

	inputBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.Border).
		Padding(0, 1).
		Width(54)

	focusedInputBoxStyle := inputBoxStyle.Copy().
		BorderForeground(m.theme.Primary)

	var content strings.Builder
	content.WriteString("\n")
	content.WriteString(lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.Primary).
		Render(fmt.Sprintf("Configure %s Profile", m.selectedProvider.DisplayName)))
	content.WriteString("\n\n")

	// Profile Name
	content.WriteString(labelStyle.Render("Profile Name"))
	content.WriteString("\n")
	if m.focusedInput == fieldProfileName {
		content.WriteString(focusedInputBoxStyle.Render(m.inputs[fieldProfileName].View()))
	} else {
		content.WriteString(inputBoxStyle.Render(m.inputs[fieldProfileName].View()))
	}
	content.WriteString("\n")
	content.WriteString(helpTextStyle.Render("A friendly name for this configuration"))
	content.WriteString("\n\n")

	// API Key
	content.WriteString(labelStyle.Render("API Key"))
	content.WriteString("\n")
	if m.focusedInput == fieldAPIKey {
		content.WriteString(focusedInputBoxStyle.Render(m.inputs[fieldAPIKey].View()))
	} else {
		content.WriteString(inputBoxStyle.Render(m.inputs[fieldAPIKey].View()))
	}
	content.WriteString("\n")
	keyURL := m.getProviderKeyURL()
	content.WriteString(helpTextStyle.Render(fmt.Sprintf("Get your key at: %s", keyURL)))
	content.WriteString("\n\n")

	// Base URL (optional)
	content.WriteString(labelStyle.Render("API Base URL"))
	content.WriteString("\n")
	if m.focusedInput == fieldBaseURL {
		content.WriteString(focusedInputBoxStyle.Render(m.inputs[fieldBaseURL].View()))
	} else {
		content.WriteString(inputBoxStyle.Render(m.inputs[fieldBaseURL].View()))
	}
	content.WriteString("\n")
	content.WriteString(helpTextStyle.Render("Optional: Use custom endpoint or leave default"))
	content.WriteString("\n\n")

	// Model (optional)
	content.WriteString(labelStyle.Render("Default Model"))
	content.WriteString("\n")
	if m.focusedInput == fieldModel {
		content.WriteString(focusedInputBoxStyle.Render(m.inputs[fieldModel].View()))
	} else {
		content.WriteString(inputBoxStyle.Render(m.inputs[fieldModel].View()))
	}
	content.WriteString("\n")
	content.WriteString(helpTextStyle.Render("Optional: Override the default model"))
	content.WriteString("\n\n")

	// Help text
	helpStyle := lipgloss.NewStyle().Foreground(m.theme.Muted)
	content.WriteString(helpStyle.Render("Tab/↑↓: Navigate • Enter: Continue/Finish • Esc: Back"))

	return content.String()
}

// getProviderKeyURL returns the URL where users can get API keys
func (m WizardModel) getProviderKeyURL() string {
	switch m.selectedProvider.Name {
	case "anthropic":
		return "console.anthropic.com"
	case "openai":
		return "platform.openai.com/api-keys"
	case "gemini":
		return "makersuite.google.com/app/apikey"
	default:
		return "provider's website"
	}
}

// viewSaving renders the saving screen
func (m WizardModel) viewSaving() string {
	if m.err != nil {
		errorStyle := lipgloss.NewStyle().
			Foreground(m.theme.Warning).
			Bold(true).
			Padding(1, 2)

		return errorStyle.Render(fmt.Sprintf("❌ Error: %v\n\nPress Esc to go back", m.err))
	}

	savingStyle := lipgloss.NewStyle().
		Foreground(m.theme.Primary).
		Padding(1, 2)

	return savingStyle.Render("💾 Saving configuration...")
}

// viewComplete renders the completion screen
func (m WizardModel) viewComplete() string {
	successStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.Success).
		Padding(1, 0).
		Align(lipgloss.Center)

	infoStyle := lipgloss.NewStyle().
		Foreground(m.theme.Text).
		Padding(1, 2).
		Align(lipgloss.Center)

	helpStyle := lipgloss.NewStyle().
		Foreground(m.theme.Muted).
		Padding(2, 0).
		Align(lipgloss.Center)

	profileName := m.inputs[fieldProfileName].Value()

	var content strings.Builder
	content.WriteString("\n\n\n")
	content.WriteString(successStyle.Render("✓ Setup Complete!"))
	content.WriteString("\n\n")
	content.WriteString(infoStyle.Render(fmt.Sprintf(
		"Your profile '%s' is ready to use\nConfiguration saved to %s",
		profileName,
		m.home,
	)))
	content.WriteString("\n\n")
	content.WriteString(helpStyle.Render("Press Enter to launch BobaMixer"))

	return content.String()
}

// RunWizard starts the setup wizard and returns whether to continue to main TUI
func RunWizard(home string) (bool, error) {
	wizard, err := NewWizard(home)
	if err != nil {
		return false, fmt.Errorf("failed to create wizard: %w", err)
	}

	p := tea.NewProgram(wizard, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return false, fmt.Errorf("wizard error: %w", err)
	}

	// Check if wizard completed successfully
	if m, ok := finalModel.(WizardModel); ok {
		return m.stage == StageComplete, nil
	}

	return false, nil
}
