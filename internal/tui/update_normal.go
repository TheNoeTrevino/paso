package tui

import (
	"context"
	"log/slog"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/services/task"
	"github.com/thenoetrevino/paso/internal/tui/commands"
	"github.com/thenoetrevino/paso/internal/tui/components"
	"github.com/thenoetrevino/paso/internal/tui/helpers"
	"github.com/thenoetrevino/paso/internal/tui/huhforms"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// taskDetailLoadTimeout is the maximum time allowed for loading task details
// when opening the edit form. Prevents indefinite hangs if the backend is slow.
const taskDetailLoadTimeout = 5 * time.Second

// loadTaskDetailForEditCmd creates a command that fetches task details and activities
// for editing. Runs asynchronously to prevent blocking the UI.
func loadTaskDetailForEditCmd(ctx context.Context, svc task.Service, taskID int) tea.Cmd {
	return func() tea.Msg {
		fetchCtx, cancel := context.WithTimeout(ctx, taskDetailLoadTimeout)
		defer cancel()

		taskDetail, err := svc.GetTaskDetail(fetchCtx, taskID)
		if err != nil {
			return taskDetailForEditError{err: err}
		}

		activities, err := svc.GetTaskActivities(fetchCtx, taskID)
		if err != nil {
			// Activities are optional, log but don't fail
			slog.Error("failed to load task activities", "error", err)
			activities = []models.ActivityItem{}
		}

		return taskDetailForEditMsg{
			taskID:     taskID,
			taskDetail: taskDetail,
			activities: activities,
		}
	}
}

func (m Model) handleNormalMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	km := m.Config.KeyMappings

	switch key {
	case km.General.Quit, "ctrl+c":
		return m.handleQuit()
	case km.General.ShowHelp:
		return m.handleShowHelp()
	case km.Tasks.AddTask:
		return m.handleAddTask()
	case km.Tasks.EditTask:
		return m.handleEditTask()
	case km.Tasks.DeleteTask:
		return m.handleDeleteTask()
	case km.Tasks.ViewTask:
		return m.handleEditTask()
	case km.Kanban.CreateColumn:
		return m.handleCreateColumn()
	case km.Kanban.RenameColumn:
		return m.handleRenameColumn()
	case km.Kanban.DeleteColumn:
		return m.handleDeleteColumn()
	case km.Navigation.ScrollViewportRight:
		return m.handleScrollRight()
	case km.Navigation.ScrollViewportLeft:
		return m.handleScrollLeft()
	case km.Navigation.MoveLeft, "left":
		return m.handleNavigateLeft()
	case km.Navigation.MoveRight, "right":
		return m.handleNavigateRight()
	case km.Navigation.MoveDown, "down":
		return m.handleNavigateDown()
	case km.Navigation.MoveUp, "up":
		return m.handleNavigateUp()
	case km.Kanban.MoveTaskRight:
		return m.handleMoveTaskRight()
	case km.Kanban.MoveTaskLeft:
		return m.handleMoveTaskLeft()
	case km.Kanban.MoveTaskUp:
		return m.handleMoveTaskUp()
	case km.Kanban.MoveTaskDown:
		return m.handleMoveTaskDown()
	case km.Navigation.PrevProject:
		return m.handlePrevProject()
	case km.Navigation.NextProject:
		return m.handleNextProject()
	case km.Projects.CreateProject:
		return m.handleCreateProject()
	case km.Projects.EditProject:
		return m.handleEditProject()
	case km.Projects.DeleteProject:
		return m.handleDeleteProject()
	case km.General.ToggleView:
		return m.handleToggleView()
	case km.Tasks.MoveTaskToProject:
		return m.handleMoveTaskToProject()
	case km.General.ConnectRemote:
		return m.handleConnectRemote()
	case km.Tasks.EditLabels:
		return m.handleQuickEditLabels()
	case km.Tasks.EditPriority:
		return m.handleQuickEditPriority()
	case km.Tasks.EditAssignee:
		return m.handleQuickEditAssignee()
	case km.Tasks.EditEstimate:
		return m.handleQuickEditEstimate()
	case km.Tasks.EditDueDate:
		return m.handleQuickEditDueDate()
	case km.Tasks.EditType:
		return m.handleQuickEditType()
	case km.General.Search:
		return m.handleEnterSearchChip()
	case km.General.FilterBar:
		return m.handleEnterFilterBar()
	}

	return m, nil
}

