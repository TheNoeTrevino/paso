package main

import (
	"fmt"
	"log"

	"github.com/thenoetrevino/paso/internal/database"
)

func main() {
	// Initialize database
	db, err := database.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Test: Retrieve all columns
	columns, err := database.GetAllColumns(db)
	if err != nil {
		log.Fatalf("Failed to get columns: %v", err)
	}

	fmt.Printf("Found %d columns:\n", len(columns))
	for _, col := range columns {
		fmt.Printf("  - %s (ID: %d, Position: %d)\n", col.Name, col.ID, col.Position)
	}

	// Test: Create a task in "Todo" column
	if len(columns) > 0 {
		task, err := database.CreateTask(db, "Test task", "This is a test task", columns[0].ID, 0)
		if err != nil {
			log.Fatalf("Failed to create task: %v", err)
		}
		fmt.Printf("\nCreated task: %s (ID: %d)\n", task.Title, task.ID)

		// Test: Retrieve all tasks in Todo column
		tasks, err := database.GetTasksByColumn(db, columns[0].ID)
		if err != nil {
			log.Fatalf("Failed to get tasks: %v", err)
		}
		fmt.Printf("Found %d tasks in %s column\n", len(tasks), columns[0].Name)
		for _, t := range tasks {
			fmt.Printf("  - %s (ID: %d)\n", t.Title, t.ID)
		}
	}

	fmt.Println("\nPhase 1 completed successfully!")
}
