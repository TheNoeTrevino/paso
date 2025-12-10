package tui

import (
	"context"
	"log"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui/huhforms"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// ============================================================================
// NORMAL MODE HANDLERS
// ============================================================================

// handleNormalMode dispatches key events in NormalMode to specific handlers.
func (m Model) handleNormalMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.notificationState.Clear()

	key := msg.String()
	km := m.config.KeyMappings

	switch key {
	case km.Quit, "ctrl+c":
		return m.handleQuit()
	case km.ShowHelp:
		return m.handleShowHelp()
	case km.AddTask:
		return m.handleAddTask()
	case km.EditTask:
		return m.handleEditTask()
	case km.DeleteTask:
		return m.handleDeleteTask()
	case km.ViewTask:
		return m.handleEditTask()
	case km.CreateColumn:
		return m.handleCreateColumn()
	case km.RenameColumn:
		return m.handleRenameColumn()
	case km.DeleteColumn:
		return m.handleDeleteColumn()
	case km.ScrollViewportRight:
		return m.handleScrollRight()
	case km.ScrollViewportLeft:
		return m.handleScrollLeft()
	case km.PrevColumn, "left":
		return m.handleNavigateLeft()
	case km.NextColumn, "right":
		return m.handleNavigateRight()
	case km.NextTask, "down":
		return m.handleNavigateDown()
	case km.PrevTask, "up":
		return m.handleNavigateUp()
	case km.MoveTaskRight:
		return m.handleMoveTaskRight()
	case km.MoveTaskLeft:
		return m.handleMoveTaskLeft()
	case km.MoveTaskUp:
		return m.handleMoveTaskUp()
	case km.MoveTaskDown:
		return m.handleMoveTaskDown()
	case km.PrevProject:
		return m.handlePrevProject()
	case km.NextProject:
		return m.handleNextProject()
	case km.CreateProject:
		return m.handleCreateProject()
	case "/":
		return m.handleEnterSearch()
	}

	return m, nil
}

// handleQuit exits the application.
func (m Model) handleQuit() (tea.Model, tea.Cmd) {
	return m, tea.Quit
}

// handleShowHelp shows the help screen.
func (m Model) handleShowHelp() (tea.Model, tea.Cmd) {
	m.uiState.SetMode(state.HelpMode)
	return m, nil
}

// handleNavigateLeft moves selection to the previous column.
func (m Model) handleNavigateLeft() (tea.Model, tea.Cmd) {
	if m.uiState.SelectedColumn() > 0 {
		m.uiState.SetSelectedColumn(m.uiState.SelectedColumn() - 1)
		m.uiState.SetSelectedTask(0)
		m.uiState.EnsureSelectionVisible(m.uiState.SelectedColumn())
	} else {
		m.notificationState.Add(state.LevelInfo, "Already at the first column")
	}
	return m, nil
}

// handleNavigateRight moves selection to the next column.
func (m Model) handleNavigateRight() (tea.Model, tea.Cmd) {
	if m.uiState.SelectedColumn() < len(m.appState.Columns())-1 {
		m.uiState.SetSelectedColumn(m.uiState.SelectedColumn() + 1)
		m.uiState.SetSelectedTask(0)
		m.uiState.EnsureSelectionVisible(m.uiState.SelectedColumn())
	} else {
		m.notificationState.Add(state.LevelInfo, "Already at the last column")
	}
	return m, nil
}

// handleNavigateUp moves selection to the previous task.
func (m Model) handleNavigateUp() (tea.Model, tea.Cmd) {
	if m.uiState.SelectedTask() > 0 {
		m.uiState.SetSelectedTask(m.uiState.SelectedTask() - 1)
	} else {
		m.notificationState.Add(state.LevelInfo, "Already at the first task")
	}
	return m, nil
}

