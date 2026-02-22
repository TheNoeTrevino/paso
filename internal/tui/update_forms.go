package tui

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"github.com/thenoetrevino/paso/internal/models"
	columnService "github.com/thenoetrevino/paso/internal/services/column"
	projectService "github.com/thenoetrevino/paso/internal/services/project"
	taskService "github.com/thenoetrevino/paso/internal/services/task"
	"github.com/thenoetrevino/paso/internal/tui/state"
	userutil "github.com/thenoetrevino/paso/internal/user"
)

type taskFormValues struct {
	title       string
	description string
	confirm     bool
	labelIDs    []int
}

// extractTaskFormValues extracts and returns form values from the task form
// Since our forms update pointers in place, we can just read from formState
func (m Model) extractTaskFormValues() taskFormValues {
	return taskFormValues{
		title:       strings.TrimSpace(m.Forms.Form.FormTitle),
		description: strings.TrimSpace(m.Forms.Form.FormDescription),
		confirm:     m.Forms.Form.FormConfirm,
		labelIDs:    m.Forms.Form.FormLabelIDs,
	}
}

// createNewTaskWithLabelsAndRelationships creates a new task, sets labels, and applies parent/child relationships
func (m *Model) createNewTaskWithLabelsAndRelationships(values taskFormValues) tea.Cmd {
	currentCol := m.getCurrentColumn()
	if currentCol == nil {
		m.UI.Notification.Add(state.LevelError, "No column selected")
		return nil
	}

	// Create context for database operations
	ctx, cancel := m.DBContext()
	defer cancel()

	// Build parent IDs with relation types
	parentIDs := make([]int, 0)
	for _, item := range m.Pickers.Parent.Items {
		if item.Selected {
			parentIDs = append(parentIDs, item.TaskRef.ID)
		}
	}

	// Build child IDs with relation types
	childIDs := make([]int, 0)
	for _, item := range m.Pickers.Child.Items {
		if item.Selected {
			childIDs = append(childIDs, item.TaskRef.ID)
		}
	}

	task, err := m.App.TaskService.CreateTask(ctx, taskService.CreateTaskRequest{
		Title:       values.title,
		Description: values.description,
		ColumnID:    currentCol.ID,
		Position:    len(m.getTasksForColumn(currentCol.ID)),
		AssigneeID:  m.Forms.Form.FormAssigneeID,
		Estimate:    m.Forms.Form.FormEstimate,
		DueDate:     m.Forms.Form.FormDueDate,
		LabelIDs:    values.labelIDs,
		ParentIDs:   parentIDs,
		ChildIDs:    childIDs,
	})
	if err != nil {
		slog.Error("failed to creating task", "error", err)
		m.UI.Notification.Add(state.LevelError, "Error creating task")
		return nil
	}

	m.applyParentRelationships(ctx, task.ID)
	m.applyChildRelationships(ctx, task.ID)

	project := m.getCurrentProject()
	if project != nil {
		tasksByColumn, err := m.fetchTasksForCurrentProject(ctx, project.ID)
		if err != nil {
			slog.Error("failed to reloading tasks", "error", err)
		} else {
			m.AppState.SetTasks(tasksByColumn)
		}
	}

	// Notify user if the new task is hidden by active filters
	if m.UI.Filter.HasAnyFilter() {
		return m.addNotification(state.LevelWarning, "Task created but hidden by active filters")
	}
	return nil
}

