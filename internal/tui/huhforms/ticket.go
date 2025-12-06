package huhforms

import (
	"charm.land/huh/v2"
	"github.com/thenoetrevino/paso/internal/models"
)

// CreateTicketForm creates a huh form for adding/editing a ticket
// The form uses pointers to update values in place, matching the existing pattern
func CreateTicketForm(
	title *string,
	description *string,
	labelIDs *[]int,
	labels []*models.Label,
	confirm *bool,
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

	// Description text area field
	fields = append(fields,
		huh.NewText().
			Key("description").
			Title("Description").
			Placeholder("Enter task description...").
			CharLimit(500).
			Lines(5).
			Value(description),
	)

	// Labels multi-select (only if labels exist)
	if len(labels) > 0 {
		var labelOptions []huh.Option[int]
		for _, label := range labels {
			labelOptions = append(labelOptions, huh.NewOption(label.Name, label.ID))
		}

		fields = append(fields,
			huh.NewMultiSelect[int]().
				Key("labels").
				Title("Labels").
				Description("Space to toggle, / to filter").
				Options(labelOptions...).
				Value(labelIDs).
				Filterable(true),
		)
	}

	// Confirmation
	fields = append(fields,
		huh.NewConfirm().
			Key("confirm").
			Title("Submit this ticket?").
			Affirmative("Yes").
			Negative("No").
			Value(confirm),
	)

	return huh.NewForm(huh.NewGroup(fields...))
}
