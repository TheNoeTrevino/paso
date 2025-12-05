package tui

import (
	"context"
	"errors"
	"log"

	tea "charm.land/bubbletea/v2"
	"github.com/thenoetrevino/paso/internal/database"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// Model represents the application state for the TUI
type Model struct {
	repo             database.DataStore
	appState         *state.AppState
	uiState          *state.UIState
	inputState       *state.InputState
	formState        *state.FormState
	labelPickerState *state.LabelPickerState
	notificationState *state.NotificationState
}

// InitialModel creates and initializes the TUI model with data from the database
func InitialModel(repo database.DataStore) Model {
	ctx := context.Background()

	// Load all projects
	projects, err := repo.GetAllProjects(ctx)
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
	columns, err := repo.GetColumnsByProject(ctx, currentProjectID)
	if err != nil {
		log.Printf("Error loading columns: %v", err)
		columns = []*models.Column{}
	}

	// Load task summaries for the entire project (includes labels)
	// Uses batch query to avoid N+1 pattern
	tasks, err := repo.GetTaskSummariesByProject(ctx, currentProjectID)
	if err != nil {
		log.Printf("Error loading tasks for project %d: %v", currentProjectID, err)
		tasks = make(map[int][]*models.TaskSummary)
	}

	// Load labels for the current project
	labels, err := repo.GetLabelsByProject(ctx, currentProjectID)
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
	notificationState := state.NewNotificationState()

	return Model{
		repo:             repo,
		appState:         appState,
		uiState:          uiState,
		inputState:       inputState,
		formState:        formState,
		labelPickerState: labelPickerState,
		notificationState: notificationState,
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
	if len(m.appState.Columns()) == 0 {
		return []*models.TaskSummary{}
	}
	if m.uiState.SelectedColumn() >= len(m.appState.Columns()) {
		return []*models.TaskSummary{}
	}
	currentCol := m.appState.Columns()[m.uiState.SelectedColumn()]
	tasks := m.appState.Tasks()[currentCol.ID]
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
	if m.uiState.SelectedTask() >= len(tasks) {
		return nil
	}
	return tasks[m.uiState.SelectedTask()]
}

// removeCurrentTask removes the currently selected task from the model's local state
// This should be called after successfully deleting a task from the database
// It adjusts the selectedTask index if necessary to keep it within bounds
func (m Model) removeCurrentTask() {
	if len(m.appState.Columns()) == 0 {
		return
	}

	currentCol := m.appState.Columns()[m.uiState.SelectedColumn()]
	tasks := m.appState.Tasks()[currentCol.ID]

	if len(tasks) == 0 || m.uiState.SelectedTask() >= len(tasks) {
		return
	}

	// Remove the task at selectedTask index
	m.appState.Tasks()[currentCol.ID] = append(tasks[:m.uiState.SelectedTask()], tasks[m.uiState.SelectedTask()+1:]...)

	// Adjust selectedTask if we removed the last task
	if m.uiState.SelectedTask() >= len(m.appState.Tasks()[currentCol.ID]) && m.uiState.SelectedTask() > 0 {
		m.uiState.SetSelectedTask(m.uiState.SelectedTask() - 1)
	}
}

// getCurrentColumn returns the currently selected column
// Returns nil if there are no columns
func (m Model) getCurrentColumn() *models.Column {
	if len(m.appState.Columns()) == 0 {
		return nil
	}
	if m.uiState.SelectedColumn() >= len(m.appState.Columns()) {
		return nil
	}
	return m.appState.Columns()[m.uiState.SelectedColumn()]
}

// removeCurrentColumn removes the currently selected column from the model's local state
// This should be called after successfully deleting a column from the database
// It adjusts the selectedColumn index if necessary to keep it within bounds
// It also adjusts the viewportOffset if needed
func (m Model) removeCurrentColumn() {
	columns := m.appState.Columns()
	selectedCol := m.uiState.SelectedColumn()

	if len(columns) == 0 || selectedCol >= len(columns) {
		return
	}

	// Remove the column at selectedColumn index
	m.appState.SetColumns(append(columns[:selectedCol], columns[selectedCol+1:]...))

	// Adjust selectedColumn if we removed the last column
	if selectedCol >= len(m.appState.Columns()) && selectedCol > 0 {
		m.uiState.SetSelectedColumn(selectedCol - 1)
	}

	// Reset task selection
	m.uiState.SetSelectedTask(0)

	// Adjust viewportOffset using UIState helper
	m.uiState.AdjustViewportAfterColumnRemoval(m.uiState.SelectedColumn(), len(m.appState.Columns()))
}

// moveTaskRight moves the currently selected task to the next column (right)
// Updates both the local state and the database using the linked list structure
// The selection follows the moved task and the viewport scrolls if needed
func (m Model) moveTaskRight() {
	// Get the current task
	task := m.getCurrentTask()
	if task == nil {
		return
	}

	// Check if there's a next column using the linked list
	currentCol := m.appState.Columns()[m.uiState.SelectedColumn()]
	if currentCol.NextID == nil {
		// Already at last column - show notification
		m.notificationState.Add(state.LevelInfo, "There are no more columns to move to.")
		return
	}

	// Use the new database function to move task
	err := m.repo.MoveTaskToNextColumn(context.Background(), task.ID)
	if err != nil {
		log.Printf("Error moving task to next column: %v", err)
		m.notificationState.Add(state.LevelError, "Failed to move task to next column")
		return
	}

	// Update local state: remove from current column
	tasks := m.appState.Tasks()[currentCol.ID]
	m.appState.Tasks()[currentCol.ID] = append(tasks[:m.uiState.SelectedTask()], tasks[m.uiState.SelectedTask()+1:]...)

	// Find the next column and add task there
	nextColID := *currentCol.NextID
	newPosition := len(m.appState.Tasks()[nextColID])
	task.ColumnID = nextColID
	task.Position = newPosition
	m.appState.Tasks()[nextColID] = append(m.appState.Tasks()[nextColID], task)

	// Move selection to follow the task
	m.uiState.SetSelectedColumn(m.uiState.SelectedColumn() + 1)
	m.uiState.SetSelectedTask(newPosition)

	// Ensure the moved task is visible (auto-scroll viewport if needed)
	if m.uiState.SelectedColumn() >= m.uiState.ViewportOffset()+m.uiState.ViewportSize() {
		m.uiState.SetViewportOffset(m.uiState.ViewportOffset() + 1)
	}
}

// moveTaskLeft moves the currently selected task to the previous column (left)
// Updates both the local state and the database using the linked list structure
// The selection follows the moved task and the viewport scrolls if needed
func (m Model) moveTaskLeft() {
	// Get the current task
	task := m.getCurrentTask()
	if task == nil {
		return
	}

	// Check if there's a previous column using the linked list
	currentCol := m.appState.Columns()[m.uiState.SelectedColumn()]
	if currentCol.PrevID == nil {
		// Already at first column - show notification
		m.notificationState.Add(state.LevelInfo, "There are no more columns to move to.")
		return
	}

	// Use the new database function to move task
	err := m.repo.MoveTaskToPrevColumn(context.Background(), task.ID)
	if err != nil {
		log.Printf("Error moving task to previous column: %v", err)
		m.notificationState.Add(state.LevelError, "Failed to move task to previous column")
		return
	}

	// Update local state: remove from current column
	tasks := m.appState.Tasks()[currentCol.ID]
	m.appState.Tasks()[currentCol.ID] = append(tasks[:m.uiState.SelectedTask()], tasks[m.uiState.SelectedTask()+1:]...)

	// Find the previous column and add task there
	prevColID := *currentCol.PrevID
	newPosition := len(m.appState.Tasks()[prevColID])
	task.ColumnID = prevColID
	task.Position = newPosition
	m.appState.Tasks()[prevColID] = append(m.appState.Tasks()[prevColID], task)

	// Move selection to follow the task
	m.uiState.SetSelectedColumn(m.uiState.SelectedColumn() - 1)
	m.uiState.SetSelectedTask(newPosition)

	// Ensure the moved task is visible (auto-scroll viewport if needed)
	if m.uiState.SelectedColumn() < m.uiState.ViewportOffset() {
		m.uiState.SetViewportOffset(m.uiState.ViewportOffset() - 1)
	}
}

// moveTaskUp moves the currently selected task up within its column
// Updates both the local state and the database
// The selection follows the moved task
func (m Model) moveTaskUp() {
	task := m.getCurrentTask()
	if task == nil {
		return
	}

	// Check if already at top (edge case handled here for quick feedback)
	if m.uiState.SelectedTask() == 0 {
		m.notificationState.Add(state.LevelInfo, "Task is already at the top")
		return
	}

	// Call database swap
	err := m.repo.SwapTaskUp(context.Background(), task.ID)
	if err != nil {
		log.Printf("Error moving task up: %v", err)
		if err != errors.New("task is already at the top of the column") {
			m.notificationState.Add(state.LevelError, "Failed to move task up")
		}
		return
	}

	// Update local state: swap tasks in slice
	currentCol := m.appState.Columns()[m.uiState.SelectedColumn()]
	tasks := m.appState.Tasks()[currentCol.ID]
	selectedIdx := m.uiState.SelectedTask()

	// Swap positions in slice
	tasks[selectedIdx], tasks[selectedIdx-1] = tasks[selectedIdx-1], tasks[selectedIdx]

	// Update position values on the task objects
	tasks[selectedIdx].Position = selectedIdx
	tasks[selectedIdx-1].Position = selectedIdx - 1

	// Move selection to follow the task
	m.uiState.SetSelectedTask(selectedIdx - 1)
}

// moveTaskDown moves the currently selected task down within its column
// Updates both the local state and the database
// The selection follows the moved task
func (m Model) moveTaskDown() {
	task := m.getCurrentTask()
	if task == nil {
		return
	}

	// Get current tasks for edge case check
	currentCol := m.appState.Columns()[m.uiState.SelectedColumn()]
	tasks := m.appState.Tasks()[currentCol.ID]
	selectedIdx := m.uiState.SelectedTask()

	// Check if already at bottom
	if selectedIdx >= len(tasks)-1 {
		m.notificationState.Add(state.LevelInfo, "Task is already at the bottom")
		return
	}

	// Call database swap
	err := m.repo.SwapTaskDown(context.Background(), task.ID)
	if err != nil {
		log.Printf("Error moving task down: %v", err)
		if err != models.ErrAlreadyLastTask {
			m.notificationState.Add(state.LevelError, "Failed to move task down")
		}
		return
	}

	// Update local state: swap tasks in slice
	tasks[selectedIdx], tasks[selectedIdx+1] = tasks[selectedIdx+1], tasks[selectedIdx]

	// Update position values on the task objects
	tasks[selectedIdx].Position = selectedIdx
	tasks[selectedIdx+1].Position = selectedIdx + 1

	// Move selection to follow the task
	m.uiState.SetSelectedTask(selectedIdx + 1)
}

// getCurrentProject returns the currently selected project
// Returns nil if there are no projects
func (m Model) getCurrentProject() *models.Project {
	return m.appState.GetCurrentProject()
}

// switchToProject switches to a different project by index and reloads columns/tasks/labels
func (m Model) switchToProject(projectIndex int) {
	if projectIndex < 0 || projectIndex >= len(m.appState.Projects()) {
		return
	}

	// Update state
	m.appState.SetSelectedProject(projectIndex)

	project := m.appState.Projects()[projectIndex]

	// Reload columns for this project
	columns, err := m.repo.GetColumnsByProject(context.Background(), project.ID)
	if err != nil {
		log.Printf("Error loading columns for project %d: %v", project.ID, err)
		columns = []*models.Column{}
	}
	m.appState.SetColumns(columns)

	// Reload task summaries for the entire project
	tasks, err := m.repo.GetTaskSummariesByProject(context.Background(), project.ID)
	if err != nil {
		log.Printf("Error loading tasks for project %d: %v", project.ID, err)
		tasks = make(map[int][]*models.TaskSummary)
	}
	m.appState.SetTasks(tasks)

	// Reload labels for this project
	labels, err := m.repo.GetLabelsByProject(context.Background(), project.ID)
	if err != nil {
		log.Printf("Error loading labels for project %d: %v", project.ID, err)
		labels = []*models.Label{}
	}
	m.appState.SetLabels(labels)

	// Reset selection state
	m.uiState.ResetSelection()
}

// reloadProjects reloads the projects list from the database
func (m Model) reloadProjects() {
	projects, err := m.repo.GetAllProjects(context.Background())
	if err != nil {
		log.Printf("Error reloading projects: %v", err)
		return
	}
	m.appState.SetProjects(projects)
}

// reloadLabels reloads the labels for the current project from the database
func (m Model) reloadLabels() {
	project := m.getCurrentProject()
	if project == nil {
		m.appState.SetLabels([]*models.Label{})
		return
	}

	labels, err := m.repo.GetLabelsByProject(context.Background(), project.ID)
	if err != nil {
		log.Printf("Error reloading labels: %v", err)
		return
	}
	m.appState.SetLabels(labels)
}

// initLabelPicker initializes the label picker for a task
// Returns false if there's no task to edit
func (m Model) initLabelPicker(taskID int) bool {
	if taskID == 0 {
		return false
	}

	// Get current labels for the task
	taskLabels, err := m.repo.GetLabelsForTask(context.Background(), taskID)
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
	items := make([]state.LabelPickerItem, len(m.appState.Labels()))
	for i, label := range m.appState.Labels() {
		items[i] = state.LabelPickerItem{
			Label:    label,
			Selected: taskLabelMap[label.ID],
		}
	}

	// Initialize LabelPickerState
	m.labelPickerState.Items = items
	m.labelPickerState.TaskID = taskID
	m.labelPickerState.Cursor = 0
	m.labelPickerState.Filter = ""
	m.labelPickerState.CreateMode = false
	m.labelPickerState.ColorIdx = 0

	return true
}

// getFilteredLabelPickerItems returns label picker items filtered by the current filter text
func (m *Model) getFilteredLabelPickerItems() []state.LabelPickerItem {
	// Delegate to LabelPickerState which now owns this logic
	return m.labelPickerState.GetFilteredItems()
}
