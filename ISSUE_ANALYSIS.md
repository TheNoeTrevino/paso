# Issue Analysis Report

This document analyzes three reported issues in the Paso task management system and provides detailed root cause analysis with line numbers and recommended fixes.

---

## Issue 1: SQLite BUSY Error

### Error Message
```
columns: querying columns for project: database is locked (5) (SQLITE_BUSY)
```

### Root Cause

**Location**: `internal/database/column_repository.go:126`

The error occurs when SQLite encounters a database lock. The root causes are:

1. **No WAL mode enabled**: The database is using SQLite's default journaling mode, which allows only one writer at a time.
   - **File**: `internal/database/db.go:28`
   - **Current code**:
     ```go
     db, err := sql.Open("sqlite", dbPath)
     ```
   - The SQLite connection is opened without enabling Write-Ahead Logging (WAL) mode, which would allow concurrent reads during writes.

2. **No busy timeout configured**: When a database is locked, SQLite immediately returns SQLITE_BUSY instead of retrying.
   - **File**: `internal/database/db.go:34` (only sets `PRAGMA foreign_keys = ON`)
   - Missing `PRAGMA busy_timeout` configuration.

3. **No connection pool limits**: The database allows unlimited concurrent connections, which can exacerbate locking issues.
   - **File**: `internal/database/db.go:28-58`
   - No `db.SetMaxOpenConns()` or `db.SetMaxIdleConns()` calls.

4. **Transaction overlap**: Multiple database operations (task movement, column queries) can run concurrently without proper synchronization, especially with the event-driven architecture where the TUI and daemon both access the database.

### Detailed Analysis

When moving a task:
1. `internal/database/task_repository.go:579-693` - `swapTask()` starts a transaction (line 580-584)
2. During this transaction, the database is locked for writing
3. If another goroutine tries to read columns (line 126 in `column_repository.go`), it encounters the lock
4. Without a busy timeout, SQLite immediately returns SQLITE_BUSY

### Fix Strategy

Enable WAL mode and configure busy timeout:

**File**: `internal/database/db.go`

Add after line 41 (after enabling foreign keys):
```go
// Enable WAL mode for better concurrency
_, err = db.ExecContext(ctx, "PRAGMA journal_mode = WAL")
if err != nil {
    log.Printf("Failed to enable WAL mode: %v", err)
    if closeErr := db.Close(); closeErr != nil {
        log.Printf("error closing db: %v", closeErr)
    }
    return nil, err
}

// Set busy timeout to 5 seconds (SQLite will retry for this duration)
_, err = db.ExecContext(ctx, "PRAGMA busy_timeout = 5000")
if err != nil {
    log.Printf("Failed to set busy timeout: %v", err)
    if closeErr := db.Close(); closeErr != nil {
        log.Printf("error closing db: %v", closeErr)
    }
    return nil, err
}
```

Add after line 48 (after ping check):
```go
// Configure connection pool to reduce contention
db.SetMaxOpenConns(1)  // SQLite benefits from a single writer connection
db.SetMaxIdleConns(1)
```

---

## Issue 2: Daemon Broken Pipe Error

### Error Message
```
23:46:56 Failed to send pong:
write unix @->/home/noetrevino/.paso/paso.sock: write: broken pipe
```

### Root Cause

**Location**: `internal/events/client.go:274`

The error occurs when the client tries to respond to a daemon ping after the connection has been closed.

### Detailed Analysis

The sequence of events:
1. **Daemon sends ping**: `internal/daemon/server.go:289-309` - Every 30 seconds, daemon sends ping to all clients
2. **Client receives ping**: `internal/events/client.go:271-276` - Client tries to send pong response
3. **Connection already closed**: The connection has been closed (by client or daemon), but the client still attempts to write

**Critical Code Flow**:

1. `internal/events/client.go:236-278` - `readEvents()` function
   - Line 246: Sets read deadline to 60 seconds
   - Line 253: Attempts to decode message
   - If connection is closed or times out, returns error
   - Line 273: Sends pong response **without checking if connection is still valid**

2. `internal/events/client.go:186-198` - `sendToSocket()` function
   - Line 190-192: Checks if `c.conn == nil` but this doesn't detect a closed connection
   - A connection can be closed but `c.conn` is still non-nil

3. **Race condition**: The connection can be closed between:
   - The check at line 240 (in `readEvents`)
   - The pong send at line 273

### Why This Happens

This error is mostly harmless and occurs when:
- Client disconnects abruptly
- Network issues cause socket closure
- Daemon removes a stale client while the client is processing a message

The error message is logged but doesn't indicate a serious problem. It's a normal part of handling disconnections.

### Fix Strategy

The fix should suppress this specific error and handle it gracefully:

**File**: `internal/events/client.go:273-275`

