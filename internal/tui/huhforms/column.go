package huhforms

import "charm.land/huh/v2"

// CreateColumnForm creates a huh form for adding or editing a column.
// The form contains a single input field for the column name.
// No confirmation field is used - the form saves on completion.
func CreateColumnForm(
	name *string,
	isEdit bool,
) *huh.Form {
	title := "New Column Name"
	if isEdit {
		title = "Rename Column"
	}

	fields := []huh.Field{
		huh.NewInput().
			Key("name").
			Title(title).
			Placeholder("Enter column name...").
			Value(name),
	}

	return huh.NewForm(huh.NewGroup(fields...))
}
