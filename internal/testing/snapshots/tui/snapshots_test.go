package tui_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/app"
	"github.com/thenoetrevino/paso/internal/config"
	"github.com/thenoetrevino/paso/internal/database"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/services/column"
	"github.com/thenoetrevino/paso/internal/services/label"
	"github.com/thenoetrevino/paso/internal/services/project"
	"github.com/thenoetrevino/paso/internal/services/task"
	"github.com/thenoetrevino/paso/internal/testing/fixtures"
	"github.com/thenoetrevino/paso/internal/testing/snapshots"
	"github.com/thenoetrevino/paso/internal/tui"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// NOTE:
// If you are here bc you changed the ui layout and tests are failing,
// run the following command to update snapshots:
// UPDATE_SNAPSHOTS=1 go test ./internal/testing/snapshots/tui -run TestSnapshots

// TestSnapshots verifies TUI rendering consistency across different application states
func TestSnapshots(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		setup func(*testing.T, *sql.DB) tui.Model
	}{
		{
			name: "empty_project_board",
			setup: func(t *testing.T, db *sql.DB) tui.Model {
				return setupEmptyProject(t, db)
			},
		},
		{
			name: "board_with_tasks",
			setup: func(t *testing.T, db *sql.DB) tui.Model {
				return setupBoardWithTasks(t, db)
			},
		},
		{
			name: "board_with_multiple_tasks",
			setup: func(t *testing.T, db *sql.DB) tui.Model {
				return setupBoardWithMultipleTasks(t, db)
			},
		},
		{
			name: "board_with_labels",
			setup: func(t *testing.T, db *sql.DB) tui.Model {
				return setupBoardWithLabels(t, db)
			},
		},
		{
			name: "no_projects",
			setup: func(t *testing.T, db *sql.DB) tui.Model {
				return setupNoProjects(t, db)
			},
		},
		{
			name: "project_no_columns",
			setup: func(t *testing.T, db *sql.DB) tui.Model {
				return setupProjectNoColumns(t, db)
			},
		},
		{
			name: "connection_disconnected",
			setup: func(t *testing.T, db *sql.DB) tui.Model {
				return setupConnectionDisconnected(t, db)
			},
		},
		{
			name: "connection_reconnecting",
			setup: func(t *testing.T, db *sql.DB) tui.Model {
				return setupConnectionReconnecting(t, db)
			},
		},
		{
			name: "notification_error",
			setup: func(t *testing.T, db *sql.DB) tui.Model {
				return setupNotificationError(t, db)
			},
		},
		{
			name: "notification_warning",
			setup: func(t *testing.T, db *sql.DB) tui.Model {
				return setupNotificationWarning(t, db)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			db := fixtures.SetupTestDB(t)

			m := tt.setup(t, db)

			// Set fixed terminal size for consistent snapshots (80x24 is standard)
			m.UIState.SetWidth(80)
			m.UIState.Height = 24

			// Render the view
			view := m.View()
			output := view.Content

			// Compare against golden file
			helper := snapshots.NewHelper(t, "testdata")
			helper.Compare(tt.name, output)
		})
	}
}

func setupEmptyProject(t *testing.T, db *sql.DB) tui.Model {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	projectID := fixtures.CreateTestProject(t, db, fixtures.SQLiteDialect(), "Test Project")
	appContainer := createAppContainer(t, db)

	_, err := appContainer.ProjectService.GetProjectByID(ctx, projectID)
	require.NoError(t, err, "Failed to get project")

	columns, err := appContainer.ColumnService.GetColumnsByProject(ctx, projectID)
	require.NoError(t, err)

	tasks, err := appContainer.TaskService.GetTaskSummariesByProject(ctx, projectID)
	require.NoError(t, err)
	labels, err := appContainer.LabelService.GetLabelsByProject(ctx, projectID)
	require.NoError(t, err)

	cfg, err := config.Load()
	require.NoError(t, err)
	m := tui.InitialModel(ctx, appContainer, cfg, nil, db)

	m.AppState.SetColumns(columns)
	m.AppState.SetTasks(tasks)
	m.AppState.SetLabels(labels)

	m.UI.CurrentTip = "Press 'p' to create a new project, 'n'/'N' to switch projects"

	return m
}

