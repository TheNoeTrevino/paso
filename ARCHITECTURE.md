# Paso Architecture

## Overview

Paso is a terminal-based kanban board for personal task management with dual interfaces: a CLI for scriptable automation and a TUI for interactive use. Built in Go with SQLite (local) and PostgreSQL (remote) support. The architecture follows strict layering: presentation (CLI/TUI) → app container → services → database abstraction → storage.

**Core technologies**: Cobra (CLI), Bubble Tea v2 (TUI), SQLC (type-safe SQL), Lipgloss v2 (styling), Goose (migrations).

**Guiding principle**: Each layer depends only on abstractions, never concrete implementations. Services use `database.Querier`, commands use `handler.Handler`, tests inject `*app.App` via context.

---

## Project Layout

```
paso/
├── main.go                          # CLI entry point, command registration, exit handling
├── cmd/
│   ├── daemon/main.go              # Daemon process for live TUI updates
│   ├── pre-commit/main.go          # Pre-commit hook binary (gofmt)
│   └── ci/main.go                  # CI tooling binary
├── internal/
│   ├── app/                        # App container, DI via functional options
│   ├── appcontext/                 # Context keys for test injection
│   ├── cli/
│   │   ├── {domain}/              # Task, project, column, label, assignee commands
│   │   ├── handler/               # Handler pattern framework
│   │   ├── styles/                # Lipgloss styling, success/error rendering
│   │   └── golden/                # Help output golden tests
│   ├── config/
│   │   └── colors/                # 5 color scheme presets (default, monochrome, wave, dragon, lotus)
│   ├── database/
│   │   ├── adapters/              # SQLite/PostgreSQL adapters
│   │   ├── generated_{sqlite,postgres}/  # SQLC codegen output
│   │   ├── migrations_{sqlite,postgres}/ # Goose migration SQL files
│   │   ├── sql_{sqlite,postgres}/ # SQLC query definitions
│   │   └── types/                 # Database-agnostic types, Querier interface
│   ├── services/                   # Business logic layer (task, project, column, label, assignee)
│   ├── models/                     # Plain DTOs, no database dependencies
│   ├── converters/                 # DB types → model types
│   ├── tui/                        # Bubble Tea TUI (state/, components/, commands/)
│   ├── testing/
│   │   ├── cli/                   # Command execution harness
│   │   ├── fixtures/              # DB setup, data helpers
│   │   └── mocks/                 # Hand-rolled service mocks
│   ├── events/                     # Event publisher for daemon communication
│   ├── git/                        # Git detector (real + mock)
│   ├── launcher/                   # TUI launcher
│   ├── logging/                    # File logging setup
│   ├── types/                      # Custom ID types
│   └── user/                       # Username detection
└── tests/
    └── acceptance/                 # Shell-based CLI validation
```

---

## Request Lifecycle

### CLI Command Flow

```
main.go:init()
  └─> rootCmd.AddCommand(task.TaskCmd())
        └─> task.TaskCmd() registers subcommands (create, list, update, etc.)

User runs: paso task create -t "Fix bug" -p 1

main.go:main()
  └─> rootCmd.Execute()
        └─> Cobra dispatches to task create command
              └─> handler.Command(&createHandler{}, parseCreateFlags)
                    ├─> parseCreateFlags(cmd) validates flags
                    ├─> parseFlagsToMap(cmd) → Arguments
                    └─> handler.Execute(ctx, args)
                          └─> cli.GetCLIFromContext(ctx)
                                ├─> [test mode] reads *app.App from context
                                └─> [prod mode] NewCLI() → config.Load() → database.InitDB() → app.New()
                                      └─> app.New(db, opts...)
                                            └─> Creates all services with database.Querier
                          └─> cliInstance.App.TaskService.CreateTask(ctx, req)
                                └─> service validates, uses database.Querier
                                      └─> database.NewQuerier(db, dbType) → adapter
                                            └─> SQLC generated code → SQL execution
                          └─> returns *taskCreateResult (implements GetID() + PrettyPrint())
                    └─> formatter.Success(result)
                          ├─> quiet: fmt.Printf("%d\n", result.GetID())
                          ├─> json: encode {"success": true, "data": {...}}
                          └─> human: result.PrettyPrint(colorScheme)
        └─> returns error (may be *cli.ExitErr)

main.go:main()
  └─> checks if error is *cli.ExitErr → os.Exit(exitErr.Code)
```

