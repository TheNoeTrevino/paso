package tui

import (
	"database/sql"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/thenoetrevino/paso/internal/database"
	"github.com/thenoetrevino/paso/internal/models"
)

// Mode represents the current interaction mode of the TUI
type Mode int

const (
	NormalMode              Mode = iota // Default navigation mode
	AddTaskMode                         // Creating a new task
	EditTaskMode                        // Editing an existing task
	DeleteConfirmMode                   // Confirming task deletion
	AddColumnMode                       // Creating a new column
	EditColumnMode                      // Renaming an existing column
	DeleteColumnConfirmMode             // Confirming column deletion
)

// Model represents the application state for the TUI
type Model struct {
	db                    *sql.DB                // Database connection
	columns               []*models.Column       // All columns ordered by position
	tasks                 map[int][]*models.Task // Tasks organized by column ID
	selectedColumn        int                    // Index of currently selected column
	selectedTask          int                    // Index of currently selected task in the column
	width                 int                    // Terminal width
	height                int                    // Terminal height
	mode                  Mode                   // Current interaction mode
	inputBuffer           string                 // Text being typed in input mode
	inputPrompt           string                 // Prompt to show in input dialog
	viewportOffset        int                    // Index of leftmost visible column
	viewportSize          int                    // Number of columns that fit on screen
	deleteColumnTaskCount int                    // Number of tasks in column being deleted
}

// InitialModel creates and initializes the TUI model with data from the database
func InitialModel(db *sql.DB) Model {
	// Load all columns from database
	columns, err := database.GetAllColumns(db)
	if err != nil {
		log.Printf("Error loading columns: %v", err)
		columns = []*models.Column{}
	}

	// Load tasks for each column
	tasks := make(map[int][]*models.Task)
	for _, col := range columns {
		columnTasks, err := database.GetTasksByColumn(db, col.ID)
		if err != nil {
			log.Printf("Error loading tasks for column %d: %v", col.ID, err)
			columnTasks = []*models.Task{}
		}
		tasks[col.ID] = columnTasks
	}

	return Model{
		db:             db,
		columns:        columns,
		tasks:          tasks,
		selectedColumn: 0,
		selectedTask:   0,
		width:          0,
		height:         0,
		mode:           NormalMode,
		inputBuffer:    "",
		inputPrompt:    "",
		viewportOffset: 0,
		viewportSize:   1, // Default to 1, will be recalculated when width is set
	}
}

// Init initializes the Bubble Tea application
// Required by tea.Model interface
func (m Model) Init() tea.Cmd {
	// No initial commands needed yet
	return nil
}

// getCurrentTasks returns the tasks for the currently selected column
// Returns an empty slice if the column has no tasks
func (m Model) getCurrentTasks() []*models.Task {
	if len(m.columns) == 0 {
		return []*models.Task{}
	}
	currentCol := m.columns[m.selectedColumn]
	tasks := m.tasks[currentCol.ID]
	if tasks == nil {
		return []*models.Task{}
	}
	return tasks
}

// getCurrentTask returns the currently selected task
// Returns nil if there are no tasks in the current column or no columns exist
func (m Model) getCurrentTask() *models.Task {
	tasks := m.getCurrentTasks()
	if len(tasks) == 0 {
		return nil
	}
	if m.selectedTask >= len(tasks) {
		return nil
	}
	return tasks[m.selectedTask]
}

// removeCurrentTask removes the currently selected task from the model's local state
// This should be called after successfully deleting a task from the database
// It adjusts the selectedTask index if necessary to keep it within bounds
func (m *Model) removeCurrentTask() {
	if len(m.columns) == 0 {
		return
	}

	currentCol := m.columns[m.selectedColumn]
	tasks := m.tasks[currentCol.ID]

	if len(tasks) == 0 || m.selectedTask >= len(tasks) {
		return
	}

	// Remove the task at selectedTask index
	m.tasks[currentCol.ID] = append(tasks[:m.selectedTask], tasks[m.selectedTask+1:]...)

	// Adjust selectedTask if we removed the last task
	if m.selectedTask >= len(m.tasks[currentCol.ID]) && m.selectedTask > 0 {
		m.selectedTask--
	}
}

