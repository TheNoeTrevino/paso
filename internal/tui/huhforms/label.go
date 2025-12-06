package huhforms

import "charm.land/huh/v2"

// LabelColorOptions returns the available color options for labels
func LabelColorOptions() []huh.Option[int] {
	return []huh.Option[int]{
		huh.NewOption("Purple", 0),
		huh.NewOption("Blue", 1),
		huh.NewOption("Green", 2),
		huh.NewOption("Yellow", 3),
		huh.NewOption("Orange", 4),
		huh.NewOption("Red", 5),
		huh.NewOption("Pink", 6),
		huh.NewOption("Cyan", 7),
		huh.NewOption("Gray", 8),
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

// CreateLabelForm creates a huh form for adding/editing a label
func CreateLabelForm(
	name *string,
	colorIndex *int,
) *huh.Form {
	if colorIndex == nil {
		defaultIdx := 0
		colorIndex = &defaultIdx
	}

	fields := []huh.Field{
		huh.NewInput().
			Key("name").
			Title("Label Name").
			Placeholder("Enter label name...").
			Value(name),

		huh.NewSelect[int]().
			Key("color").
			Title("Color").
			Options(LabelColorOptions()...).
			Value(colorIndex),
	}

	return huh.NewForm(huh.NewGroup(fields...))
}