### TUI Startup Flow

```
main.go:init() registers tui command → launcher.Launch()
  └─> loads config, connects daemon (optional), initializes DB, creates app.App
        └─> creates tui/core.App (wraps tui.Model)
              └─> tea.NewProgram(app).Run()
                    └─> Elm Architecture: Init → Update (handle msgs) → View (render)
```

---

## Layers

### Entry Point (`main.go`)

**Purpose**: Cobra root command setup, subcommand registration, exit code handling.

**Key details**:
- Root command defined at package level (line 28)
- `init()` registers all subcommands: `task`, `project`, `column`, `label`, `assignee`, `tutorial`, `setup`, `db`, `tui`, `completion`
- `main()` silences Cobra's default errors, executes root command, unwraps `*cli.ExitErr` for proper exit codes
- Build metadata (`version`, `commit`, `date`) injected via ldflags

**Files**: `main.go` (138 lines)

---

### CLI Commands (`internal/cli/{domain}/`)

**Purpose**: Define commands, parse flags, orchestrate service calls, format output.

**Structure**: Each domain (task, project, column, label, assignee) has:
- `{domain}.go` - parent command returning `*cobra.Command`
- `{action}.go` - subcommands (create, list, update, delete, etc.)
- `{action}_logic.go` - **pure business logic** (parsing, validation, formatting)
- `{action}_logic_test.go` - **unit tests** for logic functions (100% coverage target)

**Two command patterns**:

1. **Handler pattern** (preferred for complex commands):
   ```go
   RunE: handler.Command(&createHandler{}, parseCreateFlags)
   ```
   Handler implements `Execute(ctx, *Arguments) (any, error)`, result auto-formatted.

2. **Direct RunE** (simple commands):
   ```go
   RunE: func(cmd *cobra.Command, args []string) error { ... }
   ```
   Manually construct `OutputFormatter` and call `cli.GetCLIFromContext()`.

**IoC pattern**: Extract pure functions into `*_logic.go`:
- Parsing: `ParseAssignArgs(args []string) (*AssignInput, error)`
- Formatting: `FormatAssignOutput(result *AssignResult) string`
- JSON: `FormatAssignJSON(result *AssignResult) map[string]any`

Test these directly in same package (`task`, not `task_test`) for coverage metrics.

**Files**:
- `internal/cli/task/*.go` (16 commands, 15 logic files, 15 test files)
- `internal/cli/project/*.go`, `internal/cli/column/*.go`, etc.

---

### Handler Framework (`internal/cli/handler/`)

**Purpose**: Reduce command boilerplate, standardize argument parsing and output.

**`Handler` interface** (line 14-18):
```go
type Handler interface {
    Execute(ctx context.Context, args *Arguments) (any, error)
}
```

**`Command()` bridge** (line 40-70):
```go
func Command(handler Handler, parseFlags func(*cobra.Command) error) func(*cobra.Command, []string) error
```
Returns Cobra-compatible `RunE` function that:
1. Calls `parseFlags(cmd)` for validation
2. Extracts `--json`, `--quiet` flags → `OutputFormatter`
3. Converts all flags to `map[string]any` → `Arguments`
4. Calls `handler.Execute(ctx, args)`
5. Passes result to `formatter.Success(result)`

**`Arguments` struct** (line 20-25): Provides typed getters:
- `MustGetString(name)`, `GetString(name, default)`
- `MustGetInt(name)`, `GetInt(name, default)`
- `GetBool(name)`, `GetStringSlice(name, default)`, `GetIntSlice(name, default)`

**`FlagParser` helper** (`flag_parser.go`): Domain-specific parsers wrapping `*cobra.Command`:
- `ParseProjectID()`, `ParseTaskID()`, `ParseColumnID()`, `ParseLabelID()`
- `ParseString()`, `ParseColor()`