// updateExistingTaskWithLabelsAndRelationships updates task, labels, and parent/child relationships
func (m *Model) updateExistingTaskWithLabelsAndRelationships(values taskFormValues) tea.Cmd {
	// create context for database operations
	ctx, cancel := m.DBContext()
	defer cancel()
	taskID := m.Forms.Form.EditingTaskID

	// Invalidate cache entry since task is being modified
	if m.DetailCache != nil {
		m.DetailCache.Invalidate(taskID)
	}

	// update task basic fields
	title := values.title
	description := values.description
	err := m.App.TaskService.UpdateTask(ctx, taskService.UpdateTaskRequest{
		TaskID:      taskID,
		Title:       &title,
		Description: &description,
	})
	if err != nil {
		slog.Error("failed to updating task", "error", err)
		m.UI.Notification.Add(state.LevelError, "Error updating task")
		return nil
	}

	// update labels - need to handle this through detaching old and attaching new
	// First, get current labels
	taskDetail, err := m.App.TaskService.GetTaskDetail(ctx, taskID)
	if err != nil {
		slog.Error("failed to getting task detail for label sync", "error", err)
	} else {
		// Build current label map
		currentLabelMap := make(map[int]bool)
		for _, lbl := range taskDetail.Labels {
			currentLabelMap[lbl.ID] = true
		}

		// Build new label map
		newLabelMap := make(map[int]bool)
		for _, labelID := range values.labelIDs {
			newLabelMap[labelID] = true
		}

		// Detach labels that are no longer selected
		for _, lbl := range taskDetail.Labels {
			if !newLabelMap[lbl.ID] {
				if err := m.App.TaskService.DetachLabel(ctx, taskID, lbl.ID); err != nil {
					slog.Error("failed to detaching label", "error", err)
				}
			}
		}

		// Attach new labels
		for _, labelID := range values.labelIDs {
			if !currentLabelMap[labelID] {
				if err := m.App.TaskService.AttachLabel(ctx, taskID, labelID); err != nil {
					slog.Error("failed to attaching label", "error", err)
				}
			}
		}
	}

	m.syncParentRelationships(ctx, taskID)

	m.syncChildRelationships(ctx, taskID)

	// reload tasks respecting active filters
	project := m.getCurrentProject()
	if project != nil {
		tasksByColumn, err := m.fetchTasksForCurrentProject(ctx, project.ID)
		if err != nil {
			slog.Error("failed to reloading tasks", "error", err)
		} else {
			m.AppState.SetTasks(tasksByColumn)
		}
	}

	// Notify user if the updated task is hidden by active filters
	if m.UI.Filter.HasAnyFilter() {
		return m.addNotification(state.LevelWarning, "Task updated but hidden by active filters")
	}
	return nil
}

// applyParentRelationships applies parent relationships from ParentPickerState to a task
func (m *Model) applyParentRelationships(ctx context.Context, taskID int) {
	for _, item := range m.Pickers.Parent.Items {
		if item.Selected {
			relationTypeID := item.RelationTypeID
			if relationTypeID == 0 {
				relationTypeID = 1 // Default to Parent/Child
			}
			err := m.App.TaskService.AddParentRelation(ctx, taskID, item.TaskRef.ID, relationTypeID)
			if err != nil {
				slog.Error("failed to adding parent relationship", "error", err)
			}
		}
	}
}

// applyChildRelationships applies child relationships from ChildPickerState to a task
func (m *Model) applyChildRelationships(ctx context.Context, taskID int) {
	for _, item := range m.Pickers.Child.Items {
		if item.Selected {
			relationTypeID := item.RelationTypeID
			if relationTypeID == 0 {
				relationTypeID = 1 // Default to Parent/Child
			}
			err := m.App.TaskService.AddChildRelation(ctx, taskID, item.TaskRef.ID, relationTypeID)
			if err != nil {
				slog.Error("failed to adding child relationship", "error", err)
			}
		}
	}
}

// syncParentRelationships syncs parent relationships for an existing task by diffing current and new state.
func (m *Model) syncParentRelationships(ctx context.Context, taskID int) {
	// Get current parents from database
	taskDetail, err := m.App.TaskService.GetTaskDetail(ctx, taskID)
	if err != nil {
		slog.Error("failed to getting current parents", "error", err)
		return
	}
	currentParents := taskDetail.ParentTasks
	if currentParents == nil {
		currentParents = []*models.TaskReference{}
	}

	// Build maps for comparison (ID -> RelationTypeID)
	currentParentMap := make(map[int]int)
	for _, p := range currentParents {
		currentParentMap[p.ID] = p.RelationTypeID
	}

	newParentMap := make(map[int]int)
	for _, item := range m.Pickers.Parent.Items {
		if item.Selected {
			relationTypeID := item.RelationTypeID
			if relationTypeID == 0 {
				relationTypeID = 1 // Default to Parent/Child
			}
			newParentMap[item.TaskRef.ID] = relationTypeID
		}
	}

	// Remove parents that are no longer selected
	for parentID := range currentParentMap {
		if _, exists := newParentMap[parentID]; !exists {
			err = m.App.TaskService.RemoveParentRelation(ctx, taskID, parentID)
			if err != nil {
				slog.Error("failed to removing parent", "parentID", parentID, "error", err)
			}
		}
	}

	// Add or update parents
	for parentID, relationTypeID := range newParentMap {
		currentRelationType, exists := currentParentMap[parentID]
		if !exists || currentRelationType != relationTypeID {
			// Add new parent or update existing parent's relation type
			err = m.App.TaskService.AddParentRelation(ctx, taskID, parentID, relationTypeID)
			if err != nil {
				slog.Error("failed to adding/updating parent", "parentID", parentID, "error", err)
			}
		}
	}
}

