package tui

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/thenoetrevino/paso/internal/app"
	"github.com/thenoetrevino/paso/internal/config"
	"github.com/thenoetrevino/paso/internal/database"
	"github.com/thenoetrevino/paso/internal/events"
	"github.com/thenoetrevino/paso/internal/git"
	"github.com/thenoetrevino/paso/internal/models"
	projectsvc "github.com/thenoetrevino/paso/internal/services/project"
	tasksvc "github.com/thenoetrevino/paso/internal/services/task"
	"github.com/thenoetrevino/paso/internal/tui/components"
	"github.com/thenoetrevino/paso/internal/tui/helpers"
	"github.com/thenoetrevino/paso/internal/tui/renderers"
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
	Ctx                 context.Context // Application context for cancellation and timeouts
	App                 *app.App        // Application container with services
	DB                  *sql.DB         // Database connection for switching
	Config              *config.Config
	AppState            *state.AppState
	UIState             *state.UIState
	Pickers             *state.PickerStates         // Grouped picker states (Label, Parent, Child, Priority, Type, RelationType, Status)
	Forms               *state.FormStates           // Grouped form states (Form, Input, Comment)
	UI                  *state.UIElements           // Grouped UI element states (Notification, Search, ListView)
	ConnectionState     *state.ConnectionState      // Connection status to daemon
	EventClient         events.EventPublisher       // Connection to daemon for live updates
	EventChan           <-chan events.Event         // Channel for receiving events
	NotifyChan          chan events.NotificationMsg // Channel for user-facing notifications from events
	SubscriptionStarted bool                        // Track if we've started listening
	DatabasePicker      *state.DatabasePickerState  // Database picker state (for Ctrl+H)
	CurrentDBType       database.DatabaseType       // Current database type (SQLite or PostgreSQL)
	CurrentDBName       string                      // Friendly name of current database (e.g., "Local", "Production")
	LoadingGitInfo      bool                        // True when fetching git info asynchronously
	LoadingBranches     bool                        // True when fetching git branches asynchronously
	SpinnerFrame        int                         // Current frame for git loading spinner
	GitCache            *git.Cache                  // Git information cache
	DetailCache         *state.TaskDetailCache      // LRU cache for task details (panel prefetch)
	DetailPanelLoading  bool                        // True when fetching task detail for panel
	DetailPanelTaskID   int                         // ID of task currently shown in detail panel
	PrefetchDebounceSeq int                         // Sequence number for debouncing prefetch requests
	InitialPrefetchDone bool                        // True after first detail panel prefetch is triggered
}

// InitialModel creates and initializes the TUI model with data from the database
func InitialModel(ctx context.Context, application *app.App, cfg *config.Config, eventClient events.EventPublisher, db *sql.DB) Model {
	// Create child context with timeout for initial loading
	loadCtx, cancel := context.WithTimeout(ctx, timeoutInitialLoad)
	defer cancel()

	projects, err := application.ProjectService.GetAllProjects(loadCtx)
	if err != nil {
		slog.Error("failed to loading projects", "error", err)
		projects = []*models.Project{}
	}

	var currentProjectID int
	gitInfo := git.DetectGitInfo(loadCtx)

	if gitInfo.IsRepo && gitInfo.IsValidForAssociation() {
		branchProject, err := application.ProjectService.GetProjectByGitBranch(loadCtx, gitInfo.CurrentBranch)
		if err != nil {
			slog.Error("failed to get project by git branch", "branch", gitInfo.CurrentBranch, "error", err)
		}

		if branchProject != nil {
			branchExists, err := git.BranchExists(loadCtx, gitInfo.CurrentBranch)
			if err != nil {
				slog.Error("failed to verify branch existence",
					"branch", gitInfo.CurrentBranch,
					"error", err)
			}

			if branchExists {
				currentProjectID = branchProject.ID
				slog.Info("auto-selected project based on git branch",
					"project_id", currentProjectID,
					"branch", gitInfo.CurrentBranch,
					"project_name", branchProject.Name)
			} else {
				slog.Warn("project associated with deleted branch, skipping auto-selection",
					"project_name", branchProject.Name,
					"deleted_branch", gitInfo.CurrentBranch)
			}
		} else {
			slog.Info("current git branch has no associated project, falling back to default",
				"branch", gitInfo.CurrentBranch)
		}
	}

	if !gitInfo.IsRepo && currentProjectID == 0 {
		slog.Info("not in a git repository, using default project")
	} else if gitInfo.IsRepo && !gitInfo.IsValidForAssociation() && currentProjectID == 0 {
		slog.Info("git repository state invalid for project association, using default project",
			"is_detached", gitInfo.IsDetached,
			"current_branch", gitInfo.CurrentBranch,
			"has_commits", gitInfo.HasCommits)
	}

	if currentProjectID == 0 && len(projects) > 0 {
		currentProjectID = projects[0].ID
		slog.Info("selected default project",
			"project_id", currentProjectID,
			"project_name", projects[0].Name)
	}

	if currentProjectID == 0 && len(projects) == 0 {
		slog.Warn("no projects available, TUI may not function correctly")
	}

	columns, err := application.ColumnService.GetColumnsByProject(loadCtx, currentProjectID)
	if err != nil {
		slog.Error("failed to loading columns", "error", err)
		columns = []*models.Column{}
	}

	tasks, err := application.TaskService.GetTaskSummariesByProject(loadCtx, currentProjectID)
	if err != nil {
		slog.Error("failed to loading tasks for project", "project_id", currentProjectID, "error", err)
		tasks = make(map[int][]*models.TaskSummary)
	}

	labels, err := application.LabelService.GetLabelsByProject(loadCtx, currentProjectID)
	if err != nil {
		slog.Error("failed to loading labels", "error", err)
		labels = []*models.Label{}
	}

	selectedProjectIndex := 0
	for i, p := range projects {
		if p.ID == currentProjectID {
			selectedProjectIndex = i
			break
		}
	}

	appState := state.NewAppState(projects, selectedProjectIndex, columns, tasks, labels)
	uiState := state.NewUIState()
	pickerStates := state.NewPickerStates()
	formStates := state.NewFormStates()
	uiElements := state.NewUIElements()

	tipGenerator := helpers.NewTipGenerator(&cfg.KeyMappings)
	uiElements.CurrentTip = tipGenerator.SelectRandom()

	initialStatus := state.Disconnected
	if eventClient != nil {
		initialStatus = state.Connected
	}
	connectionState := state.NewConnectionState(initialStatus)

	components.InitStyles(cfg.ColorScheme)
	renderers.InitDatePickerStyles()

	// Create notification channel for events client messages
	notifyChan := make(chan events.NotificationMsg, 10)

	// Initialize git cache with default TTL
	gitCache := git.NewCache(git.DefaultCacheTTL)

	// Initialize task detail cache for panel prefetching
	detailCache := state.NewTaskDetailCache(state.DefaultCacheSize)

	// Start listening for events from daemon (optional - daemon may not be running)
	var eventChan <-chan events.Event
	if eventClient != nil {
		var err error
		eventChan, err = eventClient.Listen(ctx)
		if err != nil {
			slog.Error("failed to listen for events", "error", err)
		}

		// Subscribe to current project's events
		if currentProjectID > 0 {
			if err := eventClient.Subscribe(currentProjectID); err != nil {
				slog.Error("error subscribing to project events", "project_id", currentProjectID, "error", err)
			}
		}

		// Set up notification callback for user-facing messages from events client
		eventClient.SetNotifyFunc(func(level, message string) {
			// Send notification to channel (non-blocking)
			select {
			case notifyChan <- events.NotificationMsg{Level: level, Message: message}:
			default:
				slog.Warn("notification channel full, dropping message", "level", level, "message", message)
			}
		})
	} else {
		// No event client available - create a closed channel to avoid nil channel issues
		ch := make(chan events.Event)
		close(ch)
		eventChan = ch
		slog.Debug("event client is nil, continuing without live updates")
	}

	return Model{
		Ctx:                 ctx,
		App:                 application,
		DB:                  db,
		Config:              cfg,
		AppState:            appState,
		UIState:             uiState,
		Pickers:             pickerStates,
		Forms:               formStates,
		UI:                  uiElements,
		ConnectionState:     connectionState,
		EventClient:         eventClient,
		EventChan:           eventChan,
		NotifyChan:          notifyChan,
		SubscriptionStarted: false,
		DatabasePicker:      state.NewDatabasePickerState(),
		CurrentDBType:       database.SQLite,
		CurrentDBName:       "Local",
		GitCache:            gitCache,
		DetailCache:         detailCache,
	}
}

