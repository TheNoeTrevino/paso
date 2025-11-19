package tui

import (
	"database/sql"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/thenoetrevino/paso/internal/database"
	"github.com/thenoetrevino/paso/internal/models"
)

// Model represents the application state for the TUI
type Model struct {
	db      *sql.DB                // Database connection
	columns []*models.Column       // All columns ordered by position
	tasks   map[int][]*models.Task // Tasks organized by column ID
	width   int                    // Terminal width
	height  int                    // Terminal height
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
		db:      db,
		columns: columns,
		tasks:   tasks,
		width:   0,
		height:  0,
	}
}

// Init initializes the Bubble Tea application
// Required by tea.Model interface
func (m Model) Init() tea.Cmd {
	// No initial commands needed yet
	return nil
}
