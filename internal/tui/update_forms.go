package tui

import (
	"context"
	"log/slog"
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
func (m *Model) createNewTaskWithLabelsAndRelationships(values taskFormValues) {
	currentCol := m.getCurrentColumn()
	if currentCol == nil {
		m.UI.Notification.Add(state.LevelError, "No column selected")
		return
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

	// 1. Create the task with all data in one call
	task, err := m.App.TaskService.CreateTask(ctx, taskService.CreateTaskRequest{
		Title:       values.title,
		Description: values.description,
		ColumnID:    currentCol.ID,
		Position:    len(m.getTasksForColumn(currentCol.ID)),
		LabelIDs:    values.labelIDs,
		ParentIDs:   parentIDs,
		ChildIDs:    childIDs,
	})
	if err != nil {
		slog.Error("failed to creating task", "error", err)
		m.UI.Notification.Add(state.LevelError, "Error creating task")
		return
	}

	// 2. Apply parent relationships with correct relation types
	m.applyParentRelationships(ctx, task.ID)

	// 3. Apply child relationships with correct relation types
	m.applyChildRelationships(ctx, task.ID)

	// 4. Reload all tasks for the project to keep state consistent
	project := m.getCurrentProject()
	if project != nil {
		tasksByColumn, err := m.App.TaskService.GetTaskSummariesByProject(ctx, project.ID)
		if err != nil {
			slog.Error("failed to reloading tasks", "error", err)
		} else {
			m.AppState.SetTasks(tasksByColumn)
		}
	}
}

// updateExistingTaskWithLabelsAndRelationships updates task, labels, and parent/child relationships
func (m *Model) updateExistingTaskWithLabelsAndRelationships(values taskFormValues) {
	// create context for database operations
	ctx, cancel := m.DBContext()
	defer cancel()
	taskID := m.Forms.Form.EditingTaskID

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
		return
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

	// reload all tasks for the project to keep state consistent
	project := m.getCurrentProject()
	if project != nil {
		tasksByColumn, err := m.App.TaskService.GetTaskSummariesByProject(ctx, project.ID)
		if err != nil {
			slog.Error("failed to reloading tasks", "error", err)
		} else {
			m.AppState.SetTasks(tasksByColumn)
		}
	}
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
	onComplete func() // Called when form completes successfully
	confirmPtr *bool  // Pointer to confirmation field for quick save
}

// handleFormUpdate processes form messages generically
func (m Model) handleFormUpdate(msg tea.Msg, cfg formConfig) (tea.Model, tea.Cmd) {
	if cfg.form == nil {
		m.UIState.SetMode(state.NormalMode)
		return m, nil
	}

	// Forward to form
	model, cmd := cfg.form.Update(msg)
	cfg.setForm(model.(*huh.Form))

	// Check completion
	if cfg.form.State == huh.StateCompleted {
		cfg.onComplete()
		m.UIState.SetMode(state.NormalMode)
		cfg.setForm(nil)
		cfg.clearForm()
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
		m.UIState.SetMode(state.NormalMode)
		return m, nil
	}

	// Set confirmation to true (user wants to save)
	if cfg.confirmPtr != nil {
		*cfg.confirmPtr = true
	}

	// Mark form as completed to trigger onComplete callback
	cfg.form.State = huh.StateCompleted

	// Trigger the save logic
	cfg.onComplete()

	// Clean up and return to normal mode
	m.UIState.SetMode(state.NormalMode)
	cfg.setForm(nil)
	cfg.clearForm()

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
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "esc":
			// Check for changes before allowing abort
			if m.Forms.Form.HasTaskFormChanges() {
				// Show discard confirmation
				m.UIState.SetDiscardContext(&state.DiscardContext{
					SourceMode: state.TicketFormMode,
					Message:    "This task has unsaved changes. Discard?",
				})
				m.UIState.SetMode(state.DiscardConfirmMode)
				return m, nil
			}
			// No changes - allow immediate close
			m.UIState.SetMode(state.NormalMode)
			m.Forms.Form.ClearTaskForm()
			return m, tea.ClearScreen

		case "ctrl+p":
			// Open parent picker
			if m.initParentPickerForForm() {
				m.UIState.SetMode(state.ParentPickerMode)
			}
			return m, nil

		case "ctrl+c":
			// Open child picker
			if m.initChildPickerForForm() {
				m.UIState.SetMode(state.ChildPickerMode)
			}
			return m, nil

		case "ctrl+l":
			// Open label picker
			if m.initLabelPickerForForm() {
				m.UIState.SetMode(state.LabelPickerMode)
			}
			return m, nil

		case "ctrl+r":
			// Open priority picker
			if m.initPriorityPickerForForm() {
				m.UIState.SetMode(state.PriorityPickerMode)
			}
			return m, nil

		case "ctrl+t":
			// Open type picker
			if m.initTypePickerForForm() {
				m.UIState.SetMode(state.TypePickerMode)
			}
			return m, nil

		case "ctrl+h":
			// Open task form help menu
			m.UIState.SetMode(state.TaskFormHelpMode)
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

		case "ctrl+n":
			// Open comments view
			return m.handleOpenCommentsView()

		case m.Config.KeyMappings.SaveForm:
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
				onComplete: func() {
					values := m.extractTaskFormValues()
					if !values.confirm {
						return
					}
					if values.title != "" {
						if m.Forms.Form.EditingTaskID == 0 {
							m.createNewTaskWithLabelsAndRelationships(values)
						} else {
							m.updateExistingTaskWithLabelsAndRelationships(values)
						}
					}
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
		onComplete: func() {
			values := m.extractTaskFormValues()

			// Form submitted - check confirmation and save the task
			if !values.confirm {
				// User selected "No" on confirmation
				return
			}

			if values.title != "" {
				if m.Forms.Form.EditingTaskID == 0 {
					m.createNewTaskWithLabelsAndRelationships(values)
				} else {
					m.updateExistingTaskWithLabelsAndRelationships(values)
				}
			}
		},
		confirmPtr: &m.Forms.Form.FormConfirm,
	})
}

// updateProjectForm handles all messages when in ProjectFormMode
// This is separated out because forms need to receive ALL messages, not just KeyMsg
func (m Model) updateProjectForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Check for keyboard shortcuts before passing to form
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "esc":
			// Check for changes before allowing abort
			if m.Forms.Form.HasProjectFormChanges() {
				// Show discard confirmation
				m.UIState.SetDiscardContext(&state.DiscardContext{
					SourceMode: state.ProjectFormMode,
					Message:    "Discard project?",
				})
				m.UIState.SetMode(state.DiscardConfirmMode)
				return m, nil
			}
			// No changes - allow immediate close
			m.UIState.SetMode(state.NormalMode)
			m.Forms.Form.ClearProjectForm()
			return m, tea.ClearScreen

		case m.Config.KeyMappings.SaveForm:
			return m.handleFormSave(formConfig{
				form: m.Forms.Form.ProjectForm,
				setForm: func(f *huh.Form) {
					m.Forms.Form.ProjectForm = f
				},
				clearForm: func() {
					m.Forms.Form.ClearProjectForm()
				},
				onComplete: func() {
					name := strings.TrimSpace(m.Forms.Form.FormProjectName)
					description := strings.TrimSpace(m.Forms.Form.FormProjectDescription)
					confirm := m.Forms.Form.FormProjectConfirm

					if !confirm {
						return
					}

					if name != "" {
						ctx, cancel := m.DBContext()
						defer cancel()
						project, err := m.App.ProjectService.CreateProject(ctx, projectService.CreateProjectRequest{
							Name:        name,
							Description: description,
						})
						if err != nil {
							slog.Error("failed to creating project", "error", err)
							m.UI.Notification.Add(state.LevelError, "Error creating project")
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
				},
				confirmPtr: &m.Forms.Form.FormProjectConfirm,
			})
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
		onComplete: func() {
			// Read values from form state (forms update pointers in place)
			name := strings.TrimSpace(m.Forms.Form.FormProjectName)
			description := strings.TrimSpace(m.Forms.Form.FormProjectDescription)
			confirm := m.Forms.Form.FormProjectConfirm

			// Form submitted - check confirmation and create the project
			if !confirm {
				// User selected "No" on confirmation
				return
			}

			if name != "" {
				ctx, cancel := m.DBContext()
				defer cancel()
				project, err := m.App.ProjectService.CreateProject(ctx, projectService.CreateProjectRequest{
					Name:        name,
					Description: description,
				})
				if err != nil {
					slog.Error("failed to creating project", "error", err)
					m.UI.Notification.Add(state.LevelError, "Error creating project")
				} else {
					// Reload projects list
					m.reloadProjects()

					// Switch to the new project
					for i, p := range m.AppState.Projects() {
						if p.ID == project.ID {
							m.switchToProject(i)
							break
						}
					}
				}
			}
		},
	})
}

