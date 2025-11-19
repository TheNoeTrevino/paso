package main

import (
	"fmt"
	"log"

	"github.com/thenoetrevino/paso/internal/database"
)

func main() {
	fmt.Println("=== Phase 1: Database Foundation Test ===\n")

	// Run CRUD tests
	testCRUD()

	// Display database info
	fmt.Println("\n=== Database Information ===")
	db, err := database.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	columns, _ := database.GetAllColumns(db)
	fmt.Printf("Database location: ~/.paso/tasks.db\n")
	fmt.Printf("Default columns created: %d\n", len(columns))
	for _, col := range columns {
		tasks, _ := database.GetTasksByColumn(db, col.ID)
		fmt.Printf("  - %s (%d tasks)\n", col.Name, len(tasks))
	}

	fmt.Println("\n✅ Phase 1 completed successfully!")
}