// withTimeout creates a child context with appropriate timeout for operation type
func (m *Model) withTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(m.Ctx, timeout)
}

// DBContext creates a context for database operations with 30s timeout
func (m *Model) DBContext() (context.Context, context.CancelFunc) {
	return m.withTimeout(timeoutDB)
}

// listenForNotifications returns a tea.Cmd that listens for notification messages
// from the events client and sends them to the Update loop
func (m *Model) listenForNotifications() tea.Cmd {
	return func() tea.Msg {
		select {
		case msg := <-m.NotifyChan:
			return msg
		case <-m.Ctx.Done():
			return nil
		}
	}
}

// UIContext creates a context for UI operations with 10s timeout
func (m *Model) UIContext() (context.Context, context.CancelFunc) {
	return m.withTimeout(timeoutUI)
}

// HandleDBError handles database errors with context-aware messages
// It distinguishes between cancellation, timeout, and other errors,
// providing appropriate user feedback for each case.
// Returns a tea.Cmd that auto-dismisses the notification, or nil if no notification was added.
func (m *Model) HandleDBError(err error, operation string) tea.Cmd {
	if err == nil {
		return nil
	}

	if errors.Is(err, context.Canceled) {
		// Operation was cancelled - this is expected during shutdown
		// Don't show error notification, just log for debugging
		slog.Info("operation cancelled by user", "operation", operation)
		return nil
	}

	if errors.Is(err, context.DeadlineExceeded) {
		// Operation timed out - show user-friendly message
		return m.addNotification(state.LevelError, fmt.Sprintf("%s timed out. Please try again.", operation))
	}

	// Other errors - show detailed error message
	return m.addNotification(state.LevelError, fmt.Sprintf("%s failed: %v", operation, err))
}

// notificationDuration is how long notifications are visible before auto-dismissing
const notificationDuration = 1 * time.Second

// addNotification adds a notification and returns a tea.Cmd that will fire
// a NotificationExpiredMsg after the notification duration elapses.
// Returns nil if the notification was suppressed by debouncing.
func (m *Model) addNotification(level state.NotificationLevel, message string) tea.Cmd {
	id := m.UI.Notification.Add(level, message)
	if id == 0 {
		return nil
	}
	return tea.Tick(notificationDuration, func(t time.Time) tea.Msg {
		return NotificationExpiredMsg{ID: id}
	})
}

// Init initializes the Bubble Tea application
// Required by tea.Model interface
func (m Model) Init() tea.Cmd {
	return m.listenForNotifications()
}

// getCurrentTasks returns the task summaries for the currently selected column
// Returns an empty slice if the column has no tasks
func (m Model) getCurrentTasks() []*models.TaskSummary {
	if len(m.AppState.Columns()) == 0 {
		return []*models.TaskSummary{}
	}
	if m.UIState.SelectedColumn >= len(m.AppState.Columns()) {
		return []*models.TaskSummary{}
	}
	currentCol := m.AppState.Columns()[m.UIState.SelectedColumn]
	tasks := m.AppState.Tasks()[currentCol.ID]
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
	if m.UIState.SelectedTask >= len(tasks) {
		return nil
	}
	return tasks[m.UIState.SelectedTask]
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

	if len(tasks) == 0 || m.UIState.SelectedTask >= len(tasks) {
		return
	}

	// Remove the task at selectedTask index
	m.AppState.Tasks()[currentCol.ID] = append(tasks[:m.UIState.SelectedTask], tasks[m.UIState.SelectedTask+1:]...)

	// Adjust selectedTask if we removed the last task
	if m.UIState.SelectedTask >= len(m.AppState.Tasks()[currentCol.ID]) && m.UIState.SelectedTask > 0 {
		m.UIState.SelectedTask = m.UIState.SelectedTask - 1
	}
}

