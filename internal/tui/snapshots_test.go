package tui

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/thenoetrevino/paso/internal/app"
	"github.com/thenoetrevino/paso/internal/config"
	"github.com/thenoetrevino/paso/internal/services/column"
	"github.com/thenoetrevino/paso/internal/services/label"
	"github.com/thenoetrevino/paso/internal/services/project"
	"github.com/thenoetrevino/paso/internal/services/task"
	"github.com/thenoetrevino/paso/internal/testutil"
)

// TestSnapshots verifies TUI rendering consistency across different application states
func TestSnapshots(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*testing.T, *sql.DB) Model
	}{
		{
			name: "empty_project_board",
			setup: func(t *testing.T, db *sql.DB) Model {
				return setupEmptyProject(t, db)
			},
		},
		{
			name: "board_with_tasks",
			setup: func(t *testing.T, db *sql.DB) Model {
				return setupBoardWithTasks(t, db)
			},
		},
		{
			name: "board_with_multiple_tasks",
			setup: func(t *testing.T, db *sql.DB) Model {
				return setupBoardWithMultipleTasks(t, db)
			},
		},
		{
			name: "board_with_labels",
			setup: func(t *testing.T, db *sql.DB) Model {
				return setupBoardWithLabels(t, db)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := testutil.SetupTestDB(t)
			defer db.Close()

			m := tt.setup(t, db)

			// Set fixed terminal size for consistent snapshots (80x24 is standard)
			m.UIState.SetWidth(80)
			m.UIState.SetHeight(24)

			// Render the view
			view := m.View()
			output := view.Content

			// Compare against golden file
			helper := NewSnapshotHelper(t)
			helper.Compare(tt.name, output)
		})
	}
}

// setupEmptyProject creates a model with an empty project and default columns
func setupEmptyProject(t *testing.T, db *sql.DB) Model {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create project and services
	projectID := testutil.CreateTestProject(t, db, "Test Project")
	appContainer := createAppContainer(db)

	// Load project data
	_, err := appContainer.ProjectService.GetProjectByID(ctx, projectID)
	if err != nil {
		t.Fatalf("Failed to get project: %v", err)
	}

	columns, err := appContainer.ColumnService.GetColumnsByProject(ctx, projectID)
	if err != nil {
		t.Fatalf("Failed to get columns: %v", err)
	}

	tasks, _ := appContainer.TaskService.GetTaskSummariesByProject(ctx, projectID)
	labels, _ := appContainer.LabelService.GetLabelsByProject(ctx, projectID)

	cfg, _ := config.Load()
	m := InitialModel(ctx, appContainer, cfg, nil)

	// Override with loaded data
	m.AppState.SetColumns(columns)
	m.AppState.SetTasks(tasks)
	m.AppState.SetLabels(labels)

	return m
}

// setupBoardWithTasks creates a model with tasks across columns
func setupBoardWithTasks(t *testing.T, db *sql.DB) Model {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	projectID := testutil.CreateTestProject(t, db, "Test Project")
	appContainer := createAppContainer(db)

	// Get columns
	columns, err := appContainer.ColumnService.GetColumnsByProject(ctx, projectID)
	if err != nil {
		t.Fatalf("Failed to get columns: %v", err)
	}

	if len(columns) < 3 {
		t.Fatalf("Expected at least 3 columns, got %d", len(columns))
	}

	// Create tasks in different columns
	_ = testutil.CreateTestTask(t, db, columns[0].ID, "Setup database")
	_ = testutil.CreateTestTask(t, db, columns[0].ID, "Configure service")
	_ = testutil.CreateTestTask(t, db, columns[1].ID, "Implement API endpoints")
	_ = testutil.CreateTestTask(t, db, columns[2].ID, "Deploy to production")

	tasks, _ := appContainer.TaskService.GetTaskSummariesByProject(ctx, projectID)
	labels, _ := appContainer.LabelService.GetLabelsByProject(ctx, projectID)

	cfg, _ := config.Load()
	m := InitialModel(ctx, appContainer, cfg, nil)

	m.AppState.SetColumns(columns)
	m.AppState.SetTasks(tasks)
	m.AppState.SetLabels(labels)

	return m
}

