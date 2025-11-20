# Paso - Implementation Handoff Document

## Project Status: Phases 1-7 Complete ✅

Paso is a terminal-based kanban board with SQLite persistence and linked-list column architecture.

## What's Been Implemented

### Phase 1-3: Foundation (Complete)
- SQLite database with `modernc.org/sqlite` (pure Go, no CGo)
- Column and Task models
- Bubble Tea TUI with MVU pattern
- Lipgloss styling and kanban board layout

### Phase 4: Navigation (Complete)
- hjkl/arrow key navigation between columns and tasks
- Visual selection indicators (purple highlighting)
- Boundary handling for first/last items

### Phase 5: Task CRUD (Complete)
- `a` - Add task (input dialog)
- `e` - Edit task (pre-filled dialog)
- `d` - Delete task (confirmation dialog)
- All operations persist to SQLite immediately

### Phase 6: Viewport Scrolling (Complete)
- `[` and `]` keys for horizontal scrolling
- Auto-scroll when navigating keeps selection visible
- Scroll indicators (◀ ▶) show hidden columns

### Phase 6.5: Column CRUD (Complete) ⭐
- **Linked List Architecture**: Columns use `prev_id` and `next_id` pointers instead of position
- `C` - Create column (inserts after current, green dialog)
- `R` - Rename column (blue dialog)
- `X` - Delete column (red dialog with task count warning)
- Auto-migration from position-based to linked list

### Phase 7: Task Movement (Complete)
- `<` - Move task to previous column (uses linked list prev pointer)
- `>` - Move task to next column (uses linked list next pointer)
- Selection follows moved task
- Database functions: `MoveTaskToNextColumn()`, `MoveTaskToPrevColumn()`

### Phase 8: Animations (Skipped)
- User decided against spring animations
- Instant column snapping preferred

## Technical Details

### Database Schema
```sql
-- Columns table (linked list)
CREATE TABLE columns (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    prev_id INTEGER,  -- NULL for head
    next_id INTEGER   -- NULL for tail
);

-- Tasks table
CREATE TABLE tasks (
    id INTEGER PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    column_id INTEGER,
    position INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(column_id) REFERENCES columns(id)
);
```

### Key Files
```
paso/
├── main.go                          # Entry point
├── internal/
│   ├── database/
│   │   ├── db.go                   # Database init (~/.paso/tasks.db)
│   │   ├── migrations.go           # Schema + linked list migration
│   │   ├── repository.go           # All CRUD operations
│   │   └── repository_test.go      # 9 linked list tests (all passing)
│   ├── models/
│   │   ├── column.go               # Column struct (ID, Name, PrevID, NextID)
│   │   └── task.go                 # Task struct
│   └── tui/
│       ├── model.go                # App state + MVU
│       ├── update.go               # Event handlers
│       ├── view.go                 # Rendering
│       ├── components.go           # RenderTask, RenderColumn
│       └── styles.go               # Lipgloss styles
```

### Current Controls
```
Tasks:
  a     - Add task
  e     - Edit task
  d     - Delete task
  <     - Move task left
  >     - Move task right

Columns:
  C     - Create column (after current)
  R     - Rename column
  X     - Delete column (with warning)

Navigation:
  hjkl  - Navigate
  [ ]   - Scroll viewport
  q     - Quit
```

## What's Left

### Phase 9: Data Persistence Tests
- Write integration tests for CRUD persistence
- Test reload behavior on startup
- Verify transaction safety
- Test linked list integrity after operations

### Phase 10: Polish & UX (Optional)
- Task counts in column headers (`Todo (3)`)
- Timestamps on recently updated tasks (`2h ago`)
- Help screen with `?` key
- Status bar
- Better error messages

## Testing

Run existing tests:
```bash
go test ./internal/database -v
# 9/9 tests passing (linked list operations)
```

Build and run:
```bash
go build -o paso
./paso
```

## Notes

- Database location: `~/.paso/tasks.db`
- Default columns auto-created: "Todo", "In Progress", "Done"
- Linked list migration runs automatically on first start
- All operations are transactional and persist immediately
- No CGo dependencies - compiles anywhere
