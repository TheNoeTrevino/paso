# Paso - Terminal Kanban Board

**Paso** (Spanish for "step") is a zero-setup, terminal-based kanban board for personal task management.

## Features

- **Dual Interface**: Both CLI commands and interactive TUI
- **Agent-Friendly**: JSON output, quiet mode, structured exit codes
- **SQLite Database**: Pure Go implementation using `modernc.org/sqlite` (no CGo dependencies)
- **Auto-initialization**: Database and directory (`~/.paso/`) created automatically on first run
- **Live Updates**: Optional daemon for real-time updates across sessions
- **Shell Completion**: Auto-completion for bash, zsh, fish, and PowerShell
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

## Installation

```bash
# Build from source
go build -o bin/paso .

# Install to system path
sudo cp bin/paso /usr/local/bin/

# Or use install script (if available)
./install.sh
```

## Quick Start

```bash
# Show help
paso

# Launch interactive TUI
paso tui

# Create a project
paso project create --title="My Project"

# Create a task
paso task create --title="Fix bug" --project=1

# List tasks
paso task list --project=1
```

## CLI Usage

### Project Management

```bash
# Create a project
paso project create --title="Backend API"

# List all projects
paso project list

# List projects (JSON output)
paso project list --json

# Delete a project
paso project delete <project-id>
```

### Task Management

```bash
# Create a simple task
paso task create --title="Fix login bug" --project=1

# Create a feature task with priority
paso task create \
  --title="User authentication" \
  --description="Implement JWT auth" \
  --type=feature \
  --priority=high \
  --project=1

# Create a subtask
PARENT_ID=$(paso task create --title="Parent task" --project=1 --quiet)
paso task create --title="Subtask" --parent=$PARENT_ID --project=1

# List tasks
paso task list --project=1

# Update task
paso task update <task-id> --title="New title" --priority=critical

# Delete task
paso task delete <task-id>
```

### Column Management

```bash
# Create a column
paso column create --name="In Review" --project=1

# List columns for a project
paso column list --project=1
```

### Label Management

```bash
# Create a label
paso label create --name="bug" --color="#FF0000" --project=1

# List labels
paso label list --project=1

# Attach label to task
paso label attach <task-id> <label-id>
```

### Agent-Friendly Features

Paso is designed to work well with AI agents and shell scripts:

```bash
# Quiet mode: Returns only the ID
PROJECT_ID=$(paso project create --title="New Project" --quiet)
TASK_ID=$(paso task create --title="Task" --project=$PROJECT_ID --quiet)

# JSON mode: Structured output for parsing
paso task list --project=1 --json | jq '.tasks[] | select(.priority.name=="high")'

# Exit codes for error handling
paso task create --title="Test" --project=999 --quiet
if [ $? -eq 3 ]; then
  echo "Project not found"
fi
```

### Shell Completion

```bash
# Generate completion script for your shell
paso completion bash > /etc/bash_completion.d/paso  # Linux
paso completion bash > $(brew --prefix)/etc/bash_completion.d/paso  # macOS
paso completion zsh > "${fpath[1]}/_paso"
paso completion fish > ~/.config/fish/completions/paso.fish
```

## TUI (Terminal User Interface)

Launch the interactive TUI with:

```bash
paso tui
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

### File Locations

```
~/.paso/tasks.db              # SQLite database
~/.paso/paso.sock             # Unix socket for daemon (optional)
~/.config/paso/config.yaml    # Configuration file
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

#### Projects
- `P` - Create new project

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

## Exit Codes

Paso uses consistent exit codes for automation:

- `0` - Success
- `1` - General error
- `2` - Usage error (invalid flags)
- `3` - Not found (project/task doesn't exist)
- `5` - Validation error (invalid input)
- `6` - Dependency error (circular dependency, etc.)

## Optional Daemon

Start the daemon for live updates across sessions:

```bash
# Start daemon
paso-daemon &

# The daemon listens on ~/.paso/paso.sock
# CLI commands automatically connect to it when available
```

## Tech Stack

- **Go** - Primary language
- **modernc.org/sqlite** - Pure Go SQLite (no CGo)
- **Bubble Tea** - TUI framework (Phase 2)
- **Lipgloss** - Styling (Phase 2)
- **Harmonica** - Animations (Phase 8)

## Running Tests

### Go concurrent runner

``` bash
go run ./cmd/ci
```

### Manually

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
