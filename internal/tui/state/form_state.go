package state

import (
	"github.com/charmbracelet/huh"
)

// FormState manages all form-related state for the application.
// This includes the huh forms for tickets, projects, and labels,
// as well as their associated field values and editing state.
type FormState struct {
	// Ticket form fields (for creating/editing tasks)
	TicketForm      *huh.Form // The huh form instance
	EditingTaskID   int       // ID of task being edited (0 for new task)
	FormTitle       string    // Form field: task title
	FormDescription string    // Form field: task description
	FormLabelIDs    []int     // Form field: selected label IDs
	FormConfirm     bool      // Form field: confirmation (submit vs cancel)

	// Project form fields (for creating projects)
	ProjectForm            *huh.Form // The huh form instance
	FormProjectName        string    // Form field: project name
	FormProjectDescription string    // Form field: project description

	// Label form fields (for creating/editing labels)
	LabelForm         *huh.Form // The huh form instance
	EditingLabelID    int       // ID of label being edited (0 for new label)
	FormLabelName     string    // Form field: label name
	FormLabelColor    string    // Form field: label color (hex code)
	SelectedLabelIdx  int       // Index of selected label in label list
	LabelListMode     string    // Sub-mode: "list", "add", "edit", "delete"

	// Label assignment fields (for quick label toggling)
	AssigningLabelIDs []int // Currently selected labels for assignment
}

// NewFormState creates a new FormState with default values.
func NewFormState() *FormState {
	return &FormState{
		TicketForm:             nil,
		EditingTaskID:          0,
		FormTitle:              "",
		FormDescription:        "",
		FormLabelIDs:           []int{},
		FormConfirm:            true,
		ProjectForm:            nil,
		FormProjectName:        "",
		FormProjectDescription: "",
		LabelForm:              nil,
		EditingLabelID:         0,
		FormLabelName:          "",
		FormLabelColor:         "",
		SelectedLabelIdx:       0,
		LabelListMode:          "",
		AssigningLabelIDs:      []int{},
	}
}

// --- Ticket Form Methods ---

// ClearTicketForm resets all ticket form fields to their default values.
func (s *FormState) ClearTicketForm() {
	s.TicketForm = nil
	s.EditingTaskID = 0
	s.FormTitle = ""
	s.FormDescription = ""
	s.FormLabelIDs = []int{}
	s.FormConfirm = true
}

// IsTicketFormActive returns true if a ticket form is currently active.
func (s *FormState) IsTicketFormActive() bool {
	return s.TicketForm != nil
}

// --- Project Form Methods ---

// ClearProjectForm resets all project form fields to their default values.
func (s *FormState) ClearProjectForm() {
	s.ProjectForm = nil
	s.FormProjectName = ""
	s.FormProjectDescription = ""
}

// IsProjectFormActive returns true if a project form is currently active.
func (s *FormState) IsProjectFormActive() bool {
	return s.ProjectForm != nil
}

// --- Label Form Methods ---

// ClearLabelForm resets all label form fields to their default values.
func (s *FormState) ClearLabelForm() {
	s.LabelForm = nil
	s.EditingLabelID = 0
	s.FormLabelName = ""
	s.FormLabelColor = ""
	s.SelectedLabelIdx = 0
	s.LabelListMode = ""
}

// IsLabelFormActive returns true if a label form is currently active.
func (s *FormState) IsLabelFormActive() bool {
	return s.LabelForm != nil
}

// --- Label Assignment Methods ---

// ClearAssigningLabels resets the label assignment state.
func (s *FormState) ClearAssigningLabels() {
	s.AssigningLabelIDs = []int{}
}
