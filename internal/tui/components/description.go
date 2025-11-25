package components

import (
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

type DescriptionProps struct {
	Description string
	Width       int
}

func RenderDescription(props DescriptionProps) string {
	if props.Description != "" {
		renderer, err := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(props.Width),
		)
		if err == nil {
			renderedDesc, err := renderer.Render(props.Description)
			if err == nil {
				return strings.TrimSpace(renderedDesc)
			}
		}
		return props.Description
	}

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Italic(true).
		Render("No description")
}