**Files**: `handler/handler.go` (214 lines), `handler/flag_parser.go` (139 lines)

---

### Output Formatting (`internal/cli/output.go`)

**Purpose**: Consistent output across three modes: quiet, JSON, human-readable.

**`OutputFormatter` struct** (line 15-18):
```go
type OutputFormatter struct {
    JSON  bool  // -j flag
    Quiet bool  // -q flag
}
```

**Output modes**:

1. **Quiet** (`-q`): For shell scripting. Calls `data.(interface{ GetID() int })` and prints just the ID.
   ```bash
   TASK_ID=$(paso task create -t "Fix bug" -p 1 -q)
   ```

2. **JSON** (`-j`): For programmatic use. Encodes:
   ```json
   {"success": true, "data": {...}}
   ```
   Errors: `{"success": false, "error": {"code": "...", "message": "..."}}`

3. **Human-readable** (default): Checks if data implements `PrettyPrintable`:
   ```go
   type PrettyPrintable interface {
       PrettyPrint(colors.ColorScheme) string
   }
   ```
   Falls back to `fmt.Stringer`, then `%+v` dump.

**Result types**: Each command defines a struct implementing both:
- `GetID() int` for quiet mode
- `PrettyPrint(ColorScheme) string` for human mode

**Exit codes** (`exitcodes.go`):
```go
ExitSuccess    = 0  // Normal completion
ExitError      = 1  // General error
ExitUsage      = 2  // Incorrect usage
ExitNotFound   = 3  // Resource not found
ExitDataErr    = 4  // Invalid data
ExitValidation = 5  // Validation error
```

**`ExitErr`**: Carries exit code as regular `error`, allowing deferred cleanup to run before `os.Exit()`.

**Files**: `output.go` (119 lines), `exitcodes.go` (55 lines)

---

### Styles (`internal/cli/styles/`)

**Purpose**: Terminal styling with Lipgloss, themed rendering for success/error messages.

**Key functions**:
- `Init(ColorScheme)` - Initializes lipgloss styles (runs once via `sync.Once`)
- `RenderSuccess(message, colorScheme)` - Checkmark + message
- `RenderSuccessWithDetails(message, details, colorScheme)` - Checkmark + key-value pairs
- `RenderError(message, colorScheme)` - X mark + "Error: message"
- `ColoredText()`, `BoldColoredText()`, `RenderLabelChip()`, `RenderTaskReference()`, `RenderCard()`

**Styles**:
- `CardStyle` - Rounded border, 80-char width, accent color
- `TitleStyle`, `SubtitleStyle`, `LabelStyle`, `ValueStyle` - Text styles
- `BlockedStyle`, `SuccessStyle`, `ErrorStyle`, `WarningStyle` - Semantic styles

**Tree characters** (`tree.go`):
- `TreeBranch` (`├── `), `TreeLastBranch` (`└── `), `TreeVertical` (`│   `), `TreeSpace` (`    `)

**Golden tests**: 80+ `.golden` files in `testdata/` verify styled output across all themes.

**Files**: `styles.go` (144 lines), `success.go` (85 lines), `color.go` (50 lines), `tree.go` (13 lines)

---

### App Container (`internal/app/`)

**Purpose**: Dependency injection container holding all services.

**`App` struct** (line 19-35):
```go
type App struct {
    eventClient     events.EventPublisher  // Optional daemon connection
    dbType          database.DatabaseType
    GitDetector     git.Detector
    TaskService     taskservice.Service
    ProjectService  projectservice.Service
    ColumnService   columnservice.Service
    LabelService    labelservice.Service
    AssigneeService assigneeservice.Service
}
```

**Construction** (`app.go`, line 42-104):
```go
func New(db *sql.DB, opts ...Option) (*App, error)
```

**Functional options** (`options.go`):
- `WithEventPublisher(ec)` - Connect to daemon for live TUI updates
- `WithDatabaseType(dbType)` - SQLite or PostgreSQL
- `WithGitDetector(gd)` - Injectable git detector (default: real, test: mock)
- `WithLogger(logger)` - Custom logger

