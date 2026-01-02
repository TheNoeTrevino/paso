package task

import (
	"context"
	"database/sql"
	"testing"

	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/testutil"
)

// ============================================================================
// BENCHMARK SETUP HELPERS
// ============================================================================

// setupBenchmarkDB creates a benchmark database with test data
func setupBenchmarkDB(b *testing.B) *sql.DB {
	b.Helper()
	db := testutil.SetupTestDB(&testing.T{})
	return db
}

// createBenchmarkProject creates a project with columns for benchmarking
func createBenchmarkProject(b *testing.B, db *sql.DB) int {
	b.Helper()
	result, err := db.ExecContext(context.Background(), "INSERT INTO projects (name, description) VALUES (?, ?)", "Benchmark Project", "Description")
	if err != nil {
		b.Fatalf("Failed to create benchmark project: %v", err)
	}

	// Initialize project counter
	projectID, _ := result.LastInsertId()
	_, err = db.ExecContext(context.Background(), "INSERT INTO project_counters (project_id, next_ticket_number) VALUES (?, 1)", projectID)
	if err != nil {
		b.Fatalf("Failed to initialize project counter: %v", err)
	}

	return int(projectID)
}

// createBenchmarkTask creates a task with optional labels
func createBenchmarkTask(b *testing.B, db *sql.DB, columnID int, title string, labelIDs []int) int {
	b.Helper()

	// Get next position
	var maxPos sql.NullInt64
	err := db.QueryRowContext(context.Background(),
		"SELECT MAX(position) FROM tasks WHERE column_id = ?", columnID).Scan(&maxPos)
	if err != nil && err != sql.ErrNoRows {
		b.Fatalf("Failed to get max position: %v", err)
	}

	nextPos := 0
	if maxPos.Valid {
		nextPos = int(maxPos.Int64) + 1
	}

	result, err := db.ExecContext(context.Background(),
		"INSERT INTO tasks (column_id, title, position, type_id, priority_id) VALUES (?, ?, ?, 1, 3)",
		columnID, title, nextPos)
	if err != nil {
		b.Fatalf("Failed to create benchmark task: %v", err)
	}

	taskID, _ := result.LastInsertId()
	taskIDInt := int(taskID)

	// Attach labels if provided
	for _, labelID := range labelIDs {
		_, err := db.ExecContext(context.Background(),
			"INSERT INTO task_labels (task_id, label_id) VALUES (?, ?)", taskID, labelID)
		if err != nil {
			b.Fatalf("Failed to attach label: %v", err)
		}
	}

	return taskIDInt
}

// createBenchmarkLabel creates a label
func createBenchmarkLabel(b *testing.B, db *sql.DB, projectID int, name, color string) int {
	b.Helper()
	result, err := db.ExecContext(context.Background(),
		"INSERT INTO labels (project_id, name, color) VALUES (?, ?, ?)", projectID, name, color)
	if err != nil {
		b.Fatalf("Failed to create benchmark label: %v", err)
	}
	labelID, _ := result.LastInsertId()
	return int(labelID)
}

// createInProgressColumn creates an in-progress column
func createInProgressColumn(b *testing.B, db *sql.DB, projectID int) int {
	b.Helper()
	result, err := db.ExecContext(context.Background(),
		"INSERT INTO columns (project_id, name, holds_in_progress_tasks) VALUES (?, ?, ?)",
		projectID, "In Progress", true)
	if err != nil {
		b.Fatalf("Failed to create in-progress column: %v", err)
	}
	columnID, _ := result.LastInsertId()
	return int(columnID)
}

// createReadyColumn creates a ready column
func createReadyColumn(b *testing.B, db *sql.DB, projectID int) int {
	b.Helper()
	result, err := db.ExecContext(context.Background(),
		"INSERT INTO columns (project_id, name, holds_ready_tasks) VALUES (?, ?, ?)",
		projectID, "Ready", true)
	if err != nil {
		b.Fatalf("Failed to create ready column: %v", err)
	}
	columnID, _ := result.LastInsertId()
	return int(columnID)
}

