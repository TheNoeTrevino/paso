# Paso Refactoring Plan

**Date**: December 24, 2024  
**Status**: Planning Phase  
**Goal**: Improve code maintainability, development speed, testing, and performance

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Current State Analysis](#current-state-analysis)
3. [Goals & Success Criteria](#goals--success-criteria)
4. [Phase 1: Service Layer Implementation](#phase-1-service-layer-implementation)
5. [Phase 2: SQLC Migration](#phase-2-sqlc-migration)
6. [Phase 3: Infrastructure Improvements](#phase-3-infrastructure-improvements)
7. [Phase 4: Documentation & Polish](#phase-4-documentation--polish)
8. [Timeline & Milestones](#timeline--milestones)
9. [Risk Mitigation](#risk-mitigation)
10. [References](#references)

---

## Executive Summary

This document outlines a comprehensive refactoring plan for the Paso project, focusing on four key priorities:

1. **Code Maintainability** - Clear separation of concerns, readable code
2. **Development Speed** - Faster feature development, less boilerplate
3. **Better Testing** - Testable business logic, mock-friendly architecture
4. **Performance** - Efficient database queries, caching capabilities

### Key Improvements

- **Service Layer**: Separate business logic from data access
- **SQLC Migration**: Type-safe SQL code generation (70% code reduction)
- **App Structure**: Centralized dependency injection
- **Keep TUI as-is**: Current structure is already well-organized

### Timeline

- **Phase 1**: Service Layer (2 weeks)
- **Phase 2**: SQLC Migration (2 weeks)
- **Phase 3**: Infrastructure (1 week)
- **Phase 4**: Documentation (1 week)
- **Total**: 6 weeks

---

## Current State Analysis

### What's Working Well ✅

- **TUI Structure**: `model.go` (1,008 lines) is well-organized
- **State Management**: Separate `internal/tui/state/` package
- **Update/View Splitting**: Feature-based file organization
- **CLI Commands**: Nested structure works well for multiple subcommands
- **Testing**: Co-located tests with good coverage

### What Needs Improvement ⚠️

| Issue | Current State | Impact |
|-------|---------------|--------|
| **No Service Layer** | Business logic in repositories/TUI | Hard to test, tight coupling |
| **Hand-written SQL** | 1,090 lines in task_repository.go | Error-prone, boilerplate-heavy |
| **TUI-Database Coupling** | TUI directly calls repositories | Can't test without database |
| **Large main.go** | 121 lines with command definitions | Entry point not minimal |

### Architecture Comparison: Paso vs Crush

After analyzing the Crush codebase, key learnings:

1. **Service Layer**: Crush separates business logic (services) from data access (repositories)
2. **SQLC Usage**: All database code is generated from SQL files
3. **Component-Based TUI**: Heavy componentization (not needed for Paso)
4. **Package Documentation**: `doc.go` files explain complex packages

**Decision**: Adopt service layer and SQLC; keep TUI structure as-is.

---

## Goals & Success Criteria

### Goal 1: Code Maintainability

**Success Criteria:**
- [ ] Business logic separated into service layer
- [ ] SQL queries in separate `.sql` files
- [ ] Each package has clear, single responsibility
- [ ] Package documentation exists for complex packages

**Metrics:**
- Reduce database repository code by 60-70%
- Service layer covers 100% of business operations
- All services have interfaces for testing

### Goal 2: Development Speed

**Success Criteria:**
- [ ] New features can be added without touching multiple layers
- [ ] Database queries can be added/modified in SQL files only
- [ ] Service layer provides reusable business logic

**Metrics:**
- Time to add new database query: < 5 minutes (just write SQL, regenerate)
- Time to add new business operation: < 30 minutes (implement in service)

### Goal 3: Better Testing

**Success Criteria:**
- [ ] Services can be tested without database (using mocks)
- [ ] SQLC generates mock interfaces automatically
- [ ] Integration tests cover critical workflows

**Metrics:**
- Service layer test coverage: > 80%
- Repository test coverage: > 90%
- Integration test suite: < 10 seconds

### Goal 4: Performance

**Success Criteria:**
- [ ] SQLC-generated queries are optimized
- [ ] Service layer can cache frequently accessed data
- [ ] Database connection pooling configured

**Metrics:**
- Database query performance: No regressions
- TUI startup time: < 100ms (no regressions)
- Memory usage: No significant increase

---

## Phase 1: Service Layer Implementation

**Duration**: 2 weeks  
**Priority**: CRITICAL  
**Dependencies**: None

### Overview

Create a service layer between the UI (TUI/CLI) and the repository layer. Services encapsulate business logic, validation, and orchestrate repository operations.

### Architecture

```
Before:
TUI/CLI → Repository → Database

After:
TUI/CLI → Service → Repository → Database
```

### Directory Structure

```
internal/
├── services/
│   ├── task/
│   │   ├── service.go        # Interface + implementation
│   │   ├── service_test.go   # Unit tests (mock repository)
│   │   ├── validation.go     # Business rules
│   │   └── errors.go         # Domain errors
│   ├── project/
│   │   ├── service.go
│   │   └── service_test.go
│   ├── column/
│   │   ├── service.go
│   │   └── service_test.go
│   └── label/
│       ├── service.go
│       └── service_test.go
├── app/
│   └── app.go               # Dependency injection container
└── database/                # Existing repositories (unchanged)
```

### Step 1.1: Create Task Service Interface (Day 1)

**File**: `internal/services/task/service.go`

**Tasks:**
1. Define `Service` interface with all task operations:
   - Read: `GetTaskDetail`, `GetTaskSummariesByProject`, `GetTaskReferencesForProject`
   - Write: `CreateTask`, `UpdateTask`, `DeleteTask`
   - Movement: `MoveTaskToColumn`, `MoveTaskUp`, `MoveTaskDown`
   - Relations: `AddParentRelation`, `RemoveParentRelation`, etc.
   - Labels: `AttachLabel`, `DetachLabel`

2. Define request/response types:
   - `CreateTaskRequest` - encapsulates all creation parameters
   - `UpdateTaskRequest` - encapsulates update parameters (optional fields)

3. Define service errors:
   - `ErrEmptyTitle`
   - `ErrTitleTooLong`
   - `ErrInvalidColumnID`
   - etc.

**Example Interface:**

```go
package task

type Service interface {
    // Read operations
    GetTaskDetail(ctx context.Context, taskID int) (*models.TaskDetail, error)
    GetTaskSummariesByProject(ctx context.Context, projectID int) (map[int][]*models.TaskSummary, error)
    GetTaskReferencesForProject(ctx context.Context, projectID int) ([]*models.TaskReference, error)
    
    // Write operations
    CreateTask(ctx context.Context, req CreateTaskRequest) (*models.Task, error)
    UpdateTask(ctx context.Context, req UpdateTaskRequest) error
    DeleteTask(ctx context.Context, taskID int) error
    
    // Task movements
    MoveTaskToColumn(ctx context.Context, taskID, columnID int) error
    MoveTaskUp(ctx context.Context, taskID int) error
    MoveTaskDown(ctx context.Context, taskID int) error
    
    // Task relationships
    AddParentRelation(ctx context.Context, taskID, parentID int) error
    RemoveParentRelation(ctx context.Context, taskID, parentID int) error
    
    // Label management
    AttachLabel(ctx context.Context, taskID, labelID int) error
    DetachLabel(ctx context.Context, taskID, labelID int) error
}

type CreateTaskRequest struct {
    Title       string
    Description string
    ColumnID    int
    Position    int
    PriorityID  int
    TypeID      int
    LabelIDs    []int
    ParentIDs   []int
}

type UpdateTaskRequest struct {
    TaskID      int
    Title       *string  // nil = don't update
    Description *string
    PriorityID  *int
    TypeID      *int
}
```

### Step 1.2: Implement Task Service (Days 2-3)

**File**: `internal/services/task/service.go`

**Tasks:**
1. Create `service` struct that implements `Service` interface
2. Constructor: `NewService(repo database.DataStore, eventClient events.EventPublisher) Service`
3. Implement `CreateTask` method:
   - Validate request (use validation.go)
   - Call repository to create task
   - Set priority/type if provided
   - Attach labels
   - Add parent relationships
   - Publish event
4. Implement all other methods following same pattern

**Key Implementation Details:**
- All business logic goes here (validation, orchestration)
- Repository is just data access, no business rules
- Service handles multi-step operations (create + attach labels + add relations)
- Service publishes events (if event client available)

### Step 1.3: Write Unit Tests for Task Service (Day 4)

**File**: `internal/services/task/service_test.go`

**Tasks:**
1. Create mock repository (implements `database.DataStore`)
2. Test `CreateTask` with various scenarios:
   - Valid request → success
   - Empty title → `ErrEmptyTitle`
   - Title too long → `ErrTitleTooLong`
   - Invalid column ID → `ErrInvalidColumnID`
3. Test that repository methods are called correctly
4. Test that events are published

**Example Test:**

```go
func TestCreateTask_Success(t *testing.T) {
    mockRepo := &MockDataStore{}
    mockEvents := &MockEventPublisher{}
    svc := task.NewService(mockRepo, mockEvents)
    
    req := task.CreateTaskRequest{
        Title:    "Test Task",
        ColumnID: 1,
        Position: 0,
    }
    
    mockRepo.On("CreateTask", mock.Anything, "Test Task", "", 1, 0).
        Return(&models.Task{ID: 1, Title: "Test Task"}, nil)
    
    task, err := svc.CreateTask(context.Background(), req)
    
    assert.NoError(t, err)
    assert.NotNil(t, task)
    assert.Equal(t, "Test Task", task.Title)
    mockRepo.AssertExpectations(t)
}

func TestCreateTask_EmptyTitle(t *testing.T) {
    svc := task.NewService(nil, nil)
    
    req := task.CreateTaskRequest{
        Title:    "",
        ColumnID: 1,
    }
    
    _, err := svc.CreateTask(context.Background(), req)
    
    assert.ErrorIs(t, err, task.ErrEmptyTitle)
}
```

### Step 1.4: Create App Struct (Day 5)

**File**: `internal/app/app.go`

**Tasks:**
1. Create `App` struct that holds all services
2. Constructor: `New(ctx context.Context, db *sql.DB, cfg *config.Config) (*App, error)`
3. Wire up services with dependencies
4. Add `Shutdown()` method for cleanup

**Example:**

```go
package app

type App struct {
    // Services
    TaskService    task.Service
    ProjectService project.Service
    ColumnService  column.Service
    LabelService   label.Service
    
    // Infrastructure
    DB          *sql.DB
    Repo        database.DataStore
    EventClient events.EventPublisher
    Config      *config.Config
    
    ctx          context.Context
    cleanupFuncs []func() error
}

func New(ctx context.Context, db *sql.DB, cfg *config.Config) (*App, error) {
    repo := database.NewDataStore(db)
    
    var eventClient events.EventPublisher
    if cfg.DaemonEnabled {
        eventClient, _ = events.NewClient(cfg.DaemonAddress)
        // Don't fail if daemon unavailable - it's optional
    }
    
    taskService := task.NewService(repo, eventClient)
    projectService := project.NewService(repo, eventClient)
    columnService := column.NewService(repo, eventClient)
    labelService := label.NewService(repo, eventClient)
    
    return &App{
        TaskService:    taskService,
        ProjectService: projectService,
        ColumnService:  columnService,
        LabelService:   labelService,
        DB:             db,
        Repo:           repo,
        EventClient:    eventClient,
        Config:         cfg,
        ctx:            ctx,
    }, nil
}

func (a *App) Shutdown() error {
    for _, cleanup := range a.cleanupFuncs {
        if err := cleanup(); err != nil {
            return err
        }
    }
    return a.DB.Close()
}
```

### Step 1.5: Update TUI to Use Services (Days 6-7)

**Files**: `internal/tui/model.go`, `internal/tui/update_*.go`

**Tasks:**
1. Update `Model` struct:
   - Replace `Repo database.DataStore` with `App *app.App`
2. Update `InitialModel` to take `*app.App` instead of `DataStore`
3. Update all TUI operations to use services:
   - `m.App.TaskService.CreateTask(...)` instead of `m.Repo.CreateTask(...)`
   - Build request objects from form state
4. Update form submission handlers to use service request types

**Example Changes:**

```go
// Before
func (m Model) createTask() tea.Cmd {
    return func() tea.Msg {
        task, err := m.Repo.CreateTask(
            m.Ctx,
            m.FormState.FormTitle,
            m.FormState.FormDescription,
            m.FormState.FormColumnID,
            m.FormState.FormPosition,
        )
        if err != nil {
            return TaskErrorMsg{Err: err}
        }
        
        // Manually attach labels
        for _, labelID := range m.FormState.FormLabelIDs {
            m.Repo.AttachLabel(m.Ctx, task.ID, labelID)
        }
        
        return TaskCreatedMsg{Task: task}
    }
}

// After
func (m Model) createTask() tea.Cmd {
    return func() tea.Msg {
        req := task.CreateTaskRequest{
            Title:       m.FormState.FormTitle,
            Description: m.FormState.FormDescription,
            ColumnID:    m.FormState.FormColumnID,
            Position:    m.FormState.FormPosition,
            PriorityID:  m.FormState.FormPriorityID,
            TypeID:      m.FormState.FormTypeID,
            LabelIDs:    m.FormState.FormLabelIDs,
            ParentIDs:   m.FormState.FormParentIDs,
        }
        
        task, err := m.App.TaskService.CreateTask(m.Ctx, req)
        if err != nil {
            return TaskErrorMsg{Err: err}
        }
        
        return TaskCreatedMsg{Task: task}
    }
}
```

### Step 1.6: Update CLI Commands to Use Services (Days 8-9)

**Files**: `internal/cli/task/*.go`

**Tasks:**
1. Update all CLI commands to use services via `App`
2. Simplify command implementations (less boilerplate)
3. Better error messages from service layer

### Step 1.7: Implement Other Services (Day 10)

**Tasks:**
1. Implement `project.Service` (similar pattern to task service)
2. Implement `column.Service`
3. Implement `label.Service`
4. Write unit tests for each

### Step 1.8: Integration Testing (Days 11-14)

**Tasks:**
1. Test entire TUI workflow with service layer
2. Test CLI commands with service layer
3. Fix any bugs found
4. Performance testing - ensure no regressions

### Phase 1 Deliverables

- [ ] Service layer interfaces defined
- [ ] All services implemented
- [ ] Unit tests for all services (80%+ coverage)
- [ ] TUI updated to use services
- [ ] CLI updated to use services
- [ ] Integration tests pass
- [ ] Documentation for service layer

### Phase 1 Success Criteria

- ✅ All tests passing
- ✅ No performance regressions
- ✅ TUI works identically to before
- ✅ CLI works identically to before
- ✅ Business logic is testable without database

---

## Phase 2: SQLC Migration

**Duration**: 2 weeks  
**Priority**: CRITICAL  
**Dependencies**: Phase 1 (service layer)

### Overview

Migrate from hand-written SQL strings to SQLC-generated type-safe Go code. Start with task repository, then expand to other repositories.

### Benefits

- **70% code reduction**: 1,090 lines → ~300 lines
- **Type safety**: Compile-time SQL validation
- **Faster queries**: No reflection, prepared statements
- **Auto-generated mocks**: For testing
- **Maintainability**: SQL separate from Go

### Step 2.1: Install and Configure SQLC (Day 1)

**Tasks:**
1. Install SQLC: `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`
2. Create `sqlc.yaml` in project root:

```yaml
version: "2"
sql:
  - engine: "sqlite"
    schema: "internal/database/migrations"
    queries: "internal/database/sql"
    gen:
      go:
        package: "database"
        out: "internal/database/generated"
        emit_json_tags: true
        emit_prepared_queries: false
        emit_interface: true
        emit_exact_table_names: false
        emit_empty_slices: true
        emit_pointers_for_null_types: true
```

3. Create directory structure:

```
internal/database/
├── migrations/              # SQL schema files
├── sql/                     # SQL query files
├── generated/               # SQLC-generated code (gitignore)
├── task_repository.go       # Wraps generated code
├── column_repository.go
├── label_repository.go
├── project_repository.go
└── db.go
```

4. Add to `.gitignore`:
```
internal/database/generated/
```

### Step 2.2: Extract Migrations to SQL Files (Days 2-3)

**Current**: Migrations are embedded in Go code (`migrations.go`)  
**Target**: Sequential SQL migration files

**Tasks:**
1. Create `internal/database/migrations/001_initial.sql`
2. Extract schema from `migrations.go`:
   - Projects table
   - Project counters table
   - Columns table
   - Tasks table
   - Labels table
   - Task-label junction table
   - Priorities table
   - Types table
   - Task relations table
3. Add all indexes
4. Update migration runner to read from SQL files

**Example Migration File:**

```sql
-- internal/database/migrations/001_initial.sql

-- Projects
CREATE TABLE IF NOT EXISTS projects (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Project ticket counters
CREATE TABLE IF NOT EXISTS project_counters (
    project_id INTEGER PRIMARY KEY,
    next_ticket_number INTEGER NOT NULL DEFAULT 1,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

-- Columns (linked list structure)
CREATE TABLE IF NOT EXISTS columns (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    project_id INTEGER NOT NULL,
    position INTEGER NOT NULL,
    prev_id INTEGER,
    next_id INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    FOREIGN KEY (prev_id) REFERENCES columns(id) ON DELETE SET NULL,
    FOREIGN KEY (next_id) REFERENCES columns(id) ON DELETE SET NULL
);

-- Tasks
CREATE TABLE IF NOT EXISTS tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    description TEXT,
    column_id INTEGER NOT NULL,
    position INTEGER NOT NULL,
    ticket_number INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (column_id) REFERENCES columns(id) ON DELETE CASCADE
);

-- Labels
CREATE TABLE IF NOT EXISTS labels (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    color TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

-- Task-Label junction
CREATE TABLE IF NOT EXISTS task_labels (
    task_id INTEGER NOT NULL,
    label_id INTEGER NOT NULL,
    PRIMARY KEY (task_id, label_id),
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
    FOREIGN KEY (label_id) REFERENCES labels(id) ON DELETE CASCADE
);

-- Priorities
CREATE TABLE IF NOT EXISTS priorities (
    id INTEGER PRIMARY KEY,
    description TEXT NOT NULL UNIQUE
);

-- Task priorities (one-to-one)
CREATE TABLE IF NOT EXISTS task_priorities (
    task_id INTEGER PRIMARY KEY,
    priority_id INTEGER NOT NULL,
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
    FOREIGN KEY (priority_id) REFERENCES priorities(id)
);

-- Types
CREATE TABLE IF NOT EXISTS types (
    id INTEGER PRIMARY KEY,
    description TEXT NOT NULL UNIQUE
);

-- Task types (one-to-one)
CREATE TABLE IF NOT EXISTS task_types (
    task_id INTEGER PRIMARY KEY,
    type_id INTEGER NOT NULL,
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
    FOREIGN KEY (type_id) REFERENCES types(id)
);

-- Task relations
CREATE TABLE IF NOT EXISTS task_relations (
    task_id INTEGER NOT NULL,
    related_task_id INTEGER NOT NULL,
    relation_type TEXT NOT NULL,
    PRIMARY KEY (task_id, related_task_id, relation_type),
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
    FOREIGN KEY (related_task_id) REFERENCES tasks(id) ON DELETE CASCADE
);

-- Indexes
CREATE INDEX idx_tasks_column_id ON tasks(column_id);
CREATE INDEX idx_columns_project_id ON columns(project_id);
CREATE INDEX idx_task_labels_task_id ON task_labels(task_id);
CREATE INDEX idx_task_labels_label_id ON task_labels(label_id);
CREATE INDEX idx_labels_project_id ON labels(project_id);
CREATE INDEX idx_task_priorities_task_id ON task_priorities(task_id);
CREATE INDEX idx_task_types_task_id ON task_types(task_id);
CREATE INDEX idx_task_relations_task_id ON task_relations(task_id);
CREATE INDEX idx_task_relations_related_task_id ON task_relations(related_task_id);

-- Insert default priorities
INSERT OR IGNORE INTO priorities (id, description) VALUES 
    (1, 'Critical'),
    (2, 'High'),
    (3, 'Medium'),
    (4, 'Low');

-- Insert default types
INSERT OR IGNORE INTO types (id, description) VALUES 
    (1, 'Task'),
    (2, 'Bug'),
    (3, 'Feature');
```

### Step 2.3: Write SQL Queries for Tasks (Days 4-5)

**File**: `internal/database/sql/tasks.sql`

**Tasks:**
1. Write SQLC queries for all task operations
2. Use SQLC annotations: `-- name: QueryName :returnType`
3. Cover all operations from `task_repository.go`

**Example Query File:**

```sql
-- internal/database/sql/tasks.sql

-- name: CreateTask :one
INSERT INTO tasks (
    title, 
    description, 
    column_id, 
    position, 
    ticket_number
) VALUES (
    ?, ?, ?, ?, ?
) RETURNING *;

-- name: GetTaskByID :one
SELECT * FROM tasks
WHERE id = ? LIMIT 1;

-- name: GetTasksByColumn :many
SELECT * FROM tasks
WHERE column_id = ?
ORDER BY position ASC;

-- name: UpdateTask :exec
UPDATE tasks
SET 
    title = ?,
    description = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: DeleteTask :exec
DELETE FROM tasks WHERE id = ?;

-- name: MoveTaskToColumn :exec
UPDATE tasks
SET 
    column_id = ?, 
    position = ?, 
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: SwapTaskPositions :exec
UPDATE tasks
SET 
    position = CASE
        WHEN id = ? THEN ?
        WHEN id = ? THEN ?
    END,
    updated_at = CURRENT_TIMESTAMP
WHERE id IN (?, ?);

-- name: GetTaskSummariesByProject :many
SELECT 
    t.id,
    t.title,
    t.ticket_number,
    t.column_id,
    t.position,
    t.created_at,
    p.description as priority_description,
    tp.description as type_description,
    GROUP_CONCAT(l.id) as label_ids,
    GROUP_CONCAT(l.name) as label_names,
    GROUP_CONCAT(l.color) as label_colors
FROM tasks t
INNER JOIN columns c ON t.column_id = c.id
LEFT JOIN task_priorities tp_link ON t.id = tp_link.task_id
LEFT JOIN priorities p ON tp_link.priority_id = p.id
LEFT JOIN task_types tt_link ON t.id = tt_link.task_id
LEFT JOIN types tp ON tt_link.type_id = tp.id
LEFT JOIN task_labels tl ON t.id = tl.task_id
LEFT JOIN labels l ON tl.label_id = l.id
WHERE c.project_id = ?
GROUP BY t.id
ORDER BY t.position ASC;

-- name: GetTaskDetail :one
SELECT 
    t.id,
    t.title,
    t.description,
    t.ticket_number,
    t.column_id,
    t.position,
    t.created_at,
    t.updated_at,
    p.id as priority_id,
    p.description as priority_description,
    tp.id as type_id,
    tp.description as type_description
FROM tasks t
LEFT JOIN task_priorities tp_link ON t.id = tp_link.task_id
LEFT JOIN priorities p ON tp_link.priority_id = p.id
LEFT JOIN task_types tt_link ON t.id = tt_link.task_id
LEFT JOIN types tp ON tt_link.type_id = tp.id
WHERE t.id = ?;

-- name: GetNextTicketNumber :one
SELECT next_ticket_number 
FROM project_counters 
WHERE project_id = ?;

-- name: IncrementTicketNumber :exec
UPDATE project_counters 
SET next_ticket_number = next_ticket_number + 1 
WHERE project_id = ?;

-- name: GetTaskReferencesForProject :many
SELECT 
    t.id,
    t.title,
    t.ticket_number
FROM tasks t
INNER JOIN columns c ON t.column_id = c.id
WHERE c.project_id = ?
ORDER BY t.ticket_number ASC;

-- name: SetTaskPriority :exec
INSERT OR REPLACE INTO task_priorities (task_id, priority_id)
VALUES (?, ?);

-- name: SetTaskType :exec
INSERT OR REPLACE INTO task_types (task_id, type_id)
VALUES (?, ?);

-- name: AttachLabel :exec
INSERT OR IGNORE INTO task_labels (task_id, label_id)
VALUES (?, ?);

-- name: DetachLabel :exec
DELETE FROM task_labels
WHERE task_id = ? AND label_id = ?;

-- name: AddTaskRelation :exec
INSERT OR IGNORE INTO task_relations (task_id, related_task_id, relation_type)
VALUES (?, ?, ?);

-- name: RemoveTaskRelation :exec
DELETE FROM task_relations
WHERE task_id = ? AND related_task_id = ? AND relation_type = ?;

-- name: GetTaskParents :many
SELECT 
    t.id,
    t.title,
    t.ticket_number
FROM task_relations tr
INNER JOIN tasks t ON tr.related_task_id = t.id
WHERE tr.task_id = ? AND tr.relation_type = 'parent';

-- name: GetTaskChildren :many
SELECT 
    t.id,
    t.title,
    t.ticket_number
FROM task_relations tr
INNER JOIN tasks t ON tr.related_task_id = t.id
WHERE tr.task_id = ? AND tr.relation_type = 'child';

-- name: GetTaskBlocks :many
SELECT 
    t.id,
    t.title,
    t.ticket_number
FROM task_relations tr
INNER JOIN tasks t ON tr.related_task_id = t.id
WHERE tr.task_id = ? AND tr.relation_type = 'blocks';

-- name: GetTaskBlockedBy :many
SELECT 
    t.id,
    t.title,
    t.ticket_number
FROM task_relations tr
INNER JOIN tasks t ON tr.related_task_id = t.id
WHERE tr.task_id = ? AND tr.relation_type = 'blocked_by';
```

### Step 2.4: Generate SQLC Code (Day 6)

**Tasks:**
1. Run `sqlc generate`
2. Verify generated code in `internal/database/generated/`
3. Review generated interfaces

**Generated Files:**
- `db.go` - Database helpers
- `models.go` - Go structs for each table
- `querier.go` - Interface with all query methods
- `tasks.sql.go` - Generated task query functions

**Example Generated Interface:**

```go
// internal/database/generated/querier.go
type Querier interface {
    CreateTask(ctx context.Context, arg CreateTaskParams) (Task, error)
    GetTaskByID(ctx context.Context, id int64) (Task, error)
    GetTasksByColumn(ctx context.Context, columnID int64) ([]Task, error)
    UpdateTask(ctx context.Context, arg UpdateTaskParams) error
    DeleteTask(ctx context.Context, id int64) error
    // ... all other queries
}
```

### Step 2.5: Update Task Repository (Days 7-8)

**File**: `internal/database/task_repository.go`

**Tasks:**
1. Backup old file: `cp task_repository.go task_repository_old.go`
2. Rewrite repository to use SQLC-generated code
3. Keep event publishing logic
4. Add helper functions to convert between SQLC models and domain models

**Example:**

```go
package database

import (
    "context"
    "database/sql"
    "fmt"
    
    "github.com/thenoetrevino/paso/internal/database/generated"
    "github.com/thenoetrevino/paso/internal/events"
    "github.com/thenoetrevino/paso/internal/models"
)

type TaskRepo struct {
    db          *sql.DB
    queries     *generated.Queries  // SQLC-generated
    eventClient events.EventPublisher
}

func NewTaskRepo(db *sql.DB, eventClient events.EventPublisher) *TaskRepo {
    return &TaskRepo{
        db:          db,
        queries:     generated.New(db),
        eventClient: eventClient,
    }
}

func (r *TaskRepo) CreateTask(ctx context.Context, title, description string, columnID, position int) (*models.Task, error) {
    projectID, err := getProjectIDFromTable(ctx, r.db, "columns", columnID)
    if err != nil {
        return nil, err
    }
    
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer tx.Rollback()
    
    qtx := r.queries.WithTx(tx)
    
    // Get next ticket number (SQLC-generated)
    ticketNumber, err := qtx.GetNextTicketNumber(ctx, int64(projectID))
    if err != nil {
        return nil, fmt.Errorf("failed to get ticket number: %w", err)
    }
    
    // Increment counter (SQLC-generated)
    if err := qtx.IncrementTicketNumber(ctx, int64(projectID)); err != nil {
        return nil, fmt.Errorf("failed to increment ticket counter: %w", err)
    }
    
    // Create task (SQLC-generated)
    dbTask, err := qtx.CreateTask(ctx, generated.CreateTaskParams{
        Title:        title,
        Description:  sql.NullString{String: description, Valid: description != ""},
        ColumnID:     int64(columnID),
        Position:     int64(position),
        TicketNumber: ticketNumber,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create task: %w", err)
    }
    
    if err := tx.Commit(); err != nil {
        return nil, fmt.Errorf("failed to commit: %w", err)
    }
    
    sendEvent(r.eventClient, projectID)
    
    return toTaskModel(dbTask), nil
}

// Convert SQLC model to domain model
func toTaskModel(dbTask generated.Task) *models.Task {
    return &models.Task{
        ID:           int(dbTask.ID),
        Title:        dbTask.Title,
        Description:  dbTask.Description.String,
        ColumnID:     int(dbTask.ColumnID),
        Position:     int(dbTask.Position),
        TicketNumber: int(dbTask.TicketNumber),
        CreatedAt:    dbTask.CreatedAt.Time,
        UpdatedAt:    dbTask.UpdatedAt.Time,
    }
}

func (r *TaskRepo) GetTaskDetail(ctx context.Context, taskID int) (*models.TaskDetail, error) {
    row, err := r.queries.GetTaskDetail(ctx, int64(taskID))
    if err != nil {
        return nil, err
    }
    
    return &models.TaskDetail{
        ID:                  int(row.ID),
        Title:               row.Title,
        Description:         row.Description.String,
        TicketNumber:        int(row.TicketNumber),
        ColumnID:            int(row.ColumnID),
        Position:            int(row.Position),
        PriorityDescription: row.PriorityDescription.String,
        TypeDescription:     row.TypeDescription.String,
        CreatedAt:           row.CreatedAt.Time,
        UpdatedAt:           row.UpdatedAt.Time,
    }, nil
}

func (r *TaskRepo) SetTaskPriority(ctx context.Context, taskID, priorityID int) error {
    return r.queries.SetTaskPriority(ctx, generated.SetTaskPriorityParams{
        TaskID:     int64(taskID),
        PriorityID: int64(priorityID),
    })
}

func (r *TaskRepo) AttachLabel(ctx context.Context, taskID, labelID int) error {
    return r.queries.AttachLabel(ctx, generated.AttachLabelParams{
        TaskID:  int64(taskID),
        LabelID: int64(labelID),
    })
}

// ... implement other methods using SQLC-generated queries
```

### Step 2.6: Test Task Repository (Days 9-10)

**Tasks:**
1. Run existing repository tests
2. Fix any issues
3. Add new tests if needed
4. Verify all task operations work
5. Compare performance with old implementation

**Example:**
```bash
# Run tests
go test ./internal/database -v -run TestTask

# Benchmark
go test ./internal/database -bench=BenchmarkTask -benchmem
```

### Step 2.7: Migrate Other Repositories (Days 11-13)

**Tasks:**

**Day 11 - Columns:**
1. Write SQL queries in `sql/columns.sql`
2. Regenerate SQLC code: `sqlc generate`
3. Update `column_repository.go` to use generated code
4. Test column repository

**Day 12 - Projects:**
1. Write SQL queries in `sql/projects.sql`
2. Regenerate SQLC code
3. Update `project_repository.go` to use generated code
4. Test project repository

**Day 13 - Labels:**
1. Write SQL queries in `sql/labels.sql`
2. Regenerate SQLC code
3. Update `label_repository.go` to use generated code
4. Test label repository

### Step 2.8: Cleanup and Integration Testing (Day 14)

**Tasks:**
1. Delete old `*_old.go` files
2. Remove unused helper functions
3. Update `DataStore` interface if needed
4. Run full integration test suite
5. Performance testing
6. Update documentation

### Phase 2 Deliverables

- [ ] SQLC configured and working
- [ ] All migrations in SQL files
- [ ] All queries in SQL files
- [ ] All repositories use SQLC-generated code
- [ ] 70% code reduction achieved
- [ ] All tests passing
- [ ] Performance maintained or improved

### Phase 2 Success Criteria

- ✅ SQLC generates valid Go code
- ✅ All repository tests pass
- ✅ No performance regressions
- ✅ Code is more readable (SQL separate from Go)
- ✅ Can add new queries in < 5 minutes

---

## Phase 3: Infrastructure Improvements

**Duration**: 1 week  
**Priority**: MEDIUM  
**Dependencies**: Phases 1 & 2

### Overview

Improve infrastructure, organization, and developer experience.

### Step 3.1: Extract main.go (Day 1)

**Tasks:**
1. Create `internal/cmd/root.go`
2. Move `rootCmd` definition to `root.go`
3. Move command registration to `root.go`
4. Keep `main.go` minimal (< 30 lines)

**Before** (`main.go` - 121 lines):
```go
package main

var rootCmd = &cobra.Command{...}

func main() {
    if err := rootCmd.Execute(); err != nil {...}
}

func init() {
    rootCmd.AddCommand(task.TaskCmd())
    rootCmd.AddCommand(project.ProjectCmd())
    // ... lots of setup
}
```

**After** (`main.go` - ~20 lines):
```go
package main

import (
    "fmt"
    "os"
    "github.com/thenoetrevino/paso/internal/cmd"
)

var (
    version = "dev"
    commit  = "none"
    date    = "unknown"
)

func main() {
    if err := cmd.Execute(version, commit, date); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

**New** (`internal/cmd/root.go`):
```go
package cmd

import (
    "github.com/spf13/cobra"
    "github.com/thenoetrevino/paso/internal/cli/column"
    "github.com/thenoetrevino/paso/internal/cli/label"
    "github.com/thenoetrevino/paso/internal/cli/project"
    "github.com/thenoetrevino/paso/internal/cli/task"
    "github.com/thenoetrevino/paso/internal/launcher"
)

var rootCmd = &cobra.Command{
    Use:   "paso",
    Short: "Terminal-based Kanban board with CLI and TUI",
    Long: `Paso is a zero-setup, terminal-based kanban board for personal task management.

Use 'paso tui' to launch the interactive TUI.
Use 'paso task create ...' for CLI commands.`,
}

func Execute(version, commit, date string) error {
    rootCmd.Version = version
    rootCmd.SetVersionTemplate(fmt.Sprintf("paso version %s\n  commit: %s\n  built: %s\n", version, commit, date))
    
    return rootCmd.Execute()
}

func init() {
    // Add CLI subcommands
    rootCmd.AddCommand(task.TaskCmd())
    rootCmd.AddCommand(project.ProjectCmd())
    rootCmd.AddCommand(column.ColumnCmd())
    rootCmd.AddCommand(label.LabelCmd())

    // Add TUI subcommand
    tuiCmd := &cobra.Command{
        Use:   "tui",
        Short: "Launch the interactive TUI",
        Long:  "Launch the interactive terminal user interface for managing tasks visually.",
        RunE: func(cmd *cobra.Command, args []string) error {
            return launcher.Launch()
        },
    }
    rootCmd.AddCommand(tuiCmd)

    // Add completion command
    completionCmd := &cobra.Command{
        Use:   "completion [bash|zsh|fish|powershell]",
        Short: "Generate shell completion script",
        // ... completion logic
    }
    rootCmd.AddCommand(completionCmd)
}
```

### Step 3.2: Add Package Documentation (Days 2-3)

**Tasks:**
1. Add `doc.go` to complex packages
2. Document service layer patterns
3. Document database layer patterns

**Example** (`internal/services/doc.go`):
```go
// Package services provides the business logic layer for Paso.
//
// Services sit between the UI layer (TUI/CLI) and the data access layer (repositories).
// They encapsulate business rules, validation, and orchestrate complex operations.
//
// Architecture:
//
//     TUI/CLI → Services → Repositories → Database
//
// Each domain (task, project, column, label) has its own service package.
//
// Services are designed to be testable without a database by mocking repositories.
//
// Example usage:
//
//     svc := task.NewService(repo, eventClient)
//     task, err := svc.CreateTask(ctx, task.CreateTaskRequest{
//         Title:    "New Task",
//         ColumnID: 1,
//     })
package services
```

**Example** (`internal/database/doc.go`):
```go
// Package database provides data access layer for Paso using SQLC-generated code.
//
// All SQL queries are defined in internal/database/sql/*.sql files and
// SQLC generates type-safe Go code in internal/database/generated/.
//
// Repositories wrap the generated code and add:
// - Event publishing for live updates
// - Helper functions for complex operations
// - Domain model conversions
//
// Architecture:
//
//     Repositories → SQLC-generated queries → SQLite database
//
// To add a new query:
// 1. Write SQL in sql/*.sql with SQLC annotation
// 2. Run `sqlc generate`
// 3. Use generated function in repository
package database
```

### Step 3.3: Organize Documentation (Day 4)

**Tasks:**
1. Create `docs/` directory
2. Move architecture docs to `docs/`
3. Keep README, QUICK_START, and RELEASE_GUIDE at root
4. Update references

**Structure:**
```
docs/
├── architecture/
│   ├── SERVICE_LAYER.md
│   ├── DATABASE.md
│   └── EVENTS.md
├── development/
│   ├── TESTING_GUIDE.md
│   ├── CODING_CONVENTIONS.md
│   └── REFACTORING_PLAN.md (this file)
└── implementation/
    ├── CLI_IMPLEMENTATION_GUIDE.md
    ├── LIVE_UPDATES_IMPLEMENTATION_PLAN.md
    └── IMPLEMENTATION_SUMMARY.md
```

### Step 3.4: Add Event/PubSub System (Days 5-7)

**Optional but recommended for future scalability**

**Tasks:**
1. Create `internal/pubsub/` package (similar to Crush)
2. Implement `Broker[T]` for type-safe event publishing
3. Update services to use pubsub for internal events
4. Decouple TUI components with events

**Benefits:**
- Better component decoupling
- Easier to add features like live updates
- More testable (can observe events in tests)

**Example:**
```go
// internal/pubsub/broker.go
package pubsub

type EventType string

const (
    CreatedEvent EventType = "created"
    UpdatedEvent EventType = "updated"
    DeletedEvent EventType = "deleted"
)

type Broker[T any] struct {
    subscribers map[EventType][]chan T
    mu          sync.RWMutex
}

func NewBroker[T any]() *Broker[T] {
    return &Broker[T]{
        subscribers: make(map[EventType][]chan T),
    }
}

func (b *Broker[T]) Subscribe(eventType EventType) <-chan T {
    b.mu.Lock()
    defer b.mu.Unlock()
    
    ch := make(chan T, 10)
    b.subscribers[eventType] = append(b.subscribers[eventType], ch)
    return ch
}

func (b *Broker[T]) Publish(eventType EventType, data T) {
    b.mu.RLock()
    defer b.mu.RUnlock()
    
    for _, ch := range b.subscribers[eventType] {
        select {
        case ch <- data:
        default:
            // Channel full, skip
        }
    }
}
```

### Phase 3 Deliverables

- [ ] Minimal main.go (< 30 lines)
- [ ] Package documentation for services and database
- [ ] Organized docs directory
- [ ] (Optional) PubSub system implemented

### Phase 3 Success Criteria

- ✅ Clean entry point
- ✅ Well-documented architecture
- ✅ Easy for new developers to understand codebase

---

## Phase 4: Documentation & Polish

**Duration**: 1 week  
**Priority**: LOW  
**Dependencies**: All previous phases

### Overview

Final documentation, performance tuning, and polish.

### Step 4.1: Architecture Documentation (Days 1-2)

**Tasks:**
1. Write `docs/architecture/SERVICE_LAYER.md`
   - Service layer design
   - How to add a new service
   - Testing strategies
2. Write `docs/architecture/DATABASE.md`
   - SQLC usage guide
   - How to add queries
   - Migration guide
3. Write `docs/architecture/OVERVIEW.md`
   - System architecture diagram
   - Component responsibilities
   - Data flow

### Step 4.2: Developer Guides (Day 3)

**Tasks:**
1. Update `CONTRIBUTING.md` (if exists)
2. Add "Adding a Feature" guide
3. Add "Database Queries" guide
4. Add "Testing" guide

### Step 4.3: Performance Testing (Days 4-5)

**Tasks:**
1. Benchmark critical operations:
   - Task creation
   - Task listing
   - Task updates
   - TUI startup
2. Profile memory usage
3. Profile CPU usage
4. Optimize if needed

**Tools:**
```bash
# Benchmark
go test -bench=. -benchmem ./...

# CPU profiling
go test -cpuprofile=cpu.prof -bench=.
go tool pprof cpu.prof

# Memory profiling
go test -memprofile=mem.prof -bench=.
go tool pprof mem.prof
```

### Step 4.4: Database Optimization (Day 6)

**Tasks:**
1. Analyze slow queries
2. Add indexes where needed
3. Configure connection pooling:

```go
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(5 * time.Minute)
```

### Step 4.5: Final Review & Cleanup (Day 7)

**Tasks:**
1. Code review all changes
2. Remove dead code
3. Update all documentation
4. Create migration guide for users (if breaking changes)
5. Tag release

### Phase 4 Deliverables

- [ ] Architecture documentation complete
- [ ] Developer guides updated
- [ ] Performance benchmarks documented
- [ ] Database optimized
- [ ] Clean, production-ready codebase

### Phase 4 Success Criteria

- ✅ All documentation up-to-date
- ✅ Performance meets or exceeds baseline
- ✅ Ready for production use

---

## Timeline & Milestones

### Week 1-2: Phase 1 - Service Layer
- **Week 1**: Task service implementation + tests
- **Week 2**: Update TUI/CLI + other services

**Milestone**: Service layer operational, all tests pass

### Week 3-4: Phase 2 - SQLC Migration
- **Week 3**: Setup SQLC, migrate task repository
- **Week 4**: Migrate other repositories, testing

**Milestone**: All repositories use SQLC, 70% code reduction achieved

### Week 5: Phase 3 - Infrastructure
- Infrastructure improvements
- Documentation organization
- Optional: PubSub system

**Milestone**: Clean codebase structure

### Week 6: Phase 4 - Polish
- Documentation
- Performance tuning
- Final review

**Milestone**: Production-ready release

---

## Risk Mitigation

### Risk 1: Service Layer Breaks Existing Functionality

**Mitigation:**
- Implement comprehensive tests before refactoring
- Keep repository layer unchanged during service layer addition
- Test each service method thoroughly
- Can run old and new code paths in parallel during transition

### Risk 2: SQLC Generated Code Doesn't Match Expectations

**Mitigation:**
- Start with task repository only
- Review generated code before proceeding
- Keep old repository as backup (`*_old.go`)
- Can rollback if issues found
- SQLC is mature and widely used

### Risk 3: Performance Regressions

**Mitigation:**
- Benchmark before and after each phase
- Profile critical operations
- SQLC is generally faster than hand-written code
- Service layer adds minimal overhead

### Risk 4: Testing Gaps

**Mitigation:**
- Write tests as we go, not after
- Maintain or improve test coverage
- Integration tests catch breaking changes
- Service layer actually improves testability

### Risk 5: Timeline Overruns

**Mitigation:**
- Break into phases with clear deliverables
- Each phase is independently useful
- Can pause between phases
- Phases can be extended if needed

---

## References

### Inspiration: Crush Codebase Analysis

**Location**: `/home/noetrevino/projects/paso/feature/crush/`

**Key Learnings:**
- Service layer pattern: `internal/session/`, `internal/message/`
- SQLC usage: `internal/db/` with generated code
- App struct for DI: `internal/app/app.go`
- Package documentation: `doc.go` files

### Tools & Libraries

- **SQLC**: https://sqlc.dev/ - SQL code generator
- **Testify**: https://github.com/stretchr/testify - Testing toolkit
- **Cobra**: https://github.com/spf13/cobra - CLI framework (already in use)

### Documentation

- **SQLC Tutorial**: https://docs.sqlc.dev/en/stable/tutorials/getting-started-sqlite.html
- **Go Project Layout**: https://github.com/golang-standards/project-layout
- **Service Layer Pattern**: Martin Fowler's P of EAA

---

## Appendix A: Code Examples

### Service Layer Example

See Phase 1, Step 1.1 for complete service interface example.

### SQLC Query Example

See Phase 2, Step 2.3 for complete query file example.

### App Struct Example

See Phase 1, Step 1.4 for complete app struct example.

---

## Appendix B: Testing Strategy

### Unit Tests (Service Layer)

```go
// Mock repository
type MockDataStore struct {
    mock.Mock
}

func (m *MockDataStore) CreateTask(ctx context.Context, title, description string, columnID, position int) (*models.Task, error) {
    args := m.Called(ctx, title, description, columnID, position)
    return args.Get(0).(*models.Task), args.Error(1)
}

// Test
func TestTaskService_CreateTask(t *testing.T) {
    mockRepo := new(MockDataStore)
    svc := task.NewService(mockRepo, nil)
    
    mockRepo.On("CreateTask", mock.Anything, "Test", "", 1, 0).
        Return(&models.Task{ID: 1}, nil)
    
    result, err := svc.CreateTask(context.Background(), task.CreateTaskRequest{
        Title:    "Test",
        ColumnID: 1,
    })
    
    assert.NoError(t, err)
    assert.Equal(t, 1, result.ID)
    mockRepo.AssertExpectations(t)
}
```

### Integration Tests (Repository Layer)

```go
func TestTaskRepository_CreateTask(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    
    repo := database.NewTaskRepo(db, nil)
    
    task, err := repo.CreateTask(context.Background(), "Test Task", "Description", 1, 0)
    
    assert.NoError(t, err)
    assert.NotNil(t, task)
    assert.Greater(t, task.ID, 0)
}
```

### Benchmark Example

```go
func BenchmarkTaskService_CreateTask(b *testing.B) {
    db := setupTestDB(b)
    defer db.Close()
    
    repo := database.NewDataStore(db)
    svc := task.NewService(repo, nil)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = svc.CreateTask(context.Background(), task.CreateTaskRequest{
            Title:    fmt.Sprintf("Task %d", i),
            ColumnID: 1,
        })
    }
}
```

---

## Appendix C: Migration Checklist

Use this checklist to track progress through each phase.

### Phase 1: Service Layer
- [ ] Service interfaces defined (task, project, column, label)
- [ ] Task service implemented
- [ ] Task service tests written (80%+ coverage)
- [ ] Project service implemented
- [ ] Column service implemented
- [ ] Label service implemented
- [ ] App struct created with DI
- [ ] TUI updated to use services
- [ ] CLI updated to use services
- [ ] All tests passing
- [ ] Integration tests added
- [ ] Documentation written

### Phase 2: SQLC Migration
- [ ] SQLC installed and configured
- [ ] sqlc.yaml created
- [ ] Directory structure created
- [ ] Migrations extracted to SQL files
- [ ] Task queries written (sql/tasks.sql)
- [ ] Column queries written (sql/columns.sql)
- [ ] Project queries written (sql/projects.sql)
- [ ] Label queries written (sql/labels.sql)
- [ ] SQLC code generated successfully
- [ ] Task repository updated to use SQLC
- [ ] Column repository updated to use SQLC
- [ ] Project repository updated to use SQLC
- [ ] Label repository updated to use SQLC
- [ ] All repository tests passing
- [ ] Old code removed
- [ ] Performance benchmarks run
- [ ] Documentation updated

### Phase 3: Infrastructure
- [ ] main.go extracted to cmd/root.go
- [ ] Package documentation added (doc.go files)
- [ ] Documentation organized in docs/ directory
- [ ] (Optional) PubSub system implemented

### Phase 4: Documentation & Polish
- [ ] Architecture documentation written
- [ ] Developer guides updated
- [ ] Performance testing completed
- [ ] Database optimized
- [ ] Code review completed
- [ ] Release prepared

---

**End of Refactoring Plan**

*For questions or updates to this plan, contact the development team.*
