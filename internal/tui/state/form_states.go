package state

// FormStates groups all form-related states into a single struct.
// This reduces the number of fields in the Model and improves organization.
// Forms are used for creating/editing entities (tasks, projects, labels, comments)
type FormStates struct {
	Form    *FormState    // Main task form state (title, description, labels, etc.)
	Input   *InputState   // Input field state (editing mode, cursor position, etc.)
	Comment *CommentState // Comment state (for managing comments on tasks)
}

// NewFormStates creates a new FormStates instance with all form states initialized.
func NewFormStates() *FormStates {
	return &FormStates{
		Form:    NewFormState(),
		Input:   NewInputState(),
		Comment: NewCommentState(),
	}
}
