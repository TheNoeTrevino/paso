package components

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/thenoetrevino/paso/internal/tui/state"
	"github.com/thenoetrevino/paso/internal/tui/theme"
)

// FilterBarProps contains the data needed to render the filter bar.
type FilterBarProps struct {
	Filter       *state.FilterState
	Focused      bool     // Whether the filter bar has keyboard focus
	Width        int      // Total available width
	PriorityName string   // e.g. "High", empty if no filter
	TypeName     string   // e.g. "Bug", empty if no filter
	AssigneeName string   // e.g. "John", empty if no filter
	LabelNames   []string // e.g. ["bug", "frontend"], empty if no filter
}

// RenderFilterBar renders the filter bar with chip-style buttons.
func RenderFilterBar(props FilterBarProps) string {
	if props.Filter == nil {
		return ""
	}

	chipTexts := make([]string, state.FilterChipCount)

	chipTexts[state.FilterChipLabel] = buildChipText("Label", strings.Join(props.LabelNames, ", "))
	chipTexts[state.FilterChipPriority] = buildChipText("Priority", props.PriorityName)
	chipTexts[state.FilterChipType] = buildChipText("Type", props.TypeName)
	chipTexts[state.FilterChipAssignee] = buildChipText("Assignee", props.AssigneeName)
	chipTexts[state.FilterChipClearAll] = "Clear All"

	var chips []string
	for i := range int(state.FilterChipCount) {
		chip := state.FilterChip(i)
		text := chipTexts[i]
		styled := styleChip(text, chip, props)
		chips = append(chips, styled)
	}

	row := strings.Join(chips, "  ")

	prefix := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Subtle)).
		Render("Filter: ")

	line := prefix + row

	barStyle := lipgloss.NewStyle().
		Width(props.Width).
		PaddingLeft(1)

	return barStyle.Render(line)
}

// buildChipText creates the display text for a filter chip.
func buildChipText(field string, value string) string {
	if value == "" {
		return field
	}
	return fmt.Sprintf("%s: %s", field, value)
}

// styleChip applies the appropriate style to a chip based on its state.
func styleChip(text string, chip state.FilterChip, props FilterBarProps) string {
	hasValue := chipHasValue(chip, props)
	isFocused := props.Focused && props.Filter.FocusedChip == chip

	style := lipgloss.NewStyle()

	switch {
	case isFocused && hasValue:
		style = style.
			Bold(true).
			Foreground(lipgloss.Color(theme.Highlight)).
			Underline(true)
	case isFocused:
		style = style.
			Foreground(lipgloss.Color(theme.Normal)).
			Underline(true)
	case hasValue:
		style = style.
			Bold(true).
			Foreground(lipgloss.Color(theme.Highlight))
	default:
		style = style.
			Foreground(lipgloss.Color(theme.Subtle))
	}

	return style.Render(text)
}

// chipHasValue returns true if the given chip has an active filter value.
func chipHasValue(chip state.FilterChip, props FilterBarProps) bool {
	switch chip {
	case state.FilterChipLabel:
		return len(props.LabelNames) > 0
	case state.FilterChipPriority:
		return props.PriorityName != ""
	case state.FilterChipType:
		return props.TypeName != ""
	case state.FilterChipAssignee:
		return props.AssigneeName != ""
	case state.FilterChipClearAll:
		return props.Filter.HasAnyFilter()
	}
	return false
}
