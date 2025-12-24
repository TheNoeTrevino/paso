package handlers

import (
	"log/slog"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui"
	"github.com/thenoetrevino/paso/internal/tui/modelops"
	"github.com/thenoetrevino/paso/internal/tui/renderers"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// ============================================================================
// LABEL PICKER HANDLERS
// ============================================================================

// UpdateLabelPicker handles keyboard input in the label picker mode.
// This function processes navigation (up/down), filtering, label toggling, and label creation.
func UpdateLabelPicker(m *tui.Model, msg tea.Msg) tea.Cmd {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return nil
	}

	// Handle color picker sub-mode for creating new label
	if m.LabelPickerState.CreateMode {
		return UpdateLabelColorPicker(m, keyMsg)
	}

	// Get filtered items to determine bounds
	filteredItems := modelops.GetFilteredLabelPickerItems(m)
	maxIdx := len(filteredItems) // +1 for "create new label" option

	switch keyMsg.String() {
	case "esc":
		// Close picker and return to appropriate mode
		if m.LabelPickerState.ReturnMode == state.TicketFormMode {
			// In form mode: sync selections and return to form
			syncLabelPickerToFormState(m)
			m.UiState.SetMode(state.TicketFormMode)
		} else {
			// In view mode: return to NormalMode
			m.UiState.SetMode(state.NormalMode)
		}
		m.LabelPickerState.Filter = ""
		m.LabelPickerState.Cursor = 0
		return nil

	case "up", "k":
		// Move cursor up
		if m.LabelPickerState.Cursor > 0 {
			m.LabelPickerState.Cursor--
		}
		return nil

	case "down", "j":
		// Move cursor down
		if m.LabelPickerState.Cursor < maxIdx {
			m.LabelPickerState.Cursor++
		}
		return nil

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
						reloadCurrentColumnTasks(m)
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
		return nil

	case "backspace", "ctrl+h":
		// Remove last character from filter
		if len(m.LabelPickerState.Filter) > 0 {
			m.LabelPickerState.Filter = m.LabelPickerState.Filter[:len(m.LabelPickerState.Filter)-1]
			// Reset cursor if it's out of bounds after filter change
			newFiltered := modelops.GetFilteredLabelPickerItems(m)
			if m.LabelPickerState.Cursor > len(newFiltered) {
				m.LabelPickerState.Cursor = len(newFiltered)
			}
		}
		return nil

	default:
		// Type to filter/search
		key := keyMsg.String()
		if len(key) == 1 && len(m.LabelPickerState.Filter) < 50 {
			m.LabelPickerState.Filter += key
			// Reset cursor to 0 when filter changes
			m.LabelPickerState.Cursor = 0
		}
		return nil
	}
}

// UpdateLabelColorPicker handles keyboard input when selecting a color for new label
func UpdateLabelColorPicker(m *tui.Model, keyMsg tea.KeyMsg) tea.Cmd {
	colors := renderers.GetDefaultLabelColors()
	maxIdx := len(colors) - 1

	switch keyMsg.String() {
	case "esc":
		// Cancel and return to label list
		m.LabelPickerState.CreateMode = false
		return nil

	case "up", "k":
		if m.LabelPickerState.ColorIdx > 0 {
			m.LabelPickerState.ColorIdx--
		}
		return nil

	case "down", "j":
		if m.LabelPickerState.ColorIdx < maxIdx {
			m.LabelPickerState.ColorIdx++
		}
		return nil

	case "enter":
		// Create the new label
		color := colors[m.LabelPickerState.ColorIdx].Color
		project := modelops.GetCurrentProject(m)
		if project == nil {
			m.LabelPickerState.CreateMode = false
			return nil
		}

		ctx, cancel := m.DbContext()
		defer cancel()
		label, err := m.Repo.CreateLabel(ctx, project.ID, m.FormState.FormLabelName, color)
		if err != nil {
			slog.Error("Error creating label", "error", err)
			m.NotificationState.Add(state.LevelError, "Failed to create label")
			m.LabelPickerState.CreateMode = false
			return nil
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
		reloadCurrentColumnTasks(m)

		// Exit create mode and clear filter
		m.LabelPickerState.CreateMode = false
		m.LabelPickerState.Filter = ""
		m.LabelPickerState.Cursor = 0

		return nil
	}

	return nil
}

// ============================================================================
// PARENT PICKER HANDLERS
// ============================================================================

// UpdateParentPicker handles keyboard input in the parent picker mode.
// This function processes navigation (up/down), filtering, and selection toggling.
//
// CRITICAL - Database Parameter Ordering:
// Parent picker uses the selected task as the parent of the current task.
// When toggling relationships:
//   - AddSubtask(selectedTaskID, currentTaskID) - selected becomes parent of current
//   - RemoveSubtask(selectedTaskID, currentTaskID)
//
// This means: selectedTask BLOCKS ON currentTask (selected depends on current).
func UpdateParentPicker(m *tui.Model, msg tea.Msg) tea.Cmd {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return nil
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
			syncParentPickerToFormState(m)
		}

		m.UiState.SetMode(returnMode)
		m.ParentPickerState.Filter = ""
		m.ParentPickerState.Cursor = 0
		return nil

	case "up", "k":
		// Move cursor up
		m.ParentPickerState.MoveCursorUp()
		return nil

	case "down", "j":
		// Move cursor down
		m.ParentPickerState.MoveCursorDown(maxIdx)
		return nil

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
						reloadCurrentColumnTasks(m)
					}
					break
				}
			}
		}
		return nil

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
		return nil

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
		return nil

	default:
		// Type to filter/search
		key := keyMsg.String()
		if len(key) == 1 {
			m.ParentPickerState.AppendFilter(rune(key[0]))
			// Reset cursor to 0 when filter changes
			m.ParentPickerState.Cursor = 0
		}
		return nil
	}
}