// syncChildRelationships syncs child relationships for an existing task by diffing current and new state.
func (m *Model) syncChildRelationships(ctx context.Context, taskID int) {
	// Get current children from database
	taskDetail, err := m.App.TaskService.GetTaskDetail(ctx, taskID)
	if err != nil {
		slog.Error("failed to getting current children", "error", err)
		return
	}
	currentChildren := taskDetail.ChildTasks
	if currentChildren == nil {
		currentChildren = []*models.TaskReference{}
	}

	// Build maps for comparison (ID -> RelationTypeID)
	currentChildMap := make(map[int]int)
	for _, c := range currentChildren {
		currentChildMap[c.ID] = c.RelationTypeID
	}

	newChildMap := make(map[int]int)
	for _, item := range m.Pickers.Child.Items {
		if item.Selected {
			relationTypeID := item.RelationTypeID
			if relationTypeID == 0 {
				relationTypeID = 1 // Default to Parent/Child
			}
			newChildMap[item.TaskRef.ID] = relationTypeID
		}
	}

	// Remove children that are no longer selected
	for childID := range currentChildMap {
		if _, exists := newChildMap[childID]; !exists {
			err = m.App.TaskService.RemoveChildRelation(ctx, taskID, childID)
			if err != nil {
				slog.Error("failed to removing child", "childID", childID, "error", err)
			}
		}
	}

	// Add or update children
	for childID, relationTypeID := range newChildMap {
		currentRelationType, exists := currentChildMap[childID]
		if !exists || currentRelationType != relationTypeID {
			// Add new child or update existing child's relation type
			err = m.App.TaskService.AddChildRelation(ctx, taskID, childID, relationTypeID)
			if err != nil {
				slog.Error("failed to adding/updating child", "childID", childID, "error", err)
			}
		}
	}
}

// formConfig holds configuration for generic form handling
type formConfig struct {
	form       *huh.Form
	setForm    func(*huh.Form)
	clearForm  func()
	onComplete func() tea.Cmd // Called when form completes successfully, may return a cmd (e.g. notification)
	confirmPtr *bool          // Pointer to confirmation field for quick save
}

// handleFormUpdate processes form messages generically
func (m Model) handleFormUpdate(msg tea.Msg, cfg formConfig) (tea.Model, tea.Cmd) {
	if cfg.form == nil {
		m.UIState.Mode = state.NormalMode
		return m, nil
	}

	// Forward to form
	model, cmd := cfg.form.Update(msg)
	cfg.setForm(model.(*huh.Form))

	// Check completion
	if cfg.form.State == huh.StateCompleted {
		modeBeforeComplete := m.UIState.Mode
		completeCmd := cfg.onComplete()

		if m.UIState.Mode != modeBeforeComplete {
			return m, completeCmd
		}

		m.UIState.Mode = state.NormalMode
		cfg.setForm(nil)
		cfg.clearForm()

		if completeCmd != nil {
			return m, tea.Batch(tea.ClearScreen, completeCmd)
		}
		return m, tea.ClearScreen
	}

	// Note: StateAborted handling removed - ESC is now intercepted in updateTaskForm/updateProjectForm
	// to allow for change detection and discard confirmation

	return m, cmd
}

