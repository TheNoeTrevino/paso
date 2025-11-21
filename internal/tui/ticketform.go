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
