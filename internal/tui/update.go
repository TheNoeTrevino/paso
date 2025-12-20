package tui

import (
	"log/slog"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"github.com/thenoetrevino/paso/internal/events"
	"github.com/thenoetrevino/paso/internal/models"
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
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Check if context is cancelled (graceful shutdown)
	select {
	case <-m.ctx.Done():
		// Context cancelled, initiate graceful shutdown
		return m, tea.Quit
	default:
		// Continue normal processing
	}

	// Start listening for events on first update if not already started
	var cmd tea.Cmd
	if m.eventChan != nil && !m.subscriptionStarted {
		m.subscriptionStarted = true
		cmd = m.subscribeToEvents()
	}

	// Handle form modes first - forms need ALL messages
	if m.uiState.Mode() == state.TicketFormMode {
		return m.updateTicketForm(msg)
	}
	if m.uiState.Mode() == state.ProjectFormMode {
		return m.updateProjectForm(msg)
	}

	switch msg := msg.(type) {
	case RefreshMsg:
		// log.Printf("Received refresh event for project %d", msg.Event.ProjectID)

		// Only refresh if event is for current project
		currentProject := m.appState.GetCurrentProject()
		if currentProject != nil && msg.Event.ProjectID == currentProject.ID {
			m.reloadCurrentProject()
			// m.notificationState.Add(state.LevelInfo, "Synced with other instances")
		}

		// Continue listening for more events
		cmd = m.subscribeToEvents()
		return m, cmd

	case events.NotificationMsg:
		// Handle user-facing notification from events client
		level := state.LevelInfo
		if msg.Level == "error" {
			level = state.LevelError
		} else if msg.Level == "warning" {
			level = state.LevelWarning
		}
		m.notificationState.Add(level, msg.Message)

		// Update connection status based on notification message
		if strings.Contains(msg.Message, "Connection lost") || strings.Contains(msg.Message, "reconnecting") {
			m.connectionState.SetStatus(state.Reconnecting)
		} else if strings.Contains(msg.Message, "Reconnected") {
			m.connectionState.SetStatus(state.Connected)
		} else if strings.Contains(msg.Message, "Failed to reconnect") {
			m.connectionState.SetStatus(state.Disconnected)
		}

		// Continue listening for more notifications
		cmd = m.listenForNotifications()
		return m, cmd

	case ConnectionEstablishedMsg:
		m.connectionState.SetStatus(state.Connected)
		return m, nil

	case ConnectionLostMsg:
		m.connectionState.SetStatus(state.Disconnected)
		return m, nil

	case ConnectionReconnectingMsg:
		m.connectionState.SetStatus(state.Reconnecting)
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
	switch m.uiState.Mode() {
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
	case state.SearchMode:
		return m.handleSearchMode(msg)
	case state.StatusPickerMode:
		return m.handleStatusPickerMode(msg)
	}
	return m, nil
}

// handleWindowResize handles terminal resize events.
func (m Model) handleWindowResize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.uiState.SetWidth(msg.Width)
	m.uiState.SetHeight(msg.Height)

	// Update notification state with new window dimensions
	m.notificationState.SetWindowSize(msg.Width, msg.Height)

	// Ensure viewport offset is still valid after resize
	if m.uiState.ViewportOffset()+m.uiState.ViewportSize() > len(m.appState.Columns()) {
		m.uiState.SetViewportOffset(max(0, len(m.appState.Columns())-m.uiState.ViewportSize()))
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
		title:       strings.TrimSpace(m.formState.FormTitle),
		description: strings.TrimSpace(m.formState.FormDescription),
		confirm:     m.formState.FormConfirm,
		labelIDs:    m.formState.FormLabelIDs,
	}
}