// handleFormSave handles the C-s save shortcut for forms.
// Sets confirmation to true and completes the form, triggering the save flow.
func (m Model) handleFormSave(cfg formConfig) (tea.Model, tea.Cmd) {
	if cfg.form == nil {
		m.UIState.Mode = state.NormalMode
		return m, nil
	}

	// Set confirmation to true (user wants to save)
	if cfg.confirmPtr != nil {
		*cfg.confirmPtr = true
	}

	// Mark form as completed to trigger onComplete callback
	cfg.form.State = huh.StateCompleted

	// Trigger the save logic
	completeCmd := cfg.onComplete()

	// Clean up and return to normal mode
	m.UIState.Mode = state.NormalMode
	cfg.setForm(nil)
	cfg.clearForm()

	if completeCmd != nil {
		return m, tea.Batch(tea.ClearScreen, completeCmd)
	}
	return m, tea.ClearScreen
}

// updateTaskForm handles all messages when in TaskFormMode
// This is separated out because forms need to receive ALL messages, not just KeyMsg
func (m Model) updateTaskForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Delegate viewport events first if viewport has focus
	if m.Forms.Form.ViewportFocused && m.Forms.Form.ViewportReady {
		var cmd tea.Cmd
		m.Forms.Form.CommentsViewport, cmd = m.Forms.Form.CommentsViewport.Update(msg)

		// Check if we should release focus back to form
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if keyMsg.String() == "tab" || keyMsg.String() == "shift+tab" {
				// Tab from viewport back to form
				m.Forms.Form.ViewportFocused = false
				return m, nil
			}
		}

		// If viewport consumed the message, return
		if cmd != nil {
			return m, cmd
		}
	}

	// Check for keyboard shortcuts before passing to form
	km := m.Config.KeyMappings
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "esc":
			// Check for changes before allowing abort
			if m.Forms.Form.HasTaskFormChanges() {
				// Show discard confirmation
				m.UIState.DiscardContext = &state.DiscardContext{
					SourceMode: state.TaskFormMode,
					Message:    "This task has unsaved changes. Discard?",
				}
				m.UIState.Mode = state.DiscardConfirmMode
				return m, nil
			}
			// No changes - allow immediate close
			m.UIState.Mode = state.NormalMode
			m.Forms.Form.ClearTaskForm()
			return m, tea.ClearScreen

		case km.Forms.EditParentTask:
			// Open parent picker
			if m.initParentPickerForForm() {
				m.UIState.Mode = state.ParentPickerMode
			}
			return m, nil

		case km.Forms.EditChildTask:
			// Open child picker
			if m.initChildPickerForForm() {
				m.UIState.Mode = state.ChildPickerMode
			}
			return m, nil

		case km.Forms.EditLabels:
			// Open label picker
			if m.initLabelPicker(state.TaskFormMode) {
				m.UIState.Mode = state.LabelPickerMode
			}
			return m, nil

		case km.Forms.EditPriority:
			// Open priority picker
			if m.initPriorityPicker(state.TaskFormMode) {
				m.UIState.Mode = state.PriorityPickerMode
			}
			return m, nil

		case km.Forms.EditType:
			// Open type picker
			if m.initTypePicker(state.TaskFormMode) {
				m.UIState.Mode = state.TypePickerMode
			}
			return m, nil

		case km.Forms.EditAssignee:
			// Open assignee picker
			if m.initAssigneePicker(state.TaskFormMode) {
				m.UIState.Mode = state.AssigneePickerMode
			}
			return m, nil

		case km.Forms.EditEstimate:
			// Open estimate input layer
			m.initEstimateInputForForm()
			m.UIState.Mode = state.EstimateInputMode
			return m, nil

		case km.Forms.EditDueDate:
			// Open due date picker
			m.initDatePickerForForm()
			m.UIState.Mode = state.DatePickerMode
			return m, nil

		case km.Forms.ShowHelp:
			// Open task form help menu
			m.UIState.Mode = state.TaskFormHelpMode
			return m, nil

		case "ctrl+down":
			// Focus comments viewport (explicit)
			if len(m.Forms.Form.FormComments) > 0 && m.Forms.Form.ViewportReady {
				m.Forms.Form.ViewportFocused = true
				m.Forms.Form.CommentsViewport.GotoBottom() // Start at most recent
				return m, nil
			}
			return m, nil

		case "down":
			// Auto-focus viewport on down arrow (implicit)
			if len(m.Forms.Form.FormComments) > 0 && !m.Forms.Form.ViewportFocused && m.Forms.Form.ViewportReady {
				m.Forms.Form.ViewportFocused = true
				m.Forms.Form.CommentsViewport.GotoBottom()
				// Let viewport handle the down arrow
				var cmd tea.Cmd
				m.Forms.Form.CommentsViewport, cmd = m.Forms.Form.CommentsViewport.Update(msg)
				return m, cmd
			}
			// Otherwise let form handle it

		case km.Forms.OpenCommentsView:
			// Open comments view
			return m.handleOpenCommentsView()

		case km.Forms.SaveForm:
			// Quick save via C-s
			return m.handleFormSave(formConfig{
				form: m.Forms.Form.TaskForm,
				setForm: func(f *huh.Form) {
					m.Forms.Form.TaskForm = f
				},
				clearForm: func() {
					m.Forms.Form.ClearTaskForm()
					m.Forms.Form.EditingTaskID = 0
				},
				onComplete: func() tea.Cmd {
					values := m.extractTaskFormValues()
					if !values.confirm {
						return nil
					}
					if values.title != "" {
						if m.Forms.Form.EditingTaskID == 0 {
							return m.createNewTaskWithLabelsAndRelationships(values)
						}
						return m.updateExistingTaskWithLabelsAndRelationships(values)
					}
					return nil
				},
				confirmPtr: &m.Forms.Form.FormConfirm,
			})
		}
	}

	// Pass through to existing form handler
	return m.handleFormUpdate(msg, formConfig{
		form: m.Forms.Form.TaskForm,
		setForm: func(f *huh.Form) {
			m.Forms.Form.TaskForm = f
		},
		clearForm: func() {
			m.Forms.Form.ClearTaskForm()
			m.Forms.Form.EditingTaskID = 0
		},
		onComplete: func() tea.Cmd {
			values := m.extractTaskFormValues()

			// Form submitted - check confirmation and save the task
			if !values.confirm {
				// User selected "No" on confirmation
				return nil
			}

			if values.title != "" {
				if m.Forms.Form.EditingTaskID == 0 {
					return m.createNewTaskWithLabelsAndRelationships(values)
				}
				return m.updateExistingTaskWithLabelsAndRelationships(values)
			}
			return nil
		},
		confirmPtr: &m.Forms.Form.FormConfirm,
	})
}

