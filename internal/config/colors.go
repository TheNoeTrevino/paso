package config

import "github.com/thenoetrevino/paso/internal/config/colors"

// ColorScheme is an alias for colors.ColorScheme for backwards compatibility
type ColorScheme = colors.ColorScheme

// DefaultColorScheme returns the default color scheme (purple theme)
func DefaultColorScheme() ColorScheme {
	return *colors.Default()
}

// MonochromeColorScheme returns a black and white color scheme
func MonochromeColorScheme() ColorScheme {
	return *colors.Monochrome()
}
