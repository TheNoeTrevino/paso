package tui

import (
	"log/slog"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"github.com/thenoetrevino/paso/internal/events"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui/renderers"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// RefreshMsg is sent when data changes from other instances via the daemon
type RefreshMsg struct {
	Event events.Event
}

// ConnectionEstablishedMsg is sent when connection to daemon is established
type ConnectionEstablishedMsg struct{}

// ConnectionLostMsg is sent when connection to daemon is lost
type ConnectionLostMsg struct{}

// ConnectionReconnectingMsg is sent when attempting to reconnect to daemon
type ConnectionReconnectingMsg struct{}

// Update handles all messages and updates the model accordingly
// This implements the "Update" part of the Model-View-Update pattern
//
// ARCHITECTURE NOTE (Incremental Migration):
// This file contains the LEGACY update logic for backward compatibility with tests.
// Production code uses: core.App → handlers.Update() → handler functions (see handlers/ package).
// Tests use: Model.Update() → methods below (this file and other internal/tui/*.go handlers).
//
// The duplication is intentional to avoid breaking existing tests during incremental migration.
// Future work: Refactor tests to use core.App, then delete this legacy update logic.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Check if context is cancelled (graceful shutdown)
	select {
	case <-m.Ctx.Done():
		// Context cancelled, initiate graceful shutdown
		return m, tea.Quit
	default:
		// Continue normal processing
	}

	// Start listening for events on first update if not already started
	var cmd tea.Cmd
	if m.EventChan != nil && !m.SubscriptionStarted {
		m.SubscriptionStarted = true
		cmd = m.subscribeToEvents()
	}

	// Handle form modes first - forms need ALL messages
	if m.UiState.Mode() == state.TicketFormMode {
		return m.updateTicketForm(msg)
	}
	if m.UiState.Mode() == state.ProjectFormMode {
		return m.updateProjectForm(msg)
	}

	switch msg := msg.(type) {
	case RefreshMsg:
		// log.Printf("Received refresh event for project %d", msg.Event.ProjectID)

		// Only refresh if event is for current project
		currentProject := m.AppState.GetCurrentProject()
		if currentProject != nil && msg.Event.ProjectID == currentProject.ID {
			m.reloadCurrentProject()
			// m.NotificationState.Add(state.LevelInfo, "Synced with other instances")
		}

		// Continue listening for more events
		cmd = m.subscribeToEvents()
		return m, cmd

	case events.NotificationMsg:
		// Handle user-facing notification from events client
		level := state.LevelInfo
		switch msg.Level {
		case "error":
			level = state.LevelError
		case "warning":
			level = state.LevelWarning
		}
		m.NotificationState.Add(level, msg.Message)

		// Update connection status based on notification message
		if strings.Contains(msg.Message, "Connection lost") || strings.Contains(msg.Message, "reconnecting") {
			m.ConnectionState.SetStatus(state.Reconnecting)
		} else if strings.Contains(msg.Message, "Reconnected") {
			m.ConnectionState.SetStatus(state.Connected)
		} else if strings.Contains(msg.Message, "Failed to reconnect") {
			m.ConnectionState.SetStatus(state.Disconnected)
		}

		// Continue listening for more notifications
		cmd = m.listenForNotifications()
		return m, cmd

	case ConnectionEstablishedMsg:
		m.ConnectionState.SetStatus(state.Connected)
		return m, nil

	case ConnectionLostMsg:
		m.ConnectionState.SetStatus(state.Disconnected)
		return m, nil

	case ConnectionReconnectingMsg:
		m.ConnectionState.SetStatus(state.Reconnecting)
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case tea.WindowSizeMsg:
		return m.handleWindowResize(msg)
	}

	return m, cmd
}

