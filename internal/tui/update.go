package tui

import (
	"context"
	"log"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/thenoetrevino/paso/internal/database"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// Update handles all messages and updates the model accordingly
// This implements the "Update" part of the Model-View-Update pattern
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle form modes first - forms need ALL messages
	if m.uiState.Mode() == state.TicketFormMode {
		return m.updateTicketForm(msg)
	}
	if m.uiState.Mode() == state.ProjectFormMode {
		return m.updateProjectForm(msg)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case tea.WindowSizeMsg:
		return m.handleWindowResize(msg)
	}

	return m, nil
}

// handleKeyMsg dispatches key messages to the appropriate mode handler.
func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.uiState.Mode() {
	case state.NormalMode:
		return m.handleNormalMode(msg)
	case state.AddColumnMode, state.EditColumnMode:
		return m.handleInputMode(msg)
	case state.DeleteConfirmMode:
		return m.handleDeleteConfirm(msg)
	case state.DeleteColumnConfirmMode:
		return m.handleDeleteColumnConfirm(msg)
	case state.HelpMode:
		return m.handleHelpMode(msg)
	case state.ViewTaskMode:
		return m.handleViewTaskMode(msg)
	case state.LabelPickerMode:
		return m.updateLabelPicker(msg)
	}
	return m, nil
}

// handleWindowResize handles terminal resize events.
func (m Model) handleWindowResize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.uiState.SetWidth(msg.Width)
	m.uiState.SetHeight(msg.Height)

	// Ensure viewport offset is still valid after resize
	if m.uiState.ViewportOffset()+m.uiState.ViewportSize() > len(m.appState.Columns()) {
		m.uiState.SetViewportOffset(max(0, len(m.appState.Columns())-m.uiState.ViewportSize()))
	}
	return m, nil
}

// updateTicketForm handles all messages when in TicketFormMode
// This is separated out because huh forms need to receive ALL messages, not just KeyMsg
func (m Model) updateTicketForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.formState.TicketForm == nil {
		m.uiState.SetMode(state.NormalMode)
		return m, nil
	}

	// Check for escape key to cancel
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if keyMsg.String() == "esc" {
			m.uiState.SetMode(state.NormalMode)
			m.formState.TicketForm = nil
			return m, nil
		}
	}

	// Forward ALL messages to the form
	var cmds []tea.Cmd
	form, cmd := m.formState.TicketForm.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.formState.TicketForm = f
		cmds = append(cmds, cmd)
	}

	// Check if form is complete
	if m.formState.TicketForm.State == huh.StateCompleted {
		// Read values directly from the form (not from bound pointers which point to stale model copies)
		title := m.formState.TicketForm.GetString("title")
		description := m.formState.TicketForm.GetString("description")

		confirm := true
		if c := m.formState.TicketForm.Get("confirm"); c != nil {
			if b, ok := c.(bool); ok {
				confirm = b
			}
		}

		// Get label IDs - need type assertion since it's a generic Get
		var labelIDs []int
		if labels := m.formState.TicketForm.Get("labels"); labels != nil {
			if ids, ok := labels.([]int); ok {
				labelIDs = ids
			}
		}

		// Form submitted - check confirmation and save the task
		if !confirm {
			// User selected "No" on confirmation
			m.uiState.SetMode(state.NormalMode)
			m.formState.TicketForm = nil
			m.formState.EditingTaskID = 0
			m.formState.ClearTicketForm()
			return m, tea.ClearScreen
		}
		if strings.TrimSpace(title) != "" {
			if m.formState.EditingTaskID == 0 {
				// Create new task
				currentCol := m.appState.Columns()[m.uiState.SelectedColumn()]
				task, err := database.CreateTask(context.Background(), 
					m.db,
					strings.TrimSpace(title),
					strings.TrimSpace(description),
					currentCol.ID,
					len(m.appState.Tasks()[currentCol.ID]),
				)
				if err != nil {
					log.Printf("Error creating task: %v", err)
					m.errorState.Set("Error creating task")
				} else {
					// Set labels
					if len(labelIDs) > 0 {
						err = database.SetTaskLabels(context.Background(), m.db, task.ID, labelIDs)
						if err != nil {
							log.Printf("Error setting labels: %v", err)
						}
					}
					// Reload task summary with labels
					summaries, err := database.GetTaskSummariesByColumn(context.Background(), m.db, currentCol.ID)
					if err != nil {
						log.Printf("Error reloading tasks: %v", err)
					} else {
						m.appState.Tasks()[currentCol.ID] = summaries
					}
				}
			} else {
				// Update existing task
				err := database.UpdateTask(context.Background(), m.db, m.formState.EditingTaskID, strings.TrimSpace(title), strings.TrimSpace(description))
				if err != nil {
					log.Printf("Error updating task: %v", err)
					m.errorState.Set("Error updating task")
				} else {
					// Update labels
					err = database.SetTaskLabels(context.Background(), m.db, m.formState.EditingTaskID, labelIDs)
					if err != nil {
						log.Printf("Error setting labels: %v", err)
					}
					// Reload task summaries for the column
					currentCol := m.appState.Columns()[m.uiState.SelectedColumn()]
					summaries, err := database.GetTaskSummariesByColumn(context.Background(), m.db, currentCol.ID)
					if err != nil {
						log.Printf("Error reloading tasks: %v", err)
					} else {
						m.appState.Tasks()[currentCol.ID] = summaries
					}
				}
			}
		}
		m.uiState.SetMode(state.NormalMode)
		m.formState.TicketForm = nil
		m.formState.EditingTaskID = 0
		m.formState.ClearTicketForm()
		return m, tea.ClearScreen
	}

	// Check if form was aborted
	if m.formState.TicketForm.State == huh.StateAborted {
		m.uiState.SetMode(state.NormalMode)
		m.formState.TicketForm = nil
		m.formState.EditingTaskID = 0
		m.formState.ClearTicketForm()
		return m, tea.ClearScreen
	}

	return m, tea.Batch(cmds...)
}

