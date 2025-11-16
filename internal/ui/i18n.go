// Package ui provides internationalization support for BobaMixer TUI
package ui

import (
	"embed"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

//go:embed locales/*.json
var localesFS embed.FS

// Localizer wraps i18n.Localizer for easier usage
type Localizer struct {
	*i18n.Localizer
}

// NewLocalizer creates a new localizer for the given language
// Supported languages: en, zh-CN
// Falls back to English if the language is not supported
func NewLocalizer(lang string) (*Localizer, error) {
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	// Load embedded locale files
	localeFiles := []string{
		"locales/en.json",
		"locales/zh-CN.json",
	}

	for _, file := range localeFiles {
		data, err := localesFS.ReadFile(file)
		if err != nil {
			continue // Skip missing files
		}
		bundle.MustParseMessageFileBytes(data, filepath.Base(file))
	}

	// Try to load from external locales directory for overrides
	externalLocales := []string{"en.json", "zh-CN.json"}
	for _, file := range externalLocales {
		path := filepath.Join("locales", file)
		if data, err := os.ReadFile(path); err == nil {
			bundle.ParseMessageFileBytes(data, file)
		}
	}

	// Create localizer with fallback
	localizer := i18n.NewLocalizer(bundle, lang, language.English.String())

	return &Localizer{Localizer: localizer}, nil
}

// T is a shorthand for translating a message by ID
func (l *Localizer) T(messageID string) string {
	msg, err := l.Localize(&i18n.LocalizeConfig{
		MessageID: messageID,
	})
	if err != nil {
		return messageID // Return ID if translation fails
	}
	return msg
}

// TP translates a message with template data
func (l *Localizer) TP(messageID string, templateData map[string]interface{}) string {
	msg, err := l.Localize(&i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: templateData,
	})
	if err != nil {
		return messageID
	}
	return msg
}

// GetUserLanguage determines the user's preferred language from environment
func GetUserLanguage() string {
	// Check LANG environment variable
	if lang := os.Getenv("LANG"); lang != "" {
		// Extract language code (e.g., "zh_CN.UTF-8" -> "zh-CN")
		if len(lang) >= 5 {
			langCode := lang[:5]
			// Convert underscore to hyphen for BCP 47 format
			if langCode[2] == '_' {
				return langCode[:2] + "-" + langCode[3:5]
			}
		}
		// Return first 2 chars if format is different
		if len(lang) >= 2 {
			return lang[:2]
		}
	}

	// Check LC_ALL or LC_MESSAGES
	if lang := os.Getenv("LC_ALL"); lang != "" {
		return lang[:2]
	}
	if lang := os.Getenv("LC_MESSAGES"); lang != "" {
		return lang[:2]
	}

	// Default to English
	return "en"
}

// Example usage in welcome screen:
// func runWelcomeScreen(home string, configErr error) error {
//     localizer, _ := NewLocalizer(GetUserLanguage())
//
//     title := localizer.T("welcome.title")
//     configRequired := localizer.T("welcome.config_required")
//     step1 := localizer.T("welcome.step1_title")
//     step1Desc := localizer.TP("welcome.step1_desc", map[string]interface{}{
//         "Path": filepath.Join(home, "profiles.yaml"),
//     })
//
//     fmt.Println(title)
//     fmt.Println(configRequired)
//     // ...
// }
