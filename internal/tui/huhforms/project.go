package huhforms

import "charm.land/huh/v2"

// CreateProjectForm creates a huh form for adding a new project
func CreateProjectForm(
	name *string,
	description *string,
	confirm *bool,
) *huh.Form {
	fields := []huh.Field{
		huh.NewInput().
			Key("name").
			Title("Project Name").
			Placeholder("Enter project name...").
			Value(name),

		huh.NewText().
			Key("description").
			Title("Description (optional)").
			Placeholder("Enter project description...").
			CharLimit(500).
			Lines(3).
			Value(description),

		huh.NewConfirm().
			Key("confirm").
			Title("Create this project?").
			Affirmative("Yes").
			Negative("No").
			Value(confirm),
	}

	form := huh.NewForm(huh.NewGroup(fields...))
	return form.WithKeyMap(CreateKeyMapWithShiftEnter())
}
