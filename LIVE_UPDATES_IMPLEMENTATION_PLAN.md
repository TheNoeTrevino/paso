# Live Updates Implementation Plan
**Project:** Paso - Terminal Kanban Board  
**Feature:** Real-time synchronization across multiple terminal instances  
**Date:** 2025-12-15  
**Target OS:** Linux only

---

## Overview

Implement a daemon-based event notification system to enable live updates across multiple Paso instances. When any instance modifies the database, all other instances receive a notification to refresh their data.

### Core Principle
**Database change → Notify daemon → Broadcast to all clients → Refresh UI**

We don't care *what* changed, only *that something* changed. Each client does a full refresh of visible data.

---

## Architecture

```
┌─────────────────┐         ┌─────────────────┐
│   Paso TUI #1   │◄───────►│                 │
└─────────────────┘         │                 │
                            │  Paso Daemon    │
┌─────────────────┐         │  (Unix Socket)  │
│   Paso TUI #2   │◄───────►│                 │
└─────────────────┘         │   ~/.paso/      │
                            │   paso.sock     │
┌─────────────────┐         │                 │
│   Paso TUI #3   │◄───────►│                 │
└─────────────────┘         └─────────────────┘
                                     │
                                     ▼
                            ┌─────────────────┐
                            │  SQLite DB      │
                            │  tasks.db       │
                            └─────────────────┘
```

---

## Components

### 1. Event System (`internal/events/`)

#### `internal/events/types.go`
```go
package events

import "time"

// Event represents a database change notification
type Event struct {
    Type      EventType
    ProjectID int
    Timestamp time.Time
}

// EventType indicates what kind of change occurred
type EventType string

const (
    EventDatabaseChanged EventType = "db_changed"
)

// Message wraps events for wire protocol
type Message struct {
    Event Event
}
```

#### `internal/events/client.go`
- `Client` struct with Unix socket connection
- `NewClient(socketPath string) (*Client, error)` - Connect to daemon
- `SendEvent(event Event) error` - Notify daemon of change
- `Listen(ctx context.Context) (<-chan Event, error)` - Receive broadcasts
- `Close() error` - Cleanup connection

**Implementation details:**
- Uses `net.Dial("unix", socketPath)`
- JSON encoding for messages over socket
- Buffered channel (size 10) for incoming events
- Goroutine reads from socket, sends to channel
- Non-blocking sends to avoid blocking daemon

---

### 2. Daemon (`internal/daemon/`)

#### `internal/daemon/server.go`
```go
type Server struct {
    listener  net.Listener
    clients   map[*client]bool
    mutex     sync.RWMutex
    ctx       context.Context
    cancel    context.CancelFunc
    broadcast chan events.Event
}
```

**Key methods:**
- `NewServer(socketPath string) (*Server, error)` - Create daemon
- `Start(ctx context.Context) error` - Accept connections
- `Broadcast(event events.Event)` - Send to all clients
- `Shutdown() error` - Graceful shutdown

**Implementation details:**
- Creates Unix socket at `~/.paso/paso.sock`
- Goroutine per client reading events
- Central broadcast goroutine distributes to all clients
- Remove client on disconnect/error
- Clean up socket file on shutdown

#### Socket path location
`~/.paso/paso.sock` (same directory as database)

#### Daemon lifecycle
- Auto-start on first Paso TUI launch if not running
- PID file at `~/.paso/paso.pid` to track running daemon
- Graceful shutdown on last client disconnect (optional - or keep running)

---

### 3. Database Integration (`internal/database/`)

#### Modify `repository.go`
Add event client to Repository:
```go
type Repository struct {
    *ProjectRepo
    *ColumnRepo
    *TaskRepo
    *LabelRepo
    eventClient *events.Client // NEW
}

func NewRepository(db *sql.DB, eventClient *events.Client) *Repository {
    return &Repository{
        ProjectRepo: &ProjectRepo{db: db},
        ColumnRepo:  &ColumnRepo{db: db},
        TaskRepo:    &TaskRepo{db: db},
        LabelRepo:   &LabelRepo{db: db},
        eventClient: eventClient,
    }
}
```

#### Wrap write operations
After any mutation (INSERT/UPDATE/DELETE), send event:
```go
func (r *Repository) CreateTask(...) (*models.Task, error) {
    // ... existing code ...
    
    if r.eventClient != nil {
        _ = r.eventClient.SendEvent(events.Event{
            Type:      events.EventDatabaseChanged,
            ProjectID: projectID,
            Timestamp: time.Now(),
        })
    }
    
    return task, nil
}
```

