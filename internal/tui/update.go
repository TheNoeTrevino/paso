package tui

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/thenoetrevino/paso/internal/app"
	"github.com/thenoetrevino/paso/internal/database"
	"github.com/thenoetrevino/paso/internal/events"
	"github.com/thenoetrevino/paso/internal/git"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui/commands"
	"github.com/thenoetrevino/paso/internal/tui/components"
	"github.com/thenoetrevino/paso/internal/tui/huhforms"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

type RefreshMsg struct {
	Event events.Event
}

type ConnectionEstablishedMsg struct{}

type ConnectionLostMsg struct{}

type ConnectionReconnectingMsg struct{}

// Database switching messages
type databaseConnected struct {
	dbType database.DatabaseType
	dbName string
	newDB  *sql.DB  // The new database connection
	newApp *app.App // The new app with services pointing to newDB
}

type databaseConnectionError struct {
	err error
}

// SpinnerTickMsg is sent periodically to advance the connecting spinner animation
type SpinnerTickMsg time.Time

type dataReloaded struct {
	projects []*models.Project
	columns  []*models.Column
	tasks    map[int][]*models.TaskSummary
	labels   []*models.Label
}

// Git operation messages
type gitInfoFetched struct {
	gitInfo     git.GitInfo
	gitBranches []git.BranchInfo
	forEdit     bool
}

type gitInfoError struct {
	err     error
	forEdit bool
}

// GitSpinnerTickMsg is sent periodically to advance the git loading spinner animation
type GitSpinnerTickMsg time.Time

// taskDetailForEditMsg is sent when task details are fetched for the edit form
type taskDetailForEditMsg struct {
	taskID     int
	taskDetail *models.TaskDetail
	activities []models.ActivityItem
}

// taskDetailForEditError is sent when fetching task details fails
type taskDetailForEditError struct {
	err error
}

// prefetchDebounceMsg is sent after the debounce delay to trigger actual prefetch.
// Contains the sequence number to verify the request is still valid.
//
// Synchronization: The seq field enables safe debouncing without explicit locks.
// When a navigation occurs, PrefetchDebounceSeq is incremented and captured in this
// message. If another navigation happens before this message is processed, the new
// navigation increments PrefetchDebounceSeq again, causing this message's seq to be
// stale (msg.seq != m.PrefetchDebounceSeq) and ignored. This pattern is thread-safe
// because Bubble Tea's Update loop is single-threaded.
type prefetchDebounceMsg struct {
	seq int
}