// addCommentToTask adds a comment to a task
func addCommentToTask(b *testing.B, db *sql.DB, taskID int, author, message string) {
	b.Helper()
	_, err := db.ExecContext(context.Background(),
		"INSERT INTO task_comments (task_id, author, content) VALUES (?, ?, ?)",
		taskID, author, message)
	if err != nil {
		b.Fatalf("Failed to add comment: %v", err)
	}
}

// addBenchmarkTaskRelation adds a task relationship
func addBenchmarkTaskRelation(b *testing.B, db *sql.DB, parentID, childID, relationTypeID int) {
	b.Helper()
	_, err := db.ExecContext(context.Background(),
		"INSERT INTO task_subtasks (parent_id, child_id, relation_type_id) VALUES (?, ?, ?)",
		parentID, childID, relationTypeID)
	if err != nil {
		b.Fatalf("Failed to add task relation: %v", err)
	}
}

// ============================================================================
// BENCHMARK TESTS
// ============================================================================

// BenchmarkGetTaskDetail measures the performance of fetching a complete task detail
// This operation loads: task details, labels, parent tasks, child tasks, and comments
// The optimized version should efficiently fetch all related data
func BenchmarkGetTaskDetail(b *testing.B) {
	db := setupBenchmarkDB(b)
	defer db.Close()

	projectID := createBenchmarkProject(b, db)
	columnID := createBenchmarkColumn(b, db, projectID)

	// Create labels
	label1 := createBenchmarkLabel(b, db, projectID, "feature", "#0066FF")
	label2 := createBenchmarkLabel(b, db, projectID, "urgent", "#FF0000")

	// Create a task with labels
	taskID := createBenchmarkTask(b, db, columnID, "Test Task", []int{label1, label2})

	// Create parent and child tasks for relationships
	parentID := createBenchmarkTask(b, db, columnID, "Parent Task", []int{})
	childID := createBenchmarkTask(b, db, columnID, "Child Task", []int{})

	// Add relationships
	addBenchmarkTaskRelation(b, db, parentID, taskID, 1) // parent-child
	addBenchmarkTaskRelation(b, db, taskID, childID, 1)

	// Add comments
	addCommentToTask(b, db, taskID, "user1", "Great progress!")
	addCommentToTask(b, db, taskID, "user2", "Looks good to me")

	service := NewService(db, nil)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.GetTaskDetail(ctx, taskID)
		if err != nil {
			b.Fatalf("GetTaskDetail failed: %v", err)
		}
	}
}

// BenchmarkGetInProgressTasksByProject measures fetching all in-progress tasks
// After N+1 fix: should be a single optimized query instead of N+5 queries per task
// This is the most impactful optimization as it's a common operation
func BenchmarkGetInProgressTasksByProject(b *testing.B) {
	db := setupBenchmarkDB(b)
	defer db.Close()

	projectID := createBenchmarkProject(b, db)
	inProgressColumnID := createInProgressColumn(b, db, projectID)
	normalColumnID := createBenchmarkColumn(b, db, projectID)

	// Create labels for diversity
	label1 := createBenchmarkLabel(b, db, projectID, "frontend", "#FF00FF")
	label2 := createBenchmarkLabel(b, db, projectID, "backend", "#00FF00")
	label3 := createBenchmarkLabel(b, db, projectID, "database", "#FFFF00")

	// Create 50 in-progress tasks with varying label counts
	for i := 0; i < 50; i++ {
		title := "In Progress Task " + string(rune(i))
		var labelIDs []int
		switch i % 3 {
		case 0:
			labelIDs = []int{label1}
		case 1:
			labelIDs = []int{label1, label2}
		case 2:
			labelIDs = []int{label1, label2, label3}
		}
		createBenchmarkTask(b, db, inProgressColumnID, title, labelIDs)
	}

	// Create some tasks in other columns (should not be fetched)
	for i := 0; i < 30; i++ {
		title := "Other Task " + string(rune(i))
		createBenchmarkTask(b, db, normalColumnID, title, []int{})
	}

	service := NewService(db, nil)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.GetInProgressTasksByProject(ctx, projectID)
		if err != nil {
			b.Fatalf("GetInProgressTasksByProject failed: %v", err)
		}
	}
}