// getCurrentColumn returns the currently selected column
// Returns nil if there are no columns
func (m Model) getCurrentColumn() *models.Column {
	if len(m.AppState.Columns()) == 0 {
		return nil
	}
	selectedIdx := m.UIState.SelectedColumn
	if selectedIdx < 0 || selectedIdx >= len(m.AppState.Columns()) {
		return nil
	}
	return m.AppState.Columns()[selectedIdx]
}

// getTasksForColumn returns tasks for a specific column ID with safe map access.
// Returns an empty slice if the column ID doesn't exist in the tasks map.
func (m Model) getTasksForColumn(columnID int) []*models.TaskSummary {
	tasks, ok := m.AppState.Tasks()[columnID]
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
	columns := m.AppState.Columns()
	selectedCol := m.UIState.SelectedColumn

	if len(columns) == 0 || selectedCol >= len(columns) {
		return
	}

	// Remove the column at selectedColumn index
	m.AppState.SetColumns(append(columns[:selectedCol], columns[selectedCol+1:]...))

	// Adjust selectedColumn if we removed the last column
	if selectedCol >= len(m.AppState.Columns()) && selectedCol > 0 {
		m.UIState.SelectedColumn = selectedCol - 1
	}

	// Reset task selection
	m.UIState.SelectedTask = 0

	// Adjust viewportOffset using UIState helper
	m.UIState.AdjustViewportAfterColumnRemoval(m.UIState.SelectedColumn, len(m.AppState.Columns()))
}

// removeCurrentProject removes the currently selected project from local state
// and navigates to an adjacent project, or creates a default project if none remain.
// Returns a tea.Cmd if a notification was generated, or nil otherwise.
func (m *Model) removeCurrentProject() tea.Cmd {
	projects := m.AppState.Projects()
	selectedIdx := m.AppState.SelectedProject()

	if selectedIdx < 0 || selectedIdx >= len(projects) {
		return nil
	}

	// Remove project from slice
	newProjects := append(projects[:selectedIdx], projects[selectedIdx+1:]...)
	m.AppState.SetProjects(newProjects)

	// Clear detail cache since project context changed
	if m.DetailCache != nil {
		m.DetailCache.Clear()
	}

	if len(newProjects) == 0 {
		// Create default project
		return m.createDefaultProject()
	}

	// Navigate to appropriate project
	newIdx := selectedIdx
	if newIdx >= len(newProjects) {
		newIdx = len(newProjects) - 1
	}

	m.switchToProject(newIdx)
	return nil
}

// createDefaultProject creates a "Default" project when no projects exist.
// Returns a tea.Cmd that auto-dismisses the notification, or nil if no notification was added.
func (m *Model) createDefaultProject() tea.Cmd {
	ctx, cancel := m.DBContext()
	defer cancel()

	newProject, err := m.App.ProjectService.CreateProject(ctx, projectsvc.CreateProjectRequest{
		Name:        "Default",
		Description: "Default project",
	})
	if err != nil {
		slog.Error("failed to create default project", "error", err)
		// Clear state anyway
		m.AppState.SetColumns([]*models.Column{})
		m.AppState.SetTasks(make(map[int][]*models.TaskSummary))
		m.AppState.SetLabels([]*models.Label{})
		m.UIState.ResetSelection()
		return m.addNotification(state.LevelError, "Failed to create default project")
	}

	// Add to projects and switch to it
	m.AppState.SetProjects([]*models.Project{newProject})
	m.switchToProject(0)
	return m.addNotification(state.LevelInfo, "Created default project")
}

// moveTaskRight moves the currently selected task to the next column (right)
// Updates both the local state and the database using the linked list structure
// The selection follows the moved task and the viewport scrolls if needed.
// Returns a tea.Cmd if a notification was generated, or nil otherwise.
func (m *Model) moveTaskRight() tea.Cmd {
	// Get the current task
	task := m.getCurrentTask()
	if task == nil {
		return nil
	}

	// Check if there's a next column using the linked list
	currentCol := m.getCurrentColumn()
	if currentCol == nil {
		return nil
	}
	if currentCol.NextID == nil {
		// Already at last column - show notification
		return m.addNotification(state.LevelInfo, "There are no more columns to move to.")
	}

	// Use the new database function to move task
	ctx, cancel := m.UIContext()
	defer cancel()
	err := m.App.TaskService.MoveTaskToNextColumn(ctx, task.ID)
	if err != nil {
		slog.Error("failed to moving task to next column", "error", err)
		if err != tasksvc.ErrAlreadyLastColumn {
			return m.addNotification(state.LevelError, "Failed to move task to next column")
		}
		return nil
	}

	// Update local state: remove from current column
	tasks := m.AppState.Tasks()[currentCol.ID]
	m.AppState.Tasks()[currentCol.ID] = append(tasks[:m.UIState.SelectedTask], tasks[m.UIState.SelectedTask+1:]...)

	// Find the next column and add task there
	nextColID := *currentCol.NextID
	newPosition := len(m.AppState.Tasks()[nextColID])
	task.ColumnID = nextColID
	task.Position = newPosition
	m.AppState.Tasks()[nextColID] = append(m.AppState.Tasks()[nextColID], task)

	// Move selection to follow the task
	m.UIState.SelectedColumn = m.UIState.SelectedColumn + 1
	m.UIState.SelectedTask = newPosition

	// Ensure the moved task is visible (auto-scroll viewport if needed)
	if m.UIState.SelectedColumn >= m.UIState.ViewportOffset+m.UIState.ViewportSize() {
		m.UIState.ViewportOffset = m.UIState.ViewportOffset + 1
	}
	return nil
}

