# Code Review: Paso - Terminal Kanban Board (Go/Charm CLI)

## Executive Summary

**Overall Grade: B+ (7.5/10)**

Paso is a well-architected TUI application with strong fundamentals. The database layer is production-quality, and the project follows Go best practices in most areas. However, there are notable anti-patterns, missing defensive programming, and opportunities for performance improvements.

---

## 1. PERFORMANCE ANALYSIS

### ⚠️ **Issues Found**

#### 1.1 Linked List Column Navigation (O(n) traversal)
**File:** `internal/database/column_repository.go`

The project uses a **doubly-linked list** structure for columns (PrevID/NextID pointers) instead of simple position integers. This requires O(n) traversal to find columns by position.

```go
type Column struct {
    ID      int
    PrevID  *int  // Pointer-based linked list
    NextID  *int
}
```

**Impact:** For 10+ columns, navigation becomes noticeably slower.

**Mitigation:** Partially addressed by caching columns in `AppState.columnByID` map (O(1) lookup).

**Recommendation:** Consider hybrid approach - maintain linked list for ordering but add position cache that updates on reorder operations.

---

#### 1.2 Reload Entire Column After Single Task Operation
**File:** `internal/tui/update.go:311-320`

After operations like adding a label to a task, the code reloads the **entire column's tasks**:

```go
func (m Model) handleAddLabelToTask() (Model, tea.Cmd) {
    // ... add label to single task
    m.reloadViewingTask()           // ✓ Reload one task
    m.reloadCurrentColumnTasks()    // ✗ Reload ALL tasks in column
    return m, nil
}
```

**Impact:** Unnecessary database queries. For a column with 50 tasks, you're re-fetching 49 tasks that didn't change.

**Recommendation:** Update task in-memory instead of full reload, or use selective refresh.

---

#### 1.3 GROUP_CONCAT Could Scale Poorly
**File:** `internal/database/task_repository.go:124-178`

Smart optimization using `GROUP_CONCAT` to fetch labels in one query:

```go
SELECT t.*, GROUP_CONCAT(l.name, CHAR(31)) as label_names
FROM tasks t LEFT JOIN task_labels tl ON t.id = tl.task_id
GROUP BY t.id
```

**Concern:** If a single task has 100+ labels, the concatenated string grows large. SQLite GROUP_CONCAT has 1GB default limit, but parsing overhead increases.

**Impact:** Low for typical use (<10 labels/task), but could degrade with pathological data.

**Recommendation:** Monitor and consider pagination or separate label query if label count per task exceeds 20.

---

#### 1.4 Swap Operations Use 3 UPDATEs
**File:** `internal/database/task_repository.go:623-672`

Swapping two tasks requires setting one to position `-1` temporarily to avoid UNIQUE constraint:

```go
UPDATE tasks SET position = -1 WHERE id = ?      // Step 1
UPDATE tasks SET position = ? WHERE id = ?       // Step 2
UPDATE tasks SET position = ? WHERE id = ?       // Step 3
```

**Better Approach:** Use SQLite's `DEFERRABLE INITIALLY DEFERRED` constraint:
```sql
CREATE TABLE tasks (
    position INT NOT NULL,
    column_id INT NOT NULL,
    UNIQUE(column_id, position) DEFERRABLE INITIALLY DEFERRED
);
```
Then swap in 2 UPDATEs within a transaction.

---

### ✅ **Good Optimizations**

1. **N+1 Query Prevention:** `GetTaskSummariesByProject` uses JOIN + GROUP_CONCAT
2. **Map-based caching:** `AppState.columnByID` provides O(1) column lookup
3. **Viewport limiting:** Only renders visible columns (`calculateViewportSize`)
4. **Pure rendering functions:** No heavy computation in View() methods
5. **Batch loading:** Tasks loaded per-project, not globally

**Performance Verdict:** **Acceptable for personal use** (<1000 tasks), but optimization needed for heavy usage.

---

## 2. IDIOMATIC GO - PATTERNS & ANTI-PATTERNS

### ✅ **Strong Idiomatic Patterns**

#### 2.1 Repository Pattern with Interface Segregation
**File:** `internal/database/datastore.go`

```go
type TaskRepository interface {
    TaskReader
    TaskWriter
    TaskMover
    TaskRelationshipReader
    TaskRelationshipWriter
}
```

Clean separation of read/write concerns. Excellent for testing.

#### 2.2 Struct Embedding for Composition
```go
type Repository struct {
    *ProjectRepo
    *ColumnRepo
    *TaskRepo
    *LabelRepo
}
```

