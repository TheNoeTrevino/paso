//go:build ignore
// +build ignore

// Helper script to add test tasks to the database
// Run with: go run add_test_data.go

package main

import (
	"context"
	"log"

	"github.com/thenoetrevino/paso/internal/database"
)

func main() {
	// Initialize database
	db, err := database.InitDB(context.Background())
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Get all columns
	columns, err := database.GetAllColumns(db)
	if err != nil {
		log.Fatalf("Failed to get columns: %v", err)
	}

	if len(columns) < 3 {
		log.Fatal("Expected at least 3 columns (Todo, In Progress, Done)")
	}

	// Add tasks to Todo column (first column)
	todoCol := columns[0]
	tasks := []string{
		"Fix auth bug",
		"Refactor UI",
		"Update deps",
	}
	for i, title := range tasks {
		_, err := database.CreateTask(db, title, "", todoCol.ID, i)
		if err != nil {
			log.Printf("Error creating task '%s': %v", title, err)
		} else {
			log.Printf("Created task: %s", title)
		}
	}

	// Add tasks to In Progress column (second column)
	inProgressCol := columns[1]
	inProgressTasks := []string{
		"Add tests",
		"Review PR #42",
	}
	for i, title := range inProgressTasks {
		_, err := database.CreateTask(db, title, "", inProgressCol.ID, i)
		if err != nil {
			log.Printf("Error creating task '%s': %v", title, err)
		} else {
			log.Printf("Created task: %s", title)
		}
	}

	// Add tasks to Done column (third column)
	doneCol := columns[2]
	doneTasks := []string{
		"Deploy v1.0",
		"Hotfix prod",
	}
	for i, title := range doneTasks {
		_, err := database.CreateTask(db, title, "", doneCol.ID, i)
		if err != nil {
			log.Printf("Error creating task '%s': %v", title, err)
		} else {
			log.Printf("Created task: %s", title)
		}
	}

	log.Println("Test data added successfully!")
}