// moveTaskLeft moves the currently selected task to the previous column (left)
// Updates both the local state and the database using the linked list structure
// The selection follows the moved task and the viewport scrolls if needed.
// Returns a tea.Cmd if a notification was generated, or nil otherwise.
func (m *Model) moveTaskLeft() tea.Cmd {
	// Get the current task
	task := m.getCurrentTask()
	if task == nil {
		return nil
	}

	// Check if there's a previous column using the linked list
	currentCol := m.getCurrentColumn()
	if currentCol == nil {
		return nil
	}
	if currentCol.PrevID == nil {
		// Already at first column - show notification
		return m.addNotification(state.LevelInfo, "There are no more columns to move to.")
	}

	// Use the new database function to move task
	ctx, cancel := m.UIContext()
	defer cancel()
	err := m.App.TaskService.MoveTaskToPrevColumn(ctx, task.ID)
	if err != nil {
		slog.Error("failed to moving task to previous column", "error", err)
		if err != tasksvc.ErrAlreadyFirstColumn {
			return m.addNotification(state.LevelError, "Failed to move task to previous column")
		}
		return nil
	}

	// Update local state: remove from current column
	tasks := m.AppState.Tasks()[currentCol.ID]
	m.AppState.Tasks()[currentCol.ID] = append(tasks[:m.UIState.SelectedTask], tasks[m.UIState.SelectedTask+1:]...)

	// Find the previous column and add task there
	prevColID := *currentCol.PrevID
	newPosition := len(m.AppState.Tasks()[prevColID])
	task.ColumnID = prevColID
	task.Position = newPosition
	m.AppState.Tasks()[prevColID] = append(m.AppState.Tasks()[prevColID], task)

	// Move selection to follow the task
	m.UIState.SelectedColumn = m.UIState.SelectedColumn - 1
	m.UIState.SelectedTask = newPosition

	// Ensure the moved task is visible (auto-scroll viewport if needed)
	if m.UIState.SelectedColumn < m.UIState.ViewportOffset {
		m.UIState.ViewportOffset = m.UIState.ViewportOffset - 1
	}
	return nil
}

// moveTaskUp moves the currently selected task up within its column
// Updates both the local state and the database
// The selection follows the moved task.
// Returns a tea.Cmd if a notification was generated, or nil otherwise.
func (m *Model) moveTaskUp() tea.Cmd {
	task := m.getCurrentTask()
	if task == nil {
		return nil
	}

	// Check if already at top (edge case handled here for quick feedback)
	if m.UIState.SelectedTask == 0 {
		return m.addNotification(state.LevelInfo, "Task is already at the top")
	}

	// Call database swap
	ctx, cancel := m.UIContext()
	defer cancel()
	err := m.App.TaskService.MoveTaskUp(ctx, task.ID)
	if err != nil {
		slog.Error("failed to moving task up", "error", err)
		if err != tasksvc.ErrAlreadyFirstTask {
			return m.addNotification(state.LevelError, "Failed to move task up")
		}
		return nil
	}

	// Update local state: swap tasks in slice
	currentCol := m.getCurrentColumn()
	if currentCol == nil {
		return nil
	}

	tasks := m.getTasksForColumn(currentCol.ID)
	if len(tasks) < 2 {
		return nil
	}

	selectedIdx := m.UIState.SelectedTask
	if selectedIdx == 0 || selectedIdx >= len(tasks) {
		return nil
	}

	// Swap positions in slice
	tasks[selectedIdx], tasks[selectedIdx-1] = tasks[selectedIdx-1], tasks[selectedIdx]

	// Update position values on the task objects
	tasks[selectedIdx].Position = selectedIdx
	tasks[selectedIdx-1].Position = selectedIdx - 1

	// Move selection to follow the task
	m.UIState.SelectedTask = selectedIdx - 1
	return nil
}

// moveTaskDown moves the currently selected task down within its column
// Updates both the local state and the database
// The selection follows the moved task.
// Returns a tea.Cmd if a notification was generated, or nil otherwise.
func (m *Model) moveTaskDown() tea.Cmd {
	task := m.getCurrentTask()
	if task == nil {
		return nil
	}

	// Get current tasks for edge case check
	currentCol := m.getCurrentColumn()
	if currentCol == nil {
		return nil
	}

	tasks := m.getTasksForColumn(currentCol.ID)
	selectedIdx := m.UIState.SelectedTask

	// Check if already at bottom
	if selectedIdx >= len(tasks)-1 {
		return m.addNotification(state.LevelInfo, "Task is already at the bottom")
	}

	// Call database swap
	ctx, cancel := m.UIContext()
	defer cancel()
	err := m.App.TaskService.MoveTaskDown(ctx, task.ID)
	if err != nil {
		slog.Error("failed to moving task down", "error", err)
		if err != tasksvc.ErrAlreadyLastTask {
			return m.addNotification(state.LevelError, "Failed to move task down")
		}
		return nil
	}

	// Update local state: swap tasks in slice
	tasks[selectedIdx], tasks[selectedIdx+1] = tasks[selectedIdx+1], tasks[selectedIdx]

	// Update position values on the task objects
	tasks[selectedIdx].Position = selectedIdx
	tasks[selectedIdx+1].Position = selectedIdx + 1

	// Move selection to follow the task
	m.UIState.SelectedTask = selectedIdx + 1
	return nil
}

// getCurrentProject returns the currently selected project
// Returns nil if there are no projects
func (m Model) getCurrentProject() *models.Project {
	return m.AppState.GetCurrentProject()
}