Avoids method forwarding boilerplate.

#### 2.3 Error Wrapping (Consistent)
127 instances properly use `%w` for error chains:
```go
return fmt.Errorf("failed to get task detail for task %d: %w", taskID, err)
```

#### 2.4 Defer for Cleanup
Every transaction uses:
```go
defer tx.Rollback()  // Safe - no-op if committed
```

#### 2.5 Context Propagation
All database methods accept `context.Context` for cancellation support.

---

### ⚠️ **Anti-Patterns & Issues**

#### 2.6 String Comparison for Error Checking ❌
**File:** `internal/tui/model.go:311`

```go
if err != errors.New("task is already at the top of the column") {  // WRONG!
```

**Problem:** Creates new error object for comparison. Will NEVER match.

**Correct Approach:**
```go
// In models/errors.go:
var ErrAlreadyFirstTask = errors.New("task is already at the top")

// In model.go:
if !errors.Is(err, models.ErrAlreadyFirstTask) {
```

**Impact:** Critical bug - error handling not working as intended.

---

#### 2.7 Missing Bounds Checking on Slice Operations
**File:** `internal/tui/model.go:various`

```go
m.appState.Tasks()[currentCol.ID] = append(
    tasks[:m.uiState.SelectedTask()],
    tasks[m.uiState.SelectedTask()+1:]...
)
```

**Risk:** Panic if `SelectedTask()` is out of bounds or if `currentCol.ID` doesn't exist in map.

**Recommendation:** Add defensive checks:
```go
tasks, ok := m.appState.Tasks()[currentCol.ID]
if !ok || m.uiState.SelectedTask() >= len(tasks) {
    return m, nil  // or log error
}
```

---

#### 2.8 No Structured Logging
Uses `log.Printf` (62 instances) instead of structured logger:
```go
log.Printf("Error creating task: %v", err)  // Unstructured
```

**Better:**
```go
slog.Error("failed to create task",
    "error", err,
    "column_id", columnID,
    "user_action", "add_task")
```

**Benefits:** Structured logs enable filtering, log aggregation, error tracking.

---

#### 2.9 Large Function Files
- `task_repository.go`: **837 lines**
- `update.go`: **1067 lines**
- `migrations.go`: **662 lines**

**Recommendation:** Split by responsibility:
- `task_repository.go` → `task_crud.go`, `task_movement.go`, `task_relationships.go`
- `update.go` → `update_normal.go`, `update_forms.go`, `update_pickers.go`

---

#### 2.10 Mixed Concerns (Logging)
**File:** `internal/tui/handlers.go`

```go
func (m Model) handleAddTask() (Model, tea.Cmd) {
    task, err := m.repo.CreateTask(...)
    if err != nil {
        log.Printf("Error creating task: %v", err)  // TUI logging DB errors
        return m, nil
    }
}
```

**Problem:** TUI layer logs database errors. Logging should happen at repository layer or be returned as domain events.

**Better Approach:** Return structured error, let repository log, TUI displays notification.

---

#### 2.11 Magic Numbers Without Constants
```go
make([]*models.TaskSummary, 0, 50)  // Why 50?
```

Should be:
```go
const defaultTaskCapacity = 50
make([]*models.TaskSummary, 0, defaultTaskCapacity)
```

---

#### 2.12 No Custom Error Types
Only sentinel errors exist (`ErrAlreadyFirstTask`). Complex errors lack context.

**Better:**
```go
type ValidationError struct {
    Field   string
    Value   interface{}
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("%s: %v - %s", e.Field, e.Value, e.Message)
}
```

---

## 3. BEST PRACTICES REVIEW

### ✅ **Followed**

1. **Standard Project Layout:** Uses `/internal` for private packages ✓
2. **Dependency Injection:** Repo passed to model ✓
3. **Interface-Based Testing:** Repository interfaces enable mocks ✓
4. **Idempotent Migrations:** Checks column existence before ALTER TABLE ✓
5. **Foreign Key Constraints:** CASCADE deletes prevent orphans ✓
6. **Transaction Safety:** All mutations wrapped in transactions ✓
7. **Pure Go Dependencies:** `modernc.org/sqlite` avoids CGo ✓

### ⚠️ **Missed**

1. **No Linting Configuration:** Missing `.golangci.yml`
2. **No CI/CD:** No GitHub Actions or equivalent
3. **No Makefile:** Build commands not codified
4. **Version Embedding:** No `-ldflags` to embed version/commit
5. **No Graceful Shutdown:** SIGTERM not handled (minor for TUI)
6. **Test Coverage Gaps:** TUI layer at 25.9%, database at 59.3%

