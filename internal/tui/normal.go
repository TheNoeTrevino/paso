package tui

import (
	"log/slog"

	tea "charm.land/bubbletea/v2"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui/components"
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
	case km.ToggleView:
		return m.handleToggleView()
	case km.ChangeStatus:
		return m.handleChangeStatus()
	case km.SortList:
		return m.handleSortList()
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
	// List view navigation
	if m.listViewState.IsListView() {
		if m.listViewState.SelectedRow() > 0 {
			m.listViewState.SetSelectedRow(m.listViewState.SelectedRow() - 1)

			// Ensure row is visible by adjusting scroll offset
			listHeight := m.uiState.ContentHeight()
			const reservedHeight = 6
			visibleRows := max(listHeight-reservedHeight, 1)
			m.listViewState.EnsureRowVisible(visibleRows)
		} else {
			m.notificationState.Add(state.LevelInfo, "Already at the first task")
		}
		return m, nil
	}

	// Kanban navigation
	if m.uiState.SelectedTask() > 0 {
		m.uiState.SetSelectedTask(m.uiState.SelectedTask() - 1)

		// Ensure task is visible by adjusting column scroll offset
		if m.uiState.SelectedColumn() < len(m.appState.Columns()) {
			currentCol := m.appState.Columns()[m.uiState.SelectedColumn()]
			columnHeight := m.uiState.ContentHeight()
			const columnOverhead = 5 // Includes reserved space for top and bottom indicators
			maxTasksVisible := max((columnHeight-columnOverhead)/components.TaskCardHeight, 1)
			m.uiState.EnsureTaskVisible(currentCol.ID, m.uiState.SelectedTask(), maxTasksVisible)
		}
	} else {
		m.notificationState.Add(state.LevelInfo, "Already at the first task")
	}
	return m, nil
}

// handleNavigateDown moves selection to the next task.
func (m Model) handleNavigateDown() (tea.Model, tea.Cmd) {
	// List view navigation
	if m.listViewState.IsListView() {
		rows := m.buildListViewRows()
		if m.listViewState.SelectedRow() < len(rows)-1 {
			m.listViewState.SetSelectedRow(m.listViewState.SelectedRow() + 1)

			// Ensure row is visible by adjusting scroll offset
			listHeight := m.uiState.ContentHeight()
			const reservedHeight = 6
			visibleRows := max(listHeight-reservedHeight, 1)
			m.listViewState.EnsureRowVisible(visibleRows)
		} else if len(rows) > 0 {
			m.notificationState.Add(state.LevelInfo, "Already at the last task")
		}
		return m, nil
	}

	// Kanban navigation
	currentTasks := m.getCurrentTasks()
	if len(currentTasks) > 0 && m.uiState.SelectedTask() < len(currentTasks)-1 {
		m.uiState.SetSelectedTask(m.uiState.SelectedTask() + 1)

		// Ensure task is visible by adjusting column scroll offset
		if m.uiState.SelectedColumn() < len(m.appState.Columns()) {
			currentCol := m.appState.Columns()[m.uiState.SelectedColumn()]
			columnHeight := m.uiState.ContentHeight()
			const columnOverhead = 5 // Includes reserved space for top and bottom indicators
			maxTasksVisible := max((columnHeight-columnOverhead)/components.TaskCardHeight, 1)
			m.uiState.EnsureTaskVisible(currentCol.ID, m.uiState.SelectedTask(), maxTasksVisible)
		}
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

	ctx, cancel := m.dbContext()
	defer cancel()
	taskDetail, err := m.repo.GetTaskDetail(ctx, task.ID)
	if err != nil {
		m.handleDBError(err, "Loading task details")
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

	// Load timestamps, type, and priority for metadata display
	m.formState.FormCreatedAt = taskDetail.CreatedAt
	m.formState.FormUpdatedAt = taskDetail.UpdatedAt
	m.formState.FormTypeDescription = taskDetail.TypeDescription
	m.formState.FormPriorityDescription = taskDetail.PriorityDescription
	m.formState.FormPriorityColor = taskDetail.PriorityColor

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
	ctx, cancel := m.dbContext()
	defer cancel()
	taskCount, err := m.repo.GetTaskCountByColumn(ctx, column.ID)
	if err != nil {
		slog.Error("Error getting task count", "error", err)
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
