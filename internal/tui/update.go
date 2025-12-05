package tui

import (
	"context"
	"log"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
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

// ticketFormValues holds the extracted values from a completed ticket form
type ticketFormValues struct {
	title       string
	description string
	confirm     bool
	labelIDs    []int
}

// extractTicketFormValues extracts and returns form values from the ticket form
func (m Model) extractTicketFormValues() ticketFormValues {
	title := m.formState.TicketForm.GetString("title")
	description := m.formState.TicketForm.GetString("description")

	confirm := true
	if c := m.formState.TicketForm.Get("confirm"); c != nil {
		if b, ok := c.(bool); ok {
			confirm = b
		}
	}

	var labelIDs []int
	if labels := m.formState.TicketForm.Get("labels"); labels != nil {
		if ids, ok := labels.([]int); ok {
			labelIDs = ids
		}
	}

	return ticketFormValues{
		title:       strings.TrimSpace(title),
		description: strings.TrimSpace(description),
		confirm:     confirm,
		labelIDs:    labelIDs,
	}
}

// createNewTaskWithLabels creates a new task and sets its labels
func (m Model) createNewTaskWithLabels(values ticketFormValues) {
	currentCol := m.appState.Columns()[m.uiState.SelectedColumn()]
	task, err := m.repo.CreateTask(context.Background(),
		values.title,
		values.description,
		currentCol.ID,
		len(m.appState.Tasks()[currentCol.ID]),
	)
	if err != nil {
		log.Printf("Error creating task: %v", err)
		m.notificationState.Add(state.LevelError, "Error creating task")
		return
	}

	// Set labels
	if len(values.labelIDs) > 0 {
		err = m.repo.SetTaskLabels(context.Background(), task.ID, values.labelIDs)
		if err != nil {
			log.Printf("Error setting labels: %v", err)
		}
	}

	// Reload task summary with labels
	summaries, err := m.repo.GetTaskSummariesByColumn(context.Background(), currentCol.ID)
	if err != nil {
		log.Printf("Error reloading tasks: %v", err)
	} else {
		m.appState.Tasks()[currentCol.ID] = summaries
	}
}

// updateExistingTaskWithLabels updates an existing task and its labels
func (m Model) updateExistingTaskWithLabels(values ticketFormValues) {
	err := m.repo.UpdateTask(context.Background(), m.formState.EditingTaskID, values.title, values.description)
	if err != nil {
		log.Printf("Error updating task: %v", err)
		m.notificationState.Add(state.LevelError, "Error updating task")
		return
	}

	// Update labels
	err = m.repo.SetTaskLabels(context.Background(), m.formState.EditingTaskID, values.labelIDs)
	if err != nil {
		log.Printf("Error setting labels: %v", err)
	}

	// Reload task summaries for the column
	currentCol := m.appState.Columns()[m.uiState.SelectedColumn()]
	summaries, err := m.repo.GetTaskSummariesByColumn(context.Background(), currentCol.ID)
	if err != nil {
		log.Printf("Error reloading tasks: %v", err)
	} else {
		m.appState.Tasks()[currentCol.ID] = summaries
	}
}

// formConfig holds configuration for generic form handling
type formConfig struct {
	form       *huh.Form
	setForm    func(*huh.Form)
	clearForm  func()
	onComplete func() // Called when form completes successfully
}

// handleFormUpdate processes form messages generically
func (m Model) handleFormUpdate(msg tea.Msg, cfg formConfig) (tea.Model, tea.Cmd) {
	if cfg.form == nil {
		m.uiState.SetMode(state.NormalMode)
		return m, nil
	}

	// Escape key handling
	if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "esc" {
		m.uiState.SetMode(state.NormalMode)
		cfg.setForm(nil)
		return m, nil
	}

	// Forward to form
	var cmds []tea.Cmd
	form, cmd := cfg.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		cfg.setForm(f)
		cmds = append(cmds, cmd)
	}

	// Check completion
	if cfg.form.State == huh.StateCompleted {
		cfg.onComplete()
		m.uiState.SetMode(state.NormalMode)
		cfg.setForm(nil)
		cfg.clearForm()
		return m, tea.ClearScreen
	}

	// Check abort
	if cfg.form.State == huh.StateAborted {
		m.uiState.SetMode(state.NormalMode)
		cfg.setForm(nil)
		cfg.clearForm()
		return m, tea.ClearScreen
	}

	return m, tea.Batch(cmds...)
}

// updateTicketForm handles all messages when in TicketFormMode
// This is separated out because huh forms need to receive ALL messages, not just KeyMsg
func (m Model) updateTicketForm(msg tea.Msg) (tea.Model, tea.Cmd) {
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
					m.createNewTaskWithLabels(values)
				} else {
					m.updateExistingTaskWithLabels(values)
				}
			}
		},
	})
}

// updateProjectForm handles all messages when in ProjectFormMode
// This is separated out because huh forms need to receive ALL messages, not just KeyMsg
func (m Model) updateProjectForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.handleFormUpdate(msg, formConfig{
		form: m.formState.ProjectForm,
		setForm: func(f *huh.Form) {
			m.formState.ProjectForm = f
		},
		clearForm: func() {
			m.formState.ClearProjectForm()
		},
		onComplete: func() {
			// Read values directly from the form using GetString
			name := m.formState.ProjectForm.GetString("name")
			description := m.formState.ProjectForm.GetString("description")

			// Form submitted - create the project
			if strings.TrimSpace(name) != "" {
				project, err := m.repo.CreateProject(context.Background(), strings.TrimSpace(name), strings.TrimSpace(description))
				if err != nil {
					log.Printf("Error creating project: %v", err)
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
						err := m.repo.RemoveLabelFromTask(context.Background(), m.labelPickerState.TaskID, item.Label.ID)
						if err != nil {
							log.Printf("Error removing label: %v", err)
							m.notificationState.Add(state.LevelError, "Failed to remove label from task")
						} else {
							m.labelPickerState.Items[i].Selected = false
						}
					} else {
						// Add label to task
						err := m.repo.AddLabelToTask(context.Background(), m.labelPickerState.TaskID, item.Label.ID)
						if err != nil {
							log.Printf("Error adding label: %v", err)
							m.notificationState.Add(state.LevelError, "Failed to add label to task")
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

		label, err := m.repo.CreateLabel(context.Background(), project.ID, m.formState.FormLabelName, color)
		if err != nil {
			log.Printf("Error creating label: %v", err)
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
		err = m.repo.AddLabelToTask(context.Background(), m.labelPickerState.TaskID, label.ID)
		if err != nil {
			log.Printf("Error assigning new label to task: %v", err)
			m.notificationState.Add(state.LevelError, "Failed to assign label to task")
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

	taskDetail, err := m.repo.GetTaskDetail(context.Background(), m.uiState.ViewingTask().ID)
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
	summaries, err := m.repo.GetTaskSummariesByColumn(context.Background(), currentCol.ID)
	if err != nil {
		log.Printf("Error reloading column tasks: %v", err)
		return
	}
	m.appState.Tasks()[currentCol.ID] = summaries
}