// handleKeyMsg dispatches key messages to the appropriate mode handler.
func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.UiState.Mode() {
	case state.NormalMode:
		return m.handleNormalMode(msg)
	case state.AddColumnMode, state.EditColumnMode:
		return m.handleInputMode(msg)
	case state.DiscardConfirmMode:
		return m.handleDiscardConfirm(msg)
	case state.DeleteConfirmMode:
		return m.handleDeleteConfirm(msg)
	case state.DeleteColumnConfirmMode:
		return m.handleDeleteColumnConfirm(msg)
	case state.HelpMode:
		return m.handleHelpMode(msg)
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
	case state.RelationTypePickerMode:
		return m.updateRelationTypePicker(msg)
	case state.SearchMode:
		return m.handleSearchMode(msg)
	case state.StatusPickerMode:
		return m.handleStatusPickerMode(msg)
	}
	return m, nil
}

// handleWindowResize handles terminal resize events.
func (m Model) handleWindowResize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.UiState.SetWidth(msg.Width)
	m.UiState.SetHeight(msg.Height)

	// Update notification state with new window dimensions
	m.NotificationState.SetWindowSize(msg.Width, msg.Height)

	// Ensure viewport offset is still valid after resize
	if m.UiState.ViewportOffset()+m.UiState.ViewportSize() > len(m.AppState.Columns()) {
		m.UiState.SetViewportOffset(max(0, len(m.AppState.Columns())-m.UiState.ViewportSize()))
	}
	return m, nil
}

// ticketFormValues holds the extracted values from a completed ticket form
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
	// Create context for database operations
	ctx, cancel := m.DbContext()
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

// updateLabelPicker handles keyboard input in the label picker mode
func (m Model) updateLabelPicker(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	// Handle color picker sub-mode for creating new label
	if m.LabelPickerState.CreateMode {
		return m.updateLabelColorPicker(keyMsg)
	}

	// Get filtered items to determine bounds
	filteredItems := m.getFilteredLabelPickerItems()
	maxIdx := len(filteredItems) // +1 for "create new label" option

	switch keyMsg.String() {
	case "esc":
		// Close picker and return to appropriate mode
		if m.LabelPickerState.ReturnMode == state.TicketFormMode {
			// In form mode: sync selections and return to form
			m.syncLabelPickerToFormState()
			m.UiState.SetMode(state.TicketFormMode)
		} else {
			// In view mode: return to NormalMode
			m.UiState.SetMode(state.NormalMode)
		}
		m.LabelPickerState.Filter = ""
		m.LabelPickerState.Cursor = 0
		return m, nil

	case "up", "k":
		// Move cursor up
		if m.LabelPickerState.Cursor > 0 {
			m.LabelPickerState.Cursor--
		}
		return m, nil

	case "down", "j":
		// Move cursor down
		if m.LabelPickerState.Cursor < maxIdx {
			m.LabelPickerState.Cursor++
		}
		return m, nil

	case "enter":
		// Toggle label or create new
		if m.LabelPickerState.Cursor < len(filteredItems) {
			// Toggle this label
			item := filteredItems[m.LabelPickerState.Cursor]

			// Find the index in the unfiltered list
			for i, pi := range m.LabelPickerState.Items {
				if pi.Label.ID == item.Label.ID {
					if m.LabelPickerState.ReturnMode == state.TicketFormMode {
						// In form mode: just toggle selection state, don't update database
						m.LabelPickerState.Items[i].Selected = !m.LabelPickerState.Items[i].Selected
					} else {
						// In view mode: update database immediately
						ctx, cancel := m.UiContext()
						defer cancel()
						if m.LabelPickerState.Items[i].Selected {
							// Remove label from task
							err := m.Repo.RemoveLabelFromTask(ctx, m.LabelPickerState.TaskID, item.Label.ID)
							if err != nil {
								slog.Error("Error removing label", "error", err)
								m.NotificationState.Add(state.LevelError, "Failed to remove label from task")
							} else {
								m.LabelPickerState.Items[i].Selected = false
							}
						} else {
							// Add label to task
							err := m.Repo.AddLabelToTask(ctx, m.LabelPickerState.TaskID, item.Label.ID)
							if err != nil {
								slog.Error("Error adding label", "error", err)
								m.NotificationState.Add(state.LevelError, "Failed to add label to task")
							} else {
								m.LabelPickerState.Items[i].Selected = true
							}
						}
						// Reload task summaries for the current column
						m.reloadCurrentColumnTasks()
					}
					break
				}
			}
		} else {
			// Create new label - switch to color picker sub-mode
			if strings.TrimSpace(m.LabelPickerState.Filter) != "" {
				m.FormState.FormLabelName = strings.TrimSpace(m.LabelPickerState.Filter)
			} else {
				m.FormState.FormLabelName = "New Label"
			}
			m.LabelPickerState.CreateMode = true
			m.LabelPickerState.ColorIdx = 0
		}
		return m, nil

	case "backspace", "ctrl+h":
		// Remove last character from filter
		if len(m.LabelPickerState.Filter) > 0 {
			m.LabelPickerState.Filter = m.LabelPickerState.Filter[:len(m.LabelPickerState.Filter)-1]
			// Reset cursor if it's out of bounds after filter change
			newFiltered := m.getFilteredLabelPickerItems()
			if m.LabelPickerState.Cursor > len(newFiltered) {
				m.LabelPickerState.Cursor = len(newFiltered)
			}
		}
		return m, nil

	default:
		// Type to filter/search
		key := keyMsg.String()
		if len(key) == 1 && len(m.LabelPickerState.Filter) < 50 {
			m.LabelPickerState.Filter += key
			// Reset cursor to 0 when filter changes
			m.LabelPickerState.Cursor = 0
		}
		return m, nil
	}
}