// databaseSaved and connectionConfirmation message types are defined in update_database.go

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	select {
	case <-m.Ctx.Done():
		close(m.NotifyChan)
		return m, tea.Quit
	default:
	}

	var cmd tea.Cmd
	if m.EventChan != nil && !m.SubscriptionStarted {
		m.SubscriptionStarted = true
		cmd = m.subscribeToEvents()
	}

	if m.UIState.Mode == state.TicketFormMode {
		return m.updateTaskForm(msg)
	}
	if m.UIState.Mode == state.ProjectFormMode || m.UIState.Mode == state.EditProjectFormMode {
		return m.updateProjectForm(msg)
	}
	// Handle column forms early to prevent key binding conflicts (e.g., space key)
	if m.UIState.Mode == state.AddColumnFormMode || m.UIState.Mode == state.EditColumnFormMode {
		return m.updateColumnForm(msg)
	}
	// Handle comment form early to prevent key binding conflicts (e.g., space key)
	if m.UIState.Mode == state.CommentFormMode {
		return m.updateCommentForm(msg)
	}
	// Handle CommentEditMode early to prevent key binding conflicts (e.g., space key mapped to ViewTask)
	if m.UIState.Mode == state.CommentEditMode {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			return m.updateCommentEdit(keyMsg)
		}
		return m, nil
	}
	// Handle database create form early to prevent key binding conflicts
	if m.UIState.Mode == state.DatabaseCreateMode {
		return m.updateDatabaseCreate(msg)
	}

	switch msg := msg.(type) {
	case RefreshMsg:
		currentProject := m.AppState.GetCurrentProject()
		// Reload if event is for current project OR for all projects (0)
		if currentProject != nil {
			slog.Info("received refresh event",
				"event_project_id", msg.Event.ProjectID,
				"current_project_id", currentProject.ID,
				"will_reload", msg.Event.ProjectID == currentProject.ID || msg.Event.ProjectID == 0)
		}
		if currentProject != nil && (msg.Event.ProjectID == currentProject.ID || msg.Event.ProjectID == 0) {
			m.reloadCurrentProject()
		}

		cmd = m.subscribeToEvents()
		return m, cmd

	case events.NotificationMsg:
		return m.handleNotificationMsg(msg)

	case ConnectionEstablishedMsg:
		m.ConnectionState.SetStatus(state.Connected)
		return m, nil

	case ConnectionLostMsg:
		m.ConnectionState.SetStatus(state.Disconnected)
		return m, nil

	case ConnectionReconnectingMsg:
		m.ConnectionState.SetStatus(state.Reconnecting)
		return m, nil

	case databaseConnected:
		// Close old database connection synchronously (safe - no concurrent access)
		if m.DB != nil {
			_ = m.DB.Close()
		}

		// Swap to new database and app atomically
		slog.Info("before swap", "old_app", m.App, "new_app", msg.newApp)
		m.DB = msg.newDB
		m.App = msg.newApp
		slog.Info("after swap", "m.App", m.App, "m.App.ColumnService", m.App.ColumnService)
		m.CurrentDBType = msg.dbType
		m.CurrentDBName = msg.dbName
		m.UIState.Mode = state.NormalMode
		m.DatabasePicker.Reset()

		// Stop spinner animation and show success
		m.DatabasePicker.StopConnecting()
		m.UI.Notification.Add(state.LevelInfo, fmt.Sprintf("Connected to %s", msg.dbName))

		slog.Info("database switched successfully", "name", msg.dbName, "type", msg.dbType)

		// Reload all data from new database
		return m, m.reloadAllData()

	case databaseConnectionError:
		m.UIState.Mode = state.NormalMode
		m.DatabasePicker.Reset()
		// Stop spinner animation and show error
		m.DatabasePicker.StopConnecting()
		m.UI.Notification.Add(state.LevelError, "Database connection failed: "+msg.err.Error())
		return m, nil

	case dataReloaded:
		// Clear search when switching databases since it's a different context
		m.UI.Search.Clear()
		m.UI.Search.Deactivate()

		m.AppState.SetProjects(msg.projects)
		m.AppState.SetColumns(msg.columns)
		m.AppState.SetTasks(msg.tasks)
		m.AppState.SetLabels(msg.labels)
		m.reloadCurrentProject()
		return m, nil

	case SpinnerTickMsg:
		// Only advance spinner if still connecting
		if m.DatabasePicker.IsConnecting() {
			m.DatabasePicker.AdvanceSpinnerFrame()
			// Return next tick to continue animation
			return m, tickConnectingSpinner()
		}
		// If not connecting anymore, stop animation
		return m, nil

	case GitSpinnerTickMsg:
		if m.LoadingGitInfo || m.LoadingBranches {
			m.SpinnerFrame = (m.SpinnerFrame + 1) % components.GetSpinnerFrameCount()
			return m, tickGitSpinner()
		}
		return m, nil

	case taskDetailForEditMsg:
		return m.handleTaskDetailForEdit(msg)

	case taskDetailForEditError:
		m.LoadingGitInfo = false
		m.SpinnerFrame = 0
		m.UIState.Mode = state.NormalMode
		m.HandleDBError(msg.err, "Loading task details")
		return m, nil

	case gitInfoFetched:
		m.LoadingGitInfo = false
		m.LoadingBranches = false
		return m.handleGitInfoFetched(msg)

	case gitInfoError:
		m.LoadingGitInfo = false
		m.LoadingBranches = false

		if errors.Is(msg.err, context.Canceled) {
			slog.Info("git info fetch cancelled", "forEdit", msg.forEdit)
		} else if errors.Is(msg.err, context.DeadlineExceeded) {
			slog.Warn("git info fetch timed out", "forEdit", msg.forEdit)
			m.UI.Notification.Add(state.LevelWarning, "Git operation timed out. Creating form without branch list.")
		} else {
			slog.Warn("failed to fetch git info", "error", msg.err, "forEdit", msg.forEdit)
			m.UI.Notification.Add(state.LevelWarning, "Could not load git branches. Creating form without branch list.")
		}

		return m.createProjectFormWithoutGit(msg.forEdit)

	case connectionConfirmation:
		if msg.connect {
			// User wants to connect to the new connection
			m.DatabasePicker.StartConnecting(msg.config.Name)
			m.UIState.Mode = state.NormalMode
			return m, m.switchToDatabaseConfig(msg.config)
		} else {
			// User wants to stay in selection view
			m.DatabasePicker.PendingConnection = nil
			m.UIState.Mode = state.DatabaseSelectMode
			return m, nil
		}

	case prefetchDebounceMsg:
		// Only trigger prefetch if this is still the current sequence.
		// This check invalidates requests from previous navigations, ensuring only
		// the most recent navigation triggers a prefetch after the debounce delay.
		if msg.seq != m.PrefetchDebounceSeq {
			return m, nil // Stale request, ignore
		}
		return m, m.executePrefetch()

	case commands.TaskDetailsPrefetchedMsg:
		// Stop loading spinner
		m.DetailPanelLoading = false

		// Update cache with fetched details
		if m.DetailCache != nil && len(msg.Details) > 0 {
			m.DetailCache.SetBatch(msg.Details)
		}

		// Log any errors that occurred during prefetching
		for taskID, err := range msg.Errors {
			slog.Warn("failed to prefetch task detail", "task_id", taskID, "error", err)
		}

		return m, nil

	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case tea.WindowSizeMsg:
		return m.handleWindowResize(msg)
	}

	return m, cmd
}