// handleNavigateDown moves selection to the next task.
func (m Model) handleNavigateDown() (tea.Model, tea.Cmd) {
	currentTasks := m.getCurrentTasks()
	if len(currentTasks) > 0 && m.uiState.SelectedTask() < len(currentTasks)-1 {
		m.uiState.SetSelectedTask(m.uiState.SelectedTask() + 1)
	} else if len(currentTasks) > 0 {
		m.notificationState.Add(state.LevelInfo, "Already at the last task")
	}
	return m, nil
}

// handleScrollRight scrolls the viewport right.
func (m Model) handleScrollRight() (tea.Model, tea.Cmd) {
	if m.uiState.ViewportOffset()+m.uiState.ViewportSize() < len(m.appState.Columns()) {
		m.uiState.SetViewportOffset(m.uiState.ViewportOffset() + 1)
		if m.uiState.SelectedColumn() < m.uiState.ViewportOffset() {
			m.uiState.SetSelectedColumn(m.uiState.ViewportOffset())
			m.uiState.SetSelectedTask(0)
		}
	} else {
		m.notificationState.Add(state.LevelInfo, "Already at the rightmost view")
	}
	return m, nil
}

// handleScrollLeft scrolls the viewport left.
func (m Model) handleScrollLeft() (tea.Model, tea.Cmd) {
	if m.uiState.ViewportOffset() > 0 {
		m.uiState.SetViewportOffset(m.uiState.ViewportOffset() - 1)
		if m.uiState.SelectedColumn() >= m.uiState.ViewportOffset()+m.uiState.ViewportSize() {
			m.uiState.SetSelectedColumn(m.uiState.ViewportOffset() + m.uiState.ViewportSize() - 1)
			m.uiState.SetSelectedTask(0)
		}
	} else {
		m.notificationState.Add(state.LevelInfo, "Already at the leftmost view")
	}
	return m, nil
}

// handleAddTask initiates adding a new task.
func (m Model) handleAddTask() (tea.Model, tea.Cmd) {
	if len(m.appState.Columns()) == 0 {
		m.notificationState.Add(state.LevelError, "Cannot add task: No columns exist. Create a column first with 'C'")
		return m, nil
	}
	m.formState.FormTitle = ""
	m.formState.FormDescription = ""
	m.formState.FormLabelIDs = []int{}
	m.formState.FormParentIDs = []int{}
	m.formState.FormChildIDs = []int{}
	m.formState.FormParentRefs = []*models.TaskReference{}
	m.formState.FormChildRefs = []*models.TaskReference{}
	m.formState.FormConfirm = true
	m.formState.EditingTaskID = 0

	// Calculate description height (will be dynamic in Phase 4)
	descriptionLines := 10

	m.formState.TicketForm = huhforms.CreateTicketForm(
		&m.formState.FormTitle,
		&m.formState.FormDescription,
		&m.formState.FormConfirm,
		descriptionLines,
	).WithTheme(huhforms.CreatePasoTheme(m.config.ColorScheme))
	m.formState.SnapshotTicketFormInitialValues() // Snapshot for change detection
	m.uiState.SetMode(state.TicketFormMode)
	return m, m.formState.TicketForm.Init()
}