// updateLabelColorPicker handles keyboard input when selecting a color for new label
func (m Model) updateLabelColorPicker(keyMsg tea.KeyMsg) (tea.Model, tea.Cmd) {
	colors := renderers.GetDefaultLabelColors()
	maxIdx := len(colors) - 1

	switch keyMsg.String() {
	case "esc":
		// Cancel and return to label list
		m.LabelPickerState.CreateMode = false
		return m, nil

	case "up", "k":
		if m.LabelPickerState.ColorIdx > 0 {
			m.LabelPickerState.ColorIdx--
		}
		return m, nil

	case "down", "j":
		if m.LabelPickerState.ColorIdx < maxIdx {
			m.LabelPickerState.ColorIdx++
		}
		return m, nil

	case "enter":
		// Create the new label
		color := colors[m.LabelPickerState.ColorIdx].Color
		project := m.getCurrentProject()
		if project == nil {
			m.LabelPickerState.CreateMode = false
			return m, nil
		}

		ctx, cancel := m.DbContext()
		defer cancel()
		label, err := m.Repo.CreateLabel(ctx, project.ID, m.FormState.FormLabelName, color)
		if err != nil {
			slog.Error("Error creating label", "error", err)
			m.NotificationState.Add(state.LevelError, "Failed to create label")
			m.LabelPickerState.CreateMode = false
			return m, nil
		}

		// Add to labels list
		m.AppState.SetLabels(append(m.AppState.Labels(), label))

		// Add to picker items (selected by default)
		m.LabelPickerState.Items = append(m.LabelPickerState.Items, state.LabelPickerItem{
			Label:    label,
			Selected: true,
		})

		// Assign to current task
		err = m.Repo.AddLabelToTask(ctx, m.LabelPickerState.TaskID, label.ID)
		if err != nil {
			slog.Error("Error assigning new label to task", "error", err)
			m.NotificationState.Add(state.LevelError, "Failed to assign label to task")
		}

		// Reload task summaries for the current column
		m.reloadCurrentColumnTasks()

		// Exit create mode and clear filter
		m.LabelPickerState.CreateMode = false
		m.LabelPickerState.Filter = ""
		m.LabelPickerState.Cursor = 0

		return m, nil
	}

	return m, nil
}

