# Paso - Terminal Kanban Board

**Paso** (Spanish for "step") is a zero-setup, terminal-based kanban board for personal task management.

## Current Status: Phase 1 Complete ✅

Phase 1 implements the database foundation with SQLite persistence and CRUD operations.

## Features Implemented

- **SQLite Database**: Pure Go implementation using `modernc.org/sqlite` (no CGo dependencies)
- **Auto-initialization**: Database and directory (`~/.paso/`) created automatically on first run
- **Default Columns**: Three columns pre-seeded: "Todo", "In Progress", "Done"
- **Full CRUD Operations**: Create, read, update, and delete tasks and columns
- **Proper Schema**: Foreign key constraints, indexes, and timestamps

## Project Structure

```
paso/
├── main.go                    # Entry point with test code
├── test_crud.go              # CRUD operation tests
├── internal/
│   ├── database/
│   │   ├── db.go             # Database initialization
│   │   ├── migrations.go     # Schema definitions and seeding
│   │   └── repository.go     # CRUD operations
│   └── models/
│       ├── column.go         # Column struct
│       └── task.go           # Task struct
└── README.md
```

## Building and Running

```bash
# Build the project
go build

# Run the test program
./paso

# The database will be created at:
~/.paso/tasks.db
```

## Database Schema

### Columns Table
- `id` (INTEGER PRIMARY KEY AUTOINCREMENT)
- `name` (TEXT NOT NULL)
- `position` (INTEGER NOT NULL)

### Tasks Table
- `id` (INTEGER PRIMARY KEY AUTOINCREMENT)
- `title` (TEXT NOT NULL)
- `description` (TEXT)
- `column_id` (INTEGER NOT NULL, FK to columns.id)
- `position` (INTEGER NOT NULL)
- `created_at` (DATETIME DEFAULT CURRENT_TIMESTAMP)
- `updated_at` (DATETIME DEFAULT CURRENT_TIMESTAMP)

## API Reference

### Database Functions

```go
// Initialize database connection
db, err := database.InitDB()

// Column operations
column, err := database.CreateColumn(db, "Backlog", 0)
columns, err := database.GetAllColumns(db)

// Task operations
task, err := database.CreateTask(db, "Fix bug", "Description", columnID, position)
tasks, err := database.GetTasksByColumn(db, columnID)
err := database.UpdateTaskColumn(db, taskID, newColumnID, newPosition)
err := database.DeleteTask(db, taskID)
```

## Next Steps: Phase 2

Phase 2 will implement the Bubble Tea TUI framework to create the interactive terminal interface.

## Tech Stack

- **Go** - Primary language
- **modernc.org/sqlite** - Pure Go SQLite (no CGo)
- **Bubble Tea** - TUI framework (Phase 2)
- **Lipgloss** - Styling (Phase 2)
- **Harmonica** - Animations (Phase 8)

## License

MIT
