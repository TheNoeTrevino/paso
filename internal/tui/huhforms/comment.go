package huhforms

import "charm.land/huh/v2"

// CreateCommentForm creates a huh form for adding or editing a comment/note.
// The form contains a multi-line text field for the comment message.
// No confirmation field is used - the form saves on completion.
func CreateCommentForm(
	message *string,
	isEdit bool,
) *huh.Form {
	title := "New Note"
	if isEdit {
		title = "Edit Note"
	}

	fields := []huh.Field{
		huh.NewText().
			Key("message").
			Title(title).
			Placeholder("Enter note text...").
			Value(message).
			CharLimit(1000), // Reasonable limit for notes
	}

	form := huh.NewForm(huh.NewGroup(fields...))
	return form.WithKeyMap(CreateKeyMapWithShiftEnter())
}
