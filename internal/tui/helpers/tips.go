package helpers

import (
	"fmt"
	"math/rand"
	"reflect"

	"github.com/thenoetrevino/paso/internal/config"
)

// Tip represents a single tip/trick with dynamic keybinding support
type Tip struct {
	// Template is format string
	Template string
	// KeyFields are the names of KeyMappings struct fields to substitute
	//
	// Example: []string{"CreateProject", "NextProject"}
	KeyFields []string
}

// TipGenerator generates tip text with user's actual keybindings
type TipGenerator struct {
	keyMappings *config.KeyMappings
	tips        []Tip
}

// NewTipGenerator creates a generator with all available tips
func NewTipGenerator(keyMappings *config.KeyMappings) *TipGenerator {
	return &TipGenerator{
		keyMappings: keyMappings,
		tips: []Tip{
			{
				Template:  "Press '%s' to create a new project, '%s'/'%s' to switch projects",
				KeyFields: []string{"CreateProject", "NextProject", "PrevProject"},
			},
			{
				Template:  "Press '%s' to add task, '%s' to edit, '%s' to delete",
				KeyFields: []string{"AddTask", "EditTask", "DeleteTask"},
			},
			{
				Template:  "Navigate with '%s'/'%s' (columns), '%s'/'%s' (tasks)",
				KeyFields: []string{"PrevColumn", "NextColumn", "PrevTask", "NextTask"},
			},
			{
				Template:  "Move tasks with '%s'/'%s' (horizontal), '%s'/'%s' (vertical)",
				KeyFields: []string{"MoveTaskLeft", "MoveTaskRight", "MoveTaskUp", "MoveTaskDown"},
			},
			{
				Template:  "Press '%s' to create column, '%s' to rename, '%s' to delete",
				KeyFields: []string{"CreateColumn", "RenameColumn", "DeleteColumn"},
			},
			{
				Template:  "Press '%s' to view task details, '%s' to toggle view mode",
				KeyFields: []string{"ViewTask", "ToggleView"},
			},
			{
				Template:  "Use '%s'/'%s' to scroll viewport when many columns exist",
				KeyFields: []string{"ScrollViewportLeft", "ScrollViewportRight"},
			},
			{
				Template:  "Press '%s' for help screen with all keyboard shortcuts",
				KeyFields: []string{"ShowHelp"},
			},
		},
	}
}

// SelectRandom picks a random tip and generates text with user's keybindings
func (tg *TipGenerator) SelectRandom() string {
	tip := tg.tips[rand.Intn(len(tg.tips))]
	return tg.generateTipText(tip)
}

// generateTipText replaces placeholders with actual keybindings
func (tg *TipGenerator) generateTipText(tip Tip) string {
	keys := make([]any, len(tip.KeyFields))
	for i, fieldName := range tip.KeyFields {
		keys[i] = tg.getKeyBinding(fieldName)
	}

	return fmt.Sprintf(tip.Template, keys...)
}

// getKeyBinding retrieves keybinding value by field name using reflection
func (tg *TipGenerator) getKeyBinding(fieldName string) string {
	v := reflect.ValueOf(tg.keyMappings).Elem()
	field := v.FieldByName(fieldName)
	if !field.IsValid() {
		return "?"
	}
	return field.String()
}