// updateProjectForm handles all messages when in ProjectFormMode or EditProjectFormMode
// This is separated out because forms need to receive ALL messages, not just KeyMsg
func (m Model) updateProjectForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Check for keyboard shortcuts before passing to form
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "esc":
			// Check for changes before allowing abort
			if m.Forms.Form.HasProjectFormChanges() {
				sourceMode := state.ProjectFormMode
				if m.UIState.Mode == state.EditProjectFormMode {
					sourceMode = state.EditProjectFormMode
				}
				m.UIState.DiscardContext = &state.DiscardContext{
					SourceMode: sourceMode,
					Message:    "Discard project changes?",
				}
				m.UIState.Mode = state.DiscardConfirmMode
				return m, nil
			}
			// No changes - allow immediate close
			m.UIState.Mode = state.NormalMode
			m.Forms.Form.ClearProjectForm()
			return m, tea.ClearScreen

		case m.Config.KeyMappings.Forms.SaveForm:
			return m.handleFormSave(formConfig{
				form: m.Forms.Form.ProjectForm,
				setForm: func(f *huh.Form) {
					m.Forms.Form.ProjectForm = f
				},
				clearForm: func() {
					m.Forms.Form.ClearProjectForm()
				},
				onComplete: func() tea.Cmd {
					m.submitProjectForm()
					return nil
				},
				confirmPtr: &m.Forms.Form.FormProjectConfirm,
			})

		case m.Config.KeyMappings.Forms.RefreshGitData:
			return m.handleRefreshGitData()
		}
	}

	return m.handleFormUpdate(msg, formConfig{
		form: m.Forms.Form.ProjectForm,
		setForm: func(f *huh.Form) {
			m.Forms.Form.ProjectForm = f
		},
		clearForm: func() {
			m.Forms.Form.ClearProjectForm()
		},
		onComplete: func() tea.Cmd {
			m.submitProjectForm()
			return nil
		},
	})
}