**Affected methods:**
- `CreateTask`, `UpdateTask`, `DeleteTask`
- `CreateColumn`, `UpdateColumn`, `DeleteColumn`
- `CreateProject`, `UpdateProject`, `DeleteProject`
- `CreateLabel`, `UpdateLabel`, `DeleteLabel`
- `UpdateTaskColumn` (task movement)
- Any other write operations

**Error handling:**
- Event send failures are logged but don't fail the operation
- Database write succeeds even if notification fails
- Gracefully handle nil eventClient (daemon not running)

---

### 4. TUI Integration (`internal/tui/`)

#### Modify `model.go`
Add event listener channel:
```go
type Model struct {
    // ... existing fields ...
    eventClient   *events.Client
    eventChan     <-chan events.Event
}

func InitialModel(ctx context.Context, repo database.DataStore, cfg *config.Config, eventClient *events.Client) Model {
    // ... existing code ...
    
    // Start listening for events
    var eventChan <-chan events.Event
    if eventClient != nil {
        var err error
        eventChan, err = eventClient.Listen(ctx)
        if err != nil {
            log.Printf("Failed to listen for events: %v", err)
        }
    }
    
    return Model{
        // ... existing fields ...
        eventClient: eventClient,
        eventChan:   eventChan,
    }
}
```

#### Add subscription command
```go
// subscribeToEvents returns a command that waits for database change events
func (m Model) subscribeToEvents() tea.Cmd {
    if m.eventChan == nil {
        return nil
    }
    
    return func() tea.Msg {
        event := <-m.eventChan
        return RefreshMsg{Event: event}
    }
}

// RefreshMsg signals that data should be reloaded
type RefreshMsg struct {
    Event events.Event
}
```

#### Modify `update.go`
Handle refresh messages:
```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // ... existing context check ...
    
    switch msg := msg.(type) {
    case RefreshMsg:
        log.Printf("Received refresh event for project %d", msg.Event.ProjectID)
        
        // Reload current project data
        m = m.reloadCurrentProject()
        
        // Continue listening for next event
        return m, m.subscribeToEvents()
    
    case tea.KeyMsg:
        // ... existing handlers ...
    }
    
    return m, nil
}
```

#### Add reload method
```go
func (m Model) reloadCurrentProject() Model {
    ctx, cancel := context.WithTimeout(m.ctx, timeoutDB)
    defer cancel()
    
    currentProject := m.appState.CurrentProject()
    if currentProject == nil {
        return m
    }
    
    // Reload columns
    columns, err := m.repo.GetColumnsByProject(ctx, currentProject.ID)
    if err != nil {
        log.Printf("Error reloading columns: %v", err)
        return m
    }
    
    // Reload tasks
    tasks, err := m.repo.GetTaskSummariesByProject(ctx, currentProject.ID)
    if err != nil {
        log.Printf("Error reloading tasks: %v", err)
        return m
    }
    
    // Reload labels
    labels, err := m.repo.GetLabelsByProject(ctx, currentProject.ID)
    if err != nil {
        log.Printf("Error reloading labels: %v", err)
        return m
    }
    
    // Update state (preserve cursor position if possible)
    m.appState.SetColumns(columns)
    m.appState.SetTasks(tasks)
    m.appState.SetLabels(labels)
    
    return m
}
```

#### Start subscription in InitialModel
At the end of `InitialModel`, return command to start listening:
```go
// In main.go after creating model:
p := tea.NewProgram(model, tea.WithContext(ctx))

// Add initial command to start event subscription
if model.eventChan != nil {
    go func() {
        // This ensures we start listening immediately
        p.Send(model.subscribeToEvents()())
    }()
}
```

---

### 5. Main Entry Point (`main.go` & `cmd/daemon/`)

