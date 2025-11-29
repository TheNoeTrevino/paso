package tui

import (
	"log"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/thenoetrevino/paso/internal/database"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// Update handles all messages and updates the model accordingly
// This implements the "Update" part of the Model-View-Update pattern
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle TicketFormMode first - form needs ALL messages, not just KeyMsg
	if m.uiState.Mode() == state.TicketFormMode {
		return m.updateTicketForm(msg)
	}

	// Handle ProjectFormMode - form needs ALL messages, not just KeyMsg
	if m.uiState.Mode() == state.ProjectFormMode {
		return m.updateProjectForm(msg)
	}

	switch msg := msg.(type) {

	case tea.KeyMsg:
		// Handle keyboard input based on current mode
		if m.uiState.Mode() == state.NormalMode {
			// Normal mode: navigation and command keys
			// Clear any previous error messages
			m.errorState.Clear()

			switch msg.String() {
			case "q", "ctrl+c":
				// Quit the application
				return m, tea.Quit

			case "?":
				// Show help screen
				m.uiState.SetMode(state.HelpMode)
				return m, nil

			case "a":
				// Add new task in current column using ticket form
				if len(m.appState.Columns()) == 0 {
					m.errorState.Set("Cannot add task: No columns exist. Create a column first with 'C'")
					return m, nil
				}
				// Initialize form fields (old fields still needed for form binding)
				m.formTitle = ""
				m.formDescription = ""
				m.formLabelIDs = []int{}
				m.formConfirm = true
				m.editingTaskID = 0
				// Sync to FormState
				m.formState.SetFormTitle(m.formTitle)
				m.formState.SetFormDescription(m.formDescription)
				m.formState.SetFormLabelIDs(m.formLabelIDs)
				m.formState.SetFormConfirm(m.formConfirm)
				m.formState.SetEditingTaskID(m.editingTaskID)
				// Create the form (binds to old fields)
				m.ticketForm = CreateTicketForm(
					&m.formTitle,
					&m.formDescription,
					&m.formLabelIDs,
					m.appState.Labels(),
					&m.formConfirm,
				)
				m.formState.SetTicketForm(m.ticketForm)
				m.uiState.SetMode(state.TicketFormMode)
				return m, m.ticketForm.Init()

			case "e":
				// Edit selected task using ticket form
				task := m.getCurrentTask()
				if task != nil {
					// Load full task details
					taskDetail, err := database.GetTaskDetail(m.db, task.ID)
					if err != nil {
						log.Printf("Error loading task details: %v", err)
						m.errorState.Set("Error loading task details")
						return m, nil
					}
					// Initialize form fields with existing data
					m.formTitle = taskDetail.Title
					m.formDescription = taskDetail.Description
					m.formLabelIDs = make([]int, len(taskDetail.Labels))
					for i, label := range taskDetail.Labels {
						m.formLabelIDs[i] = label.ID
					}
					m.formConfirm = true
					m.editingTaskID = task.ID
					// Sync to FormState
					m.formState.SetFormTitle(m.formTitle)
					m.formState.SetFormDescription(m.formDescription)
					m.formState.SetFormLabelIDs(m.formLabelIDs)
					m.formState.SetFormConfirm(m.formConfirm)
					m.formState.SetEditingTaskID(m.editingTaskID)
					// Create the form (binds to old fields)
					m.ticketForm = CreateTicketForm(
						&m.formTitle,
						&m.formDescription,
						&m.formLabelIDs,
						m.appState.Labels(),
						&m.formConfirm,
					)
					m.formState.SetTicketForm(m.ticketForm)
					m.uiState.SetMode(state.TicketFormMode)
					return m, m.ticketForm.Init()
				} else {
					m.errorState.Set("No task selected to edit")
				}
				return m, nil

			case "d":
				// Delete selected task (if one exists)
				if m.getCurrentTask() != nil {
					m.uiState.SetMode(state.DeleteConfirmMode)
				} else {
					m.errorState.Set("No task selected to delete")
				}
				return m, nil

			case " ": // Space key
				// View task details (if one exists)
				task := m.getCurrentTask()
				if task != nil {
					// Load full task detail from database
					taskDetail, err := database.GetTaskDetail(m.db, task.ID)
					if err != nil {
						log.Printf("Error loading task details: %v", err)
						m.errorState.Set("Error loading task details")
						return m, nil
					}
					m.uiState.SetViewingTask(taskDetail)
					m.uiState.SetMode(state.ViewTaskMode)
				} else {
					m.errorState.Set("No task selected to view")
				}
				return m, nil

			case "C":
				// Create new column
				m.uiState.SetMode(state.AddColumnMode)
				m.inputPrompt = "New column name:"
				m.inputBuffer = ""
				// Sync to InputState
				m.inputState.SetPrompt(m.inputPrompt)
				m.inputState.SetBuffer(m.inputBuffer)
				return m, nil

			case "R":
				// Rename selected column (if one exists)
				column := m.getCurrentColumn()
				if column != nil {
					m.uiState.SetMode(state.EditColumnMode)
					m.inputBuffer = column.Name
					m.inputPrompt = "Rename column:"
					// Sync to InputState
					m.inputState.SetBuffer(m.inputBuffer)
					m.inputState.SetPrompt(m.inputPrompt)
				} else {
					m.errorState.Set("No column selected to rename")
				}
				return m, nil

			case "X":
				// Delete selected column (if one exists)
				column := m.getCurrentColumn()
				if column != nil {
					// Get task count for warning
					taskCount, err := database.GetTaskCountByColumn(m.db, column.ID)
					if err != nil {
						log.Printf("Error getting task count: %v", err)
						m.errorState.Set("Error getting column info")
						return m, nil
					}
					m.deleteColumnTaskCount = taskCount
					m.inputState.SetDeleteColumnTaskCount(taskCount)
					m.uiState.SetMode(state.DeleteColumnConfirmMode)
				} else {
					m.errorState.Set("No column selected to delete")
				}
				return m, nil

			// Viewport scrolling: Move the viewport window
			case "]":
				// Scroll viewport right (show columns to the right)
				if m.uiState.ViewportOffset()+m.uiState.ViewportSize() < len(m.appState.Columns()) {
					m.uiState.SetViewportOffset(m.uiState.ViewportOffset() + 1)
					// Adjust selectedColumn if it's now off-screen to the left
					if m.uiState.SelectedColumn() < m.uiState.ViewportOffset() {
						m.uiState.SetSelectedColumn(m.uiState.ViewportOffset())
						m.uiState.SetSelectedTask(0)
					}
				}

			case "[":
				// Scroll viewport left (show columns to the left)
				if m.uiState.ViewportOffset() > 0 {
					m.uiState.SetViewportOffset(m.uiState.ViewportOffset() - 1)
					// Adjust selectedColumn if it's now off-screen to the right
					if m.uiState.SelectedColumn() >= m.uiState.ViewportOffset()+m.uiState.ViewportSize() {
						m.uiState.SetSelectedColumn(m.uiState.ViewportOffset() + m.uiState.ViewportSize() - 1)
						m.uiState.SetSelectedTask(0)
					}
				}

			// Left/Right navigation: Move between columns
			case "h", "left":
				// Move to previous column if not at first column
				if m.uiState.SelectedColumn() > 0 {
					m.uiState.SetSelectedColumn(m.uiState.SelectedColumn() - 1)
					m.uiState.SetSelectedTask(0) // Reset task selection when switching columns
					// Auto-scroll viewport if selected column is now off-screen to the left
					if m.uiState.SelectedColumn() < m.uiState.ViewportOffset() {
						m.uiState.SetViewportOffset(m.uiState.SelectedColumn())
					}
				}

			case "l", "right":
				// Move to next column if not at last column
				if m.uiState.SelectedColumn() < len(m.appState.Columns())-1 {
					m.uiState.SetSelectedColumn(m.uiState.SelectedColumn() + 1)
					m.uiState.SetSelectedTask(0) // Reset task selection when switching columns
					// Auto-scroll viewport if selected column is now off-screen to the right
					if m.uiState.SelectedColumn() >= m.uiState.ViewportOffset()+m.uiState.ViewportSize() {
						m.uiState.SetViewportOffset(m.uiState.SelectedColumn() - m.uiState.ViewportSize() + 1)
					}
				}

			// Up/Down navigation: Move between tasks in current column
			case "j", "down":
				// Move to next task in current column
				currentTasks := m.getCurrentTasks()
				if len(currentTasks) > 0 && m.uiState.SelectedTask() < len(currentTasks)-1 {
					m.uiState.SetSelectedTask(m.uiState.SelectedTask() + 1)
				}

			case "k", "up":
				// Move to previous task in current column
				if m.uiState.SelectedTask() > 0 {
					m.uiState.SetSelectedTask(m.uiState.SelectedTask() - 1)
				}

			// Task movement: Move tasks between columns
			case ">", "L":
				// Move task to next column (right)
				if m.getCurrentTask() != nil {
					m.moveTaskRight()
				}

			case "<", "H":
				// Move task to previous column (left)
				if m.getCurrentTask() != nil {
					m.moveTaskLeft()
				}

			// Project tab navigation
			case "{":
				// Switch to previous project
				if m.appState.SelectedProject() > 0 {
					m.switchToProject(m.appState.SelectedProject() - 1)
				}

			case "}":
				// Switch to next project
				if m.appState.SelectedProject() < len(m.appState.Projects())-1 {
					m.switchToProject(m.appState.SelectedProject() + 1)
				}

			case "ctrl+p":
				// Create new project
				m.formProjectName = ""
				m.formProjectDescription = ""
				// Sync to FormState
				m.formState.SetFormProjectName(m.formProjectName)
				m.formState.SetFormProjectDescription(m.formProjectDescription)
				m.projectForm = CreateProjectForm(
					&m.formProjectName,
					&m.formProjectDescription,
				)
				m.formState.SetProjectForm(m.projectForm)
				m.uiState.SetMode(state.ProjectFormMode)
				return m, m.projectForm.Init()
			}

		} else if m.uiState.Mode() == state.AddColumnMode || m.uiState.Mode() == state.EditColumnMode {
			// Input modes: handle text input for column operations
			switch msg.String() {
			case "enter":
				// Confirm input and create/edit column
				if strings.TrimSpace(m.inputBuffer) != "" {
					if m.uiState.Mode() == state.AddColumnMode {
						// Create new column after the current column in the current project
						var afterColumnID *int
						if len(m.appState.Columns()) > 0 {
							currentCol := m.appState.Columns()[m.uiState.SelectedColumn()]
							afterColumnID = &currentCol.ID
						}
						// Get current project ID
						projectID := 0
						if project := m.getCurrentProject(); project != nil {
							projectID = project.ID
						}
						column, err := database.CreateColumn(m.db, strings.TrimSpace(m.inputBuffer), projectID, afterColumnID)
						if err != nil {
							log.Printf("Error creating column: %v", err)
						} else {
							// Reload columns from database to get correct order
							columns, err := database.GetColumnsByProject(m.db, projectID)
							if err != nil {
								log.Printf("Error reloading columns: %v", err)
							}
							m.appState.SetColumns(columns)
							m.appState.Tasks()[column.ID] = []*models.TaskSummary{}
							// Move selection to new column (it will be after current)
							if afterColumnID != nil {
								m.uiState.SetSelectedColumn(m.uiState.SelectedColumn() + 1)
							}
						}
					} else if m.uiState.Mode() == state.EditColumnMode {
						// Update existing column
						column := m.getCurrentColumn()
						if column != nil {
							err := database.UpdateColumnName(m.db, column.ID, strings.TrimSpace(m.inputBuffer))
							if err != nil {
								log.Printf("Error updating column: %v", err)
							} else {
								column.Name = strings.TrimSpace(m.inputBuffer)
							}
						}
					}
				}
				// Return to normal mode
				m.uiState.SetMode(state.NormalMode)
				m.inputBuffer = ""
				m.inputPrompt = ""
				m.inputState.Clear()
				return m, nil

			case "esc":
				// Cancel input
				m.uiState.SetMode(state.NormalMode)
				m.inputBuffer = ""
				m.inputPrompt = ""
				m.inputState.Clear()
				return m, nil

			case "backspace", "ctrl+h":
				// Remove last character
				if len(m.inputBuffer) > 0 {
					m.inputBuffer = m.inputBuffer[:len(m.inputBuffer)-1]
					m.inputState.Backspace() // Keep InputState synced
				}
				return m, nil

			default:
				// Append character to input buffer
				// Only accept printable characters and limit length
				key := msg.String()
				if len(key) == 1 && len(m.inputBuffer) < 100 {
					m.inputBuffer += key
					if len(key) == 1 {
						m.inputState.AppendChar(rune(key[0])) // Keep InputState synced
					}
				}
				return m, nil
			}

		} else if m.uiState.Mode() == state.DeleteConfirmMode {
			// Delete confirmation mode (for tasks)
			switch msg.String() {
			case "y", "Y":
				// Confirm deletion
				task := m.getCurrentTask()
				if task != nil {
					err := database.DeleteTask(m.db, task.ID)
					if err != nil {
						log.Printf("Error deleting task: %v", err)
					} else {
						m.removeCurrentTask()
					}
				}
				m.uiState.SetMode(state.NormalMode)
				return m, nil

			case "n", "N", "esc":
				// Cancel deletion
				m.uiState.SetMode(state.NormalMode)
				return m, nil
			}

		} else if m.uiState.Mode() == state.DeleteColumnConfirmMode {
			// Delete confirmation mode (for columns)
			switch msg.String() {
			case "y", "Y":
				// Confirm deletion
				column := m.getCurrentColumn()
				if column != nil {
					err := database.DeleteColumn(m.db, column.ID)
					if err != nil {
						log.Printf("Error deleting column: %v", err)
					} else {
						// Delete tasks from local state
						delete(m.appState.Tasks(), column.ID)
						// Remove column from local state
						m.removeCurrentColumn()
					}
				}
				m.uiState.SetMode(state.NormalMode)
				return m, nil

			case "n", "N", "esc":
				// Cancel deletion
				m.uiState.SetMode(state.NormalMode)
				return m, nil
			}

		} else if m.uiState.Mode() == state.HelpMode {
			// Help screen mode - any key returns to normal mode
			switch msg.String() {
			case "?", "q", "esc", "enter", " ":
				m.uiState.SetMode(state.NormalMode)
				return m, nil
			}

		} else if m.uiState.Mode() == state.ViewTaskMode {
			// View task mode - close popup on esc or space, 'l' opens label picker
			switch msg.String() {
			case "esc", " ", "q":
				m.uiState.SetMode(state.NormalMode)
				m.uiState.SetViewingTask(nil)
				return m, nil
			case "l":
				// Open label picker for this task
				if m.uiState.ViewingTask() != nil {
					if m.initLabelPicker(m.uiState.ViewingTask().ID) {
						m.uiState.SetMode(state.LabelPickerMode)
					}
				}
				return m, nil
			}

		} else if m.uiState.Mode() == state.LabelPickerMode {
			// Label picker mode - navigate, toggle, create
			return m.updateLabelPicker(msg)
		}
		// Note: TicketFormMode is handled at the top of Update()

	case tea.WindowSizeMsg:
		// Handle terminal resize events
		m.uiState.SetWidth(msg.Width)
		m.uiState.SetHeight(msg.Height)

		// Recalculate how many columns fit in the new width (UIState handles this internally)
		m.calculateViewportSize()

		// Ensure viewport offset is still valid after resize
		if m.uiState.ViewportOffset()+m.uiState.ViewportSize() > len(m.appState.Columns()) {
			m.uiState.SetViewportOffset(max(0, len(m.appState.Columns())-m.uiState.ViewportSize()))
		}
	}

	// Return updated model with no command
	return m, nil
}

