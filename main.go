package main

import (
	"fmt"
	"log"

	tea "charm.land/bubbletea/v2"
	"github.com/thenoetrevino/paso/internal/config"
	"github.com/thenoetrevino/paso/internal/database"
	"github.com/thenoetrevino/paso/internal/tui"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database
	db, err := database.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Create repository wrapping the database
	repo := database.NewRepository(db)

	// Create initial TUI model with repository and config
	model := tui.InitialModel(repo, cfg)

	// Create and run Bubble Tea program
	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		log.Fatal(err)
	}
}
