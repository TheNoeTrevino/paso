package task

import (
	"context"
	"database/sql"
	"testing"

	"github.com/thenoetrevino/paso/internal/database"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/testing/fixtures"
)

// newBenchmarkService creates a new service for benchmarking
func newBenchmarkService(b *testing.B, db *sql.DB) Service {
	b.Helper()
	svc, err := NewService(db, database.SQLite, nil, nil)
	if err != nil {
		b.Fatalf("failed to create benchmark service: %v", err)
	}
	return svc
}

// createBenchmarkTaskWithLabels creates a task and attaches the given labels.
// Use fixtures.CreateTestTask directly when no labels are needed.
func createBenchmarkTaskWithLabels(b *testing.B, db *sql.DB, columnID int, title string, labelIDs []int) int {
	b.Helper()
	taskID := fixtures.CreateTestTask(b, db, testDialect, columnID, title)
	for _, labelID := range labelIDs {
		fixtures.AttachLabelToTask(b, db, testDialect, taskID, labelID)
	}
	return taskID
}

func BenchmarkGetTaskDetail(b *testing.B) {
	db := fixtures.SetupTestDB(b)

	projectID := fixtures.CreateBareProject(b, db, testDialect, "Benchmark Project")
	columnID := fixtures.CreateTestColumn(b, db, testDialect, projectID, "Column")

	label1 := fixtures.CreateTestLabel(b, db, testDialect, projectID, "feature", "#0066FF")
	label2 := fixtures.CreateTestLabel(b, db, testDialect, projectID, "urgent", "#FF0000")

	taskID := createBenchmarkTaskWithLabels(b, db, columnID, "Test Task", []int{label1, label2})

	parentID := fixtures.CreateTestTask(b, db, testDialect, columnID, "Parent Task")
	childID := fixtures.CreateTestTask(b, db, testDialect, columnID, "Child Task")

	fixtures.AddTaskSubtask(b, db, testDialect, parentID, taskID, 1)
	fixtures.AddTaskSubtask(b, db, testDialect, taskID, childID, 1)

	fixtures.CreateTestComment(b, db, testDialect, taskID, "Great progress!", "user1")
	fixtures.CreateTestComment(b, db, testDialect, taskID, "Looks good to me", "user2")

	service := newBenchmarkService(b, db)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.GetTaskDetail(ctx, taskID)
		if err != nil {
			b.Fatalf("GetTaskDetail failed: %v", err)
		}
	}
}

func BenchmarkGetInProgressTasksByProject(b *testing.B) {
	db := fixtures.SetupTestDB(b)

	projectID := fixtures.CreateBareProject(b, db, testDialect, "Benchmark Project")
	inProgressColumnID := fixtures.CreateColumnWithFlags(b, db, testDialect, projectID, "In Progress", false, false, true)
	normalColumnID := fixtures.CreateTestColumn(b, db, testDialect, projectID, "Column")

	label1 := fixtures.CreateTestLabel(b, db, testDialect, projectID, "frontend", "#FF00FF")
	label2 := fixtures.CreateTestLabel(b, db, testDialect, projectID, "backend", "#00FF00")
	label3 := fixtures.CreateTestLabel(b, db, testDialect, projectID, "database", "#FFFF00")

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
		createBenchmarkTaskWithLabels(b, db, inProgressColumnID, title, labelIDs)
	}

	for i := 0; i < 30; i++ {
		title := "Other Task " + string(rune(i))
		fixtures.CreateTestTask(b, db, testDialect, normalColumnID, title)
	}

	service := newBenchmarkService(b, db)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.GetInProgressTasksByProject(ctx, projectID)
		if err != nil {
			b.Fatalf("GetInProgressTasksByProject failed: %v", err)
		}
	}
}

func BenchmarkGetTaskSummariesByProject(b *testing.B) {
	db := fixtures.SetupTestDB(b)

	projectID := fixtures.CreateBareProject(b, db, testDialect, "Benchmark Project")

	columns := make([]int, 5)
	for i := 0; i < 5; i++ {
		colName := "Column " + string(rune('A'+i))
		columns[i] = fixtures.CreateTestColumn(b, db, testDialect, projectID, colName)
	}

	labels := make([]int, 4)
	labelColors := []string{"#FF0000", "#00FF00", "#0000FF", "#FFFF00"}
	for i := 0; i < 4; i++ {
		labels[i] = fixtures.CreateTestLabel(b, db, testDialect, projectID, "label"+string(rune('a'+i)), labelColors[i])
	}

	for col := 0; col < 5; col++ {
		for i := 0; i < 40; i++ {
			title := "Task " + string(rune(col)) + "-" + string(rune(i%10))
			labelCount := (i + col) % 4
			var labelIDs []int
			for j := 0; j <= labelCount; j++ {
				labelIDs = append(labelIDs, labels[j])
			}
			createBenchmarkTaskWithLabels(b, db, columns[col], title, labelIDs)
		}
	}

	service := newBenchmarkService(b, db)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.GetTaskSummariesByProject(ctx, projectID)
		if err != nil {
			b.Fatalf("GetTaskSummariesByProject failed: %v", err)
		}
	}
}

