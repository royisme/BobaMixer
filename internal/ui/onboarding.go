package ui

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/royisme/bobamixer/internal/domain/core"
)

// OnboardingStage represents the current stage in the setup wizard
type OnboardingStage int

const (
	StageWelcome OnboardingStage = iota
	StageScanning
	StageScanResults
	StageProviderSelect
	StageAPIKeyInput
	StageComplete
)

// ToolDetection represents the result of scanning for a tool
type ToolDetection struct {
	Tool       core.Tool
	Found      bool
	Path       string
	ConfigPath string
}

// OnboardingModel is the Bubble Tea model for the onboarding wizard
type OnboardingModel struct {
	stage     OnboardingStage
	home      string
	theme     Theme
	localizer *Localizer

	// Scanning state
	spinner    spinner.Model
	scanning   bool
	detections []ToolDetection

	// Provider selection
	providerList     list.Model
	providers        []core.Provider
	selectedProvider *core.Provider

	// API key input
	apiKeyInput   textinput.Model
	apiKeyFromEnv bool
	apiKeyValue   string

	// Dimensions
	width  int
	height int

	// State
	err      error
	quitting bool
}

// Messages for async operations
type scanCompleteMsg struct {
	detections []ToolDetection
}

type providerItem struct {
	provider core.Provider
}

func (p providerItem) FilterValue() string {
	return p.provider.DisplayName
}

func (p providerItem) Title() string {
	return p.provider.DisplayName
}

func (p providerItem) Description() string {
	baseURL := p.provider.BaseURL
	if len(baseURL) > 50 {
		baseURL = baseURL[:47] + "..."
	}
	return fmt.Sprintf("Model: %s • %s", p.provider.DefaultModel, baseURL)
}

// NewOnboarding creates a new onboarding wizard
func NewOnboarding(home string) (*OnboardingModel, error) {
	// Load theme and localizer
	theme := loadTheme(home)
	localizer, err := NewLocalizer(GetUserLanguage())
	if err != nil {
		localizer, _ = NewLocalizer("en")
	}

	// Create spinner for scanning
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(theme.Primary)

	return &OnboardingModel{
		stage:     StageWelcome,
		home:      home,
		theme:     theme,
		localizer: localizer,
		spinner:   s,
	}, nil
}