// updateParentPicker handles keyboard input in the parent picker mode.
// This function processes navigation (up/down), filtering, and selection toggling.
//
// CRITICAL - Database Parameter Ordering:
// Parent picker uses the selected task as the parent of the current task.
// When toggling relationships:
//   - AddSubtask(selectedTaskID, currentTaskID) - selected becomes parent of current
//   - RemoveSubtask(selectedTaskID, currentTaskID)
//
// This means: selectedTask BLOCKS ON currentTask (selected depends on current).
func (m Model) updateParentPicker(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	// Get filtered items to determine bounds
	filteredItems := m.ParentPickerState.GetFilteredItems()
	maxIdx := len(filteredItems) - 1

	switch keyMsg.String() {
	case "esc":
		// Return to the mode specified by ReturnMode
		returnMode := m.ParentPickerState.ReturnMode
		if returnMode == state.Mode(0) { // Default to NormalMode
			returnMode = state.NormalMode
		}

		// If returning to TicketFormMode, sync selections back to FormState
		if returnMode == state.TicketFormMode {
			m.syncParentPickerToFormState()
		}

		m.UiState.SetMode(returnMode)
		m.ParentPickerState.Filter = ""
		m.ParentPickerState.Cursor = 0
		return m, nil

	case "up", "k":
		// Move cursor up
		m.ParentPickerState.MoveCursorUp()
		return m, nil

	case "down", "j":
		// Move cursor down
		m.ParentPickerState.MoveCursorDown(maxIdx)
		return m, nil

	case "enter":
		// Toggle parent relationship
		if m.ParentPickerState.Cursor < len(filteredItems) {
			item := filteredItems[m.ParentPickerState.Cursor]

			// Find the index in the unfiltered list
			for i, pi := range m.ParentPickerState.Items {
				if pi.TaskRef.ID == item.TaskRef.ID {
					// Determine if we're in form mode or view mode
					if m.ParentPickerState.ReturnMode == state.TicketFormMode {
						// Form mode: just toggle the selection state
						// Actual database changes happen on form submission
						m.ParentPickerState.Items[i].Selected = !m.ParentPickerState.Items[i].Selected
						// Set default relation type when selecting (if not already set)
						if m.ParentPickerState.Items[i].Selected && m.ParentPickerState.Items[i].RelationTypeID == 0 {
							m.ParentPickerState.Items[i].RelationTypeID = 1 // Default to Parent/Child
						}
					} else {
						// View mode: apply changes to database immediately (existing behavior)
						ctx, cancel := m.UiContext()
						defer cancel()
						if m.ParentPickerState.Items[i].Selected {
							// Remove parent relationship
							// CRITICAL: RemoveSubtask(parentID, childID)
							// selectedTask (parent) blocks on currentTask (child)
							err := m.Repo.RemoveSubtask(ctx, item.TaskRef.ID, m.ParentPickerState.TaskID)
							if err != nil {
								slog.Error("Error removing parent", "error", err)
								m.NotificationState.Add(state.LevelError, "Failed to remove parent from task")
							} else {
								m.ParentPickerState.Items[i].Selected = false
							}
						} else {
							// Add parent relationship - selected task becomes parent of current task
							// CRITICAL: AddSubtask(parentID, childID)
							// This makes selectedTask (parent) block on currentTask (child)
							// Meaning: selectedTask depends on completion of currentTask
							err := m.Repo.AddSubtask(ctx, item.TaskRef.ID, m.ParentPickerState.TaskID)
							if err != nil {
								slog.Error("Error adding parent", "error", err)
								m.NotificationState.Add(state.LevelError, "Failed to add parent to task")
							} else {
								m.ParentPickerState.Items[i].Selected = true
							}
						}

						// Reload task summaries
						m.reloadCurrentColumnTasks()
					}
					break
				}
			}
		}
		return m, nil

	case "tab":
		// Open relation type picker for the currently highlighted item
		if m.ParentPickerState.Cursor < len(filteredItems) {
			item := filteredItems[m.ParentPickerState.Cursor]

			// Initialize relation type picker
			currentRelationTypeID := 1 // Default to Parent/Child
			if item.RelationTypeID > 0 {
				currentRelationTypeID = item.RelationTypeID
			}

			m.RelationTypePickerState.SetSelectedRelationTypeID(currentRelationTypeID)
			m.RelationTypePickerState.SetCurrentTaskPickerIndex(m.ParentPickerState.Cursor)
			m.RelationTypePickerState.SetReturnMode(state.ParentPickerMode)

			// Set cursor to match selected relation type
			relationTypes := renderers.GetRelationTypeOptions()
			for i, rt := range relationTypes {
				if rt.ID == currentRelationTypeID {
					m.RelationTypePickerState.SetCursor(i)
					break
				}
			}

			m.UiState.SetMode(state.RelationTypePickerMode)
		}
		return m, nil

	case "backspace", "ctrl+h":
		// Remove last character from filter
		m.ParentPickerState.BackspaceFilter()
		// Reset cursor if it's out of bounds after filter change
		newFiltered := m.ParentPickerState.GetFilteredItems()
		if m.ParentPickerState.Cursor >= len(newFiltered) && len(newFiltered) > 0 {
			m.ParentPickerState.Cursor = len(newFiltered) - 1
		} else if len(newFiltered) == 0 {
			m.ParentPickerState.Cursor = 0
		}
		return m, nil

	default:
		// Type to filter/search
		key := keyMsg.String()
		if len(key) == 1 {
			m.ParentPickerState.AppendFilter(rune(key[0]))
			// Reset cursor to 0 when filter changes
			m.ParentPickerState.Cursor = 0
		}
		return m, nil
	}
}