// updateTicketForm handles all messages when in TicketFormMode
// This is separated out because huh forms need to receive ALL messages, not just KeyMsg
func (m Model) updateTicketForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.ticketForm == nil {
		m.uiState.SetMode(state.NormalMode)
		return m, nil
	}

	// Check for escape key to cancel
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if keyMsg.String() == "esc" {
			m.uiState.SetMode(state.NormalMode)
			m.ticketForm = nil
			return m, nil
		}
	}

	// Forward ALL messages to the form
	var cmds []tea.Cmd
	form, cmd := m.ticketForm.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.ticketForm = f
		cmds = append(cmds, cmd)
	}

	// Check if form is complete
	if m.ticketForm.State == huh.StateCompleted {
		// Read values directly from the form (not from bound pointers which point to stale model copies)
		title := m.ticketForm.GetString("title")
		description := m.ticketForm.GetString("description")

		confirm := true
		if c := m.ticketForm.Get("confirm"); c != nil {
			if b, ok := c.(bool); ok {
				confirm = b
			}
		}

		// Get label IDs - need type assertion since it's a generic Get
		var labelIDs []int
		if labels := m.ticketForm.Get("labels"); labels != nil {
			if ids, ok := labels.([]int); ok {
				labelIDs = ids
			}
		}

		// Form submitted - check confirmation and save the task
		if !confirm {
			// User selected "No" on confirmation
			m.uiState.SetMode(state.NormalMode)
			m.ticketForm = nil
			m.editingTaskID = 0
			m.formState.ClearTicketForm()
			return m, tea.ClearScreen
		}
		if strings.TrimSpace(title) != "" {
			if m.editingTaskID == 0 {
				// Create new task
				currentCol := m.appState.Columns()[m.uiState.SelectedColumn()]
				task, err := database.CreateTask(
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
						err = database.SetTaskLabels(m.db, task.ID, labelIDs)
						if err != nil {
							log.Printf("Error setting labels: %v", err)
						}
					}
					// Reload task summary with labels
					summaries, err := database.GetTaskSummariesByColumn(m.db, currentCol.ID)
					if err != nil {
						log.Printf("Error reloading tasks: %v", err)
					} else {
						m.appState.Tasks()[currentCol.ID] = summaries
					}
				}
			} else {
				// Update existing task
				err := database.UpdateTask(m.db, m.editingTaskID, strings.TrimSpace(title), strings.TrimSpace(description))
				if err != nil {
					log.Printf("Error updating task: %v", err)
					m.errorState.Set("Error updating task")
				} else {
					// Update labels
					err = database.SetTaskLabels(m.db, m.editingTaskID, labelIDs)
					if err != nil {
						log.Printf("Error setting labels: %v", err)
					}
					// Reload task summaries for the column
					currentCol := m.appState.Columns()[m.uiState.SelectedColumn()]
					summaries, err := database.GetTaskSummariesByColumn(m.db, currentCol.ID)
					if err != nil {
						log.Printf("Error reloading tasks: %v", err)
					} else {
						m.appState.Tasks()[currentCol.ID] = summaries
					}
				}
			}
		}
		m.uiState.SetMode(state.NormalMode)
		m.ticketForm = nil
		m.editingTaskID = 0
		m.formState.ClearTicketForm()
		return m, tea.ClearScreen
	}

	// Check if form was aborted
	if m.ticketForm.State == huh.StateAborted {
		m.uiState.SetMode(state.NormalMode)
		m.ticketForm = nil
		m.editingTaskID = 0
		m.formState.ClearTicketForm()
		return m, tea.ClearScreen
	}

	return m, tea.Batch(cmds...)
}

