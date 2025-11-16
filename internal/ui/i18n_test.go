package ui

import (
	"os"
	"testing"
)

func TestI18nEnglish(t *testing.T) {
	os.Setenv("LANG", "en_US.UTF-8")

	localizer, err := NewLocalizer(GetUserLanguage())
	if err != nil {
		t.Fatalf("Failed to create localizer: %v", err)
	}

	title := localizer.T("welcome.title")
	expected := "🧋 Welcome to BobaMixer!"

	if title != expected {
		t.Errorf("Expected %q, got %q", expected, title)
	}
}

func TestI18nChinese(t *testing.T) {
	os.Setenv("LANG", "zh_CN.UTF-8")

	localizer, err := NewLocalizer(GetUserLanguage())
	if err != nil {
		t.Fatalf("Failed to create localizer: %v", err)
	}

	title := localizer.T("welcome.title")
	expected := "🧋 欢迎使用 BobaMixer！"

	if title != expected {
		t.Errorf("Expected %q, got %q", expected, title)
	}
}

func TestGetUserLanguage(t *testing.T) {
	tests := []struct{
		name string
		lang string
		want string
	}{
		{"English US", "en_US.UTF-8", "en"},
		{"Chinese Simplified", "zh_CN.UTF-8", "zh-CN"},
		{"Japanese", "ja_JP.UTF-8", "ja"},
		{"Default fallback", "", "en"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.lang != "" {
				os.Setenv("LANG", tt.lang)
			} else {
				os.Unsetenv("LANG")
			}

			got := GetUserLanguage()
			if got != tt.want {
				t.Errorf("GetUserLanguage() = %q, want %q", got, tt.want)
			}
		})
	}
}
