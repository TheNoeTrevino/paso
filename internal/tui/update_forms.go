package tui

import (
	"context"
	"log/slog"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

type ticketFormValues struct {
	title       string
	description string
	confirm     bool
	labelIDs    []int
}

// extractTicketFormValues extracts and returns form values from the ticket form
// Since our forms update pointers in place, we can just read from formState
func (m Model) extractTicketFormValues() ticketFormValues {
	return ticketFormValues{
		title:       strings.TrimSpace(m.FormState.FormTitle),
		description: strings.TrimSpace(m.FormState.FormDescription),
		confirm:     m.FormState.FormConfirm,
		labelIDs:    m.FormState.FormLabelIDs,
	}
}

// createNewTaskWithLabelsAndRelationships creates a new task, sets labels, and applies parent/child relationships
func (m *Model) createNewTaskWithLabelsAndRelationships(values ticketFormValues) {
	currentCol := m.getCurrentColumn()
	if currentCol == nil {
		m.NotificationState.Add(state.LevelError, "No column selected")
		return
	}

	// Create context for database operations
	ctx, cancel := m.DbContext()
	defer cancel()

	// 1. Create the task
	task, err := m.Repo.CreateTask(ctx,
		values.title,
		values.description,
		currentCol.ID,
		len(m.getTasksForColumn(currentCol.ID)),
	)
	if err != nil {
		slog.Error("Error creating task", "error", err)
		m.NotificationState.Add(state.LevelError, "Error creating task")
		return
	}

	// 2. Set labels
	if len(values.labelIDs) > 0 {
		err = m.Repo.SetTaskLabels(ctx, task.ID, values.labelIDs)
		if err != nil {
			slog.Error("Error setting labels", "error", err)
		}
	}

	m.applyParentRelationships(ctx, task.ID)

	m.applyChildRelationships(ctx, task.ID)

	// 5. Reload all tasks for the project to keep state consistent
	project := m.getCurrentProject()
	if project != nil {
		tasksByColumn, err := m.Repo.GetTaskSummariesByProject(ctx, project.ID)
		if err != nil {
			slog.Error("Error reloading tasks", "error", err)
		} else {
			m.AppState.SetTasks(tasksByColumn)
		}
	}
}

// updateExistingTaskWithLabelsAndRelationships updates task, labels, and parent/child relationships
func (m *Model) updateExistingTaskWithLabelsAndRelationships(values ticketFormValues) {
	// create context for database operations
	ctx, cancel := m.DbContext()
	defer cancel()
	taskID := m.FormState.EditingTaskID

	// update task basic fields
	err := m.Repo.UpdateTask(ctx, taskID, values.title, values.description)
	if err != nil {
		slog.Error("Error updating task", "error", err)
		m.NotificationState.Add(state.LevelError, "Error updating task")
		return
	}

	// update labels
	err = m.Repo.SetTaskLabels(ctx, taskID, values.labelIDs)
	if err != nil {
		slog.Error("Error setting labels", "error", err)
	}

	m.syncParentRelationships(ctx, taskID)

	m.syncChildRelationships(ctx, taskID)

	// reload all tasks for the project to keep state consistent
	project := m.getCurrentProject()
	if project != nil {
		tasksByColumn, err := m.Repo.GetTaskSummariesByProject(ctx, project.ID)
		if err != nil {
			slog.Error("Error reloading tasks", "error", err)
		} else {
			m.AppState.SetTasks(tasksByColumn)
		}
	}
}

// applyParentRelationships applies parent relationships from ParentPickerState to a task
func (m *Model) applyParentRelationships(ctx context.Context, taskID int) {
	for _, item := range m.ParentPickerState.Items {
		if item.Selected {
			relationTypeID := item.RelationTypeID
			if relationTypeID == 0 {
				relationTypeID = 1 // Default to Parent/Child
			}
			err := m.Repo.AddSubtaskWithRelationType(ctx, item.TaskRef.ID, taskID, relationTypeID)
			if err != nil {
				slog.Error("Error adding parent relationship", "error", err)
			}
		}
	}
}

// applyChildRelationships applies child relationships from ChildPickerState to a task
func (m *Model) applyChildRelationships(ctx context.Context, taskID int) {
	for _, item := range m.ChildPickerState.Items {
		if item.Selected {
			relationTypeID := item.RelationTypeID
			if relationTypeID == 0 {
				relationTypeID = 1 // Default to Parent/Child
			}
			err := m.Repo.AddSubtaskWithRelationType(ctx, taskID, item.TaskRef.ID, relationTypeID)
			if err != nil {
				slog.Error("Error adding child relationship", "error", err)
			}
		}
	}
}

// syncParentRelationships syncs parent relationships for an existing task by diffing current and new state.
func (m *Model) syncParentRelationships(ctx context.Context, taskID int) {
	// Get current parents from database
	currentParents, err := m.Repo.GetParentTasks(ctx, taskID)
	if err != nil {
		slog.Error("Error getting current parents", "error", err)
		currentParents = []*models.TaskReference{}
	}

	// Build maps for comparison (ID -> RelationTypeID)
	currentParentMap := make(map[int]int)
	for _, p := range currentParents {
		currentParentMap[p.ID] = p.RelationTypeID
	}

	newParentMap := make(map[int]int)
	for _, item := range m.ParentPickerState.Items {
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
			err = m.Repo.RemoveSubtask(ctx, parentID, taskID)
			if err != nil {
				slog.Error("Error removing parent", "parentID", parentID, "error", err)
			}
		}
	}

	// Add or update parents (AddSubtaskWithRelationType uses INSERT OR REPLACE)
	for parentID, relationTypeID := range newParentMap {
		currentRelationType, exists := currentParentMap[parentID]
		if !exists || currentRelationType != relationTypeID {
			// Add new parent or update existing parent's relation type
			err = m.Repo.AddSubtaskWithRelationType(ctx, parentID, taskID, relationTypeID)
			if err != nil {
				slog.Error("Error adding/updating parent", "parentID", parentID, "error", err)
			}
		}
	}
}

