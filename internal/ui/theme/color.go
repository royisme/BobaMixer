// Package theme provides color schemes and styling for the BobaMixer TUI.
package theme

import "github.com/charmbracelet/lipgloss"

// Theme defines the color scheme for the TUI.
type Theme struct {
	Primary lipgloss.AdaptiveColor
	Success lipgloss.AdaptiveColor
	Warning lipgloss.AdaptiveColor
	Danger  lipgloss.AdaptiveColor
	Muted   lipgloss.AdaptiveColor
	Text    lipgloss.AdaptiveColor
	Border  lipgloss.AdaptiveColor
}

// DefaultTheme returns an adaptive palette that works on both light and dark terminals.
func DefaultTheme() Theme {
	return Theme{
		Primary: lipgloss.AdaptiveColor{
			Light: "#5A56E0",
			Dark:  "#7C3AED",
		},
		Success: lipgloss.AdaptiveColor{
			Light: "#059669",
			Dark:  "#10B981",
		},
		Warning: lipgloss.AdaptiveColor{
			Light: "#D97706",
			Dark:  "#F59E0B",
		},
		Danger: lipgloss.AdaptiveColor{
			Light: "#DC2626",
			Dark:  "#EF4444",
		},
		Muted: lipgloss.AdaptiveColor{
			Light: "#6B7280",
			Dark:  "#9CA3AF",
		},
		Text: lipgloss.AdaptiveColor{
			Light: "#1F2937",
			Dark:  "#E5E7EB",
		},
		Border: lipgloss.AdaptiveColor{
			Light: "#D1D5DB",
			Dark:  "#4B5563",
		},
	}
}

// CatppuccinTheme returns a Catppuccin-inspired palette.
func CatppuccinTheme() Theme {
	return Theme{
		Primary: lipgloss.AdaptiveColor{
			Light: "#8839EF",
			Dark:  "#CBA6F7",
		},
		Success: lipgloss.AdaptiveColor{
			Light: "#40A02B",
			Dark:  "#A6E3A1",
		},
		Warning: lipgloss.AdaptiveColor{
			Light: "#DF8E1D",
			Dark:  "#F9E2AF",
		},
		Danger: lipgloss.AdaptiveColor{
			Light: "#D20F39",
			Dark:  "#F38BA8",
		},
		Muted: lipgloss.AdaptiveColor{
			Light: "#6C6F85",
			Dark:  "#A6ADC8",
		},
		Text: lipgloss.AdaptiveColor{
			Light: "#4C4F69",
			Dark:  "#CDD6F4",
		},
		Border: lipgloss.AdaptiveColor{
			Light: "#DCE0E8",
			Dark:  "#45475A",
		},
	}
}

// DraculaTheme returns a Dracula-inspired palette.
func DraculaTheme() Theme {
	return Theme{
		Primary: lipgloss.AdaptiveColor{
			Light: "#6272A4",
			Dark:  "#BD93F9",
		},
		Success: lipgloss.AdaptiveColor{
			Light: "#50FA7B",
			Dark:  "#50FA7B",
		},
		Warning: lipgloss.AdaptiveColor{
			Light: "#F1FA8C",
			Dark:  "#F1FA8C",
		},
		Danger: lipgloss.AdaptiveColor{
			Light: "#FF5555",
			Dark:  "#FF5555",
		},
		Muted: lipgloss.AdaptiveColor{
			Light: "#6272A4",
			Dark:  "#6272A4",
		},
		Text: lipgloss.AdaptiveColor{
			Light: "#44475A",
			Dark:  "#F8F8F2",
		},
		Border: lipgloss.AdaptiveColor{
			Light: "#6272A4",
			Dark:  "#44475A",
		},
	}
}

// GetTheme returns the palette that matches the provided theme name.
func GetTheme(themeName string) Theme {
	switch themeName {
	case "catppuccin":
		return CatppuccinTheme()
	case "dracula":
		return DraculaTheme()
	default:
		return DefaultTheme()
	}
}
