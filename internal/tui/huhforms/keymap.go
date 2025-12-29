package huhforms

import (
	"charm.land/bubbles/v2/key"
	"charm.land/huh/v2"
)

// CreateKeyMapWithShiftEnter creates a custom keymap that includes shift+enter
// for newlines in text fields, in addition to the default alt+enter and ctrl+j.
func CreateKeyMapWithShiftEnter() *huh.KeyMap {
	keymap := huh.NewDefaultKeyMap()

	// Add shift+enter to the existing newline keys (alt+enter, ctrl+j)
	keymap.Text.NewLine = key.NewBinding(
		key.WithKeys("shift+enter", "alt+enter", "ctrl+j"),
		key.WithHelp("shift+enter / alt+enter / ctrl+j", "new line"),
	)

	return keymap
}
