package huhforms

import "charm.land/huh/v2"

type ProjectFormProps struct {
	Name        *string
	Description *string
	Confirm     *bool
	IsEditing   bool
}

func CreateProjectForm(props ProjectFormProps) *huh.Form {
	confirmTitle := "Create this project?"
	if props.IsEditing {
		confirmTitle = "Save changes?"
	}

	fields := []huh.Field{
		huh.NewInput().
			Key("name").
			Title("Project Name").
			Placeholder("Enter project name...").
			Value(props.Name),

		huh.NewText().
			Key("description").
			Title("Description (optional)").
			Placeholder("Enter project description...").
			CharLimit(500).
			Lines(3).
			Value(props.Description),

		huh.NewConfirm().
			Key("confirm").
			Title(confirmTitle).
			Affirmative("Yes").
			Negative("No").
			Value(props.Confirm),
	}

	form := huh.NewForm(huh.NewGroup(fields...))
	return form.WithKeyMap(CreateKeyMapWithShiftEnter())
}