// handleEditTask initiates editing the selected task.
func (m Model) handleEditTask() (tea.Model, tea.Cmd) {
	task := m.getCurrentTask()
	if task == nil {
		m.notificationState.Add(state.LevelError, "No task selected to edit")
		return m, nil
	}

	taskDetail, err := m.repo.GetTaskDetail(context.Background(), task.ID)
	if err != nil {
		log.Printf("Error loading task details: %v", err)
		m.notificationState.Add(state.LevelError, "Error loading task details")
		return m, nil
	}

	m.formState.FormTitle = taskDetail.Title
	m.formState.FormDescription = taskDetail.Description
	m.formState.FormLabelIDs = make([]int, len(taskDetail.Labels))
	for i, label := range taskDetail.Labels {
		m.formState.FormLabelIDs[i] = label.ID
	}

	// Load parent relationships
	m.formState.FormParentIDs = make([]int, len(taskDetail.ParentTasks))
	m.formState.FormParentRefs = taskDetail.ParentTasks
	for i, parent := range taskDetail.ParentTasks {
		m.formState.FormParentIDs[i] = parent.ID
	}

	// Load child relationships
	m.formState.FormChildIDs = make([]int, len(taskDetail.ChildTasks))
	m.formState.FormChildRefs = taskDetail.ChildTasks
	for i, child := range taskDetail.ChildTasks {
		m.formState.FormChildIDs[i] = child.ID
	}

	// Load timestamps for metadata display
	m.formState.FormCreatedAt = taskDetail.CreatedAt
	m.formState.FormUpdatedAt = taskDetail.UpdatedAt

	m.formState.FormConfirm = true
	m.formState.EditingTaskID = task.ID

	// Calculate description height (will be dynamic in Phase 4)
	descriptionLines := 10

	m.formState.TicketForm = huhforms.CreateTicketForm(
		&m.formState.FormTitle,
		&m.formState.FormDescription,
		&m.formState.FormConfirm,
		descriptionLines,
	).WithTheme(huhforms.CreatePasoTheme(m.config.ColorScheme))
	m.formState.SnapshotTicketFormInitialValues() // Snapshot for change detection
	m.uiState.SetMode(state.TicketFormMode)
	return m, m.formState.TicketForm.Init()
}

// handleDeleteTask initiates task deletion confirmation.
func (m Model) handleDeleteTask() (tea.Model, tea.Cmd) {
	if m.getCurrentTask() == nil {
		m.notificationState.Add(state.LevelError, "No task selected to delete")
		return m, nil
	}
	m.uiState.SetMode(state.DeleteConfirmMode)
	return m, nil
}

// handleMoveTaskRight moves the task to the next column.
func (m Model) handleMoveTaskRight() (tea.Model, tea.Cmd) {
	if m.getCurrentTask() != nil {
		m.moveTaskRight()
	}
	return m, nil
}

// handleMoveTaskLeft moves the task to the previous column.
func (m Model) handleMoveTaskLeft() (tea.Model, tea.Cmd) {
	if m.getCurrentTask() != nil {
		m.moveTaskLeft()
	}
	return m, nil
}

// handleMoveTaskUp moves the task up within its column.
func (m Model) handleMoveTaskUp() (tea.Model, tea.Cmd) {
	if m.getCurrentTask() != nil {
		m.moveTaskUp()
	}
	return m, nil
}

// handleMoveTaskDown moves the task down within its column.
func (m Model) handleMoveTaskDown() (tea.Model, tea.Cmd) {
	if m.getCurrentTask() != nil {
		m.moveTaskDown()
	}
	return m, nil
}

// handleCreateColumn initiates column creation.
func (m Model) handleCreateColumn() (tea.Model, tea.Cmd) {
	m.uiState.SetMode(state.AddColumnMode)
	m.inputState.Prompt = "New column name:"
	m.inputState.Buffer = ""
	return m, nil
}

// handleRenameColumn initiates column renaming.
func (m Model) handleRenameColumn() (tea.Model, tea.Cmd) {
	column := m.getCurrentColumn()
	if column == nil {
		m.notificationState.Add(state.LevelError, "No column selected to rename")
		return m, nil
	}
	m.uiState.SetMode(state.EditColumnMode)
	m.inputState.Buffer = column.Name
	m.inputState.Prompt = "Rename column:"
	m.inputState.SnapshotInitialBuffer() // Snapshot for change detection
	return m, nil
}

// handleDeleteColumn initiates column deletion confirmation.
func (m Model) handleDeleteColumn() (tea.Model, tea.Cmd) {
	column := m.getCurrentColumn()
	if column == nil {
		m.notificationState.Add(state.LevelError, "No column selected to delete")
		return m, nil
	}
	taskCount, err := m.repo.GetTaskCountByColumn(context.Background(), column.ID)
	if err != nil {
		log.Printf("Error getting task count: %v", err)
		m.notificationState.Add(state.LevelError, "Error getting column info")
		return m, nil
	}
	m.inputState.DeleteColumnTaskCount = taskCount
	m.uiState.SetMode(state.DeleteColumnConfirmMode)
	return m, nil
}

