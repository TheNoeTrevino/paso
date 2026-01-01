// Package huhforms contains form that use the huh library for the TUI app
package huhforms

import "charm.land/huh/v2"

// CreateCommentForm creates a huh form for adding or editing a comment.
// No confirmation field is used - the form saves on completion.
func CreateCommentForm(
	message *string,
	isEdit bool,
) *huh.Form {
	title := "New Comment"
	if isEdit {
		title = "Edit Comment"
	}

	fields := []huh.Field{
		huh.NewText().
			Key("message").
			Title(title).
			Placeholder("Enter comment text...").
			Value(message).
			CharLimit(1000), // Reasonable limit for comments
	}

	form := huh.NewForm(huh.NewGroup(fields...))
	return form.WithKeyMap(CreateKeyMapWithShiftEnter())
}
