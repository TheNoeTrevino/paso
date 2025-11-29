package state

import (
	"github.com/charmbracelet/huh"
)

// FormState manages all form-related state for the application.
// This includes the huh forms for tickets, projects, and labels,
// as well as their associated field values and editing state.
type FormState struct {
	// Ticket form fields (for creating/editing tasks)
	ticketForm      *huh.Form // The huh form instance
	editingTaskID   int       // ID of task being edited (0 for new task)
	formTitle       string    // Form field: task title
	formDescription string    // Form field: task description
	formLabelIDs    []int     // Form field: selected label IDs
	formConfirm     bool      // Form field: confirmation (submit vs cancel)

	// Project form fields (for creating projects)
	projectForm            *huh.Form // The huh form instance
	formProjectName        string    // Form field: project name
	formProjectDescription string    // Form field: project description

	// Label form fields (for creating/editing labels)
	labelForm         *huh.Form // The huh form instance
	editingLabelID    int       // ID of label being edited (0 for new label)
	formLabelName     string    // Form field: label name
	formLabelColor    string    // Form field: label color (hex code)
	selectedLabelIdx  int       // Index of selected label in label list
	labelListMode     string    // Sub-mode: "list", "add", "edit", "delete"

	// Label assignment fields (for quick label toggling)
	assigningLabelIDs []int // Currently selected labels for assignment
}

// NewFormState creates a new FormState with default values.
func NewFormState() *FormState {
	return &FormState{
		ticketForm:             nil,
		editingTaskID:          0,
		formTitle:              "",
		formDescription:        "",
		formLabelIDs:           []int{},
		formConfirm:            true,
		projectForm:            nil,
		formProjectName:        "",
		formProjectDescription: "",
		labelForm:              nil,
		editingLabelID:         0,
		formLabelName:          "",
		formLabelColor:         "",
		selectedLabelIdx:       0,
		labelListMode:          "",
		assigningLabelIDs:      []int{},
	}
}

// --- Ticket Form Methods ---

// TicketForm returns the current ticket form instance.
func (s *FormState) TicketForm() *huh.Form {
	return s.ticketForm
}

// SetTicketForm sets the ticket form instance.
func (s *FormState) SetTicketForm(form *huh.Form) {
	s.ticketForm = form
}

// EditingTaskID returns the ID of the task being edited.
func (s *FormState) EditingTaskID() int {
	return s.editingTaskID
}

// SetEditingTaskID sets the task ID being edited.
func (s *FormState) SetEditingTaskID(id int) {
	s.editingTaskID = id
}

// FormTitle returns the current form title value.
func (s *FormState) FormTitle() string {
	return s.formTitle
}

// SetFormTitle sets the form title value.
func (s *FormState) SetFormTitle(title string) {
	s.formTitle = title
}

// FormDescription returns the current form description value.
func (s *FormState) FormDescription() string {
	return s.formDescription
}

// SetFormDescription sets the form description value.
func (s *FormState) SetFormDescription(description string) {
	s.formDescription = description
}

// FormLabelIDs returns the selected label IDs.
func (s *FormState) FormLabelIDs() []int {
	return s.formLabelIDs
}

// SetFormLabelIDs sets the selected label IDs.
func (s *FormState) SetFormLabelIDs(ids []int) {
	s.formLabelIDs = ids
}

// FormConfirm returns the confirmation value.
func (s *FormState) FormConfirm() bool {
	return s.formConfirm
}

// SetFormConfirm sets the confirmation value.
func (s *FormState) SetFormConfirm(confirm bool) {
	s.formConfirm = confirm
}

// ClearTicketForm resets all ticket form fields to their default values.
func (s *FormState) ClearTicketForm() {
	s.ticketForm = nil
	s.editingTaskID = 0
	s.formTitle = ""
	s.formDescription = ""
	s.formLabelIDs = []int{}
	s.formConfirm = true
}

