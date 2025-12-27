package state

import (
	"strings"
	"time"

	"charm.land/huh/v2"
	"github.com/thenoetrevino/paso/internal/models"
)

// FormState manages all form-related state for the application.
// This includes the custom forms for tickets, projects, and labels,
// as well as their associated field values and editing state.
type FormState struct {
	// Ticket form fields (for creating/editing tasks)
	TicketForm      *huh.Form // The form instance
	EditingTaskID   int       // ID of task being edited (0 for new task)
	FormTitle       string    // Form field: task title
	FormDescription string    // Form field: task description
	FormLabelIDs    []int     // Form field: selected label IDs
	FormConfirm     bool      // Form field: confirmation (submit vs cancel)

	// Parent/child issue tracking for ticket form
	FormParentIDs  []int                   // Selected parent task IDs
	FormChildIDs   []int                   // Selected child task IDs
	FormParentRefs []*models.TaskReference // Parent task references for display
	FormChildRefs  []*models.TaskReference // Child task references for display

	// Comments/notes for ticket form
	FormComments        []*models.Comment // Current notes loaded from DB
	InitialFormComments []*models.Comment // Snapshot for change detection

	// Task metadata for display (edit mode only)
	FormCreatedAt           time.Time // Task creation timestamp (only populated in edit mode)
	FormUpdatedAt           time.Time // Task last update timestamp (only populated in edit mode)
	FormTypeDescription     string    // Task type (e.g., "task", "feature")
	FormPriorityDescription string    // Task priority (e.g., "low", "high", "critical")
	FormPriorityColor       string    // Task priority color (hex code)

	// Ticket form initial values (for change detection)
	InitialFormTitle       string // Initial title value when form was created
	InitialFormDescription string // Initial description value when form was created
	InitialFormLabelIDs    []int  // Initial label IDs when form was created
	InitialFormParentIDs   []int  // Initial parent IDs when form was created
	InitialFormChildIDs    []int  // Initial child IDs when form was created

	// Project form fields (for creating projects)
	ProjectForm            *huh.Form // The form instance
	FormProjectName        string    // Form field: project name
	FormProjectDescription string    // Form field: project description
	FormProjectConfirm     bool      // Form field: confirmation (submit vs cancel)

	// Project form initial values (for change detection)
	InitialFormProjectName        string // Initial project name when form was created
	InitialFormProjectDescription string // Initial project description when form was created

	// Label form fields (for creating/editing labels)
	LabelForm        *huh.Form // The form instance
	EditingLabelID   int       // ID of label being edited (0 for new label)
	FormLabelName    string    // Form field: label name
	FormLabelColor   string    // Form field: label color (hex code)
	SelectedLabelIdx int       // Index of selected label in label list
	LabelListMode    string    // Sub-mode: "list", "add", "edit", "delete"

	// Label assignment fields (for quick label toggling)
	AssigningLabelIDs []int // Currently selected labels for assignment

	// Column form fields (for creating/renaming columns)
	ColumnForm            *huh.Form // The form instance
	FormColumnName        string    // Form field: column name
	EditingColumnID       int       // ID of column being edited (0 for new column)
	InitialFormColumnName string    // Initial column name for change detection

	// Comment form fields (for creating/editing comments/notes)
	CommentForm               *huh.Form // The form instance
	FormCommentMessage        string    // Form field: comment message text
	EditingCommentID          int       // ID of comment being edited (0 for new comment)
	InitialFormCommentMessage string    // Initial comment message for change detection
}

