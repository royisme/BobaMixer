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
// This theme is designed to be modern, vibrant, and high-contrast.
func DefaultTheme() Theme {
	return Theme{
		Primary: lipgloss.AdaptiveColor{
			Light: "#4F46E5", // Indigo 600
			Dark:  "#818CF8", // Indigo 400 (Much brighter)
		},
		Success: lipgloss.AdaptiveColor{
			Light: "#059669",
			Dark:  "#34D399", // Emerald 400
		},
		Warning: lipgloss.AdaptiveColor{
			Light: "#D97706",
			Dark:  "#FBBF24", // Amber 400
		},
		Danger: lipgloss.AdaptiveColor{
			Light: "#DC2626",
			Dark:  "#F87171", // Red 400
		},
		Muted: lipgloss.AdaptiveColor{
			Light: "#6B7280",
			Dark:  "#9CA3AF", // Gray 400
		},
		Text: lipgloss.AdaptiveColor{
			Light: "#111827",
			Dark:  "#F3F4F6", // Gray 100 (Brighter)
		},
		Border: lipgloss.AdaptiveColor{
			Light: "#E5E7EB",
			Dark:  "#4B5563", // Gray 600 (Darker contrast)
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
