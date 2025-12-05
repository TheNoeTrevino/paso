package tui

import (
	tea "charm.land/bubbletea/v2"
	oldtea "github.com/charmbracelet/bubbletea"
)

// wrapV1Cmd wraps a bubbletea v1 Cmd to work with v2
func wrapV1Cmd(v1Cmd oldtea.Cmd) tea.Cmd {
	if v1Cmd == nil {
		return nil
	}
	return func() tea.Msg {
		return v1Cmd()
	}
}