// BenchmarkGetTaskSummariesByProject measures fetching task summaries for kanban view
// This is the primary display operation for the board
// The optimized version efficiently groups tasks by column
func BenchmarkGetTaskSummariesByProject(b *testing.B) {
	db := setupBenchmarkDB(b)
	defer db.Close()

	projectID := createBenchmarkProject(b, db)

	// Create multiple columns
	columns := make([]int, 5)
	for i := 0; i < 5; i++ {
		colName := "Column " + string(rune('A'+i))
		columns[i] = createBenchmarkColumn(b, db, projectID)
		// Name the column (update via direct SQL since we don't have updateColumn)
		_, _ = db.ExecContext(context.Background(),
			"UPDATE columns SET name = ? WHERE id = ?", colName, columns[i])
	}

	// Create labels
	labels := make([]int, 4)
	labelColors := []string{"#FF0000", "#00FF00", "#0000FF", "#FFFF00"}
	for i := 0; i < 4; i++ {
		labels[i] = createBenchmarkLabel(b, db, projectID, "label"+string(rune('a'+i)), labelColors[i])
	}

	// Create 200 tasks distributed across columns with varying labels
	for col := 0; col < 5; col++ {
		for i := 0; i < 40; i++ {
			title := "Task " + string(rune(col)) + "-" + string(rune(i%10))
			labelCount := (i + col) % 4
			var labelIDs []int
			for j := 0; j <= labelCount; j++ {
				labelIDs = append(labelIDs, labels[j])
			}
			createBenchmarkTask(b, db, columns[col], title, labelIDs)
		}
	}

	service := NewService(db, nil)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.GetTaskSummariesByProject(ctx, projectID)
		if err != nil {
			b.Fatalf("GetTaskSummariesByProject failed: %v", err)
		}
	}
}

// BenchmarkGetTaskSummariesByProjectFiltered measures filtered task summary retrieval
// Tests the search/filter functionality on the board
func BenchmarkGetTaskSummariesByProjectFiltered(b *testing.B) {
	db := setupBenchmarkDB(b)
	defer db.Close()

	projectID := createBenchmarkProject(b, db)
	columnID := createBenchmarkColumn(b, db, projectID)

	// Create 100 tasks with varied names
	for i := 0; i < 100; i++ {
		title := "Database Migration Task " + string(rune(i%10))
		createBenchmarkTask(b, db, columnID, title, []int{})
	}

	service := NewService(db, nil)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.GetTaskSummariesByProjectFiltered(ctx, projectID, "Database")
		if err != nil {
			b.Fatalf("GetTaskSummariesByProjectFiltered failed: %v", err)
		}
	}
}

// BenchmarkGetTaskTreeByProject measures hierarchical task tree construction
// This is a complex recursive operation with multiple queries
// Performance depends on the depth and breadth of the task hierarchy
func BenchmarkGetTaskTreeByProject(b *testing.B) {
	db := setupBenchmarkDB(b)
	defer db.Close()

	projectID := createBenchmarkProject(b, db)
	columnID := createBenchmarkColumn(b, db, projectID)

	// Create a task hierarchy
	// 5 root tasks
	// Each root has 3 children
	// Each child has 2 grandchildren
	rootTasks := make([]int, 5)
	for i := 0; i < 5; i++ {
		rootTasks[i] = createBenchmarkTask(b, db, columnID, "Root Task "+string(rune(i)), []int{})

		childTasks := make([]int, 3)
		for j := 0; j < 3; j++ {
			childTasks[j] = createBenchmarkTask(b, db, columnID, "Child "+string(rune(i))+"-"+string(rune(j)), []int{})
			addBenchmarkTaskRelation(b, db, rootTasks[i], childTasks[j], 1)

			// Add grandchildren
			for k := 0; k < 2; k++ {
				grandchildID := createBenchmarkTask(b, db, columnID, "Grandchild "+string(rune(i))+"-"+string(rune(j))+"-"+string(rune(k)), []int{})
				addBenchmarkTaskRelation(b, db, childTasks[j], grandchildID, 1)
			}
		}
	}

	service := NewService(db, nil)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.GetTaskTreeByProject(ctx, projectID)
		if err != nil {
			b.Fatalf("GetTaskTreeByProject failed: %v", err)
		}
	}
}

