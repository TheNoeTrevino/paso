package config

import "github.com/thenoetrevino/paso/internal/config/colors"

// DefaultColorScheme returns the default color scheme (purple theme)
func DefaultColorScheme() colors.ColorScheme {
	return *colors.Default()
}

// MonochromeColorScheme returns a black and white color scheme
func MonochromeColorScheme() colors.ColorScheme {
	return *colors.Monochrome()
}