// handlePrevProject switches to the previous project.
func (m Model) handlePrevProject() (tea.Model, tea.Cmd) {
	if m.appState.SelectedProject() > 0 {
		m.switchToProject(m.appState.SelectedProject() - 1)
	} else {
		m.notificationState.Add(state.LevelInfo, "Already at the first project")
	}
	return m, nil
}

// handleNextProject switches to the next project.
func (m Model) handleNextProject() (tea.Model, tea.Cmd) {
	if m.appState.SelectedProject() < len(m.appState.Projects())-1 {
		m.switchToProject(m.appState.SelectedProject() + 1)
	} else {
		m.notificationState.Add(state.LevelInfo, "Already at the last project")
	}
	return m, nil
}

// handleCreateProject initiates project creation.
func (m Model) handleCreateProject() (tea.Model, tea.Cmd) {
	m.formState.FormProjectName = ""
	m.formState.FormProjectDescription = ""
	m.formState.FormProjectConfirm = true
	m.formState.ProjectForm = huhforms.CreateProjectForm(
		&m.formState.FormProjectName,
		&m.formState.FormProjectDescription,
		&m.formState.FormProjectConfirm,
	).WithTheme(huhforms.CreatePasoTheme(m.config.ColorScheme))
	m.formState.SnapshotProjectFormInitialValues() // Snapshot for change detection
	m.uiState.SetMode(state.ProjectFormMode)
	return m, m.formState.ProjectForm.Init()
}

// ============================================================================
// INPUT MODE HANDLERS
// ============================================================================

// handleInputMode handles text input for column creation/editing.
func (m Model) handleInputMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		return m.handleInputConfirm()
	case "esc":
		// Check for changes before closing
		shouldConfirm := false
		contextMsg := ""

		if m.uiState.Mode() == state.AddColumnMode {
			// AddColumnMode: confirm if user typed anything
			shouldConfirm = !m.inputState.IsEmpty()
			contextMsg = "Discard column?"
		} else if m.uiState.Mode() == state.EditColumnMode {
			// EditColumnMode: confirm if text changed from original
			shouldConfirm = m.inputState.HasInputChanges()
			contextMsg = "Discard changes?"
		}

		if shouldConfirm {
			m.uiState.SetDiscardContext(&state.DiscardContext{
				SourceMode: m.uiState.Mode(),
				Message:    contextMsg,
			})
			m.uiState.SetMode(state.DiscardConfirmMode)
			return m, nil
		}

		// No changes - immediate close
		return m.handleInputCancel()
	case "backspace", "ctrl+h":
		m.inputState.Backspace()
		return m, nil
	default:
		key := msg.String()
		if len(key) == 1 {
			m.inputState.AppendChar(rune(key[0]))
		}
		return m, nil
	}
}

// handleInputConfirm processes the input and creates/renames column.
func (m Model) handleInputConfirm() (tea.Model, tea.Cmd) {
	if strings.TrimSpace(m.inputState.Buffer) == "" {
		m.uiState.SetMode(state.NormalMode)
		m.inputState.Clear()
		return m, nil
	}

	if m.uiState.Mode() == state.AddColumnMode {
		return m.createColumn()
	}
	return m.renameColumn()
}

// handleInputCancel cancels the input operation.
func (m Model) handleInputCancel() (tea.Model, tea.Cmd) {
	m.uiState.SetMode(state.NormalMode)
	m.inputState.Clear()
	return m, nil
}

