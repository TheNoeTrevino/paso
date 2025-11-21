package tui

import (
	"log"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/thenoetrevino/paso/internal/database"
	"github.com/thenoetrevino/paso/internal/models"
)

// Update handles all messages and updates the model accordingly
// This implements the "Update" part of the Model-View-Update pattern
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle TicketFormMode first - form needs ALL messages, not just KeyMsg
	if m.mode == TicketFormMode {
		return m.updateTicketForm(msg)
	}

	switch msg := msg.(type) {

	case tea.KeyMsg:
		// Handle keyboard input based on current mode
		if m.mode == NormalMode {
			// Normal mode: navigation and command keys
			// Clear any previous error messages
			m.errorMessage = ""

			switch msg.String() {
			case "q", "ctrl+c":
				// Quit the application
				return m, tea.Quit

			case "?":
				// Show help screen
				m.mode = HelpMode
				return m, nil

			case "a":
				// Add new task in current column using ticket form
				if len(m.columns) == 0 {
					m.errorMessage = "Cannot add task: No columns exist. Create a column first with 'C'"
					return m, nil
				}
				// Initialize form fields
				m.formTitle = ""
				m.formDescription = ""
				m.formLabelIDs = []int{}
				m.formConfirm = true // Default to "Yes" so user can just hit Enter to save
				m.editingTaskID = 0  // 0 means new task
				// Create the form
				m.ticketForm = CreateTicketForm(
					&m.formTitle,
					&m.formDescription,
					&m.formLabelIDs,
					m.labels,
					&m.formConfirm,
				)
				m.mode = TicketFormMode
				return m, m.ticketForm.Init()

			case "e":
				// Edit selected task using ticket form
				task := m.getCurrentTask()
				if task != nil {
					// Load full task details
					taskDetail, err := database.GetTaskDetail(m.db, task.ID)
					if err != nil {
						log.Printf("Error loading task details: %v", err)
						m.errorMessage = "Error loading task details"
						return m, nil
					}
					// Initialize form fields with existing data
					m.formTitle = taskDetail.Title
					m.formDescription = taskDetail.Description
					m.formLabelIDs = make([]int, len(taskDetail.Labels))
					for i, label := range taskDetail.Labels {
						m.formLabelIDs[i] = label.ID
					}
					m.formConfirm = true // Default to "Yes" so user can just hit Enter to save
					m.editingTaskID = task.ID
					// Create the form
					m.ticketForm = CreateTicketForm(
						&m.formTitle,
						&m.formDescription,
						&m.formLabelIDs,
						m.labels,
						&m.formConfirm,
					)
					m.mode = TicketFormMode
					return m, m.ticketForm.Init()
				} else {
					m.errorMessage = "No task selected to edit"
				}
				return m, nil

			case "d":
				// Delete selected task (if one exists)
				if m.getCurrentTask() != nil {
					m.mode = DeleteConfirmMode
				} else {
					m.errorMessage = "No task selected to delete"
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
						m.errorMessage = "Error loading task details"
						return m, nil
					}
					m.viewingTask = taskDetail
					m.mode = ViewTaskMode
				} else {
					m.errorMessage = "No task selected to view"
				}
				return m, nil

			case "C":
				// Create new column
				m.mode = AddColumnMode
				m.inputPrompt = "New column name:"
				m.inputBuffer = ""
				return m, nil

			case "R":
				// Rename selected column (if one exists)
				column := m.getCurrentColumn()
				if column != nil {
					m.mode = EditColumnMode
					m.inputBuffer = column.Name
					m.inputPrompt = "Rename column:"
				} else {
					m.errorMessage = "No column selected to rename"
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
						m.errorMessage = "Error getting column info"
						return m, nil
					}
					m.deleteColumnTaskCount = taskCount
					m.mode = DeleteColumnConfirmMode
				} else {
					m.errorMessage = "No column selected to delete"
				}
				return m, nil

			// Viewport scrolling: Move the viewport window
			case "]":
				// Scroll viewport right (show columns to the right)
				if m.viewportOffset+m.viewportSize < len(m.columns) {
					m.viewportOffset++
					// Adjust selectedColumn if it's now off-screen to the left
					if m.selectedColumn < m.viewportOffset {
						m.selectedColumn = m.viewportOffset
						m.selectedTask = 0
					}
				}

			case "[":
				// Scroll viewport left (show columns to the left)
				if m.viewportOffset > 0 {
					m.viewportOffset--
					// Adjust selectedColumn if it's now off-screen to the right
					if m.selectedColumn >= m.viewportOffset+m.viewportSize {
						m.selectedColumn = m.viewportOffset + m.viewportSize - 1
						m.selectedTask = 0
					}
				}

			// Left/Right navigation: Move between columns
			case "h", "left":
				// Move to previous column if not at first column
				if m.selectedColumn > 0 {
					m.selectedColumn--
					m.selectedTask = 0 // Reset task selection when switching columns
					// Auto-scroll viewport if selected column is now off-screen to the left
					if m.selectedColumn < m.viewportOffset {
						m.viewportOffset = m.selectedColumn
					}
				}

			case "l", "right":
				// Move to next column if not at last column
				if m.selectedColumn < len(m.columns)-1 {
					m.selectedColumn++
					m.selectedTask = 0 // Reset task selection when switching columns
					// Auto-scroll viewport if selected column is now off-screen to the right
					if m.selectedColumn >= m.viewportOffset+m.viewportSize {
						m.viewportOffset = m.selectedColumn - m.viewportSize + 1
					}
				}

			// Up/Down navigation: Move between tasks in current column
			case "j", "down":
				// Move to next task in current column
				currentTasks := m.getCurrentTasks()
				if len(currentTasks) > 0 && m.selectedTask < len(currentTasks)-1 {
					m.selectedTask++
				}

			case "k", "up":
				// Move to previous task in current column
				if m.selectedTask > 0 {
					m.selectedTask--
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
			}

		} else if m.mode == AddColumnMode || m.mode == EditColumnMode {
			// Input modes: handle text input for column operations
			switch msg.String() {
			case "enter":
				// Confirm input and create/edit column
				if strings.TrimSpace(m.inputBuffer) != "" {
					if m.mode == AddColumnMode {
						// Create new column after the current column
						var afterColumnID *int
						if len(m.columns) > 0 {
							currentCol := m.columns[m.selectedColumn]
							afterColumnID = &currentCol.ID
						}
						column, err := database.CreateColumn(m.db, strings.TrimSpace(m.inputBuffer), afterColumnID)
						if err != nil {
							log.Printf("Error creating column: %v", err)
						} else {
							// Reload columns from database to get correct order
							m.columns, err = database.GetAllColumns(m.db)
							if err != nil {
								log.Printf("Error reloading columns: %v", err)
							}
							m.tasks[column.ID] = []*models.TaskSummary{}
							// Move selection to new column (it will be after current)
							if afterColumnID != nil {
								m.selectedColumn++
							}
						}
					} else if m.mode == EditColumnMode {
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
				m.mode = NormalMode
				m.inputBuffer = ""
				m.inputPrompt = ""
				return m, nil

			case "esc":
				// Cancel input
				m.mode = NormalMode
				m.inputBuffer = ""
				m.inputPrompt = ""
				return m, nil

			case "backspace", "ctrl+h":
				// Remove last character
				if len(m.inputBuffer) > 0 {
					m.inputBuffer = m.inputBuffer[:len(m.inputBuffer)-1]
				}
				return m, nil

			default:
				// Append character to input buffer
				// Only accept printable characters and limit length
				key := msg.String()
				if len(key) == 1 && len(m.inputBuffer) < 100 {
					m.inputBuffer += key
				}
				return m, nil
			}

		} else if m.mode == DeleteConfirmMode {
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
				m.mode = NormalMode
				return m, nil

			case "n", "N", "esc":
				// Cancel deletion
				m.mode = NormalMode
				return m, nil
			}

		} else if m.mode == DeleteColumnConfirmMode {
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
						delete(m.tasks, column.ID)
						// Remove column from local state
						m.removeCurrentColumn()
					}
				}
				m.mode = NormalMode
				return m, nil

			case "n", "N", "esc":
				// Cancel deletion
				m.mode = NormalMode
				return m, nil
			}

		} else if m.mode == HelpMode {
			// Help screen mode - any key returns to normal mode
			switch msg.String() {
			case "?", "q", "esc", "enter", " ":
				m.mode = NormalMode
				return m, nil
			}

		} else if m.mode == ViewTaskMode {
			// View task mode - close popup on esc or space
			switch msg.String() {
			case "esc", " ", "q":
				m.mode = NormalMode
				m.viewingTask = nil
				return m, nil
			}
		}
		// Note: TicketFormMode is handled at the top of Update()

	case tea.WindowSizeMsg:
		// Handle terminal resize events
		m.width = msg.Width
		m.height = msg.Height

		// Recalculate how many columns fit in the new width
		m.calculateViewportSize()

		// Ensure viewport offset is still valid after resize
		if m.viewportOffset+m.viewportSize > len(m.columns) {
			m.viewportOffset = max(0, len(m.columns)-m.viewportSize)
		}
	}

	// Return updated model with no command
	return m, nil
}