// switchToProject switches to a different project by index and reloads columns/tasks/labels
func (m Model) switchToProject(projectIndex int) {
	if projectIndex < 0 || projectIndex >= len(m.AppState.Projects()) {
		return
	}

	m.AppState.SetSelectedProject(projectIndex)

	// Clear task detail cache since we're switching projects
	if m.DetailCache != nil {
		m.DetailCache.Clear()
	}

	project := m.AppState.Projects()[projectIndex]

	ctx, cancel := m.DBContext()
	defer cancel()

	columns, err := m.App.ColumnService.GetColumnsByProject(ctx, project.ID)
	if err != nil {
		slog.Error("failed to loading columns for project", "project_id", project.ID, "error", err)
		columns = []*models.Column{}
	}

	tasks, err := m.fetchTasksForCurrentProject(ctx, project.ID)
	if err != nil {
		slog.Error("failed to loading tasks for project", "project_id", project.ID, "error", err)
		tasks = make(map[int][]*models.TaskSummary)
	}

	labels, err := m.App.LabelService.GetLabelsByProject(ctx, project.ID)
	if err != nil {
		slog.Error("failed to loading labels for project", "project_id", project.ID, "error", err)
		labels = []*models.Label{}
	}

	// Update daemon subscription to the new project
	// TODO: hmm idk about this
	if m.EventClient != nil && project.ID > 0 {
		slog.Info("subscribing to project events", "project_id", project.ID, "project_name", project.Name)
		if err := m.EventClient.Subscribe(project.ID); err != nil {
			slog.Error("error subscribing to project events", "project_id", project.ID, "error", err)
		} else {
			slog.Info("successfully subscribed to project events", "project_id", project.ID)
		}
	}

	m.AppState.SetColumns(columns)
	m.AppState.SetTasks(tasks)
	m.AppState.SetLabels(labels)

	// Clear field filters when switching projects since they're project-scoped.
	// Search query persists across project switches.
	m.UI.Filter.ClearFieldFilters()
	m.UI.Filter.IsActive = false

	m.UIState.ResetSelection()
}

// reloadProjects reloads the projects list from the database
func (m Model) reloadProjects() {
	ctx, cancel := m.DBContext()
	defer cancel()
	projects, err := m.App.ProjectService.GetAllProjects(ctx)
	if err != nil {
		slog.Error("failed to reloading projects", "error", err)
		return
	}
	m.AppState.SetProjects(projects)
}

// initParentPickerForForm initializes the parent picker for use in task form mode.
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
	ctx, cancel := m.DBContext()
	defer cancel()
	allTasks, err := m.App.TaskService.GetTaskReferencesForProject(ctx, project.ID)
	if err != nil {
		slog.Error("failed to loading project tasks", "error", err)
		return false
	}

	// Build map of currently selected parent task IDs and their relation types from form state
	parentTaskMap := make(map[int]int) // map[taskID]relationTypeID
	for _, parentRef := range m.Forms.Form.FormParentRefs {
		parentTaskMap[parentRef.ID] = parentRef.RelationTypeID
	}

	// Build map of child task IDs to exclude (prevent circular dependencies)
	// A task that is already a child cannot become a parent
	childTaskMap := make(map[int]bool)
	for _, childID := range m.Forms.Form.FormChildIDs {
		childTaskMap[childID] = true
	}

	// Build picker items from all tasks, excluding current task (if editing)
	// and excluding tasks that are already children (circular dependency prevention)
	items := make([]state.TaskPickerItem, 0, len(allTasks))
	for _, task := range allTasks {
		// In edit mode, exclude the task being edited
		if m.Forms.Form.EditingTaskID != 0 && task.ID == m.Forms.Form.EditingTaskID {
			continue
		}

		if childTaskMap[task.ID] {
			continue
		}

		relationTypeID, isSelected := parentTaskMap[task.ID]

		items = append(items, state.TaskPickerItem{
			TaskRef:        task,
			Selected:       isSelected,
			RelationTypeID: relationTypeID,
		})
	}

	// Initialize ParentPickerState
	m.Pickers.Parent.Items = items
	m.Pickers.Parent.TaskID = m.Forms.Form.EditingTaskID // 0 for create mode
	m.Pickers.Parent.Cursor = 0
	m.Pickers.Parent.Filter = ""
	m.Pickers.Parent.PickerType = "parent"
	m.Pickers.Parent.ReturnMode = state.TaskFormMode

	return true
}

// initChildPickerForForm initializes the child picker for use in task form mode.
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
	ctx, cancel := m.DBContext()
	defer cancel()
	allTasks, err := m.App.TaskService.GetTaskReferencesForProject(ctx, project.ID)
	if err != nil {
		slog.Error("failed to loading project tasks", "error", err)
		return false
	}

	// Build map of currently selected child task IDs and their relation types from form state
	childTaskMap := make(map[int]int) // map[taskID]relationTypeID
	for _, childRef := range m.Forms.Form.FormChildRefs {
		childTaskMap[childRef.ID] = childRef.RelationTypeID
	}

	// Build map of parent task IDs to exclude (prevent circular dependencies)
	// A task that is already a parent cannot become a child
	parentTaskMap := make(map[int]bool)
	for _, parentID := range m.Forms.Form.FormParentIDs {
		parentTaskMap[parentID] = true
	}

	// Build picker items from all tasks, excluding current task (if editing)
	// and excluding tasks that are already parents (circular dependency prevention)
	items := make([]state.TaskPickerItem, 0, len(allTasks))
	for _, task := range allTasks {
		// In edit mode, exclude the task being edited
		if m.Forms.Form.EditingTaskID != 0 && task.ID == m.Forms.Form.EditingTaskID {
			continue
		}
		if parentTaskMap[task.ID] {
			continue
		}

		relationTypeID, isSelected := childTaskMap[task.ID]

		items = append(items, state.TaskPickerItem{
			TaskRef:        task,
			Selected:       isSelected,
			RelationTypeID: relationTypeID,
		})
	}

	// Initialize ChildPickerState
	m.Pickers.Child.Items = items
	m.Pickers.Child.TaskID = m.Forms.Form.EditingTaskID // 0 for create mode
	m.Pickers.Child.Cursor = 0
	m.Pickers.Child.Filter = ""
	m.Pickers.Child.PickerType = "child"
	m.Pickers.Child.ReturnMode = state.TaskFormMode

	return true
}