// createColumn creates a new column with the input buffer as name.
func (m Model) createColumn() (tea.Model, tea.Cmd) {
	var afterColumnID *int
	if len(m.appState.Columns()) > 0 {
		currentCol := m.getCurrentColumn()
		if currentCol != nil {
			afterColumnID = &currentCol.ID
		}
	}

	projectID := 0
	if project := m.getCurrentProject(); project != nil {
		projectID = project.ID
	}

	column, err := m.repo.CreateColumn(context.Background(), strings.TrimSpace(m.inputState.Buffer), projectID, afterColumnID)
	if err != nil {
		log.Printf("Error creating column: %v", err)
		m.notificationState.Add(state.LevelError, "Failed to create column")
	} else {
		columns, err := m.repo.GetColumnsByProject(context.Background(), projectID)
		if err != nil {
			log.Printf("Error reloading columns: %v", err)
			m.notificationState.Add(state.LevelError, "Failed to reload columns")
		}
		m.appState.SetColumns(columns)
		m.appState.Tasks()[column.ID] = []*models.TaskSummary{}
		if afterColumnID != nil {
			m.uiState.SetSelectedColumn(m.uiState.SelectedColumn() + 1)
		}
	}

	m.uiState.SetMode(state.NormalMode)
	m.inputState.Clear()
	return m, nil
}

// renameColumn renames the current column with the input buffer.
func (m Model) renameColumn() (tea.Model, tea.Cmd) {
	column := m.getCurrentColumn()
	if column != nil {
		err := m.repo.UpdateColumnName(context.Background(), column.ID, strings.TrimSpace(m.inputState.Buffer))
		if err != nil {
			log.Printf("Error updating column: %v", err)
			m.notificationState.Add(state.LevelError, "Failed to rename column")
		} else {
			column.Name = strings.TrimSpace(m.inputState.Buffer)
		}
	}

	m.uiState.SetMode(state.NormalMode)
	m.inputState.Clear()
	return m, nil
}

// ============================================================================
// CONFIRMATION HANDLERS
// ============================================================================

// handleDeleteConfirm handles task deletion confirmation.
func (m Model) handleDeleteConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		return m.confirmDeleteTask()
	case "n", "N", "esc":
		m.uiState.SetMode(state.NormalMode)
		return m, nil
	}
	return m, nil
}

// confirmDeleteTask performs the actual task deletion.
func (m Model) confirmDeleteTask() (tea.Model, tea.Cmd) {
	task := m.getCurrentTask()
	if task != nil {
		err := m.repo.DeleteTask(context.Background(), task.ID)
		if err != nil {
			log.Printf("Error deleting task: %v", err)
			m.notificationState.Add(state.LevelError, "Failed to delete task")
		} else {
			m.removeCurrentTask()
		}
	}
	m.uiState.SetMode(state.NormalMode)
	return m, nil
}

// handleDiscardConfirm handles discard confirmation for forms and inputs.
// This provides a generic Y/N/ESC handler that works for all discard scenarios.
func (m Model) handleDiscardConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	ctx := m.uiState.DiscardContext()
	if ctx == nil {
		// Safety: if context is missing, return to normal mode
		m.uiState.SetMode(state.NormalMode)
		return m, nil
	}

	switch msg.String() {
	case "y", "Y":
		// User confirmed discard - clear form/input and return to normal mode
		return m.confirmDiscard()

	case "n", "N", "esc":
		// User cancelled - return to source mode without clearing
		m.uiState.SetMode(ctx.SourceMode)
		m.uiState.ClearDiscardContext()
		return m, nil
	}

	return m, nil
}

// confirmDiscard performs the actual discard operation based on context.
func (m Model) confirmDiscard() (tea.Model, tea.Cmd) {
	ctx := m.uiState.DiscardContext()
	if ctx == nil {
		m.uiState.SetMode(state.NormalMode)
		return m, nil
	}

	// Clear the appropriate form/input based on source mode
	switch ctx.SourceMode {
	case state.TicketFormMode:
		m.formState.ClearTicketForm()

	case state.ProjectFormMode:
		m.formState.ClearProjectForm()

	case state.AddColumnMode, state.EditColumnMode:
		m.inputState.Clear()
	}

	// Always return to normal mode after discard
	m.uiState.SetMode(state.NormalMode)
	m.uiState.ClearDiscardContext()

	return m, tea.ClearScreen
}

