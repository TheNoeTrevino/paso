package cli

import (
	"charm.land/huh/v2"
)

// Pickable is implemented by any type that can be presented in an interactive picker.
// The interface lives here in the consumer package (cli), following Go convention.
// Implementations live on the concrete model structs in the models package.
type Pickable interface {
	PickerLabel() string
	PickerValue() string
}

// BuildPickerOptions converts a slice of Pickable items into huh.Option items for an interactive picker.
func BuildPickerOptions[T Pickable](items []T) []huh.Option[string] {
	options := make([]huh.Option[string], len(items))
	for i, item := range items {
		options[i] = huh.NewOption(item.PickerLabel(), item.PickerValue())
	}
	return options
}

// RunPick presents an interactive picker with the given title and items,
// returning the selected item's value. The picker renders to stderr so only
// the selected value goes to stdout when used in command substitution.
func RunPick[T Pickable](title string, items []T) (string, error) {
	options := BuildPickerOptions(items)
	var selected string
	err := huh.NewSelect[string]().
		Title(title).
		Options(options...).
		Filtering(true).
		Value(&selected).
		Run()
	if err != nil {
		return "", err
	}
	return selected, nil
}
