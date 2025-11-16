# BobaMixer Phase 1 - UI/UX Design (Claude Only)

Version: 1.0
Date: 2025-01-16
Scope: Phase 1 - Core Control Plane (Claude CLI only)

---

## Design Philosophy

Following the principles in `CLAUDE.md`:
1. **Research First**: Study Bubble Tea table examples
2. **Design Before Code**: Complete user flows before implementation
3. **Modern TUI Standards**: Interactive, visual feedback, smart defaults
4. **Seamless Experience**: No manual file editing required

---

## 1. Onboarding Flow (First Run)

### 1.1 User Journey

```
User runs `boba` (first time)
    ↓
┌─────────────────────────────────────────┐
│ 🧋 Welcome to BobaMixer!                │
│                                         │
│ Scanning your environment...           │
│ ⠋ Checking for AI CLI tools...         │
└─────────────────────────────────────────┘
    ↓
System scans for:
  - `claude` in PATH
  - ~/.claude/settings.json
  - ANTHROPIC_API_KEY in env
    ↓
┌─────────────────────────────────────────┐
│ Tool Detection Results                  │
│                                         │
│ ✓ Claude Code CLI                       │
│   Location: /usr/local/bin/claude       │
│   Config: ~/.claude/settings.json       │
│   Status: Ready                         │
│                                         │
│ Press Enter to continue...              │
└─────────────────────────────────────────┘
    ↓
┌─────────────────────────────────────────┐
│ Configure Provider for Claude           │
│                                         │
│ Choose your Anthropic provider:         │
│                                         │
│ > ● Anthropic (Official)                │
│   ○ Claude via Z.AI (GLM-4.6)          │
│   ○ Custom...                           │
│                                         │
│ ↑/↓: Navigate  Enter: Select            │
└─────────────────────────────────────────┘
    ↓
If API key not detected:
┌─────────────────────────────────────────┐
│ Anthropic API Key                       │
│                                         │
│ > •••••••••••••••••••••_                │
│                                         │
│ Get your key at: console.anthropic.com  │
│                                         │
│ Enter: Continue  Esc: Back              │
└─────────────────────────────────────────┘
    ↓
If API key already in env:
┌─────────────────────────────────────────┐
│ API Key Detected                        │
│                                         │
│ Found ANTHROPIC_API_KEY in environment  │
│ Key: sk-ant-••••••••••••••••            │
│                                         │
│ [Y] Use this key  [N] Enter different   │
└─────────────────────────────────────────┘
    ↓
┌─────────────────────────────────────────┐
│ ✓ Setup Complete!                       │
│                                         │
│ Claude is now configured with:          │
│ • Provider: Anthropic (Official)        │
│ • Model: claude-3-5-sonnet-20241022     │
│                                         │
│ Launching Control Plane...              │
└─────────────────────────────────────────┘
    ↓
Dashboard
```

### 1.2 State Machine

```go
type OnboardingStage int

const (
    StageWelcome OnboardingStage = iota
    StageScanning
    StageScanResults
    StageProviderSelect
    StageAPIKey
    StageComplete
)
```

### 1.3 Key Features

- **Auto-detection**: Scan PATH for `claude` binary
- **Smart defaults**: Use existing env vars if available
- **Visual feedback**: Spinner during scanning, checkmarks for success
- **Skip unnecessary steps**: If key exists in env, just confirm
- **Error handling**: Clear messages if `claude` not found

---

## 2. Dashboard (Main Interface)

### 2.1 Layout

```
┌─────────────────────────────────────────────────────────────────────┐
│ BobaMixer - AI CLI Control Plane                                    │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│ Tool      Provider                Model                   Status   │
│ ──────────────────────────────────────────────────────────────────  │
│ > claude  Anthropic (Official)    claude-3-5-sonnet      ✓ Ready   │
│                                                                     │
│                                                                     │
│                                                                     │
│                                                                     │
│                                                                     │
│                                                                     │
├─────────────────────────────────────────────────────────────────────┤
│ [R] Run  [B] Change Binding  [P] Providers  [?] Help  [Q] Quit     │
└─────────────────────────────────────────────────────────────────────┘
```

### 2.2 Interactions

#### Select Row + Press `R` (Run)

```
User presses R on claude row
    ↓
┌─────────────────────────────────────────┐
│ Run Claude Code CLI                     │
│                                         │
│ Command: boba run claude                │
│                                         │
│ [Enter] Run  [Esc] Cancel               │
└─────────────────────────────────────────┘
    ↓
If user presses Enter:
    Exit TUI, exec `boba run claude` in current shell
```

#### Select Row + Press `B` (Change Binding)

