package tui

import (
	"log"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/thenoetrevino/paso/internal/database"
	"github.com/thenoetrevino/paso/internal/models"
)

// Update handles all messages and updates the model accordingly
// This implements the "Update" part of the Model-View-Update pattern
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				// Add new task in current column
				if len(m.columns) == 0 {
					m.errorMessage = "Cannot add task: No columns exist. Create a column first with 'C'"
					return m, nil
				}
				m.mode = AddTaskMode
				m.inputPrompt = "New task title:"
				m.inputBuffer = ""
				return m, nil

			case "e":
				// Edit selected task (if one exists)
				task := m.getCurrentTask()
				if task != nil {
					m.mode = EditTaskMode
					m.inputBuffer = task.Title
					m.inputPrompt = "Edit task title:"
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

		} else if m.mode == AddTaskMode || m.mode == EditTaskMode || m.mode == AddColumnMode || m.mode == EditColumnMode {
			// Input modes: handle text input
			switch msg.String() {
			case "enter":
				// Confirm input and create/edit task or column
				if strings.TrimSpace(m.inputBuffer) != "" {
					if m.mode == AddTaskMode {
						// Create new task
						currentCol := m.columns[m.selectedColumn]
						task, err := database.CreateTask(
							m.db,
							strings.TrimSpace(m.inputBuffer),
							"",
							currentCol.ID,
							len(m.tasks[currentCol.ID]),
						)
						if err != nil {
							log.Printf("Error creating task: %v", err)
						} else {
							m.tasks[currentCol.ID] = append(m.tasks[currentCol.ID], task)
						}
					} else if m.mode == EditTaskMode {
						// Update existing task
						task := m.getCurrentTask()
						if task != nil {
							err := database.UpdateTaskTitle(m.db, task.ID, strings.TrimSpace(m.inputBuffer))
							if err != nil {
								log.Printf("Error updating task: %v", err)
							} else {
								task.Title = strings.TrimSpace(m.inputBuffer)
							}
						}
					} else if m.mode == AddColumnMode {
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
							m.tasks[column.ID] = []*models.Task{}
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
		}

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