// initLabelPicker initializes the label picker for both form and board modes.
// - In form mode: Uses FormLabelIDs from form state (edit or create)
// - In board mode: Uses labels from the currently selected task (quick edit)
//
// Returns (false, cmd) if there's no current project or (in board mode) no selected task.
// Returns (true, nil) on success.
func (m *Model) initLabelPicker(mode state.Mode) (bool, tea.Cmd) {
	project := m.getCurrentProject()
	if project == nil {
		return false, m.addNotification(state.LevelError, "No project selected")
	}

	var labelIDMap map[int]bool
	var taskID int

	if mode == state.TaskFormMode {
		// Form mode: use FormLabelIDs
		labelIDMap = make(map[int]bool)
		for _, labelID := range m.Forms.Form.FormLabelIDs {
			labelIDMap[labelID] = true
		}
		taskID = m.Forms.Form.EditingTaskID // 0 for create mode
	} else {
		// Board mode: use current task's labels
		task := m.getCurrentTask()
		if task == nil {
			return false, m.addNotification(state.LevelError, "No task selected")
		}
		labelIDMap = make(map[int]bool)
		for _, label := range task.Labels {
			labelIDMap[label.ID] = true
		}
		taskID = task.ID
		m.Forms.Form.EditingTaskID = taskID // Set for board mode updates
	}

	// Build picker items from all available labels
	var items []state.LabelPickerItem
	for _, label := range m.AppState.Labels() {
		items = append(items, state.LabelPickerItem{
			Label:    label,
			Selected: labelIDMap[label.ID],
		})
	}

	// Initialize LabelPickerState
	m.Pickers.Label.Items = items
	m.Pickers.Label.TaskID = taskID
	m.Pickers.Label.Cursor = 0
	m.Pickers.Label.Filter = ""
	m.Pickers.Label.ReturnMode = mode

	return true, nil
}

// getFilteredLabelPickerItems returns label picker items filtered by the current filter text
func (m *Model) getFilteredLabelPickerItems() []state.LabelPickerItem {
	// Delegate to LabelPickerState which now owns this logic
	return m.Pickers.Label.GetFilteredItems()
}

// initPriorityPicker initializes the priority picker for both form and board modes.
// - In form mode: Defaults to medium priority (id=3) for new tasks, loads from DB for editing
// - In board mode: Loads priority from the currently selected task
//
// Returns (false, cmd) if there's a database error or (in board mode) no selected task.
// Returns (true, nil) on success.
func (m *Model) initPriorityPicker(mode state.Mode) (bool, tea.Cmd) {
	var currentPriorityID int
	var taskID int

	if mode == state.TaskFormMode {
		// Form mode: default to medium priority for new tasks
		currentPriorityID = 3
		taskID = m.Forms.Form.EditingTaskID

		// If editing an existing task, get the current priority ID from database
		if taskID != 0 {
			ctx, cancel := m.DBContext()
			defer cancel()

			_, priorityID, err := m.App.TaskService.GetTaskTypeAndPriorityIDs(ctx, taskID)
			if err != nil {
				slog.Error("failed to get task priority ID for priority picker", "error", err)
				return false, nil
			}
			currentPriorityID = priorityID
		}
	} else {
		// Board mode: use current task's priority
		task := m.getCurrentTask()
		if task == nil {
			return false, m.addNotification(state.LevelError, "No task selected")
		}

		ctx, cancel := m.DBContext()
		defer cancel()

		_, priorityID, err := m.App.TaskService.GetTaskTypeAndPriorityIDs(ctx, task.ID)
		if err != nil {
			slog.Error("failed to get task priority ID for board picker", "error", err)
			return false, m.addNotification(state.LevelError, "Failed to load task priority")
		}

		currentPriorityID = priorityID
		taskID = task.ID
		m.Forms.Form.EditingTaskID = taskID // Set for board mode updates
	}

	// Initialize PriorityPickerState
	m.Pickers.Priority.SetSelectedPriorityID(currentPriorityID)
	m.Pickers.Priority.SetCursor(currentPriorityID - 1) // 0-indexed
	m.Pickers.Priority.ReturnMode = mode

	return true, nil
}

// initTypePicker initializes the type picker for both form and board modes.
// - In form mode: Defaults to task (id=1) for new tasks, loads from DB for editing
// - In board mode: Loads type from the currently selected task
//
// Returns (false, cmd) if there's a database error or (in board mode) no selected task.
// Returns (true, nil) on success.
func (m *Model) initTypePicker(mode state.Mode) (bool, tea.Cmd) {
	var currentTypeID int
	var taskID int

	if mode == state.TaskFormMode {
		currentTypeID = 1 // Default to task
		taskID = m.Forms.Form.EditingTaskID

		if taskID != 0 {
			ctx, cancel := m.DBContext()
			defer cancel()

			typeID, _, err := m.App.TaskService.GetTaskTypeAndPriorityIDs(ctx, taskID)
			if err != nil {
				slog.Error("failed to get task type ID for type picker", "error", err)
				return false, nil
			}
			currentTypeID = typeID
		}
	} else {
		task := m.getCurrentTask()
		if task == nil {
			return false, m.addNotification(state.LevelError, "No task selected")
		}

		ctx, cancel := m.DBContext()
		defer cancel()

		typeID, _, err := m.App.TaskService.GetTaskTypeAndPriorityIDs(ctx, task.ID)
		if err != nil {
			slog.Error("failed to get task type ID for board picker", "error", err)
			return false, m.addNotification(state.LevelError, "Failed to load task type")
		}

		currentTypeID = typeID
		taskID = task.ID
		m.Forms.Form.EditingTaskID = taskID
	}

	m.Pickers.Type.SetSelectedTypeID(currentTypeID)
	m.Pickers.Type.SetCursor(currentTypeID - 1)
	m.Pickers.Type.ReturnMode = mode

	return true, nil
}