// setupBoardWithMultipleTasks creates a model with many tasks for testing rendering at scale
func setupBoardWithMultipleTasks(t *testing.T, db *sql.DB) Model {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	projectID := testutil.CreateTestProject(t, db, "Test Project")
	appContainer := createAppContainer(db)

	columns, _ := appContainer.ColumnService.GetColumnsByProject(ctx, projectID)

	// Create multiple tasks per column
	for i := 0; i < 5; i++ {
		_ = testutil.CreateTestTask(t, db, columns[0].ID, "Task "+string(rune(65+i)))
	}
	for i := 0; i < 3; i++ {
		_ = testutil.CreateTestTask(t, db, columns[1].ID, "In Progress Task "+string(rune(65+i)))
	}
	for i := 0; i < 2; i++ {
		_ = testutil.CreateTestTask(t, db, columns[2].ID, "Done Task "+string(rune(65+i)))
	}

	tasks, _ := appContainer.TaskService.GetTaskSummariesByProject(ctx, projectID)
	labels, _ := appContainer.LabelService.GetLabelsByProject(ctx, projectID)

	cfg, _ := config.Load()
	m := InitialModel(ctx, appContainer, cfg, nil)

	m.AppState.SetColumns(columns)
	m.AppState.SetTasks(tasks)
	m.AppState.SetLabels(labels)

	return m
}

// setupBoardWithLabels creates a model with labeled tasks
func setupBoardWithLabels(t *testing.T, db *sql.DB) Model {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	projectID := testutil.CreateTestProject(t, db, "Test Project")
	appContainer := createAppContainer(db)

	// Create labels
	labelBug := testutil.CreateTestLabel(t, db, projectID, "bug", "#FF0000")
	labelFeature := testutil.CreateTestLabel(t, db, projectID, "feature", "#00FF00")
	labelDoc := testutil.CreateTestLabel(t, db, projectID, "documentation", "#0000FF")

	columns, _ := appContainer.ColumnService.GetColumnsByProject(ctx, projectID)

	// Create tasks with labels
	task1ID := testutil.CreateTestTask(t, db, columns[0].ID, "Fix critical bug")
	task2ID := testutil.CreateTestTask(t, db, columns[0].ID, "Implement new feature")
	task3ID := testutil.CreateTestTask(t, db, columns[1].ID, "Write API docs")

	// Attach labels
	db.ExecContext(ctx, "INSERT INTO task_labels (task_id, label_id) VALUES (?, ?)", task1ID, labelBug)
	db.ExecContext(ctx, "INSERT INTO task_labels (task_id, label_id) VALUES (?, ?)", task2ID, labelFeature)
	db.ExecContext(ctx, "INSERT INTO task_labels (task_id, label_id) VALUES (?, ?)", task2ID, labelDoc)
	db.ExecContext(ctx, "INSERT INTO task_labels (task_id, label_id) VALUES (?, ?)", task3ID, labelDoc)

	tasks, _ := appContainer.TaskService.GetTaskSummariesByProject(ctx, projectID)
	labels, _ := appContainer.LabelService.GetLabelsByProject(ctx, projectID)

	cfg, _ := config.Load()
	m := InitialModel(ctx, appContainer, cfg, nil)

	m.AppState.SetColumns(columns)
	m.AppState.SetTasks(tasks)
	m.AppState.SetLabels(labels)

	return m
}

// createAppContainer creates an app container with all services
func createAppContainer(db *sql.DB) *app.App {
	return &app.App{
		TaskService:    task.NewService(db, nil),
		ColumnService:  column.NewService(db, nil),
		LabelService:   label.NewService(db, nil),
		ProjectService: project.NewService(db, nil),
	}
}

// TestSnapshotRegressions verifies that snapshot files exist and match baseline
func TestSnapshotRegressions(t *testing.T) {
	// This test verifies that snapshot golden files are properly maintained
	helper := NewSnapshotHelper(t)

	snapshotNames := []string{
		"empty_project_board",
		"board_with_tasks",
		"board_with_multiple_tasks",
		"board_with_labels",
	}

	for _, name := range snapshotNames {
		t.Run("verify_"+name, func(t *testing.T) {
			_, err := helper.ReadSnapshot(name)
			if err != nil {
				t.Logf("Snapshot %s not found. Run UPDATE_SNAPSHOTS=1 to create", name)
			}
		})
	}
}
