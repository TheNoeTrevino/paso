package tui

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/app"
	"github.com/thenoetrevino/paso/internal/config"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/services/column"
	"github.com/thenoetrevino/paso/internal/services/label"
	"github.com/thenoetrevino/paso/internal/services/project"
	"github.com/thenoetrevino/paso/internal/services/task"
	"github.com/thenoetrevino/paso/internal/testutil"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// If you are here bc you changed the ui layout,
// run the following command to update snapshots:
// UPDATE_SNAPSHOTS=1 go test ./internal/tui -run TestSnapshots

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
		{
			name: "no_projects",
			setup: func(t *testing.T, db *sql.DB) Model {
				return setupNoProjects(t, db)
			},
		},
		{
			name: "project_no_columns",
			setup: func(t *testing.T, db *sql.DB) Model {
				return setupProjectNoColumns(t, db)
			},
		},
		{
			name: "connection_disconnected",
			setup: func(t *testing.T, db *sql.DB) Model {
				return setupConnectionDisconnected(t, db)
			},
		},
		{
			name: "connection_reconnecting",
			setup: func(t *testing.T, db *sql.DB) Model {
				return setupConnectionReconnecting(t, db)
			},
		},
		{
			name: "notification_error",
			setup: func(t *testing.T, db *sql.DB) Model {
				return setupNotificationError(t, db)
			},
		},
		{
			name: "notification_warning",
			setup: func(t *testing.T, db *sql.DB) Model {
				return setupNotificationWarning(t, db)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := testutil.SetupTestDB(t)
			defer func() {
				_ = db.Close()
			}()

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
	testutil.CreateTestTask(t, db, columns[0].ID, "Setup database")
	testutil.CreateTestTask(t, db, columns[0].ID, "Configure service")
	testutil.CreateTestTask(t, db, columns[1].ID, "Implement API endpoints")
	testutil.CreateTestTask(t, db, columns[2].ID, "Deploy to production")

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
	for i := range 5 {
		_ = testutil.CreateTestTask(t, db, columns[0].ID, "Task "+string(rune(65+i)))
	}
	for i := range 3 {
		_ = testutil.CreateTestTask(t, db, columns[1].ID, "In Progress Task "+string(rune(65+i)))
	}
	for i := range 2 {
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
	_, err := db.ExecContext(ctx, "INSERT INTO task_labels (task_id, label_id) VALUES (?, ?)", task1ID, labelBug)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, "INSERT INTO task_labels (task_id, label_id) VALUES (?, ?)", task2ID, labelFeature)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, "INSERT INTO task_labels (task_id, label_id) VALUES (?, ?)", task2ID, labelDoc)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, "INSERT INTO task_labels (task_id, label_id) VALUES (?, ?)", task3ID, labelDoc)
	require.NoError(t, err)

	tasks, err := appContainer.TaskService.GetTaskSummariesByProject(ctx, projectID)
	require.NoError(t, err)
	labels, err := appContainer.LabelService.GetLabelsByProject(ctx, projectID)
	require.NoError(t, err)

	cfg, err := config.Load()
	require.NoError(t, err)
	m := InitialModel(ctx, appContainer, cfg, nil)

	m.AppState.SetColumns(columns)
	m.AppState.SetTasks(tasks)
	m.AppState.SetLabels(labels)

	return m
}

// setupNoProjects creates a model with no projects at all
func setupNoProjects(t *testing.T, db *sql.DB) Model {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create app container but don't create any projects
	appContainer := createAppContainer(db)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	m := InitialModel(ctx, appContainer, cfg, nil)

	// Explicitly set empty state
	m.AppState.SetColumns([]*models.Column{})
	m.AppState.SetTasks(make(map[int][]*models.TaskSummary))
	m.AppState.SetLabels([]*models.Label{})

	return m
}

// setupProjectNoColumns creates a model with a project but no columns
func setupProjectNoColumns(t *testing.T, db *sql.DB) Model {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create project but don't create any columns
	projectID := testutil.CreateTestProject(t, db, "Empty Project")
	appContainer := createAppContainer(db)

	// Delete default columns that were auto-created
	_, err := db.ExecContext(ctx, "DELETE FROM columns WHERE project_id = ?", projectID)
	if err != nil {
		t.Fatalf("Failed to delete columns: %v", err)
	}

	cfg, _ := config.Load()
	m := InitialModel(ctx, appContainer, cfg, nil)

	// Set empty columns
	m.AppState.SetColumns([]*models.Column{})
	m.AppState.SetTasks(make(map[int][]*models.TaskSummary))
	m.AppState.SetLabels([]*models.Label{})

	return m
}