// calculateViewportSize calculates how many columns can fit in the terminal width
// Column width: 30 (content) + 2 (padding) + 2 (border) = 34 chars
// Spacing between columns: 2 chars
// Total per column: 36 chars
// This method ensures at least 1 column is always visible
func (m *Model) calculateViewportSize() {
	if m.width == 0 {
		m.viewportSize = 1
		return
	}

	const columnWidth = 36 // 30 content + 2 padding + 2 border + 2 spacing
	// Reserve 4 chars for margins and scroll indicators
	availableWidth := m.width - 4

	// Calculate how many columns fit, with minimum of 1
	m.viewportSize = max(1, availableWidth/columnWidth)
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// getCurrentColumn returns the currently selected column
// Returns nil if there are no columns
func (m Model) getCurrentColumn() *models.Column {
	if len(m.columns) == 0 {
		return nil
	}
	if m.selectedColumn >= len(m.columns) {
		return nil
	}
	return m.columns[m.selectedColumn]
}

// removeCurrentColumn removes the currently selected column from the model's local state
// This should be called after successfully deleting a column from the database
// It adjusts the selectedColumn index if necessary to keep it within bounds
// It also adjusts the viewportOffset if needed
func (m *Model) removeCurrentColumn() {
	if len(m.columns) == 0 || m.selectedColumn >= len(m.columns) {
		return
	}

	// Remove the column at selectedColumn index
	m.columns = append(m.columns[:m.selectedColumn], m.columns[m.selectedColumn+1:]...)

	// Adjust selectedColumn if we removed the last column
	if m.selectedColumn >= len(m.columns) && m.selectedColumn > 0 {
		m.selectedColumn--
	}

	// Reset task selection
	m.selectedTask = 0

	// Adjust viewportOffset if needed to keep selection visible
	if len(m.columns) > 0 {
		// If selected column is before viewport, move viewport left
		if m.selectedColumn < m.viewportOffset {
			m.viewportOffset = m.selectedColumn
		}
		// If viewport offset is now beyond available columns, adjust it
		if m.viewportOffset+m.viewportSize > len(m.columns) {
			m.viewportOffset = max(0, len(m.columns)-m.viewportSize)
		}
	} else {
		m.viewportOffset = 0
	}
}

// moveTaskRight moves the currently selected task to the next column (right)
// Updates both the local state and the database using the linked list structure
// The selection follows the moved task and the viewport scrolls if needed
func (m *Model) moveTaskRight() {
	// Get the current task
	task := m.getCurrentTask()
	if task == nil {
		return
	}

	// Check if there's a next column using the linked list
	currentCol := m.columns[m.selectedColumn]
	if currentCol.NextID == nil {
		// Already at last column
		return
	}

	// Use the new database function to move task
	err := database.MoveTaskToNextColumn(m.db, task.ID)
	if err != nil {
		log.Printf("Error moving task to next column: %v", err)
		return
	}

	// Update local state: remove from current column
	tasks := m.tasks[currentCol.ID]
	m.tasks[currentCol.ID] = append(tasks[:m.selectedTask], tasks[m.selectedTask+1:]...)

	// Find the next column and add task there
	nextColID := *currentCol.NextID
	newPosition := len(m.tasks[nextColID])
	task.ColumnID = nextColID
	task.Position = newPosition
	m.tasks[nextColID] = append(m.tasks[nextColID], task)

	// Move selection to follow the task
	m.selectedColumn++
	m.selectedTask = newPosition

	// Ensure the moved task is visible (auto-scroll viewport if needed)
	if m.selectedColumn >= m.viewportOffset+m.viewportSize {
		m.viewportOffset++
	}
}

// moveTaskLeft moves the currently selected task to the previous column (left)
// Updates both the local state and the database using the linked list structure
// The selection follows the moved task and the viewport scrolls if needed
func (m *Model) moveTaskLeft() {
	// Get the current task
	task := m.getCurrentTask()
	if task == nil {
		return
	}

	// Check if there's a previous column using the linked list
	currentCol := m.columns[m.selectedColumn]
	if currentCol.PrevID == nil {
		// Already at first column
		return
	}

	// Use the new database function to move task
	err := database.MoveTaskToPrevColumn(m.db, task.ID)
	if err != nil {
		log.Printf("Error moving task to previous column: %v", err)
		return
	}

	// Update local state: remove from current column
	tasks := m.tasks[currentCol.ID]
	m.tasks[currentCol.ID] = append(tasks[:m.selectedTask], tasks[m.selectedTask+1:]...)

	// Find the previous column and add task there
	prevColID := *currentCol.PrevID
	newPosition := len(m.tasks[prevColID])
	task.ColumnID = prevColID
	task.Position = newPosition
	m.tasks[prevColID] = append(m.tasks[prevColID], task)

	// Move selection to follow the task
	m.selectedColumn--
	m.selectedTask = newPosition

	// Ensure the moved task is visible (auto-scroll viewport if needed)
	if m.selectedColumn < m.viewportOffset {
		m.viewportOffset--
	}
}
