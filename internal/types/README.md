# Type Aliases for ID Types

This package defines semantic type aliases for ID values throughout the Paso application. These aliases provide type safety, improve code readability, and enable future optimizations.

## Motivation

The codebase frequently converts between `int64` (from database/SQLC queries) and `int` (used in domain models). This creates several problems:

1. **Repetitive casting**: Dozens of `int()` conversions scatter throughout the codebase
2. **Semantic confusion**: It's unclear what each integer represents (is it a task ID, column ID, etc.?)
3. **Missed opportunities**: Type system can't help catch mistakes (passing a ProjectID where a ColumnID is expected)
4. **Difficult refactoring**: If we ever want to change ID types, we must manually refactor dozens of locations

## Solution: Type Aliases

We define semantic type aliases for each ID type:

```go
type ProjectID int
type ColumnID int
type TaskID int
type LabelID int
// ... etc
```

### Benefits

1. **Semantic clarity**: Code like `task.ProjectID` is clearer than `task.ProjectInt`
2. **Type safety**: Compiler helps catch mistakes where wrong ID type is passed
3. **Future-proof**: If we need 64-bit IDs, we can change the underlying type once
4. **Reversible**: Old code using `int` can coexist with new code using type aliases
5. **Zero runtime cost**: Type aliases have no runtime overhead in Go

## Usage Patterns

### Reading from Database

When SQLC returns `int64` values, convert immediately to semantic types:

```go
// Before: confusing int() casts
createdTask := &Task{
    ID:        int(row.ID),           // What kind of ID?
    ColumnID:  int(row.ColumnID),     // Which column?
    PriorityID: int(row.PriorityID),  // What's the priority?
}

// After: semantic type aliases
createdTask := &Task{
    ID:         types.TaskID(row.ID),
    ColumnID:   types.ColumnID(row.ColumnID),
    PriorityID: types.PriorityID(row.PriorityID),
}
```

### Known Constants

Use typed constants instead of magic numbers:

```go
// Before: magic number
if task.PriorityID == 4 {  // What does 4 mean?
    sendAlert()
}

// After: semantic constants
if task.PriorityID == types.PriorityHigh {
    sendAlert()
}
```

### Converting Back

For compatibility with legacy code, conversion methods are provided:

```go
id := types.TaskID(42)
intID := id.ToInt()  // Returns int(42)
```

Or use constructor functions:

```go
intID := 42
taskID := types.TaskIDFromInt(intID)
```

## Migration Strategy

This is a non-breaking addition:

1. **Phase 1** (Current): Introduce type aliases, keep all existing `int` types
2. **Phase 2** (Optional): Update hot paths to use type aliases
3. **Phase 3** (Optional): Migrate remaining code gradually
4. **Phase 4** (Optional): Remove legacy int types when fully migrated

Existing code using `int` continues to work unchanged. New code can use type aliases.

## Type Alias Hierarchy

```
ProjectID ─────┬─── ColumnID ──┬─── TaskID ──┬─── LabelID
               │               │             └─── CommentID
               │               │
               │               └─── TypeID
               │               └─── PriorityID
               │               └─── RelationTypeID
```

## Constants vs Functions

- **Constants**: For well-known values (PriorityHigh, TaskTypeFeature, etc.)
- **Functions**: For values computed from user input or database queries

## Performance Considerations

- Zero runtime cost: Type aliases compile to the underlying type
- No heap allocation: These are value types
- Inlining friendly: Conversions are inlined by the compiler
- No interface overhead: No dynamic dispatch involved

## Testing

Type aliases integrate seamlessly with existing test infrastructure:

```go
func TestTask(t *testing.T) {
    task := &Task{
        ID: types.TaskID(1),
        ColumnID: types.ColumnID(2),
    }

    if task.ID.ToInt() != 1 {
        t.Fail()
    }
}
```

## Future Extensions

This foundation enables type-safe wrappers for common patterns:

```go
// Future: strongly typed validation
func (id types.ProjectID) Valid() error {
    if id <= 0 {
        return errors.New("invalid project ID")
    }
    return nil
}
```

## Related Documents

- [CODING_CONVENTIONS.md](../CODING_CONVENTIONS.md) - Code quality standards
- [PERFORMANCE.md](../PERFORMANCE.md) - Performance guidelines
