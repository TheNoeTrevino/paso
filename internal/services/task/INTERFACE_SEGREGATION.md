# Task Service Interface Segregation

## Overview

The original `Service` interface had 24 methods covering disparate concerns, violating the Interface Segregation Principle (ISP). This refactoring splits the monolithic interface into 6 focused, segregated interfaces while maintaining backward compatibility.

## Problem Statement

The original interface forced clients to depend on all 24 methods even when they only needed a subset:

```go
// OLD: 24 methods in one interface
type Service interface {
    GetTaskDetail()
    GetTaskSummariesByProject()
    // ... 22 more methods
}
```

Issues:
- **Testing burden**: Mocking required implementing all 24 methods
- **Coupling**: Components depended on capabilities they didn't use
- **Readability**: Hard to understand what a client actually needed
- **Evolution**: Adding new methods affected all implementations

## Solution: Interface Segregation

Split into 6 focused interfaces, each with a single responsibility:

### 1. TaskReader (7 methods)
**Purpose**: Read-only operations for retrieving task data

Contains all GET operations:
- `GetTaskDetail()` - Get full task details
- `GetTaskSummariesByProject()` - Get task list grouped by column
- `GetTaskSummariesByProjectFiltered()` - Get filtered tasks
- `GetReadyTaskSummariesByProject()` - Get ready tasks
- `GetInProgressTasksByProject()` - Get in-progress tasks
- `GetTaskReferencesForProject()` - Get task references
- `GetTaskTreeByProject()` - Get task hierarchy

**Usage**: Components that only need to read task information

### 2. TaskWriter (3 methods)
**Purpose**: CRUD operations for task lifecycle

- `CreateTask()` - Create new task
- `UpdateTask()` - Update task fields
- `DeleteTask()` - Delete task

**Usage**: Task creation/modification components

### 3. TaskMover (8 methods)
**Purpose**: Task movement operations within the workflow

- `MoveTaskToNextColumn()` - Advance to next column
- `MoveTaskToPrevColumn()` - Move back to previous column
- `MoveTaskToColumn()` - Move to specific column
- `MoveTaskToReadyColumn()` - Move to ready state
- `MoveTaskToCompletedColumn()` - Mark as completed
- `MoveTaskToInProgressColumn()` - Start working on task
- `MoveTaskUp()` - Reorder within column
- `MoveTaskDown()` - Reorder within column

**Usage**: Workflow/kanban board components

### 4. TaskRelationer (4 methods)
**Purpose**: Task relationship management (dependencies, blocking)

- `AddParentRelation()` - Add dependency
- `AddChildRelation()` - Add dependent task
- `RemoveParentRelation()` - Remove dependency
- `RemoveChildRelation()` - Remove dependent task

**Usage**: Dependency graph management components

### 5. TaskLabeler (2 methods)
**Purpose**: Task label/tag management

- `AttachLabel()` - Add label to task
- `DetachLabel()` - Remove label from task

**Usage**: Label management components

### 6. TaskCommenter (4 methods)
**Purpose**: Task comment operations

- `CreateComment()` - Add comment to task
- `UpdateComment()` - Modify comment
- `DeleteComment()` - Remove comment
- `GetCommentsByTask()` - Retrieve comments

**Usage**: Comment/discussion components

## Composition Interface

The original `Service` interface is preserved as a **composition** of all segregated interfaces:

```go
type Service interface {
    TaskReader
    TaskWriter
    TaskMover
    TaskRelationer
    TaskLabeler
    TaskCommenter
}
```

**Benefits**:
- Backward compatible with existing code using `Service`
- The concrete `service` struct implements all interfaces automatically
- Clients can depend on specific interfaces for better design
- No breaking changes to existing code

## Migration Path

### Current Code (Still Works)
```go
// Old way - still valid
var s task.Service = task.NewService(db, eventClient)
tasks, err := s.GetTaskSummariesByProject(ctx, projectID)
```

### Recommended (New Way)
```go
// Better: Depend on what you actually need
var reader task.TaskReader = task.NewService(db, eventClient)
tasks, err := reader.GetTaskSummariesByProject(ctx, projectID)
```

## Benefits of This Refactoring

1. **Easier Testing**: Mock only required methods
   ```go
   type mockReader struct {}
   func (m *mockReader) GetTaskDetail(...) {...}
   // No need to implement 20 other methods
   ```

2. **Better Documentation**: Interface name clarifies intent
   ```go
   func (h *Handler) renderBoard(ctx context.Context, reader task.TaskReader) error {
       // Clear: handler only reads tasks
   }
   ```

3. **Reduced Coupling**: Components depend on minimal interface
   ```go
   // Only what's needed
   type Commenter interface {
       CreateComment(ctx context.Context, req CreateCommentRequest) (*models.Comment, error)
   }
   ```

4. **Clearer Responsibilities**: Each interface has one job
   - TaskReader: Retrieval
   - TaskWriter: CRUD
   - TaskMover: Workflow
   - TaskRelationer: Dependencies
   - TaskLabeler: Labels
   - TaskCommenter: Comments

5. **Future Evolution**: Easy to split further
   - Can create `TaskTreeReader` if tree operations grow
   - Can create `TaskSummaryReader` if summary operations diverge
   - No impact on existing implementations

## Statistics

| Metric | Before | After |
|--------|--------|-------|
| Max methods per interface | 24 | 8 (TaskMover) |
| Min methods per interface | 24 | 2 (TaskLabeler) |
| Avg methods per interface | 24 | 4 |
| Number of concerns | 1 | 6 |
| Cohesion | Low | High |

## Implementation Notes

- The `service` struct implementation unchanged
- All methods remain in the same file
- Tests all pass without modification
- No breaking changes to public API
- Composition maintains full backward compatibility

## ISP Principle Compliance

This refactoring follows Robert C. Martin's Interface Segregation Principle:
> "Clients should not be forced to depend upon interfaces they do not use."

Each interface now represents a specific client concern, with no unused methods.