// BenchmarkUpdateTask measures the performance of task updates
// Tests single field updates vs. multi-field updates
func BenchmarkUpdateTask(b *testing.B) {
	db := setupBenchmarkDB(b)
	defer db.Close()

	projectID := createBenchmarkProject(b, db)
	columnID := createBenchmarkColumn(b, db, projectID)
	taskID := createBenchmarkTask(b, db, columnID, "Original Title", []int{})

	service := NewService(db, nil)
	ctx := context.Background()

	newTitle := "Updated Title"
	newDesc := "Updated Description"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := service.UpdateTask(ctx, UpdateTaskRequest{
			TaskID:      taskID,
			Title:       &newTitle,
			Description: &newDesc,
		})
		if err != nil {
			b.Fatalf("UpdateTask failed: %v", err)
		}
	}
}

// BenchmarkCreateTask measures the performance of task creation
// Tests the full creation pipeline including label attachment
func BenchmarkCreateTask(b *testing.B) {
	db := setupBenchmarkDB(b)
	defer db.Close()

	projectID := createBenchmarkProject(b, db)
	columnID := createBenchmarkColumn(b, db, projectID)

	// Create labels to attach
	label1 := createBenchmarkLabel(b, db, projectID, "feature", "#FF0000")
	label2 := createBenchmarkLabel(b, db, projectID, "important", "#00FF00")

	service := NewService(db, nil)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.CreateTask(ctx, CreateTaskRequest{
			Title:       "Benchmark Task",
			Description: "This is a benchmark task for performance testing",
			ColumnID:    columnID,
			Position:    i,
			TypeID:      1,
			PriorityID:  3,
			LabelIDs:    []int{label1, label2},
		})
		if err != nil {
			b.Fatalf("CreateTask failed: %v", err)
		}
	}
}

// BenchmarkAttachLabel measures label attachment performance
// This is frequently used when editing tasks
func BenchmarkAttachLabel(b *testing.B) {
	db := setupBenchmarkDB(b)
	defer db.Close()

	projectID := createBenchmarkProject(b, db)
	columnID := createBenchmarkColumn(b, db, projectID)
	taskID := createBenchmarkTask(b, db, columnID, "Test Task", []int{})

	// Create multiple labels
	labels := make([]int, b.N)
	for i := 0; i < b.N; i++ {
		labels[i] = createBenchmarkLabel(b, db, projectID, "label_"+string(rune(i)), "#FF00FF")
	}

	service := NewService(db, nil)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := service.AttachLabel(ctx, taskID, labels[i])
		if err != nil {
			b.Fatalf("AttachLabel failed: %v", err)
		}
	}
}

// BenchmarkDetachLabel measures label detachment performance
func BenchmarkDetachLabel(b *testing.B) {
	db := setupBenchmarkDB(b)
	defer db.Close()

	projectID := createBenchmarkProject(b, db)
	columnID := createBenchmarkColumn(b, db, projectID)
	taskID := createBenchmarkTask(b, db, columnID, "Test Task", []int{})

	// Create and attach multiple labels
	labels := make([]int, b.N)
	for i := 0; i < b.N; i++ {
		labelID := createBenchmarkLabel(b, db, projectID, "label_"+string(rune(i)), "#FF00FF")
		labels[i] = labelID
		_, _ = db.ExecContext(context.Background(),
			"INSERT INTO task_labels (task_id, label_id) VALUES (?, ?)", taskID, labelID)
	}

	service := NewService(db, nil)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := service.DetachLabel(ctx, taskID, labels[i])
		if err != nil {
			b.Fatalf("DetachLabel failed: %v", err)
		}
	}
}

// BenchmarkMoveTaskToColumn measures column movement performance
// This is frequent during task workflow changes
func BenchmarkMoveTaskToColumn(b *testing.B) {
	db := setupBenchmarkDB(b)
	defer db.Close()

	projectID := createBenchmarkProject(b, db)
	columnID := createBenchmarkColumn(b, db, projectID)
	targetColumn := createBenchmarkColumn(b, db, projectID)

	// Create multiple tasks to move
	tasks := make([]int, b.N)
	for i := 0; i < b.N; i++ {
		tasks[i] = createBenchmarkTask(b, db, columnID, "Move Task "+string(rune(i)), []int{})
	}

	service := NewService(db, nil)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := service.MoveTaskToColumn(ctx, tasks[i], targetColumn)
		if err != nil {
			b.Fatalf("MoveTaskToColumn failed: %v", err)
		}
	}
}

