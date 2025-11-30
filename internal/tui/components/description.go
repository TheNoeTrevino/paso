package components

import (
	"strings"
	"sync"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

type DescriptionProps struct {
	Description string
	Width       int
}

// Cache Glamour renderers by width to avoid expensive re-creation
var (
	rendererCache sync.Map // map[int]*glamour.TermRenderer
)

// getRenderer returns a cached renderer for the given width
func getRenderer(width int) (*glamour.TermRenderer, error) {
	// Check cache first
	if cached, ok := rendererCache.Load(width); ok {
		return cached.(*glamour.TermRenderer), nil
	}

	// Create new renderer
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return nil, err
	}

	// Store in cache
	rendererCache.Store(width, renderer)
	return renderer, nil
}

func RenderDescription(props DescriptionProps) string {
	if props.Description != "" {
		renderer, err := getRenderer(props.Width)
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