// NewFormState creates a new FormState with default values.
func NewFormState() *FormState {
	return &FormState{
		TicketForm:                nil,
		EditingTaskID:             0,
		FormTitle:                 "",
		FormDescription:           "",
		FormLabelIDs:              []int{},
		FormConfirm:               true,
		FormParentIDs:             []int{},
		FormChildIDs:              []int{},
		FormParentRefs:            []*models.TaskReference{},
		FormChildRefs:             []*models.TaskReference{},
		FormComments:              []*models.Comment{},
		InitialFormComments:       []*models.Comment{},
		ProjectForm:               nil,
		FormProjectName:           "",
		FormProjectDescription:    "",
		FormProjectConfirm:        true,
		LabelForm:                 nil,
		EditingLabelID:            0,
		FormLabelName:             "",
		FormLabelColor:            "",
		SelectedLabelIdx:          0,
		LabelListMode:             "",
		AssigningLabelIDs:         []int{},
		ColumnForm:                nil,
		FormColumnName:            "",
		EditingColumnID:           0,
		InitialFormColumnName:     "",
		CommentForm:               nil,
		FormCommentMessage:        "",
		EditingCommentID:          0,
		InitialFormCommentMessage: "",
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
	s.FormParentIDs = []int{}
	s.FormChildIDs = []int{}
	s.FormParentRefs = []*models.TaskReference{}
	s.FormChildRefs = []*models.TaskReference{}
	s.FormComments = []*models.Comment{}
	s.InitialFormComments = []*models.Comment{}
	// Clear initial values
	s.InitialFormTitle = ""
	s.InitialFormDescription = ""
	s.InitialFormLabelIDs = []int{}
	s.InitialFormParentIDs = []int{}
	s.InitialFormChildIDs = []int{}
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
	s.FormProjectConfirm = true
	// Clear initial values
	s.InitialFormProjectName = ""
	s.InitialFormProjectDescription = ""
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

// --- Change Detection Methods ---

// HasTicketFormChanges returns true if the ticket form has unsaved changes.
// Compares current field values against initial snapshots.
func (s *FormState) HasTicketFormChanges() bool {
	if s.TicketForm == nil {
		return false
	}

	if strings.TrimSpace(s.FormTitle) != strings.TrimSpace(s.InitialFormTitle) {
		return true
	}

	if strings.TrimSpace(s.FormDescription) != strings.TrimSpace(s.InitialFormDescription) {
		return true
	}

	if !slicesEqual(s.FormLabelIDs, s.InitialFormLabelIDs) {
		return true
	}

	if !slicesEqual(s.FormParentIDs, s.InitialFormParentIDs) {
		return true
	}

	if !slicesEqual(s.FormChildIDs, s.InitialFormChildIDs) {
		return true
	}

	return false
}

// HasProjectFormChanges returns true if the project form has unsaved changes.
func (s *FormState) HasProjectFormChanges() bool {
	if s.ProjectForm == nil {
		return false
	}

	if strings.TrimSpace(s.FormProjectName) != strings.TrimSpace(s.InitialFormProjectName) {
		return true
	}

	if strings.TrimSpace(s.FormProjectDescription) != strings.TrimSpace(s.InitialFormProjectDescription) {
		return true
	}

	return false
}

// SnapshotTicketFormInitialValues stores current form values as initial state.
// Call this when the form is first created/initialized.
func (s *FormState) SnapshotTicketFormInitialValues() {
	s.InitialFormTitle = s.FormTitle
	s.InitialFormDescription = s.FormDescription
	s.InitialFormLabelIDs = append([]int{}, s.FormLabelIDs...)   // Copy slice
	s.InitialFormParentIDs = append([]int{}, s.FormParentIDs...) // Copy slice
	s.InitialFormChildIDs = append([]int{}, s.FormChildIDs...)   // Copy slice
}

// SnapshotProjectFormInitialValues stores current project form values as initial state.
func (s *FormState) SnapshotProjectFormInitialValues() {
	s.InitialFormProjectName = s.FormProjectName
	s.InitialFormProjectDescription = s.FormProjectDescription
}

// --- Helper Functions ---

// slicesEqual compares two int slices for equality (order-independent).
func slicesEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}

	// Create frequency maps
	aMap := make(map[int]int)
	bMap := make(map[int]int)

	for _, v := range a {
		aMap[v]++
	}
	for _, v := range b {
		bMap[v]++
	}

	// Compare maps
	for k, v := range aMap {
		if bMap[k] != v {
			return false
		}
	}

	return true
}

// --- Column Form Methods ---

// ClearColumnForm resets all column form fields to their default values.
func (s *FormState) ClearColumnForm() {
	s.ColumnForm = nil
	s.FormColumnName = ""
	s.EditingColumnID = 0
	s.InitialFormColumnName = ""
}

// SnapshotColumnFormInitialValues saves the current column form values for change detection.
func (s *FormState) SnapshotColumnFormInitialValues() {
	s.InitialFormColumnName = s.FormColumnName
}

// HasColumnFormChanges returns true if the column form has unsaved changes.
func (s *FormState) HasColumnFormChanges() bool {
	if s.ColumnForm == nil {
		return false
	}
	return strings.TrimSpace(s.FormColumnName) != strings.TrimSpace(s.InitialFormColumnName)
}

// --- Comment Form Methods ---

// ClearCommentForm resets all comment form fields to their default values.
func (s *FormState) ClearCommentForm() {
	s.CommentForm = nil
	s.FormCommentMessage = ""
	s.EditingCommentID = 0
	s.InitialFormCommentMessage = ""
}

// SnapshotCommentFormInitialValues saves the current comment form values for change detection.
func (s *FormState) SnapshotCommentFormInitialValues() {
	s.InitialFormCommentMessage = s.FormCommentMessage
}

// HasCommentFormChanges returns true if the comment form has unsaved changes.
func (s *FormState) HasCommentFormChanges() bool {
	if s.CommentForm == nil {
		return false
	}
	return strings.TrimSpace(s.FormCommentMessage) != strings.TrimSpace(s.InitialFormCommentMessage)
}
