package huhforms

import (
	"charm.land/huh/v2"
)

// CreateTaskForm creates a huh form for adding/editing a task
// The form uses pointers to update values in place, matching the existing pattern
func CreateTaskForm(
	title *string,
	description *string,
	confirm *bool,
	descriptionLines int,
) *huh.Form {
	var fields []huh.Field

	// Title input field
	fields = append(fields,
		huh.NewInput().
			Key("title").
			Title("Title").
			Placeholder("Enter task title...").
			Value(title),
	)

	// Description text area field with dynamic height
	fields = append(fields,
		huh.NewText().
			Key("description").
			Title("Description").
			Placeholder("Enter task description...").
			CharLimit(5000).
			Lines(descriptionLines).
			Value(description),
	)

	// Confirmation
	fields = append(fields,
		huh.NewConfirm().
			Key("confirm").
			Title("Submit this task?").
			Affirmative("Yes").
			Negative("No").
			Value(confirm),
	)

	form := huh.NewForm(huh.NewGroup(fields...))
	return form.WithKeyMap(CreateKeyMapWithShiftEnter()).WithShowHelp(false)
}
