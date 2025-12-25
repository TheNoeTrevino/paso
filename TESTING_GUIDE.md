# Live Updates Testing Guide

## Prerequisites

1. **Build both binaries:**
   ```bash
   go build -o bin/paso-daemon ./cmd/daemon
   go build -o bin/paso .
   ```

2. **Set PASO_EVENT_DEBOUNCE_MS (optional):**
   ```bash
   export PASO_EVENT_DEBOUNCE_MS=100  # default, 50-200ms recommended
   ```

## Test 1: Single Instance (Daemon Not Running)

**Goal:** Verify paso works without daemon (graceful degradation)

```bash
# Terminal 1: Start paso (no daemon running)
./bin/paso

# Expected behavior:
# - Prints warnings about daemon connection but continues
# - Suggests: "Start daemon: systemctl --user start paso"
# - UI loads normally
# - No live updates (expected)
# - Can create/edit/delete tasks normally
```

## Test 2: Daemon Installation

**Goal:** Verify systemd service installation

```bash
# Copy daemon to system location (requires sudo)
sudo cp bin/paso-daemon /usr/local/bin/

# Run installer script
chmod +x scripts/install_systemd.sh
./scripts/install_systemd.sh

# Verify service is running
systemctl --user status paso

# View logs
journalctl --user -u paso -f

# Expected output:
# - Service active (running)
# - Logs show "Paso daemon starting on ~/.paso/paso.sock"
# - Logs show "Process ID: XXXX"
```

## Test 3: Basic Live Updates (2 Instances)

**Goal:** Verify real-time sync between two paso instances

**Terminal 1: Start paso instance #1**
```bash
./bin/paso
```

**Terminal 2: Start paso instance #2**
```bash
./bin/paso
```

**In Terminal 1: Create a task**
- Press 'n' to create new task
- Type task name: "Test live updates"
- Press Enter
- Press Tab to confirm

**Expected in Terminal 2:**
- ✅ New task appears within 100ms
- ✅ Shows "Synced with other instances" notification
- ✅ No flashing or flickering

## Test 4: Event Batching

**Goal:** Verify rapid events are batched (not 10 refreshes for 10 moves)

**Terminal 1: Start paso instance #1**
```bash
./bin/paso
```

**Terminal 2: Start paso instance #2 (watching logs)**
```bash
PASO_EVENT_DEBOUNCE_MS=100 ./bin/paso
# Enable debug logging if available
```

**In Terminal 1: Move task 10 times rapidly**
- Select a task
- Press 'm' to move task (or arrow keys)
- Rapidly move left/right 10 times

**Expected:**
- ✅ Terminal 2 refreshes only 1-2 times, not 10 times
- ✅ Shows "Synced with other instances" once
- ✅ No lag or stutter in UI

## Test 5: Project Scoping

**Goal:** Verify events are project-scoped (no unnecessary refreshes)

**Setup:**
- Create Project A with a task
- Create Project B with a task
- Start paso instance #1 (viewing Project A)
- Start paso instance #2 (viewing Project B)

**In Terminal 1: Modify Project A task**
```bash
# Edit/move/create task in Project A
```

**Expected in Terminal 2:**
- ✅ No refresh (Project B is not affected)
- ✅ No "Synced" notification

**In Terminal 1: Switch to Project B, modify task**
```bash
# Switch project, then edit task
```

**Expected in Terminal 2:**
- ✅ Refreshes immediately
- ✅ Shows "Synced with other instances"

## Test 6: Reconnection After Daemon Restart

**Goal:** Verify clients reconnect after daemon crashes

**Terminal 1: Start paso instance #1**
```bash
./bin/paso
```

**Terminal 2: Start paso instance #2**
```bash
./bin/paso
```

**Terminal 3: Watch daemon status**
```bash
journalctl --user -u paso -f
```

**In Terminal 1: Create a task to verify sync works**
- Task appears in Terminal 2 within 100ms ✅

**Terminal 3: Restart daemon**
```bash
systemctl --user restart paso
```

