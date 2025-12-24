package handlers

import (
	"context"
	"log/slog"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui"
	"github.com/thenoetrevino/paso/internal/tui/renderers"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// Timeout constant for database operations
const timeoutDB = 30 * time.Second

// ticketFormValues holds the extracted values from a completed ticket form
type ticketFormValues struct {
	title       string
	description string
	confirm     bool
	labelIDs    []int
}

// projectFormValues holds the extracted values from a completed project form
type projectFormValues struct {
	name        string
	description string
	confirm     bool
}

// formConfig holds configuration for generic form handling
type formConfig struct {
	form       *huh.Form
	setForm    func(*huh.Form)
	clearForm  func()
	onComplete func() // Called when form completes successfully
	confirmPtr *bool  // Pointer to confirmation field for quick save
}

// Helper functions for model operations

// dbContext creates a child context with timeout for database operations
func dbContext(m *tui.Model) (context.Context, context.CancelFunc) {
	return context.WithTimeout(m.Ctx, timeoutDB)
}

// getCurrentColumn returns the currently selected column
func getCurrentColumn(m *tui.Model) *models.Column {
	if len(m.AppState.Columns()) == 0 {
		return nil
	}
	selectedIdx := m.UiState.SelectedColumn()
	if selectedIdx < 0 || selectedIdx >= len(m.AppState.Columns()) {
		return nil
	}
	return m.AppState.Columns()[selectedIdx]
}

// getTasksForColumn returns tasks for a specific column ID with safe map access
func getTasksForColumn(m *tui.Model, columnID int) []*models.TaskSummary {
	tasks, ok := m.AppState.Tasks()[columnID]
	if !ok || tasks == nil {
		return []*models.TaskSummary{}
	}
	return tasks
}

// getCurrentProject returns the currently selected project
func getCurrentProject(m *tui.Model) *models.Project {
	return m.AppState.GetCurrentProject()
}

// reloadProjects reloads the projects list from the database
func reloadProjects(m *tui.Model) {
	ctx, cancel := dbContext(m)
	defer cancel()
	projects, err := m.Repo.GetAllProjects(ctx)
	if err != nil {
		slog.Error("Error reloading projects", "error", err)
		return
	}
	m.AppState.SetProjects(projects)
}

// switchToProject switches to a different project by index and reloads columns/tasks/labels
func switchToProject(m *tui.Model, projectIndex int) {
	if projectIndex < 0 || projectIndex >= len(m.AppState.Projects()) {
		return
	}

	// Update state
	m.AppState.SetSelectedProject(projectIndex)

	project := m.AppState.Projects()[projectIndex]

	// Create context for database operations
	ctx, cancel := dbContext(m)
	defer cancel()

	// Reload columns for this project
	columns, err := m.Repo.GetColumnsByProject(ctx, project.ID)
	if err != nil {
		slog.Error("Error loading columns for project", "project_id", project.ID, "error", err)
		columns = []*models.Column{}
	}
	m.AppState.SetColumns(columns)

	// Reload task summaries for the entire project
	tasks, err := m.Repo.GetTaskSummariesByProject(ctx, project.ID)
	if err != nil {
		slog.Error("Error loading tasks for project", "project_id", project.ID, "error", err)
		tasks = make(map[int][]*models.TaskSummary)
	}
	m.AppState.SetTasks(tasks)

	// Reload labels for this project
	labels, err := m.Repo.GetLabelsByProject(ctx, project.ID)
	if err != nil {
		slog.Error("Error loading labels for project", "project_id", project.ID, "error", err)
		labels = []*models.Label{}
	}
	m.AppState.SetLabels(labels)

	// Reset selection state
	m.UiState.ResetSelection()
}

// initParentPickerForForm initializes the parent picker for use in TicketFormMode
func initParentPickerForForm(m *tui.Model) bool {
	project := getCurrentProject(m)
	if project == nil {
		return false
	}

	// Get all task references for the entire project
	ctx, cancel := dbContext(m)
	defer cancel()
	allTasks, err := m.Repo.GetTaskReferencesForProject(ctx, project.ID)
	if err != nil {
		slog.Error("Error loading project tasks", "error", err)
		return false
	}

	// Build map of currently selected parent task IDs from form state
	parentTaskMap := make(map[int]bool)
	for _, parentID := range m.FormState.FormParentIDs {
		parentTaskMap[parentID] = true
	}

	// Build picker items from all tasks, excluding current task (if editing)
	items := make([]state.TaskPickerItem, 0, len(allTasks))
	for _, task := range allTasks {
		// In edit mode, exclude the task being edited
		if m.FormState.EditingTaskID != 0 && task.ID == m.FormState.EditingTaskID {
			continue
		}
		items = append(items, state.TaskPickerItem{
			TaskRef:  task,
			Selected: parentTaskMap[task.ID],
		})
	}

	// Initialize ParentPickerState
	m.ParentPickerState.Items = items
	m.ParentPickerState.TaskID = m.FormState.EditingTaskID // 0 for create mode
	m.ParentPickerState.Cursor = 0
	m.ParentPickerState.Filter = ""
	m.ParentPickerState.PickerType = "parent"
	m.ParentPickerState.ReturnMode = state.TicketFormMode

	return true
}

// initChildPickerForForm initializes the child picker for use in TicketFormMode
func initChildPickerForForm(m *tui.Model) bool {
	project := getCurrentProject(m)
	if project == nil {
		return false
	}

	// Get all task references for the entire project
	ctx, cancel := dbContext(m)
	defer cancel()
	allTasks, err := m.Repo.GetTaskReferencesForProject(ctx, project.ID)
	if err != nil {
		slog.Error("Error loading project tasks", "error", err)
		return false
	}

	// Build map of currently selected child task IDs from form state
	childTaskMap := make(map[int]bool)
	for _, childID := range m.FormState.FormChildIDs {
		childTaskMap[childID] = true
	}

	// Build picker items from all tasks, excluding current task (if editing)
	items := make([]state.TaskPickerItem, 0, len(allTasks))
	for _, task := range allTasks {
		// In edit mode, exclude the task being edited
		if m.FormState.EditingTaskID != 0 && task.ID == m.FormState.EditingTaskID {
			continue
		}
		items = append(items, state.TaskPickerItem{
			TaskRef:  task,
			Selected: childTaskMap[task.ID],
		})
	}

	// Initialize ChildPickerState
	m.ChildPickerState.Items = items
	m.ChildPickerState.TaskID = m.FormState.EditingTaskID // 0 for create mode
	m.ChildPickerState.Cursor = 0
	m.ChildPickerState.Filter = ""
	m.ChildPickerState.PickerType = "child"
	m.ChildPickerState.ReturnMode = state.TicketFormMode

	return true
}

// initLabelPickerForForm initializes the label picker for use in TicketFormMode
func initLabelPickerForForm(m *tui.Model) bool {
	project := getCurrentProject(m)
	if project == nil {
		return false
	}

	// Build map of currently selected label IDs from form state
	labelIDMap := make(map[int]bool)
	for _, labelID := range m.FormState.FormLabelIDs {
		labelIDMap[labelID] = true
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
	m.LabelPickerState.Items = items
	m.LabelPickerState.TaskID = m.FormState.EditingTaskID // 0 for create mode
	m.LabelPickerState.Cursor = 0
	m.LabelPickerState.Filter = ""
	m.LabelPickerState.ReturnMode = state.TicketFormMode

	return true
}

// initPriorityPickerForForm initializes the priority picker for use in TicketFormMode
func initPriorityPickerForForm(m *tui.Model) bool {
	// Get current priority ID from form state
	// If we're editing a task, get it from the task detail
	// Otherwise, default to medium (id=3)
	currentPriorityID := 3 // Default to medium

	// If editing an existing task, we need to get the current priority from database
	if m.FormState.EditingTaskID != 0 {
		ctx, cancel := dbContext(m)
		defer cancel()

		taskDetail, err := m.Repo.GetTaskDetail(ctx, m.FormState.EditingTaskID)
		if err != nil {
			slog.Error("Error loading task detail for priority picker", "error", err)
			return false
		}

		// Find the priority ID from the priority description
		// We need to match it against our priority options
		priorities := renderers.GetPriorityOptions()
		for _, p := range priorities {
			if p.Description == taskDetail.PriorityDescription {
				currentPriorityID = p.ID
				break
			}
		}
	}

	// Initialize PriorityPickerState
	m.PriorityPickerState.SetSelectedPriorityID(currentPriorityID)
	// Set cursor to match the selected priority (adjust for 0-indexing)
	m.PriorityPickerState.SetCursor(currentPriorityID - 1)
	m.PriorityPickerState.SetReturnMode(state.TicketFormMode)

	return true
}

// initTypePickerForForm initializes the type picker for use in TicketFormMode
func initTypePickerForForm(m *tui.Model) bool {
	// Get current type ID from form state
	// If we're editing a task, get it from the task detail
	// Otherwise, default to task (id=1)
	currentTypeID := 1 // Default to task

	// If editing an existing task, we need to get the current type from database
	if m.FormState.EditingTaskID != 0 {
		ctx, cancel := dbContext(m)
		defer cancel()

		taskDetail, err := m.Repo.GetTaskDetail(ctx, m.FormState.EditingTaskID)
		if err != nil {
			slog.Error("Error loading task detail for type picker", "error", err)
			return false
		}

		// Find the type ID from the type description
		// We need to match it against our type options
		types := renderers.GetTypeOptions()
		for _, t := range types {
			if t.Description == taskDetail.TypeDescription {
				currentTypeID = t.ID
				break
			}
		}
	}

	// Initialize TypePickerState
	m.TypePickerState.SetSelectedTypeID(currentTypeID)
	// Set cursor to match the selected type (adjust for 0-indexing)
	m.TypePickerState.SetCursor(currentTypeID - 1)
	m.TypePickerState.SetReturnMode(state.TicketFormMode)

	return true
}

// extractTicketFormValues extracts and returns form values from the ticket form
// Since our forms update pointers in place, we can just read from formState
func extractTicketFormValues(m *tui.Model) ticketFormValues {
	return ticketFormValues{
		title:       strings.TrimSpace(m.FormState.FormTitle),
		description: strings.TrimSpace(m.FormState.FormDescription),
		confirm:     m.FormState.FormConfirm,
		labelIDs:    m.FormState.FormLabelIDs,
	}
}

// extractProjectFormValues extracts and returns form values from the project form
func extractProjectFormValues(m *tui.Model) projectFormValues {
	return projectFormValues{
		name:        strings.TrimSpace(m.FormState.FormProjectName),
		description: strings.TrimSpace(m.FormState.FormProjectDescription),
		confirm:     m.FormState.FormProjectConfirm,
	}
}

// createNewTaskWithLabelsAndRelationships creates a new task, sets labels, and applies parent/child relationships
func createNewTaskWithLabelsAndRelationships(m *tui.Model, values ticketFormValues) {
	currentCol := getCurrentColumn(m)
	if currentCol == nil {
		m.NotificationState.Add(state.LevelError, "No column selected")
		return
	}

	// Create context for database operations
	ctx, cancel := dbContext(m)
	defer cancel()

	// 1. Create the task
	task, err := m.Repo.CreateTask(ctx,
		values.title,
		values.description,
		currentCol.ID,
		len(getTasksForColumn(m, currentCol.ID)),
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

	// 3. Apply parent relationships with relation types
	// CRITICAL: Parent picker means selected task BLOCKS ON current task
	// So: AddSubtaskWithRelationType(parentID, currentTaskID, relationTypeID)
	for _, item := range m.ParentPickerState.Items {
		if item.Selected {
			relationTypeID := item.RelationTypeID
			if relationTypeID == 0 {
				relationTypeID = 1 // Default to Parent/Child
			}
			err = m.Repo.AddSubtaskWithRelationType(ctx, item.TaskRef.ID, task.ID, relationTypeID)
			if err != nil {
				slog.Error("Error adding parent relationship", "error", err)
			}
		}
	}

	// 4. Apply child relationships with relation types
	// CRITICAL: Child picker means current task BLOCKS ON selected task
	// So: AddSubtaskWithRelationType(currentTaskID, childID, relationTypeID)
	for _, item := range m.ChildPickerState.Items {
		if item.Selected {
			relationTypeID := item.RelationTypeID
			if relationTypeID == 0 {
				relationTypeID = 1 // Default to Parent/Child
			}
			err = m.Repo.AddSubtaskWithRelationType(ctx, task.ID, item.TaskRef.ID, relationTypeID)
			if err != nil {
				slog.Error("Error adding child relationship", "error", err)
			}
		}
	}

	// 5. Reload all tasks for the project to keep state consistent
	project := getCurrentProject(m)
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
func updateExistingTaskWithLabelsAndRelationships(m *tui.Model, values ticketFormValues) {
	// Create context for database operations
	ctx, cancel := dbContext(m)
	defer cancel()
	taskID := m.FormState.EditingTaskID

	// 1. Update task basic fields
	err := m.Repo.UpdateTask(ctx, taskID, values.title, values.description)
	if err != nil {
		slog.Error("Error updating task", "error", err)
		m.NotificationState.Add(state.LevelError, "Error updating task")
		return
	}

	// 2. Update labels
	err = m.Repo.SetTaskLabels(ctx, taskID, values.labelIDs)
	if err != nil {
		slog.Error("Error setting labels", "error", err)
	}

	// 3. Sync parent relationships with relation types
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

	// 4. Sync child relationships with relation types
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

	// 5. Reload all tasks for the project to keep state consistent
	project := getCurrentProject(m)
	if project != nil {
		tasksByColumn, err := m.Repo.GetTaskSummariesByProject(ctx, project.ID)
		if err != nil {
			slog.Error("Error reloading tasks", "error", err)
		} else {
			m.AppState.SetTasks(tasksByColumn)
		}
	}
}

// handleFormUpdate processes form messages generically
func handleFormUpdate(m *tui.Model, msg tea.Msg, cfg formConfig) tea.Cmd {
	if cfg.form == nil {
		m.UiState.SetMode(state.NormalMode)
		return nil
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
		return tea.ClearScreen
	}

	// Note: StateAborted handling removed - ESC is now intercepted in UpdateTicketForm/UpdateProjectForm
	// to allow for change detection and discard confirmation

	return cmd
}

// handleFormSave handles the C-s save shortcut for forms.
// Sets confirmation to true and completes the form, triggering the save flow.
func handleFormSave(m *tui.Model, cfg formConfig) tea.Cmd {
	if cfg.form == nil {
		m.UiState.SetMode(state.NormalMode)
		return nil
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

	return tea.ClearScreen
}

// UpdateTicketForm handles all messages when in TicketFormMode
// This is separated out because forms need to receive ALL messages, not just KeyMsg
func UpdateTicketForm(m *tui.Model, msg tea.Msg) tea.Cmd {
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
				return nil
			}
			// No changes - allow immediate close
			m.UiState.SetMode(state.NormalMode)
			m.FormState.ClearTicketForm()
			return tea.ClearScreen

		case "ctrl+p":
			// Open parent picker
			if initParentPickerForForm(m) {
				m.UiState.SetMode(state.ParentPickerMode)
			}
			return nil

		case "ctrl+c":
			// Open child picker
			if initChildPickerForForm(m) {
				m.UiState.SetMode(state.ChildPickerMode)
			}
			return nil

		case "ctrl+l":
			// Open label picker
			if initLabelPickerForForm(m) {
				m.UiState.SetMode(state.LabelPickerMode)
			}
			return nil

		case "ctrl+r":
			// Open priority picker
			if initPriorityPickerForForm(m) {
				m.UiState.SetMode(state.PriorityPickerMode)
			}
			return nil

		case "ctrl+t":
			// Open type picker
			if initTypePickerForForm(m) {
				m.UiState.SetMode(state.TypePickerMode)
			}
			return nil

		case m.Config.KeyMappings.SaveForm:
			// Quick save via C-s
			return handleFormSave(m, formConfig{
				form: m.FormState.TicketForm,
				setForm: func(f *huh.Form) {
					m.FormState.TicketForm = f
				},
				clearForm: func() {
					m.FormState.ClearTicketForm()
					m.FormState.EditingTaskID = 0
				},
				onComplete: func() {
					values := extractTicketFormValues(m)
					if !values.confirm {
						return
					}
					if values.title != "" {
						if m.FormState.EditingTaskID == 0 {
							createNewTaskWithLabelsAndRelationships(m, values)
						} else {
							updateExistingTaskWithLabelsAndRelationships(m, values)
						}
					}
				},
				confirmPtr: &m.FormState.FormConfirm,
			})
		}
	}

	// Pass through to existing form handler
	return handleFormUpdate(m, msg, formConfig{
		form: m.FormState.TicketForm,
		setForm: func(f *huh.Form) {
			m.FormState.TicketForm = f
		},
		clearForm: func() {
			m.FormState.ClearTicketForm()
			m.FormState.EditingTaskID = 0
		},
		onComplete: func() {
			values := extractTicketFormValues(m)

			// Form submitted - check confirmation and save the task
			if !values.confirm {
				// User selected "No" on confirmation
				return
			}

			if values.title != "" {
				if m.FormState.EditingTaskID == 0 {
					createNewTaskWithLabelsAndRelationships(m, values)
				} else {
					updateExistingTaskWithLabelsAndRelationships(m, values)
				}
			}
		},
	})
}

// UpdateProjectForm handles all messages when in ProjectFormMode
// This is separated out because forms need to receive ALL messages, not just KeyMsg
func UpdateProjectForm(m *tui.Model, msg tea.Msg) tea.Cmd {
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
				return nil
			}
			// No changes - allow immediate close
			m.UiState.SetMode(state.NormalMode)
			m.FormState.ClearProjectForm()
			return tea.ClearScreen

		case m.Config.KeyMappings.SaveForm:
			return handleFormSave(m, formConfig{
				form: m.FormState.ProjectForm,
				setForm: func(f *huh.Form) {
					m.FormState.ProjectForm = f
				},
				clearForm: func() {
					m.FormState.ClearProjectForm()
				},
				onComplete: func() {
					values := extractProjectFormValues(m)

					if !values.confirm {
						return
					}

					if values.name != "" {
						ctx, cancel := dbContext(m)
						defer cancel()
						project, err := m.Repo.CreateProject(ctx, values.name, values.description)
						if err != nil {
							slog.Error("Error creating project", "error", err)
							m.NotificationState.Add(state.LevelError, "Error creating project")
						} else {
							reloadProjects(m)
							for i, p := range m.AppState.Projects() {
								if p.ID == project.ID {
									switchToProject(m, i)
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

	return handleFormUpdate(m, msg, formConfig{
		form: m.FormState.ProjectForm,
		setForm: func(f *huh.Form) {
			m.FormState.ProjectForm = f
		},
		clearForm: func() {
			m.FormState.ClearProjectForm()
		},
		onComplete: func() {
			values := extractProjectFormValues(m)

			// Form submitted - check confirmation and create the project
			if !values.confirm {
				// User selected "No" on confirmation
				return
			}

			if values.name != "" {
				ctx, cancel := dbContext(m)
				defer cancel()
				project, err := m.Repo.CreateProject(ctx, values.name, values.description)
				if err != nil {
					slog.Error("Error creating project", "error", err)
					m.NotificationState.Add(state.LevelError, "Error creating project")
				} else {
					// Reload projects list
					reloadProjects(m)

					// Switch to the new project
					for i, p := range m.AppState.Projects() {
						if p.ID == project.ID {
							switchToProject(m, i)
							break
						}
					}
				}
			}
		},
	})
}