```
User presses B on claude row
    ↓
┌─────────────────────────────────────────┐
│ Change Provider for Claude              │
│                                         │
│ Current: Anthropic (Official)           │
│                                         │
│ Available Providers:                    │
│ ────────────────────────────────────    │
│ > ● Anthropic (Official)                │
│   ○ Claude via Z.AI                     │
│   ○ Add new provider...                 │
│                                         │
│ Enter: Select  Esc: Cancel              │
└─────────────────────────────────────────┘
    ↓
Update binding in memory + save to bindings.yaml
    ↓
Return to Dashboard (updated)
```

#### Press `P` (View Providers)

```
User presses P
    ↓
┌─────────────────────────────────────────────────────────────┐
│ Provider Management                                         │
│                                                             │
│ Provider              Type        Base URL            Key   │
│ ─────────────────────────────────────────────────────────── │
│ > Anthropic (Official) anthropic  api.anthropic.com   ✓ env │
│   Claude via Z.AI     anthropic   api.z.ai/...        ✗     │
│                                                             │
│ [Enter] Edit  [N] New  [D] Delete  [Esc] Back              │
└─────────────────────────────────────────────────────────────┘
```

### 2.3 State & Data Model

```go
type DashboardModel struct {
    // Data
    tools      []Tool
    providers  []Provider
    bindings   map[string]Binding  // tool_id -> binding

    // UI State
    table      table.Model
    viewMode   ViewMode  // ViewDashboard | ViewProviders | ViewBindingEdit

    // Dimensions
    width, height int

    // Theme
    theme      Theme

    // Error
    err        error
}

type ViewMode int
const (
    ViewDashboard ViewMode = iota
    ViewProviders
    ViewBindingEdit
)
```

---

## 3. Provider Management View

### 3.1 Provider List

```
┌─────────────────────────────────────────────────────────────────────┐
│ Provider Management                                                 │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│ Provider              Type          Base URL               Key      │
│ ────────────────────────────────────────────────────────────────    │
│ > Anthropic (Official) anthropic    api.anthropic.com      ✓ env    │
│   Claude via Z.AI     anthropic-c.  api.z.ai/api/anth...  ✗        │
│                                                                     │
│ ✓ = Configured   ✗ = Missing   env = From environment              │
│                                                                     │
├─────────────────────────────────────────────────────────────────────┤
│ [Enter] Edit  [N] New Provider  [D] Delete  [Esc] Back             │
└─────────────────────────────────────────────────────────────────────┘
```

### 3.2 Add New Provider Flow

```
User presses N
    ↓
┌─────────────────────────────────────────┐
│ Add New Provider                        │
│                                         │
│ Provider Type:                          │
│ > ● Anthropic (Official)                │
│   ○ Anthropic-compatible               │
│   ○ Custom                              │
│                                         │
│ Enter: Select  Esc: Cancel              │
└─────────────────────────────────────────┘
    ↓
┌─────────────────────────────────────────┐
│ Provider Details                        │
│                                         │
│ Display Name:                           │
│ > Custom Anthropic___                   │
│                                         │
│ Base URL:                               │
│ > https://api.example.com___            │
│                                         │
│ API Key (env var name):                 │
│ > CUSTOM_API_KEY___                     │
│                                         │
│ Tab: Next  Enter: Save  Esc: Cancel     │
└─────────────────────────────────────────┘
    ↓
Save to providers.yaml
    ↓
Return to Provider List
```

---

## 4. Components & Styling

### 4.1 Components Used

From `charmbracelet/bubbles`:
- **table**: Main dashboard table
- **list**: Provider selection lists
- **textinput**: API key input, provider details
- **spinner**: Loading states during scanning

### 4.2 Theme Integration

Use existing `Theme` from `internal/ui/theme.go`:

```go
// Table styles
selectedRowStyle := lipgloss.NewStyle().
    Foreground(theme.Text).
    Background(theme.Primary).
    Bold(true)

headerStyle := lipgloss.NewStyle().
    Foreground(theme.Primary).
    BorderStyle(lipgloss.NormalBorder()).
    BorderBottom(true).
    BorderForeground(theme.Border)

// Status indicators
readyStyle := lipgloss.NewStyle().Foreground(theme.Success)    // ✓ Ready
errorStyle := lipgloss.NewStyle().Foreground(theme.Warning)    // ✗ Error
mutedStyle := lipgloss.NewStyle().Foreground(theme.Muted)      // env, hints
```

### 4.3 Responsive Layout

```go
func (m DashboardModel) View() string {
    // Calculate column widths based on terminal size
    toolWidth := 10
    providerWidth := max(20, m.width/4)
    modelWidth := max(25, m.width/3)
    statusWidth := m.width - toolWidth - providerWidth - modelWidth - 10

    // Update table columns dynamically
    m.table.SetColumns([]table.Column{
        {Title: "Tool", Width: toolWidth},
        {Title: "Provider", Width: providerWidth},
        {Title: "Model", Width: modelWidth},
        {Title: "Status", Width: statusWidth},
    })
}
```

