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
- **Customizable Key Mappings**: Configure keyboard shortcuts via YAML config file

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

# Run the application
./paso

# The database will be created at:
~/.paso/tasks.db

# Configuration file (optional):
~/.config/paso/config.yaml
```

## Configuration

Paso supports customizable key mappings via a YAML configuration file. See `config.example.yaml` for an example.

### Creating Your Config

```bash
# Copy the example config
mkdir -p ~/.config/paso
cp config.example.yaml ~/.config/paso/config.yaml

# Edit to customize your key bindings
nano ~/.config/paso/config.yaml
```

### Default Key Bindings

#### Tasks
- `a` - Add new task
- `e` - Edit selected task
- `d` - Delete selected task
- `L` - Move task to previous column
- `H` - Move task to next column
- `K` - Move task up in column
- `J` - Move task down in column
- `space` - View task details
- `l` - Edit labels (when viewing task)

#### Columns
- `C` - Create new column
- `R` - Rename current column
- `X` - Delete current column

#### Navigation
- `h` - Move to previous column
- `l` - Move to next column
- `k` - Move to previous task
- `j` - Move to next task
- `[` - Scroll viewport left
- `]` - Scroll viewport right
- `{` - Move to next project
- `}` - Move to previous project

#### Other
- `?` - Show help screen
- `q` - Quit application

All key bindings can be customized in `~/.config/paso/config.yaml`.

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

## Running Tests

Run all tests in all packages

``` bash
  go test ./...
```

Run with verbose output (shows each test name)

``` bash
  go test -v ./...
```

Run with coverage report
``` bash
  go test -cover ./...
```

Run with detailed coverage
``` bash
  go test -coverprofile=coverage.out ./...
  go tool cover -html=coverage.out
```

Run specific package
```
  go test ./internal/database
```

Run specific test
``` go
  go test -v ./internal/database -run TestTaskCRUDPersistence
```

## License

MIT