// updateChildPicker handles keyboard input in the child picker mode.
// This function processes navigation (up/down), filtering, and selection toggling.
//
// CRITICAL - Database Parameter Ordering (REVERSED from parent picker):
// Child picker uses the current task as the parent of the selected task.
// When toggling relationships:
//   - AddSubtask(currentTaskID, selectedTaskID) - current becomes parent of selected
//   - RemoveSubtask(currentTaskID, selectedTaskID)
//
// This means: currentTask BLOCKS ON selectedTask (current depends on selected).
func (m Model) updateChildPicker(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	// Get filtered items to determine bounds
	filteredItems := m.ChildPickerState.GetFilteredItems()
	maxIdx := len(filteredItems) - 1

	switch keyMsg.String() {
	case "esc":
		// Return to the mode specified by ReturnMode
		returnMode := m.ChildPickerState.ReturnMode
		if returnMode == state.Mode(0) { // Default to NormalMode
			returnMode = state.NormalMode
		}

		// If returning to TicketFormMode, sync selections back to FormState
		if returnMode == state.TicketFormMode {
			m.syncChildPickerToFormState()
		}

		m.UiState.SetMode(returnMode)
		m.ChildPickerState.Filter = ""
		m.ChildPickerState.Cursor = 0
		return m, nil

	case "up", "k":
		// Move cursor up
		m.ChildPickerState.MoveCursorUp()
		return m, nil

	case "down", "j":
		// Move cursor down
		m.ChildPickerState.MoveCursorDown(maxIdx)
		return m, nil

	case "enter":
		// Toggle child relationship
		if m.ChildPickerState.Cursor < len(filteredItems) {
			item := filteredItems[m.ChildPickerState.Cursor]

			// Find the index in the unfiltered list
			for i, pi := range m.ChildPickerState.Items {
				if pi.TaskRef.ID == item.TaskRef.ID {
					// Determine if we're in form mode or view mode
					if m.ChildPickerState.ReturnMode == state.TicketFormMode {
						// Form mode: just toggle the selection state
						// Actual database changes happen on form submission
						m.ChildPickerState.Items[i].Selected = !m.ChildPickerState.Items[i].Selected
						// Set default relation type when selecting (if not already set)
						if m.ChildPickerState.Items[i].Selected && m.ChildPickerState.Items[i].RelationTypeID == 0 {
							m.ChildPickerState.Items[i].RelationTypeID = 1 // Default to Parent/Child
						}
					} else {
						// View mode: apply changes to database immediately (existing behavior)
						ctx, cancel := m.UiContext()
						defer cancel()
						if m.ChildPickerState.Items[i].Selected {
							// Remove child relationship
							// CRITICAL: RemoveSubtask(parentID, childID) - REVERSED parameter order from parent picker
							// currentTask (parent) blocks on selectedTask (child)
							err := m.Repo.RemoveSubtask(ctx, m.ChildPickerState.TaskID, item.TaskRef.ID)
							if err != nil {
								slog.Error("Error removing child", "error", err)
								m.NotificationState.Add(state.LevelError, "Failed to remove child from task")
							} else {
								m.ChildPickerState.Items[i].Selected = false
							}
						} else {
							// Add child relationship - current task becomes parent of selected task
							// CRITICAL: AddSubtask(parentID, childID) - REVERSED parameter order from parent picker
							// This makes currentTask (parent) block on selectedTask (child)
							// Meaning: currentTask depends on completion of selectedTask
							err := m.Repo.AddSubtask(ctx, m.ChildPickerState.TaskID, item.TaskRef.ID)
							if err != nil {
								slog.Error("Error adding child", "error", err)
								m.NotificationState.Add(state.LevelError, "Failed to add child to task")
							} else {
								m.ChildPickerState.Items[i].Selected = true
							}
						}

						// Reload task summaries
						m.reloadCurrentColumnTasks()
					}
					break
				}
			}
		}
		return m, nil

	case "tab":
		// Open relation type picker for the currently highlighted item
		if m.ChildPickerState.Cursor < len(filteredItems) {
			item := filteredItems[m.ChildPickerState.Cursor]

			// Initialize relation type picker
			currentRelationTypeID := 1 // Default to Parent/Child
			if item.RelationTypeID > 0 {
				currentRelationTypeID = item.RelationTypeID
			}

			m.RelationTypePickerState.SetSelectedRelationTypeID(currentRelationTypeID)
			m.RelationTypePickerState.SetCurrentTaskPickerIndex(m.ChildPickerState.Cursor)
			m.RelationTypePickerState.SetReturnMode(state.ChildPickerMode)

			// Set cursor to match selected relation type
			relationTypes := renderers.GetRelationTypeOptions()
			for i, rt := range relationTypes {
				if rt.ID == currentRelationTypeID {
					m.RelationTypePickerState.SetCursor(i)
					break
				}
			}

			m.UiState.SetMode(state.RelationTypePickerMode)
		}
		return m, nil

	case "backspace", "ctrl+h":
		// Remove last character from filter
		m.ChildPickerState.BackspaceFilter()
		// Reset cursor if it's out of bounds after filter change
		newFiltered := m.ChildPickerState.GetFilteredItems()
		if m.ChildPickerState.Cursor >= len(newFiltered) && len(newFiltered) > 0 {
			m.ChildPickerState.Cursor = len(newFiltered) - 1
		} else if len(newFiltered) == 0 {
			m.ChildPickerState.Cursor = 0
		}
		return m, nil

	default:
		// Type to filter/search
		key := keyMsg.String()
		if len(key) == 1 {
			m.ChildPickerState.AppendFilter(rune(key[0]))
			// Reset cursor to 0 when filter changes
			m.ChildPickerState.Cursor = 0
		}
		return m, nil
	}
}

