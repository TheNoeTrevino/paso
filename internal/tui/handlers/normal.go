package handlers

import (
	"log/slog"

	tea "charm.land/bubbletea/v2"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui"
	"github.com/thenoetrevino/paso/internal/tui/components"
	"github.com/thenoetrevino/paso/internal/tui/huhforms"
	"github.com/thenoetrevino/paso/internal/tui/modelops"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// ============================================================================
// NORMAL MODE HANDLERS
// ============================================================================

// HandleNormalMode dispatches key events in NormalMode to specific handlers.
func HandleNormalMode(m *tui.Model, msg tea.KeyMsg) tea.Cmd {
	m.NotificationState.Clear()

	key := msg.String()
	km := m.Config.KeyMappings

	switch key {
	case km.Quit, "ctrl+c":
		return handleQuit(m)
	case km.ShowHelp:
		return handleShowHelp(m)
	case km.AddTask:
		return handleAddTask(m)
	case km.EditTask:
		return handleEditTask(m)
	case km.DeleteTask:
		return handleDeleteTask(m)
	case km.ViewTask:
		return handleEditTask(m)
	case km.CreateColumn:
		return handleCreateColumn(m)
	case km.RenameColumn:
		return handleRenameColumn(m)
	case km.DeleteColumn:
		return handleDeleteColumn(m)
	case km.ScrollViewportRight:
		return handleScrollRight(m)
	case km.ScrollViewportLeft:
		return handleScrollLeft(m)
	case km.PrevColumn, "left":
		return handleNavigateLeft(m)
	case km.NextColumn, "right":
		return handleNavigateRight(m)
	case km.NextTask, "down":
		return handleNavigateDown(m)
	case km.PrevTask, "up":
		return handleNavigateUp(m)
	case km.MoveTaskRight:
		return handleMoveTaskRight(m)
	case km.MoveTaskLeft:
		return handleMoveTaskLeft(m)
	case km.MoveTaskUp:
		return handleMoveTaskUp(m)
	case km.MoveTaskDown:
		return handleMoveTaskDown(m)
	case km.PrevProject:
		return handlePrevProject(m)
	case km.NextProject:
		return handleNextProject(m)
	case km.CreateProject:
		return handleCreateProject(m)
	case km.ToggleView:
		return HandleToggleView(m)
	case km.ChangeStatus:
		return HandleChangeStatus(m)
	case km.SortList:
		return HandleSortList(m)
	case "/":
		return HandleEnterSearch(m)
	}

	return nil
}

// handleQuit exits the application.
func handleQuit(m *tui.Model) tea.Cmd {
	return tea.Quit
}

// handleShowHelp shows the help screen.
func handleShowHelp(m *tui.Model) tea.Cmd {
	m.UiState.SetMode(state.HelpMode)
	return nil
}

// handleNavigateLeft moves selection to the previous column.
func handleNavigateLeft(m *tui.Model) tea.Cmd {
	if m.UiState.SelectedColumn() > 0 {
		m.UiState.SetSelectedColumn(m.UiState.SelectedColumn() - 1)
		m.UiState.SetSelectedTask(0)
		m.UiState.EnsureSelectionVisible(m.UiState.SelectedColumn())
	} else {
		m.NotificationState.Add(state.LevelInfo, "Already at the first column")
	}
	return nil
}

// handleNavigateRight moves selection to the next column.
func handleNavigateRight(m *tui.Model) tea.Cmd {
	if m.UiState.SelectedColumn() < len(m.AppState.Columns())-1 {
		m.UiState.SetSelectedColumn(m.UiState.SelectedColumn() + 1)
		m.UiState.SetSelectedTask(0)
		m.UiState.EnsureSelectionVisible(m.UiState.SelectedColumn())
	} else {
		m.NotificationState.Add(state.LevelInfo, "Already at the last column")
	}
	return nil
}

// handleNavigateUp moves selection to the previous task.
func handleNavigateUp(m *tui.Model) tea.Cmd {
	// List view navigation
	if m.ListViewState.IsListView() {
		if m.ListViewState.SelectedRow() > 0 {
			m.ListViewState.SetSelectedRow(m.ListViewState.SelectedRow() - 1)

			// Ensure row is visible by adjusting scroll offset
			listHeight := m.UiState.ContentHeight()
			const reservedHeight = 6
			visibleRows := max(listHeight-reservedHeight, 1)
			m.ListViewState.EnsureRowVisible(visibleRows)
		} else {
			m.NotificationState.Add(state.LevelInfo, "Already at the first task")
		}
		return nil
	}

	// Kanban navigation
	if m.UiState.SelectedTask() > 0 {
		m.UiState.SetSelectedTask(m.UiState.SelectedTask() - 1)

		// Ensure task is visible by adjusting column scroll offset
		if m.UiState.SelectedColumn() < len(m.AppState.Columns()) {
			currentCol := m.AppState.Columns()[m.UiState.SelectedColumn()]
			columnHeight := m.UiState.ContentHeight()
			const columnOverhead = 5 // Includes reserved space for top and bottom indicators
			maxTasksVisible := max((columnHeight-columnOverhead)/components.TaskCardHeight, 1)
			m.UiState.EnsureTaskVisible(currentCol.ID, m.UiState.SelectedTask(), maxTasksVisible)
		}
	} else {
		m.NotificationState.Add(state.LevelInfo, "Already at the first task")
	}
	return nil
}

// handleNavigateDown moves selection to the next task.
func handleNavigateDown(m *tui.Model) tea.Cmd {
	// List view navigation
	if m.ListViewState.IsListView() {
		rows := modelops.BuildListViewRows(m)
		if m.ListViewState.SelectedRow() < len(rows)-1 {
			m.ListViewState.SetSelectedRow(m.ListViewState.SelectedRow() + 1)

			// Ensure row is visible by adjusting scroll offset
			listHeight := m.UiState.ContentHeight()
			const reservedHeight = 6
			visibleRows := max(listHeight-reservedHeight, 1)
			m.ListViewState.EnsureRowVisible(visibleRows)
		} else if len(rows) > 0 {
			m.NotificationState.Add(state.LevelInfo, "Already at the last task")
		}
		return nil
	}

	// Kanban navigation
	currentTasks := modelops.GetCurrentTasks(m)
	if len(currentTasks) > 0 && m.UiState.SelectedTask() < len(currentTasks)-1 {
		m.UiState.SetSelectedTask(m.UiState.SelectedTask() + 1)

		// Ensure task is visible by adjusting column scroll offset
		if m.UiState.SelectedColumn() < len(m.AppState.Columns()) {
			currentCol := m.AppState.Columns()[m.UiState.SelectedColumn()]
			columnHeight := m.UiState.ContentHeight()
			const columnOverhead = 5 // Includes reserved space for top and bottom indicators
			maxTasksVisible := max((columnHeight-columnOverhead)/components.TaskCardHeight, 1)
			m.UiState.EnsureTaskVisible(currentCol.ID, m.UiState.SelectedTask(), maxTasksVisible)
		}
	} else if len(currentTasks) > 0 {
		m.NotificationState.Add(state.LevelInfo, "Already at the last task")
	}
	return nil
}

// handleScrollRight scrolls the viewport right.
func handleScrollRight(m *tui.Model) tea.Cmd {
	if m.UiState.ViewportOffset()+m.UiState.ViewportSize() < len(m.AppState.Columns()) {
		m.UiState.SetViewportOffset(m.UiState.ViewportOffset() + 1)
		if m.UiState.SelectedColumn() < m.UiState.ViewportOffset() {
			m.UiState.SetSelectedColumn(m.UiState.ViewportOffset())
			m.UiState.SetSelectedTask(0)
		}
	} else {
		m.NotificationState.Add(state.LevelInfo, "Already at the rightmost view")
	}
	return nil
}

// handleScrollLeft scrolls the viewport left.
func handleScrollLeft(m *tui.Model) tea.Cmd {
	if m.UiState.ViewportOffset() > 0 {
		m.UiState.SetViewportOffset(m.UiState.ViewportOffset() - 1)
		if m.UiState.SelectedColumn() >= m.UiState.ViewportOffset()+m.UiState.ViewportSize() {
			m.UiState.SetSelectedColumn(m.UiState.ViewportOffset() + m.UiState.ViewportSize() - 1)
			m.UiState.SetSelectedTask(0)
		}
	} else {
		m.NotificationState.Add(state.LevelInfo, "Already at the leftmost view")
	}
	return nil
}

// handleAddTask initiates adding a new task.
func handleAddTask(m *tui.Model) tea.Cmd {
	if len(m.AppState.Columns()) == 0 {
		m.NotificationState.Add(state.LevelError, "Cannot add task: No columns exist. Create a column first with 'C'")
		return nil
	}
	m.FormState.FormTitle = ""
	m.FormState.FormDescription = ""
	m.FormState.FormLabelIDs = []int{}
	m.FormState.FormParentIDs = []int{}
	m.FormState.FormChildIDs = []int{}
	m.FormState.FormParentRefs = []*models.TaskReference{}
	m.FormState.FormChildRefs = []*models.TaskReference{}
	m.FormState.FormConfirm = true
	m.FormState.EditingTaskID = 0

	// Calculate description height (will be dynamic in Phase 4)
	descriptionLines := 10

	m.FormState.TicketForm = huhforms.CreateTicketForm(
		&m.FormState.FormTitle,
		&m.FormState.FormDescription,
		&m.FormState.FormConfirm,
		descriptionLines,
	).WithTheme(huhforms.CreatePasoTheme(m.Config.ColorScheme))
	m.FormState.SnapshotTicketFormInitialValues() // Snapshot for change detection
	m.UiState.SetMode(state.TicketFormMode)
	return m.FormState.TicketForm.Init()
}

// handleEditTask initiates editing the selected task.
func handleEditTask(m *tui.Model) tea.Cmd {
	task := modelops.GetCurrentTask(m)
	if task == nil {
		m.NotificationState.Add(state.LevelError, "No task selected to edit")
		return nil
	}

	ctx, cancel := m.DbContext()
	defer cancel()
	taskDetail, err := m.Repo.GetTaskDetail(ctx, task.ID)
	if err != nil {
		m.HandleDBError(err, "Loading task details")
		return nil
	}

	m.FormState.FormTitle = taskDetail.Title
	m.FormState.FormDescription = taskDetail.Description
	m.FormState.FormLabelIDs = make([]int, len(taskDetail.Labels))
	for i, label := range taskDetail.Labels {
		m.FormState.FormLabelIDs[i] = label.ID
	}

	// Load parent relationships
	m.FormState.FormParentIDs = make([]int, len(taskDetail.ParentTasks))
	m.FormState.FormParentRefs = taskDetail.ParentTasks
	for i, parent := range taskDetail.ParentTasks {
		m.FormState.FormParentIDs[i] = parent.ID
	}

	// Load child relationships
	m.FormState.FormChildIDs = make([]int, len(taskDetail.ChildTasks))
	m.FormState.FormChildRefs = taskDetail.ChildTasks
	for i, child := range taskDetail.ChildTasks {
		m.FormState.FormChildIDs[i] = child.ID
	}

	// Load timestamps, type, and priority for metadata display
	m.FormState.FormCreatedAt = taskDetail.CreatedAt
	m.FormState.FormUpdatedAt = taskDetail.UpdatedAt
	m.FormState.FormTypeDescription = taskDetail.TypeDescription
	m.FormState.FormPriorityDescription = taskDetail.PriorityDescription
	m.FormState.FormPriorityColor = taskDetail.PriorityColor

	m.FormState.FormConfirm = true
	m.FormState.EditingTaskID = task.ID

	// Calculate description height (will be dynamic in Phase 4)
	descriptionLines := 10

	m.FormState.TicketForm = huhforms.CreateTicketForm(
		&m.FormState.FormTitle,
		&m.FormState.FormDescription,
		&m.FormState.FormConfirm,
		descriptionLines,
	).WithTheme(huhforms.CreatePasoTheme(m.Config.ColorScheme))
	m.FormState.SnapshotTicketFormInitialValues() // Snapshot for change detection
	m.UiState.SetMode(state.TicketFormMode)
	return m.FormState.TicketForm.Init()
}

// handleDeleteTask initiates task deletion confirmation.
func handleDeleteTask(m *tui.Model) tea.Cmd {
	if modelops.GetCurrentTask(m) == nil {
		m.NotificationState.Add(state.LevelError, "No task selected to delete")
		return nil
	}
	m.UiState.SetMode(state.DeleteConfirmMode)
	return nil
}

// handleMoveTaskRight moves the task to the next column.
func handleMoveTaskRight(m *tui.Model) tea.Cmd {
	if modelops.GetCurrentTask(m) != nil {
		modelops.MoveTaskRight(m)
	}
	return nil
}

// handleMoveTaskLeft moves the task to the previous column.
func handleMoveTaskLeft(m *tui.Model) tea.Cmd {
	if modelops.GetCurrentTask(m) != nil {
		modelops.MoveTaskLeft(m)
	}
	return nil
}

// handleMoveTaskUp moves the task up within its column.
func handleMoveTaskUp(m *tui.Model) tea.Cmd {
	if modelops.GetCurrentTask(m) != nil {
		modelops.MoveTaskUp(m)
	}
	return nil
}

// handleMoveTaskDown moves the task down within its column.
func handleMoveTaskDown(m *tui.Model) tea.Cmd {
	if modelops.GetCurrentTask(m) != nil {
		modelops.MoveTaskDown(m)
	}
	return nil
}

// handleCreateColumn initiates column creation.
func handleCreateColumn(m *tui.Model) tea.Cmd {
	m.UiState.SetMode(state.AddColumnMode)
	m.InputState.Prompt = "New column name:"
	m.InputState.Buffer = ""
	return nil
}

// handleRenameColumn initiates column renaming.
func handleRenameColumn(m *tui.Model) tea.Cmd {
	column := modelops.GetCurrentColumn(m)
	if column == nil {
		m.NotificationState.Add(state.LevelError, "No column selected to rename")
		return nil
	}
	m.UiState.SetMode(state.EditColumnMode)
	m.InputState.Buffer = column.Name
	m.InputState.Prompt = "Rename column:"
	m.InputState.SnapshotInitialBuffer() // Snapshot for change detection
	return nil
}

// handleDeleteColumn initiates column deletion confirmation.
func handleDeleteColumn(m *tui.Model) tea.Cmd {
	column := modelops.GetCurrentColumn(m)
	if column == nil {
		m.NotificationState.Add(state.LevelError, "No column selected to delete")
		return nil
	}
	ctx, cancel := m.DbContext()
	defer cancel()
	taskCount, err := m.Repo.GetTaskCountByColumn(ctx, column.ID)
	if err != nil {
		slog.Error("Error getting task count", "error", err)
		m.NotificationState.Add(state.LevelError, "Error getting column info")
		return nil
	}
	m.InputState.DeleteColumnTaskCount = taskCount
	m.UiState.SetMode(state.DeleteColumnConfirmMode)
	return nil
}

// handlePrevProject switches to the previous project.
func handlePrevProject(m *tui.Model) tea.Cmd {
	if m.AppState.SelectedProject() > 0 {
		modelops.SwitchToProject(m, m.AppState.SelectedProject()-1)
	} else {
		m.NotificationState.Add(state.LevelInfo, "Already at the first project")
	}
	return nil
}

// handleNextProject switches to the next project.
func handleNextProject(m *tui.Model) tea.Cmd {
	if m.AppState.SelectedProject() < len(m.AppState.Projects())-1 {
		modelops.SwitchToProject(m, m.AppState.SelectedProject()+1)
	} else {
		m.NotificationState.Add(state.LevelInfo, "Already at the last project")
	}
	return nil
}

// handleCreateProject initiates project creation.
func handleCreateProject(m *tui.Model) tea.Cmd {
	m.FormState.FormProjectName = ""
	m.FormState.FormProjectDescription = ""
	m.FormState.FormProjectConfirm = true
	m.FormState.ProjectForm = huhforms.CreateProjectForm(
		&m.FormState.FormProjectName,
		&m.FormState.FormProjectDescription,
		&m.FormState.FormProjectConfirm,
	).WithTheme(huhforms.CreatePasoTheme(m.Config.ColorScheme))
	m.FormState.SnapshotProjectFormInitialValues() // Snapshot for change detection
	m.UiState.SetMode(state.ProjectFormMode)
	return m.FormState.ProjectForm.Init()
}
