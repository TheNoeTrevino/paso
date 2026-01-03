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
	m.UI.Notification.Clear()

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
	m.UIState.SetMode(state.HelpMode)
	return m, nil
}

func (m Model) handleNavigateLeft() (tea.Model, tea.Cmd) {
	if m.UIState.SelectedColumn() > 0 {
		m.UIState.SetSelectedColumn(m.UIState.SelectedColumn() - 1)
		m.UIState.SetSelectedTask(0)
		m.UIState.EnsureSelectionVisible(m.UIState.SelectedColumn())
	} else {
		m.UI.Notification.Add(state.LevelInfo, "Already at the first column")
	}
	return m, nil
}

func (m Model) handleNavigateRight() (tea.Model, tea.Cmd) {
	if m.UIState.SelectedColumn() < len(m.AppState.Columns())-1 {
		m.UIState.SetSelectedColumn(m.UIState.SelectedColumn() + 1)
		m.UIState.SetSelectedTask(0)
		m.UIState.EnsureSelectionVisible(m.UIState.SelectedColumn())
	} else {
		m.UI.Notification.Add(state.LevelInfo, "Already at the last column")
	}
	return m, nil
}

func (m Model) handleNavigateUp() (tea.Model, tea.Cmd) {
	if m.UI.ListView.IsListView() {
		if m.UI.ListView.SelectedRow() > 0 {
			m.UI.ListView.SetSelectedRow(m.UI.ListView.SelectedRow() - 1)

			listHeight := m.UIState.ContentHeight()
			const reservedHeight = 6
			visibleRows := max(listHeight-reservedHeight, 1)
			m.UI.ListView.EnsureRowVisible(visibleRows)
		} else {
			m.UI.Notification.Add(state.LevelInfo, "Already at the first task")
		}
		return m, nil
	}

	if m.UIState.SelectedTask() > 0 {
		m.UIState.SetSelectedTask(m.UIState.SelectedTask() - 1)

		if m.UIState.SelectedColumn() < len(m.AppState.Columns()) {
			currentCol := m.AppState.Columns()[m.UIState.SelectedColumn()]
			columnHeight := m.UIState.ContentHeight()
			const columnOverhead = 5
			maxTasksVisible := max((columnHeight-columnOverhead)/components.TaskCardHeight, 1)
			m.UIState.EnsureTaskVisible(currentCol.ID, m.UIState.SelectedTask(), maxTasksVisible)
		}
	} else {
		m.UI.Notification.Add(state.LevelInfo, "Already at the first task")
	}
	return m, nil
}

func (m Model) handleNavigateDown() (tea.Model, tea.Cmd) {
	if m.UI.ListView.IsListView() {
		rows := m.buildListViewRows()
		if m.UI.ListView.SelectedRow() < len(rows)-1 {
			m.UI.ListView.SetSelectedRow(m.UI.ListView.SelectedRow() + 1)

			listHeight := m.UIState.ContentHeight()
			const reservedHeight = 6
			visibleRows := max(listHeight-reservedHeight, 1)
			m.UI.ListView.EnsureRowVisible(visibleRows)
		} else if len(rows) > 0 {
			m.UI.Notification.Add(state.LevelInfo, "Already at the last task")
		}
		return m, nil
	}

	currentTasks := m.getCurrentTasks()
	if len(currentTasks) > 0 && m.UIState.SelectedTask() < len(currentTasks)-1 {
		m.UIState.SetSelectedTask(m.UIState.SelectedTask() + 1)

		if m.UIState.SelectedColumn() < len(m.AppState.Columns()) {
			currentCol := m.AppState.Columns()[m.UIState.SelectedColumn()]
			columnHeight := m.UIState.ContentHeight()
			const columnOverhead = 5
			maxTasksVisible := max((columnHeight-columnOverhead)/components.TaskCardHeight, 1)
			m.UIState.EnsureTaskVisible(currentCol.ID, m.UIState.SelectedTask(), maxTasksVisible)
		}
	} else if len(currentTasks) > 0 {
		m.UI.Notification.Add(state.LevelInfo, "Already at the last task")
	}
	return m, nil
}

func (m Model) handleScrollRight() (tea.Model, tea.Cmd) {
	if m.UIState.ViewportOffset()+m.UIState.ViewportSize() < len(m.AppState.Columns()) {
		m.UIState.SetViewportOffset(m.UIState.ViewportOffset() + 1)
		if m.UIState.SelectedColumn() < m.UIState.ViewportOffset() {
			m.UIState.SetSelectedColumn(m.UIState.ViewportOffset())
			m.UIState.SetSelectedTask(0)
		}
	} else {
		m.UI.Notification.Add(state.LevelInfo, "Already at the rightmost view")
	}
	return m, nil
}