**Expected:**
- ✅ Clients continue working (show "Reconnecting..." or similar)
- ✅ After restart, new tasks sync again
- ✅ No crashes or errors

**In Terminal 1: Create another task**
- Task should appear in Terminal 2 again ✅

## Test 7: Daemon Crash Recovery (5 Retries)

**Goal:** Verify reconnection with exponential backoff

**Setup from Test 6:**

**Terminal 3: Stop daemon (don't restart)**
```bash
systemctl --user stop paso
```

**In Terminal 1: Create a task**
- Task saves locally (works)
- No sync notification (expected, daemon unavailable)

**Terminal 3: Start daemon again (after a few seconds)**
```bash
systemctl --user start paso
```

**Expected:**
- ✅ Both instances reconnect automatically
- ✅ New tasks sync when daemon restarts
- ✅ No manual restart needed for paso

## Test 8: Concurrent Operations

**Goal:** Verify data consistency with multiple simultaneous edits

**Terminal 1 & 2: Both running paso**

**Simultaneous in both:**
- Terminal 1: Press 'n', type "Task from T1", press Enter
- Terminal 2: Press 'n', type "Task from T2", press Enter
- Both press Tab to confirm (at same time)

**Expected:**
- ✅ Both tasks appear in both terminals
- ✅ Both sync notifications show
- ✅ No data corruption or duplicates

## Test 9: Column Operations

**Goal:** Verify column changes sync

**Terminal 1 & 2: Both running paso**

**In Terminal 1:**
- Press 'a' to add column
- Type "In Review"
- Confirm

**Expected in Terminal 2:**
- ✅ New column appears immediately
- ✅ Shows "Synced with other instances"

## Test 10: Label Operations

**Goal:** Verify label changes sync

**Terminal 1 & 2: Both running paso**

**In Terminal 1:**
- Open a task
- Assign labels (e.g., "urgent", "blocking")

**Expected in Terminal 2:**
- ✅ Labels appear on task immediately
- ✅ Shows "Synced with other instances"

## Test 11: Memory & Performance

**Goal:** Verify no memory leaks or performance degradation

**Run for 5 minutes:**
```bash
# Terminal 1: Start paso
./bin/paso

# Terminal 2: Watch process memory
watch -n 1 'ps aux | grep paso | grep -v grep'

# Terminal 1: Rapidly perform operations
# - Create 20 tasks
# - Move them 5 times each
# - Add labels to 10 tasks
# - Switch projects 10 times
```

**Expected:**
- ✅ Memory usage stable (no growth)
- ✅ CPU usage returns to 0% between operations
- ✅ No goroutine leaks (check logs)

## Test 12: Error Scenarios

**Test reconnection failure handling:**

```bash
# Start paso when daemon socket is broken
rm ~/.paso/paso.sock
./bin/paso

# Expected:
# - Shows helpful error message
# - Suggests "systemctl --user start paso"
# - Continues without crashes
```

## Verification Checklist

- [ ] Single instance works without daemon
- [ ] Daemon installs successfully via systemd
- [ ] Two instances sync in <150ms
- [ ] 10 rapid events batch into 1-2 refreshes
- [ ] Project scoping prevents cross-project refreshes
- [ ] Clients reconnect after daemon restart
- [ ] Exponential backoff works (5 retries)
- [ ] Concurrent edits don't corrupt data
- [ ] Column operations sync
- [ ] Label operations sync
- [ ] Memory usage stable over time
- [ ] Error messages are helpful

## Performance Baseline

On modern hardware, you should see:
- **Latency:** 50-150ms from edit to visible in other instance
- **Batching:** 10 rapid operations → 1-2 refreshes max
- **Memory:** ~50MB per paso instance (stable)
- **CPU:** 0% when idle, <5% during operations

## Known Limitations (by design)

- If daemon crashes and doesn't restart, clients show "Synced" notifications even if sync failed (non-blocking)
- Project switching doesn't auto-resubscribe (only happens at init time)
- No explicit "connection status" indicator in UI (use journalctl for diagnostics)

## Troubleshooting

**Daemon not starting:**
```bash
journalctl --user -u paso -n 50
systemctl --user restart paso
```

**No syncs happening:**
```bash
# Check daemon is running
systemctl --user status paso

# Restart it
systemctl --user restart paso

# Check socket exists
ls -la ~/.paso/paso.sock
```

**Memory leak suspected:**
```bash
# Check for goroutine leaks
pprof http://localhost:6060/debug/pprof/goroutine
```

---

## Automated Test Suite: Real-Time Updates Bug Fix

### Overview

This section documents the comprehensive automated tests added to prevent regression of the real-time updates bug where only Project 1 received live updates. The bug was caused by lingering write deadlines on socket connections that weren't properly cleared.

### Running the Tests

**Run all tests:**
```bash
go test ./...
```

**Run only short tests (skips slow reconnection tests):**
```bash
go test ./... -short
```

**Run with race detector:**
```bash
go test ./... -race
```

**Run specific test suites:**
```bash
# Events client tests
go test ./internal/events/... -v

# Daemon server tests  
go test ./internal/daemon/... -v

# TUI subscription tests
go test ./internal/tui/... -v -run "Subscription|RefreshMsg"
```

### Test Coverage (9 Tests Total)

#### 1. Events Client Tests (`internal/events/client_test.go`) - 4 tests

**TestSubscribe_AfterLongDelay** ⭐ CRITICAL
- Validates Subscribe works after write deadline expires
- Reproduces original bug without fix
- Runtime: ~1s

**TestSubscribe_ClearsWriteDeadline**
- Verifies Subscribe clears deadline after encoding
- Runtime: ~600ms

**TestSubscribe_MultipleRapidChanges**
- Tests rapid project switching (9 changes)
- Runtime: ~100ms

**TestSendEvent_ClearsDeadline**
- Verifies SendEvent clears deadline
- Runtime: ~1s

#### 2. Daemon Server Tests (`internal/daemon/server_test.go`) - 2 tests

**TestSubscription_Switching** ⭐ CRITICAL
- Validates server routes events based on subscription
- Runtime: ~700ms

**TestSubscription_ProjectZeroHandling**
- Verifies broadcast (ProjectID=0) reaches all clients
- Runtime: ~400ms

#### 3. TUI Tests (`internal/tui/model_subscriptions_test.go`) - 3 tests

**TestSwitchToProject_UpdatesSubscription** ⭐ CRITICAL
- Verifies subscription tracking when switching projects
- Runtime: <10ms

**TestRefreshMsg_HandlesProjectZero**
- Unit test for broadcast filtering logic
- Runtime: <10ms

**TestRefreshMsg_IgnoresWrongProject**
- Validates filtering of wrong-project events
- Runtime: <10ms

### Test Infrastructure

**MockEventPublisher Enhancement** (`internal/events/mock_test.go`)
- Added subscription tracking: `SubscriptionHistory`, `CurrentSubscription`
- Added helpers: `GetSubscriptionHistory()`, `GetCurrentSubscription()`

**Configurable Write Deadline** (`internal/events/client.go`)
- Production: 5 seconds
- Tests: 500ms (via `setWriteDeadlineForTest()`)

**TUI Test Fixtures** (`internal/tui/testdata.go`)
- `createTestProjects()`, `createTestColumns()`, `createTestTasks()`
- `createTestModel()`, `createTestModelWithProjects()`

### Execution Time

- **Total:** ~3 seconds
- **Short mode:** <3 seconds
- **Events tests:** ~3s
- **Daemon tests:** ~1s
- **TUI tests:** <10ms

### Bug Fix Validated

✅ Write deadlines cleared after operations  
✅ Subscribe works after deadline expiration  
✅ Rapid subscription changes work  
✅ Server routes events correctly  
✅ TUI updates subscriptions  
✅ Broadcast events reach all clients