#### Modify `main.go`
```go
func main() {
    ctx, cancel := signal.NotifyContext(...)
    defer cancel()
    
    cfg, err := config.Load()
    // ... error handling ...
    
    // Ensure daemon is running
    socketPath := filepath.Join(os.Getenv("HOME"), ".paso", "paso.sock")
    if err := ensureDaemonRunning(socketPath); err != nil {
        log.Printf("Warning: Failed to ensure daemon running: %v", err)
    }
    
    // Connect to daemon
    eventClient, err := events.NewClient(socketPath)
    if err != nil {
        log.Printf("Warning: Failed to connect to daemon: %v", err)
        eventClient = nil // Continue without live updates
    }
    defer func() {
        if eventClient != nil {
            eventClient.Close()
        }
    }()
    
    db, err := database.InitDB(ctx)
    // ... error handling ...
    
    repo := database.NewRepository(db, eventClient)
    model := tui.InitialModel(ctx, repo, cfg, eventClient)
    
    // ... rest of program ...
}

func ensureDaemonRunning(socketPath string) error {
    // Check if socket exists
    if _, err := os.Stat(socketPath); err == nil {
        // Socket exists, try to connect
        conn, err := net.Dial("unix", socketPath)
        if err == nil {
            conn.Close()
            return nil // Daemon is running
        }
        // Socket exists but can't connect - remove stale socket
        os.Remove(socketPath)
    }
    
    // Start daemon
    cmd := exec.Command("paso", "daemon")
    cmd.Stdout = nil
    cmd.Stderr = nil
    if err := cmd.Start(); err != nil {
        return fmt.Errorf("failed to start daemon: %w", err)
    }
    
    // Detach daemon
    cmd.Process.Release()
    
    // Wait for socket to be created (max 2 seconds)
    for i := 0; i < 20; i++ {
        if _, err := os.Stat(socketPath); err == nil {
            return nil
        }
        time.Sleep(100 * time.Millisecond)
    }
    
    return fmt.Errorf("daemon did not create socket in time")
}
```

#### Create `cmd/daemon/main.go`
```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "os/signal"
    "path/filepath"
    "syscall"
    
    "github.com/thenoetrevino/paso/internal/daemon"
)

func main() {
    ctx, cancel := signal.NotifyContext(
        context.Background(),
        os.Interrupt,
        syscall.SIGTERM,
    )
    defer cancel()
    
    home, _ := os.UserHomeDir()
    socketPath := filepath.Join(home, ".paso", "paso.sock")
    
    server, err := daemon.NewServer(socketPath)
    if err != nil {
        log.Fatalf("Failed to create daemon: %v", err)
    }
    
    log.Printf("Paso daemon starting on %s", socketPath)
    
    if err := server.Start(ctx); err != nil {
        log.Fatalf("Daemon error: %v", err)
    }
    
    log.Println("Paso daemon shutting down")
}
```

#### Update build
Modify `go.mod` if needed, ensure `cmd/daemon` builds:
```bash
go build -o bin/paso-daemon ./cmd/daemon
```

---

## Implementation Steps

### Phase 1: Event Infrastructure
1. Create `internal/events/types.go` with event definitions
2. Create `internal/events/client.go` with client implementation
3. Add tests for event client

### Phase 2: Daemon Server
4. Create `internal/daemon/server.go` with Unix socket server
5. Implement client management and broadcasting
6. Add daemon tests (mock socket connections)
7. Create `cmd/daemon/main.go` entry point

### Phase 3: Database Integration
8. Modify `internal/database/repository.go` to accept event client
9. Add event notifications to all write operations
10. Update tests to handle optional event client

### Phase 4: TUI Integration
11. Modify `internal/tui/model.go` to accept event client
12. Add `RefreshMsg` and `subscribeToEvents()` command
13. Implement `reloadCurrentProject()` method
14. Update `Update()` to handle refresh messages
15. Add tests for refresh behavior

### Phase 5: Main Integration
16. Modify `main.go` to start/connect to daemon
17. Implement `ensureDaemonRunning()` helper
18. Update initialization to pass event client through stack
19. Handle graceful degradation if daemon unavailable

### Phase 6: Testing & Polish
20. Manual testing with multiple terminal instances
21. Test daemon crash recovery
22. Test socket cleanup on shutdown
23. Add logging for debugging
24. Update README with daemon information

---

## Error Handling

### Daemon not running
- TUI continues to work without live updates
- Log warning but don't crash
- Event client is nil throughout stack

### Daemon crashes
- Clients detect disconnection
- Log error and continue without live updates
- User can restart Paso to reconnect

### Socket permission issues
- Fail fast with clear error message
- Check `~/.paso/` permissions on startup

### Network errors
- Non-blocking event sends
- Log failures but don't fail DB operations
- Clients auto-reconnect (future enhancement)

---

## Performance Considerations

### Event frequency
- No throttling initially (simple broadcast)
- Future: Debounce events (100ms window)

### Refresh scope
- Full project reload (columns, tasks, labels)
- Simple but slightly inefficient
- Future: Granular updates based on event type

### Memory
- Daemon uses ~5-10MB RAM
- Each client connection: ~100KB
- Event channel buffer: 10 events

### CPU
- Minimal overhead (<1% CPU)
- JSON encoding/decoding is fast for small events
- No polling - event-driven only

---

## Testing Strategy

### Unit Tests
- `events.Client`: Mock socket connections
- `daemon.Server`: Test broadcast logic
- Repository event sending: Verify events sent after writes

### Integration Tests
- Start real daemon, connect multiple clients
- Verify broadcasts received
- Test daemon shutdown with connected clients

