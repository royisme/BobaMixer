package ui

import "github.com/royisme/bobamixer/internal/ui/theme"

// Theme re-exports the theme palette type for existing callers while the refactor is in progress.
type Theme = theme.Theme

// Styles re-exports the default style collection derived from a palette.
type Styles = theme.Styles

// DefaultTheme wraps theme.DefaultTheme for backward compatibility.
func DefaultTheme() Theme {
	return theme.DefaultTheme()
}

// CatppuccinTheme wraps theme.CatppuccinTheme for backward compatibility.
func CatppuccinTheme() Theme {
	return theme.CatppuccinTheme()
}

// DraculaTheme wraps theme.DraculaTheme for backward compatibility.
func DraculaTheme() Theme {
	return theme.DraculaTheme()
}

// GetTheme wraps theme.GetTheme for backward compatibility.
func GetTheme(themeName string) Theme {
	return theme.GetTheme(themeName)
}

// NewStyles exposes the shared styles builder.
func NewStyles(palette Theme) Styles {
	return theme.NewStyles(palette)
}