func (m Model) handleScrollLeft() (tea.Model, tea.Cmd) {
	if m.UIState.ViewportOffset() > 0 {
		m.UIState.SetViewportOffset(m.UIState.ViewportOffset() - 1)
		if m.UIState.SelectedColumn() >= m.UIState.ViewportOffset()+m.UIState.ViewportSize() {
			m.UIState.SetSelectedColumn(m.UIState.ViewportOffset() + m.UIState.ViewportSize() - 1)
			m.UIState.SetSelectedTask(0)
		}
	} else {
		m.UI.Notification.Add(state.LevelInfo, "Already at the leftmost view")
	}
	return m, nil
}

func (m Model) handleAddTask() (tea.Model, tea.Cmd) {
	if len(m.AppState.Columns()) == 0 {
		m.UI.Notification.Add(state.LevelError, "Cannot add task: No columns exist. Create a column first with 'C'")
		return m, nil
	}
	m.Forms.Form.FormTitle = ""
	m.Forms.Form.FormDescription = ""
	m.Forms.Form.FormLabelIDs = []int{}
	m.Forms.Form.FormParentIDs = []int{}
	m.Forms.Form.FormChildIDs = []int{}
	m.Forms.Form.FormParentRefs = []*models.TaskReference{}
	m.Forms.Form.FormChildRefs = []*models.TaskReference{}
	m.Forms.Form.FormConfirm = true
	m.Forms.Form.EditingTaskID = 0

	// Calculate description lines based on current screen size
	descriptionLines := m.calculateDescriptionLines()

	m.Forms.Form.TaskForm = huhforms.CreateTaskForm(
		&m.Forms.Form.FormTitle,
		&m.Forms.Form.FormDescription,
		&m.Forms.Form.FormConfirm,
		descriptionLines,
	).WithTheme(huhforms.CreatePasoTheme(m.Config.ColorScheme))
	m.Forms.Form.SnapshotTaskFormInitialValues()
	m.UIState.SetMode(state.TicketFormMode)
	return m, m.Forms.Form.TaskForm.Init()
}

func (m Model) handleEditTask() (tea.Model, tea.Cmd) {
	task := m.getCurrentTask()
	if task == nil {
		m.UI.Notification.Add(state.LevelError, "No task selected to edit")
		return m, nil
	}

	ctx, cancel := m.DBContext()
	defer cancel()
	taskDetail, err := m.App.TaskService.GetTaskDetail(ctx, task.ID)
	if err != nil {
		m.HandleDBError(err, "Loading task details")
		return m, nil
	}

	m.Forms.Form.FormTitle = taskDetail.Title
	m.Forms.Form.FormDescription = taskDetail.Description
	m.Forms.Form.FormLabelIDs = make([]int, len(taskDetail.Labels))
	for i, label := range taskDetail.Labels {
		m.Forms.Form.FormLabelIDs[i] = label.ID
	}

	m.Forms.Form.FormParentIDs = make([]int, len(taskDetail.ParentTasks))
	m.Forms.Form.FormParentRefs = taskDetail.ParentTasks
	for i, parent := range taskDetail.ParentTasks {
		m.Forms.Form.FormParentIDs[i] = parent.ID
	}

	m.Forms.Form.FormChildIDs = make([]int, len(taskDetail.ChildTasks))
	m.Forms.Form.FormChildRefs = taskDetail.ChildTasks
	for i, child := range taskDetail.ChildTasks {
		m.Forms.Form.FormChildIDs[i] = child.ID
	}

	// Load comments for the task
	m.Forms.Form.FormComments = taskDetail.Comments
	m.Forms.Form.InitialFormComments = make([]*models.Comment, len(taskDetail.Comments))
	copy(m.Forms.Form.InitialFormComments, taskDetail.Comments)

	m.Forms.Form.FormCreatedAt = taskDetail.CreatedAt
	m.Forms.Form.FormUpdatedAt = taskDetail.UpdatedAt
	m.Forms.Form.FormTypeDescription = taskDetail.TypeDescription
	m.Forms.Form.FormPriorityDescription = taskDetail.PriorityDescription
	m.Forms.Form.FormPriorityColor = taskDetail.PriorityColor

	m.Forms.Form.FormConfirm = true
	m.Forms.Form.EditingTaskID = task.ID

	// Calculate description lines based on current screen size
	descriptionLines := m.calculateDescriptionLines()

	m.Forms.Form.TaskForm = huhforms.CreateTaskForm(
		&m.Forms.Form.FormTitle,
		&m.Forms.Form.FormDescription,
		&m.Forms.Form.FormConfirm,
		descriptionLines,
	).WithTheme(huhforms.CreatePasoTheme(m.Config.ColorScheme))
	m.Forms.Form.SnapshotTaskFormInitialValues()
	m.UIState.SetMode(state.TicketFormMode)
	return m, m.Forms.Form.TaskForm.Init()
}

