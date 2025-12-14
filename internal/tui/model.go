package tui

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/thenoetrevino/paso/internal/config"
	"github.com/thenoetrevino/paso/internal/database"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// Timeout constants for context operations
const (
	timeoutInitialLoad = 30 * time.Second
	timeoutUI          = 10 * time.Second
	timeoutDB          = 30 * time.Second
)

// Model represents the application state for the TUI
type Model struct {
	ctx               context.Context // Application context for cancellation and timeouts
	repo              database.DataStore
	config             *config.Config
	appState           *state.AppState
	uiState            *state.UIState
	inputState         *state.InputState
	formState          *state.FormState
	labelPickerState   *state.LabelPickerState
	parentPickerState  *state.TaskPickerState
	childPickerState   *state.TaskPickerState
	priorityPickerState *state.PriorityPickerState
	notificationState  *state.NotificationState
	searchState        *state.SearchState
	listViewState      *state.ListViewState
	statusPickerState  *state.StatusPickerState
}

// InitialModel creates and initializes the TUI model with data from the database
func InitialModel(ctx context.Context, repo database.DataStore, cfg *config.Config) Model {
	// Create child context with timeout for initial loading
	loadCtx, cancel := context.WithTimeout(ctx, timeoutInitialLoad)
	defer cancel()

	// Load all projects
	projects, err := repo.GetAllProjects(loadCtx)
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
	columns, err := repo.GetColumnsByProject(loadCtx, currentProjectID)
	if err != nil {
		log.Printf("Error loading columns: %v", err)
		columns = []*models.Column{}
	}

	// Load task summaries for the entire project (includes labels)
	// Uses batch query to avoid N+1 pattern
	tasks, err := repo.GetTaskSummariesByProject(loadCtx, currentProjectID)
	if err != nil {
		log.Printf("Error loading tasks for project %d: %v", currentProjectID, err)
		tasks = make(map[int][]*models.TaskSummary)
	}

	// Load labels for the current project
	labels, err := repo.GetLabelsByProject(loadCtx, currentProjectID)
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
	parentPickerState := state.NewTaskPickerState()
	childPickerState := state.NewTaskPickerState()
	priorityPickerState := state.NewPriorityPickerState()
	notificationState := state.NewNotificationState()
	searchState := state.NewSearchState()
	listViewState := state.NewListViewState()
	statusPickerState := state.NewStatusPickerState()

	// Initialize styles with color scheme from config
	InitStyles(cfg.ColorScheme)

	return Model{
		ctx:                 ctx, // Store root context
		repo:                repo,
		config:              cfg,
		appState:            appState,
		uiState:             uiState,
		inputState:          inputState,
		formState:           formState,
		labelPickerState:    labelPickerState,
		parentPickerState:   parentPickerState,
		childPickerState:    childPickerState,
		priorityPickerState: priorityPickerState,
		notificationState:   notificationState,
		searchState:         searchState,
		listViewState:     listViewState,
		statusPickerState: statusPickerState,
	}
}

// withTimeout creates a child context with appropriate timeout for operation type
func (m *Model) withTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(m.ctx, timeout)
}

// dbContext creates a context for database operations with 30s timeout
func (m *Model) dbContext() (context.Context, context.CancelFunc) {
	return m.withTimeout(timeoutDB)
}

// uiContext creates a context for UI operations with 10s timeout
func (m *Model) uiContext() (context.Context, context.CancelFunc) {
	return m.withTimeout(timeoutUI)
}

// handleDBError handles database errors with context-aware messages
// It distinguishes between cancellation, timeout, and other errors,
// providing appropriate user feedback for each case.
func (m *Model) handleDBError(err error, operation string) {
	if err == nil {
		return
	}

	if errors.Is(err, context.Canceled) {
		// Operation was cancelled - this is expected during shutdown
		// Don't show error notification, just log for debugging
		log.Printf("%s cancelled by user", operation)
		return
	}

	if errors.Is(err, context.DeadlineExceeded) {
		// Operation timed out - show user-friendly message
		m.notificationState.Add(state.LevelError, fmt.Sprintf("%s timed out. Please try again.", operation))
		return
	}

	// Other errors - show detailed error message
	m.notificationState.Add(state.LevelError, fmt.Sprintf("%s failed: %v", operation, err))
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
	currentCol := m.getCurrentColumn()
	if currentCol == nil {
		return
	}

	tasks := m.getTasksForColumn(currentCol.ID)

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
	selectedIdx := m.uiState.SelectedColumn()
	if selectedIdx < 0 || selectedIdx >= len(m.appState.Columns()) {
		return nil
	}
	return m.appState.Columns()[selectedIdx]
}