// initAssigneePicker initializes the assignee picker for both form and board modes.
// - In form mode: Uses FormAssigneeID from form state
// - In board mode: Uses assignee from the currently selected task
//
// Returns (false, cmd) if there's a database error or (in board mode) no selected task.
// Returns (true, nil) on success.
func (m *Model) initAssigneePicker(mode state.Mode) (bool, tea.Cmd) {
	ctx, cancel := m.DBContext()
	defer cancel()

	assignees, err := m.App.AssigneeService.List(ctx)
	if err != nil {
		slog.Error("failed to load assignees for picker", "error", err)
		return false, m.addNotification(state.LevelError, "Failed to load assignees")
	}

	var currentAssigneeID int
	var taskID int

	if mode == state.TaskFormMode {
		// Form mode: use FormAssigneeID
		currentAssigneeID = m.Forms.Form.FormAssigneeID
	} else {
		// Board mode: use current task's assignee
		task := m.getCurrentTask()
		if task == nil {
			return false, m.addNotification(state.LevelError, "No task selected")
		}

		if task.AssigneeID != nil {
			currentAssigneeID = *task.AssigneeID
		}
		taskID = task.ID
		m.Forms.Form.EditingTaskID = taskID // Set for board mode updates
	}

	m.Pickers.Assignee.SetAssignees(assignees)
	m.Pickers.Assignee.SetSelectedID(currentAssigneeID)

	// Position cursor to match current assignee (offset by 1 for the clear option)
	cursorPos := 0 // Default to "clear assignee" option
	for i, a := range assignees {
		if a.ID == currentAssigneeID {
			cursorPos = i + 1 // +1 because index 0 is the clear option
			break
		}
	}
	m.Pickers.Assignee.SetCursor(cursorPos)
	m.Pickers.Assignee.ReturnMode = mode

	return true, nil
}

// initProjectPicker initializes the project picker for moving a task to another project.
// It loads all projects, filters out the current project, and sets up the picker state.
// Returns (false, cmd) if there's a database error, no current project, no other projects, or no selected task.
// Returns (true, nil) on success.
func (m *Model) initProjectPicker() (bool, tea.Cmd) {
	ctx, cancel := m.DBContext()
	defer cancel()

	projects, err := m.App.ProjectService.GetAllProjects(ctx)
	if err != nil {
		slog.Error("failed to load projects for picker", "error", err)
		return false, m.addNotification(state.LevelError, "Failed to load projects")
	}

	currentProject := m.getCurrentProject()
	if currentProject == nil {
		return false, m.addNotification(state.LevelError, "No project selected")
	}

	// Filter out the current project
	filtered := make([]*models.Project, 0, len(projects))
	for _, p := range projects {
		if p.ID != currentProject.ID {
			filtered = append(filtered, p)
		}
	}

	if len(filtered) == 0 {
		return false, m.addNotification(state.LevelInfo, "No other projects available")
	}

	task := m.getCurrentTask()
	if task == nil {
		return false, m.addNotification(state.LevelError, "No task selected")
	}

	m.Forms.Form.EditingTaskID = task.ID
	m.Pickers.Project.SetProjects(filtered)
	m.Pickers.Project.SetCursor(0)
	m.Pickers.Project.ReturnMode = state.NormalMode

	return true, nil
}

// initEstimateInput initializes the estimate input for both form and board modes.
// - In form mode: Uses FormEstimate from form state (edit or create)
// - In board mode: Uses estimate from the currently selected task (quick edit)
//
// Returns (false, cmd) if there's no selected task in board mode.
// Returns (true, nil) on success.
func (m *Model) initEstimateInput(mode state.Mode) (bool, tea.Cmd) {
	var estimateValue string
	var taskID int

	if mode == state.TaskFormMode {
		// Form mode: use FormEstimate
		estimateValue = m.Forms.Form.FormEstimate
	} else {
		// Board mode: use current task's estimate
		task := m.getCurrentTask()
		if task == nil {
			return false, m.addNotification(state.LevelError, "No task selected")
		}
		if task.Estimate != nil {
			estimateValue = *task.Estimate
		} else {
			estimateValue = ""
		}
		taskID = task.ID
		m.Forms.Form.EditingTaskID = taskID // Set for board mode updates
	}

	m.Pickers.Estimate.SetBuffer(estimateValue)
	m.Pickers.Estimate.SetError("")
	m.Pickers.Estimate.ReturnMode = mode

	return true, nil
}

// initEstimateInputForForm initializes the estimate input for use in task form mode.
// Pre-fills the input buffer with the current estimate value.
// Deprecated: Use initEstimateInput(state.TaskFormMode) instead.
func (m *Model) initEstimateInputForForm() {
	_, _ = m.initEstimateInput(state.TaskFormMode)
}

// initDatePickerForForm initializes the date picker for use in task form mode.
// Pre-fills the picker with the current due date value from FormDueDate.
// If FormDueDate is nil, defaults to today's date.
func (m *Model) initDatePickerForForm() {
	m.Pickers.DatePicker.InitFromDate(m.Forms.Form.FormDueDate)
	m.Pickers.DatePicker.ReturnMode = state.TaskFormMode
}

func (m *Model) initDatePicker(mode state.Mode) (bool, tea.Cmd) {
	var currentDueDate *time.Time

	if mode == state.TaskFormMode {
		// Form mode: use the form's due date
		currentDueDate = m.Forms.Form.FormDueDate
	} else {
		// Board mode: use current task's due date
		task := m.getCurrentTask()
		if task == nil {
			return false, m.addNotification(state.LevelError, "No task selected")
		}

		currentDueDate = task.DueDate
		m.Forms.Form.EditingTaskID = task.ID // Set for board mode updates
	}

	// Initialize DatePickerState
	m.Pickers.DatePicker.InitFromDate(currentDueDate)
	m.Pickers.DatePicker.ReturnMode = mode

	return true, nil
}