**Pattern**: All services receive `*sql.DB` and `database.DatabaseType`, create their `database.Querier` internally.

**Files**: `app.go` (112 lines), `options.go` (50 lines)

---

### Services (`internal/services/{domain}/`)

**Purpose**: Business logic layer, transaction management, validation, event publishing.

**Pattern**: Interface segregation. Example from `services/task/service.go`:

```go
type Service interface {
    TaskReader      // GetTaskDetail, GetTaskSummariesByProject, etc.
    TaskMover       // MoveTaskToNextColumn, MoveTaskUp, etc.
    TaskWriter      // CreateTask, UpdateTask, DeleteTask
    TaskRelationer  // AddParentRelation, RemoveChildRelation
    TaskLabeler     // AttachLabel, DetachLabel
    TaskCommenter   // CreateComment, UpdateComment, DeleteComment
}
```

**Each service package has**:
- `service.go` - Main implementation, transaction management
- `validation.go` - Input validation (separated from service code)
- `errors.go` - Sentinel error types (e.g., `ErrTaskNotFound`)

**Request structs**: Use pointers for optional fields:
```go
type UpdateTaskRequest struct {
    TaskID      int
    Title       *string  // nil means don't update
    Description *string
    PriorityID  *int
    TypeID      *int
}
```

**Database access**: Services use `database.Querier` interface, obtained via:
```go
queries, err := database.NewQuerier(db, s.dbType)
```

**Event publishing**: Optional `events.EventPublisher` for live TUI updates when data changes.

**Files**:
- `services/task/service.go` (30+ methods)
- `services/project/service.go`, `services/column/service.go`, etc.

---

### Database Layer (`internal/database/`)

**Purpose**: Dual-database support (SQLite/PostgreSQL) with type-safe SQL via SQLC.

**Three-tier abstraction**:

1. **SQLC codegen** - SQL queries per dialect:
   - `sql_sqlite/*.sql` → `generated_sqlite/` (Go code)
   - `sql_postgres/*.sql` → `generated_postgres/` (Go code)

2. **Adapters** - Implement unified `types.Querier` interface:
   - `adapters/sqlite/adapter.go` - Wraps `generated_sqlite.Queries`
   - `adapters/postgres/adapter.go` - Wraps `generated_postgres.Queries`
   - Each has `convert_to.go` (params) and `convert_from.go` (results)

3. **Factory** (`factory.go`, line 53-77):
   ```go
   func NewQuerier(db any, dbType DatabaseType) (Querier, error)
   ```
   Accepts `*sql.DB` or `*sql.Tx`, returns appropriate adapter.

**`Querier` interface** (`types/querier.go`, 228 lines): 80+ database operations using database-agnostic types. Services depend on this interface, never concrete adapters.

**Database initialization** (`db.go`, line 27-62):
```go
func InitDB(ctx context.Context, cfg Config, name string) (*sql.DB, error)
```
1. Opens connection (SQLite: `modernc.org/sqlite`, PostgreSQL: `lib/pq`)
2. Configures SQLite: foreign keys ON, WAL mode, busy timeout 5s, single writer
3. Configures PostgreSQL: version check (min v12), connection pool (25 open, 5 idle)
4. Runs migrations via goose

**Migrations** (`migrations.go`):
- `migrations_sqlite/` - 9 migration files (00001-00009)
- `migrations_postgres/` - 8 migration files (00001-00009, no 00004)
- Embedded via `//go:embed`
- `sync.Mutex` protects goose's global state for parallel test safety
- Seeds default data: default project, 3 columns (Todo, In Progress, Done), GitHub-style labels

**Files**:
- `db.go` (209 lines), `factory.go` (89 lines), `migrations.go` (138 lines)
- `types/querier.go` (228 lines)

---

### Models (`internal/models/`)

**Purpose**: Plain data transfer objects with no database dependencies.

