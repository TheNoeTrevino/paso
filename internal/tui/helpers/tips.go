package helpers

import (
	"fmt"
	"math/rand"
	"reflect"
	"strings"

	"github.com/thenoetrevino/paso/internal/config"
)

// Tip represents a single tip/trick with dynamic keybinding support
type Tip struct {
	// Template is format string
	Template string
	// KeyFields are dot-separated paths into the nested KeyMappings struct.
	//
	// Example: []string{"Projects.CreateProject", "Navigation.NextProject"}
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
				KeyFields: []string{"Projects.CreateProject", "Navigation.NextProject", "Navigation.PrevProject"},
			},
			{
				Template:  "Press '%s' to add task, '%s' to edit, '%s' to delete",
				KeyFields: []string{"Tasks.AddTask", "Tasks.EditTask", "Tasks.DeleteTask"},
			},
			{
				Template:  "Navigate with '%s'/'%s' (columns), '%s'/'%s' (tasks)",
				KeyFields: []string{"Navigation.MoveLeft", "Navigation.MoveRight", "Navigation.MoveUp", "Navigation.MoveDown"},
			},
			{
				Template:  "Move tasks with '%s'/'%s' (horizontal), '%s'/'%s' (vertical)",
				KeyFields: []string{"Kanban.MoveTaskLeft", "Kanban.MoveTaskRight", "Kanban.MoveTaskUp", "Kanban.MoveTaskDown"},
			},
			{
				Template:  "Press '%s' to create column, '%s' to rename, '%s' to delete",
				KeyFields: []string{"Kanban.CreateColumn", "Kanban.RenameColumn", "Kanban.DeleteColumn"},
			},
			{
				Template:  "Press '%s' to view task details, '%s' to toggle view mode",
				KeyFields: []string{"Tasks.ViewTask", "General.ToggleView"},
			},
			{
				Template:  "Use '%s'/'%s' to scroll viewport when many columns exist",
				KeyFields: []string{"Navigation.ScrollViewportLeft", "Navigation.ScrollViewportRight"},
			},
			{
				Template:  "Press '%s' for help screen with all keyboard shortcuts",
				KeyFields: []string{"General.ShowHelp"},
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

// getKeyBinding retrieves a keybinding value by a dot-separated path
// into the nested KeyMappings struct (e.g. "Navigation.MoveUp").
func (tg *TipGenerator) getKeyBinding(path string) string {
	v := reflect.ValueOf(tg.keyMappings).Elem()

	for segment := range strings.SplitSeq(path, ".") {
		v = v.FieldByName(segment)
		if !v.IsValid() {
			return "?"
		}
	}

	return v.String()
}