// IsTicketFormActive returns true if a ticket form is currently active.
func (s *FormState) IsTicketFormActive() bool {
	return s.ticketForm != nil
}

// --- Project Form Methods ---

// ProjectForm returns the current project form instance.
func (s *FormState) ProjectForm() *huh.Form {
	return s.projectForm
}

// SetProjectForm sets the project form instance.
func (s *FormState) SetProjectForm(form *huh.Form) {
	s.projectForm = form
}

// FormProjectName returns the current project name value.
func (s *FormState) FormProjectName() string {
	return s.formProjectName
}

// SetFormProjectName sets the project name value.
func (s *FormState) SetFormProjectName(name string) {
	s.formProjectName = name
}

// FormProjectDescription returns the current project description value.
func (s *FormState) FormProjectDescription() string {
	return s.formProjectDescription
}

// SetFormProjectDescription sets the project description value.
func (s *FormState) SetFormProjectDescription(description string) {
	s.formProjectDescription = description
}

// ClearProjectForm resets all project form fields to their default values.
func (s *FormState) ClearProjectForm() {
	s.projectForm = nil
	s.formProjectName = ""
	s.formProjectDescription = ""
}

// IsProjectFormActive returns true if a project form is currently active.
func (s *FormState) IsProjectFormActive() bool {
	return s.projectForm != nil
}

// --- Label Form Methods ---

// LabelForm returns the current label form instance.
func (s *FormState) LabelForm() *huh.Form {
	return s.labelForm
}

// SetLabelForm sets the label form instance.
func (s *FormState) SetLabelForm(form *huh.Form) {
	s.labelForm = form
}

// EditingLabelID returns the ID of the label being edited.
func (s *FormState) EditingLabelID() int {
	return s.editingLabelID
}

// SetEditingLabelID sets the label ID being edited.
func (s *FormState) SetEditingLabelID(id int) {
	s.editingLabelID = id
}

// FormLabelName returns the current label name value.
func (s *FormState) FormLabelName() string {
	return s.formLabelName
}

// SetFormLabelName sets the label name value.
func (s *FormState) SetFormLabelName(name string) {
	s.formLabelName = name
}

// FormLabelColor returns the current label color value.
func (s *FormState) FormLabelColor() string {
	return s.formLabelColor
}

// SetFormLabelColor sets the label color value.
func (s *FormState) SetFormLabelColor(color string) {
	s.formLabelColor = color
}

// SelectedLabelIdx returns the selected label index.
func (s *FormState) SelectedLabelIdx() int {
	return s.selectedLabelIdx
}

// SetSelectedLabelIdx sets the selected label index.
func (s *FormState) SetSelectedLabelIdx(idx int) {
	s.selectedLabelIdx = idx
}

// LabelListMode returns the current label list mode.
func (s *FormState) LabelListMode() string {
	return s.labelListMode
}

// SetLabelListMode sets the label list mode.
func (s *FormState) SetLabelListMode(mode string) {
	s.labelListMode = mode
}

// ClearLabelForm resets all label form fields to their default values.
func (s *FormState) ClearLabelForm() {
	s.labelForm = nil
	s.editingLabelID = 0
	s.formLabelName = ""
	s.formLabelColor = ""
	s.selectedLabelIdx = 0
	s.labelListMode = ""
}

// IsLabelFormActive returns true if a label form is currently active.
func (s *FormState) IsLabelFormActive() bool {
	return s.labelForm != nil
}

// --- Label Assignment Methods ---

// AssigningLabelIDs returns the currently selected labels for assignment.
func (s *FormState) AssigningLabelIDs() []int {
	return s.assigningLabelIDs
}

// SetAssigningLabelIDs sets the labels for assignment.
func (s *FormState) SetAssigningLabelIDs(ids []int) {
	s.assigningLabelIDs = ids
}

// ClearAssigningLabels resets the label assignment state.
func (s *FormState) ClearAssigningLabels() {
	s.assigningLabelIDs = []int{}
}
