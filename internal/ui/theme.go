// Package ui provides terminal user interface theming
package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// Theme defines the color scheme for the TUI
type Theme struct {
	Primary lipgloss.AdaptiveColor
	Success lipgloss.AdaptiveColor
	Warning lipgloss.AdaptiveColor
	Danger  lipgloss.AdaptiveColor
	Muted   lipgloss.AdaptiveColor
	Text    lipgloss.AdaptiveColor
	Border  lipgloss.AdaptiveColor
}

// DefaultTheme returns an adaptive theme that works on both light and dark terminals
func DefaultTheme() Theme {
	return Theme{
		Primary: lipgloss.AdaptiveColor{
			Light: "#5A56E0", // Darker purple for light backgrounds
			Dark:  "#7C3AED", // Brighter purple for dark backgrounds
		},
		Success: lipgloss.AdaptiveColor{
			Light: "#059669", // Darker green
			Dark:  "#10B981", // Brighter green
		},
		Warning: lipgloss.AdaptiveColor{
			Light: "#D97706", // Darker amber
			Dark:  "#F59E0B", // Brighter amber
		},
		Danger: lipgloss.AdaptiveColor{
			Light: "#DC2626", // Darker red
			Dark:  "#EF4444", // Brighter red
		},
		Muted: lipgloss.AdaptiveColor{
			Light: "#6B7280", // Medium gray (same on both)
			Dark:  "#9CA3AF", // Lighter gray for dark backgrounds
		},
		Text: lipgloss.AdaptiveColor{
			Light: "#1F2937", // Dark gray text on light background
			Dark:  "#E5E7EB", // Light gray text on dark background
		},
		Border: lipgloss.AdaptiveColor{
			Light: "#D1D5DB", // Light gray border
			Dark:  "#4B5563", // Dark gray border
		},
	}
}

// CatppuccinTheme returns a Catppuccin-inspired theme
// Catppuccin Latte for light mode, Catppuccin Mocha for dark mode
func CatppuccinTheme() Theme {
	return Theme{
		Primary: lipgloss.AdaptiveColor{
			Light: "#8839EF", // Latte Mauve
			Dark:  "#CBA6F7", // Mocha Mauve
		},
		Success: lipgloss.AdaptiveColor{
			Light: "#40A02B", // Latte Green
			Dark:  "#A6E3A1", // Mocha Green
		},
		Warning: lipgloss.AdaptiveColor{
			Light: "#DF8E1D", // Latte Yellow
			Dark:  "#F9E2AF", // Mocha Yellow
		},
		Danger: lipgloss.AdaptiveColor{
			Light: "#D20F39", // Latte Red
			Dark:  "#F38BA8", // Mocha Red
		},
		Muted: lipgloss.AdaptiveColor{
			Light: "#6C6F85", // Latte Subtext1
			Dark:  "#A6ADC8", // Mocha Subtext1
		},
		Text: lipgloss.AdaptiveColor{
			Light: "#4C4F69", // Latte Text
			Dark:  "#CDD6F4", // Mocha Text
		},
		Border: lipgloss.AdaptiveColor{
			Light: "#DCE0E8", // Latte Surface0
			Dark:  "#45475A", // Mocha Surface0
		},
	}
}

// DraculaTheme returns a Dracula-inspired theme
func DraculaTheme() Theme {
	return Theme{
		Primary: lipgloss.AdaptiveColor{
			Light: "#6272A4", // Dracula Comment (darker for light mode)
			Dark:  "#BD93F9", // Dracula Purple
		},
		Success: lipgloss.AdaptiveColor{
			Light: "#50FA7B", // Dracula Green (works on both)
			Dark:  "#50FA7B",
		},
		Warning: lipgloss.AdaptiveColor{
			Light: "#F1FA8C", // Dracula Yellow (works on both)
			Dark:  "#F1FA8C",
		},
		Danger: lipgloss.AdaptiveColor{
			Light: "#FF5555", // Dracula Red (works on both)
			Dark:  "#FF5555",
		},
		Muted: lipgloss.AdaptiveColor{
			Light: "#6272A4", // Dracula Comment
			Dark:  "#6272A4",
		},
		Text: lipgloss.AdaptiveColor{
			Light: "#44475A", // Dracula Current Line (darker)
			Dark:  "#F8F8F2", // Dracula Foreground
		},
		Border: lipgloss.AdaptiveColor{
			Light: "#6272A4", // Dracula Comment
			Dark:  "#44475A", // Dracula Current Line
		},
	}
}

// GetTheme returns the appropriate theme based on user settings
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