// ============================================================================
// CHILD PICKER HANDLERS
// ============================================================================

// UpdateChildPicker handles keyboard input in the child picker mode.
// This function processes navigation (up/down), filtering, and selection toggling.
//
// CRITICAL - Database Parameter Ordering (REVERSED from parent picker):
// Child picker uses the current task as the parent of the selected task.
// When toggling relationships:
//   - AddSubtask(currentTaskID, selectedTaskID) - current becomes parent of selected
//   - RemoveSubtask(currentTaskID, selectedTaskID)
//
// This means: currentTask BLOCKS ON selectedTask (current depends on selected).
func UpdateChildPicker(m *tui.Model, msg tea.Msg) tea.Cmd {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return nil
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
			syncChildPickerToFormState(m)
		}

		m.UiState.SetMode(returnMode)
		m.ChildPickerState.Filter = ""
		m.ChildPickerState.Cursor = 0
		return nil

	case "up", "k":
		// Move cursor up
		m.ChildPickerState.MoveCursorUp()
		return nil

	case "down", "j":
		// Move cursor down
		m.ChildPickerState.MoveCursorDown(maxIdx)
		return nil

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
						reloadCurrentColumnTasks(m)
					}
					break
				}
			}
		}
		return nil

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
		return nil

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
		return nil

	default:
		// Type to filter/search
		key := keyMsg.String()
		if len(key) == 1 {
			m.ChildPickerState.AppendFilter(rune(key[0]))
			// Reset cursor to 0 when filter changes
			m.ChildPickerState.Cursor = 0
		}
		return nil
	}
}

// ============================================================================
// PRIORITY PICKER HANDLERS
// ============================================================================

// UpdatePriorityPicker handles keyboard input in the priority picker mode.
// This function processes navigation (up/down) and selection.
func UpdatePriorityPicker(m *tui.Model, msg tea.Msg) tea.Cmd {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return nil
	}

	switch keyMsg.String() {
	case "esc":
		// Return to ticket form mode without changing priority
		m.UiState.SetMode(m.PriorityPickerState.ReturnMode())
		m.PriorityPickerState.Reset()
		return nil

	case "up", "k":
		// Move cursor up
		m.PriorityPickerState.MoveUp()
		return nil

	case "down", "j":
		// Move cursor down
		m.PriorityPickerState.MoveDown()
		return nil

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
					reloadCurrentColumnTasks(m)
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
		return nil
	}

	return nil
}

// ============================================================================
// TYPE PICKER HANDLERS
// ============================================================================

// UpdateTypePicker handles keyboard input in the type picker mode.
// This function processes navigation (up/down) and selection.
func UpdateTypePicker(m *tui.Model, msg tea.Msg) tea.Cmd {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return nil
	}

	switch keyMsg.String() {
	case "esc":
		// Return to ticket form mode without changing type
		m.UiState.SetMode(m.TypePickerState.ReturnMode())
		m.TypePickerState.Reset()
		return nil

	case "up", "k":
		// Move cursor up
		m.TypePickerState.MoveUp()
		return nil

	case "down", "j":
		// Move cursor down
		m.TypePickerState.MoveDown()
		return nil

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
					reloadCurrentColumnTasks(m)
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
		return nil
	}

	return nil
}

// ============================================================================
// RELATION TYPE PICKER HANDLERS
// ============================================================================

// UpdateRelationTypePicker handles keyboard input in the relation type picker mode
func UpdateRelationTypePicker(m *tui.Model, msg tea.Msg) tea.Cmd {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return nil
	}

	switch keyMsg.String() {
	case "esc":
		// Return to previous picker (parent or child) without changing relation type
		m.UiState.SetMode(m.RelationTypePickerState.ReturnMode())
		m.RelationTypePickerState.Reset()
		return nil

	case "up", "k":
		// Move cursor up
		m.RelationTypePickerState.MoveUp()
		return nil

	case "down", "j":
		// Move cursor down
		m.RelationTypePickerState.MoveDown()
		return nil

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
		return nil
	}

	return nil
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// syncParentPickerToFormState syncs parent picker selections back to form state.
// Extracts all selected task IDs and references from the picker and updates FormState.
func syncParentPickerToFormState(m *tui.Model) {
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
func syncChildPickerToFormState(m *tui.Model) {
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
func syncLabelPickerToFormState(m *tui.Model) {
	var labelIDs []int

	for _, item := range m.LabelPickerState.Items {
		if item.Selected {
			labelIDs = append(labelIDs, item.Label.ID)
		}
	}

	m.FormState.FormLabelIDs = labelIDs
}

// reloadCurrentColumnTasks reloads all task summaries for the project to keep state consistent
func reloadCurrentColumnTasks(m *tui.Model) {
	project := modelops.GetCurrentProject(m)
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