// updateProjectForm handles all messages when in ProjectFormMode
// This is separated out because huh forms need to receive ALL messages, not just KeyMsg
func (m Model) updateProjectForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.projectForm == nil {
		m.uiState.SetMode(state.NormalMode)
		return m, nil
	}

	// Check for escape key to cancel
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if keyMsg.String() == "esc" {
			m.uiState.SetMode(state.NormalMode)
			m.projectForm = nil
			return m, nil
		}
	}

	// Forward ALL messages to the form
	var cmds []tea.Cmd
	form, cmd := m.projectForm.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.projectForm = f
		cmds = append(cmds, cmd)
	}

	// Check if form is complete
	if m.projectForm.State == huh.StateCompleted {
		// Read values directly from the form using GetString
		name := m.projectForm.GetString("name")
		description := m.projectForm.GetString("description")

		// Form submitted - create the project
		if strings.TrimSpace(name) != "" {
			project, err := database.CreateProject(m.db, strings.TrimSpace(name), strings.TrimSpace(description))
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
		m.projectForm = nil
		m.formState.ClearProjectForm()
		return m, tea.ClearScreen
	}

	// Check if form was aborted
	if m.projectForm.State == huh.StateAborted {
		m.uiState.SetMode(state.NormalMode)
		m.projectForm = nil
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
	if m.labelPickerCreateMode {
		return m.updateLabelColorPicker(keyMsg)
	}

	// Get filtered items to determine bounds
	filteredItems := m.getFilteredLabelPickerItems()
	maxIdx := len(filteredItems) // +1 for "create new label" option

	switch keyMsg.String() {
	case "esc":
		// Close picker and return to ViewTaskMode
		m.uiState.SetMode(state.ViewTaskMode)
		m.labelPickerFilter = ""
		m.labelPickerCursor = 0
		return m, nil

	case "up", "k":
		// Move cursor up
		if m.labelPickerCursor > 0 {
			m.labelPickerCursor--
		}
		return m, nil

	case "down", "j":
		// Move cursor down
		if m.labelPickerCursor < maxIdx {
			m.labelPickerCursor++
		}
		return m, nil

	case "enter":
		// Toggle label or create new
		if m.labelPickerCursor < len(filteredItems) {
			// Toggle this label
			item := filteredItems[m.labelPickerCursor]

			// Find the index in the unfiltered list
			for i, pi := range m.labelPickerItems {
				if pi.Label.ID == item.Label.ID {
					if m.labelPickerItems[i].Selected {
						// Remove label from task
						err := database.RemoveLabelFromTask(m.db, m.labelPickerTaskID, item.Label.ID)
						if err != nil {
							log.Printf("Error removing label: %v", err)
						} else {
							m.labelPickerItems[i].Selected = false
						}
					} else {
						// Add label to task
						err := database.AddLabelToTask(m.db, m.labelPickerTaskID, item.Label.ID)
						if err != nil {
							log.Printf("Error adding label: %v", err)
						} else {
							m.labelPickerItems[i].Selected = true
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
			if strings.TrimSpace(m.labelPickerFilter) != "" {
				m.formLabelName = strings.TrimSpace(m.labelPickerFilter)
			} else {
				m.formLabelName = "New Label"
			}
			m.labelPickerCreateMode = true
			m.labelPickerColorIdx = 0
		}
		return m, nil

	case "backspace", "ctrl+h":
		// Remove last character from filter
		if len(m.labelPickerFilter) > 0 {
			m.labelPickerFilter = m.labelPickerFilter[:len(m.labelPickerFilter)-1]
			// Reset cursor if it's out of bounds after filter change
			newFiltered := m.getFilteredLabelPickerItems()
			if m.labelPickerCursor > len(newFiltered) {
				m.labelPickerCursor = len(newFiltered)
			}
		}
		return m, nil

	default:
		// Type to filter/search
		key := keyMsg.String()
		if len(key) == 1 && len(m.labelPickerFilter) < 50 {
			m.labelPickerFilter += key
			// Reset cursor to 0 when filter changes
			m.labelPickerCursor = 0
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
		m.labelPickerCreateMode = false
		return m, nil

	case "up", "k":
		if m.labelPickerColorIdx > 0 {
			m.labelPickerColorIdx--
		}
		return m, nil

	case "down", "j":
		if m.labelPickerColorIdx < maxIdx {
			m.labelPickerColorIdx++
		}
		return m, nil

	case "enter":
		// Create the new label
		color := colors[m.labelPickerColorIdx].Color
		project := m.getCurrentProject()
		if project == nil {
			m.labelPickerCreateMode = false
			return m, nil
		}

		label, err := database.CreateLabel(m.db, project.ID, m.formLabelName, color)
		if err != nil {
			log.Printf("Error creating label: %v", err)
			m.labelPickerCreateMode = false
			return m, nil
		}

		// Add to labels list
		m.labels = append(m.labels, label)

		// Add to picker items (selected by default)
		m.labelPickerItems = append(m.labelPickerItems, state.LabelPickerItem{
			Label:    label,
			Selected: true,
		})

		// Assign to current task
		err = database.AddLabelToTask(m.db, m.labelPickerTaskID, label.ID)
		if err != nil {
			log.Printf("Error assigning new label to task: %v", err)
		}

		// Reload task detail to update the view
		m.reloadViewingTask()
		// Reload task summaries for the current column
		m.reloadCurrentColumnTasks()

		// Exit create mode and clear filter
		m.labelPickerCreateMode = false
		m.labelPickerFilter = ""
		m.labelPickerCursor = 0

		return m, nil
	}

	return m, nil
}

// reloadViewingTask reloads the task detail being viewed
func (m *Model) reloadViewingTask() {
	if m.viewingTask == nil {
		return
	}

	taskDetail, err := database.GetTaskDetail(m.db, m.viewingTask.ID)
	if err != nil {
		log.Printf("Error reloading task detail: %v", err)
		return
	}
	m.viewingTask = taskDetail
}

// reloadCurrentColumnTasks reloads task summaries for the current column
func (m *Model) reloadCurrentColumnTasks() {
	if len(m.appState.Columns()) == 0 || m.uiState.SelectedColumn() >= len(m.appState.Columns()) {
		return
	}

	currentCol := m.appState.Columns()[m.uiState.SelectedColumn()]
	summaries, err := database.GetTaskSummariesByColumn(m.db, currentCol.ID)
	if err != nil {
		log.Printf("Error reloading column tasks: %v", err)
		return
	}
	m.appState.Tasks()[currentCol.ID] = summaries
}