// updateTicketForm handles all messages when in TicketFormMode
// This is separated out because huh forms need to receive ALL messages, not just KeyMsg
func (m Model) updateTicketForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.ticketForm == nil {
		m.mode = NormalMode
		return m, nil
	}

	// Check for escape key to cancel
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if keyMsg.String() == "esc" {
			m.mode = NormalMode
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
			m.mode = NormalMode
			m.ticketForm = nil
			m.editingTaskID = 0
			return m, tea.ClearScreen
		}
		if strings.TrimSpace(title) != "" {
			if m.editingTaskID == 0 {
				// Create new task
				currentCol := m.columns[m.selectedColumn]
				task, err := database.CreateTask(
					m.db,
					strings.TrimSpace(title),
					strings.TrimSpace(description),
					currentCol.ID,
					len(m.tasks[currentCol.ID]),
				)
				if err != nil {
					log.Printf("Error creating task: %v", err)
					m.errorMessage = "Error creating task"
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
						m.tasks[currentCol.ID] = summaries
					}
				}
			} else {
				// Update existing task
				err := database.UpdateTask(m.db, m.editingTaskID, strings.TrimSpace(title), strings.TrimSpace(description))
				if err != nil {
					log.Printf("Error updating task: %v", err)
					m.errorMessage = "Error updating task"
				} else {
					// Update labels
					err = database.SetTaskLabels(m.db, m.editingTaskID, labelIDs)
					if err != nil {
						log.Printf("Error setting labels: %v", err)
					}
					// Reload task summaries for the column
					currentCol := m.columns[m.selectedColumn]
					summaries, err := database.GetTaskSummariesByColumn(m.db, currentCol.ID)
					if err != nil {
						log.Printf("Error reloading tasks: %v", err)
					} else {
						m.tasks[currentCol.ID] = summaries
					}
				}
			}
		}
		m.mode = NormalMode
		m.ticketForm = nil
		m.editingTaskID = 0
		return m, tea.ClearScreen
	}

	// Check if form was aborted
	if m.ticketForm.State == huh.StateAborted {
		m.mode = NormalMode
		m.ticketForm = nil
		m.editingTaskID = 0
		return m, tea.ClearScreen
	}

	return m, tea.Batch(cmds...)
}