---

## 4. CODE STRUCTURE & MODULARITY

### ✅ **Strengths**

#### 4.1 Layered Architecture
```
internal/
├── models/         # Domain models (clean)
├── database/       # Persistence (repository pattern)
├── config/         # Configuration (YAML + defaults)
└── tui/            # Presentation (Bubble Tea)
    ├── state/      # 8 focused state objects
    ├── components/ # Pure rendering functions
    ├── layers/     # Modal system
    └── huhforms/   # Form builders
```

**Assessment:** Excellent separation. Each layer has clear responsibility.

#### 4.2 State Management (8 Objects)
- `AppState`: Domain data (projects, columns, tasks)
- `UIState`: Selection, viewport, mode
- `FormState`: Form lifecycle
- `InputState`: Simple text input
- `LabelPickerState`: Label selection
- `TaskPickerState`: Parent/child picker
- `NotificationState`: Toast notifications
- `SearchState`: Vim-style search

**Assessment:** Each state object has single responsibility. Well-organized.

#### 4.3 Pure Rendering Components
```go
func RenderTask(task *models.TaskSummary, selected bool) string
func RenderColumn(column *models.Column, tasks []*models.TaskSummary) string
```

Pure functions (input → output, no side effects). Highly testable.

---

### ⚠️ **Weaknesses**

#### 4.4 Tight Coupling (TUI → Database)
**File:** `internal/tui/model.go`

```go
type Model struct {
    repo database.DataStore  // Direct import
}
```

**Problem:** TUI layer tightly coupled to database package.

**Better Approach:**
```go
// In internal/tui/repository.go
type Repository interface {
    GetTasks(ctx context.Context, projectID int) ([]*models.TaskSummary, error)
    CreateTask(ctx context.Context, task *models.Task) error
    // ... only methods TUI needs
}

type Model struct {
    repo Repository  // Interface, not concrete type
}
```

**Benefits:**
- TUI tests don't need database
- Can swap implementations (API backend, in-memory)
- Clear contract of what TUI uses

---

#### 4.5 State Mutation Leaks
**File:** `internal/tui/state/app_state.go`

```go
func (s *AppState) Tasks() map[int][]*models.TaskSummary {
    return s.tasks  // Returns internal map!
}
```

**Problem:** Caller can mutate internal state:
```go
tasks := appState.Tasks()
tasks[5] = append(tasks[5], &models.TaskSummary{})  // Untracked mutation
```

**Fix:** Return copy or read-only view:
```go
func (s *AppState) Tasks() map[int][]*models.TaskSummary {
    copy := make(map[int][]*models.TaskSummary, len(s.tasks))
    for k, v := range s.tasks {
        copy[k] = append([]*models.TaskSummary{}, v...)
    }
    return copy
}
```

---

#### 4.6 Code Duplication - Picker Logic
**Files:** `state/label_picker_state.go` (151 lines), `state/task_picker_state.go` (165 lines)

Both implement identical filter/cursor/selection logic:
- `GetFilteredItems()`
- `MoveCursorUp/Down()`
- `AppendFilter()`
- `BackspaceFilter()`

**Recommendation:** Create generic `PickerState[T]` with parameterized item type:
```go
type PickerState[T any] struct {
    items       []T
    filter      string
    cursor      int
    getLabel    func(T) string
    getMatch    func(T, string) bool
}
```

Then specialize:
```go
type LabelPickerState struct {
    *PickerState[*models.Label]
    selectedIDs map[int]bool
}
```

**Savings:** ~100 lines of duplicated code eliminated.

---

## 5. CRITICAL BUGS FOUND

### 🔴 **Critical Issue #1: Map Access Without Existence Check**
**File:** `internal/tui/model.go:154`

```go
if m.uiState.SelectedTask() >= len(m.appState.Tasks()[currentCol.ID]) {
    m.uiState.SetSelectedTask(m.uiState.SelectedTask() - 1)
}
```

**Bug:** No check that `currentCol.ID` exists in Tasks map. **Will panic** if column has no tasks entry.

**Fix:**
```go
tasks, ok := m.appState.Tasks()[currentCol.ID]
if !ok || m.uiState.SelectedTask() >= len(tasks) {
    m.uiState.SetSelectedTask(max(0, len(tasks)-1))
    return m, nil
}
```

---

### 🟡 **High Priority Issue #2: Error Comparison Broken**
Already covered in section 2.6. Error handling not working due to string comparison.