// Init initializes the onboarding wizard
func (m OnboardingModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the wizard state
//
//nolint:gocyclo // Bubble Tea Update function handles multiple stages and message types
func (m OnboardingModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "esc":
			// Allow going back in most stages
			if m.stage > StageWelcome && m.stage < StageComplete {
				m.stage--
			}
			return m, nil

		case "enter":
			return m.handleEnter()

		case "y", "n":
			if m.stage == StageAPIKeyInput && m.apiKeyFromEnv {
				if msg.String() == "y" {
					// Use env key
					m.stage = StageComplete
					return m.saveConfiguration()
				} else {
					// Enter different key
					m.apiKeyFromEnv = false
					m.initializeAPIKeyInput()
				}
				return m, nil
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.providerList.Items() != nil {
			m.providerList.SetSize(msg.Width-4, msg.Height-10)
		}
		return m, nil

	case scanCompleteMsg:
		m.detections = msg.detections
		m.scanning = false
		m.stage = StageScanResults
		return m, nil

	case spinner.TickMsg:
		if m.scanning {
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	// Update active component based on stage
	switch m.stage {
	case StageProviderSelect:
		m.providerList, cmd = m.providerList.Update(msg)
		cmds = append(cmds, cmd)

	case StageAPIKeyInput:
		if !m.apiKeyFromEnv {
			m.apiKeyInput, cmd = m.apiKeyInput.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

// handleEnter processes Enter key based on current stage
func (m OnboardingModel) handleEnter() (tea.Model, tea.Cmd) {
	switch m.stage {
	case StageWelcome:
		// Start scanning
		m.stage = StageScanning
		m.scanning = true
		return m, tea.Batch(m.spinner.Tick, m.scanTools)

	case StageScanResults:
		// Check if any tools were found
		foundAny := false
		for _, det := range m.detections {
			if det.Found {
				foundAny = true
				break
			}
		}

		if !foundAny {
			// No tools found, show error and quit
			m.err = fmt.Errorf("no supported CLI tools found")
			return m, tea.Quit
		}

		// Move to provider selection
		m.stage = StageProviderSelect
		m.initializeProviderList()
		return m, nil

	case StageProviderSelect:
		// Get selected provider
		if item, ok := m.providerList.SelectedItem().(providerItem); ok {
			m.selectedProvider = &item.provider
			m.stage = StageAPIKeyInput
			m.checkAPIKey()
		}
		return m, nil

	case StageAPIKeyInput:
		if m.apiKeyFromEnv {
			// Already handled by y/n keys
			return m, nil
		}

		// Validate API key input
		key := strings.TrimSpace(m.apiKeyInput.Value())
		if key == "" {
			return m, nil // Don't proceed with empty key
		}

		m.apiKeyValue = key
		m.stage = StageComplete
		return m.saveConfiguration()

	case StageComplete:
		m.quitting = true
		return m, tea.Quit
	}

	return m, nil
}

// scanTools scans the system for CLI tools
func (m OnboardingModel) scanTools() tea.Msg {
	// Define tools to scan for (Phase 1.5: Claude, Codex, Gemini)
	toolsToScan := []core.Tool{
		{
			ID:          "claude",
			Name:        "Claude Code CLI",
			Exec:        "claude",
			Kind:        core.ToolKindClaude,
			ConfigType:  core.ConfigTypeClaudeSettingsJSON,
			ConfigPath:  "~/.claude/settings.json",
			Description: "Claude Code CLI for AI-assisted coding",
		},
		{
			ID:          "codex",
			Name:        "OpenAI Codex CLI",
			Exec:        "codex",
			Kind:        core.ToolKindCodex,
			ConfigType:  core.ConfigTypeCodexConfigTOML,
			ConfigPath:  "~/.codex/config.toml",
			Description: "OpenAI Codex CLI for AI-powered development",
		},
		{
			ID:          "gemini",
			Name:        "Google Gemini CLI",
			Exec:        "gemini",
			Kind:        core.ToolKindGemini,
			ConfigType:  core.ConfigTypeGeminiSettingsJSON,
			ConfigPath:  "~/.gemini/settings.json",
			Description: "Google Gemini CLI for multimodal AI assistance",
		},
	}

	detections := make([]ToolDetection, 0, len(toolsToScan))

	for _, tool := range toolsToScan {
		path, err := exec.LookPath(tool.Exec)
		detection := ToolDetection{
			Tool:       tool,
			Found:      err == nil,
			Path:       path,
			ConfigPath: tool.ConfigPath,
		}
		detections = append(detections, detection)
	}

	return scanCompleteMsg{detections: detections}
}

// initializeProviderList creates the provider selection list
func (m *OnboardingModel) initializeProviderList() {
	// Load providers from config or use defaults
	providersConfig, err := core.LoadProviders(m.home)
	if err != nil || len(providersConfig.Providers) == 0 {
		// Use default providers for all detected tools
		m.providers = []core.Provider{
			{
				ID:          "claude-anthropic-official",
				Kind:        core.ProviderKindAnthropic,
				DisplayName: "Anthropic (Official)",
				BaseURL:     "https://api.anthropic.com",
				APIKey: core.APIKeyConfig{
					Source: core.APIKeySourceEnv,
					EnvVar: "ANTHROPIC_API_KEY",
				},
				DefaultModel: "claude-3-5-sonnet-20241022",
				Enabled:      true,
			},
			{
				ID:          "claude-zai",
				Kind:        core.ProviderKindAnthropicCompatible,
				DisplayName: "Claude via Z.AI",
				BaseURL:     "https://api.z.ai/api/anthropic",
				APIKey: core.APIKeyConfig{
					Source: core.APIKeySourceEnv,
					EnvVar: "ANTHROPIC_AUTH_TOKEN",
				},
				DefaultModel: "glm-4.6",
				Enabled:      true,
			},
			{
				ID:          "openai-official",
				Kind:        core.ProviderKindOpenAI,
				DisplayName: "OpenAI (Official)",
				BaseURL:     "https://api.openai.com/v1",
				APIKey: core.APIKeyConfig{
					Source: core.APIKeySourceEnv,
					EnvVar: "OPENAI_API_KEY",
				},
				DefaultModel: "gpt-4-turbo-preview",
				Enabled:      true,
			},
			{
				ID:          "gemini-official",
				Kind:        core.ProviderKindGemini,
				DisplayName: "Google Gemini (Official)",
				BaseURL:     "https://generativelanguage.googleapis.com/v1",
				APIKey: core.APIKeyConfig{
					Source: core.APIKeySourceEnv,
					EnvVar: "GEMINI_API_KEY",
				},
				DefaultModel: "gemini-1.5-pro",
				Enabled:      true,
			},
		}
	} else {
		// Use loaded providers, include all supported provider kinds
		for _, p := range providersConfig.Providers {
			if p.Enabled {
				m.providers = append(m.providers, p)
			}
		}
	}

	// Create list items
	items := make([]list.Item, len(m.providers))
	for i, p := range m.providers {
		items[i] = providerItem{provider: p}
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().
		Foreground(m.theme.Primary).
		BorderLeft(true).
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(m.theme.Primary).
		Padding(0, 0, 0, 1)

	delegate.Styles.SelectedDesc = lipgloss.NewStyle().
		Foreground(m.theme.Muted).
		BorderLeft(true).
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(m.theme.Primary).
		Padding(0, 0, 0, 1)

	m.providerList = list.New(items, delegate, 0, 0)
	m.providerList.Title = "Choose Your AI Provider"
	m.providerList.SetShowStatusBar(false)
	m.providerList.SetFilteringEnabled(false)
	m.providerList.SetShowHelp(false)

	if m.width > 0 {
		m.providerList.SetSize(m.width-4, m.height-10)
	}
}

// checkAPIKey checks if API key is available in environment
func (m *OnboardingModel) checkAPIKey() {
	secrets, err := core.LoadSecrets(m.home)
	if err != nil {
		// Secrets file doesn't exist yet, use empty config
		secrets = &core.SecretsConfig{}
	}
	if key, err := core.ResolveAPIKey(m.selectedProvider, secrets); err == nil {
		// Key found in env
		m.apiKeyFromEnv = true
		m.apiKeyValue = key
	} else {
		// Need to input key
		m.apiKeyFromEnv = false
		m.initializeAPIKeyInput()
	}
}

// initializeAPIKeyInput creates the API key input field
func (m *OnboardingModel) initializeAPIKeyInput() {
	ti := textinput.New()
	ti.Placeholder = fmt.Sprintf("Your %s API key", m.selectedProvider.DisplayName)
	ti.CharLimit = 200
	ti.Width = 50
	ti.Prompt = "│ "
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '•'
	ti.Focus()

	m.apiKeyInput = ti
}

// saveConfiguration saves the configuration and creates bindings
//
//nolint:gocyclo // Configuration saving involves multiple files and validation steps
func (m *OnboardingModel) saveConfiguration() (tea.Model, tea.Cmd) {
	// Initialize default configs if they don't exist
	if err := core.InitDefaultConfigs(m.home); err != nil {
		m.err = fmt.Errorf("failed to initialize configs: %w", err)
		return m, nil
	}

	// Load configs
	providers, err := core.LoadProviders(m.home)
	if err != nil {
		m.err = fmt.Errorf("failed to load providers: %w", err)
		return m, nil
	}

	tools, err := core.LoadTools(m.home)
	if err != nil {
		m.err = fmt.Errorf("failed to load tools: %w", err)
		return m, nil
	}

	bindings, err := core.LoadBindings(m.home)
	if err != nil {
		m.err = fmt.Errorf("failed to load bindings: %w", err)
		return m, nil
	}

	secrets, err := core.LoadSecrets(m.home)
	if err != nil {
		m.err = fmt.Errorf("failed to load secrets: %w", err)
		return m, nil
	}

	// Add provider if it doesn't exist
	providerExists := false
	for _, p := range providers.Providers {
		if p.ID == m.selectedProvider.ID {
			providerExists = true
			break
		}
	}
	if !providerExists {
		providers.Providers = append(providers.Providers, *m.selectedProvider)
		if err := core.SaveProviders(m.home, providers); err != nil {
			m.err = fmt.Errorf("failed to save providers: %w", err)
			return m, nil
		}
	}

	// Add detected tools if they don't exist
	for _, det := range m.detections {
		if !det.Found {
			continue
		}

		toolExists := false
		for _, t := range tools.Tools {
			if t.ID == det.Tool.ID {
				toolExists = true
				break
			}
		}

		if !toolExists {
			tools.Tools = append(tools.Tools, det.Tool)
		}
	}
	if err := core.SaveTools(m.home, tools); err != nil {
		m.err = fmt.Errorf("failed to save tools: %w", err)
		return m, nil
	}

	// Create bindings for detected tools
	for _, det := range m.detections {
		if !det.Found {
			continue
		}

		// Check if binding exists
		bindingExists := false
		for i, b := range bindings.Bindings {
			if b.ToolID == det.Tool.ID {
				// Update existing binding
				bindings.Bindings[i].ProviderID = m.selectedProvider.ID
				bindingExists = true
				break
			}
		}

		if !bindingExists {
			// Create new binding
			bindings.Bindings = append(bindings.Bindings, core.Binding{
				ToolID:     det.Tool.ID,
				ProviderID: m.selectedProvider.ID,
				UseProxy:   false,
				Options:    core.BindingOptions{},
			})
		}
	}
	if err := core.SaveBindings(m.home, bindings); err != nil {
		m.err = fmt.Errorf("failed to save bindings: %w", err)
		return m, nil
	}

	// Save API key if not from env
	if !m.apiKeyFromEnv && m.apiKeyValue != "" {
		if secrets.Secrets == nil {
			secrets.Secrets = make(map[string]core.Secret)
		}
		secrets.Secrets[m.selectedProvider.ID] = core.Secret{
			APIKey: m.apiKeyValue,
		}
		if err := core.SaveSecrets(m.home, secrets); err != nil {
			m.err = fmt.Errorf("failed to save secrets: %w", err)
			return m, nil
		}
	}

	return m, nil
}

// View renders the wizard UI
func (m OnboardingModel) View() string {
	if m.quitting {
		return ""
	}

	switch m.stage {
	case StageWelcome:
		return m.viewWelcome()
	case StageScanning:
		return m.viewScanning()
	case StageScanResults:
		return m.viewScanResults()
	case StageProviderSelect:
		return m.viewProviderSelect()
	case StageAPIKeyInput:
		return m.viewAPIKeyInput()
	case StageComplete:
		return m.viewComplete()
	}

	return ""
}

// viewWelcome renders the welcome screen
func (m OnboardingModel) viewWelcome() string {
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
	content.WriteString(subtitleStyle.Render("AI CLI Control Plane"))
	content.WriteString("\n\n")
	content.WriteString(promptStyle.Render("Let's get you set up in just a few steps"))
	content.WriteString("\n\n")
	content.WriteString(helpStyle.Render("Press Enter to continue • Ctrl+C to exit"))

	return content.String()
}

// viewScanning renders the scanning screen
func (m OnboardingModel) viewScanning() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.Primary).
		Padding(1, 0)

	var content strings.Builder
	content.WriteString("\n\n")
	content.WriteString(titleStyle.Render("Scanning your environment..."))
	content.WriteString("\n\n")
	content.WriteString(m.spinner.View() + " Checking for AI CLI tools...")
	content.WriteString("\n")

	return content.String()
}

// viewScanResults renders the scan results screen
func (m OnboardingModel) viewScanResults() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.Primary).
		Padding(1, 0)

	successStyle := lipgloss.NewStyle().Foreground(m.theme.Success)
	errorStyle := lipgloss.NewStyle().Foreground(m.theme.Warning)
	mutedStyle := lipgloss.NewStyle().Foreground(m.theme.Muted)
	helpStyle := lipgloss.NewStyle().
		Foreground(m.theme.Muted).
		Padding(1, 0)

	var content strings.Builder
	content.WriteString("\n")
	content.WriteString(titleStyle.Render("Tool Detection Results"))
	content.WriteString("\n\n")

	foundAny := false
	for _, det := range m.detections {
		if det.Found {
			content.WriteString(successStyle.Render("✓ " + det.Tool.Name))
			content.WriteString("\n")
			content.WriteString(mutedStyle.Render("  Location: " + det.Path))
			content.WriteString("\n")
			content.WriteString(mutedStyle.Render("  Status: Ready"))
			content.WriteString("\n\n")
			foundAny = true
		} else {
			content.WriteString(errorStyle.Render("✗ " + det.Tool.Name))
			content.WriteString("\n")
			content.WriteString(mutedStyle.Render("  Status: Not found in PATH"))
			content.WriteString("\n\n")
		}
	}

	if foundAny {
		content.WriteString(helpStyle.Render("Press Enter to continue • Esc to exit"))
	} else {
		content.WriteString(errorStyle.Render("No supported CLI tools found."))
		content.WriteString("\n\n")
		content.WriteString(mutedStyle.Render("Please install Claude Code CLI from:"))
		content.WriteString("\n")
		content.WriteString(mutedStyle.Render("https://claude.ai/download"))
		content.WriteString("\n\n")
		content.WriteString(helpStyle.Render("Press Ctrl+C to exit"))
	}

	return content.String()
}

// viewProviderSelect renders the provider selection screen
func (m OnboardingModel) viewProviderSelect() string {
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

// viewAPIKeyInput renders the API key input screen
func (m OnboardingModel) viewAPIKeyInput() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.Primary).
		Padding(1, 0)

	labelStyle := lipgloss.NewStyle().
		Foreground(m.theme.Primary).
		Bold(true)

	helpTextStyle := lipgloss.NewStyle().
		Foreground(m.theme.Muted).
		Italic(true).
		MarginLeft(2)

	inputBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.Primary).
		Padding(0, 1).
		Width(54)

	var content strings.Builder
	content.WriteString("\n")
	content.WriteString(titleStyle.Render(fmt.Sprintf("Configure %s", m.selectedProvider.DisplayName)))
	content.WriteString("\n\n")

	if m.apiKeyFromEnv {
		// API key found in environment
		content.WriteString(labelStyle.Render("API Key Detected"))
		content.WriteString("\n\n")
		content.WriteString(fmt.Sprintf("Found %s in environment\n", m.selectedProvider.APIKey.EnvVar))
		maskedKey := m.apiKeyValue
		if len(maskedKey) > 10 {
			maskedKey = maskedKey[:10] + "••••••••••••••••"
		}
		content.WriteString(fmt.Sprintf("Key: %s\n", maskedKey))
		content.WriteString("\n")
		content.WriteString(helpTextStyle.Render("[Y] Use this key  [N] Enter different key"))
	} else {
		// Need to input API key
		content.WriteString(labelStyle.Render("API Key"))
		content.WriteString("\n")
		content.WriteString(inputBoxStyle.Render(m.apiKeyInput.View()))
		content.WriteString("\n")
		keyURL := getProviderKeyURL(m.selectedProvider)
		content.WriteString(helpTextStyle.Render(fmt.Sprintf("Get your key at: %s", keyURL)))
		content.WriteString("\n\n")
		content.WriteString(helpTextStyle.Render("Enter: Continue • Esc: Back"))
	}

	return content.String()
}

// viewComplete renders the completion screen
func (m OnboardingModel) viewComplete() string {
	if m.err != nil {
		errorStyle := lipgloss.NewStyle().
			Foreground(m.theme.Warning).
			Bold(true).
			Padding(1, 2)

		return errorStyle.Render(fmt.Sprintf("❌ Error: %v\n\nPress Ctrl+C to exit", m.err))
	}

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

	var toolNames []string
	for _, det := range m.detections {
		if det.Found {
			toolNames = append(toolNames, det.Tool.Name)
		}
	}

	var content strings.Builder
	content.WriteString("\n\n\n")
	content.WriteString(successStyle.Render("✓ Setup Complete!"))
	content.WriteString("\n\n")
	content.WriteString(infoStyle.Render(fmt.Sprintf(
		"%s is now configured with:\n• Provider: %s\n• Model: %s\n\nConfiguration saved to %s",
		strings.Join(toolNames, ", "),
		m.selectedProvider.DisplayName,
		m.selectedProvider.DefaultModel,
		m.home,
	)))
	content.WriteString("\n\n")
	content.WriteString(helpStyle.Render("Press Enter to launch Dashboard"))

	return content.String()
}

// getProviderKeyURL returns the URL where users can get API keys
func getProviderKeyURL(provider *core.Provider) string {
	switch provider.Kind {
	case core.ProviderKindAnthropic:
		return "console.anthropic.com"
	case core.ProviderKindAnthropicCompatible:
		if strings.Contains(provider.BaseURL, "z.ai") {
			return "z.ai"
		}
		return "provider's website"
	case core.ProviderKindOpenAI:
		return "platform.openai.com/api-keys"
	case core.ProviderKindGemini:
		return "makersuite.google.com/app/apikey"
	default:
		return "provider's website"
	}
}

// RunOnboarding starts the onboarding wizard and returns whether to continue to dashboard
func RunOnboarding(home string) (bool, error) {
	wizard, err := NewOnboarding(home)
	if err != nil {
		return false, fmt.Errorf("failed to create onboarding wizard: %w", err)
	}

	p := tea.NewProgram(wizard, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return false, fmt.Errorf("onboarding error: %w", err)
	}

	// Check if wizard completed successfully
	if m, ok := finalModel.(OnboardingModel); ok {
		if m.err != nil {
			return false, m.err
		}
		return m.stage == StageComplete, nil
	}

	return false, nil
}