func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.UIState.Mode {
	case state.NormalMode:
		return m.handleNormalMode(msg)
	case state.DiscardConfirmMode:
		return m.handleDiscardConfirm(msg)
	case state.DeleteConfirmMode:
		return m.handleDeleteConfirm(msg)
	case state.DeleteColumnConfirmMode:
		return m.handleDeleteColumnConfirm(msg)
	case state.DeleteProjectConfirmMode:
		return m.handleDeleteProjectConfirm(msg)
	case state.DeleteLabelConfirmMode:
		return m.handleDeleteLabelConfirm(msg)
	case state.ProjectBranchConfirmMode:
		return m.handleProjectBranchConfirm(msg)
	case state.TicketFormLoadingMode:
		switch msg.String() {
		case "esc", m.Config.KeyMappings.Quit:
			m.LoadingGitInfo = false
			m.SpinnerFrame = 0
			m.UIState.Mode = state.NormalMode
			return m, tea.ClearScreen
		}
		return m, nil
	case state.ProjectFormLoadingMode:
		switch msg.String() {
		case "esc", m.Config.KeyMappings.Quit:
			m.LoadingGitInfo = false
			m.LoadingBranches = false
			m.Forms.Form.ClearProjectForm()
			m.UIState.Mode = state.NormalMode
			return m, tea.ClearScreen
		}
		return m, nil
	case state.CommentsViewMode:
		return m.handleCommentsViewInput(msg)
	case state.HelpMode:
		switch msg.String() {
		case m.Config.KeyMappings.ShowHelp, m.Config.KeyMappings.Quit, "esc", "enter", " ":
			m.UIState.Mode = state.NormalMode
			return m, nil
		}
		return m, nil
	case state.TaskFormHelpMode:
		switch msg.String() {
		case "ctrl+h", "esc":
			m.UIState.Mode = state.TicketFormMode
			return m, nil
		}
		return m, nil
	case state.LabelPickerMode:
		return m.updateLabelPicker(msg)
	case state.ParentPickerMode:
		return m.updateParentPicker(msg)
	case state.ChildPickerMode:
		return m.updateChildPicker(msg)
	case state.PriorityPickerMode:
		return m.updatePriorityPicker(msg)
	case state.TypePickerMode:
		return m.updateTypePicker(msg)
	case state.AssigneePickerMode:
		return m.updateAssigneePicker(msg)
	case state.EstimateInputMode:
		return m.updateEstimateInput(msg)
	case state.DatePickerMode:
		return m.updateDatePicker(msg)
	case state.RelationTypePickerMode:
		return m.updateRelationTypePicker(msg)
	case state.DatabaseSelectMode:
		return m.updateDatabaseSelect(msg)
	case state.DatabaseConnectConfirmMode:
		return m.updateDatabaseConnectConfirm(msg)
	case state.DatabaseDeleteConfirmMode:
		return m.updateDatabaseDeleteConfirm(msg)
	case state.SearchMode:
		return m.handleSearchMode(msg)
	case state.StatusPickerMode:
		return m.handleStatusPickerMode(msg)
	}
	return m, nil
}