**Key models**:
- `Task`, `TaskReference`, `TaskSummary`, `TaskDetail`, `TaskTreeNode`, `TaskRelation`
- `Project`, `Column` (doubly-linked list with `PrevID`/`NextID`)
- `Label`, `Assignee`, `Comment`, `TaskEvent`, `ActivityItem`

**Constants** (`constants.go`):
```go
// Relation types
RelationBlocks     = 1  // This task blocks another
RelationBlockedBy  = 2  // This task is blocked by another
RelationRelated    = 3  // Related tasks

// Task types
TaskTypeTask    = 1
TaskTypeFeature = 2

// Priorities
PriorityTrivial = 1
PriorityLow     = 2
PriorityMedium  = 3
PriorityHigh    = 4
PriorityCritical = 5

DefaultTaskPosition = 9999
```

**Metadata**: `priority_type.go`, `task_type.go`, `relation_type.go` provide human-readable names.

**Files**: 10 model files (15-100 lines each)

---

### Converters (`internal/converters/`)

**Purpose**: Translate between `database/types` result types and `models` types.

One file per domain: `task.go`, `column.go`, `label.go`, `assignee.go`, `task_event.go`.

Example:
```go
func TaskDetailFromDB(dbTask *types.TaskDetailResult) *models.TaskDetail
```

**Files**: 5 converter files (20-80 lines each)

---

### Config (`internal/config/`)

**Purpose**: YAML configuration, color scheme presets, key mappings.

**`Config` struct** (line 22-28):
```go
type Config struct {
    KeyMappings    KeyMappings        `yaml:"key_mappings"`
    ColorScheme    colors.ColorScheme `yaml:"theme"`
    Databases      []DatabaseConfig   `yaml:"databases"`
    ActiveDatabase string             `yaml:"active_database,omitempty"`
    ActiveAssignee string             `yaml:"active_assignee,omitempty"`
}
```

**Location**: `$XDG_CONFIG_HOME/paso/config.yaml` or `~/.config/paso/config.yaml`

**`Load()`** (line 57-124):
1. Returns defaults if file doesn't exist
2. Parses YAML, warns if permissions too permissive (should be 0600)
3. Loads optional theme override from `PASO_THEME_FILE` env var
4. Auto-detects database type for legacy configs

**Color schemes** (`config/colors/`):
- `ColorScheme` struct with 23 configurable hex colors
- 5 built-in presets: **default** (purple), **monochrome** (grayscale), **wave**/**dragon**/**lotus** (Kanagawa palettes)
- `ApplyDefaults()` fills missing fields, `MergeFrom()` overrides individual colors
- Shared palette in `palette.go` (244 lines, 120+ named colors)

**Files**: `config.go` (251 lines), `colors/scheme.go` (218 lines), 5 preset files, `palette.go`

---

### TUI (`internal/tui/`)

**Purpose**: Interactive terminal UI built on Bubble Tea v2 (Elm Architecture).

**Framework**: Bubble Tea v2 (`bubbletea`), Lipgloss v2 (styling), Huh v2 (forms).

**Main model** (`model.go`): 60+ fields holding app state, UI state, services.

**Elm Architecture**:
- `Init()` - Initial state
- `Update(tea.Msg)` - Handle messages (keyboard, events, async results)
- `View()` - Render current state to string

**Update handlers** (split by concern):
- `update.go` - Main message dispatch
- `update_normal.go` - Keyboard navigation
- `update_forms.go` - Form input
- `update_pickers.go` - Picker selection
- `update_search.go`, `update_confirmations.go`, `update_comments.go`, etc.

**View renderers**:
- `view_board.go` - Kanban board
- `view_forms.go`, `view_confirmations.go`, `view_comments.go`

**Sub-packages**:
- `state/` - State structs (AppState, UIState, PickerStates, FormStates, etc.)
- `components/` - Reusable UI (column, task card, detail panel, status bar, tabs, label chip)
- `huhforms/` - Form definitions (task, project, column, label, comment, database)
- `commands/` - Async commands (`tea.Cmd` for prefetch)
- `helpers/` - Navigation, scroll, tips
- `notifications/` - Notification rendering
- `theme/` - TUI-specific color setup
- `renderers/`, `layout/`, `text/`, `modelops/`