Change from:
```go
if err := c.sendToSocket(Event{Type: EventPong}); err != nil {
    log.Printf("Failed to send pong: %v", err)
}
```

To:
```go
if err := c.sendToSocket(Event{Type: EventPong}); err != nil {
    // Broken pipe/connection closed is expected during disconnection
    if !isConnectionError(err) {
        log.Printf("Failed to send pong: %v", err)
    }
}
```

Add helper function in `internal/events/client.go`:
```go
// isConnectionError checks if an error is a network connection error
func isConnectionError(err error) bool {
    if err == nil {
        return false
    }
    errStr := err.Error()
    return strings.Contains(errStr, "broken pipe") ||
           strings.Contains(errStr, "connection reset") ||
           strings.Contains(errStr, "use of closed network connection")
}
```

**Alternative/Better Fix**: Detect closed connections before attempting to write:

**File**: `internal/events/client.go:186-198`

Change `sendToSocket()`:
```go
func (c *Client) sendToSocket(event Event) error {
    c.mu.Lock()
    defer c.mu.Unlock()

    if c.conn == nil {
        return fmt.Errorf("not connected to daemon")
    }

    // Set a short write deadline to detect dead connections
    if err := c.conn.SetWriteDeadline(time.Now().Add(5 * time.Second)); err != nil {
        return fmt.Errorf("connection error: %w", err)
    }

    msg := Message{
        Type:  "event",
        Event: &event,
    }
    return c.encoder.Encode(msg)
}
```

---

## Issue 3: Task Movement Outside Visible Range

### Error Message
```
2025/12/17 23:55:25 Error moving task up: failed to find task above position 2: sql: no rows in result set
```

### Root Cause

**Location**: `internal/database/task_repository.go:611-617`

**The Problem**: The task movement logic assumes that tasks are **consecutively positioned** (0, 1, 2, 3...) in the database, but the TUI may only load a **subset of tasks** for display, creating gaps in the visible positions.

### Detailed Analysis

**How tasks are loaded**:

1. **Initial load** (`internal/tui/model.go:76-81`):
   ```go
   tasks, err := repo.GetTaskSummariesByProject(loadCtx, currentProjectID)
   ```
   ✅ Loads ALL tasks for the project, grouped by column

2. **After creating a task** (`internal/tui/update.go:190`):
   ```go
   summaries, err := m.repo.GetTaskSummariesByColumn(ctx, currentCol.ID)
   ```
   ❌ Only loads tasks for the CURRENT column, not the entire project

3. **After editing a task** (`internal/tui/update.go:301`):
   ```go
   summaries, err := m.repo.GetTaskSummariesByColumn(ctx, currentCol.ID)
   ```
   ❌ Only loads tasks for the CURRENT column

4. **After swapping tasks** (`internal/tui/update.go:1156`):
   ```go
   summaries, err := m.repo.GetTaskSummariesByColumn(ctx, currentCol.ID)
   ```
   ❌ Only loads tasks for the CURRENT column

**The Real Problem**: When you use `GetTaskSummariesByColumn()` for individual columns, the TUI's local state (`m.appState.Tasks()`) only contains tasks for **some columns**, not all columns in the project.

**Example scenario**:
```
Project has columns: Todo, In Progress, Done
Initial load: All tasks loaded for all 3 columns
User creates a task in "In Progress"
  → Only "In Progress" tasks are reloaded
  → "Todo" and "Done" tasks are stale in local state

User switches to "Todo" column and tries to move a task
  → TUI has stale data for "Todo"
  → Database has current data
  → Position mismatch occurs
```

### Why The Error Occurs

**File**: `internal/database/task_repository.go:611-617`

```go
// Find the task above (position - 1)
adjacentPos = currentPos - 1
newPos = currentPos - 1
err = tx.QueryRowContext(ctx,
    `SELECT id FROM tasks WHERE column_id = ? AND position = ?`,
    columnID, adjacentPos,
).Scan(&adjacentTaskID)
if err != nil {
    return fmt.Errorf("failed to find task above position %d: %w", currentPos, err)
}
```

This query **assumes continuous positions**. If:
- Task at position 2 is being moved up
- The query looks for position 1
- But position 1 doesn't exist (gap in positions)
- Returns `sql.ErrNoRows`

**However**, this shouldn't happen in practice because:
1. Tasks are created with sequential positions (0, 1, 2, ...)
2. When tasks are deleted, positions aren't recompacted
3. When tasks are moved, positions are swapped

**The REAL issue**: The error message is misleading. The actual problem is that the **task ID being passed to SwapTaskUp** doesn't exist or is in a different column than the TUI thinks it is.

### Root Cause Deep Dive

Looking at the swap logic:

**File**: `internal/tui/model.go:418-454`

