package tui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/huh"
	"github.com/thenoetrevino/paso/internal/models"
)

// CreateTicketForm creates a huh form for adding/editing a ticket
// The form uses pointers to the model's fields so they're updated in place
func CreateTicketForm(
	title *string,
	description *string,
	labelIDs *[]int,
	labels []*models.Label,
	confirm *bool,
) *huh.Form {
	// Convert labels to huh options
	var labelOptions []huh.Option[int]
	for _, label := range labels {
		labelOptions = append(labelOptions, huh.NewOption(label.Name, label.ID))
	}

	// Build form fields
	var formFields []huh.Field

	formFields = append(formFields,
		huh.NewInput().
			Key("title").
			Title("Title").
			Placeholder("Enter task title...").
			Value(title),

		huh.NewText().
			Key("description").
			Title("Description").
			Placeholder("Enter task description...").
			CharLimit(500).
			Value(description),
	)

	// Only add label multi-select if there are labels available
	if len(labelOptions) > 0 {
		formFields = append(formFields,
			huh.NewMultiSelect[int]().
				Key("labels").
				Title("Labels").
				Options(labelOptions...).
				Value(labelIDs),
		)
	}

	// Add confirmation at the end
	formFields = append(formFields,
		huh.NewConfirm().
			Key("confirm").
			Title("Submit this ticket?").
			Affirmative("Yes").
			Negative("No").
			Value(confirm),
	)

	// Create custom keymap where Enter creates newlines in Text field
	// and Tab moves between fields
	customKeyMap := huh.NewDefaultKeyMap()

	// Override Text field keybindings:
	// - Enter creates a new line (instead of moving to next field)
	// - Tab moves to next field
	customKeyMap.Text.Next = key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next"),
	)
	customKeyMap.Text.NewLine = key.NewBinding(
		key.WithKeys("enter", "ctrl+j"),
		key.WithHelp("enter", "new line"),
	)

	form := huh.NewForm(
		huh.NewGroup(formFields...),
	).WithTheme(huh.ThemeDracula()).
		WithKeyMap(customKeyMap)

	return form
}

// CreateProjectForm creates a huh form for adding a new project
func CreateProjectForm(
	name *string,
	description *string,
) *huh.Form {
	form := huh.NewForm(
		huh.NewGroup(
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
				Value(description),
		),
	).WithTheme(huh.ThemeDracula())

	return form
}

// LabelColorOptions returns the available color options for labels
func LabelColorOptions() []huh.Option[string] {
	return []huh.Option[string]{
		huh.NewOption("Purple", "#7D56F4"),
		huh.NewOption("Blue", "#3B82F6"),
		huh.NewOption("Green", "#22C55E"),
		huh.NewOption("Yellow", "#EAB308"),
		huh.NewOption("Orange", "#F97316"),
		huh.NewOption("Red", "#EF4444"),
		huh.NewOption("Pink", "#EC4899"),
		huh.NewOption("Cyan", "#06B6D4"),
		huh.NewOption("Gray", "#6B7280"),
	}
}

// CreateLabelForm creates a huh form for adding/editing a label
func CreateLabelForm(
	name *string,
	color *string,
) *huh.Form {
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Key("name").
				Title("Label Name").
				Placeholder("Enter label name...").
				Value(name),

			huh.NewSelect[string]().
				Key("color").
				Title("Color").
				Options(LabelColorOptions()...).
				Value(color),
		),
	).WithTheme(huh.ThemeDracula())

	return form
}