// buildListViewRows creates a flat list of all tasks with their column names.
// The list is sorted according to the current sort settings in listViewState.
func (m Model) buildListViewRows() []renderers.ListViewRow {
	var rows []renderers.ListViewRow
	for _, col := range m.AppState.Columns() {
		tasks := m.AppState.Tasks()[col.ID]
		for _, task := range tasks {
			rows = append(rows, renderers.ListViewRow{
				Task:       task,
				ColumnName: col.Name,
				ColumnID:   col.ID,
			})
		}
	}

	m.sortListViewRows(rows)
	return rows
}

// sortListViewRows sorts the rows based on current sort settings.
func (m Model) sortListViewRows(rows []renderers.ListViewRow) {
	if m.UI.ListView.SortField() == state.SortNone {
		return
	}

	sort.Slice(rows, func(i, j int) bool {
		var cmp int
		switch m.UI.ListView.SortField() {
		case state.SortByTitle:
			cmp = strings.Compare(rows[i].Task.Title, rows[j].Task.Title)
		case state.SortByStatus:
			cmp = strings.Compare(rows[i].ColumnName, rows[j].ColumnName)
		default:
			return false
		}

		if m.UI.ListView.SortOrder() == state.SortDesc {
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
		m.UI.ListView.SetSelectedRow(0)
		return
	}

	// find task that matches the current kanban selection
	currentTask := m.getCurrentTask()
	if currentTask == nil {
		m.UI.ListView.SetSelectedRow(0)
		return
	}

	for i, row := range rows {
		if row.Task.ID == currentTask.ID {
			m.UI.ListView.SetSelectedRow(i)
			return
		}
	}
	m.UI.ListView.SetSelectedRow(0)
}

// syncListToKanbanSelection maps the current list row to kanban column/task selection.
// This should be called when switching from list to kanban view.
func (m *Model) syncListToKanbanSelection() {
	rows := m.buildListViewRows()
	if len(rows) == 0 {
		return
	}

	selectedRow := m.UI.ListView.SelectedRow()
	if selectedRow >= len(rows) {
		selectedRow = len(rows) - 1
	}
	if selectedRow < 0 {
		return
	}

	selectedTask := rows[selectedRow].Task

	// column and task position in kanban view
	for colIdx, col := range m.AppState.Columns() {
		tasks := m.AppState.Tasks()[col.ID]
		for taskIdx, task := range tasks {
			if task.ID == selectedTask.ID {
				m.UIState.SelectedColumn = colIdx
				m.UIState.SelectedTask = taskIdx
				m.UIState.EnsureSelectionVisible(colIdx)
				return
			}
		}
	}
}

// subscribeToEvents returns a command that listens for events from the daemon
// and sends RefreshMsg when data changes
func (m Model) subscribeToEvents() tea.Cmd {
	if m.EventChan == nil {
		return nil
	}

	return func() tea.Msg {
		select {
		case event, ok := <-m.EventChan:
			if !ok {
				return nil
			}
			return RefreshMsg{Event: event}
		case <-m.Ctx.Done():
			return nil
		}
	}
}

// reloadCurrentProject refreshes columns, tasks, and labels for the current project
func (m *Model) reloadCurrentProject() {
	currentProject := m.AppState.GetCurrentProject()
	if currentProject == nil {
		return
	}

	ctx, cancel := m.DBContext()
	defer cancel()

	columns, err := m.App.ColumnService.GetColumnsByProject(ctx, currentProject.ID)
	if err != nil {
		slog.Error("failed to reloading columns", "error", err)
		m.HandleDBError(err, "reload columns")
		return
	}

	tasks, err := m.fetchTasksForCurrentProject(ctx, currentProject.ID)
	if err != nil {
		slog.Error("failed to reloading tasks", "error", err)
		m.HandleDBError(err, "reload tasks")
		return
	}

	labels, err := m.App.LabelService.GetLabelsByProject(ctx, currentProject.ID)
	if err != nil {
		slog.Error("failed to reloading labels", "error", err)
		m.HandleDBError(err, "reload labels")
		return
	}

	m.AppState.SetColumns(columns)
	m.AppState.SetTasks(tasks)
	m.AppState.SetLabels(labels)
}

// fetchTasksForCurrentProject fetches tasks for the given project ID,
// respecting all active filters from the filter bar (including search query).
// Uses the unified filter query so search + filters combine with AND logic.
func (m *Model) fetchTasksForCurrentProject(ctx context.Context, projectID int) (map[int][]*models.TaskSummary, error) {
	if !m.UI.Filter.HasAnyFilter() {
		return m.App.TaskService.GetTaskSummariesByProject(ctx, projectID)
	}

	params := m.buildFilterParams(projectID)
	return m.App.TaskService.GetTaskSummariesWithFilters(ctx, params)
}

// buildFilterParams constructs TaskFilterParams from the current filter state (including search).
func (m *Model) buildFilterParams(projectID int) tasksvc.TaskFilterParams {
	params := tasksvc.TaskFilterParams{
		ProjectID: projectID,
	}

	if m.UI.Filter.SearchQuery != "" {
		q := m.UI.Filter.SearchQuery
		params.Title = &q
	}

	params.PriorityID = m.UI.Filter.PriorityID
	params.TypeID = m.UI.Filter.TypeID
	params.AssigneeID = m.UI.Filter.AssigneeID
	params.LabelIDs = m.UI.Filter.LabelIDs
	params.ShowArchived = m.UI.Filter.ShowArchived

	return params
}

// calculateDescriptionLines calculates the number of lines for the
// description field in the task form based on current screen height
func (m Model) calculateDescriptionLines() int {
	const ( // WARN: keep in sync with renderTaskFormLayer
		chromeHeight       = 6  // border (2) + padding (2) + title (1) + blanks (1)
		descChromeOverhead = 9  // title field (3) + confirmation (3) + spacing (3)
		minDescLines       = 5  // minimum lines to view description
		maxDescLines       = 15 // max description lines
	)

	layerHeight := m.UIState.Height * 8 / 10
	innerHeight := layerHeight - chromeHeight
	topLeftHeight := innerHeight * 7 / 10

	return min(
		max(topLeftHeight-descChromeOverhead, minDescLines),
		maxDescLines,
	)
}