func BenchmarkGetTaskSummariesByProjectFiltered(b *testing.B) {
	db := fixtures.SetupTestDB(b)

	projectID := fixtures.CreateBareProject(b, db, testDialect, "Benchmark Project")
	columnID := fixtures.CreateTestColumn(b, db, testDialect, projectID, "Column")

	for i := 0; i < 100; i++ {
		title := "Database Migration Task " + string(rune(i%10))
		fixtures.CreateTestTask(b, db, testDialect, columnID, title)
	}

	service := newBenchmarkService(b, db)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.GetTaskSummariesByProjectFiltered(ctx, projectID, "Database")
		if err != nil {
			b.Fatalf("GetTaskSummariesByProjectFiltered failed: %v", err)
		}
	}
}

func BenchmarkGetTaskTreeByProject(b *testing.B) {
	db := fixtures.SetupTestDB(b)

	projectID := fixtures.CreateBareProject(b, db, testDialect, "Benchmark Project")
	columnID := fixtures.CreateTestColumn(b, db, testDialect, projectID, "Column")

	rootTasks := make([]int, 5)
	for i := 0; i < 5; i++ {
		rootTasks[i] = fixtures.CreateTestTask(b, db, testDialect, columnID, "Root Task "+string(rune(i)))

		childTasks := make([]int, 3)
		for j := 0; j < 3; j++ {
			childTasks[j] = fixtures.CreateTestTask(b, db, testDialect, columnID, "Child "+string(rune(i))+"-"+string(rune(j)))
			fixtures.AddTaskSubtask(b, db, testDialect, rootTasks[i], childTasks[j], 1)

			for k := 0; k < 2; k++ {
				grandchildID := fixtures.CreateTestTask(b, db, testDialect, columnID, "Grandchild "+string(rune(i))+"-"+string(rune(j))+"-"+string(rune(k)))
				fixtures.AddTaskSubtask(b, db, testDialect, childTasks[j], grandchildID, 1)
			}
		}
	}

	service := newBenchmarkService(b, db)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.GetTaskTreeByProject(ctx, projectID)
		if err != nil {
			b.Fatalf("GetTaskTreeByProject failed: %v", err)
		}
	}
}

func BenchmarkUpdateTask(b *testing.B) {
	db := fixtures.SetupTestDB(b)

	projectID := fixtures.CreateBareProject(b, db, testDialect, "Benchmark Project")
	columnID := fixtures.CreateTestColumn(b, db, testDialect, projectID, "Column")
	taskID := fixtures.CreateTestTask(b, db, testDialect, columnID, "Original Title")

	service := newBenchmarkService(b, db)
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

func BenchmarkCreateTask(b *testing.B) {
	db := fixtures.SetupTestDB(b)

	projectID := fixtures.CreateBareProject(b, db, testDialect, "Benchmark Project")
	columnID := fixtures.CreateTestColumn(b, db, testDialect, projectID, "Column")

	label1 := fixtures.CreateTestLabel(b, db, testDialect, projectID, "feature", "#FF0000")
	label2 := fixtures.CreateTestLabel(b, db, testDialect, projectID, "important", "#00FF00")

	service := newBenchmarkService(b, db)
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

func BenchmarkAttachLabel(b *testing.B) {
	db := fixtures.SetupTestDB(b)

	projectID := fixtures.CreateBareProject(b, db, testDialect, "Benchmark Project")
	columnID := fixtures.CreateTestColumn(b, db, testDialect, projectID, "Column")
	taskID := fixtures.CreateTestTask(b, db, testDialect, columnID, "Test Task")

	labels := make([]int, b.N)
	for i := 0; i < b.N; i++ {
		labels[i] = fixtures.CreateTestLabel(b, db, testDialect, projectID, "label_"+string(rune(i)), "#FF00FF")
	}

	service := newBenchmarkService(b, db)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := service.AttachLabel(ctx, taskID, labels[i])
		if err != nil {
			b.Fatalf("AttachLabel failed: %v", err)
		}
	}
}

func BenchmarkDetachLabel(b *testing.B) {
	db := fixtures.SetupTestDB(b)

	projectID := fixtures.CreateBareProject(b, db, testDialect, "Benchmark Project")
	columnID := fixtures.CreateTestColumn(b, db, testDialect, projectID, "Column")
	taskID := fixtures.CreateTestTask(b, db, testDialect, columnID, "Test Task")

	labels := make([]int, b.N)
	for i := 0; i < b.N; i++ {
		labelID := fixtures.CreateTestLabel(b, db, testDialect, projectID, "label_"+string(rune(i)), "#FF00FF")
		labels[i] = labelID
		fixtures.AttachLabelToTask(b, db, testDialect, taskID, labelID)
	}

	service := newBenchmarkService(b, db)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := service.DetachLabel(ctx, taskID, labels[i])
		if err != nil {
			b.Fatalf("DetachLabel failed: %v", err)
		}
	}
}

