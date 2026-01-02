## 8. SPECIFIC IMPROVEMENTS & RECOMMENDATIONS

### HIGH PRIORITY (Industry Standard Improvements)

#### 8.1 CLI Command Test Coverage
**Current**: 6.5% (task), 14.6% (project)
**Target**: 40%+

**Recommendation**:
```go
// Example: Create integration tests for common workflows
func TestTaskCreate(t *testing.T) {
    // Setup test DB
    db := setupTestDB(t)

    // Execute CLI command with mock args
    app := createTestApp(db)
    err := task.CreateTask(app, []string{"--title", "Test Task"})

    // Verify task created
    assert.NoError(t, err)
    // Verify output format
}
```

**Benefit**: Catch CLI arg parsing issues, output formatting bugs early.

#### 8.2 Error Path Testing
**Current**: Many error paths untested
**Recommendation**: Add negative test cases for:
- Invalid column IDs
- Circular relationships
- Duplicate labels
- Permission/authorization (if applicable)

**Example**:
```go
func TestCreateTaskInvalidColumn(t *testing.T) {
    service := setupService()
    _, err := service.CreateTask(ctx, CreateTaskRequest{
        ColumnID: 9999, // Non-existent
    })
    assert.ErrorIs(t, err, models.ErrInvalidColumnID)
}
```

### MEDIUM PRIORITY (Code Organization)

#### 8.3 TUI Update Functions - Decomposition
**Issue**: `update_forms.go` (855 lines), `update_pickers.go` (858 lines)

**Current Structure**:
```
update_forms.go     855 lines (many form handlers)
update_pickers.go   858 lines (many picker handlers)
```

**Recommendation**: Further decompose by feature:
```
update_forms/
├── task_form.go
├── project_form.go
├── column_form.go
└── comment_form.go

update_pickers/
├── label_picker.go
├── parent_picker.go
├── child_picker.go
└── priority_picker.go
```

**Benefits**:
- Each file ~150-200 lines (more manageable)
- Easier to locate handler logic
- Reduced cognitive load when modifying forms
- Better for parallel development

#### 8.4 Converters Package Documentation
**Issue**: Converters module lacks clear documentation

**Recommendation**: Add doc comments explaining conversion strategy:
```go
// Package converters provides type-safe conversion between
// database models (from SQLC) and domain models.
//
// All conversions handle:
// - NULL database values (sql.Null* types)
// - Type coercions (int64 from database to int in domain)
// - Relationship parsing (GROUP_CONCAT labels)
//
// Conversion failures are explicit - never silent type coercions.
package converters
```

**Benefit**: Future maintainers understand why converters exist and how to add new ones.

#### 8.5 Service Interface Documentation
**Current**: Good but could be better

**Recommendation**: Add usage examples in doc comments:
```go
// TaskReader defines read-only operations for retrieving task data.
//
// Example usage:
//
//    reader := app.TaskService.(taskservice.TaskReader)
//    task, err := reader.GetTaskDetail(ctx, 42)
//    if err != nil {
//        return err
//    }
type TaskReader interface {
    GetTaskDetail(ctx context.Context, taskID int) (*models.TaskDetail, error)
    // ...
}
```

### LOW PRIORITY (Polish)

#### 8.6 Logging Consistency
**Status**: Good, but could be more consistent

**Observation**: Mix of contextual logging styles:
```go
// Style 1
slog.Error("Failed to enable foreign keys", "error", err)

// Style 2
slog.Error("Error loading projects", "error", err)

// Style 3
slog.Info("received refresh event", "event_project_id", msg.Event.ProjectID, "current_project_id", currentProjectID)
```

**Recommendation**: Adopt consistent pattern:
```go
// Consistent: Verb + What + Error
slog.Error("failed to enable foreign keys", "error", err)
slog.Error("failed to load projects", "error", err)
slog.Info("received refresh event", "event_project_id", msg.Event.ProjectID, "current_project_id", currentProjectID)
```

**Benefit**: Easier log filtering and parsing.

#### 8.7 Configuration Validation
**Status**: Works, but no schema validation

**Recommendation**: Add config schema validation at load time:
```go
func (c *Config) Validate() error {
    if c.Theme == "" {
        return errors.New("theme cannot be empty")
    }
    if len(c.Keymaps) == 0 {
        return errors.New("keymaps cannot be empty")
    }
    return nil
}
```

**Benefit**: Catch bad configs early with helpful error messages.

#### 8.8 Handler Consistency
**Status**: Good in general, but some variation

**Observation**: CLI handlers use different patterns for option parsing

**Recommendation**: Standardize on one approach (likely the most common one already in use)