// updatePriorityPicker handles keyboard input in the priority picker mode.
// This function processes navigation (up/down) and selection.
func (m Model) updatePriorityPicker(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch keyMsg.String() {
	case "esc":
		// Return to ticket form mode without changing priority
		m.UiState.SetMode(m.PriorityPickerState.ReturnMode())
		m.PriorityPickerState.Reset()
		return m, nil

	case "up", "k":
		// Move cursor up
		m.PriorityPickerState.MoveUp()
		return m, nil

	case "down", "j":
		// Move cursor down
		m.PriorityPickerState.MoveDown()
		return m, nil

	case "enter":
		// Select the priority at cursor position
		priorities := renderers.GetPriorityOptions()
		cursorIdx := m.PriorityPickerState.Cursor()

		if cursorIdx >= 0 && cursorIdx < len(priorities) {
			selectedPriority := priorities[cursorIdx]

			// If we're editing a task, update it in the database
			if m.FormState.EditingTaskID != 0 {
				ctx, cancel := m.DbContext()
				defer cancel()

				// Update the task's priority_id in the database
				err := m.Repo.UpdateTaskPriority(ctx, m.FormState.EditingTaskID, selectedPriority.ID)

				if err != nil {
					slog.Error("Error updating task priority", "error", err)
					m.NotificationState.Add(state.LevelError, "Failed to update priority")
				} else {
					// Update form state with new priority
					m.FormState.FormPriorityDescription = selectedPriority.Description
					m.FormState.FormPriorityColor = selectedPriority.Color
					m.NotificationState.Add(state.LevelInfo, "Priority updated to "+selectedPriority.Description)

					// Reload tasks to reflect the change
					m.reloadCurrentColumnTasks()
				}
			} else {
				// For new tasks, just update the form state
				m.FormState.FormPriorityDescription = selectedPriority.Description
				m.FormState.FormPriorityColor = selectedPriority.Color
				m.NotificationState.Add(state.LevelInfo, "Priority set to "+selectedPriority.Description)
			}

			// Update the selected priority ID in picker state
			m.PriorityPickerState.SetSelectedPriorityID(selectedPriority.ID)
		}

		// Return to ticket form mode
		m.UiState.SetMode(m.PriorityPickerState.ReturnMode())
		return m, nil
	}

	return m, nil
}