// handleDeleteColumnConfirm handles column deletion confirmation.
func (m Model) handleDeleteColumnConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		return m.confirmDeleteColumn()
	case "n", "N", "esc":
		m.uiState.SetMode(state.NormalMode)
		return m, nil
	}
	return m, nil
}

// confirmDeleteColumn performs the actual column deletion.
func (m Model) confirmDeleteColumn() (tea.Model, tea.Cmd) {
	column := m.getCurrentColumn()
	if column != nil {
		err := m.repo.DeleteColumn(context.Background(), column.ID)
		if err != nil {
			log.Printf("Error deleting column: %v", err)
			m.notificationState.Add(state.LevelError, "Failed to delete column")
		} else {
			delete(m.appState.Tasks(), column.ID)
			m.removeCurrentColumn()
		}
	}
	m.uiState.SetMode(state.NormalMode)
	return m, nil
}

// ============================================================================
// VIEW/HELP MODE HANDLERS
// ============================================================================

// handleHelpMode handles input in the help screen.
func (m Model) handleHelpMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case m.config.KeyMappings.ShowHelp, m.config.KeyMappings.Quit, "esc", "enter", " ":
		m.uiState.SetMode(state.NormalMode)
		return m, nil
	}
	return m, nil
}

// ============================================================================
// SEARCH MODE HANDLERS
// ============================================================================

// handleEnterSearch enters search mode and clears any previous search state.
func (m Model) handleEnterSearch() (tea.Model, tea.Cmd) {
	m.searchState.Clear()
	m.searchState.Deactivate()
	m.uiState.SetMode(state.SearchMode)
	return m, nil
}

// handleSearchMode handles keyboard input in search mode.
func (m Model) handleSearchMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		return m.handleSearchConfirm()
	case "esc":
		return m.handleSearchCancel()
	case "backspace", "ctrl+h":
		if m.searchState.Backspace() {
			return m.executeSearch()
		}
		return m, nil
	default:
		key := msg.String()
		if len(key) == 1 {
			if m.searchState.AppendChar(rune(key[0])) {
				return m.executeSearch()
			}
		}
		return m, nil
	}
}

// handleSearchConfirm activates the filter and returns to normal mode.
// The search query persists and continues to filter the kanban view.
func (m Model) handleSearchConfirm() (tea.Model, tea.Cmd) {
	m.searchState.Activate()
	m.uiState.SetMode(state.NormalMode)
	return m, nil
}

// handleSearchCancel clears the search and returns to normal mode.
// All tasks are shown again.
func (m Model) handleSearchCancel() (tea.Model, tea.Cmd) {
	m.searchState.Clear()
	m.searchState.Deactivate()
	m.uiState.SetMode(state.NormalMode)
	return m.executeSearch() // Reload all tasks
}

// executeSearch runs the search query and updates the task list.
// If the query is empty, all tasks are loaded. Otherwise, only matching tasks are loaded.
func (m Model) executeSearch() (tea.Model, tea.Cmd) {
	project := m.getCurrentProject()
	if project == nil {
		return m, nil
	}

	ctx := context.Background()
	var tasksByColumn map[int][]*models.TaskSummary
	var err error

	if m.searchState.Query == "" {
		tasksByColumn, err = m.repo.GetTaskSummariesByProject(ctx, project.ID)
	} else {
		tasksByColumn, err = m.repo.GetTaskSummariesByProjectFiltered(ctx, project.ID, m.searchState.Query)
	}

	if err != nil {
		log.Printf("Error filtering tasks: %v", err)
		return m, nil
	}

	m.appState.SetTasks(tasksByColumn)
	// Reset task selection to 0 to avoid out-of-bounds
	m.uiState.SetSelectedTask(0)

	return m, nil
}