**Launch**: `launcher.Launch()` creates signal context, loads config, connects daemon, initializes DB, creates `app.App`, constructs `core.App`, runs `tea.NewProgram()`.

**Files**: `model.go` (360 lines), 20+ update/view files, 10+ state files, 10+ component files

---

## Key Patterns

### IoC for Testability

**Problem**: Cobra command handlers mix CLI concerns with business logic, making 0% test coverage.

**Solution**: Extract pure functions into `*_logic.go` files:

```go
// assign_logic.go
type AssignInput struct {
    TaskID       int
    AssigneeName string
}

type AssignResult struct {
    TaskID       int
    AssigneeName string
    Cleared      bool
}

func ParseAssignArgs(args []string) (*AssignInput, error)
func FormatAssignOutput(result *AssignResult) string
func FormatAssignJSON(result *AssignResult) map[string]any
```

Command handler becomes thin orchestration:
```go
func runAssign(cmd *cobra.Command, args []string) error {
    input, err := ParseAssignArgs(args)  // Pure function
    // ... service call ...
    result := &AssignResult{...}
    cli.PrintSuccess(FormatAssignOutput(result))  // Pure function
}
```

**Test** pure functions directly in same package (`task`, not `task_test`) for coverage:
```go
// assign_logic_test.go
package task  // Same package!

func TestParseAssignArgs(t *testing.T) {
    tests := []struct {
        name string
        args []string
        want *AssignInput
        err  bool
    }{
        {"valid", []string{"42", "alice"}, &AssignInput{42, "alice"}, false},
        // ...
    }
    // ...
}
```

**Result**: 33.2% coverage on `internal/cli/task` package from 0%.

---

### Handler Pattern