func setupBoardWithTasks(t *testing.T, db *sql.DB) tui.Model {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	projectID := fixtures.CreateTestProject(t, db, fixtures.SQLiteDialect(), "Test Project")
	appContainer := createAppContainer(t, db)

	columns, err := appContainer.ColumnService.GetColumnsByProject(ctx, projectID)
	require.NoError(t, err)

	require.GreaterOrEqual(t, len(columns), 3, "Expected at least 3 columns")

	fixtures.CreateTestTask(t, db, fixtures.SQLiteDialect(), columns[0].ID, "Setup database")
	fixtures.CreateTestTask(t, db, fixtures.SQLiteDialect(), columns[0].ID, "Configure service")
	fixtures.CreateTestTask(t, db, fixtures.SQLiteDialect(), columns[1].ID, "Implement API endpoints")
	fixtures.CreateTestTask(t, db, fixtures.SQLiteDialect(), columns[2].ID, "Deploy to production")

	tasks, err := appContainer.TaskService.GetTaskSummariesByProject(ctx, projectID)
	require.NoError(t, err)
	labels, err := appContainer.LabelService.GetLabelsByProject(ctx, projectID)
	require.NoError(t, err)

	cfg, err := config.Load()
	require.NoError(t, err)
	m := tui.InitialModel(ctx, appContainer, cfg, nil, db)

	m.AppState.SetColumns(columns)
	m.AppState.SetTasks(tasks)
	m.AppState.SetLabels(labels)

	m.UI.CurrentTip = "Press 'p' to create a new project, 'n'/'N' to switch projects"

	return m
}

func setupBoardWithMultipleTasks(t *testing.T, db *sql.DB) tui.Model {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	projectID := fixtures.CreateTestProject(t, db, fixtures.SQLiteDialect(), "Test Project")
	appContainer := createAppContainer(t, db)

	columns, err := appContainer.ColumnService.GetColumnsByProject(ctx, projectID)
	require.NoError(t, err)

	for i := range 5 {
		fixtures.CreateTestTask(t, db, fixtures.SQLiteDialect(), columns[0].ID, "Task "+string(rune(65+i)))
	}
	for i := range 3 {
		fixtures.CreateTestTask(t, db, fixtures.SQLiteDialect(), columns[1].ID, "In Progress Task "+string(rune(65+i)))
	}
	for i := range 2 {
		fixtures.CreateTestTask(t, db, fixtures.SQLiteDialect(), columns[2].ID, "Done Task "+string(rune(65+i)))
	}

	tasks, err := appContainer.TaskService.GetTaskSummariesByProject(ctx, projectID)
	require.NoError(t, err)
	labels, err := appContainer.LabelService.GetLabelsByProject(ctx, projectID)
	require.NoError(t, err)

	cfg, err := config.Load()
	require.NoError(t, err)
	m := tui.InitialModel(ctx, appContainer, cfg, nil, db)

	m.AppState.SetColumns(columns)
	m.AppState.SetTasks(tasks)
	m.AppState.SetLabels(labels)

	m.UI.CurrentTip = "Press 'p' to create a new project, 'n'/'N' to switch projects"

	return m
}

func setupBoardWithLabels(t *testing.T, db *sql.DB) tui.Model {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	projectID := fixtures.CreateTestProject(t, db, fixtures.SQLiteDialect(), "Test Project")
	appContainer := createAppContainer(t, db)

	labelBug := fixtures.CreateTestLabel(t, db, fixtures.SQLiteDialect(), projectID, "bug", "#FF0000")
	labelFeature := fixtures.CreateTestLabel(t, db, fixtures.SQLiteDialect(), projectID, "feature", "#00FF00")
	labelDoc := fixtures.CreateTestLabel(t, db, fixtures.SQLiteDialect(), projectID, "documentation", "#0000FF")

	columns, err := appContainer.ColumnService.GetColumnsByProject(ctx, projectID)
	require.NoError(t, err)

	task1ID := fixtures.CreateTestTask(t, db, fixtures.SQLiteDialect(), columns[0].ID, "Fix critical bug")
	task2ID := fixtures.CreateTestTask(t, db, fixtures.SQLiteDialect(), columns[0].ID, "Implement new feature")
	task3ID := fixtures.CreateTestTask(t, db, fixtures.SQLiteDialect(), columns[1].ID, "Write API docs")

	_, err = db.ExecContext(ctx, "INSERT INTO task_labels (task_id, label_id) VALUES (?, ?)", task1ID, labelBug)
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
	m := tui.InitialModel(ctx, appContainer, cfg, nil, db)

	m.AppState.SetColumns(columns)
	m.AppState.SetTasks(tasks)
	m.AppState.SetLabels(labels)

	m.UI.CurrentTip = "Press 'p' to create a new project, 'n'/'N' to switch projects"

	return m
}

