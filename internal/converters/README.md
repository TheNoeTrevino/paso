# Converters Package

The `converters` package provides centralized conversion functions for transforming database-generated types to domain models. This eliminates duplication across service packages and ensures consistent conversion patterns.

## Overview

This package consolidates conversion logic from multiple services:
- `task.go` - Task and task-related conversions
- `column.go` - Column conversions
- `label.go` - Label conversions

## Structure

### Task Conversions (`task.go`)

**Single Entity Conversions:**
- `TaskToModel(t generated.Task) *models.Task` - Converts database task to domain model
- `CommentsToModels(comments []generated.TaskComment) []*models.Comment` - Converts comment rows to domain models

**Relationship Conversions:**
- `ParentTasksToReferences(rows []generated.GetParentTasksRow) []*models.TaskReference` - Converts parent task relationships
- `ChildTasksToReferences(rows []generated.GetChildTasksRow) []*models.TaskReference` - Converts child task relationships

**Summary Conversions:**
- `TaskSummaryFromRowToModel(row generated.GetTaskSummariesByProjectRow) *models.TaskSummary` - Converts standard task summary
- `ReadyTaskSummaryFromRowToModel(row generated.GetReadyTaskSummariesByProjectRow) *models.TaskSummary` - Converts ready task summary
- `FilteredTaskSummaryFromRowToModel(row generated.GetTaskSummariesByProjectFilteredRow) *models.TaskSummary` - Converts filtered task summary

**Utility Functions:**
- `ParseLabelsFromConcatenated(ids, names, colors string) []*models.Label` - Parses GROUP_CONCAT label data

### Column Conversions (`column.go`)

- `ColumnToModel(c generated.Column) *models.Column` - Converts generated column to domain model
- `ColumnFromIDRowToModel(r generated.GetColumnByIDRow) *models.Column` - Converts column by ID query result
- `ColumnsFromRowsToModels(rows []generated.GetColumnsByProjectRow) []*models.Column` - Converts multiple columns from project query

### Label Conversions (`label.go`)

- `LabelToModel(l generated.Label) *models.Label` - Converts single label to domain model
- `LabelsToModels(labels []generated.Label) []*models.Label` - Converts label slice to domain models

## Patterns

### Single Entity Conversion
```go
func TaskToModel(t generated.Task) *models.Task {
    task := &models.Task{
        ID:         int(t.ID),
        Title:      t.Title,
        // ... other fields
    }

    // Handle optional fields
    if t.Description.Valid {
        task.Description = t.Description.String
    }

    return task
}
```

### Slice Conversion
```go
func LabelsToModels(labels []generated.Label) []*models.Label {
    result := make([]*models.Label, len(labels))
    for i, l := range labels {
        result[i] = LabelToModel(l)
    }
    return result
}
```

### Row-Specific Conversion
Different database queries may return the same data in different row types (e.g., `GetColumnByIDRow` vs `GetColumnsByProjectRow`). Separate converter functions handle these cases while maintaining consistent output.

## Type Conversion Considerations

### SQL null.* types
When converting from SQLC-generated types with `null.String` or `null.Time`:
- Check the `.Valid` field before accessing the value
- Only set the domain model field if `.Valid` is true

### Type Casting
All ID fields are cast from `int64` to `int`:
- Database: `int64` (SQLC default)
- Domain models: `int` (API convenience)

### Special Cases

**InterfaceToIntPtr**: The `database.InterfaceToIntPtr()` helper is used for converting nullable pointer fields in columns:
```go
PrevID: database.InterfaceToIntPtr(c.PrevID),
NextID: database.InterfaceToIntPtr(c.NextID),
```

**GROUP_CONCAT Parsing**: Task summaries include concatenated label data that must be parsed:
```go
Labels: ParseLabelsFromConcatenated(row.LabelIds, row.LabelNames, row.LabelColors)
```

## Usage

Import the converters package in your service:
```go
import "github.com/thenoetrevino/paso/internal/converters"
```

Use in your service methods:
```go
func (s *service) GetLabel(ctx context.Context, id int) (*models.Label, error) {
    label, err := s.queries.GetLabelByID(ctx, int64(id))
    if err != nil {
        return nil, err
    }
    return converters.LabelToModel(label), nil
}
```

## Benefits

1. **DRY Principle**: Single source of truth for each conversion
2. **Consistency**: All services use identical conversion logic
3. **Maintainability**: Changes to conversion logic in one place
4. **Testability**: Conversion logic can be unit tested independently
5. **Type Safety**: Explicit type conversions with proper null handling