func (m Model) handleWindowResize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.UIState.SetWidth(msg.Width)
	m.UIState.Height = msg.Height

	m.UI.Notification.SetWindowSize(msg.Width, msg.Height)

	if m.UIState.ViewportOffset+m.UIState.ViewportSize() > len(m.AppState.Columns()) {
		m.UIState.ViewportOffset = max(0, len(m.AppState.Columns())-m.UIState.ViewportSize())
	}

	if !m.InitialPrefetchDone && m.UIState.ShouldShowDetailPanel() {
		m.InitialPrefetchDone = true
		return m, m.triggerDetailPanelPrefetch()
	}

	return m, nil
}

func (m Model) handleNotificationMsg(msg events.NotificationMsg) (tea.Model, tea.Cmd) {
	level := state.LevelInfo
	switch msg.Level {
	case "error":
		level = state.LevelError
	case "warning":
		level = state.LevelWarning
	}
	m.UI.Notification.Add(level, msg.Message)

	m.updateConnectionStateFromMessage(msg.Message)

	return m, m.listenForNotifications()
}

func (m *Model) updateConnectionStateFromMessage(message string) {
	if strings.Contains(message, "Connection lost") || strings.Contains(message, "reconnecting") {
		m.ConnectionState.SetStatus(state.Reconnecting)
	} else if strings.Contains(message, "Reconnected") {
		m.ConnectionState.SetStatus(state.Connected)
	} else if strings.Contains(message, "Failed to reconnect") {
		m.ConnectionState.SetStatus(state.Disconnected)
	}
}

// tickGitSpinner returns a command that sends a GitSpinnerTickMsg after a short delay
func tickGitSpinner() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return GitSpinnerTickMsg(t)
	})
}

