package tui

import (
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui/forms"
)

// CreateTicketForm creates a form for adding/editing a ticket
// The form uses pointers to the model's fields so they're updated in place
func CreateTicketForm(
	title *string,
	description *string,
	labelIDs *[]int,
	labels []*models.Label,
	confirm *bool,
) *forms.Form {
	// Build form fields
	var formFields []forms.Field

	formFields = append(formFields,
		forms.NewTextInput("title", "Title", "Enter task title...", title),
		forms.NewTextArea("description", "Description", "Enter task description...", 500, description),
	)

	// Only add label multi-select if there are labels available
	if len(labels) > 0 {
		var labelOptions []forms.Option
		for _, label := range labels {
			labelOptions = append(labelOptions, forms.Option{
				Label: label.Name,
				Value: label.ID,
			})
		}

		formFields = append(formFields,
			forms.NewMultiSelect("labels", "Labels (space to toggle, up/down to navigate)", labelOptions, labelIDs),
		)
	}

	// Add confirmation at the end
	formFields = append(formFields,
		forms.NewConfirm("confirm", "Submit this ticket?", "Yes", "No", confirm),
	)

	return forms.NewForm(formFields...)
}

// CreateProjectForm creates a form for adding a new project
func CreateProjectForm(
	name *string,
	description *string,
) *forms.Form {
	formFields := []forms.Field{
		forms.NewTextInput("name", "Project Name", "Enter project name...", name),
		forms.NewTextArea("description", "Description (optional)", "Enter project description...", 500, description),
	}

	return forms.NewForm(formFields...)
}

// LabelColorOptions returns the available color options for labels
func LabelColorOptions() []forms.Option {
	return []forms.Option{
		{Label: "Purple", Value: 0},
		{Label: "Blue", Value: 1},
		{Label: "Green", Value: 2},
		{Label: "Yellow", Value: 3},
		{Label: "Orange", Value: 4},
		{Label: "Red", Value: 5},
		{Label: "Pink", Value: 6},
		{Label: "Cyan", Value: 7},
		{Label: "Gray", Value: 8},
	}
}

// LabelColorValues maps option values to hex colors
func LabelColorValues() map[int]string {
	return map[int]string{
		0: "#7D56F4", // Purple
		1: "#3B82F6", // Blue
		2: "#22C55E", // Green
		3: "#EAB308", // Yellow
		4: "#F97316", // Orange
		5: "#EF4444", // Red
		6: "#EC4899", // Pink
		7: "#06B6D4", // Cyan
		8: "#6B7280", // Gray
	}
}

// CreateLabelForm creates a form for adding/editing a label
// Note: For labels, we'll use a single-select implemented as a MultiSelect with max 1
func CreateLabelForm(
	name *string,
	colorIndex *int,
) *forms.Form {
	if colorIndex == nil {
		defaultIdx := 0
		colorIndex = &defaultIdx
	}

	// Convert colorIndex to []int for multiselect (we'll only allow 1 selection)
	selectedColors := []int{*colorIndex}

	formFields := []forms.Field{
		forms.NewTextInput("name", "Label Name", "Enter label name...", name),
		forms.NewMultiSelect("color", "Color (select one)", LabelColorOptions(), &selectedColors),
	}

	return forms.NewForm(formFields...)
}