// syncChildRelationships syncs child relationships for an existing task by diffing current and new state.
func (m *Model) syncChildRelationships(ctx context.Context, taskID int) {
	// Get current children from database
	currentChildren, err := m.Repo.GetChildTasks(ctx, taskID)
	if err != nil {
		slog.Error("Error getting current children", "error", err)
		currentChildren = []*models.TaskReference{}
	}

	// Build maps for comparison (ID -> RelationTypeID)
	currentChildMap := make(map[int]int)
	for _, c := range currentChildren {
		currentChildMap[c.ID] = c.RelationTypeID
	}

	newChildMap := make(map[int]int)
	for _, item := range m.ChildPickerState.Items {
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
			err = m.Repo.RemoveSubtask(ctx, taskID, childID)
			if err != nil {
				slog.Error("Error removing child", "childID", childID, "error", err)
			}
		}
	}

	// Add or update children (AddSubtaskWithRelationType uses INSERT OR REPLACE)
	for childID, relationTypeID := range newChildMap {
		currentRelationType, exists := currentChildMap[childID]
		if !exists || currentRelationType != relationTypeID {
			// Add new child or update existing child's relation type
			err = m.Repo.AddSubtaskWithRelationType(ctx, taskID, childID, relationTypeID)
			if err != nil {
				slog.Error("Error adding/updating child", "childID", childID, "error", err)
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
		m.UiState.SetMode(state.NormalMode)
		return m, nil
	}

	// Forward to form
	model, cmd := cfg.form.Update(msg)
	cfg.setForm(model.(*huh.Form))

	// Check completion
	if cfg.form.State == huh.StateCompleted {
		cfg.onComplete()
		m.UiState.SetMode(state.NormalMode)
		cfg.setForm(nil)
		cfg.clearForm()
		return m, tea.ClearScreen
	}

	// Note: StateAborted handling removed - ESC is now intercepted in updateTicketForm/updateProjectForm
	// to allow for change detection and discard confirmation

	return m, cmd
}

// handleFormSave handles the C-s save shortcut for forms.
// Sets confirmation to true and completes the form, triggering the save flow.
func (m Model) handleFormSave(cfg formConfig) (tea.Model, tea.Cmd) {
	if cfg.form == nil {
		m.UiState.SetMode(state.NormalMode)
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
	m.UiState.SetMode(state.NormalMode)
	cfg.setForm(nil)
	cfg.clearForm()

	return m, tea.ClearScreen
}

// updateTicketForm handles all messages when in TicketFormMode
// This is separated out because forms need to receive ALL messages, not just KeyMsg
func (m Model) updateTicketForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Check for keyboard shortcuts before passing to form
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "esc":
			// Check for changes before allowing abort
			if m.FormState.HasTicketFormChanges() {
				// Show discard confirmation
				m.UiState.SetDiscardContext(&state.DiscardContext{
					SourceMode: state.TicketFormMode,
					Message:    "Discard task?",
				})
				m.UiState.SetMode(state.DiscardConfirmMode)
				return m, nil
			}
			// No changes - allow immediate close
			m.UiState.SetMode(state.NormalMode)
			m.FormState.ClearTicketForm()
			return m, tea.ClearScreen

		case "ctrl+p":
			// Open parent picker
			if m.initParentPickerForForm() {
				m.UiState.SetMode(state.ParentPickerMode)
			}
			return m, nil

		case "ctrl+c":
			// Open child picker
			if m.initChildPickerForForm() {
				m.UiState.SetMode(state.ChildPickerMode)
			}
			return m, nil

		case "ctrl+l":
			// Open label picker
			if m.initLabelPickerForForm() {
				m.UiState.SetMode(state.LabelPickerMode)
			}
			return m, nil

		case "ctrl+r":
			// Open priority picker
			if m.initPriorityPickerForForm() {
				m.UiState.SetMode(state.PriorityPickerMode)
			}
			return m, nil

		case "ctrl+t":
			// Open type picker
			if m.initTypePickerForForm() {
				m.UiState.SetMode(state.TypePickerMode)
			}
			return m, nil

		case m.Config.KeyMappings.SaveForm:
			// Quick save via C-s
			return m.handleFormSave(formConfig{
				form: m.FormState.TicketForm,
				setForm: func(f *huh.Form) {
					m.FormState.TicketForm = f
				},
				clearForm: func() {
					m.FormState.ClearTicketForm()
					m.FormState.EditingTaskID = 0
				},
				onComplete: func() {
					values := m.extractTicketFormValues()
					if !values.confirm {
						return
					}
					if values.title != "" {
						if m.FormState.EditingTaskID == 0 {
							m.createNewTaskWithLabelsAndRelationships(values)
						} else {
							m.updateExistingTaskWithLabelsAndRelationships(values)
						}
					}
				},
				confirmPtr: &m.FormState.FormConfirm,
			})
		}
	}

	// Pass through to existing form handler
	return m.handleFormUpdate(msg, formConfig{
		form: m.FormState.TicketForm,
		setForm: func(f *huh.Form) {
			m.FormState.TicketForm = f
		},
		clearForm: func() {
			m.FormState.ClearTicketForm()
			m.FormState.EditingTaskID = 0
		},
		onComplete: func() {
			values := m.extractTicketFormValues()

			// Form submitted - check confirmation and save the task
			if !values.confirm {
				// User selected "No" on confirmation
				return
			}

			if values.title != "" {
				if m.FormState.EditingTaskID == 0 {
					m.createNewTaskWithLabelsAndRelationships(values)
				} else {
					m.updateExistingTaskWithLabelsAndRelationships(values)
				}
			}
		},
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
			if m.FormState.HasProjectFormChanges() {
				// Show discard confirmation
				m.UiState.SetDiscardContext(&state.DiscardContext{
					SourceMode: state.ProjectFormMode,
					Message:    "Discard project?",
				})
				m.UiState.SetMode(state.DiscardConfirmMode)
				return m, nil
			}
			// No changes - allow immediate close
			m.UiState.SetMode(state.NormalMode)
			m.FormState.ClearProjectForm()
			return m, tea.ClearScreen

		case m.Config.KeyMappings.SaveForm:
			return m.handleFormSave(formConfig{
				form: m.FormState.ProjectForm,
				setForm: func(f *huh.Form) {
					m.FormState.ProjectForm = f
				},
				clearForm: func() {
					m.FormState.ClearProjectForm()
				},
				onComplete: func() {
					name := strings.TrimSpace(m.FormState.FormProjectName)
					description := strings.TrimSpace(m.FormState.FormProjectDescription)
					confirm := m.FormState.FormProjectConfirm

					if !confirm {
						return
					}

					if name != "" {
						ctx, cancel := m.DbContext()
						defer cancel()
						project, err := m.Repo.CreateProject(ctx, name, description)
						if err != nil {
							slog.Error("Error creating project", "error", err)
							m.NotificationState.Add(state.LevelError, "Error creating project")
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
				confirmPtr: &m.FormState.FormProjectConfirm,
			})
		}
	}

	return m.handleFormUpdate(msg, formConfig{
		form: m.FormState.ProjectForm,
		setForm: func(f *huh.Form) {
			m.FormState.ProjectForm = f
		},
		clearForm: func() {
			m.FormState.ClearProjectForm()
		},
		onComplete: func() {
			// Read values from form state (forms update pointers in place)
			name := strings.TrimSpace(m.FormState.FormProjectName)
			description := strings.TrimSpace(m.FormState.FormProjectDescription)
			confirm := m.FormState.FormProjectConfirm

			// Form submitted - check confirmation and create the project
			if !confirm {
				// User selected "No" on confirmation
				return
			}

			if name != "" {
				ctx, cancel := m.DbContext()
				defer cancel()
				project, err := m.Repo.CreateProject(ctx, name, description)
				if err != nil {
					slog.Error("Error creating project", "error", err)
					m.NotificationState.Add(state.LevelError, "Error creating project")
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