func setupNoProjects(t *testing.T, db *sql.DB) tui.Model {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	appContainer := createAppContainer(t, db)

	cfg, err := config.Load()
	require.NoError(t, err, "Failed to load config")

	m := tui.InitialModel(ctx, appContainer, cfg, nil, db)

	m.AppState.SetColumns([]*models.Column{})
	m.AppState.SetTasks(make(map[int][]*models.TaskSummary))
	m.AppState.SetLabels([]*models.Label{})

	m.UI.CurrentTip = "Press 'p' to create a new project, 'n'/'N' to switch projects"

	return m
}

func setupProjectNoColumns(t *testing.T, db *sql.DB) tui.Model {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	projectID := fixtures.CreateTestProject(t, db, fixtures.SQLiteDialect(), "Empty Project")
	appContainer := createAppContainer(t, db)

	_, err := db.ExecContext(ctx, "DELETE FROM columns WHERE project_id = ?", projectID)
	require.NoError(t, err)

	cfg, err := config.Load()
	require.NoError(t, err)
	m := tui.InitialModel(ctx, appContainer, cfg, nil, db)

	m.AppState.SetColumns([]*models.Column{})
	m.AppState.SetTasks(make(map[int][]*models.TaskSummary))
	m.AppState.SetLabels([]*models.Label{})

	m.UI.CurrentTip = "Press 'p' to create a new project, 'n'/'N' to switch projects"

	return m
}

func setupConnectionDisconnected(t *testing.T, db *sql.DB) tui.Model {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	projectID := fixtures.CreateTestProject(t, db, fixtures.SQLiteDialect(), "Test Project")
	appContainer := createAppContainer(t, db)

	columns, err := appContainer.ColumnService.GetColumnsByProject(ctx, projectID)
	require.NoError(t, err)
	tasks, err := appContainer.TaskService.GetTaskSummariesByProject(ctx, projectID)
	require.NoError(t, err)
	labels, err := appContainer.LabelService.GetLabelsByProject(ctx, projectID)
	require.NoError(t, err)

	cfg, err := config.Load()
	require.NoError(t, err)
	m := tui.InitialModel(ctx, appContainer, cfg, nil, db)

	m.AppState.SetColumns(columns)
	m.AppState.SetTasks(tasks)
	m.AppState.SetLabels(labels)

	m.ConnectionState.SetStatus(state.Disconnected)

	m.UI.CurrentTip = "Press 'p' to create a new project, 'n'/'N' to switch projects"

	return m
}

func setupConnectionReconnecting(t *testing.T, db *sql.DB) tui.Model {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	projectID := fixtures.CreateTestProject(t, db, fixtures.SQLiteDialect(), "Test Project")
	appContainer := createAppContainer(t, db)

	columns, err := appContainer.ColumnService.GetColumnsByProject(ctx, projectID)
	require.NoError(t, err)
	tasks, err := appContainer.TaskService.GetTaskSummariesByProject(ctx, projectID)
	require.NoError(t, err)
	labels, err := appContainer.LabelService.GetLabelsByProject(ctx, projectID)
	require.NoError(t, err)

	cfg, err := config.Load()
	require.NoError(t, err)
	m := tui.InitialModel(ctx, appContainer, cfg, nil, db)

	m.AppState.SetColumns(columns)
	m.AppState.SetTasks(tasks)
	m.AppState.SetLabels(labels)

	m.ConnectionState.SetStatus(state.Reconnecting)

	m.UI.CurrentTip = "Press 'p' to create a new project, 'n'/'N' to switch projects"

	return m
}