// createNewTaskWithLabelsAndRelationships creates a new task, sets labels, and applies parent/child relationships
func (m *Model) createNewTaskWithLabelsAndRelationships(values ticketFormValues) {
	currentCol := m.getCurrentColumn()
	if currentCol == nil {
		m.notificationState.Add(state.LevelError, "No column selected")
		return
	}

	// Create context for database operations
	ctx, cancel := m.dbContext()
	defer cancel()

	// 1. Create the task
	task, err := m.repo.CreateTask(ctx,
		values.title,
		values.description,
		currentCol.ID,
		len(m.getTasksForColumn(currentCol.ID)),
	)
	if err != nil {
		slog.Error("Error creating task", "error", err)
		m.notificationState.Add(state.LevelError, "Error creating task")
		return
	}

	// 2. Set labels
	if len(values.labelIDs) > 0 {
		err = m.repo.SetTaskLabels(ctx, task.ID, values.labelIDs)
		if err != nil {
			slog.Error("Error setting labels", "error", err)
		}
	}

	// 3. Apply parent relationships
	// CRITICAL: Parent picker means selected task BLOCKS ON current task
	// So: AddSubtask(parentID, currentTaskID)
	for _, parentID := range m.formState.FormParentIDs {
		err = m.repo.AddSubtask(ctx, parentID, task.ID)
		if err != nil {
			slog.Error("Error adding parent relationship", "error", err)
		}
	}

	// 4. Apply child relationships
	// CRITICAL: Child picker means current task BLOCKS ON selected task
	// So: AddSubtask(currentTaskID, childID)
	for _, childID := range m.formState.FormChildIDs {
		err = m.repo.AddSubtask(ctx, task.ID, childID)
		if err != nil {
			slog.Error("Error adding child relationship", "error", err)
		}
	}

	// 5. Reload all tasks for the project to keep state consistent
	project := m.getCurrentProject()
	if project != nil {
		tasksByColumn, err := m.repo.GetTaskSummariesByProject(ctx, project.ID)
		if err != nil {
			slog.Error("Error reloading tasks", "error", err)
		} else {
			m.appState.SetTasks(tasksByColumn)
		}
	}
}