// updateProjectForm handles all messages when in ProjectFormMode
// This is separated out because huh forms need to receive ALL messages, not just KeyMsg
func (m Model) updateProjectForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.formState.ProjectForm == nil {
		m.uiState.SetMode(state.NormalMode)
		return m, nil
	}

	// Check for escape key to cancel
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if keyMsg.String() == "esc" {
			m.uiState.SetMode(state.NormalMode)
			m.formState.ProjectForm = nil
			return m, nil
		}
	}

	// Forward ALL messages to the form
	var cmds []tea.Cmd
	form, cmd := m.formState.ProjectForm.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.formState.ProjectForm = f
		cmds = append(cmds, cmd)
	}

	// Check if form is complete
	if m.formState.ProjectForm.State == huh.StateCompleted {
		// Read values directly from the form using GetString
		name := m.formState.ProjectForm.GetString("name")
		description := m.formState.ProjectForm.GetString("description")

		// Form submitted - create the project
		if strings.TrimSpace(name) != "" {
			project, err := database.CreateProject(context.Background(), m.db, strings.TrimSpace(name), strings.TrimSpace(description))
			if err != nil {
				log.Printf("Error creating project: %v", err)
				m.errorState.Set("Error creating project")
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
		m.uiState.SetMode(state.NormalMode)
		m.formState.ProjectForm = nil
		m.formState.ClearProjectForm()
		return m, tea.ClearScreen
	}

	// Check if form was aborted
	if m.formState.ProjectForm.State == huh.StateAborted {
		m.uiState.SetMode(state.NormalMode)
		m.formState.ProjectForm = nil
		m.formState.ClearProjectForm()
		return m, tea.ClearScreen
	}

	return m, tea.Batch(cmds...)
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
		// Close picker and return to ViewTaskMode
		m.uiState.SetMode(state.ViewTaskMode)
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
					if m.labelPickerState.Items[i].Selected {
						// Remove label from task
						err := database.RemoveLabelFromTask(context.Background(), m.db, m.labelPickerState.TaskID, item.Label.ID)
						if err != nil {
							log.Printf("Error removing label: %v", err)
						} else {
							m.labelPickerState.Items[i].Selected = false
						}
					} else {
						// Add label to task
						err := database.AddLabelToTask(context.Background(), m.db, m.labelPickerState.TaskID, item.Label.ID)
						if err != nil {
							log.Printf("Error adding label: %v", err)
						} else {
							m.labelPickerState.Items[i].Selected = true
						}
					}
					break
				}
			}

			// Reload task detail to update the view
			m.reloadViewingTask()
			// Reload task summaries for the current column
			m.reloadCurrentColumnTasks()
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

		label, err := database.CreateLabel(context.Background(), m.db, project.ID, m.formState.FormLabelName, color)
		if err != nil {
			log.Printf("Error creating label: %v", err)
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
		err = database.AddLabelToTask(context.Background(), m.db, m.labelPickerState.TaskID, label.ID)
		if err != nil {
			log.Printf("Error assigning new label to task: %v", err)
		}

		// Reload task detail to update the view
		m.reloadViewingTask()
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

// reloadViewingTask reloads the task detail being viewed
func (m *Model) reloadViewingTask() {
	if m.uiState.ViewingTask() == nil {
		return
	}

	taskDetail, err := database.GetTaskDetail(context.Background(), m.db, m.uiState.ViewingTask().ID)
	if err != nil {
		log.Printf("Error reloading task detail: %v", err)
		return
	}
	m.uiState.SetViewingTask(taskDetail)
}

// reloadCurrentColumnTasks reloads task summaries for the current column
func (m *Model) reloadCurrentColumnTasks() {
	if len(m.appState.Columns()) == 0 || m.uiState.SelectedColumn() >= len(m.appState.Columns()) {
		return
	}

	currentCol := m.appState.Columns()[m.uiState.SelectedColumn()]
	summaries, err := database.GetTaskSummariesByColumn(context.Background(), m.db, currentCol.ID)
	if err != nil {
		log.Printf("Error reloading column tasks: %v", err)
		return
	}
	m.appState.Tasks()[currentCol.ID] = summaries
}