---

### 🟡 **High Priority Issue #3: Race in Ticket Number Allocation?**
**File:** `internal/database/task_repository.go:19-75`

The code uses transactions to allocate ticket numbers, but if two processes open the same database file, SQLite's default locking may not prevent races.

**Current:**
```go
tx, err := r.db.BeginTx(ctx, nil)  // Default isolation
```

**Better:**
```go
tx, err := r.db.BeginTx(ctx, &sql.TxOptions{
    Isolation: sql.LevelSerializable,
})
```

**Note:** SQLite defaults to SERIALIZABLE, so this may already be safe. Verify with concurrent testing.

---

## 6. DEPENDENCY STABILITY

### ⚠️ **Pre-Release Dependencies**

**File:** `go.mod`

```go
charm.land/bubbletea/v2 v2.0.0-rc.2      // Release Candidate
charm.land/bubbles/v2 v2.0.0-rc.1        // Release Candidate
charm.land/lipgloss/v2 v2.0.0-beta.3     // Beta
charm.land/huh/v2 v2.0.0-20251118        // Development snapshot
```

**Risk:** Breaking changes possible before stable release.

**Impact:**
- API changes could break builds
- Bugs in pre-release versions
- Harder to find help (docs may be outdated)

**Recommendation:**
1. Document why pre-release versions are needed (v2 features?)
2. Pin exact versions (already done via `go.sum`)
3. Plan migration to stable v2 when released
4. Monitor Charm.sh release notes

---

## 7. TESTING ISSUES

### Test Coverage Results
```
internal/config:    82.5% ✓
internal/database:  59.3% ⚠️
internal/tui/state: 25.9% ❌
internal/tui:       Build failure (import issue)
```

### Issues

1. **TUI Tests Broken:** Import cycle or outdated import path
2. **Low State Coverage:** UI state objects under-tested (25.9%)
3. **No Integration Tests:** No full user flow testing
4. **No Error Path Tests:** Tests mostly cover happy paths

### Recommendations

1. Fix TUI test imports
2. Add table-driven tests for state transitions
3. Add integration tests: Create task → Move → Delete flow
4. Test error cases: DB failures, invalid input, concurrent access

---

## 8. MISSING ERROR HANDLING

### Examples

**1. No check for empty task list before access:**
```go
selectedTask := tasks[m.uiState.SelectedTask()]  // Could panic if empty
```

**2. Ignored errors in defer:**
```go
defer db.Close()  // Error ignored
```
Should be:
```go
defer func() {
    if err := db.Close(); err != nil {
        slog.Error("failed to close database", "error", err)
    }
}()
```

**3. No validation before database operations:**
```go
func (r *TaskRepo) CreateTask(ctx context.Context, task *models.Task) error {
    // No check: title not empty, description not too long, etc.
    _, err := r.db.ExecContext(ctx, insertSQL, task.Title, task.Description)
}
```

---

## 9. SUGGESTIONS FOR IMPROVEMENT

### Immediate (High Impact, Low Effort)

1. **Fix error comparison bug** (`model.go:311`) - Use `errors.Is()`
2. **Add bounds checking** before slice operations - Prevent panics
3. **Add `.golangci.yml`** with recommended linters - Catch issues early
4. **Add Makefile** with `build`, `test`, `lint`, `install` targets
5. **Fix TUI test imports** - Restore test coverage visibility

### Short-Term (High Impact, Medium Effort)

6. **Structured logging** - Replace `log.Printf` with `slog`
7. **Custom error types** - Add context to errors (ValidationError, NotFoundError)
8. **State getter defensive copies** - Prevent unintended mutations
9. **Refactor large files** - Split 800+ line files into focused modules
10. **Add integration tests** - Test full user workflows
11. **Version embedding** - Add `-ldflags` with version/commit

### Medium-Term (Medium Impact, High Effort)

12. **Decouple TUI from database** - Use interface in TUI layer
13. **Generic picker abstraction** - Eliminate duplication between label/task pickers
14. **Optimize column reloading** - Update in-memory instead of full DB fetch
15. **Add CI/CD pipeline** - GitHub Actions for test/lint/build
16. **Performance profiling** - Use `pprof` to identify actual bottlenecks
17. **Consider DEFERRABLE constraints** - Simplify swap operations

### Long-Term (Nice to Have)

18. **API backend option** - Remote storage instead of local SQLite
19. **Plugin system** - Custom columns, task types, etc.
20. **Undo/redo** - Event sourcing or command pattern
21. **Multi-user sync** - Conflict resolution, operational transforms

