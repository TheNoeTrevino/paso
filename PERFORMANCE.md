# Performance Benchmarks

This document tracks performance improvements and benchmarks for critical paths in the Paso application.

## Benchmark Results

Run benchmarks with:
```bash
go test -bench=. -benchmem ./internal/services/task
```

### Latest Benchmark Results

Baseline measurements with optimizations (N+1 fixes, indexing):

| Operation | ns/op | allocs/op | B/op | Notes |
|-----------|-------|-----------|------|-------|
| GetTaskDetail | 262,313 | 256 | 9,096 | Full task with labels, comments, relationships |
| GetInProgressTasksByProject | 1,020,784 | 3,016 | 149,673 | Single optimized query (was N+5 queries) |
| GetTaskSummariesByProject | 2,920,841 | 6,819 | 294,788 | Kanban board display (200 tasks, 5 columns) |
| GetTaskSummariesByProjectFiltered | 971,773 | 1,444 | 96,192 | Search query (100 tasks) |
| GetTaskTreeByProject | 346,328 | 1,105 | 59,208 | Hierarchical tree (35 tasks, 3 levels) |
| UpdateTask | 17,168 | 11 | 328 | Update task title/description |
| CreateTask | 247,441 | 147 | 5,118 | Create task with labels (2 labels) |
| AttachLabel | 31,173 | 6 | 191 | Label attachment |
| DetachLabel | 31,970 | 6 | 191 | Label detachment |
| MoveTaskToColumn | 469,585 | 39 | 1,230 | Move task to different column |
| CreateComment | 85,009 | 97 | 3,584 | Add comment to task |
| GetReadyTaskSummariesByProject | 799,769 | 1,432 | 92,120 | "Start Work" flow (100 ready tasks) |
| AddParentRelation | 44,195 | 6 | 255 | Add parent-child relationship |
| GetTaskReferencesForProject | 464,984 | 1,626 | 81,328 | Task references for linking (200 tasks) |

## Performance Optimizations

### N+1 Query Fix: GetInProgressTasksByProject
**Status:** Implemented

**Before:** Called GetTaskDetail for each task, which made 5+ additional queries per task
- N tasks + (N × 5+) queries = O(N*5+) complexity
- 50 in-progress tasks = 250+ queries

**After:** Single optimized query fetches all details in one query
- Single query with aggregated label data = O(1)
- 50 in-progress tasks = 1 query
- **Improvement:** 99.6% reduction in database queries

### Redundant Query Fix: UpdateTask
**Status:** Implemented

**Before:** Unnecessarily called GetTaskDetail when only updating title/description
**After:** Only queries existing values when needed, optimized conditional logic
- **Improvement:** Reduced from multiple queries to 1-2 queries

### Memory Optimization: Allocation Patterns
**Status:** Baseline established

**Current behavior:**
- PreAllocate slices where size is known
- Use pool allocators for frequently created objects
- Minimize intermediate allocations in loops

## Database Indexes

The following indexes support these operations:

```sql
-- Primary composite indexes
CREATE INDEX idx_tasks_column ON tasks(column_id, position);
CREATE INDEX idx_columns_project ON columns(project_id);
CREATE INDEX idx_labels_project ON labels(project_id);
CREATE INDEX idx_task_labels_label ON task_labels(label_id);
CREATE INDEX idx_task_subtasks_parent ON task_subtasks(parent_id);
CREATE INDEX idx_task_subtasks_child ON task_subtasks(child_id);
CREATE INDEX idx_task_comments_task ON task_comments(task_id);

-- Unique partial indexes for column type queries
CREATE UNIQUE INDEX idx_columns_ready_per_project ON columns(project_id)
  WHERE holds_ready_tasks = 1;
CREATE UNIQUE INDEX idx_columns_completed_per_project ON columns(project_id)
  WHERE holds_completed_tasks = 1;
CREATE UNIQUE INDEX idx_columns_in_progress_per_project ON columns(project_id)
  WHERE holds_in_progress_tasks = 1;

-- Additional performance indexes
CREATE INDEX idx_tasks_column_id ON tasks(column_id);
CREATE INDEX idx_task_labels_task_id ON task_labels(task_id);
CREATE INDEX idx_labels_project_id ON labels(project_id);
CREATE INDEX idx_columns_project_id ON columns(project_id);
CREATE INDEX idx_task_subtasks_child_id ON task_subtasks(child_id);
CREATE INDEX idx_task_comments_task_id ON task_comments(task_id);
CREATE INDEX idx_tasks_type_id ON tasks(type_id);
CREATE INDEX idx_tasks_priority_id ON tasks(priority_id);
```

## Critical Paths

### User-Facing Operations (Should be <500ms)
- GetTaskDetail: ~260µs ✓
- GetInProgressTasksByProject: ~1ms ✓
- GetTaskSummariesByProject: ~2.9ms ✓
- GetReadyTaskSummariesByProject: ~800µs ✓

### Workflow Operations (Should be <1s)
- MoveTaskToColumn: ~470µs ✓
- UpdateTask: ~17µs ✓
- CreateTask: ~250µs ✓

## Profiling Instructions

### CPU Profiling
```bash
# Run with CPU profiling
go test -cpuprofile=cpu.prof -bench=BenchmarkGetInProgressTasksByProject \
  -benchtime=10s ./internal/services/task

# Analyze results
go tool pprof cpu.prof
```

### Memory Profiling
```bash
# Run with memory profiling
go test -memprofile=mem.prof -bench=BenchmarkGetTaskSummariesByProject \
  -benchtime=10s ./internal/services/task

# Analyze results
go tool pprof mem.prof
```

### Trace Profiling
```bash
# Generate execution trace
go test -trace=trace.out -bench=. ./internal/services/task

# View trace
go tool trace trace.out
```

## Regression Testing

To track performance regressions:

```bash
# Establish baseline
go test -bench=. -benchmem ./internal/services/task > baseline.txt

# After changes
go test -bench=. -benchmem ./internal/services/task > current.txt

# Compare (using benchstat)
benchstat baseline.txt current.txt
```

### Installation of benchstat
```bash
go install golang.org/x/perf/cmd/benchstat@latest
```

## Query Count Verification

To verify query counts match optimization claims:

1. Enable SQL logging in tests:
```go
db, _ := sql.Open("sqlite", ":memory:")
// Add query logger
```

2. Run specific benchmark with logging
3. Count SQL statements in output

## Future Optimizations

### Possible Improvements
- [ ] Implement connection pooling tuning for parallel operations
- [ ] Add caching layer for frequently accessed task references
- [ ] Consider query result pagination for large projects
- [ ] Optimize tree building with iterative approach instead of recursive
- [ ] Profile and optimize label aggregation query

### Monitoring Suggestions
- Track p99 latency for GetTaskSummariesByProject
- Monitor memory allocations under sustained load
- Set alerts for queries exceeding baseline by >20%

## Related Documents

- [CODING_CONVENTIONS.md](./CODING_CONVENTIONS.md) - Code quality standards
- [CLI_IMPLEMENTATION_GUIDE.md](./CLI_IMPLEMENTATION_GUIDE.md) - CLI performance guidelines
- [AI_INTEGRATION.md](./AI_INTEGRATION.md) - Integration patterns

## Historical Performance Notes

### Sprint 4 Optimizations
- Fixed N+1 query problem in GetInProgressTasksByProject
- Added database indexes for common queries
- Optimized UpdateTask to reduce unnecessary queries
- Established baseline benchmarks for critical paths