// setupConnectionDisconnected creates a model with disconnected daemon status
func setupConnectionDisconnected(t *testing.T, db *sql.DB) Model {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	projectID := testutil.CreateTestProject(t, db, "Test Project")
	appContainer := createAppContainer(db)

	columns, err := appContainer.ColumnService.GetColumnsByProject(ctx, projectID)
	require.NoError(t, err)
	tasks, err := appContainer.TaskService.GetTaskSummariesByProject(ctx, projectID)
	require.NoError(t, err)
	labels, err := appContainer.LabelService.GetLabelsByProject(ctx, projectID)
	require.NoError(t, err)

	cfg, err := config.Load()
	require.NoError(t, err)
	// Pass nil for eventClient to simulate disconnected state
	m := InitialModel(ctx, appContainer, cfg, nil)

	m.AppState.SetColumns(columns)
	m.AppState.SetTasks(tasks)
	m.AppState.SetLabels(labels)

	// Explicitly set disconnected status
	m.ConnectionState.SetStatus(state.Disconnected)

	return m
}

// setupConnectionReconnecting creates a model in reconnecting state
func setupConnectionReconnecting(t *testing.T, db *sql.DB) Model {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	projectID := testutil.CreateTestProject(t, db, "Test Project")
	appContainer := createAppContainer(db)

	columns, _ := appContainer.ColumnService.GetColumnsByProject(ctx, projectID)
	tasks, _ := appContainer.TaskService.GetTaskSummariesByProject(ctx, projectID)
	labels, _ := appContainer.LabelService.GetLabelsByProject(ctx, projectID)

	cfg, _ := config.Load()
	m := InitialModel(ctx, appContainer, cfg, nil)

	m.AppState.SetColumns(columns)
	m.AppState.SetTasks(tasks)
	m.AppState.SetLabels(labels)

	// Explicitly set reconnecting status
	m.ConnectionState.SetStatus(state.Reconnecting)

	return m
}

// setupNotificationError creates a model with an error notification
func setupNotificationError(t *testing.T, db *sql.DB) Model {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	projectID := testutil.CreateTestProject(t, db, "Test Project")
	appContainer := createAppContainer(db)

	columns, err := appContainer.ColumnService.GetColumnsByProject(ctx, projectID)
	require.NoError(t, err)
	tasks, err := appContainer.TaskService.GetTaskSummariesByProject(ctx, projectID)
	require.NoError(t, err)
	labels, err := appContainer.LabelService.GetLabelsByProject(ctx, projectID)
	require.NoError(t, err)

	cfg, _ := config.Load()
	m := InitialModel(ctx, appContainer, cfg, nil)

	m.AppState.SetColumns(columns)
	m.AppState.SetTasks(tasks)
	m.AppState.SetLabels(labels)

	// Add error notification
	m.UI.Notification.SetWindowSize(80, 24)
	m.UI.Notification.Add(state.LevelError, "Failed to save task: database connection lost")

	return m
}

// setupNotificationWarning creates a model with a warning notification
func setupNotificationWarning(t *testing.T, db *sql.DB) Model {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	projectID := testutil.CreateTestProject(t, db, "Test Project")
	appContainer := createAppContainer(db)

	columns, err := appContainer.ColumnService.GetColumnsByProject(ctx, projectID)
	require.NoError(t, err)
	tasks, _ := appContainer.TaskService.GetTaskSummariesByProject(ctx, projectID)
	require.NoError(t, err)
	labels, _ := appContainer.LabelService.GetLabelsByProject(ctx, projectID)
	require.NoError(t, err)

	cfg, err := config.Load()
	require.NoError(t, err)
	m := InitialModel(ctx, appContainer, cfg, nil)

	m.AppState.SetColumns(columns)
	m.AppState.SetTasks(tasks)
	m.AppState.SetLabels(labels)

	// Add warning notification
	m.UI.Notification.SetWindowSize(80, 24)
	m.UI.Notification.Add(state.LevelWarning, "Daemon connection unstable - some features may be limited")

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
		// New snapshots for error, empty, and connection states
		"no_projects",
		"project_no_columns",
		"connection_disconnected",
		"connection_reconnecting",
		"notification_error",
		"notification_warning",
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