// updateTypePicker handles keyboard input in the type picker mode.
// This function processes navigation (up/down) and selection.
func (m Model) updateTypePicker(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch keyMsg.String() {
	case "esc":
		// Return to ticket form mode without changing type
		m.UiState.SetMode(m.TypePickerState.ReturnMode())
		m.TypePickerState.Reset()
		return m, nil

	case "up", "k":
		// Move cursor up
		m.TypePickerState.MoveUp()
		return m, nil

	case "down", "j":
		// Move cursor down
		m.TypePickerState.MoveDown()
		return m, nil

	case "enter":
		// Select the type at cursor position
		types := renderers.GetTypeOptions()
		cursorIdx := m.TypePickerState.Cursor()

		if cursorIdx >= 0 && cursorIdx < len(types) {
			selectedType := types[cursorIdx]

			// If we're editing a task, update it in the database
			if m.FormState.EditingTaskID != 0 {
				ctx, cancel := m.DbContext()
				defer cancel()

				// Update the task's type_id in the database
				err := m.Repo.UpdateTaskType(ctx, m.FormState.EditingTaskID, selectedType.ID)

				if err != nil {
					slog.Error("Error updating task type", "error", err)
					m.NotificationState.Add(state.LevelError, "Failed to update type")
				} else {
					// Update form state with new type
					m.FormState.FormTypeDescription = selectedType.Description
					m.NotificationState.Add(state.LevelInfo, "Type updated to "+selectedType.Description)

					// Reload tasks to reflect the change
					m.reloadCurrentColumnTasks()
				}
			} else {
				// For new tasks, just update the form state
				m.FormState.FormTypeDescription = selectedType.Description
				m.NotificationState.Add(state.LevelInfo, "Type set to "+selectedType.Description)
			}

			// Update the selected type ID in picker state
			m.TypePickerState.SetSelectedTypeID(selectedType.ID)
		}

		// Return to ticket form mode
		m.UiState.SetMode(m.TypePickerState.ReturnMode())
		return m, nil
	}

	return m, nil
}

