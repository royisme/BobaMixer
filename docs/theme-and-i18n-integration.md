# BobaMixer Theme & i18n Integration Guide

This document explains how to use the new theme system and internationalization features in BobaMixer TUI.

## 🎨 Theme System

### Overview

BobaMixer now supports **adaptive themes** that automatically adjust to your terminal's background color (light or dark). This follows Bubble Tea best practices using `lipgloss.AdaptiveColor`.

### Available Themes

1. **default** - Clean, modern adaptive theme (recommended)
2. **catppuccin** - Soothing pastel theme (Latte for light, Mocha for dark)
3. **dracula** - Popular dark theme with good contrast

### Configuration

Add to your `~/.boba/settings.yaml`:

```yaml
mode: observer
theme: default  # or catppuccin, dracula
explore:
  enabled: true
  rate: 0.03
```

### Code Example

```go
package main

import (
    "github.com/royisme/bobamixer/internal/ui"
    "github.com/charmbracelet/lipgloss"
)

func main() {
    // Get theme from settings
    theme := ui.GetTheme("default")

    // Use adaptive colors in your styles
    titleStyle := lipgloss.NewStyle().
        Bold(true).
        Foreground(theme.Primary).  // Automatically adapts!
        MarginBottom(1)

    successStyle := lipgloss.NewStyle().
        Foreground(theme.Success).
        Bold(true)

    // Render will use appropriate colors based on terminal background
    fmt.Println(titleStyle.Render("Welcome!"))
    fmt.Println(successStyle.Render("✓ Setup complete"))
}
```

### How It Works

Lipgloss automatically:
1. Detects your terminal's background color (light/dark)
2. Selects the appropriate color from the theme
3. Degrades colors gracefully (TrueColor → ANSI256 → ANSI)

**Example:**
```go
Primary: lipgloss.AdaptiveColor{
    Light: "#5A56E0",  // Used on light terminals (white background)
    Dark:  "#7C3AED",  // Used on dark terminals (black background)
}
```

---

## 🌍 Internationalization (i18n)

### Overview

BobaMixer supports multiple languages using `go-i18n`. The user's language is automatically detected from the `LANG` environment variable.

### Supported Languages

- **English** (en) - Default
- **简体中文** (zh-CN) - Simplified Chinese
- More languages can be added easily!

### Auto-Detection

The system automatically detects language from:
1. `LANG` environment variable (e.g., `zh_CN.UTF-8`)
2. `LC_ALL` or `LC_MESSAGES`
3. Falls back to English if not detected

### Code Example

```go
package main

import (
    "github.com/royisme/bobamixer/internal/ui"
)

func main() {
    // Auto-detect user's language
    lang := ui.GetUserLanguage()  // Returns "zh-CN" or "en"

    // Create localizer
    localizer, _ := ui.NewLocalizer(lang)

    // Simple translation
    title := localizer.T("welcome.title")
    // English: "🧋 Welcome to BobaMixer!"
    // Chinese: "🧋 欢迎使用 BobaMixer！"

    // Translation with data
    message := localizer.TP("welcome.step1_location", map[string]interface{}{
        "Path": "/home/user/.boba/profiles.yaml",
    })
    // English: "Location: /home/user/.boba/profiles.yaml"
    // Chinese: "位置：/home/user/.boba/profiles.yaml"

    fmt.Println(title)
    fmt.Println(message)
}
```

### Adding New Translations

1. **Add translation file**: `locales/ja.json` (for Japanese)

```json
[
  {
    "id": "welcome.title",
    "translation": "🧋 BobaMixerへようこそ！"
  },
  {
    "id": "welcome.step1_location",
    "translation": "場所: {{.Path}}"
  }
]
```

2. **Update i18n.go**: Add file to `localeFiles` array

3. **Done!** The language will be auto-detected

---

## 📋 Integration Example: Welcome Screen

### Before (Hardcoded):

```go
func runWelcomeScreen(home string, configErr error) error {
    fmt.Println(titleStyle.Render("🧋 Welcome to BobaMixer!"))
    fmt.Println(colorize(warningColor, "⚠ Configuration Required"))
    fmt.Println("Step 1: Review profiles.yaml")
    fmt.Printf("  Location: %s\n", filepath.Join(home, "profiles.yaml"))
    return nil
}
```

### After (Theme + i18n):

```go
func runWelcomeScreen(home string, configErr error) error {
    // Setup theme and i18n
    theme := ui.GetTheme(getThemeFromSettings())
    localizer, _ := ui.NewLocalizer(ui.GetUserLanguage())

    // Create adaptive styles
    titleStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.Primary)
    warningStyle := lipgloss.NewStyle().Foreground(theme.Warning)

    // Render with translations
    fmt.Println(titleStyle.Render(localizer.T("welcome.title")))
    fmt.Println(warningStyle.Render(localizer.T("welcome.config_required")))
    fmt.Println(localizer.T("welcome.step1_title"))
    fmt.Println(localizer.TP("welcome.step1_location", map[string]interface{}{
        "Path": filepath.Join(home, "profiles.yaml"),
    }))

    return nil
}
```

**Benefits:**
- ✅ Works on both light and dark terminals
- ✅ Automatically translates based on user's language
- ✅ Easy to add new themes and languages
- ✅ Follows Bubble Tea best practices

---

## 🚀 Migration Checklist

To integrate these features into existing TUI code:

- [ ] Replace `lipgloss.Color("#xxx")` with adaptive colors
- [ ] Create theme instance: `theme := ui.GetTheme("default")`
- [ ] Use theme colors: `style.Foreground(theme.Primary)`
- [ ] Create localizer: `localizer, _ := ui.NewLocalizer(ui.GetUserLanguage())`
- [ ] Replace hardcoded strings with `localizer.T("message.id")`
- [ ] Add translations to `locales/en.json` and `locales/zh-CN.json`
- [ ] Test in both light and dark terminals
- [ ] Test with `LANG=zh_CN.UTF-8` for Chinese

---

## 📚 Resources

- [Lipgloss Documentation](https://github.com/charmbracelet/lipgloss)
- [Bubble Tea Best Practices](https://github.com/charmbracelet/bubbletea)
- [go-i18n Guide](https://github.com/nicksnyder/go-i18n)
- [BobaMixer Themes](https://github.com/royisme/bobamixer/internal/ui/theme.go)

---

## 🤝 Contributing

Want to add a new language or theme?

1. **New Language**: Add `locales/{lang-code}.json` with translations
2. **New Theme**: Add theme function to `internal/ui/theme.go`
3. **Submit PR**: Share with the community!

Happy theming and translating! 🎨🌍
