package state

// PickerStates groups all picker-related states into a single struct.
// This reduces the number of fields in the Model and improves organization.
// Pickers are used for selecting items from lists (labels, priorities, etc.)
type PickerStates struct {
	Label        *LabelPickerState        // Picker for selecting labels on tasks
	Parent       *TaskPickerState         // Picker for selecting parent task relationships
	Child        *TaskPickerState         // Picker for selecting child task relationships
	Priority     *PriorityPickerState     // Picker for selecting task priority
	Type         *TypePickerState         // Picker for selecting task type (task, bug, feature, etc.)
	RelationType *RelationTypePickerState // Picker for selecting relationship types (blocking, related, etc.)
	Status       *StatusPickerState       // Picker for selecting task status/column
}

// NewPickerStates creates a new PickerStates instance with all pickers initialized.
func NewPickerStates() *PickerStates {
	return &PickerStates{
		Label:        NewLabelPickerState(),
		Parent:       NewTaskPickerState(),
		Child:        NewTaskPickerState(),
		Priority:     NewPriorityPickerState(),
		Type:         NewTypePickerState(),
		RelationType: NewRelationTypePickerState(),
		Status:       NewStatusPickerState(),
	}
}
