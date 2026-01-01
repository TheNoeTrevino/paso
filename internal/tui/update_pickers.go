package tui

import (
	"log/slog"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/thenoetrevino/paso/internal/models"
	labelservice "github.com/thenoetrevino/paso/internal/services/label"
	taskservice "github.com/thenoetrevino/paso/internal/services/task"
	"github.com/thenoetrevino/paso/internal/tui/renderers"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

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
			m.UIState.SetMode(state.TicketFormMode)
		} else {
			// In view mode: return to NormalMode
			m.UIState.SetMode(state.NormalMode)
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
						ctx, cancel := m.UIContext()
						defer cancel()
						if m.LabelPickerState.Items[i].Selected {
							// Remove label from task
							err := m.App.TaskService.DetachLabel(ctx, m.LabelPickerState.TaskID, item.Label.ID)
							if err != nil {
								slog.Error("Error removing label", "error", err)
								m.NotificationState.Add(state.LevelError, "Failed to remove label from task")
							} else {
								m.LabelPickerState.Items[i].Selected = false
							}
						} else {
							// Add label to task
							err := m.App.TaskService.AttachLabel(ctx, m.LabelPickerState.TaskID, item.Label.ID)
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

		ctx, cancel := m.DBContext()
		defer cancel()
		label, err := m.App.LabelService.CreateLabel(ctx, labelservice.CreateLabelRequest{
			ProjectID: project.ID,
			Name:      m.FormState.FormLabelName,
			Color:     color,
		})
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
		err = m.App.TaskService.AttachLabel(ctx, m.LabelPickerState.TaskID, label.ID)
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

		m.UIState.SetMode(returnMode)
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
						ctx, cancel := m.UIContext()
						defer cancel()
						if m.ParentPickerState.Items[i].Selected {
							// Remove parent relationship
							// CRITICAL: RemoveParentRelation(childID, parentID)
							// selectedTask (parent) blocks on currentTask (child)
							err := m.App.TaskService.RemoveParentRelation(ctx, m.ParentPickerState.TaskID, item.TaskRef.ID)
							if err != nil {
								slog.Error("Error removing parent", "error", err)
								m.NotificationState.Add(state.LevelError, "Failed to remove parent from task")
							} else {
								m.ParentPickerState.Items[i].Selected = false
							}
						} else {
							// Add parent relationship - selected task becomes parent of current task
							// CRITICAL: AddParentRelation(childID, parentID, relationTypeID)
							// This makes selectedTask (parent) block on currentTask (child)
							// Meaning: selectedTask depends on completion of currentTask
							err := m.App.TaskService.AddParentRelation(ctx, m.ParentPickerState.TaskID, item.TaskRef.ID, 1)
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
			m.RelationTypePickerState.ReturnMode = state.ParentPickerMode

			// Set cursor to match selected relation type
			relationTypes := renderers.GetRelationTypeOptions()
			for i, rt := range relationTypes {
				if rt.ID == currentRelationTypeID {
					m.RelationTypePickerState.SetCursor(i)
					break
				}
			}

			m.UIState.SetMode(state.RelationTypePickerMode)
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

		m.UIState.SetMode(returnMode)
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
						ctx, cancel := m.UIContext()
						defer cancel()
						if m.ChildPickerState.Items[i].Selected {
							// Remove child relationship
							// CRITICAL: RemoveChildRelation(parentID, childID)
							// currentTask (parent) blocks on selectedTask (child)
							err := m.App.TaskService.RemoveChildRelation(ctx, m.ChildPickerState.TaskID, item.TaskRef.ID)
							if err != nil {
								slog.Error("Error removing child", "error", err)
								m.NotificationState.Add(state.LevelError, "Failed to remove child from task")
							} else {
								m.ChildPickerState.Items[i].Selected = false
							}
						} else {
							// Add child relationship - current task becomes parent of selected task
							// CRITICAL: AddChildRelation(parentID, childID, relationTypeID)
							// This makes currentTask (parent) block on selectedTask (child)
							// Meaning: currentTask depends on completion of selectedTask
							err := m.App.TaskService.AddChildRelation(ctx, m.ChildPickerState.TaskID, item.TaskRef.ID, 1)
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
			m.RelationTypePickerState.ReturnMode = state.ChildPickerMode

			// Set cursor to match selected relation type
			relationTypes := renderers.GetRelationTypeOptions()
			for i, rt := range relationTypes {
				if rt.ID == currentRelationTypeID {
					m.RelationTypePickerState.SetCursor(i)
					break
				}
			}

			m.UIState.SetMode(state.RelationTypePickerMode)
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
		m.UIState.SetMode(m.PriorityPickerState.ReturnMode)
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
				ctx, cancel := m.DBContext()
				defer cancel()

				// Update the task's priority in the database
				priorityID := selectedPriority.ID
				err := m.App.TaskService.UpdateTask(ctx, taskservice.UpdateTaskRequest{
					TaskID:     m.FormState.EditingTaskID,
					PriorityID: &priorityID,
				})

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
		m.UIState.SetMode(m.PriorityPickerState.ReturnMode)
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
		m.UIState.SetMode(m.TypePickerState.ReturnMode)
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
				ctx, cancel := m.DBContext()
				defer cancel()

				// Update the task's type in the database
				typeID := selectedType.ID
				err := m.App.TaskService.UpdateTask(ctx, taskservice.UpdateTaskRequest{
					TaskID: m.FormState.EditingTaskID,
					TypeID: &typeID,
				})

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
		m.UIState.SetMode(m.TypePickerState.ReturnMode)
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
		m.UIState.SetMode(m.RelationTypePickerState.ReturnMode)
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
			returnMode := m.RelationTypePickerState.ReturnMode

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
		m.UIState.SetMode(m.RelationTypePickerState.ReturnMode)
		return m, nil
	}

	return m, nil
}

// buildTaskRefWithRelationType creates a TaskReference with populated relation type fields.
// It looks up the relation type by ID and populates the appropriate label based on perspective.
func buildTaskRefWithRelationType(
	taskRef *models.TaskReference,
	relationTypeID int,
	relationMap map[int]renderers.RelationTypeOption,
	useParentPerspective bool,
) *models.TaskReference {
	ref := &models.TaskReference{
		ID:             taskRef.ID,
		TicketNumber:   taskRef.TicketNumber,
		Title:          taskRef.Title,
		ProjectName:    taskRef.ProjectName,
		RelationTypeID: relationTypeID,
	}

	// Populate relation type display fields from the relation type ID
	if relOpt, ok := relationMap[relationTypeID]; ok {
		if useParentPerspective {
			ref.RelationLabel = relOpt.PToCLabel // Parent's perspective
		} else {
			ref.RelationLabel = relOpt.CToPLabel // Child's perspective
		}
		ref.RelationColor = relOpt.Color
		ref.IsBlocking = relOpt.IsBlocking
	}

	return ref
}

// getRelationTypeMap returns a map of relation type IDs to RelationTypeOptions.
// This avoids repeatedly building the same map from the slice.
func getRelationTypeMap() map[int]renderers.RelationTypeOption {
	relationOptions := renderers.GetRelationTypeOptions()
	relationMap := make(map[int]renderers.RelationTypeOption, len(relationOptions))
	for _, opt := range relationOptions {
		relationMap[opt.ID] = opt
	}
	return relationMap
}

// syncParentPickerToFormState syncs parent picker selections back to form state.
// Extracts all selected task IDs and references from the picker and updates FormState.
func (m *Model) syncParentPickerToFormState() {
	var parentIDs []int
	var parentRefs []*models.TaskReference

	relationMap := getRelationTypeMap()

	for _, item := range m.ParentPickerState.Items {
		if item.Selected {
			parentIDs = append(parentIDs, item.TaskRef.ID)
			ref := buildTaskRefWithRelationType(item.TaskRef, item.RelationTypeID, relationMap, true)
			parentRefs = append(parentRefs, ref)
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

	relationMap := getRelationTypeMap()

	for _, item := range m.ChildPickerState.Items {
		if item.Selected {
			childIDs = append(childIDs, item.TaskRef.ID)
			ref := buildTaskRefWithRelationType(item.TaskRef, item.RelationTypeID, relationMap, false)
			childRefs = append(childRefs, ref)
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

func (m *Model) reloadCurrentColumnTasks() {
	project := m.getCurrentProject()
	if project == nil {
		return
	}

	ctx, cancel := m.DBContext()
	defer cancel()
	tasksByColumn, err := m.App.TaskService.GetTaskSummariesByProject(ctx, project.ID)
	if err != nil {
		slog.Error("Error reloading tasks", "error", err)
		return
	}
	m.AppState.SetTasks(tasksByColumn)
}