// updateRelationTypePicker handles keyboard input in the relation type picker mode
func (m Model) updateRelationTypePicker(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch keyMsg.String() {
	case "esc":
		// Return to previous picker (parent or child) without changing relation type
		m.UiState.SetMode(m.RelationTypePickerState.ReturnMode())
		m.RelationTypePickerState.Reset()
		return m, nil

	case "up", "k":
		// Move cursor up
		m.RelationTypePickerState.MoveUp()
		return m, nil

	case "down", "j":
		// Move cursor down
		m.RelationTypePickerState.MoveDown()
		return m, nil

	case "enter":
		// Select the relation type at cursor position
		relationTypes := renderers.GetRelationTypeOptions()
		cursorIdx := m.RelationTypePickerState.Cursor()

		if cursorIdx >= 0 && cursorIdx < len(relationTypes) {
			selectedRelationType := relationTypes[cursorIdx]

			// Update the TaskPickerItem's RelationTypeID
			itemIdx := m.RelationTypePickerState.CurrentTaskPickerIndex()
			returnMode := m.RelationTypePickerState.ReturnMode()

			if returnMode == state.ParentPickerMode {
				// Update parent picker item
				filteredItems := m.ParentPickerState.GetFilteredItems()
				if itemIdx >= 0 && itemIdx < len(filteredItems) {
					// Find the item in the original items list and update it
					taskID := filteredItems[itemIdx].TaskRef.ID
					for i := range m.ParentPickerState.Items {
						if m.ParentPickerState.Items[i].TaskRef.ID == taskID {
							m.ParentPickerState.Items[i].RelationTypeID = selectedRelationType.ID
							break
						}
					}
				}
			} else if returnMode == state.ChildPickerMode {
				// Update child picker item
				filteredItems := m.ChildPickerState.GetFilteredItems()
				if itemIdx >= 0 && itemIdx < len(filteredItems) {
					// Find the item in the original items list and update it
					taskID := filteredItems[itemIdx].TaskRef.ID
					for i := range m.ChildPickerState.Items {
						if m.ChildPickerState.Items[i].TaskRef.ID == taskID {
							m.ChildPickerState.Items[i].RelationTypeID = selectedRelationType.ID
							break
						}
					}
				}
			}

			// Update the selected relation type ID in picker state
			m.RelationTypePickerState.SetSelectedRelationTypeID(selectedRelationType.ID)
		}

		// Return to previous picker mode
		m.UiState.SetMode(m.RelationTypePickerState.ReturnMode())
		return m, nil
	}

	return m, nil
}

// syncParentPickerToFormState syncs parent picker selections back to form state.
// Extracts all selected task IDs and references from the picker and updates FormState.
func (m *Model) syncParentPickerToFormState() {
	var parentIDs []int
	var parentRefs []*models.TaskReference

	for _, item := range m.ParentPickerState.Items {
		if item.Selected {
			parentIDs = append(parentIDs, item.TaskRef.ID)
			parentRefs = append(parentRefs, item.TaskRef)
		}
	}

	m.FormState.FormParentIDs = parentIDs
	m.FormState.FormParentRefs = parentRefs
}

// syncChildPickerToFormState syncs child picker selections back to form state.
// Extracts all selected task IDs and references from the picker and updates FormState.
func (m *Model) syncChildPickerToFormState() {
	var childIDs []int
	var childRefs []*models.TaskReference

	for _, item := range m.ChildPickerState.Items {
		if item.Selected {
			childIDs = append(childIDs, item.TaskRef.ID)
			childRefs = append(childRefs, item.TaskRef)
		}
	}

	m.FormState.FormChildIDs = childIDs
	m.FormState.FormChildRefs = childRefs
}

// syncLabelPickerToFormState syncs label picker selections back to form state.
// Extracts all selected label IDs from the picker and updates FormState.
func (m *Model) syncLabelPickerToFormState() {
	var labelIDs []int

	for _, item := range m.LabelPickerState.Items {
		if item.Selected {
			labelIDs = append(labelIDs, item.Label.ID)
		}
	}

	m.FormState.FormLabelIDs = labelIDs
}

// reloadCurrentColumnTasks reloads all task summaries for the project to keep state consistent
func (m *Model) reloadCurrentColumnTasks() {
	project := m.getCurrentProject()
	if project == nil {
		return
	}

	ctx, cancel := m.DbContext()
	defer cancel()
	tasksByColumn, err := m.Repo.GetTaskSummariesByProject(ctx, project.ID)
	if err != nil {
		slog.Error("Error reloading tasks", "error", err)
		return
	}
	m.AppState.SetTasks(tasksByColumn)
}
