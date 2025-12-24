package handlers

import (
	"log/slog"

	tea "charm.land/bubbletea/v2"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui/components"
	"github.com/thenoetrevino/paso/internal/tui/huhforms"
	"github.com/thenoetrevino/paso/internal/tui/modelops"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// ============================================================================
// NORMAL MODE HANDLERS
// ============================================================================

// HandleNormalMode dispatches key events in NormalMode to specific handlers.
func (w *Wrapper) HandleNormalMode(msg tea.KeyMsg) (*Wrapper, tea.Cmd) {
	w.NotificationState.Clear()

	key := msg.String()
	km := w.Config.KeyMappings

	switch key {
	case km.Quit, "ctrl+c":
		return w.handleQuit()
	case km.ShowHelp:
		return w.handleShowHelp()
	case km.AddTask:
		return w.handleAddTask()
	case km.EditTask:
		return w.handleEditTask()
	case km.DeleteTask:
		return w.handleDeleteTask()
	case km.ViewTask:
		return w.handleEditTask()
	case km.CreateColumn:
		return w.handleCreateColumn()
	case km.RenameColumn:
		return w.handleRenameColumn()
	case km.DeleteColumn:
		return w.handleDeleteColumn()
	case km.ScrollViewportRight:
		return w.handleScrollRight()
	case km.ScrollViewportLeft:
		return w.handleScrollLeft()
	case km.PrevColumn, "left":
		return w.handleNavigateLeft()
	case km.NextColumn, "right":
		return w.handleNavigateRight()
	case km.NextTask, "down":
		return w.handleNavigateDown()
	case km.PrevTask, "up":
		return w.handleNavigateUp()
	case km.MoveTaskRight:
		return w.handleMoveTaskRight()
	case km.MoveTaskLeft:
		return w.handleMoveTaskLeft()
	case km.MoveTaskUp:
		return w.handleMoveTaskUp()
	case km.MoveTaskDown:
		return w.handleMoveTaskDown()
	case km.PrevProject:
		return w.handlePrevProject()
	case km.NextProject:
		return w.handleNextProject()
	case km.CreateProject:
		return w.handleCreateProject()
	case km.ToggleView:
		return w.HandleToggleView()
	case km.ChangeStatus:
		return w.HandleChangeStatus()
	case km.SortList:
		return w.HandleSortList()
	case "/":
		return w.HandleEnterSearch()
	}

	return w, nil
}

// handleQuit exits the application.
func (w *Wrapper) handleQuit() (*Wrapper, tea.Cmd) {
	return w, tea.Quit
}

// handleShowHelp shows the help screen.
func (w *Wrapper) handleShowHelp() (*Wrapper, tea.Cmd) {
	w.UiState.SetMode(state.HelpMode)
	return w, nil
}

// handleNavigateLeft moves selection to the previous column.
func (w *Wrapper) handleNavigateLeft() (*Wrapper, tea.Cmd) {
	if w.UiState.SelectedColumn() > 0 {
		w.UiState.SetSelectedColumn(w.UiState.SelectedColumn() - 1)
		w.UiState.SetSelectedTask(0)
		w.UiState.EnsureSelectionVisible(w.UiState.SelectedColumn())
	} else {
		w.NotificationState.Add(state.LevelInfo, "Already at the first column")
	}
	return w, nil
}

// handleNavigateRight moves selection to the next column.
func (w *Wrapper) handleNavigateRight() (*Wrapper, tea.Cmd) {
	if w.UiState.SelectedColumn() < len(w.AppState.Columns())-1 {
		w.UiState.SetSelectedColumn(w.UiState.SelectedColumn() + 1)
		w.UiState.SetSelectedTask(0)
		w.UiState.EnsureSelectionVisible(w.UiState.SelectedColumn())
	} else {
		w.NotificationState.Add(state.LevelInfo, "Already at the last column")
	}
	return w, nil
}

// handleNavigateUp moves selection to the previous task.
func (w *Wrapper) handleNavigateUp() (*Wrapper, tea.Cmd) {
	// List view navigation
	if w.ListViewState.IsListView() {
		if w.ListViewState.SelectedRow() > 0 {
			w.ListViewState.SetSelectedRow(w.ListViewState.SelectedRow() - 1)

			// Ensure row is visible by adjusting scroll offset
			listHeight := w.UiState.ContentHeight()
			const reservedHeight = 6
			visibleRows := max(listHeight-reservedHeight, 1)
			w.ListViewState.EnsureRowVisible(visibleRows)
		} else {
			w.NotificationState.Add(state.LevelInfo, "Already at the first task")
		}
		return w, nil
	}

	// Kanban navigation
	if w.UiState.SelectedTask() > 0 {
		w.UiState.SetSelectedTask(w.UiState.SelectedTask() - 1)

		// Ensure task is visible by adjusting column scroll offset
		if w.UiState.SelectedColumn() < len(w.AppState.Columns()) {
			currentCol := w.AppState.Columns()[w.UiState.SelectedColumn()]
			columnHeight := w.UiState.ContentHeight()
			const columnOverhead = 5 // Includes reserved space for top and bottom indicators
			maxTasksVisible := max((columnHeight-columnOverhead)/components.TaskCardHeight, 1)
			w.UiState.EnsureTaskVisible(currentCol.ID, w.UiState.SelectedTask(), maxTasksVisible)
		}
	} else {
		w.NotificationState.Add(state.LevelInfo, "Already at the first task")
	}
	return w, nil
}

// handleNavigateDown moves selection to the next task.
func (w *Wrapper) handleNavigateDown() (*Wrapper, tea.Cmd) {
	ops := modelops.New(w.Model)

	// List view navigation
	if w.ListViewState.IsListView() {
		rows := ops.BuildListViewRows()
		if w.ListViewState.SelectedRow() < len(rows)-1 {
			w.ListViewState.SetSelectedRow(w.ListViewState.SelectedRow() + 1)

			// Ensure row is visible by adjusting scroll offset
			listHeight := w.UiState.ContentHeight()
			const reservedHeight = 6
			visibleRows := max(listHeight-reservedHeight, 1)
			w.ListViewState.EnsureRowVisible(visibleRows)
		} else if len(rows) > 0 {
			w.NotificationState.Add(state.LevelInfo, "Already at the last task")
		}
		return w, nil
	}

	// Kanban navigation
	currentTasks := ops.GetCurrentTasks()
	if len(currentTasks) > 0 && w.UiState.SelectedTask() < len(currentTasks)-1 {
		w.UiState.SetSelectedTask(w.UiState.SelectedTask() + 1)

		// Ensure task is visible by adjusting column scroll offset
		if w.UiState.SelectedColumn() < len(w.AppState.Columns()) {
			currentCol := w.AppState.Columns()[w.UiState.SelectedColumn()]
			columnHeight := w.UiState.ContentHeight()
			const columnOverhead = 5 // Includes reserved space for top and bottom indicators
			maxTasksVisible := max((columnHeight-columnOverhead)/components.TaskCardHeight, 1)
			w.UiState.EnsureTaskVisible(currentCol.ID, w.UiState.SelectedTask(), maxTasksVisible)
		}
	} else if len(currentTasks) > 0 {
		w.NotificationState.Add(state.LevelInfo, "Already at the last task")
	}
	return w, nil
}

// handleScrollRight scrolls the viewport right.
func (w *Wrapper) handleScrollRight() (*Wrapper, tea.Cmd) {
	if w.UiState.ViewportOffset()+w.UiState.ViewportSize() < len(w.AppState.Columns()) {
		w.UiState.SetViewportOffset(w.UiState.ViewportOffset() + 1)
		if w.UiState.SelectedColumn() < w.UiState.ViewportOffset() {
			w.UiState.SetSelectedColumn(w.UiState.ViewportOffset())
			w.UiState.SetSelectedTask(0)
		}
	} else {
		w.NotificationState.Add(state.LevelInfo, "Already at the rightmost view")
	}
	return w, nil
}

// handleScrollLeft scrolls the viewport left.
func (w *Wrapper) handleScrollLeft() (*Wrapper, tea.Cmd) {
	if w.UiState.ViewportOffset() > 0 {
		w.UiState.SetViewportOffset(w.UiState.ViewportOffset() - 1)
		if w.UiState.SelectedColumn() >= w.UiState.ViewportOffset()+w.UiState.ViewportSize() {
			w.UiState.SetSelectedColumn(w.UiState.ViewportOffset() + w.UiState.ViewportSize() - 1)
			w.UiState.SetSelectedTask(0)
		}
	} else {
		w.NotificationState.Add(state.LevelInfo, "Already at the leftmost view")
	}
	return w, nil
}

// handleAddTask initiates adding a new task.
func (w *Wrapper) handleAddTask() (*Wrapper, tea.Cmd) {
	if len(w.AppState.Columns()) == 0 {
		w.NotificationState.Add(state.LevelError, "Cannot add task: No columns exist. Create a column first with 'C'")
		return w, nil
	}
	w.FormState.FormTitle = ""
	w.FormState.FormDescription = ""
	w.FormState.FormLabelIDs = []int{}
	w.FormState.FormParentIDs = []int{}
	w.FormState.FormChildIDs = []int{}
	w.FormState.FormParentRefs = []*models.TaskReference{}
	w.FormState.FormChildRefs = []*models.TaskReference{}
	w.FormState.FormConfirm = true
	w.FormState.EditingTaskID = 0

	// Calculate description height (will be dynamic in Phase 4)
	descriptionLines := 10

	w.FormState.TicketForm = huhforms.CreateTicketForm(
		&w.FormState.FormTitle,
		&w.FormState.FormDescription,
		&w.FormState.FormConfirm,
		descriptionLines,
	).WithTheme(huhforms.CreatePasoTheme(w.Config.ColorScheme))
	w.FormState.SnapshotTicketFormInitialValues() // Snapshot for change detection
	w.UiState.SetMode(state.TicketFormMode)
	return w, w.FormState.TicketForm.Init()
}

// handleEditTask initiates editing the selected task.
func (w *Wrapper) handleEditTask() (*Wrapper, tea.Cmd) {
	ops := modelops.New(w.Model)
	task := ops.GetCurrentTask()
	if task == nil {
		w.NotificationState.Add(state.LevelError, "No task selected to edit")
		return w, nil
	}

	ctx, cancel := w.DbContext()
	defer cancel()
	taskDetail, err := w.Repo.GetTaskDetail(ctx, task.ID)
	if err != nil {
		w.HandleDBError(err, "Loading task details")
		return w, nil
	}

	w.FormState.FormTitle = taskDetail.Title
	w.FormState.FormDescription = taskDetail.Description
	w.FormState.FormLabelIDs = make([]int, len(taskDetail.Labels))
	for i, label := range taskDetail.Labels {
		w.FormState.FormLabelIDs[i] = label.ID
	}

	// Load parent relationships
	w.FormState.FormParentIDs = make([]int, len(taskDetail.ParentTasks))
	w.FormState.FormParentRefs = taskDetail.ParentTasks
	for i, parent := range taskDetail.ParentTasks {
		w.FormState.FormParentIDs[i] = parent.ID
	}

	// Load child relationships
	w.FormState.FormChildIDs = make([]int, len(taskDetail.ChildTasks))
	w.FormState.FormChildRefs = taskDetail.ChildTasks
	for i, child := range taskDetail.ChildTasks {
		w.FormState.FormChildIDs[i] = child.ID
	}

	// Load timestamps, type, and priority for metadata display
	w.FormState.FormCreatedAt = taskDetail.CreatedAt
	w.FormState.FormUpdatedAt = taskDetail.UpdatedAt
	w.FormState.FormTypeDescription = taskDetail.TypeDescription
	w.FormState.FormPriorityDescription = taskDetail.PriorityDescription
	w.FormState.FormPriorityColor = taskDetail.PriorityColor

	w.FormState.FormConfirm = true
	w.FormState.EditingTaskID = task.ID

	// Calculate description height (will be dynamic in Phase 4)
	descriptionLines := 10

	w.FormState.TicketForm = huhforms.CreateTicketForm(
		&w.FormState.FormTitle,
		&w.FormState.FormDescription,
		&w.FormState.FormConfirm,
		descriptionLines,
	).WithTheme(huhforms.CreatePasoTheme(w.Config.ColorScheme))
	w.FormState.SnapshotTicketFormInitialValues() // Snapshot for change detection
	w.UiState.SetMode(state.TicketFormMode)
	return w, w.FormState.TicketForm.Init()
}

// handleDeleteTask initiates task deletion confirmation.
func (w *Wrapper) handleDeleteTask() (*Wrapper, tea.Cmd) {
	ops := modelops.New(w.Model)
	if ops.GetCurrentTask() == nil {
		w.NotificationState.Add(state.LevelError, "No task selected to delete")
		return w, nil
	}
	w.UiState.SetMode(state.DeleteConfirmMode)
	return w, nil
}

// handleMoveTaskRight moves the task to the next column.
func (w *Wrapper) handleMoveTaskRight() (*Wrapper, tea.Cmd) {
	ops := modelops.New(w.Model)
	if ops.GetCurrentTask() != nil {
		ops.MoveTaskRight()
	}
	return w, nil
}

// handleMoveTaskLeft moves the task to the previous column.
func (w *Wrapper) handleMoveTaskLeft() (*Wrapper, tea.Cmd) {
	ops := modelops.New(w.Model)
	if ops.GetCurrentTask() != nil {
		ops.MoveTaskLeft()
	}
	return w, nil
}

// handleMoveTaskUp moves the task up within its column.
func (w *Wrapper) handleMoveTaskUp() (*Wrapper, tea.Cmd) {
	ops := modelops.New(w.Model)
	if ops.GetCurrentTask() != nil {
		ops.MoveTaskUp()
	}
	return w, nil
}

// handleMoveTaskDown moves the task down within its column.
func (w *Wrapper) handleMoveTaskDown() (*Wrapper, tea.Cmd) {
	ops := modelops.New(w.Model)
	if ops.GetCurrentTask() != nil {
		ops.MoveTaskDown()
	}
	return w, nil
}

// handleCreateColumn initiates column creation.
func (w *Wrapper) handleCreateColumn() (*Wrapper, tea.Cmd) {
	w.UiState.SetMode(state.AddColumnMode)
	w.InputState.Prompt = "New column name:"
	w.InputState.Buffer = ""
	return w, nil
}

// handleRenameColumn initiates column renaming.
func (w *Wrapper) handleRenameColumn() (*Wrapper, tea.Cmd) {
	ops := modelops.New(w.Model)
	column := ops.GetCurrentColumn()
	if column == nil {
		w.NotificationState.Add(state.LevelError, "No column selected to rename")
		return w, nil
	}
	w.UiState.SetMode(state.EditColumnMode)
	w.InputState.Buffer = column.Name
	w.InputState.Prompt = "Rename column:"
	w.InputState.SnapshotInitialBuffer() // Snapshot for change detection
	return w, nil
}

// handleDeleteColumn initiates column deletion confirmation.
func (w *Wrapper) handleDeleteColumn() (*Wrapper, tea.Cmd) {
	ops := modelops.New(w.Model)
	column := ops.GetCurrentColumn()
	if column == nil {
		w.NotificationState.Add(state.LevelError, "No column selected to delete")
		return w, nil
	}
	ctx, cancel := w.DbContext()
	defer cancel()
	taskCount, err := w.Repo.GetTaskCountByColumn(ctx, column.ID)
	if err != nil {
		slog.Error("Error getting task count", "error", err)
		w.NotificationState.Add(state.LevelError, "Error getting column info")
		return w, nil
	}
	w.InputState.DeleteColumnTaskCount = taskCount
	w.UiState.SetMode(state.DeleteColumnConfirmMode)
	return w, nil
}

// handlePrevProject switches to the previous project.
func (w *Wrapper) handlePrevProject() (*Wrapper, tea.Cmd) {
	ops := modelops.New(w.Model)
	if w.AppState.SelectedProject() > 0 {
		ops.SwitchToProject(w.AppState.SelectedProject() - 1)
	} else {
		w.NotificationState.Add(state.LevelInfo, "Already at the first project")
	}
	return w, nil
}

// handleNextProject switches to the next project.
func (w *Wrapper) handleNextProject() (*Wrapper, tea.Cmd) {
	ops := modelops.New(w.Model)
	if w.AppState.SelectedProject() < len(w.AppState.Projects())-1 {
		ops.SwitchToProject(w.AppState.SelectedProject() + 1)
	} else {
		w.NotificationState.Add(state.LevelInfo, "Already at the last project")
	}
	return w, nil
}

// handleCreateProject initiates project creation.
func (w *Wrapper) handleCreateProject() (*Wrapper, tea.Cmd) {
	w.FormState.FormProjectName = ""
	w.FormState.FormProjectDescription = ""
	w.FormState.FormProjectConfirm = true
	w.FormState.ProjectForm = huhforms.CreateProjectForm(
		&w.FormState.FormProjectName,
		&w.FormState.FormProjectDescription,
		&w.FormState.FormProjectConfirm,
	).WithTheme(huhforms.CreatePasoTheme(w.Config.ColorScheme))
	w.FormState.SnapshotProjectFormInitialValues() // Snapshot for change detection
	w.UiState.SetMode(state.ProjectFormMode)
	return w, w.FormState.ProjectForm.Init()
}
