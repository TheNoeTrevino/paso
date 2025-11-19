package main

import (
	"fmt"
	"log"

	"github.com/thenoetrevino/paso/internal/database"
)

func testCRUD() {
	db, err := database.InitDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Test 1: Get columns
	columns, err := database.GetAllColumns(db)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✓ GetAllColumns: Found %d columns\n", len(columns))

	// Test 2: Create a task
	task, err := database.CreateTask(db, "Test CRUD", "Testing all operations", columns[0].ID, 0)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✓ CreateTask: Created task ID %d\n", task.ID)

	// Test 3: Get tasks by column
	tasks, err := database.GetTasksByColumn(db, columns[0].ID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✓ GetTasksByColumn: Found %d tasks\n", len(tasks))

	// Test 4: Update task column
	err = database.UpdateTaskColumn(db, task.ID, columns[1].ID, 0)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✓ UpdateTaskColumn: Moved task to column %d\n", columns[1].ID)

	// Verify the move
	tasks, _ = database.GetTasksByColumn(db, columns[1].ID)
	fmt.Printf("✓ Verification: Found %d tasks in destination column\n", len(tasks))

	// Test 5: Delete task
	err = database.DeleteTask(db, task.ID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✓ DeleteTask: Deleted task ID %d\n", task.ID)

	fmt.Println("\nAll CRUD operations passed! ✅")
}