// handleEnterSearchChip jumps directly into the filter bar with the search chip in input mode.
func (m Model) handleEnterSearchChip() (tea.Model, tea.Cmd) {
	m.UI.Filter.IsActive = true
	m.UI.Filter.FocusedChip = state.FilterChipSearch
	m.UI.Filter.SearchInputActive = true
	m.UI.Filter.SearchQuery = ""
	m.UIState.Mode = state.FilterBarMode
	return m, nil
}

func (m Model) handleEnterFilterBar() (tea.Model, tea.Cmd) {
	m.UI.Filter.IsActive = true
	m.UI.Filter.FocusedChip = state.FilterChipSearch
	m.UIState.Mode = state.FilterBarMode
	return m, nil
}

func (m Model) handleQuit() (tea.Model, tea.Cmd) {
	return m, tea.Quit
}

func (m Model) handleShowHelp() (tea.Model, tea.Cmd) {
	m.UIState.Mode = state.HelpMode
	return m, nil
}

func (m Model) handleNavigateLeft() (tea.Model, tea.Cmd) {
	if m.UIState.SelectedColumn == 0 {
		return m, m.addNotification(state.LevelInfo, "Already at the first column")
	}
	m.UIState.SelectedColumn = m.UIState.SelectedColumn - 1
	m.UIState.SelectedTask = 0
	m.UIState.EnsureSelectionVisible(m.UIState.SelectedColumn)
	return m, m.triggerDetailPanelPrefetch()
}

func (m Model) handleNavigateRight() (tea.Model, tea.Cmd) {
	if m.UIState.SelectedColumn >= len(m.AppState.Columns())-1 {
		return m, m.addNotification(state.LevelInfo, "Already at the last column")
	}
	m.UIState.SelectedColumn = m.UIState.SelectedColumn + 1
	m.UIState.SelectedTask = 0
	m.UIState.EnsureSelectionVisible(m.UIState.SelectedColumn)
	return m, m.triggerDetailPanelPrefetch()
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
			return m, m.addNotification(state.LevelInfo, "Already at the first task")
		}
		return m, nil
	}

	if m.UIState.SelectedTask == 0 {
		return m, m.addNotification(state.LevelInfo, "Already at the first task")
	}

	m.UIState.SelectedTask = m.UIState.SelectedTask - 1
	m.ensureCurrentTaskVisible()
	return m, m.triggerDetailPanelPrefetch()
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
			return m, m.addNotification(state.LevelInfo, "Already at the last task")
		}
		return m, nil
	}

	currentTasks := m.getCurrentTasks()
	if len(currentTasks) == 0 || m.UIState.SelectedTask >= len(currentTasks)-1 {
		if len(currentTasks) > 0 {
			return m, m.addNotification(state.LevelInfo, "Already at the last task")
		}
		return m, nil
	}

	m.UIState.SelectedTask = m.UIState.SelectedTask + 1
	m.ensureCurrentTaskVisible()
	return m, m.triggerDetailPanelPrefetch()
}

func (m Model) handleScrollRight() (tea.Model, tea.Cmd) {
	if m.UIState.ViewportOffset+m.UIState.ViewportSize() >= len(m.AppState.Columns()) {
		return m, m.addNotification(state.LevelInfo, "Already at the rightmost view")
	}
	m.UIState.ViewportOffset = m.UIState.ViewportOffset + 1
	if m.UIState.SelectedColumn < m.UIState.ViewportOffset {
		m.UIState.SelectedColumn = m.UIState.ViewportOffset
		m.UIState.SelectedTask = 0
	}
	return m, nil
}

