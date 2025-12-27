package tui

import (
	"log/slog"

	tea "charm.land/bubbletea/v2"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui/components"
	"github.com/thenoetrevino/paso/internal/tui/huhforms"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

func (m Model) handleNormalMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.NotificationState.Clear()

	key := msg.String()
	km := m.Config.KeyMappings

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

func (m Model) handleQuit() (tea.Model, tea.Cmd) {
	return m, tea.Quit
}

func (m Model) handleShowHelp() (tea.Model, tea.Cmd) {
	m.UiState.SetMode(state.HelpMode)
	return m, nil
}

func (m Model) handleNavigateLeft() (tea.Model, tea.Cmd) {
	if m.UiState.SelectedColumn() > 0 {
		m.UiState.SetSelectedColumn(m.UiState.SelectedColumn() - 1)
		m.UiState.SetSelectedTask(0)
		m.UiState.EnsureSelectionVisible(m.UiState.SelectedColumn())
	} else {
		m.NotificationState.Add(state.LevelInfo, "Already at the first column")
	}
	return m, nil
}

func (m Model) handleNavigateRight() (tea.Model, tea.Cmd) {
	if m.UiState.SelectedColumn() < len(m.AppState.Columns())-1 {
		m.UiState.SetSelectedColumn(m.UiState.SelectedColumn() + 1)
		m.UiState.SetSelectedTask(0)
		m.UiState.EnsureSelectionVisible(m.UiState.SelectedColumn())
	} else {
		m.NotificationState.Add(state.LevelInfo, "Already at the last column")
	}
	return m, nil
}

func (m Model) handleNavigateUp() (tea.Model, tea.Cmd) {
	if m.ListViewState.IsListView() {
		if m.ListViewState.SelectedRow() > 0 {
			m.ListViewState.SetSelectedRow(m.ListViewState.SelectedRow() - 1)

			listHeight := m.UiState.ContentHeight()
			const reservedHeight = 6
			visibleRows := max(listHeight-reservedHeight, 1)
			m.ListViewState.EnsureRowVisible(visibleRows)
		} else {
			m.NotificationState.Add(state.LevelInfo, "Already at the first task")
		}
		return m, nil
	}

	if m.UiState.SelectedTask() > 0 {
		m.UiState.SetSelectedTask(m.UiState.SelectedTask() - 1)

		if m.UiState.SelectedColumn() < len(m.AppState.Columns()) {
			currentCol := m.AppState.Columns()[m.UiState.SelectedColumn()]
			columnHeight := m.UiState.ContentHeight()
			const columnOverhead = 5
			maxTasksVisible := max((columnHeight-columnOverhead)/components.TaskCardHeight, 1)
			m.UiState.EnsureTaskVisible(currentCol.ID, m.UiState.SelectedTask(), maxTasksVisible)
		}
	} else {
		m.NotificationState.Add(state.LevelInfo, "Already at the first task")
	}
	return m, nil
}

func (m Model) handleNavigateDown() (tea.Model, tea.Cmd) {
	if m.ListViewState.IsListView() {
		rows := m.buildListViewRows()
		if m.ListViewState.SelectedRow() < len(rows)-1 {
			m.ListViewState.SetSelectedRow(m.ListViewState.SelectedRow() + 1)

			listHeight := m.UiState.ContentHeight()
			const reservedHeight = 6
			visibleRows := max(listHeight-reservedHeight, 1)
			m.ListViewState.EnsureRowVisible(visibleRows)
		} else if len(rows) > 0 {
			m.NotificationState.Add(state.LevelInfo, "Already at the last task")
		}
		return m, nil
	}

	currentTasks := m.getCurrentTasks()
	if len(currentTasks) > 0 && m.UiState.SelectedTask() < len(currentTasks)-1 {
		m.UiState.SetSelectedTask(m.UiState.SelectedTask() + 1)

		if m.UiState.SelectedColumn() < len(m.AppState.Columns()) {
			currentCol := m.AppState.Columns()[m.UiState.SelectedColumn()]
			columnHeight := m.UiState.ContentHeight()
			const columnOverhead = 5
			maxTasksVisible := max((columnHeight-columnOverhead)/components.TaskCardHeight, 1)
			m.UiState.EnsureTaskVisible(currentCol.ID, m.UiState.SelectedTask(), maxTasksVisible)
		}
	} else if len(currentTasks) > 0 {
		m.NotificationState.Add(state.LevelInfo, "Already at the last task")
	}
	return m, nil
}

func (m Model) handleScrollRight() (tea.Model, tea.Cmd) {
	if m.UiState.ViewportOffset()+m.UiState.ViewportSize() < len(m.AppState.Columns()) {
		m.UiState.SetViewportOffset(m.UiState.ViewportOffset() + 1)
		if m.UiState.SelectedColumn() < m.UiState.ViewportOffset() {
			m.UiState.SetSelectedColumn(m.UiState.ViewportOffset())
			m.UiState.SetSelectedTask(0)
		}
	} else {
		m.NotificationState.Add(state.LevelInfo, "Already at the rightmost view")
	}
	return m, nil
}

func (m Model) handleScrollLeft() (tea.Model, tea.Cmd) {
	if m.UiState.ViewportOffset() > 0 {
		m.UiState.SetViewportOffset(m.UiState.ViewportOffset() - 1)
		if m.UiState.SelectedColumn() >= m.UiState.ViewportOffset()+m.UiState.ViewportSize() {
			m.UiState.SetSelectedColumn(m.UiState.ViewportOffset() + m.UiState.ViewportSize() - 1)
			m.UiState.SetSelectedTask(0)
		}
	} else {
		m.NotificationState.Add(state.LevelInfo, "Already at the leftmost view")
	}
	return m, nil
}

func (m Model) handleAddTask() (tea.Model, tea.Cmd) {
	if len(m.AppState.Columns()) == 0 {
		m.NotificationState.Add(state.LevelError, "Cannot add task: No columns exist. Create a column first with 'C'")
		return m, nil
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

	descriptionLines := 10

	m.FormState.TicketForm = huhforms.CreateTicketForm(
		&m.FormState.FormTitle,
		&m.FormState.FormDescription,
		&m.FormState.FormConfirm,
		descriptionLines,
	).WithTheme(huhforms.CreatePasoTheme(m.Config.ColorScheme))
	m.FormState.SnapshotTicketFormInitialValues()
	m.UiState.SetMode(state.TicketFormMode)
	return m, m.FormState.TicketForm.Init()
}

func (m Model) handleEditTask() (tea.Model, tea.Cmd) {
	task := m.getCurrentTask()
	if task == nil {
		m.NotificationState.Add(state.LevelError, "No task selected to edit")
		return m, nil
	}

	ctx, cancel := m.DbContext()
	defer cancel()
	taskDetail, err := m.App.TaskService.GetTaskDetail(ctx, task.ID)
	if err != nil {
		m.HandleDBError(err, "Loading task details")
		return m, nil
	}

	m.FormState.FormTitle = taskDetail.Title
	m.FormState.FormDescription = taskDetail.Description
	m.FormState.FormLabelIDs = make([]int, len(taskDetail.Labels))
	for i, label := range taskDetail.Labels {
		m.FormState.FormLabelIDs[i] = label.ID
	}

	m.FormState.FormParentIDs = make([]int, len(taskDetail.ParentTasks))
	m.FormState.FormParentRefs = taskDetail.ParentTasks
	for i, parent := range taskDetail.ParentTasks {
		m.FormState.FormParentIDs[i] = parent.ID
	}

	m.FormState.FormChildIDs = make([]int, len(taskDetail.ChildTasks))
	m.FormState.FormChildRefs = taskDetail.ChildTasks
	for i, child := range taskDetail.ChildTasks {
		m.FormState.FormChildIDs[i] = child.ID
	}

	// Load comments for the task
	m.FormState.FormComments = taskDetail.Comments
	m.FormState.InitialFormComments = make([]*models.Comment, len(taskDetail.Comments))
	copy(m.FormState.InitialFormComments, taskDetail.Comments)

	m.FormState.FormCreatedAt = taskDetail.CreatedAt
	m.FormState.FormUpdatedAt = taskDetail.UpdatedAt
	m.FormState.FormTypeDescription = taskDetail.TypeDescription
	m.FormState.FormPriorityDescription = taskDetail.PriorityDescription
	m.FormState.FormPriorityColor = taskDetail.PriorityColor

	m.FormState.FormConfirm = true
	m.FormState.EditingTaskID = task.ID

	descriptionLines := 10

	m.FormState.TicketForm = huhforms.CreateTicketForm(
		&m.FormState.FormTitle,
		&m.FormState.FormDescription,
		&m.FormState.FormConfirm,
		descriptionLines,
	).WithTheme(huhforms.CreatePasoTheme(m.Config.ColorScheme))
	m.FormState.SnapshotTicketFormInitialValues()
	m.UiState.SetMode(state.TicketFormMode)
	return m, m.FormState.TicketForm.Init()
}

func (m Model) handleDeleteTask() (tea.Model, tea.Cmd) {
	if m.getCurrentTask() == nil {
		m.NotificationState.Add(state.LevelError, "No task selected to delete")
		return m, nil
	}
	m.UiState.SetMode(state.DeleteConfirmMode)
	return m, nil
}

func (m Model) handleMoveTaskRight() (tea.Model, tea.Cmd) {
	if m.getCurrentTask() != nil {
		m.moveTaskRight()
	}
	return m, nil
}

func (m Model) handleMoveTaskLeft() (tea.Model, tea.Cmd) {
	if m.getCurrentTask() != nil {
		m.moveTaskLeft()
	}
	return m, nil
}

func (m Model) handleMoveTaskUp() (tea.Model, tea.Cmd) {
	if m.getCurrentTask() != nil {
		m.moveTaskUp()
	}
	return m, nil
}

func (m Model) handleMoveTaskDown() (tea.Model, tea.Cmd) {
	if m.getCurrentTask() != nil {
		m.moveTaskDown()
	}
	return m, nil
}

func (m Model) handleCreateColumn() (tea.Model, tea.Cmd) {
	m.FormState.FormColumnName = ""
	m.FormState.EditingColumnID = 0
	m.FormState.ColumnForm = huhforms.CreateColumnForm(&m.FormState.FormColumnName, false).WithTheme(huhforms.CreatePasoTheme(m.Config.ColorScheme))
	m.FormState.SnapshotColumnFormInitialValues()
	m.UiState.SetMode(state.AddColumnFormMode)
	return m, m.FormState.ColumnForm.Init()
}

func (m Model) handleRenameColumn() (tea.Model, tea.Cmd) {
	column := m.getCurrentColumn()
	if column == nil {
		m.NotificationState.Add(state.LevelError, "No column selected to rename")
		return m, nil
	}
	m.FormState.FormColumnName = column.Name
	m.FormState.EditingColumnID = column.ID
	m.FormState.ColumnForm = huhforms.CreateColumnForm(&m.FormState.FormColumnName, true).WithTheme(huhforms.CreatePasoTheme(m.Config.ColorScheme))
	m.FormState.SnapshotColumnFormInitialValues()
	m.UiState.SetMode(state.EditColumnFormMode)
	return m, m.FormState.ColumnForm.Init()
}

func (m Model) handleDeleteColumn() (tea.Model, tea.Cmd) {
	column := m.getCurrentColumn()
	if column == nil {
		m.NotificationState.Add(state.LevelError, "No column selected to delete")
		return m, nil
	}
	// Count tasks in the column from current state
	taskCount := len(m.AppState.Tasks()[column.ID])
	m.InputState.DeleteColumnTaskCount = taskCount
	m.UiState.SetMode(state.DeleteColumnConfirmMode)
	return m, nil
}

func (m Model) handlePrevProject() (tea.Model, tea.Cmd) {
	if m.AppState.SelectedProject() > 0 {
		newIndex := m.AppState.SelectedProject() - 1
		slog.Info("navigating to previous project", "current_index", m.AppState.SelectedProject(), "new_index", newIndex)
		m.switchToProject(newIndex)
	} else {
		m.NotificationState.Add(state.LevelInfo, "Already at the first project")
	}
	return m, nil
}

func (m Model) handleNextProject() (tea.Model, tea.Cmd) {
	if m.AppState.SelectedProject() < len(m.AppState.Projects())-1 {
		newIndex := m.AppState.SelectedProject() + 1
		slog.Info("navigating to next project", "current_index", m.AppState.SelectedProject(), "new_index", newIndex)
		m.switchToProject(newIndex)
	} else {
		m.NotificationState.Add(state.LevelInfo, "Already at the last project")
	}
	return m, nil
}

func (m Model) handleCreateProject() (tea.Model, tea.Cmd) {
	m.FormState.FormProjectName = ""
	m.FormState.FormProjectDescription = ""
	m.FormState.FormProjectConfirm = true
	m.FormState.ProjectForm = huhforms.CreateProjectForm(
		&m.FormState.FormProjectName,
		&m.FormState.FormProjectDescription,
		&m.FormState.FormProjectConfirm,
	).WithTheme(huhforms.CreatePasoTheme(m.Config.ColorScheme))
	m.FormState.SnapshotProjectFormInitialValues()
	m.UiState.SetMode(state.ProjectFormMode)
	return m, m.FormState.ProjectForm.Init()
}