func (m *Model) submitProjectForm() {
	name := strings.TrimSpace(m.Forms.Form.FormProjectName)
	description := strings.TrimSpace(m.Forms.Form.FormProjectDescription)
	gitBranch := strings.TrimSpace(m.Forms.Form.FormProjectGitBranch)
	confirm := m.Forms.Form.FormProjectConfirm

	if !confirm || name == "" {
		return
	}

	if m.Forms.Form.Git.EditingProjectID != 0 {
		m.updateProject(name, description, gitBranch)
	} else {
		if gitBranch != "" {
			ctx, cancel := m.DBContext()
			defer cancel()
			existingProject, err := m.App.ProjectService.GetProjectByGitBranch(ctx, gitBranch)
			if err != nil {
				slog.Error("failed to check branch association", "branch", gitBranch, "error", err)
			}

			if existingProject != nil {
				m.UIState.ProjectBranchContext = &state.ProjectBranchContext{
					ProjectName:        name,
					ProjectDescription: description,
					GitBranch:          gitBranch,
					ExistingProject:    existingProject,
				}
				m.UIState.Mode = state.ProjectBranchConfirmMode
				return
			}
		}

		m.createProjectWithoutDialog(name, description, gitBranch)
	}
}

func (m *Model) updateProject(name, description, gitBranch string) {
	ctx, cancel := m.DBContext()
	defer cancel()

	var branchPtr *string
	if gitBranch != "" {
		branchPtr = &gitBranch
	} else {
		emptyBranch := ""
		branchPtr = &emptyBranch
	}

	err := m.App.ProjectService.UpdateProject(ctx, projectService.UpdateProjectRequest{
		ID:          m.Forms.Form.Git.EditingProjectID,
		Name:        &name,
		Description: &description,
		GitBranch:   branchPtr,
	})

	if err != nil {
		if errors.Is(err, projectService.ErrGitBranchAlreadyAssociated) && gitBranch != "" {
			existingProject, lookupErr := m.App.ProjectService.GetProjectByGitBranch(ctx, gitBranch)
			if lookupErr == nil && existingProject != nil {
				m.UI.Notification.Add(state.LevelError,
					fmt.Sprintf("Error: Branch already in use by project: %s", existingProject.Name))
			} else {
				m.UI.Notification.Add(state.LevelError, "Error: Branch already in use by another project")
			}
		} else {
			slog.Error("failed updating project", "error", err)
			m.UI.Notification.Add(state.LevelError, "Error updating project")
		}
	} else {
		m.reloadProjects()
		m.UI.Notification.Add(state.LevelInfo, "Project updated successfully")
	}
}

func (m *Model) createProjectWithoutDialog(name, description, gitBranch string) {
	ctx, cancel := m.DBContext()
	defer cancel()
	project, err := m.App.ProjectService.CreateProject(ctx, projectService.CreateProjectRequest{
		Name:        name,
		Description: description,
		GitBranch:   gitBranch,
	})
	if err != nil {
		if errors.Is(err, projectService.ErrGitBranchAlreadyAssociated) && gitBranch != "" {
			existingProject, lookupErr := m.App.ProjectService.GetProjectByGitBranch(ctx, gitBranch)
			if lookupErr == nil && existingProject != nil {
				m.UI.Notification.Add(state.LevelError,
					fmt.Sprintf("Error: Branch already in use by project: %s", existingProject.Name))
			} else {
				m.UI.Notification.Add(state.LevelError, "Error: Branch already in use by another project")
			}
		} else {
			slog.Error("failed to creating project", "error", err)
			m.UI.Notification.Add(state.LevelError, "Error creating project")
		}
	} else {
		m.reloadProjects()
		for i, p := range m.AppState.Projects() {
			if p.ID == project.ID {
				m.switchToProject(i)
				break
			}
		}
	}
}

