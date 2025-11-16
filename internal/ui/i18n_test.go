package ui

import (
	"os"
	"testing"
)

func TestI18nEnglish(t *testing.T) {
	if err := os.Setenv("LANG", "en_US.UTF-8"); err != nil {
		t.Fatalf("Failed to set LANG: %v", err)
	}

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
	if err := os.Setenv("LANG", "zh_CN.UTF-8"); err != nil {
		t.Fatalf("Failed to set LANG: %v", err)
	}

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
		{"English US", "en_US.UTF-8", "en-US"},
		{"Chinese Simplified", "zh_CN.UTF-8", "zh-CN"},
		{"Japanese", "ja_JP.UTF-8", "ja-JP"},
		{"Default fallback", "", "en"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.lang != "" {
				if err := os.Setenv("LANG", tt.lang); err != nil {
					t.Fatalf("Failed to set LANG: %v", err)
				}
			} else {
				if err := os.Unsetenv("LANG"); err != nil {
					t.Fatalf("Failed to unset LANG: %v", err)
				}
			}

			got := GetUserLanguage()
			if got != tt.want {
				t.Errorf("GetUserLanguage() = %q, want %q", got, tt.want)
			}
		})
	}
}