// updateColumnForm handles all messages when in AddColumnFormMode or EditColumnFormMode
// This is separated out because forms need to receive ALL messages, not just KeyMsg
func (m Model) updateColumnForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Check for keyboard shortcuts before passing to form
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "esc":
			// For edit mode, check for changes before allowing abort
			if m.UIState.Mode() == state.EditColumnFormMode && m.Forms.Form.HasColumnFormChanges() {
				// Show discard confirmation
				m.UIState.SetDiscardContext(&state.DiscardContext{
					SourceMode: state.EditColumnFormMode,
					Message:    "Discard changes to column?",
				})
				m.UIState.SetMode(state.DiscardConfirmMode)
				return m, nil
			}
			// No changes or in create mode - allow immediate close
			m.UIState.SetMode(state.NormalMode)
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
		onComplete: func() {
			// Read values from form state (forms update pointers in place)
			name := strings.TrimSpace(m.Forms.Form.FormColumnName)

			if name == "" {
				return
			}

			ctx, cancel := m.DBContext()
			defer cancel()

			if m.Forms.Form.EditingColumnID == 0 {
				// Create new column
				project := m.getCurrentProject()
				if project == nil {
					m.UI.Notification.Add(state.LevelError, "No project selected")
					return
				}

				_, err := m.App.ColumnService.CreateColumn(ctx, columnService.CreateColumnRequest{
					Name:      name,
					ProjectID: project.ID,
					AfterID:   nil, // Append to end
				})
				if err != nil {
					slog.Error("failed to creating column", "error", err)
					m.UI.Notification.Add(state.LevelError, "Error creating column")
					return
				}

				// Reload columns
				m.reloadCurrentProject()
			} else {
				// Rename existing column
				err := m.App.ColumnService.UpdateColumnName(ctx, m.Forms.Form.EditingColumnID, name)
				if err != nil {
					slog.Error("failed to renaming column", "error", err)
					m.UI.Notification.Add(state.LevelError, "Error renaming column")
					return
				}

				// Reload columns
				m.reloadCurrentProject()
			}
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
				m.UIState.SetDiscardContext(&state.DiscardContext{
					SourceMode: state.CommentFormMode,
					Message:    "Discard changes to comment?",
				})
				m.UIState.SetMode(state.DiscardConfirmMode)
				return m, nil
			}
			// No changes or in create mode - return to appropriate mode
			returnMode := m.Forms.Form.CommentFormReturnMode
			if returnMode == state.CommentsViewMode {
				m.Forms.Comment.SetComments(m.Forms.Form.FormComments)
				m.UIState.SetMode(state.CommentsViewMode)
			} else {
				m.UIState.SetMode(state.TicketFormMode)
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
		onComplete: func() {
			// Read values from form state (forms update pointers in place)
			message := strings.TrimSpace(m.Forms.Form.FormCommentMessage)

			if message == "" {
				return
			}

			ctx, cancel := m.DBContext()
			defer cancel()

			if m.Forms.Form.EditingCommentID == 0 {
				// Create new comment
				taskID := m.Forms.Form.EditingTaskID
				if taskID == 0 {
					m.UI.Notification.Add(state.LevelError, "No task selected")
					return
				}

				_, err := m.App.TaskService.CreateComment(ctx, taskService.CreateCommentRequest{
					TaskID:  taskID,
					Message: message,
					Author:  userutil.GetCurrentUsername(),
				})
				if err != nil {
					slog.Error("failed to creating comment", "error", err)
					m.UI.Notification.Add(state.LevelError, "Error creating comment")
					return
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
					return
				}

				m.UI.Notification.Add(state.LevelInfo, "Comment updated")
			}

			// Reload comments
			taskID := m.Forms.Form.EditingTaskID
			comments, err := m.App.TaskService.GetCommentsByTask(ctx, taskID)
			if err != nil {
				slog.Error("failed to reloading comments", "error", err)
				m.UI.Notification.Add(state.LevelError, "Failed to reload comments")
			} else {
				m.Forms.Form.FormComments = comments
				m.Forms.Comment.Items = convertToCommentItems(comments)
			}

			// Return to appropriate mode based on where we came from
			returnMode := m.Forms.Form.CommentFormReturnMode
			if returnMode == state.CommentsViewMode {
				// Refresh comments view and return to it
				m.Forms.Comment.SetComments(m.Forms.Form.FormComments)
				m.UIState.SetMode(state.CommentsViewMode)
			} else {
				m.UIState.SetMode(state.TicketFormMode)
			}
		},
		confirmPtr: nil, // Comment forms don't have confirmation field
	})
}

// handleOpenCommentsView opens the full-screen comments view
func (m Model) handleOpenCommentsView() (tea.Model, tea.Cmd) {
	// Set up comments view state
	m.Forms.Comment.TaskID = m.Forms.Form.EditingTaskID
	m.Forms.Comment.SetComments(m.Forms.Form.FormComments)
	m.Forms.Comment.Cursor = 0
	m.Forms.Comment.ScrollOffset = 0

	// Switch to comments view mode
	m.UIState.SetMode(state.CommentsViewMode)

	return m, nil
}
