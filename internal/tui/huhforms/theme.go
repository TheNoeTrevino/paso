package huhforms

import (
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"github.com/thenoetrevino/paso/internal/config/colors"
)

// CreatePasoTheme creates a custom huh theme matching paso's color scheme
func CreatePasoTheme(colorScheme colors.ColorScheme) huh.Theme {
	return huh.ThemeFunc(func(isDark bool) *huh.Styles {
		t := huh.ThemeBase(isDark)

		accent := lipgloss.Color(colorScheme.Accent)
		create := lipgloss.Color(colorScheme.Create)
		subtle := lipgloss.Color(colorScheme.Subtle)
		normal := lipgloss.Color(colorScheme.Normal)
		errorColor := lipgloss.Color(colorScheme.Delete)
		title := lipgloss.Color(colorScheme.Title)

		// Focused field styles
		t.Focused.Base = t.Focused.Base.BorderForeground(accent)
		t.Focused.Title = t.Focused.Title.Foreground(title).Bold(true)
		t.Focused.Description = t.Focused.Description.Foreground(subtle)
		t.Focused.ErrorIndicator = t.Focused.ErrorIndicator.Foreground(errorColor)
		t.Focused.ErrorMessage = t.Focused.ErrorMessage.Foreground(errorColor)
		t.Focused.SelectSelector = t.Focused.SelectSelector.Foreground(accent)
		t.Focused.MultiSelectSelector = t.Focused.MultiSelectSelector.Foreground(accent)
		t.Focused.SelectedOption = t.Focused.SelectedOption.Foreground(create)
		t.Focused.SelectedPrefix = t.Focused.SelectedPrefix.Foreground(create)
		t.Focused.UnselectedOption = t.Focused.UnselectedOption.Foreground(normal)
		t.Focused.UnselectedPrefix = t.Focused.UnselectedPrefix.Foreground(subtle)
		t.Focused.FocusedButton = t.Focused.FocusedButton.
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(accent).
			Bold(true)
		t.Focused.BlurredButton = t.Focused.BlurredButton.
			Foreground(normal).
			Background(subtle)

		// TextInput styles
		t.Focused.TextInput.Cursor = t.Focused.TextInput.Cursor.Foreground(accent)
		t.Focused.TextInput.Placeholder = t.Focused.TextInput.Placeholder.Foreground(subtle)
		t.Focused.TextInput.Prompt = t.Focused.TextInput.Prompt.Foreground(accent)

		// Blurred field styles (inherit from focused but with hidden border)
		t.Blurred = t.Focused
		t.Blurred.Base = t.Blurred.Base.BorderStyle(lipgloss.HiddenBorder())
		t.Blurred.Title = t.Blurred.Title.Foreground(subtle)

		return t
	})
}