// updateExistingTaskWithLabelsAndRelationships updates task, labels, and parent/child relationships
func (m *Model) updateExistingTaskWithLabelsAndRelationships(values ticketFormValues) {
	// Create context for database operations
	ctx, cancel := m.dbContext()
	defer cancel()
	taskID := m.formState.EditingTaskID

	// 1. Update task basic fields
	err := m.repo.UpdateTask(ctx, taskID, values.title, values.description)
	if err != nil {
		slog.Error("Error updating task", "error", err)
		m.notificationState.Add(state.LevelError, "Error updating task")
		return
	}

	// 2. Update labels
	err = m.repo.SetTaskLabels(ctx, taskID, values.labelIDs)
	if err != nil {
		slog.Error("Error setting labels", "error", err)
	}

	// 3. Sync parent relationships
	// Get current parents from database
	currentParents, err := m.repo.GetParentTasks(ctx, taskID)
	if err != nil {
		slog.Error("Error getting current parents", "error", err)
		currentParents = []*models.TaskReference{}
	}

	// Build sets for comparison
	currentParentIDs := make(map[int]bool)
	for _, p := range currentParents {
		currentParentIDs[p.ID] = true
	}

	newParentIDs := make(map[int]bool)
	for _, id := range m.formState.FormParentIDs {
		newParentIDs[id] = true
	}

	// Remove parents that are no longer selected
	for parentID := range currentParentIDs {
		if !newParentIDs[parentID] {
			err = m.repo.RemoveSubtask(ctx, parentID, taskID)
			if err != nil {
				slog.Error("Error removing parent %d", "error", parentID, err)
			}
		}
	}

	// Add new parents
	for parentID := range newParentIDs {
		if !currentParentIDs[parentID] {
			err = m.repo.AddSubtask(ctx, parentID, taskID)
			if err != nil {
				slog.Error("Error adding parent %d", "error", parentID, err)
			}
		}
	}

	// 4. Sync child relationships (same pattern)
	currentChildren, err := m.repo.GetChildTasks(ctx, taskID)
	if err != nil {
		slog.Error("Error getting current children", "error", err)
		currentChildren = []*models.TaskReference{}
	}

	currentChildIDs := make(map[int]bool)
	for _, c := range currentChildren {
		currentChildIDs[c.ID] = true
	}

	newChildIDs := make(map[int]bool)
	for _, id := range m.formState.FormChildIDs {
		newChildIDs[id] = true
	}

	// Remove children that are no longer selected
	for childID := range currentChildIDs {
		if !newChildIDs[childID] {
			err = m.repo.RemoveSubtask(ctx, taskID, childID)
			if err != nil {
				slog.Error("Error removing child %d", "error", childID, err)
			}
		}
	}

	// Add new children
	for childID := range newChildIDs {
		if !currentChildIDs[childID] {
			err = m.repo.AddSubtask(ctx, taskID, childID)
			if err != nil {
				slog.Error("Error adding child %d", "error", childID, err)
			}
		}
	}

	// 5. Reload all tasks for the project to keep state consistent
	project := m.getCurrentProject()
	if project != nil {
		tasksByColumn, err := m.repo.GetTaskSummariesByProject(ctx, project.ID)
		if err != nil {
			slog.Error("Error reloading tasks", "error", err)
		} else {
			m.appState.SetTasks(tasksByColumn)
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
		m.uiState.SetMode(state.NormalMode)
		return m, nil
	}

	// Forward to form
	model, cmd := cfg.form.Update(msg)
	cfg.setForm(model.(*huh.Form))

	// Check completion
	if cfg.form.State == huh.StateCompleted {
		cfg.onComplete()
		m.uiState.SetMode(state.NormalMode)
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
		m.uiState.SetMode(state.NormalMode)
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
	m.uiState.SetMode(state.NormalMode)
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
			if m.formState.HasTicketFormChanges() {
				// Show discard confirmation
				m.uiState.SetDiscardContext(&state.DiscardContext{
					SourceMode: state.TicketFormMode,
					Message:    "Discard task?",
				})
				m.uiState.SetMode(state.DiscardConfirmMode)
				return m, nil
			}
			// No changes - allow immediate close
			m.uiState.SetMode(state.NormalMode)
			m.formState.ClearTicketForm()
			return m, tea.ClearScreen

		case "ctrl+p":
			// Open parent picker
			if m.initParentPickerForForm() {
				m.uiState.SetMode(state.ParentPickerMode)
			}
			return m, nil

		case "ctrl+c":
			// Open child picker
			if m.initChildPickerForForm() {
				m.uiState.SetMode(state.ChildPickerMode)
			}
			return m, nil

		case "ctrl+l":
			// Open label picker
			if m.initLabelPickerForForm() {
				m.uiState.SetMode(state.LabelPickerMode)
			}
			return m, nil

		case "ctrl+r":
			// Open priority picker
			if m.initPriorityPickerForForm() {
				m.uiState.SetMode(state.PriorityPickerMode)
			}
			return m, nil

		case "ctrl+t":
			// Open type picker
			if m.initTypePickerForForm() {
				m.uiState.SetMode(state.TypePickerMode)
			}
			return m, nil

		case m.config.KeyMappings.SaveForm:
			// Quick save via C-s
			return m.handleFormSave(formConfig{
				form: m.formState.TicketForm,
				setForm: func(f *huh.Form) {
					m.formState.TicketForm = f
				},
				clearForm: func() {
					m.formState.ClearTicketForm()
					m.formState.EditingTaskID = 0
				},
				onComplete: func() {
					values := m.extractTicketFormValues()
					if !values.confirm {
						return
					}
					if values.title != "" {
						if m.formState.EditingTaskID == 0 {
							m.createNewTaskWithLabelsAndRelationships(values)
						} else {
							m.updateExistingTaskWithLabelsAndRelationships(values)
						}
					}
				},
				confirmPtr: &m.formState.FormConfirm,
			})
		}
	}

	// Pass through to existing form handler
	return m.handleFormUpdate(msg, formConfig{
		form: m.formState.TicketForm,
		setForm: func(f *huh.Form) {
			m.formState.TicketForm = f
		},
		clearForm: func() {
			m.formState.ClearTicketForm()
			m.formState.EditingTaskID = 0
		},
		onComplete: func() {
			values := m.extractTicketFormValues()

			// Form submitted - check confirmation and save the task
			if !values.confirm {
				// User selected "No" on confirmation
				return
			}

			if values.title != "" {
				if m.formState.EditingTaskID == 0 {
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
			if m.formState.HasProjectFormChanges() {
				// Show discard confirmation
				m.uiState.SetDiscardContext(&state.DiscardContext{
					SourceMode: state.ProjectFormMode,
					Message:    "Discard project?",
				})
				m.uiState.SetMode(state.DiscardConfirmMode)
				return m, nil
			}
			// No changes - allow immediate close
			m.uiState.SetMode(state.NormalMode)
			m.formState.ClearProjectForm()
			return m, tea.ClearScreen

		case m.config.KeyMappings.SaveForm:
			return m.handleFormSave(formConfig{
				form: m.formState.ProjectForm,
				setForm: func(f *huh.Form) {
					m.formState.ProjectForm = f
				},
				clearForm: func() {
					m.formState.ClearProjectForm()
				},
				onComplete: func() {
					name := strings.TrimSpace(m.formState.FormProjectName)
					description := strings.TrimSpace(m.formState.FormProjectDescription)
					confirm := m.formState.FormProjectConfirm

					if !confirm {
						return
					}

					if name != "" {
						ctx, cancel := m.dbContext()
						defer cancel()
						project, err := m.repo.CreateProject(ctx, name, description)
						if err != nil {
							slog.Error("Error creating project", "error", err)
							m.notificationState.Add(state.LevelError, "Error creating project")
						} else {
							m.reloadProjects()
							for i, p := range m.appState.Projects() {
								if p.ID == project.ID {
									m.switchToProject(i)
									break
								}
							}
						}
					}
				},
				confirmPtr: &m.formState.FormProjectConfirm,
			})
		}
	}

	return m.handleFormUpdate(msg, formConfig{
		form: m.formState.ProjectForm,
		setForm: func(f *huh.Form) {
			m.formState.ProjectForm = f
		},
		clearForm: func() {
			m.formState.ClearProjectForm()
		},
		onComplete: func() {
			// Read values from form state (forms update pointers in place)
			name := strings.TrimSpace(m.formState.FormProjectName)
			description := strings.TrimSpace(m.formState.FormProjectDescription)
			confirm := m.formState.FormProjectConfirm

			// Form submitted - check confirmation and create the project
			if !confirm {
				// User selected "No" on confirmation
				return
			}

			if name != "" {
				ctx, cancel := m.dbContext()
				defer cancel()
				project, err := m.repo.CreateProject(ctx, name, description)
				if err != nil {
					slog.Error("Error creating project", "error", err)
					m.notificationState.Add(state.LevelError, "Error creating project")
				} else {
					// Reload projects list
					m.reloadProjects()

					// Switch to the new project
					for i, p := range m.appState.Projects() {
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
	if m.labelPickerState.CreateMode {
		return m.updateLabelColorPicker(keyMsg)
	}

	// Get filtered items to determine bounds
	filteredItems := m.getFilteredLabelPickerItems()
	maxIdx := len(filteredItems) // +1 for "create new label" option

	switch keyMsg.String() {
	case "esc":
		// Close picker and return to appropriate mode
		if m.labelPickerState.ReturnMode == state.TicketFormMode {
			// In form mode: sync selections and return to form
			m.syncLabelPickerToFormState()
			m.uiState.SetMode(state.TicketFormMode)
		} else {
			// In view mode: return to NormalMode
			m.uiState.SetMode(state.NormalMode)
		}
		m.labelPickerState.Filter = ""
		m.labelPickerState.Cursor = 0
		return m, nil

	case "up", "k":
		// Move cursor up
		if m.labelPickerState.Cursor > 0 {
			m.labelPickerState.Cursor--
		}
		return m, nil

	case "down", "j":
		// Move cursor down
		if m.labelPickerState.Cursor < maxIdx {
			m.labelPickerState.Cursor++
		}
		return m, nil

	case "enter":
		// Toggle label or create new
		if m.labelPickerState.Cursor < len(filteredItems) {
			// Toggle this label
			item := filteredItems[m.labelPickerState.Cursor]

			// Find the index in the unfiltered list
			for i, pi := range m.labelPickerState.Items {
				if pi.Label.ID == item.Label.ID {
					if m.labelPickerState.ReturnMode == state.TicketFormMode {
						// In form mode: just toggle selection state, don't update database
						m.labelPickerState.Items[i].Selected = !m.labelPickerState.Items[i].Selected
					} else {
						// In view mode: update database immediately
						ctx, cancel := m.uiContext()
						defer cancel()
						if m.labelPickerState.Items[i].Selected {
							// Remove label from task
							err := m.repo.RemoveLabelFromTask(ctx, m.labelPickerState.TaskID, item.Label.ID)
							if err != nil {
								slog.Error("Error removing label", "error", err)
								m.notificationState.Add(state.LevelError, "Failed to remove label from task")
							} else {
								m.labelPickerState.Items[i].Selected = false
							}
						} else {
							// Add label to task
							err := m.repo.AddLabelToTask(ctx, m.labelPickerState.TaskID, item.Label.ID)
							if err != nil {
								slog.Error("Error adding label", "error", err)
								m.notificationState.Add(state.LevelError, "Failed to add label to task")
							} else {
								m.labelPickerState.Items[i].Selected = true
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
			if strings.TrimSpace(m.labelPickerState.Filter) != "" {
				m.formState.FormLabelName = strings.TrimSpace(m.labelPickerState.Filter)
			} else {
				m.formState.FormLabelName = "New Label"
			}
			m.labelPickerState.CreateMode = true
			m.labelPickerState.ColorIdx = 0
		}
		return m, nil

	case "backspace", "ctrl+h":
		// Remove last character from filter
		if len(m.labelPickerState.Filter) > 0 {
			m.labelPickerState.Filter = m.labelPickerState.Filter[:len(m.labelPickerState.Filter)-1]
			// Reset cursor if it's out of bounds after filter change
			newFiltered := m.getFilteredLabelPickerItems()
			if m.labelPickerState.Cursor > len(newFiltered) {
				m.labelPickerState.Cursor = len(newFiltered)
			}
		}
		return m, nil

	default:
		// Type to filter/search
		key := keyMsg.String()
		if len(key) == 1 && len(m.labelPickerState.Filter) < 50 {
			m.labelPickerState.Filter += key
			// Reset cursor to 0 when filter changes
			m.labelPickerState.Cursor = 0
		}
		return m, nil
	}
}

// updateLabelColorPicker handles keyboard input when selecting a color for new label
func (m Model) updateLabelColorPicker(keyMsg tea.KeyMsg) (tea.Model, tea.Cmd) {
	colors := GetDefaultLabelColors()
	maxIdx := len(colors) - 1

	switch keyMsg.String() {
	case "esc":
		// Cancel and return to label list
		m.labelPickerState.CreateMode = false
		return m, nil

	case "up", "k":
		if m.labelPickerState.ColorIdx > 0 {
			m.labelPickerState.ColorIdx--
		}
		return m, nil

	case "down", "j":
		if m.labelPickerState.ColorIdx < maxIdx {
			m.labelPickerState.ColorIdx++
		}
		return m, nil

	case "enter":
		// Create the new label
		color := colors[m.labelPickerState.ColorIdx].Color
		project := m.getCurrentProject()
		if project == nil {
			m.labelPickerState.CreateMode = false
			return m, nil
		}

		ctx, cancel := m.dbContext()
		defer cancel()
		label, err := m.repo.CreateLabel(ctx, project.ID, m.formState.FormLabelName, color)
		if err != nil {
			slog.Error("Error creating label", "error", err)
			m.notificationState.Add(state.LevelError, "Failed to create label")
			m.labelPickerState.CreateMode = false
			return m, nil
		}

		// Add to labels list
		m.appState.SetLabels(append(m.appState.Labels(), label))

		// Add to picker items (selected by default)
		m.labelPickerState.Items = append(m.labelPickerState.Items, state.LabelPickerItem{
			Label:    label,
			Selected: true,
		})

		// Assign to current task
		err = m.repo.AddLabelToTask(ctx, m.labelPickerState.TaskID, label.ID)
		if err != nil {
			slog.Error("Error assigning new label to task", "error", err)
			m.notificationState.Add(state.LevelError, "Failed to assign label to task")
		}

		// Reload task summaries for the current column
		m.reloadCurrentColumnTasks()

		// Exit create mode and clear filter
		m.labelPickerState.CreateMode = false
		m.labelPickerState.Filter = ""
		m.labelPickerState.Cursor = 0

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
	filteredItems := m.parentPickerState.GetFilteredItems()
	maxIdx := len(filteredItems) - 1

	switch keyMsg.String() {
	case "esc":
		// Return to the mode specified by ReturnMode
		returnMode := m.parentPickerState.ReturnMode
		if returnMode == state.Mode(0) { // Default to NormalMode
			returnMode = state.NormalMode
		}

		// If returning to TicketFormMode, sync selections back to FormState
		if returnMode == state.TicketFormMode {
			m.syncParentPickerToFormState()
		}

		m.uiState.SetMode(returnMode)
		m.parentPickerState.Filter = ""
		m.parentPickerState.Cursor = 0
		return m, nil

	case "up", "k":
		// Move cursor up
		m.parentPickerState.MoveCursorUp()
		return m, nil

	case "down", "j":
		// Move cursor down
		m.parentPickerState.MoveCursorDown(maxIdx)
		return m, nil

	case "enter":
		// Toggle parent relationship
		if m.parentPickerState.Cursor < len(filteredItems) {
			item := filteredItems[m.parentPickerState.Cursor]

			// Find the index in the unfiltered list
			for i, pi := range m.parentPickerState.Items {
				if pi.TaskRef.ID == item.TaskRef.ID {
					// Determine if we're in form mode or view mode
					if m.parentPickerState.ReturnMode == state.TicketFormMode {
						// Form mode: just toggle the selection state
						// Actual database changes happen on form submission
						m.parentPickerState.Items[i].Selected = !m.parentPickerState.Items[i].Selected
					} else {
						// View mode: apply changes to database immediately (existing behavior)
						ctx, cancel := m.uiContext()
						defer cancel()
						if m.parentPickerState.Items[i].Selected {
							// Remove parent relationship
							// CRITICAL: RemoveSubtask(parentID, childID)
							// selectedTask (parent) blocks on currentTask (child)
							err := m.repo.RemoveSubtask(ctx, item.TaskRef.ID, m.parentPickerState.TaskID)
							if err != nil {
								slog.Error("Error removing parent", "error", err)
								m.notificationState.Add(state.LevelError, "Failed to remove parent from task")
							} else {
								m.parentPickerState.Items[i].Selected = false
							}
						} else {
							// Add parent relationship - selected task becomes parent of current task
							// CRITICAL: AddSubtask(parentID, childID)
							// This makes selectedTask (parent) block on currentTask (child)
							// Meaning: selectedTask depends on completion of currentTask
							err := m.repo.AddSubtask(ctx, item.TaskRef.ID, m.parentPickerState.TaskID)
							if err != nil {
								slog.Error("Error adding parent", "error", err)
								m.notificationState.Add(state.LevelError, "Failed to add parent to task")
							} else {
								m.parentPickerState.Items[i].Selected = true
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

	case "backspace", "ctrl+h":
		// Remove last character from filter
		m.parentPickerState.BackspaceFilter()
		// Reset cursor if it's out of bounds after filter change
		newFiltered := m.parentPickerState.GetFilteredItems()
		if m.parentPickerState.Cursor >= len(newFiltered) && len(newFiltered) > 0 {
			m.parentPickerState.Cursor = len(newFiltered) - 1
		} else if len(newFiltered) == 0 {
			m.parentPickerState.Cursor = 0
		}
		return m, nil

	default:
		// Type to filter/search
		key := keyMsg.String()
		if len(key) == 1 {
			m.parentPickerState.AppendFilter(rune(key[0]))
			// Reset cursor to 0 when filter changes
			m.parentPickerState.Cursor = 0
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
	filteredItems := m.childPickerState.GetFilteredItems()
	maxIdx := len(filteredItems) - 1

	switch keyMsg.String() {
	case "esc":
		// Return to the mode specified by ReturnMode
		returnMode := m.childPickerState.ReturnMode
		if returnMode == state.Mode(0) { // Default to NormalMode
			returnMode = state.NormalMode
		}

		// If returning to TicketFormMode, sync selections back to FormState
		if returnMode == state.TicketFormMode {
			m.syncChildPickerToFormState()
		}

		m.uiState.SetMode(returnMode)
		m.childPickerState.Filter = ""
		m.childPickerState.Cursor = 0
		return m, nil

	case "up", "k":
		// Move cursor up
		m.childPickerState.MoveCursorUp()
		return m, nil

	case "down", "j":
		// Move cursor down
		m.childPickerState.MoveCursorDown(maxIdx)
		return m, nil

	case "enter":
		// Toggle child relationship
		if m.childPickerState.Cursor < len(filteredItems) {
			item := filteredItems[m.childPickerState.Cursor]

			// Find the index in the unfiltered list
			for i, pi := range m.childPickerState.Items {
				if pi.TaskRef.ID == item.TaskRef.ID {
					// Determine if we're in form mode or view mode
					if m.childPickerState.ReturnMode == state.TicketFormMode {
						// Form mode: just toggle the selection state
						// Actual database changes happen on form submission
						m.childPickerState.Items[i].Selected = !m.childPickerState.Items[i].Selected
					} else {
						// View mode: apply changes to database immediately (existing behavior)
						ctx, cancel := m.uiContext()
						defer cancel()
						if m.childPickerState.Items[i].Selected {
							// Remove child relationship
							// CRITICAL: RemoveSubtask(parentID, childID) - REVERSED parameter order from parent picker
							// currentTask (parent) blocks on selectedTask (child)
							err := m.repo.RemoveSubtask(ctx, m.childPickerState.TaskID, item.TaskRef.ID)
							if err != nil {
								slog.Error("Error removing child", "error", err)
								m.notificationState.Add(state.LevelError, "Failed to remove child from task")
							} else {
								m.childPickerState.Items[i].Selected = false
							}
						} else {
							// Add child relationship - current task becomes parent of selected task
							// CRITICAL: AddSubtask(parentID, childID) - REVERSED parameter order from parent picker
							// This makes currentTask (parent) block on selectedTask (child)
							// Meaning: currentTask depends on completion of selectedTask
							err := m.repo.AddSubtask(ctx, m.childPickerState.TaskID, item.TaskRef.ID)
							if err != nil {
								slog.Error("Error adding child", "error", err)
								m.notificationState.Add(state.LevelError, "Failed to add child to task")
							} else {
								m.childPickerState.Items[i].Selected = true
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

	case "backspace", "ctrl+h":
		// Remove last character from filter
		m.childPickerState.BackspaceFilter()
		// Reset cursor if it's out of bounds after filter change
		newFiltered := m.childPickerState.GetFilteredItems()
		if m.childPickerState.Cursor >= len(newFiltered) && len(newFiltered) > 0 {
			m.childPickerState.Cursor = len(newFiltered) - 1
		} else if len(newFiltered) == 0 {
			m.childPickerState.Cursor = 0
		}
		return m, nil

	default:
		// Type to filter/search
		key := keyMsg.String()
		if len(key) == 1 {
			m.childPickerState.AppendFilter(rune(key[0]))
			// Reset cursor to 0 when filter changes
			m.childPickerState.Cursor = 0
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
		m.uiState.SetMode(m.priorityPickerState.ReturnMode())
		m.priorityPickerState.Reset()
		return m, nil

	case "up", "k":
		// Move cursor up
		m.priorityPickerState.MoveUp()
		return m, nil

	case "down", "j":
		// Move cursor down
		m.priorityPickerState.MoveDown()
		return m, nil

	case "enter":
		// Select the priority at cursor position
		priorities := GetPriorityOptions()
		cursorIdx := m.priorityPickerState.Cursor()

		if cursorIdx >= 0 && cursorIdx < len(priorities) {
			selectedPriority := priorities[cursorIdx]

			// If we're editing a task, update it in the database
			if m.formState.EditingTaskID != 0 {
				ctx, cancel := m.dbContext()
				defer cancel()

				// Update the task's priority_id in the database
				err := m.repo.UpdateTaskPriority(ctx, m.formState.EditingTaskID, selectedPriority.ID)

				if err != nil {
					slog.Error("Error updating task priority", "error", err)
					m.notificationState.Add(state.LevelError, "Failed to update priority")
				} else {
					// Update form state with new priority
					m.formState.FormPriorityDescription = selectedPriority.Description
					m.formState.FormPriorityColor = selectedPriority.Color
					m.notificationState.Add(state.LevelInfo, "Priority updated to "+selectedPriority.Description)

					// Reload tasks to reflect the change
					m.reloadCurrentColumnTasks()
				}
			} else {
				// For new tasks, just update the form state
				m.formState.FormPriorityDescription = selectedPriority.Description
				m.formState.FormPriorityColor = selectedPriority.Color
				m.notificationState.Add(state.LevelInfo, "Priority set to "+selectedPriority.Description)
			}

			// Update the selected priority ID in picker state
			m.priorityPickerState.SetSelectedPriorityID(selectedPriority.ID)
		}

		// Return to ticket form mode
		m.uiState.SetMode(m.priorityPickerState.ReturnMode())
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
		m.uiState.SetMode(m.typePickerState.ReturnMode())
		m.typePickerState.Reset()
		return m, nil

	case "up", "k":
		// Move cursor up
		m.typePickerState.MoveUp()
		return m, nil

	case "down", "j":
		// Move cursor down
		m.typePickerState.MoveDown()
		return m, nil

	case "enter":
		// Select the type at cursor position
		types := GetTypeOptions()
		cursorIdx := m.typePickerState.Cursor()

		if cursorIdx >= 0 && cursorIdx < len(types) {
			selectedType := types[cursorIdx]

			// If we're editing a task, update it in the database
			if m.formState.EditingTaskID != 0 {
				ctx, cancel := m.dbContext()
				defer cancel()

				// Update the task's type_id in the database
				err := m.repo.UpdateTaskType(ctx, m.formState.EditingTaskID, selectedType.ID)

				if err != nil {
					slog.Error("Error updating task type", "error", err)
					m.notificationState.Add(state.LevelError, "Failed to update type")
				} else {
					// Update form state with new type
					m.formState.FormTypeDescription = selectedType.Description
					m.notificationState.Add(state.LevelInfo, "Type updated to "+selectedType.Description)

					// Reload tasks to reflect the change
					m.reloadCurrentColumnTasks()
				}
			} else {
				// For new tasks, just update the form state
				m.formState.FormTypeDescription = selectedType.Description
				m.notificationState.Add(state.LevelInfo, "Type set to "+selectedType.Description)
			}

			// Update the selected type ID in picker state
			m.typePickerState.SetSelectedTypeID(selectedType.ID)
		}

		// Return to ticket form mode
		m.uiState.SetMode(m.typePickerState.ReturnMode())
		return m, nil
	}

	return m, nil
}

// syncParentPickerToFormState syncs parent picker selections back to form state.
// Extracts all selected task IDs and references from the picker and updates FormState.
func (m *Model) syncParentPickerToFormState() {
	var parentIDs []int
	var parentRefs []*models.TaskReference

	for _, item := range m.parentPickerState.Items {
		if item.Selected {
			parentIDs = append(parentIDs, item.TaskRef.ID)
			parentRefs = append(parentRefs, item.TaskRef)
		}
	}

	m.formState.FormParentIDs = parentIDs
	m.formState.FormParentRefs = parentRefs
}

// syncChildPickerToFormState syncs child picker selections back to form state.
// Extracts all selected task IDs and references from the picker and updates FormState.
func (m *Model) syncChildPickerToFormState() {
	var childIDs []int
	var childRefs []*models.TaskReference

	for _, item := range m.childPickerState.Items {
		if item.Selected {
			childIDs = append(childIDs, item.TaskRef.ID)
			childRefs = append(childRefs, item.TaskRef)
		}
	}

	m.formState.FormChildIDs = childIDs
	m.formState.FormChildRefs = childRefs
}

// syncLabelPickerToFormState syncs label picker selections back to form state.
// Extracts all selected label IDs from the picker and updates FormState.
func (m *Model) syncLabelPickerToFormState() {
	var labelIDs []int

	for _, item := range m.labelPickerState.Items {
		if item.Selected {
			labelIDs = append(labelIDs, item.Label.ID)
		}
	}

	m.formState.FormLabelIDs = labelIDs
}

// reloadCurrentColumnTasks reloads all task summaries for the project to keep state consistent
func (m *Model) reloadCurrentColumnTasks() {
	project := m.getCurrentProject()
	if project == nil {
		return
	}

	ctx, cancel := m.dbContext()
	defer cancel()
	tasksByColumn, err := m.repo.GetTaskSummariesByProject(ctx, project.ID)
	if err != nil {
		slog.Error("Error reloading tasks", "error", err)
		return
	}
	m.appState.SetTasks(tasksByColumn)
}
