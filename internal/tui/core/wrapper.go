package core

import "github.com/thenoetrevino/paso/internal/tui"

// Wrapper wraps a TUI Model to provide core MVC methods (Init, Update, View)
// while maintaining access to all Model fields and methods.
type Wrapper struct {
	*tui.Model
}

// New creates a new Wrapper around the given Model
func New(m *tui.Model) *Wrapper {
	return &Wrapper{Model: m}
}