func setupNotificationError(t *testing.T, db *sql.DB) tui.Model {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	projectID := fixtures.CreateTestProject(t, db, fixtures.SQLiteDialect(), "Test Project")
	appContainer := createAppContainer(t, db)

	columns, err := appContainer.ColumnService.GetColumnsByProject(ctx, projectID)
	require.NoError(t, err)
	tasks, err := appContainer.TaskService.GetTaskSummariesByProject(ctx, projectID)
	require.NoError(t, err)
	labels, err := appContainer.LabelService.GetLabelsByProject(ctx, projectID)
	require.NoError(t, err)

	cfg, err := config.Load()
	require.NoError(t, err)
	m := tui.InitialModel(ctx, appContainer, cfg, nil, db)

	m.AppState.SetColumns(columns)
	m.AppState.SetTasks(tasks)
	m.AppState.SetLabels(labels)

	m.UI.Notification.SetWindowSize(80, 24)
	m.UI.Notification.Add(state.LevelError, "Failed to save task: database connection lost")

	m.UI.CurrentTip = "Press 'p' to create a new project, 'n'/'N' to switch projects"

	return m
}

func setupNotificationWarning(t *testing.T, db *sql.DB) tui.Model {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	projectID := fixtures.CreateTestProject(t, db, fixtures.SQLiteDialect(), "Test Project")
	appContainer := createAppContainer(t, db)

	columns, err := appContainer.ColumnService.GetColumnsByProject(ctx, projectID)
	require.NoError(t, err)
	tasks, err := appContainer.TaskService.GetTaskSummariesByProject(ctx, projectID)
	require.NoError(t, err)
	labels, err := appContainer.LabelService.GetLabelsByProject(ctx, projectID)
	require.NoError(t, err)

	cfg, err := config.Load()
	require.NoError(t, err)
	m := tui.InitialModel(ctx, appContainer, cfg, nil, db)

	m.AppState.SetColumns(columns)
	m.AppState.SetTasks(tasks)
	m.AppState.SetLabels(labels)

	m.UI.Notification.SetWindowSize(80, 24)
	m.UI.Notification.Add(state.LevelWarning, "Daemon connection unstable - some features may be limited")

	m.UI.CurrentTip = "Press 'p' to create a new project, 'n'/'N' to switch projects"

	return m
}

func createAppContainer(t *testing.T, db *sql.DB) *app.App {
	t.Helper()
	dbType := database.SQLite
	taskSvc, err := task.NewService(db, dbType, nil, nil)
	require.NoError(t, err, "failed to create task service")
	columnSvc, err := column.NewService(db, dbType, nil)
	require.NoError(t, err, "failed to create column service")
	labelSvc, err := label.NewService(db, dbType, nil)
	require.NoError(t, err, "failed to create label service")
	projectSvc, err := project.NewService(db, dbType, nil, nil)
	require.NoError(t, err, "failed to create project service")
	return &app.App{
		TaskService:    taskSvc,
		ColumnService:  columnSvc,
		LabelService:   labelSvc,
		ProjectService: projectSvc,
	}
}

// TestSnapshotRegressions verifies that snapshot golden files are properly maintained
func TestSnapshotRegressions(t *testing.T) {
	t.Parallel()
	helper := snapshots.NewHelper(t, "testdata")

	snapshotNames := []string{
		"empty_project_board",
		"board_with_tasks",
		"board_with_multiple_tasks",
		"board_with_labels",
		"no_projects",
		"project_no_columns",
		"connection_disconnected",
		"connection_reconnecting",
		"notification_error",
		"notification_warning",
	}

	for _, name := range snapshotNames {
		t.Run("verify_"+name, func(t *testing.T) {
			t.Parallel()
			_, err := helper.Read(name)
			if err != nil {
				t.Logf("Snapshot %s not found. Run UPDATE_SNAPSHOTS=1 to create", name)
			}
		})
	}
}