---

## 5. Error Handling & Edge Cases

### 5.1 Claude Not Found

```
┌─────────────────────────────────────────┐
│ ⚠ Claude CLI Not Detected               │
│                                         │
│ BobaMixer couldn't find Claude Code CLI │
│ installed on your system.               │
│                                         │
│ Install Claude Code:                    │
│ https://claude.ai/download              │
│                                         │
│ After installing, run `boba` again.     │
│                                         │
│ [Q] Quit  [R] Retry Scan                │
└─────────────────────────────────────────┘
```

### 5.2 API Key Invalid

```
┌─────────────────────────────────────────┐
│ ❌ API Key Validation Failed            │
│                                         │
│ The provided API key appears invalid.   │
│                                         │
│ Error: Authentication failed (401)      │
│                                         │
│ [E] Edit Key  [H] Help  [Esc] Cancel    │
└─────────────────────────────────────────┘
```

### 5.3 Binding Conflict

When user tries to bind a tool already bound:

```
┌─────────────────────────────────────────┐
│ Update Binding                          │
│                                         │
│ Claude is currently bound to:           │
│ • Anthropic (Official)                  │
│                                         │
│ Change to:                              │
│ • Claude via Z.AI                       │
│                                         │
│ [Y] Confirm  [N] Cancel                 │
└─────────────────────────────────────────┘
```

---

## 6. Navigation & Keybindings

### Global Keys
- `q`: Quit application
- `?`: Show help overlay
- `ctrl+c`: Force quit

### Dashboard Keys
- `↑/↓`: Navigate table rows
- `r`: Run selected tool
- `b`: Change binding for selected tool
- `p`: View providers
- `tab`: Cycle between sections (future: when we have stats/logs)

### Provider View Keys
- `↑/↓`: Navigate provider list
- `enter`: Edit selected provider
- `n`: New provider
- `d`: Delete provider
- `esc`: Back to dashboard

### Form Keys (API key input, provider details)
- `tab`: Next field
- `shift+tab`: Previous field
- `enter`: Submit / Continue
- `esc`: Cancel

---

## 7. Implementation Priority

### Phase 1.1: Onboarding (Week 1)
1. Welcome screen with spinner
2. Tool scanning logic (`exec.LookPath("claude")`)
3. Provider selection (list component)
4. API key input (if needed)
5. Complete screen

### Phase 1.2: Dashboard (Week 2)
1. Table layout with tool bindings
2. Row selection and highlighting
3. Run action (R key)
4. Binding edit action (B key)
5. Status indicators

### Phase 1.3: Provider Management (Week 3)
1. Provider list view
2. Add new provider flow
3. Edit provider
4. Delete provider (with confirmation)

---

## 8. Testing Scenarios

### Happy Path
1. First run → Claude detected → Key in env → Auto-configured → Dashboard
2. Dashboard → Press R → Launches `boba run claude`
3. Dashboard → Press B → Change provider → Binding updated

### Error Paths
1. First run → Claude not found → Error message → Retry/Quit
2. First run → No API key → Manual entry → Validation
3. Dashboard → Invalid binding → Clear error → Recovery

### Edge Cases
1. Terminal too small → Graceful degradation (min width warning)
2. Config file corrupted → Clear error → Offer to reinitialize
3. Multiple providers with same name → Append ID to display

---

## 9. Localization Support

Current wizard uses i18n. Maintain consistency:

```json
{
  "onboarding.welcome": "🧋 Welcome to BobaMixer!",
  "onboarding.scanning": "Scanning your environment...",
  "onboarding.claude_found": "✓ Claude Code CLI",
  "onboarding.claude_not_found": "⚠ Claude CLI Not Detected",

  "dashboard.title": "BobaMixer - AI CLI Control Plane",
  "dashboard.run": "Run",
  "dashboard.bind": "Change Binding",
  "dashboard.providers": "Providers",

  "provider.add": "Add New Provider",
  "provider.edit": "Edit Provider",
  "provider.delete": "Delete Provider"
}
```

---

## 10. Success Metrics

### User Experience
- ✅ Zero manual file editing required
- ✅ First run to working state in < 2 minutes
- ✅ Clear visual feedback at every step
- ✅ Graceful error handling with recovery paths

### Technical
- ✅ All state changes persisted to YAML
- ✅ Theme-consistent styling across all views
- ✅ Responsive layout (80-200 cols width)
- ✅ No crashes on invalid input

---

This design document serves as the blueprint for Phase 1 implementation. All UI components will be built following the patterns established in `CLAUDE.md`, using official Bubble Tea components, and maintaining visual consistency with the existing theme system.