---

## 10. ARCHITECTURE RECOMMENDATIONS

### Consider These Patterns

#### 10.1 Command Pattern for Operations
Instead of direct repo calls in handlers:

```go
type Command interface {
    Execute(ctx context.Context) error
    Undo(ctx context.Context) error
}

type MoveTaskCommand struct {
    repo   database.DataStore
    taskID int
    fromCol int
    toCol   int
}

func (c *MoveTaskCommand) Execute(ctx context.Context) error {
    return c.repo.MoveTask(ctx, c.taskID, c.toCol)
}
```

**Benefits:** Undo/redo, command history, testability

---

#### 10.2 Event-Driven State Updates
Instead of direct state mutations:

```go
type Event interface {
    Type() string
}

type TaskCreatedEvent struct {
    Task *models.Task
}

func (m Model) handleEvent(event Event) Model {
    switch e := event.(type) {
    case *TaskCreatedEvent:
        m.appState.AddTask(e.Task)
        m.notificationState.Success("Task created")
    }
    return m
}
```

**Benefits:** Audit log, event replay, easier debugging

---

#### 10.3 Repository Caching Layer
```go
type CachedRepo struct {
    underlying database.DataStore
    cache      *cache.Cache
}

func (r *CachedRepo) GetTaskSummaries(ctx context.Context, projectID int) ([]*models.TaskSummary, error) {
    key := fmt.Sprintf("tasks:%d", projectID)
    if cached, ok := r.cache.Get(key); ok {
        return cached.([]*models.TaskSummary), nil
    }

    tasks, err := r.underlying.GetTaskSummaries(ctx, projectID)
    if err == nil {
        r.cache.Set(key, tasks, 5*time.Minute)
    }
    return tasks, err
}
```

**Benefits:** Reduce database queries, faster UI

---

## FINAL VERDICT

### Strengths (What You Did Right)

1. **Clean architecture** - Proper layering and separation of concerns
2. **Smart query optimization** - GROUP_CONCAT, batch loading, N+1 prevention
3. **Idiomatic repository pattern** - Interface segregation, dependency injection
4. **Consistent error wrapping** - Proper use of `%w` throughout
5. **Pure rendering components** - Testable, reusable UI components
6. **Modern TUI stack** - Bubble Tea v2, layer-based modals
7. **Comprehensive features** - Multi-project, labels, relationships, search

### Critical Weaknesses

1. **Broken error handling** - String comparison instead of `errors.Is()`
2. **Missing bounds checks** - Slice operations could panic
3. **Low test coverage** - TUI layer at 25.9%, database at 59.3%
4. **No structured logging** - Debugging and monitoring difficult
5. **Tight coupling** - TUI directly depends on database package
6. **Code duplication** - Picker implementations repeat logic

### Performance Rating: **B** (7/10)
Acceptable for personal use, but O(n) linked list traversal and full column reloads are inefficient. Optimize before scaling to heavy usage.

### Idiomatic Go Rating: **B+** (8/10)
Mostly idiomatic with excellent repository pattern, error wrapping, and struct composition. Loses points for error comparison bug, missing bounds checks, and large files.

### Best Practices Rating: **B-** (7/10)
Follows architectural best practices (layering, interfaces, migrations) but missing CI/CD, linting, structured logging, and has test coverage gaps.

### Code Structure Rating: **A-** (9/10)
Excellent organization with clear layering, focused state objects, and pure components. Minor deductions for tight coupling and state mutation leaks.

---

## RECOMMENDED PRIORITY ORDER

### Week 1 (Critical Fixes)
1. Fix error comparison bug (`errors.Is()`)
2. Add bounds checking before slice operations
3. Fix TUI test imports
4. Add `.golangci.yml` and run linters

### Week 2 (Quality Improvements)
5. Implement structured logging (`slog`)
6. Add defensive copies to state getters
7. Split large files (800+ lines)
8. Increase test coverage to >60%

### Week 3 (Architecture)
9. Decouple TUI from database (interface)
10. Create generic picker abstraction
11. Add integration tests
12. Optimize column reloading strategy

### Ongoing
- Monitor Charm.sh v2 stable releases
- Profile performance with `pprof`
- Plan CI/CD pipeline
- Consider event-driven architecture for future features

---

**End of Review**

This is a solid foundation with production-quality database code and clean architecture. Address the critical bugs and testing gaps, and you'll have an excellent TUI application.


TODO: why does the terminal kind of crash when we try to open a task?
