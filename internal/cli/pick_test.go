package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockPickable struct {
	label string
	value string
}

func (m mockPickable) PickerLabel() string { return m.label }
func (m mockPickable) PickerValue() string { return m.value }

func TestBuildPickerOptions(t *testing.T) {
	items := []mockPickable{
		{label: "Fix login bug (#14) — high, bug", value: "14"},
		{label: "Add dark mode (#15) — medium, feature", value: "15"},
	}

	options := BuildPickerOptions(items)

	assert.Len(t, options, 2)
	assert.Equal(t, "Fix login bug (#14) — high, bug", options[0].Key)
	assert.Equal(t, "14", options[0].Value)
	assert.Equal(t, "Add dark mode (#15) — medium, feature", options[1].Key)
	assert.Equal(t, "15", options[1].Value)
}

func TestBuildPickOptionsEmpty(t *testing.T) {
	options := BuildPickerOptions([]mockPickable{})
	assert.Empty(t, options)
}

func TestBuildPickOptionsSingle(t *testing.T) {
	items := []mockPickable{
		{label: "Only item (#1)", value: "1"},
	}

	options := BuildPickerOptions(items)

	assert.Len(t, options, 1)
	assert.Equal(t, "Only item (#1)", options[0].Key)
	assert.Equal(t, "1", options[0].Value)
}

func TestBuildPickOptionsWithPointers(t *testing.T) {
	items := []*mockPickable{
		{label: "Pointer item (#1)", value: "1"},
		{label: "Pointer item (#2)", value: "2"},
	}

	options := BuildPickerOptions(items)

	assert.Len(t, options, 2)
	assert.Equal(t, "Pointer item (#1)", options[0].Key)
	assert.Equal(t, "1", options[0].Value)
	assert.Equal(t, "Pointer item (#2)", options[1].Key)
	assert.Equal(t, "2", options[1].Value)
}