**When to use**:
- Commands with complex flag parsing
- Commands needing consistent output formatting
- New commands (it's the recommended default)

**Structure**:
```go
type createHandler struct{}

func (h *createHandler) Execute(ctx context.Context, args *Arguments) (any, error) {
    cliInstance, _ := cli.GetCLIFromContext(ctx)
    // Business logic...
    return &taskCreateResult{...}, nil
}

func CreateCmd() *cobra.Command {
    return &cobra.Command{
        Use: "create",
        RunE: handler.Command(&createHandler{}, parseCreateFlags),
    }
}
```

**When to use direct `RunE`**:
- Very simple commands (e.g., `whoami`, commands with no flags)
- Commands that need custom output logic not fitting the handler pattern

---

### Adapter Pattern (Database Abstraction)

**Goal**: Support both SQLite and PostgreSQL with zero changes to service code.

**How it works**:

1. **Services depend on `database.Querier` interface**:
   ```go
   queries, _ := database.NewQuerier(db, s.dbType)
   tasks, _ := queries.GetTasksByColumn(ctx, columnID)
   ```

2. **Factory returns appropriate adapter**:
   ```go
   // SQLite: returns sqlite.Adapter wrapping generated_sqlite.Queries
   // PostgreSQL: returns postgres.Adapter wrapping generated_postgres.Queries
   ```

3. **Adapters convert types**:
   - `convert_to.go`: `types.CreateTaskParams` → `generated_sqlite.CreateTaskParams`
   - `convert_from.go`: `generated_sqlite.Task` → `types.TaskResult`

4. **SQLC generates dialect-specific code**:
   - `sql_sqlite/tasks.sql` → `generated_sqlite/tasks.sql.go`
   - `sql_postgres/tasks.sql` → `generated_postgres/tasks.sql.go`

**Result**: Services never know which database they're using. Switching is transparent.

---

### Context Injection (Test Mode)

**Pattern**: Use context to inject test dependencies.

**Production**:
```go
cli.GetCLIFromContext(ctx)  // ctx has nothing → creates new CLI, connects DB
```

**Test**:
```go
db, app := cli.SetupCLITest(t)
ctx := context.WithValue(context.Background(), appcontext.AppKey, app)
cli.GetCLIFromContext(ctx)  // Returns CLI wrapping injected app
```

**Benefit**: No database connection, no migrations, tests run in milliseconds.

**Files**: `appcontext/appcontext.go` (single constant), `cli/cli.go` (line 30-43)

---

### Functional Options

**Pattern**: Configure structs with variadic option functions.

**Example** (`app.New`):
```go
func New(db *sql.DB, opts ...Option) (*App, error)

type Option func(*appConfig)

func WithEventPublisher(ec events.EventPublisher) Option
func WithDatabaseType(dbType database.DatabaseType) Option
func WithGitDetector(gd git.Detector) Option
```

**Usage**:
```go
// Production
app, _ := app.New(db)

// Test with mocks
app, _ := app.New(db, app.WithGitDetector(mockGit), app.WithEventPublisher(nil))
```

**Benefits**: Clean API, extensible without breaking changes, testable.

---

### Interface Segregation

**Pattern**: Compose large interfaces from focused sub-interfaces.

**Example** (`services/task/service.go`):
```go
type TaskReader interface {
    GetTaskDetail(ctx, taskID) (*TaskDetail, error)
    GetTaskSummariesByProject(ctx, projectID) (map[int][]*TaskSummary, error)
    // ... 8 read methods
}

type TaskWriter interface {
    CreateTask(ctx, req) (*Task, error)
    UpdateTask(ctx, req) error
    DeleteTask(ctx, taskID) error
}

type Service interface {
    TaskReader
    TaskWriter
    TaskMover
    TaskRelationer
    TaskLabeler
    TaskCommenter
}
```

**Benefits**: Tests can mock just `TaskReader`, code depends on minimal interface needed.

---

## Testing

### Test Categories

1. **Unit tests** (`*_logic_test.go`, `*_test.go` within packages):
   - Pure function tests (validation, parsing, formatting)
   - No database, no I/O, runs in microseconds
   - **Same package** (`task`, not `task_test`) for coverage metrics
   - Example: `internal/cli/task/assign_logic_test.go` (100% coverage on `assign_logic.go`)

2. **Integration tests** (within service/database packages):
   - Use in-memory SQLite via `fixtures.SetupTestDB()`
   - Test service methods with real database operations
   - Example: `internal/services/task/task_crud_test.go`

3. **Golden tests**:
   - `internal/cli/golden/golden_test.go` - Help output snapshots (5 `.golden` files)
   - `internal/cli/styles/testdata/*.golden` - Styled output across all themes (80+ files)
   - Run with `-update` flag to regenerate: `go test ./internal/cli/golden -update`

4. **Acceptance tests**:
   - `tests/acceptance/cli_validation.sh` - Shell-based CLI validation

### Test Infrastructure

**`internal/testing/fixtures/`** (DB setup, data helpers):
- `SetupTestDB(tb)` - Creates in-memory SQLite with production migrations (no seed data)
- `CreateTestProject(t, db, dialect, name)` - Parameterized with `Dialect` for SQLite/PostgreSQL portability
- `CreateTestTask()`, `CreateTestLabel()`, `CreateTestColumn()`, etc.
- `AssertTaskInColumn()`, `AssertTaskLabelCount()` - Test assertions

**`internal/testing/cli/`** (Command execution harness):
- `SetupCLITest(tb)` - Returns in-memory DB + `*app.App`
- `SetupCLITestWithGit(tb)` - Adds mock git detector
- `ExecuteCLICommand(t, testApp, cmd, args)` - Injects app via context, captures stdout/stderr
- `ExecuteCLICommandWithMocks(t, services, cmd, args)` - Pure unit testing with mock services
- `CaptureOutputFunc()` - Thread-safe stdout/stderr capture with mutex

**`internal/testing/mocks/`** (Hand-rolled mocks):
- `MockTaskService`, `MockProjectService`, `MockColumnService`, `MockLabelService`, `MockAssigneeService`
- `MockEventPublisher`, `MockGitDetector`, `MockTaskEventService`
- Thread-safe, call recording, per-method error injection, `HasCall()`/`CallCount()` helpers

### Coverage Strategy

**Problem**: External package tests (`task_test`) don't count toward `internal/cli/task` coverage.

**Solution**: Write unit tests in **same package** (`package task`):
```go
// assign_logic_test.go
package task  // Same as assign_logic.go!

func TestParseAssignArgs(t *testing.T) { /* ... */ }
```

**Result**: Coverage metrics reflect actual tested code. Went from 0% → 33.2% on CLI task package.

### Conventions

- **`t.Parallel()`**: All tests run concurrently
- **Table-driven tests**: Use subtests with structured test cases
- **`stretchr/testify`**: `assert` (soft), `require` (hard) assertions
- **`goleak`**: Every `main_test.go` uses `goleak.VerifyTestMain(m)` for goroutine leak detection
- **Test naming**: `Test{FunctionName}` or `Test{Type}_{Method}` (e.g., `TestTaskService_CreateTask`)

---

## Adding a New CLI Command

**Checklist**:

1. **Create command file** (`internal/cli/task/my_command.go`):
   ```go
   func MyCommandCmd() *cobra.Command {
       return &cobra.Command{
           Use:   "my-command <arg>",
           Short: "Brief description",
           RunE:  handler.Command(&myCommandHandler{}, parseMyCommandFlags),
       }
   }
   
   type myCommandHandler struct{}
   
   func (h *myCommandHandler) Execute(ctx context.Context, args *Arguments) (any, error) {
       // Thin orchestration only
       input, err := ParseMyCommandArgs(args.Args)  // Pure function
       cliInstance, _ := cli.GetCLIFromContext(ctx)
       // Service call...
       result := &MyCommandResult{...}
       return result, nil
   }
   ```

2. **Extract pure logic** (`internal/cli/task/my_command_logic.go`):
   ```go
   package task
   
   type MyCommandInput struct { /* ... */ }
   type MyCommandResult struct { /* ... */ }
   
   func ParseMyCommandArgs(args []string) (*MyCommandInput, error) { /* ... */ }
   func FormatMyCommandOutput(result *MyCommandResult) string { /* ... */ }
   func FormatMyCommandJSON(result *MyCommandResult) map[string]any { /* ... */ }
   ```

3. **Write tests** (`internal/cli/task/my_command_logic_test.go`):
   ```go
   package task  // Same package!
   
   func TestParseMyCommandArgs(t *testing.T) {
       t.Parallel()
       tests := []struct {
           name string
           args []string
           want *MyCommandInput
           err  bool
       }{
           // Test cases...
       }
       for _, tt := range tests {
           t.Run(tt.name, func(t *testing.T) {
               t.Parallel()
               got, err := ParseMyCommandArgs(tt.args)
               if tt.err {
                   require.Error(t, err)
               } else {
                   require.NoError(t, err)
                   assert.Equal(t, tt.want, got)
               }
           })
       }
   }
   ```

4. **Register command** (`internal/cli/task/task.go`):
   ```go
   cmd.AddCommand(MyCommandCmd())
   ```

5. **Implement result interfaces**:
   ```go
   func (r *MyCommandResult) GetID() int { return r.TaskID }
   func (r *MyCommandResult) PrettyPrint(cs colors.ColorScheme) string { /* ... */ }
   ```

6. **Run tests**:
   ```bash
   go test ./internal/cli/task -v
   go test ./internal/cli/task -coverprofile=coverage.out
   go tool cover -func=coverage.out | grep my_command_logic
   ```

7. **Update golden tests** (if command has help output):
   ```bash
   go test ./internal/cli/golden -update
   ```

8. **Verify**: `go test ./...` passes, coverage on logic functions is 100%.

---

## Additional Resources

- **SQLC docs**: https://docs.sqlc.dev/
- **Cobra docs**: https://cobra.dev/
- **Bubble Tea tutorial**: https://github.com/charmbracelet/bubbletea
- **Lipgloss examples**: https://github.com/charmbracelet/lipgloss
- **Goose migrations**: https://github.com/pressly/goose

**Contributing**: See individual package READMEs (if present) for domain-specific patterns.