func (m Model) handleDeleteTask() (tea.Model, tea.Cmd) {
	if m.getCurrentTask() == nil {
		m.UI.Notification.Add(state.LevelError, "No task selected to delete")
		return m, nil
	}
	m.UIState.SetMode(state.DeleteConfirmMode)
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
	m.Forms.Form.FormColumnName = ""
	m.Forms.Form.EditingColumnID = 0
	m.Forms.Form.ColumnForm = huhforms.CreateColumnForm(&m.Forms.Form.FormColumnName, false).WithTheme(huhforms.CreatePasoTheme(m.Config.ColorScheme))
	m.Forms.Form.SnapshotColumnFormInitialValues()
	m.UIState.SetMode(state.AddColumnFormMode)
	return m, m.Forms.Form.ColumnForm.Init()
}

func (m Model) handleRenameColumn() (tea.Model, tea.Cmd) {
	column := m.getCurrentColumn()
	if column == nil {
		m.UI.Notification.Add(state.LevelError, "No column selected to rename")
		return m, nil
	}
	m.Forms.Form.FormColumnName = column.Name
	m.Forms.Form.EditingColumnID = column.ID
	m.Forms.Form.ColumnForm = huhforms.CreateColumnForm(&m.Forms.Form.FormColumnName, true).WithTheme(huhforms.CreatePasoTheme(m.Config.ColorScheme))
	m.Forms.Form.SnapshotColumnFormInitialValues()
	m.UIState.SetMode(state.EditColumnFormMode)
	return m, m.Forms.Form.ColumnForm.Init()
}

func (m Model) handleDeleteColumn() (tea.Model, tea.Cmd) {
	column := m.getCurrentColumn()
	if column == nil {
		m.UI.Notification.Add(state.LevelError, "No column selected to delete")
		return m, nil
	}
	// Count tasks in the column from current state
	taskCount := len(m.AppState.Tasks()[column.ID])
	m.Forms.Input.DeleteColumnTaskCount = taskCount
	m.UIState.SetMode(state.DeleteColumnConfirmMode)
	return m, nil
}

func (m Model) handlePrevProject() (tea.Model, tea.Cmd) {
	if m.AppState.SelectedProject() > 0 {
		newIndex := m.AppState.SelectedProject() - 1
		slog.Info("navigating to previous project", "current_index", m.AppState.SelectedProject(), "new_index", newIndex)
		m.switchToProject(newIndex)
	} else {
		m.UI.Notification.Add(state.LevelInfo, "Already at the first project")
	}
	return m, nil
}

func (m Model) handleNextProject() (tea.Model, tea.Cmd) {
	if m.AppState.SelectedProject() < len(m.AppState.Projects())-1 {
		newIndex := m.AppState.SelectedProject() + 1
		slog.Info("navigating to next project", "current_index", m.AppState.SelectedProject(), "new_index", newIndex)
		m.switchToProject(newIndex)
	} else {
		m.UI.Notification.Add(state.LevelInfo, "Already at the last project")
	}
	return m, nil
}

func (m Model) handleCreateProject() (tea.Model, tea.Cmd) {
	m.Forms.Form.FormProjectName = ""
	m.Forms.Form.FormProjectDescription = ""
	m.Forms.Form.FormProjectConfirm = true
	m.Forms.Form.ProjectForm = huhforms.CreateProjectForm(
		&m.Forms.Form.FormProjectName,
		&m.Forms.Form.FormProjectDescription,
		&m.Forms.Form.FormProjectConfirm,
	).WithTheme(huhforms.CreatePasoTheme(m.Config.ColorScheme))
	m.Forms.Form.SnapshotProjectFormInitialValues()
	m.UIState.SetMode(state.ProjectFormMode)
	return m, m.Forms.Form.ProjectForm.Init()
}
