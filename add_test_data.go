//go:build ignore
// +build ignore

// Helper script to add test tasks to the database
// Run with: go run add_test_data.go

package main

import (
	"context"
	"log/slog"

	"github.com/thenoetrevino/paso/internal/database"
)

func main() {
	// Initialize database
	db, err := database.InitDB(context.Background())
	if err != nil {
		slog.Error("failed to initialize database", "error", err)
		return
	}
	defer db.Close()

	// Get all columns
	columns, err := database.GetAllColumns(db)
	if err != nil {
		slog.Error("failed to get columns", "error", err)
		return
	}

	if len(columns) < 3 {
		slog.Error("Expected at least 3 columns (Todo, In Progress, Done)")
		return
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
			slog.Error("failed to create task", "title", title, "error", err)
		} else {
			slog.Info("Created task", "title", title)
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
			slog.Error("failed to create task", "title", title, "error", err)
		} else {
			slog.Info("Created task", "title", title)
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
			slog.Error("failed to create task", "title", title, "error", err)
		} else {
			slog.Info("Created task", "title", title)
		}
	}

	slog.Info("Test data added successfully")
}
