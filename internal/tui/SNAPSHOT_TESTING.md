# TUI Snapshot Testing

This directory contains snapshot tests for the Paso TUI (Terminal User Interface) layer using Bubble Tea's teatest framework.

## Overview

Snapshot tests verify that TUI rendering produces consistent output across different application states and modes. They use golden files to store baseline snapshots and detect regressions.

## Key Patterns

### Color Profile Standardization

All snapshot tests use ASCII color profile for CI/CD consistency:
```go
lipgloss.SetColorProfile(termenv.Ascii)
```

This prevents test failures when running in environments with different terminal color capabilities.

### Golden Files

Golden files store baseline snapshots in `testdata/snapshots/` directory:
- Files use `.golden` extension
- Git attributes ensure consistent line endings: `*.golden -text`
- Files are committed to version control as baseline references

### Test Execution

Run snapshot tests normally to compare against baselines:
```bash
go test ./internal/tui -run TestSnapshots
```

This will fail if output differs from baseline, showing the expected vs actual output.

## Updating Snapshots

When intentionally changing TUI rendering (new feature, styling update), update snapshots:

```bash
UPDATE_SNAPSHOTS=1 go test ./internal/tui -run TestSnapshots
```

This will create/update all golden files with current output.

## Test Coverage

Current snapshot tests cover:

1. **empty_project** - Default view with no tasks
2. **board_with_tasks** - Kanban board with multiple tasks across columns
3. **board_with_labels** - Tasks with labels rendered on board
4. **task_form_modal** - Task creation/editing form overlay
5. **priority_picker_modal** - Priority selection modal
6. **notification_overlay** - Success/error notification display

## Adding New Snapshots

To add a new snapshot test:

1. Add setup function in `snapshots_test.go`:
```go
{
    name: "new_feature",
    setup: func(t *testing.T, db *sql.DB) Model {
        return setupNewFeature(t, db)
    },
}
```

2. Run with UPDATE_SNAPSHOTS to create baseline:
```bash
UPDATE_SNAPSHOTS=1 go test ./internal/tui -run TestSnapshots/new_feature
```

3. Review the generated golden file to ensure it's correct
4. Commit both test code and golden file

## Terminal Dimensions

Snapshots use fixed terminal size (80x24) for consistency:
- Width: 80 characters (standard terminal width)
- Height: 24 lines (standard terminal height)

This ensures snapshots are reproducible across different test environments.

## CI/CD Integration

Snapshot tests in CI should fail if output differs from baseline, catching:
- Unintended visual regressions
- Formatting/styling changes
- Component rendering issues

Run with explicit comparison (no UPDATE_SNAPSHOTS):
```bash
go test ./internal/tui -run TestSnapshots -v
```

## Implementation Details

### SnapshotHelper

`snapshot_helper.go` provides utilities:
- `NewSnapshotHelper()` - Create helper with UPDATE_SNAPSHOTS detection
- `Compare()` - Compare output to golden file
- `WriteSnapshot()` - Write snapshot file directly
- `ReadSnapshot()` - Read existing snapshot

### Test Model

`testModel` struct wraps teatest functionality:
- `RunUntil()` - Run model until condition met or timeout
- `FinalOutput()` - Get rendered view output

## Related Documents

- [PERFORMANCE.md](../../PERFORMANCE.md) - Performance benchmarking
- [CODING_CONVENTIONS.md](../../CODING_CONVENTIONS.md) - Code quality standards
- Bubble Tea Testing: https://carlosbecker.com/posts/teatest/