func BenchmarkMoveTaskToColumn(b *testing.B) {
	db := fixtures.SetupTestDB(b)

	projectID := fixtures.CreateBareProject(b, db, testDialect, "Benchmark Project")
	columnID := fixtures.CreateTestColumn(b, db, testDialect, projectID, "Column")
	targetColumn := fixtures.CreateTestColumn(b, db, testDialect, projectID, "Target")

	tasks := make([]int, b.N)
	for i := 0; i < b.N; i++ {
		tasks[i] = fixtures.CreateTestTask(b, db, testDialect, columnID, "Move Task "+string(rune(i)))
	}

	service := newBenchmarkService(b, db)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := service.MoveTaskToColumn(ctx, tasks[i], targetColumn)
		if err != nil {
			b.Fatalf("MoveTaskToColumn failed: %v", err)
		}
	}
}

func BenchmarkCreateComment(b *testing.B) {
	db := fixtures.SetupTestDB(b)

	projectID := fixtures.CreateBareProject(b, db, testDialect, "Benchmark Project")
	columnID := fixtures.CreateTestColumn(b, db, testDialect, projectID, "Column")
	taskID := fixtures.CreateTestTask(b, db, testDialect, columnID, "Test Task")

	service := newBenchmarkService(b, db)
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

func BenchmarkGetReadyTaskSummariesByProject(b *testing.B) {
	db := fixtures.SetupTestDB(b)

	projectID := fixtures.CreateBareProject(b, db, testDialect, "Benchmark Project")
	readyColumnID := fixtures.CreateColumnWithFlags(b, db, testDialect, projectID, "Ready", true, false, false)
	normalColumnID := fixtures.CreateTestColumn(b, db, testDialect, projectID, "Column")

	for i := 0; i < 100; i++ {
		title := "Ready Task " + string(rune(i%10))
		fixtures.CreateTestTask(b, db, testDialect, readyColumnID, title)
	}

	for i := 0; i < 50; i++ {
		title := "Normal Task " + string(rune(i%10))
		fixtures.CreateTestTask(b, db, testDialect, normalColumnID, title)
	}

	service := newBenchmarkService(b, db)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.GetReadyTaskSummariesByProject(ctx, projectID)
		if err != nil {
			b.Fatalf("GetReadyTaskSummariesByProject failed: %v", err)
		}
	}
}

func BenchmarkAddParentRelation(b *testing.B) {
	db := fixtures.SetupTestDB(b)

	projectID := fixtures.CreateBareProject(b, db, testDialect, "Benchmark Project")
	columnID := fixtures.CreateTestColumn(b, db, testDialect, projectID, "Column")

	parentID := fixtures.CreateTestTask(b, db, testDialect, columnID, "Parent Task")

	childTasks := make([]int, b.N)
	for i := 0; i < b.N; i++ {
		childTasks[i] = fixtures.CreateTestTask(b, db, testDialect, columnID, "Child Task "+string(rune(i)))
	}

	service := newBenchmarkService(b, db)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := service.AddParentRelation(ctx, childTasks[i], parentID, models.RelationTypeParentChild)
		if err != nil {
			b.Fatalf("AddParentRelation failed: %v", err)
		}
	}
}

func BenchmarkGetTaskReferencesForProject(b *testing.B) {
	db := fixtures.SetupTestDB(b)

	projectID := fixtures.CreateBareProject(b, db, testDialect, "Benchmark Project")
	columnID := fixtures.CreateTestColumn(b, db, testDialect, projectID, "Column")

	for i := 0; i < 200; i++ {
		title := "Task " + string(rune(i%10))
		fixtures.CreateTestTask(b, db, testDialect, columnID, title)
	}

	service := newBenchmarkService(b, db)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.GetTaskReferencesForProject(ctx, projectID)
		if err != nil {
			b.Fatalf("GetTaskReferencesForProject failed: %v", err)
		}
	}
}
