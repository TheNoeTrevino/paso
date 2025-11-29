package tui

import (
	"database/sql"
	"log"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/thenoetrevino/paso/internal/database"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// Mode represents the current interaction mode of the TUI
type Mode int

const (
	NormalMode              Mode = iota // Default navigation mode
	DeleteConfirmMode                   // Confirming task deletion
	AddColumnMode                       // Creating a new column
	EditColumnMode                      // Renaming an existing column
	DeleteColumnConfirmMode             // Confirming column deletion
	HelpMode                            // Displaying help screen
	ViewTaskMode                        // Viewing full task details
	TicketFormMode                      // Full ticket form with huh
	ProjectFormMode                     // Creating a new project with huh
	LabelManagementMode                 // Managing labels (create/edit/delete)
	LabelAssignMode                     // Quick label assignment to task
	LabelPickerMode                     // GitHub-style label picker popup
)

// Model represents the application state for the TUI
type Model struct {
	db                    *sql.DB                       // Database connection
	columns               []*models.Column              // All columns ordered by position
	tasks                 map[int][]*models.TaskSummary // Task summaries organized by column ID
	labels                []*models.Label               // All available labels
	selectedColumn        int                           // Index of currently selected column
	selectedTask          int                           // Index of currently selected task in the column
	width                 int                           // Terminal width
	height                int                           // Terminal height
	mode                  Mode                          // Current interaction mode
	inputBuffer           string                        // Text being typed in input mode
	inputPrompt           string                        // Prompt to show in input dialog
	viewportOffset        int                           // Index of leftmost visible column
	viewportSize          int                           // Number of columns that fit on screen
	deleteColumnTaskCount int                           // Number of tasks in column being deleted
	errorMessage          string                        // Current error message to display
	viewingTask           *models.TaskDetail            // Task being viewed in detail mode

	// Project fields
	projects        []*models.Project // All available projects
	selectedProject int               // Index of currently selected project

	// Ticket form fields (for huh form)
	ticketForm      *huh.Form // The huh form for adding/editing tickets
	editingTaskID   int       // ID of task being edited (0 for new task)
	formTitle       string    // Form field: title
	formDescription string    // Form field: description
	formLabelIDs    []int     // Form field: selected label IDs
	formConfirm     bool      // Form field: confirmation

	// Project form fields (for huh form)
	projectForm            *huh.Form // The huh form for creating projects
	formProjectName        string    // Form field: project name
	formProjectDescription string    // Form field: project description

	// Label management fields
	labelForm         *huh.Form // The huh form for creating/editing labels
	editingLabelID    int       // ID of label being edited (0 for new label)
	formLabelName     string    // Form field: label name
	formLabelColor    string    // Form field: label color
	selectedLabelIdx  int       // Index of selected label in label list
	labelListMode     string    // "list", "add", "edit", "delete" - sub-mode for label management

	// Label assignment fields (quick toggle)
	assigningLabelIDs []int // Currently selected labels for assignment

	// Label picker fields
	labelPickerItems      []LabelPickerItem // Items in the label picker
	labelPickerCursor     int               // Current cursor position in picker
	labelPickerFilter     string            // Filter text for label search
	labelPickerTaskID     int               // Task being edited in picker
	labelPickerColorIdx   int               // Cursor for color picker
	labelPickerCreateMode bool              // True when in color selection for new label

	// New state management fields (Phase 2: Refactoring to avoid God Object)
	// These will eventually replace the individual fields above
	appState         *state.AppState
	uiState          *state.UIState
	inputState       *state.InputState
	formState        *state.FormState
	labelPickerState *state.LabelPickerState
	errorState       *state.ErrorState
}

// InitialModel creates and initializes the TUI model with data from the database
func InitialModel(db *sql.DB) Model {
	// Load all projects
	projects, err := database.GetAllProjects(db)
	if err != nil {
		log.Printf("Error loading projects: %v", err)
		projects = []*models.Project{}
	}

	// Get the first project's ID (or 0 if no projects)
	var currentProjectID int
	if len(projects) > 0 {
		currentProjectID = projects[0].ID
	}

	// Load columns for the current project
	columns, err := database.GetColumnsByProject(db, currentProjectID)
	if err != nil {
		log.Printf("Error loading columns: %v", err)
		columns = []*models.Column{}
	}

	// Load task summaries for each column (includes labels)
	tasks := make(map[int][]*models.TaskSummary)
	for _, col := range columns {
		columnTasks, err := database.GetTaskSummariesByColumn(db, col.ID)
		if err != nil {
			log.Printf("Error loading tasks for column %d: %v", col.ID, err)
			columnTasks = []*models.TaskSummary{}
		}
		tasks[col.ID] = columnTasks
	}

	// Load labels for the current project
	labels, err := database.GetLabelsByProject(db, currentProjectID)
	if err != nil {
		log.Printf("Error loading labels: %v", err)
		labels = []*models.Label{}
	}

	// Initialize new state objects
	appState := state.NewAppState(projects, 0, columns, tasks, labels)
	uiState := state.NewUIState()
	inputState := state.NewInputState()
	formState := state.NewFormState()
	labelPickerState := state.NewLabelPickerState()
	errorState := state.NewErrorState()

	return Model{
		db:              db,
		projects:        projects,
		selectedProject: 0,
		columns:         columns,
		tasks:           tasks,
		labels:          labels,
		selectedColumn:  0,
		selectedTask:    0,
		width:           0,
		height:          0,
		mode:            NormalMode,
		inputBuffer:     "",
		inputPrompt:     "",
		viewportOffset:  0,
		viewportSize:    1, // Default to 1, will be recalculated when width is set

		// New state fields
		appState:         appState,
		uiState:          uiState,
		inputState:       inputState,
		formState:        formState,
		labelPickerState: labelPickerState,
		errorState:       errorState,
	}
}

// Init initializes the Bubble Tea application
// Required by tea.Model interface
func (m Model) Init() tea.Cmd {
	// No initial commands needed yet
	return nil
}

// getCurrentTasks returns the task summaries for the currently selected column
// Returns an empty slice if the column has no tasks
func (m Model) getCurrentTasks() []*models.TaskSummary {
	if len(m.columns) == 0 {
		return []*models.TaskSummary{}
	}
	currentCol := m.columns[m.selectedColumn]
	tasks := m.tasks[currentCol.ID]
	if tasks == nil {
		return []*models.TaskSummary{}
	}
	return tasks
}

// getCurrentTask returns the currently selected task summary
// Returns nil if there are no tasks in the current column or no columns exist
func (m Model) getCurrentTask() *models.TaskSummary {
	tasks := m.getCurrentTasks()
	if len(tasks) == 0 {
		return nil
	}
	if m.selectedTask >= len(tasks) {
		return nil
	}
	return tasks[m.selectedTask]
}

// removeCurrentTask removes the currently selected task from the model's local state
// This should be called after successfully deleting a task from the database
// It adjusts the selectedTask index if necessary to keep it within bounds
func (m *Model) removeCurrentTask() {
	if len(m.columns) == 0 {
		return
	}

	currentCol := m.columns[m.selectedColumn]
	tasks := m.tasks[currentCol.ID]

	if len(tasks) == 0 || m.selectedTask >= len(tasks) {
		return
	}

	// Remove the task at selectedTask index
	m.tasks[currentCol.ID] = append(tasks[:m.selectedTask], tasks[m.selectedTask+1:]...)

	// Adjust selectedTask if we removed the last task
	if m.selectedTask >= len(m.tasks[currentCol.ID]) && m.selectedTask > 0 {
		m.selectedTask--
	}
}

// calculateViewportSize calculates how many columns can fit in the terminal width
// Column width: 40 (content) + 2 (padding) + 2 (border) = 44 chars
// Spacing between columns: 2 chars
// Total per column: 46 chars
// This method ensures at least 1 column is always visible
func (m *Model) calculateViewportSize() {
	// Delegate to UIState which now owns this logic
	// Keep old field updated for backwards compatibility (will be removed in Phase 4)
	m.uiState.SetWidth(m.width)
	m.viewportSize = m.uiState.ViewportSize()
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// getCurrentColumn returns the currently selected column
// Returns nil if there are no columns
func (m Model) getCurrentColumn() *models.Column {
	if len(m.columns) == 0 {
		return nil
	}
	if m.selectedColumn >= len(m.columns) {
		return nil
	}
	return m.columns[m.selectedColumn]
}

// removeCurrentColumn removes the currently selected column from the model's local state
// This should be called after successfully deleting a column from the database
// It adjusts the selectedColumn index if necessary to keep it within bounds
// It also adjusts the viewportOffset if needed
func (m *Model) removeCurrentColumn() {
	if len(m.columns) == 0 || m.selectedColumn >= len(m.columns) {
		return
	}

	// Remove the column at selectedColumn index
	m.columns = append(m.columns[:m.selectedColumn], m.columns[m.selectedColumn+1:]...)

	// Adjust selectedColumn if we removed the last column
	if m.selectedColumn >= len(m.columns) && m.selectedColumn > 0 {
		m.selectedColumn--
	}

	// Reset task selection
	m.selectedTask = 0

	// Adjust viewportOffset using UIState helper (keeps old field synced)
	m.uiState.AdjustViewportAfterColumnRemoval(m.selectedColumn, len(m.columns))
	m.viewportOffset = m.uiState.ViewportOffset()
}

// moveTaskRight moves the currently selected task to the next column (right)
// Updates both the local state and the database using the linked list structure
// The selection follows the moved task and the viewport scrolls if needed
func (m *Model) moveTaskRight() {
	// Get the current task
	task := m.getCurrentTask()
	if task == nil {
		return
	}

	// Check if there's a next column using the linked list
	currentCol := m.columns[m.selectedColumn]
	if currentCol.NextID == nil {
		// Already at last column
		return
	}

	// Use the new database function to move task
	err := database.MoveTaskToNextColumn(m.db, task.ID)
	if err != nil {
		log.Printf("Error moving task to next column: %v", err)
		return
	}

	// Update local state: remove from current column
	tasks := m.tasks[currentCol.ID]
	m.tasks[currentCol.ID] = append(tasks[:m.selectedTask], tasks[m.selectedTask+1:]...)

	// Find the next column and add task there
	nextColID := *currentCol.NextID
	newPosition := len(m.tasks[nextColID])
	task.ColumnID = nextColID
	task.Position = newPosition
	m.tasks[nextColID] = append(m.tasks[nextColID], task)

	// Move selection to follow the task
	m.selectedColumn++
	m.selectedTask = newPosition

	// Ensure the moved task is visible (auto-scroll viewport if needed)
	if m.selectedColumn >= m.viewportOffset+m.viewportSize {
		m.viewportOffset++
		m.uiState.SetViewportOffset(m.viewportOffset) // Keep UIState synced
	}
}

// moveTaskLeft moves the currently selected task to the previous column (left)
// Updates both the local state and the database using the linked list structure
// The selection follows the moved task and the viewport scrolls if needed
func (m *Model) moveTaskLeft() {
	// Get the current task
	task := m.getCurrentTask()
	if task == nil {
		return
	}

	// Check if there's a previous column using the linked list
	currentCol := m.columns[m.selectedColumn]
	if currentCol.PrevID == nil {
		// Already at first column
		return
	}

	// Use the new database function to move task
	err := database.MoveTaskToPrevColumn(m.db, task.ID)
	if err != nil {
		log.Printf("Error moving task to previous column: %v", err)
		return
	}

	// Update local state: remove from current column
	tasks := m.tasks[currentCol.ID]
	m.tasks[currentCol.ID] = append(tasks[:m.selectedTask], tasks[m.selectedTask+1:]...)

	// Find the previous column and add task there
	prevColID := *currentCol.PrevID
	newPosition := len(m.tasks[prevColID])
	task.ColumnID = prevColID
	task.Position = newPosition
	m.tasks[prevColID] = append(m.tasks[prevColID], task)

	// Move selection to follow the task
	m.selectedColumn--
	m.selectedTask = newPosition

	// Ensure the moved task is visible (auto-scroll viewport if needed)
	if m.selectedColumn < m.viewportOffset {
		m.viewportOffset--
		m.uiState.SetViewportOffset(m.viewportOffset) // Keep UIState synced
	}
}

// getCurrentProject returns the currently selected project
// Returns nil if there are no projects
func (m Model) getCurrentProject() *models.Project {
	// Use new appState (while keeping old fields for backwards compatibility)
	return m.appState.GetCurrentProject()
}

// switchToProject switches to a different project by index and reloads columns/tasks/labels
func (m *Model) switchToProject(projectIndex int) {
	if projectIndex < 0 || projectIndex >= len(m.projects) {
		return
	}

	// Update both old and new state for backwards compatibility
	m.selectedProject = projectIndex
	m.appState.SetSelectedProject(projectIndex)

	project := m.projects[projectIndex]

	// Reload columns for this project
	columns, err := database.GetColumnsByProject(m.db, project.ID)
	if err != nil {
		log.Printf("Error loading columns for project %d: %v", project.ID, err)
		columns = []*models.Column{}
	}
	m.columns = columns
	m.appState.SetColumns(columns)

	// Reload task summaries for each column
	tasks := make(map[int][]*models.TaskSummary)
	for _, col := range columns {
		columnTasks, err := database.GetTaskSummariesByColumn(m.db, col.ID)
		if err != nil {
			log.Printf("Error loading tasks for column %d: %v", col.ID, err)
			columnTasks = []*models.TaskSummary{}
		}
		tasks[col.ID] = columnTasks
	}
	m.tasks = tasks
	m.appState.SetTasks(tasks)

	// Reload labels for this project
	labels, err := database.GetLabelsByProject(m.db, project.ID)
	if err != nil {
		log.Printf("Error loading labels for project %d: %v", project.ID, err)
		labels = []*models.Label{}
	}
	m.labels = labels
	m.appState.SetLabels(labels)

	// Reset selection state
	m.selectedColumn = 0
	m.selectedTask = 0
	m.viewportOffset = 0
}

// reloadProjects reloads the projects list from the database
func (m *Model) reloadProjects() {
	projects, err := database.GetAllProjects(m.db)
	if err != nil {
		log.Printf("Error reloading projects: %v", err)
		return
	}
	// Update both old and new state for backwards compatibility
	m.projects = projects
	m.appState.SetProjects(projects)
}

// reloadLabels reloads the labels for the current project from the database
func (m *Model) reloadLabels() {
	project := m.getCurrentProject()
	if project == nil {
		m.labels = []*models.Label{}
		m.appState.SetLabels([]*models.Label{})
		return
	}

	labels, err := database.GetLabelsByProject(m.db, project.ID)
	if err != nil {
		log.Printf("Error reloading labels: %v", err)
		return
	}
	// Update both old and new state for backwards compatibility
	m.labels = labels
	m.appState.SetLabels(labels)
}

// initLabelPicker initializes the label picker for a task
// Returns false if there's no task to edit
func (m *Model) initLabelPicker(taskID int) bool {
	if taskID == 0 {
		return false
	}

	// Get current labels for the task
	taskLabels, err := database.GetLabelsForTask(m.db, taskID)
	if err != nil {
		log.Printf("Error loading task labels: %v", err)
		taskLabels = []*models.Label{}
	}

	// Build a map of task label IDs for quick lookup
	taskLabelMap := make(map[int]bool)
	for _, label := range taskLabels {
		taskLabelMap[label.ID] = true
	}

	// Build picker items from all project labels
	m.labelPickerItems = make([]LabelPickerItem, len(m.labels))
	for i, label := range m.labels {
		m.labelPickerItems[i] = LabelPickerItem{
			Label:    label,
			Selected: taskLabelMap[label.ID],
		}
	}

	m.labelPickerTaskID = taskID
	m.labelPickerCursor = 0
	m.labelPickerFilter = ""
	m.labelPickerCreateMode = false
	m.labelPickerColorIdx = 0

	return true
}

// getFilteredLabelPickerItems returns label picker items filtered by the current filter text
func (m *Model) getFilteredLabelPickerItems() []LabelPickerItem {
	if m.labelPickerFilter == "" {
		return m.labelPickerItems
	}

	lowerFilter := strings.ToLower(m.labelPickerFilter)
	var filtered []LabelPickerItem
	for _, item := range m.labelPickerItems {
		if strings.Contains(strings.ToLower(item.Label.Name), lowerFilter) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}