func (m Model) handleRefreshGitData() (tea.Model, tea.Cmd) {
	workDir, err := os.Getwd()
	if err != nil {
		slog.Warn("failed to get working directory for cache invalidation", "error", err)
		m.UI.Notification.Add(state.LevelWarning, "Could not refresh git data")
		return m, nil
	}

	m.GitCache.Invalidate(workDir)
	slog.Info("git cache invalidated, reloading git data", "workDir", workDir)

	forEdit := m.UIState.Mode == state.EditProjectFormMode

	m.UIState.Mode = state.ProjectFormLoadingMode
	m.LoadingGitInfo = true
	m.LoadingBranches = true
	m.SpinnerFrame = 0

	return m, tea.Batch(
		loadGitDataForProjectFormCmd(m.Ctx, m.GitCache, forEdit),
		tickGitSpinner(),
	)
}

// updateColumnForm handles all messages when in AddColumnFormMode or EditColumnFormMode
// This is separated out because forms need to receive ALL messages, not just KeyMsg
func (m Model) updateColumnForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Check for keyboard shortcuts before passing to form
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "esc":
			// For edit mode, check for changes before allowing abort
			if m.UIState.Mode == state.EditColumnFormMode && m.Forms.Form.HasColumnFormChanges() {
				// Show discard confirmation
				m.UIState.DiscardContext = &state.DiscardContext{
					SourceMode: state.EditColumnFormMode,
					Message:    "Discard changes to column?",
				}
				m.UIState.Mode = state.DiscardConfirmMode
				return m, nil
			}
			// No changes or in create mode - allow immediate close
			m.UIState.Mode = state.NormalMode
			m.Forms.Form.ClearColumnForm()
			return m, tea.ClearScreen
		}
	}

	return m.handleFormUpdate(msg, formConfig{
		form: m.Forms.Form.ColumnForm,
		setForm: func(f *huh.Form) {
			m.Forms.Form.ColumnForm = f
		},
		clearForm: func() {
			m.Forms.Form.ClearColumnForm()
		},
		onComplete: func() tea.Cmd {
			// Read values from form state (forms update pointers in place)
			name := strings.TrimSpace(m.Forms.Form.FormColumnName)

			if name == "" {
				return nil
			}

			ctx, cancel := m.DBContext()
			defer cancel()

			if m.Forms.Form.EditingColumnID == 0 {
				// Create new column
				project := m.getCurrentProject()
				if project == nil {
					m.UI.Notification.Add(state.LevelError, "No project selected")
					return nil
				}

				_, err := m.App.ColumnService.CreateColumn(ctx, columnService.CreateColumnRequest{
					Name:      name,
					ProjectID: project.ID,
					AfterID:   nil, // Append to end
				})
				if err != nil {
					slog.Error("failed to creating column", "error", err)
					m.UI.Notification.Add(state.LevelError, "Error creating column")
					return nil
				}

				// Reload columns
				m.reloadCurrentProject()
			} else {
				// Rename existing column
				err := m.App.ColumnService.UpdateColumnName(ctx, m.Forms.Form.EditingColumnID, name)
				if err != nil {
					slog.Error("failed to renaming column", "error", err)
					m.UI.Notification.Add(state.LevelError, "Error renaming column")
					return nil
				}

				// Reload columns
				m.reloadCurrentProject()
			}
			return nil
		},
		confirmPtr: nil, // Column forms don't have confirmation field
	})
}