// getTasksForColumn returns tasks for a specific column ID with safe map access.
// Returns an empty slice if the column ID doesn't exist in the tasks map.
func (m Model) getTasksForColumn(columnID int) []*models.TaskSummary {
	tasks, ok := m.appState.Tasks()[columnID]
	if !ok || tasks == nil {
		return []*models.TaskSummary{}
	}
	return tasks
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
	currentCol := m.getCurrentColumn()
	if currentCol == nil {
		return
	}
	if currentCol.NextID == nil {
		// Already at last column - show notification
		m.notificationState.Add(state.LevelInfo, "There are no more columns to move to.")
		return
	}

	// Use the new database function to move task
	ctx, cancel := m.uiContext()
	defer cancel()
	err := m.repo.MoveTaskToNextColumn(ctx, task.ID)
	if err != nil {
		log.Printf("Error moving task to next column: %v", err)
		if err != models.ErrAlreadyLastColumn {
			m.notificationState.Add(state.LevelError, "Failed to move task to next column")
		}
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
	currentCol := m.getCurrentColumn()
	if currentCol == nil {
		return
	}
	if currentCol.PrevID == nil {
		// Already at first column - show notification
		m.notificationState.Add(state.LevelInfo, "There are no more columns to move to.")
		return
	}

	// Use the new database function to move task
	ctx, cancel := m.uiContext()
	defer cancel()
	err := m.repo.MoveTaskToPrevColumn(ctx, task.ID)
	if err != nil {
		log.Printf("Error moving task to previous column: %v", err)
		if err != models.ErrAlreadyFirstColumn {
			m.notificationState.Add(state.LevelError, "Failed to move task to previous column")
		}
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
	ctx, cancel := m.uiContext()
	defer cancel()
	err := m.repo.SwapTaskUp(ctx, task.ID)
	if err != nil {
		log.Printf("Error moving task up: %v", err)
		if err != models.ErrAlreadyFirstTask {
			m.notificationState.Add(state.LevelError, "Failed to move task up")
		}
		return
	}

	// Update local state: swap tasks in slice
	currentCol := m.getCurrentColumn()
	if currentCol == nil {
		return
	}

	tasks := m.getTasksForColumn(currentCol.ID)
	if len(tasks) < 2 {
		return
	}

	selectedIdx := m.uiState.SelectedTask()
	if selectedIdx == 0 || selectedIdx >= len(tasks) {
		return
	}

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
	currentCol := m.getCurrentColumn()
	if currentCol == nil {
		return
	}

	tasks := m.getTasksForColumn(currentCol.ID)
	selectedIdx := m.uiState.SelectedTask()

	// Check if already at bottom
	if selectedIdx >= len(tasks)-1 {
		m.notificationState.Add(state.LevelInfo, "Task is already at the bottom")
		return
	}

	// Call database swap
	ctx, cancel := m.uiContext()
	defer cancel()
	err := m.repo.SwapTaskDown(ctx, task.ID)
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

	// Create context for database operations
	ctx, cancel := m.dbContext()
	defer cancel()

	// Reload columns for this project
	columns, err := m.repo.GetColumnsByProject(ctx, project.ID)
	if err != nil {
		log.Printf("Error loading columns for project %d: %v", project.ID, err)
		columns = []*models.Column{}
	}
	m.appState.SetColumns(columns)

	// Reload task summaries for the entire project
	tasks, err := m.repo.GetTaskSummariesByProject(ctx, project.ID)
	if err != nil {
		log.Printf("Error loading tasks for project %d: %v", project.ID, err)
		tasks = make(map[int][]*models.TaskSummary)
	}
	m.appState.SetTasks(tasks)

	// Reload labels for this project
	labels, err := m.repo.GetLabelsByProject(ctx, project.ID)
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
	ctx, cancel := m.dbContext()
	defer cancel()
	projects, err := m.repo.GetAllProjects(ctx)
	if err != nil {
		log.Printf("Error reloading projects: %v", err)
		return
	}
	m.appState.SetProjects(projects)
}

// initParentPickerForForm initializes the parent picker for use in TicketFormMode.
// In edit mode: loads existing parent relationships from FormState.
// In create mode: starts with empty selection (relationships applied after task creation).
//
// Returns false if there's no current project.
func (m *Model) initParentPickerForForm() bool {
	project := m.getCurrentProject()
	if project == nil {
		return false
	}

	// Get all task references for the entire project
	ctx, cancel := m.dbContext()
	defer cancel()
	allTasks, err := m.repo.GetTaskReferencesForProject(ctx, project.ID)
	if err != nil {
		log.Printf("Error loading project tasks: %v", err)
		return false
	}

	// Build map of currently selected parent task IDs from form state
	parentTaskMap := make(map[int]bool)
	for _, parentID := range m.formState.FormParentIDs {
		parentTaskMap[parentID] = true
	}

	// Build picker items from all tasks, excluding current task (if editing)
	items := make([]state.TaskPickerItem, 0, len(allTasks))
	for _, task := range allTasks {
		// In edit mode, exclude the task being edited
		if m.formState.EditingTaskID != 0 && task.ID == m.formState.EditingTaskID {
			continue
		}
		items = append(items, state.TaskPickerItem{
			TaskRef:  task,
			Selected: parentTaskMap[task.ID],
		})
	}

	// Initialize ParentPickerState
	m.parentPickerState.Items = items
	m.parentPickerState.TaskID = m.formState.EditingTaskID // 0 for create mode
	m.parentPickerState.Cursor = 0
	m.parentPickerState.Filter = ""
	m.parentPickerState.PickerType = "parent"
	m.parentPickerState.ReturnMode = state.TicketFormMode

	return true
}

// initChildPickerForForm initializes the child picker for use in TicketFormMode.
// In edit mode: loads existing child relationships from FormState.
// In create mode: starts with empty selection (relationships applied after task creation).
//
// Returns false if there's no current project.
func (m *Model) initChildPickerForForm() bool {
	project := m.getCurrentProject()
	if project == nil {
		return false
	}

	// Get all task references for the entire project
	ctx, cancel := m.dbContext()
	defer cancel()
	allTasks, err := m.repo.GetTaskReferencesForProject(ctx, project.ID)
	if err != nil {
		log.Printf("Error loading project tasks: %v", err)
		return false
	}

	// Build map of currently selected child task IDs from form state
	childTaskMap := make(map[int]bool)
	for _, childID := range m.formState.FormChildIDs {
		childTaskMap[childID] = true
	}

	// Build picker items from all tasks, excluding current task (if editing)
	items := make([]state.TaskPickerItem, 0, len(allTasks))
	for _, task := range allTasks {
		// In edit mode, exclude the task being edited
		if m.formState.EditingTaskID != 0 && task.ID == m.formState.EditingTaskID {
			continue
		}
		items = append(items, state.TaskPickerItem{
			TaskRef:  task,
			Selected: childTaskMap[task.ID],
		})
	}

	// Initialize ChildPickerState
	m.childPickerState.Items = items
	m.childPickerState.TaskID = m.formState.EditingTaskID // 0 for create mode
	m.childPickerState.Cursor = 0
	m.childPickerState.Filter = ""
	m.childPickerState.PickerType = "child"
	m.childPickerState.ReturnMode = state.TicketFormMode

	return true
}

// initLabelPickerForForm initializes the label picker for use in TicketFormMode.
// In edit mode: loads existing label selections from FormState.
// In create mode: starts with empty selection (labels applied on form submission).
//
// Returns false if there's no current project.
func (m *Model) initLabelPickerForForm() bool {
	project := m.getCurrentProject()
	if project == nil {
		return false
	}

	// Build map of currently selected label IDs from form state
	labelIDMap := make(map[int]bool)
	for _, labelID := range m.formState.FormLabelIDs {
		labelIDMap[labelID] = true
	}

	// Build picker items from all available labels
	var items []state.LabelPickerItem
	for _, label := range m.appState.Labels() {
		items = append(items, state.LabelPickerItem{
			Label:    label,
			Selected: labelIDMap[label.ID],
		})
	}

	// Initialize LabelPickerState
	m.labelPickerState.Items = items
	m.labelPickerState.TaskID = m.formState.EditingTaskID // 0 for create mode
	m.labelPickerState.Cursor = 0
	m.labelPickerState.Filter = ""
	m.labelPickerState.ReturnMode = state.TicketFormMode

	return true
}

// getFilteredLabelPickerItems returns label picker items filtered by the current filter text
func (m *Model) getFilteredLabelPickerItems() []state.LabelPickerItem {
	// Delegate to LabelPickerState which now owns this logic
	return m.labelPickerState.GetFilteredItems()
}

// initPriorityPickerForForm initializes the priority picker for use in TicketFormMode.
// Loads the current priority from the form state.
func (m *Model) initPriorityPickerForForm() bool {
	// Get current priority ID from form state
	// If we're editing a task, get it from the task detail
	// Otherwise, default to medium (id=3)
	currentPriorityID := 3 // Default to medium

	// If editing an existing task, we need to get the current priority from database
	if m.formState.EditingTaskID != 0 {
		ctx, cancel := m.dbContext()
		defer cancel()

		taskDetail, err := m.repo.GetTaskDetail(ctx, m.formState.EditingTaskID)
		if err != nil {
			log.Printf("Error loading task detail for priority picker: %v", err)
			return false
		}

		// Find the priority ID from the priority description
		// We need to match it against our priority options
		priorities := GetPriorityOptions()
		for _, p := range priorities {
			if p.Description == taskDetail.PriorityDescription {
				currentPriorityID = p.ID
				break
			}
		}
	}

	// Initialize PriorityPickerState
	m.priorityPickerState.SetSelectedPriorityID(currentPriorityID)
	// Set cursor to match the selected priority (adjust for 0-indexing)
	m.priorityPickerState.SetCursor(currentPriorityID - 1)
	m.priorityPickerState.SetReturnMode(state.TicketFormMode)

	return true
}

// buildListViewRows creates a flat list of all tasks with their column names.
// The list is sorted according to the current sort settings in listViewState.
func (m Model) buildListViewRows() []ListViewRow {
	var rows []ListViewRow
	for _, col := range m.appState.Columns() {
		tasks := m.appState.Tasks()[col.ID]
		for _, task := range tasks {
			rows = append(rows, ListViewRow{
				Task:       task,
				ColumnName: col.Name,
				ColumnID:   col.ID,
			})
		}
	}

	// Apply sorting
	m.sortListViewRows(rows)
	return rows
}

// sortListViewRows sorts the rows based on current sort settings.
func (m Model) sortListViewRows(rows []ListViewRow) {
	if m.listViewState.SortField() == state.SortNone {
		return
	}

	sort.Slice(rows, func(i, j int) bool {
		var cmp int
		switch m.listViewState.SortField() {
		case state.SortByTitle:
			cmp = strings.Compare(rows[i].Task.Title, rows[j].Task.Title)
		case state.SortByStatus:
			cmp = strings.Compare(rows[i].ColumnName, rows[j].ColumnName)
		default:
			return false
		}

		if m.listViewState.SortOrder() == state.SortDesc {
			cmp = -cmp
		}
		return cmp < 0
	})
}

// syncKanbanToListSelection maps the current kanban selection to a list row index.
// This should be called when switching from kanban to list view.
func (m *Model) syncKanbanToListSelection() {
	rows := m.buildListViewRows()
	if len(rows) == 0 {
		m.listViewState.SetSelectedRow(0)
		return
	}

	// Find the task that matches the current kanban selection
	currentTask := m.getCurrentTask()
	if currentTask == nil {
		m.listViewState.SetSelectedRow(0)
		return
	}

	for i, row := range rows {
		if row.Task.ID == currentTask.ID {
			m.listViewState.SetSelectedRow(i)
			return
		}
	}
	m.listViewState.SetSelectedRow(0)
}

// syncListToKanbanSelection maps the current list row to kanban column/task selection.
// This should be called when switching from list to kanban view.
func (m *Model) syncListToKanbanSelection() {
	rows := m.buildListViewRows()
	if len(rows) == 0 {
		return
	}

	selectedRow := m.listViewState.SelectedRow()
	if selectedRow >= len(rows) {
		selectedRow = len(rows) - 1
	}
	if selectedRow < 0 {
		return
	}

	selectedTask := rows[selectedRow].Task

	// Find the column and task position in kanban view
	for colIdx, col := range m.appState.Columns() {
		tasks := m.appState.Tasks()[col.ID]
		for taskIdx, task := range tasks {
			if task.ID == selectedTask.ID {
				m.uiState.SetSelectedColumn(colIdx)
				m.uiState.SetSelectedTask(taskIdx)
				m.uiState.EnsureSelectionVisible(colIdx)
				return
			}
		}
	}
}

// getTaskFromListRow returns the task at the given list row index.
// Returns nil if the index is out of bounds or no tasks exist.
func (m Model) getTaskFromListRow(rowIdx int) *models.TaskSummary {
	rows := m.buildListViewRows()
	if rowIdx < 0 || rowIdx >= len(rows) {
		return nil
	}
	return rows[rowIdx].Task
}

// getSelectedListTask returns the currently selected task in list view.
// This is a convenience method that uses getTaskFromListRow with the current selection.
func (m Model) getSelectedListTask() *models.TaskSummary {
	return m.getTaskFromListRow(m.listViewState.SelectedRow())
}