// handleTaskDetailForEdit handles the taskDetailForEditMsg and creates the task edit form
func (m Model) handleTaskDetailForEdit(msg taskDetailForEditMsg) (tea.Model, tea.Cmd) {
	m.LoadingGitInfo = false
	m.SpinnerFrame = 0

	m.Forms.Form.FormTitle = msg.taskDetail.Title
	m.Forms.Form.FormDescription = msg.taskDetail.Description
	m.Forms.Form.FormLabelIDs = make([]int, len(msg.taskDetail.Labels))
	for i, label := range msg.taskDetail.Labels {
		m.Forms.Form.FormLabelIDs[i] = label.ID
	}

	m.Forms.Form.FormParentIDs = make([]int, len(msg.taskDetail.ParentTasks))
	m.Forms.Form.FormParentRefs = msg.taskDetail.ParentTasks
	for i, parent := range msg.taskDetail.ParentTasks {
		m.Forms.Form.FormParentIDs[i] = parent.ID
	}

	m.Forms.Form.FormChildIDs = make([]int, len(msg.taskDetail.ChildTasks))
	m.Forms.Form.FormChildRefs = msg.taskDetail.ChildTasks
	for i, child := range msg.taskDetail.ChildTasks {
		m.Forms.Form.FormChildIDs[i] = child.ID
	}

	m.Forms.Form.FormComments = msg.taskDetail.Comments
	m.Forms.Form.InitialFormComments = make([]*models.Comment, len(msg.taskDetail.Comments))
	copy(m.Forms.Form.InitialFormComments, msg.taskDetail.Comments)

	m.Forms.Form.FormActivities = msg.activities

	m.Forms.Form.FormCreatedAt = msg.taskDetail.CreatedAt
	m.Forms.Form.FormUpdatedAt = msg.taskDetail.UpdatedAt
	m.Forms.Form.FormTypeDescription = msg.taskDetail.TypeDescription
	m.Forms.Form.FormPriorityDescription = msg.taskDetail.PriorityDescription
	m.Forms.Form.FormPriorityColor = msg.taskDetail.PriorityColor
	if msg.taskDetail.AssigneeID != nil {
		m.Forms.Form.FormAssigneeID = *msg.taskDetail.AssigneeID
	}
	if msg.taskDetail.AssigneeName != nil {
		m.Forms.Form.FormAssigneeName = *msg.taskDetail.AssigneeName
	}
	if msg.taskDetail.Estimate != nil {
		m.Forms.Form.FormEstimate = *msg.taskDetail.Estimate
	}
	if msg.taskDetail.DueDate != nil {
		m.Forms.Form.FormDueDate = msg.taskDetail.DueDate
	}

	m.Forms.Form.FormConfirm = true
	m.Forms.Form.EditingTaskID = msg.taskID

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

// handleGitInfoFetched handles the gitInfoFetched message and creates the project form
func (m Model) handleGitInfoFetched(msg gitInfoFetched) (tea.Model, tea.Cmd) {
	if msg.forEdit {
		return m.createEditProjectFormWithGit(msg.gitInfo, msg.gitBranches)
	}

	// For create mode, set the default branch if in a valid git state
	if msg.gitInfo.IsRepo && msg.gitInfo.IsValidForAssociation() {
		m.Forms.Form.FormProjectGitBranch = msg.gitInfo.CurrentBranch
	}

	return m.createNewProjectFormWithGit(msg.gitInfo, msg.gitBranches)
}

// createProjectFormWithoutGit creates a project form without git branch selection
func (m Model) createProjectFormWithoutGit(forEdit bool) (tea.Model, tea.Cmd) {
	if forEdit {
		return m.createEditProjectFormWithGit(git.GitInfo{IsRepo: false}, nil)
	}
	return m.createNewProjectFormWithGit(git.GitInfo{IsRepo: false}, nil)
}

// createNewProjectFormWithGit creates a new project form with git info
func (m Model) createNewProjectFormWithGit(gitInfo git.GitInfo, gitBranches []git.BranchInfo) (tea.Model, tea.Cmd) {
	m.Forms.Form.Git.EditingProjectID = 0
	m.Forms.Form.FormProjectName = ""
	m.Forms.Form.FormProjectDescription = ""
	m.Forms.Form.FormProjectGitBranch = ""
	m.Forms.Form.FormProjectConfirm = true

	if gitInfo.IsRepo && gitInfo.IsValidForAssociation() {
		m.Forms.Form.FormProjectGitBranch = gitInfo.CurrentBranch
	}

	m.Forms.Form.ProjectForm = huhforms.CreateProjectForm(huhforms.ProjectFormProps{
		Name:        &m.Forms.Form.FormProjectName,
		Description: &m.Forms.Form.FormProjectDescription,
		GitBranch:   &m.Forms.Form.FormProjectGitBranch,
		Confirm:     &m.Forms.Form.FormProjectConfirm,
		IsEditing:   false,
		GitBranches: gitBranches,
		IsGitRepo:   gitInfo.IsRepo,
	}).WithTheme(huhforms.CreatePasoTheme(m.Config.ColorScheme))
	m.Forms.Form.SnapshotProjectFormInitialValues()
	m.UIState.Mode = state.ProjectFormMode
	return m, m.Forms.Form.ProjectForm.Init()
}

// createEditProjectFormWithGit creates an edit project form with git info
func (m Model) createEditProjectFormWithGit(gitInfo git.GitInfo, gitBranches []git.BranchInfo) (tea.Model, tea.Cmd) {
	currentProject := m.AppState.GetCurrentProject()
	if currentProject == nil {
		m.UI.Notification.Add(state.LevelWarning, "No project selected")
		m.LoadingGitInfo = false
		m.LoadingBranches = false
		return m, nil
	}

	m.Forms.Form.Git.EditingProjectID = currentProject.ID
	m.Forms.Form.FormProjectName = currentProject.Name
	m.Forms.Form.FormProjectDescription = currentProject.Description
	m.Forms.Form.FormProjectGitBranch = currentProject.GitBranch
	m.Forms.Form.FormProjectConfirm = true

	m.Forms.Form.ProjectForm = huhforms.CreateProjectForm(huhforms.ProjectFormProps{
		Name:        &m.Forms.Form.FormProjectName,
		Description: &m.Forms.Form.FormProjectDescription,
		GitBranch:   &m.Forms.Form.FormProjectGitBranch,
		Confirm:     &m.Forms.Form.FormProjectConfirm,
		IsEditing:   true,
		GitBranches: gitBranches,
		IsGitRepo:   gitInfo.IsRepo,
	}).WithTheme(huhforms.CreatePasoTheme(m.Config.ColorScheme))
	m.Forms.Form.SnapshotProjectFormInitialValues()
	m.UIState.Mode = state.EditProjectFormMode
	return m, m.Forms.Form.ProjectForm.Init()
}