// BenchmarkCreateComment measures comment creation performance
func BenchmarkCreateComment(b *testing.B) {
	db := setupBenchmarkDB(b)
	defer db.Close()

	projectID := createBenchmarkProject(b, db)
	columnID := createBenchmarkColumn(b, db, projectID)
	taskID := createBenchmarkTask(b, db, columnID, "Test Task", []int{})

	service := NewService(db, nil)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.CreateComment(ctx, CreateCommentRequest{
			TaskID:  taskID,
			Message: "This is a benchmark comment",
			Author:  "benchmark_user",
		})
		if err != nil {
			b.Fatalf("CreateComment failed: %v", err)
		}
	}
}

// BenchmarkGetReadyTaskSummariesByProject measures fetching ready tasks (for "Start Work" flow)
func BenchmarkGetReadyTaskSummariesByProject(b *testing.B) {
	db := setupBenchmarkDB(b)
	defer db.Close()

	projectID := createBenchmarkProject(b, db)
	readyColumnID := createReadyColumn(b, db, projectID)
	normalColumnID := createBenchmarkColumn(b, db, projectID)

	// Create 100 ready tasks
	for i := 0; i < 100; i++ {
		title := "Ready Task " + string(rune(i%10))
		createBenchmarkTask(b, db, readyColumnID, title, []int{})
	}

	// Create some normal tasks (should not be fetched)
	for i := 0; i < 50; i++ {
		title := "Normal Task " + string(rune(i%10))
		createBenchmarkTask(b, db, normalColumnID, title, []int{})
	}

	service := NewService(db, nil)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.GetReadyTaskSummariesByProject(ctx, projectID)
		if err != nil {
			b.Fatalf("GetReadyTaskSummariesByProject failed: %v", err)
		}
	}
}

// BenchmarkAddParentRelation measures adding parent-child relationships
func BenchmarkAddParentRelation(b *testing.B) {
	db := setupBenchmarkDB(b)
	defer db.Close()

	projectID := createBenchmarkProject(b, db)
	columnID := createBenchmarkColumn(b, db, projectID)

	// Create parent task
	parentID := createBenchmarkTask(b, db, columnID, "Parent Task", []int{})

	// Create child tasks
	childTasks := make([]int, b.N)
	for i := 0; i < b.N; i++ {
		childTasks[i] = createBenchmarkTask(b, db, columnID, "Child Task "+string(rune(i)), []int{})
	}

	service := NewService(db, nil)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := service.AddParentRelation(ctx, childTasks[i], parentID, models.RelationTypeParentChild)
		if err != nil {
			b.Fatalf("AddParentRelation failed: %v", err)
		}
	}
}

// BenchmarkGetTaskReferencesForProject measures fetching all task references for a project
// This is used in the task linking interface
func BenchmarkGetTaskReferencesForProject(b *testing.B) {
	db := setupBenchmarkDB(b)
	defer db.Close()

	projectID := createBenchmarkProject(b, db)
	columnID := createBenchmarkColumn(b, db, projectID)

	// Create 200 tasks
	for i := 0; i < 200; i++ {
		title := "Task " + string(rune(i%10))
		createBenchmarkTask(b, db, columnID, title, []int{})
	}

	service := NewService(db, nil)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.GetTaskReferencesForProject(ctx, projectID)
		if err != nil {
			b.Fatalf("GetTaskReferencesForProject failed: %v", err)
		}
	}
}

// Helper function to create benchmark column
func createBenchmarkColumn(b *testing.B, db *sql.DB, projectID int) int {
	b.Helper()
	result, err := db.ExecContext(context.Background(),
		"INSERT INTO columns (project_id, name) VALUES (?, ?)", projectID, "Column")
	if err != nil {
		b.Fatalf("Failed to create benchmark column: %v", err)
	}
	columnID, _ := result.LastInsertId()
	return int(columnID)
}