func (m Model) handleScrollLeft() (tea.Model, tea.Cmd) {
	if m.UIState.ViewportOffset == 0 {
		return m, m.addNotification(state.LevelInfo, "Already at the leftmost view")
	}
	m.UIState.ViewportOffset = m.UIState.ViewportOffset - 1
	if m.UIState.SelectedColumn >= m.UIState.ViewportOffset+m.UIState.ViewportSize() {
		m.UIState.SelectedColumn = m.UIState.ViewportOffset + m.UIState.ViewportSize() - 1
		m.UIState.SelectedTask = 0
	}
	return m, nil
}

func (m Model) handleAddTask() (tea.Model, tea.Cmd) {
	if len(m.AppState.Columns()) == 0 {
		return m, m.addNotification(state.LevelError, "Cannot add task: No columns exist. Create a column first with 'C'")
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
	m.UIState.Mode = state.TicketFormMode
	return m, m.Forms.Form.TaskForm.Init()
}

func (m Model) handleEditTask() (tea.Model, tea.Cmd) {
	task := m.getCurrentTask()
	if task == nil {
		return m, m.addNotification(state.LevelError, "No task selected to edit")
	}

	// Enter loading state and fetch task details asynchronously
	m.UIState.Mode = state.TicketFormLoadingMode
	m.LoadingGitInfo = true // Reuse this flag for the loading spinner
	m.SpinnerFrame = 0

	return m, tea.Batch(
		loadTaskDetailForEditCmd(m.Ctx, m.App.TaskService, task.ID),
		tickGitSpinner(),
	)
}

func (m Model) handleDeleteTask() (tea.Model, tea.Cmd) {
	if m.getCurrentTask() == nil {
		return m, m.addNotification(state.LevelError, "No task selected to delete")
	}
	m.UIState.Mode = state.DeleteConfirmMode
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
	m.UIState.Mode = state.AddColumnFormMode
	return m, m.Forms.Form.ColumnForm.Init()
}

func (m Model) handleRenameColumn() (tea.Model, tea.Cmd) {
	column := m.getCurrentColumn()
	if column == nil {
		return m, m.addNotification(state.LevelError, "No column selected to rename")
	}
	m.Forms.Form.FormColumnName = column.Name
	m.Forms.Form.EditingColumnID = column.ID
	m.Forms.Form.ColumnForm = huhforms.CreateColumnForm(&m.Forms.Form.FormColumnName, true).WithTheme(huhforms.CreatePasoTheme(m.Config.ColorScheme))
	m.Forms.Form.SnapshotColumnFormInitialValues()
	m.UIState.Mode = state.EditColumnFormMode
	return m, m.Forms.Form.ColumnForm.Init()
}

func (m Model) handleDeleteColumn() (tea.Model, tea.Cmd) {
	column := m.getCurrentColumn()
	if column == nil {
		return m, m.addNotification(state.LevelError, "No column selected to delete")
	}
	// Count tasks in the column from current state
	taskCount := len(m.AppState.Tasks()[column.ID])
	m.Forms.Input.DeleteColumnTaskCount = taskCount
	m.UIState.Mode = state.DeleteColumnConfirmMode
	return m, nil
}

func (m Model) handlePrevProject() (tea.Model, tea.Cmd) {
	if m.AppState.SelectedProject() == 0 {
		return m, m.addNotification(state.LevelInfo, "Already at the first project")
	}
	newIndex := m.AppState.SelectedProject() - 1
	slog.Info("navigating to previous project", "current_index", m.AppState.SelectedProject(), "new_index", newIndex)
	m.switchToProject(newIndex)
	return m, m.triggerDetailPanelPrefetch()
}

func (m Model) handleNextProject() (tea.Model, tea.Cmd) {
	if m.AppState.SelectedProject() >= len(m.AppState.Projects())-1 {
		return m, m.addNotification(state.LevelInfo, "Already at the last project")
	}
	newIndex := m.AppState.SelectedProject() + 1
	slog.Info("navigating to next project", "current_index", m.AppState.SelectedProject(), "new_index", newIndex)
	m.switchToProject(newIndex)
	return m, m.triggerDetailPanelPrefetch()
}

func (m Model) handleCreateProject() (tea.Model, tea.Cmd) {
	m.Forms.Form.Git.EditingProjectID = 0
	m.Forms.Form.FormProjectName = ""
	m.Forms.Form.FormProjectDescription = ""
	m.Forms.Form.FormProjectGitBranch = ""
	m.Forms.Form.FormProjectConfirm = true

	m.UIState.Mode = state.ProjectFormLoadingMode
	m.LoadingGitInfo = true
	m.LoadingBranches = true
	m.SpinnerFrame = 0
	return m, tea.Batch(
		loadGitDataForProjectFormCmd(m.Ctx, m.GitCache, false),
		tickGitSpinner(),
	)
}

func (m Model) handleEditProject() (tea.Model, tea.Cmd) {
	currentProject := m.AppState.GetCurrentProject()
	if currentProject == nil {
		return m, m.addNotification(state.LevelWarning, "No project selected")
	}

	m.Forms.Form.Git.EditingProjectID = currentProject.ID
	m.Forms.Form.FormProjectName = currentProject.Name
	m.Forms.Form.FormProjectDescription = currentProject.Description
	m.Forms.Form.FormProjectGitBranch = currentProject.GitBranch
	m.Forms.Form.FormProjectConfirm = true

	m.UIState.Mode = state.ProjectFormLoadingMode
	m.LoadingGitInfo = true
	m.LoadingBranches = true
	m.SpinnerFrame = 0
	return m, tea.Batch(
		loadGitDataForProjectFormCmd(m.Ctx, m.GitCache, true),
		tickGitSpinner(),
	)
}

func (m Model) handleDeleteProject() (tea.Model, tea.Cmd) {
	currentProject := m.AppState.GetCurrentProject()
	if currentProject == nil {
		return m, m.addNotification(state.LevelError, "No project selected to delete")
	}

	ctx, cancel := m.DBContext()
	defer cancel()
	taskCount, err := m.App.ProjectService.GetTaskCount(ctx, currentProject.ID)
	if err != nil {
		m.HandleDBError(err, "Checking project tasks")
		return m, nil
	}

	m.Forms.Input.DeleteProjectTaskCount = taskCount
	m.UIState.Mode = state.DeleteProjectConfirmMode
	return m, nil
}

// prefetchDebounceDelay is the delay before triggering a prefetch after navigation.
// This coalesces rapid navigation (e.g., holding arrow key) into a single prefetch.
const prefetchDebounceDelay = 75 * time.Millisecond

// triggerDetailPanelPrefetch returns a command to prefetch task details for the detail panel.
// Uses debouncing to avoid excessive API calls during rapid navigation.
// Only triggers if the detail panel should be shown and there's a valid task selection.
// Returns nil if panel is not visible or no task is selected.
//
// Synchronization: This method is safe from race conditions because Bubble Tea processes
// messages sequentially in a single goroutine. PrefetchDebounceSeq is only mutated within
// Update handlers, and the sequence comparison in prefetchDebounceMsg handling occurs in
// the same goroutine. Each navigation increments the counter, invalidating any pending
// debounce timers from previous navigations.
func (m *Model) triggerDetailPanelPrefetch() tea.Cmd {
	// Only prefetch if panel is visible
	if !m.UIState.ShouldShowDetailPanel() {
		return nil
	}

	// Get adjacent task IDs for prefetching
	adjacent := helpers.GetAdjacentTaskIDs(
		m.AppState.Columns(),
		m.AppState.Tasks(),
		m.UIState.SelectedColumn,
		m.UIState.SelectedTask,
	)

	// No current task selected
	if adjacent.Current == 0 {
		return nil
	}

	// Update the detail panel task ID immediately for UI responsiveness
	m.DetailPanelTaskID = adjacent.Current

	// Check if current task is already cached
	if m.DetailCache != nil && m.DetailCache.Has(adjacent.Current) {
		// Current task is cached, but still prefetch adjacent for smooth navigation
		m.DetailPanelLoading = false
	} else {
		// Current task not cached, show loading state
		m.DetailPanelLoading = true
	}

	// Increment debounce sequence and schedule delayed prefetch
	m.PrefetchDebounceSeq++
	seq := m.PrefetchDebounceSeq

	return tea.Tick(prefetchDebounceDelay, func(t time.Time) tea.Msg {
		return prefetchDebounceMsg{seq: seq}
	})
}

// executePrefetch performs the actual prefetch after debounce delay.
// This is called from the Update function when a prefetchDebounceMsg is received.
func (m *Model) executePrefetch() tea.Cmd {
	// Only prefetch if panel is still visible
	if !m.UIState.ShouldShowDetailPanel() {
		return nil
	}

	// Get adjacent task IDs for prefetching
	adjacent := helpers.GetAdjacentTaskIDs(
		m.AppState.Columns(),
		m.AppState.Tasks(),
		m.UIState.SelectedColumn,
		m.UIState.SelectedTask,
	)

	// No current task selected
	if adjacent.Current == 0 {
		return nil
	}

	// Get cached IDs to avoid refetching
	var cachedIDs []int
	if m.DetailCache != nil {
		cachedIDs = m.DetailCache.CachedIDs()
	}

	return commands.FetchTaskDetailsCmd(m.App.TaskService, adjacent, cachedIDs)
}

// ensureCurrentTaskVisible ensures the selected task is visible within the column viewport.
// Calculates the available space for tasks by subtracting UI overhead (borders, headers, indicators)
// and scrolls the column viewport if necessary to keep the selected task in view.
func (m *Model) ensureCurrentTaskVisible() {
	if m.UIState.SelectedColumn >= len(m.AppState.Columns()) {
		return
	}

	currentCol := m.AppState.Columns()[m.UIState.SelectedColumn]
	columnHeight := m.UIState.ContentHeight()

	const (
		columnBorderLines    = 2
		columnHeaderLines    = 2
		columnIndicatorLines = 1
		columnHeightOverhead = columnBorderLines + columnHeaderLines + columnIndicatorLines
	)

	maxTasksVisible := max((columnHeight-columnHeightOverhead)/components.TaskCardHeight, 1)
	m.UIState.EnsureTaskVisible(currentCol.ID, m.UIState.SelectedTask, maxTasksVisible)
}

func (m Model) handleQuickEditLabels() (tea.Model, tea.Cmd) {
	if !m.initLabelPicker(state.NormalMode) {
		return m, nil
	}
	m.UIState.Mode = state.LabelPickerMode
	return m, nil
}

func (m Model) handleQuickEditPriority() (tea.Model, tea.Cmd) {
	if !m.initPriorityPicker(state.NormalMode) {
		return m, nil
	}
	m.UIState.Mode = state.PriorityPickerMode
	return m, nil
}

func (m Model) handleQuickEditAssignee() (tea.Model, tea.Cmd) {
	if !m.initAssigneePicker(state.NormalMode) {
		return m, nil
	}
	m.UIState.Mode = state.AssigneePickerMode
	return m, nil
}

func (m Model) handleQuickEditEstimate() (tea.Model, tea.Cmd) {
	if !m.initEstimateInput(state.NormalMode) {
		return m, nil
	}
	m.UIState.Mode = state.EstimateInputMode
	return m, nil
}

func (m Model) handleQuickEditDueDate() (tea.Model, tea.Cmd) {
	if !m.initDatePicker(state.NormalMode) {
		return m, nil
	}
	m.UIState.Mode = state.DatePickerMode
	return m, nil
}

func (m Model) handleQuickEditType() (tea.Model, tea.Cmd) {
	if !m.initTypePicker(state.NormalMode) {
		return m, nil
	}
	m.UIState.Mode = state.TypePickerMode
	return m, nil
}