// updateCommentForm handles all messages when in CommentFormMode
// This is separated out because forms need to receive ALL messages, not just KeyMsg
func (m Model) updateCommentForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Check for keyboard shortcuts before passing to form
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "esc":
			// For edit mode, check for changes before allowing abort
			if m.Forms.Form.EditingCommentID != 0 && m.Forms.Form.HasCommentFormChanges() {
				// Show discard confirmation
				m.UIState.DiscardContext = &state.DiscardContext{
					SourceMode: state.CommentFormMode,
					Message:    "Discard changes to comment?",
				}
				m.UIState.Mode = state.DiscardConfirmMode
				return m, nil
			}
			// No changes or in create mode - return to appropriate mode
			returnMode := m.Forms.Form.CommentFormReturnMode
			if returnMode == state.CommentsViewMode {
				// Reload activities for the comments view
				ctx, cancel := m.DBContext()
				activities, err := m.App.TaskService.GetTaskActivities(ctx, m.Forms.Comment.TaskID)
				cancel()
				if err != nil {
					m.UI.Notification.Add(state.LevelError, "Failed to refresh activity")
				} else {
					m.Forms.Comment.SetActivities(activities)
				}
				m.UIState.Mode = state.CommentsViewMode
			} else {
				m.UIState.Mode = state.TaskFormMode
			}
			m.Forms.Form.ClearCommentForm()
			return m, tea.ClearScreen
		}
	}

	return m.handleFormUpdate(msg, formConfig{
		form: m.Forms.Form.CommentForm,
		setForm: func(f *huh.Form) {
			m.Forms.Form.CommentForm = f
		},
		clearForm: func() {
			m.Forms.Form.ClearCommentForm()
		},
		onComplete: func() tea.Cmd {
			// Read values from form state (forms update pointers in place)
			message := strings.TrimSpace(m.Forms.Form.FormCommentMessage)

			if message == "" {
				return nil
			}

			ctx, cancel := m.DBContext()
			defer cancel()

			if m.Forms.Form.EditingCommentID == 0 {
				// Create new comment
				taskID := m.Forms.Form.EditingTaskID
				if taskID == 0 {
					m.UI.Notification.Add(state.LevelError, "No task selected")
					return nil
				}

				_, err := m.App.TaskService.CreateComment(ctx, taskService.CreateCommentRequest{
					TaskID:  taskID,
					Message: message,
					Author:  userutil.GetCurrentUsername(),
				})
				if err != nil {
					slog.Error("failed to creating comment", "error", err)
					m.UI.Notification.Add(state.LevelError, "Error creating comment")
					return nil
				}

				m.UI.Notification.Add(state.LevelInfo, "Comment added")
			} else {
				// Update existing comment
				err := m.App.TaskService.UpdateComment(ctx, taskService.UpdateCommentRequest{
					CommentID: m.Forms.Form.EditingCommentID,
					Message:   message,
				})
				if err != nil {
					slog.Error("failed to updating comment", "error", err)
					m.UI.Notification.Add(state.LevelError, "Error updating comment")
					return nil
				}

				m.UI.Notification.Add(state.LevelInfo, "Comment updated")
			}

			// Reload comments for task form viewport
			taskID := m.Forms.Form.EditingTaskID
			comments, err := m.App.TaskService.GetCommentsByTask(ctx, taskID)
			if err != nil {
				slog.Error("failed to reloading comments", "error", err)
				m.UI.Notification.Add(state.LevelError, "Failed to reload comments")
			} else {
				m.Forms.Form.FormComments = comments
			}

			// Refresh activities for task form preview and comments view
			activities, actErr := m.App.TaskService.GetTaskActivities(ctx, taskID)
			if actErr == nil {
				m.Forms.Form.FormActivities = activities
				m.Forms.Comment.SetActivities(activities)
			}

			// Return to appropriate mode based on where we came from
			returnMode := m.Forms.Form.CommentFormReturnMode
			if returnMode == state.CommentsViewMode {
				m.UIState.Mode = state.CommentsViewMode
			} else {
				m.UIState.Mode = state.TaskFormMode
			}
			return nil
		},
		confirmPtr: nil, // Comment forms don't have confirmation field
	})
}

// handleOpenCommentsView opens the full-screen comments view
func (m Model) handleOpenCommentsView() (tea.Model, tea.Cmd) {
	taskID := m.Forms.Form.EditingTaskID
	if taskID == 0 {
		m.UI.Notification.Add(state.LevelError, "No task selected")
		return m, nil
	}

	// Fetch activities (events + comments) from the service
	ctx, cancel := m.DBContext()
	defer cancel()

	activities, err := m.App.TaskService.GetTaskActivities(ctx, taskID)
	if err != nil {
		slog.Error("failed to fetch task activities", "error", err)
		m.UI.Notification.Add(state.LevelError, "Failed to load activity")
		return m, nil
	}

	// Set up comments view state with activities
	m.Forms.Comment.TaskID = taskID
	m.Forms.Comment.SetActivities(activities)
	m.Forms.Comment.Cursor = 0
	m.Forms.Comment.ScrollOffset = 0

	// Switch to comments view mode
	m.UIState.Mode = state.CommentsViewMode

	return m, nil
}