```go
func (m Model) moveTaskUp() {
    task := m.getCurrentTask()  // Line 410: Gets task from LOCAL state
    if task == nil {
        return
    }
    
    // ... validation checks ...
    
    ctx, cancel := m.uiContext()
    defer cancel()
    err := m.repo.SwapTaskUp(ctx, task.ID)  // Line 421: Uses task from LOCAL state
```

The issue:
1. `getCurrentTask()` returns a task from `m.appState.Tasks()` (local TUI state)
2. This local state might be **stale** if only one column was reloaded
3. The task ID and position in local state don't match the database
4. The database query fails because the actual position is different

### Why User Report Says "Moving Task Outside Visible Range"

The user is correct! The issue is:
1. User scrolls or filters to show only some tasks
2. TUI reloads only the visible column's tasks
3. User tries to move a task in a different column
4. That column's tasks are stale
5. Mismatch between TUI state and DB state

### Fix Strategy

**Solution 1 (Recommended)**: Always reload ALL tasks for the project after any task operation

**Files to change**:
- `internal/tui/update.go:190` (after creating task)
- `internal/tui/update.go:301` (after editing task)
- `internal/tui/update.go:1156` (after swapping tasks)

Change from:
```go
summaries, err := m.repo.GetTaskSummariesByColumn(ctx, currentCol.ID)
if err != nil {
    log.Printf("Error refreshing tasks: %v", err)
    return m, nil
}
m.appState.SetTasksForColumn(currentCol.ID, summaries)
```

To:
```go
// Reload all tasks for the project to keep state consistent
project := m.getCurrentProject()
if project != nil {
    tasksByColumn, err := m.repo.GetTaskSummariesByProject(ctx, project.ID)
    if err != nil {
        log.Printf("Error refreshing tasks: %v", err)
        return m, nil
    }
    m.appState.SetTasks(tasksByColumn)
}
```

**Solution 2 (Performance-focused)**: Fix the swap logic to handle gaps

**File**: `internal/database/task_repository.go:608-617`

Change from:
```go
// Find the task above (position - 1)
adjacentPos = currentPos - 1
newPos = currentPos - 1
err = tx.QueryRowContext(ctx,
    `SELECT id FROM tasks WHERE column_id = ? AND position = ?`,
    columnID, adjacentPos,
).Scan(&adjacentTaskID)
if err != nil {
    return fmt.Errorf("failed to find task above position %d: %w", currentPos, err)
}
```

To:
```go
// Find the task above (next smaller position)
adjacentPos = currentPos - 1
newPos = currentPos - 1
err = tx.QueryRowContext(ctx,
    `SELECT id, position FROM tasks 
     WHERE column_id = ? AND position < ?
     ORDER BY position DESC LIMIT 1`,
    columnID, currentPos,
).Scan(&adjacentTaskID, &adjacentPos)
if err != nil {
    return fmt.Errorf("failed to find task above position %d: %w", currentPos, err)
}
newPos = adjacentPos
```

This query finds the task with the **highest position less than the current task**, which handles gaps correctly.

**Solution 3 (Best for correctness)**: Implement both solutions

1. Use Solution 1 to ensure TUI state is always fresh
2. Use Solution 2 to make the database logic more robust

---

## Recommendation Priority

### High Priority
1. **SQLite BUSY** - Fix immediately as it causes data operations to fail
   - Enable WAL mode
   - Set busy timeout
   - Configure connection pool

### Medium Priority
2. **Task Movement** - Fix to prevent data consistency issues
   - Implement Solution 1 (reload all tasks) first for safety
   - Then implement Solution 2 for robustness

### Low Priority
3. **Broken Pipe** - Cosmetic error that doesn't affect functionality
   - Suppress or add better error handling
   - Consider this a low-priority polish item

---

## Testing Recommendations

After implementing fixes:

1. **SQLite BUSY**: 
   - Test concurrent task movements
   - Test moving tasks while browsing columns
   - Verify no SQLITE_BUSY errors appear

2. **Task Movement**:
   - Create multiple columns with multiple tasks
   - Move tasks between columns
   - Try moving tasks up/down in each column
   - Test after filtering/searching
   - Verify positions remain consistent

3. **Broken Pipe**:
   - Start daemon and TUI
   - Kill daemon while TUI is running
   - Restart daemon
   - Verify graceful reconnection
   - Check that errors are not logged or are suppressed

---

## Summary

The three issues have distinct root causes:

1. **SQLITE_BUSY**: Database configuration issue - needs WAL mode and busy timeout
2. **Broken Pipe**: Expected network error during disconnection - needs graceful handling
3. **Task Movement**: State synchronization issue - TUI loads partial data causing mismatches

All three issues are fixable with the recommended changes.