### Manual Testing Checklist
1. Start Paso instance #1
2. Start Paso instance #2
3. Create task in #1 → verify appears in #2
4. Move task in #2 → verify updates in #1
5. Delete column in #1 → verify removed in #2
6. Kill daemon → verify both instances continue working
7. Restart daemon → verify reconnection (manual restart required)

---

## Future Enhancements (Out of Scope)

- **Auto-reconnection**: Clients detect daemon restart and reconnect
- **Granular events**: Include what changed (task ID, column ID, etc.)
- **Event filtering**: Only refresh affected columns/tasks
- **Network support**: TCP sockets for remote collaboration
- **Event persistence**: Store events in DB for audit log
- **Conflict resolution**: Handle concurrent edits gracefully
- **Debouncing**: Batch rapid changes into single refresh

---

## File Changes Summary

### New Files
- `internal/events/types.go` (~50 lines)
- `internal/events/client.go` (~200 lines)
- `internal/events/client_test.go` (~100 lines)
- `internal/daemon/server.go` (~300 lines)
- `internal/daemon/server_test.go` (~150 lines)
- `cmd/daemon/main.go` (~50 lines)

### Modified Files
- `main.go` (~40 lines added)
- `internal/database/repository.go` (~30 lines added)
- `internal/database/*_repository.go` (~10 lines each, 6 files)
- `internal/tui/model.go` (~60 lines added)
- `internal/tui/update.go` (~20 lines added)
- `README.md` (documentation section)

### Total New Code
~1000 lines of new code
~160 lines of modifications

---

## Configuration

### No new config needed
Daemon socket path is hardcoded: `~/.paso/paso.sock`

Future: Allow custom socket path in `config.yaml`:
```yaml
daemon:
  socket_path: ~/.paso/paso.sock
  auto_start: true
```

---

## Security Considerations

### Unix socket permissions
- Socket created with `0700` permissions (owner only)
- Only user who created socket can connect
- Safe for multi-user systems

### Process isolation
- Daemon runs as same user as TUI
- No privilege escalation
- No external network access

### Input validation
- Validate all messages from socket
- Ignore malformed events
- Rate limiting (future enhancement)

---

## Compatibility

### Linux only
- Uses Unix domain sockets (`net.Dial("unix", ...)`)
- Uses standard `os/signal` for SIGTERM
- No Windows support needed

### Go version
- Requires Go 1.25+ (current project version)
- Uses standard library only (no new dependencies)

---

## Deployment

### Installation
```bash
# Build both binaries
go build -o bin/paso .
go build -o bin/paso-daemon ./cmd/daemon

# Install (optional)
sudo cp bin/paso /usr/local/bin/
sudo cp bin/paso-daemon /usr/local/bin/
```

### Systemd service (optional)
```ini
[Unit]
Description=Paso Daemon
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/paso-daemon
Restart=always
User=%u

[Install]
WantedBy=default.target
```

---

## Success Criteria

- [ ] Multiple Paso instances update in real-time
- [ ] Daemon auto-starts on first Paso launch
- [ ] No crashes when daemon is not running
- [ ] Graceful shutdown of daemon and clients
- [ ] Clean socket file management
- [ ] All existing tests pass
- [ ] New tests cover event system
- [ ] Documentation updated

---

## Rollback Plan

If implementation causes issues:
1. Feature is isolated in new packages (`events`, `daemon`)
2. Remove daemon integration from `main.go`
3. Remove event client from `repository.go`
4. Remove event handling from `tui/update.go`
5. System returns to current state with no live updates

---

## Timeline Estimate

- Phase 1 (Event Infrastructure): 2 hours
- Phase 2 (Daemon Server): 4 hours
- Phase 3 (Database Integration): 2 hours
- Phase 4 (TUI Integration): 3 hours
- Phase 5 (Main Integration): 2 hours
- Phase 6 (Testing & Polish): 3 hours

**Total: ~16 hours of development**

---

## Questions & Decisions

### Q: Should daemon persist between sessions?
**A:** Yes, daemon continues running until system reboot or manual kill. First Paso instance starts it, last instance doesn't stop it (keeps it ready).

### Q: What if two users edit simultaneously?
**A:** Last write wins (SQLite behavior). Future: Add optimistic locking or conflict detection.

### Q: Daemon startup time?
**A:** <100ms from execution to socket ready

### Q: Event message format?
**A:** JSON for simplicity and debuggability. Binary protocol if performance becomes issue.

---

## End of Plan

This plan is comprehensive and ready for implementation. All major decisions are documented, and the architecture is sound for Linux-based terminal applications.
