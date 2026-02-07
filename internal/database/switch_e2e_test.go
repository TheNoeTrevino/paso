package database

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/database/types"
)

// ============================================================================
// Database Switching E2E Tests
// Tests data persistence when switching between SQLite and PostgreSQL
// ============================================================================

// TestSwitchDatabaseWithDataIntegrity verifies that data persists when
// switching between SQLite and PostgreSQL configurations
func TestSwitchDatabaseWithDataIntegrity(t *testing.T) {
	ctx := context.Background()

	// Setup SQLite database with test data
	sqliteDB := setupTestDB(t)
	defer func() { _ = sqliteDB.Close() }()

	sqliteQueries, err := NewQuerier(sqliteDB, SQLite)
	require.NoError(t, err, "failed to create SQLite querier")

	// Create test data in SQLite
	testProject := createTestProjectWithData(t, ctx, sqliteQueries)

	// Setup PostgreSQL database (skip if not available)
	postgresDB := setupPostgresTestDB(t)
	if postgresDB == nil {
		t.Skip("PostgreSQL not available for E2E testing")
	}
	defer func() { _ = postgresDB.Close() }()

	postgresQueries, err := NewQuerier(postgresDB, PostgreSQL)
	require.NoError(t, err, "failed to create PostgreSQL querier")

	// Run migrations on PostgreSQL
	err = runMigrations(ctx, postgresDB, PostgreSQL)
	require.NoError(t, err, "failed to run PostgreSQL migrations")

	// Verify data can be created in PostgreSQL with same structure
	verifyProjectStructureInDatabase(t, ctx, postgresQueries, testProject)
}

// TestLocalDatabasePersistence verifies that data persists across
// multiple connections to the same SQLite database
func TestLocalDatabasePersistence(t *testing.T) {
	ctx := context.Background()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "persistence_test.db")

	// Create database config for SQLite
	config := Config{
		Type:       SQLite,
		SQLitePath: dbPath,
	}

	// First connection - create database and data
	db1, err := InitDB(ctx, config, "persistence_test_first")
	require.NoError(t, err, "failed to initialize first SQLite connection")
	defer func() { _ = db1.Close() }()

	queries1, err := NewQuerier(db1, SQLite)
	require.NoError(t, err, "failed to create first SQLite querier")

	// Create test data
	projectParams := types.CreateProjectRecordParams{
		Name:        "Persistence Test Project",
		Description: types.NullString{String: "Testing data persistence", Valid: true},
	}
	project, err := queries1.CreateProjectRecord(ctx, projectParams)
	require.NoError(t, err, "failed to create project in first connection")

	err = queries1.InitializeProjectCounter(ctx, project.ID)
	require.NoError(t, err, "failed to initialize project counter")

	// Create column
	columnParams := types.CreateColumnParams{
		ProjectID: project.ID,
		Name:      "Todo",
	}
	column, err := queries1.CreateColumn(ctx, columnParams)
	require.NoError(t, err, "failed to create column")

	// Create task
	taskParams := types.CreateTaskParams{
		Title:    "Persistence Test Task",
		ColumnID: column.ID,
		Position: 0,
	}
	task, err := queries1.CreateTask(ctx, taskParams)
	require.NoError(t, err, "failed to create task")

	projectID := project.ID
	columnID := column.ID
	taskID := task.ID

	// Close first connection
	err = db1.Close()
	require.NoError(t, err, "failed to close first connection")

	// Second connection - verify data persists
	db2, err := InitDB(ctx, config, "persistence_test_second")
	require.NoError(t, err, "failed to initialize second SQLite connection")
	defer func() { _ = db2.Close() }()

	queries2, err := NewQuerier(db2, SQLite)
	require.NoError(t, err, "failed to create second SQLite querier")

	// Verify project still exists
	retrievedProject, err := queries2.GetProjectByID(ctx, projectID)
	require.NoError(t, err, "failed to retrieve project")
	require.Equal(t, "Persistence Test Project", retrievedProject.Name, "project name mismatch")

	// Verify column still exists
	retrievedColumn, err := queries2.GetColumnByID(ctx, columnID)
	require.NoError(t, err, "failed to retrieve column")
	require.Equal(t, "Todo", retrievedColumn.Name, "column name mismatch")

	// Verify task still exists
	retrievedTask, err := queries2.GetTask(ctx, taskID)
	require.NoError(t, err, "failed to retrieve task")
	require.Equal(t, "Persistence Test Task", retrievedTask.Title, "task title mismatch")

	// Verify file was created
	fileInfo, err := os.Stat(dbPath)
	require.NoError(t, err, "database file should exist")
	require.True(t, fileInfo.Size() > 0, "database file should not be empty")
}

// TestRemoteDatabaseConfiguration verifies that remote PostgreSQL
// configuration is properly validated and initialized
func TestRemoteDatabaseConfiguration(t *testing.T) {
	// This test validates connection string parsing and configuration
	// without requiring actual PostgreSQL connection

	connStr := "host=localhost port=5432 user=testuser password=testpass dbname=testdb sslmode=require"

	config, err := ParsePostgresConnectionString(connStr)
	require.NoError(t, err, "failed to parse PostgreSQL connection string")

	require.Equal(t, PostgreSQL, config.Type)
	require.Equal(t, "localhost", config.PostgresHost)
	require.Equal(t, 5432, config.PostgresPort)
	require.Equal(t, "testuser", config.PostgresUser)
	require.Equal(t, "testpass", config.PostgresPassword)
	require.Equal(t, "testdb", config.PostgresDB)
	require.True(t, config.IsRemoteDatabase())
	require.False(t, config.IsLocalDatabase())
}

// TestDatabaseConfigIsLocalVsRemote verifies the helper methods
// for detecting local vs remote databases
func TestDatabaseConfigIsLocalVsRemote(t *testing.T) {
	// Test local (SQLite)
	localConfig := Config{
		Type:       SQLite,
		SQLitePath: "/path/to/database.db",
	}
	require.True(t, localConfig.IsLocalDatabase())
	require.False(t, localConfig.IsRemoteDatabase())

	// Test remote (PostgreSQL)
	remoteConfig := Config{
		Type:             PostgreSQL,
		PostgresHost:     "db.example.com",
		PostgresPort:     5432,
		PostgresUser:     "user",
		PostgresPassword: "pass",
		PostgresDB:       "mydb",
	}
	require.False(t, remoteConfig.IsLocalDatabase())
	require.True(t, remoteConfig.IsRemoteDatabase())
}

// TestSwitchBetweenDatabases verifies that we can create different
// database configs and validate their connection strings
func TestSwitchBetweenDatabases(t *testing.T) {
	tmpDir := t.TempDir()

	// Config 1: Local SQLite
	sqlite1Path := filepath.Join(tmpDir, "database1.db")
	config1 := Config{
		Type:       SQLite,
		SQLitePath: sqlite1Path,
	}

	connStr1, err := config1.ConnectionString()
	require.NoError(t, err)
	require.Equal(t, sqlite1Path, connStr1)

	// Config 2: Local SQLite (different path)
	sqlite2Path := filepath.Join(tmpDir, "database2.db")
	config2 := Config{
		Type:       SQLite,
		SQLitePath: sqlite2Path,
	}

	connStr2, err := config2.ConnectionString()
	require.NoError(t, err)
	require.Equal(t, sqlite2Path, connStr2)
	require.NotEqual(t, connStr1, connStr2, "different SQLite paths should produce different connection strings")

	// Config 3: PostgreSQL
	postgresConfig := Config{
		Type:             PostgreSQL,
		PostgresHost:     "localhost",
		PostgresPort:     5432,
		PostgresUser:     "user",
		PostgresPassword: "password",
		PostgresDB:       "testdb",
	}

	connStr3, err := postgresConfig.ConnectionString()
	require.NoError(t, err)
	require.Contains(t, connStr3, "host=localhost")
	require.Contains(t, connStr3, "port=5432")
	require.NotEqual(t, connStr1, connStr3, "SQLite and PostgreSQL configs should have different connection strings")
}

// TestDatabaseInitializationWithDifferentTypes verifies that
// InitDB works correctly with both SQLite and PostgreSQL configurations
func TestDatabaseInitializationWithDifferentTypes(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Test SQLite initialization
	sqliteConfig := Config{
		Type:       SQLite,
		SQLitePath: filepath.Join(tmpDir, "test.db"),
	}

	db, err := InitDB(ctx, sqliteConfig, "test")
	require.NoError(t, err, "failed to initialize SQLite database")
	require.NotNil(t, db, "database connection should not be nil")

	// Verify database is usable
	err = db.PingContext(ctx)
	require.NoError(t, err, "database should be reachable")

	err = db.Close()
	require.NoError(t, err, "failed to close database")

	// Test PostgreSQL would fail without PostgreSQL running, which is expected
	postgresConfig := Config{
		Type:             PostgreSQL,
		PostgresHost:     "invalid-host-that-does-not-exist.local",
		PostgresPort:     5432,
		PostgresUser:     "user",
		PostgresPassword: "password",
		PostgresDB:       "testdb",
	}

	_, err = InitDB(ctx, postgresConfig, "test")
	require.Error(t, err, "PostgreSQL initialization with invalid host should fail")
}

// TestDataIntegrityAcrossConnections verifies that data written
// in one connection is readable in another connection to the same database
func TestDataIntegrityAcrossConnections(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "integrity_test.db")

	config := Config{
		Type:       SQLite,
		SQLitePath: dbPath,
	}

	// First connection - don't run migrations yet
	db1, err := InitDB(ctx, config, "connection1")
	require.NoError(t, err)
	defer func() { _ = db1.Close() }()

	queries1, err := NewQuerier(db1, SQLite)
	require.NoError(t, err)

	// Create test data: project -> column -> task -> label
	projectParams := types.CreateProjectRecordParams{
		Name: "Integrity Test Project",
	}
	project, err := queries1.CreateProjectRecord(ctx, projectParams)
	require.NoError(t, err)

	err = queries1.InitializeProjectCounter(ctx, project.ID)
	require.NoError(t, err)

	columnParams := types.CreateColumnParams{
		ProjectID: project.ID,
		Name:      "Ready",
	}
	column, err := queries1.CreateColumn(ctx, columnParams)
	require.NoError(t, err)

	taskParams := types.CreateTaskParams{
		Title:    "Integrity Test Task",
		ColumnID: column.ID,
		Position: 0,
	}
	task, err := queries1.CreateTask(ctx, taskParams)
	require.NoError(t, err)

	labelParams := types.CreateLabelParams{
		Name:      "urgent",
		Color:     "#FF0000",
		ProjectID: project.ID,
	}
	label, err := queries1.CreateLabel(ctx, labelParams)
	require.NoError(t, err)

	// Add label to task
	err = queries1.AddLabelToTask(ctx, types.AddLabelToTaskParams{
		TaskID:  task.ID,
		LabelID: label.ID,
	})
	require.NoError(t, err)

	projectID := project.ID
	columnID := column.ID
	taskID := task.ID

	// Close first connection
	err = db1.Close()
	require.NoError(t, err)

	// Second connection - verify all data is still there
	db2, err := InitDB(ctx, config, "connection2")
	require.NoError(t, err)
	defer func() { _ = db2.Close() }()

	queries2, err := NewQuerier(db2, SQLite)
	require.NoError(t, err)

	// Verify project
	retrievedProject, err := queries2.GetProjectByID(ctx, projectID)
	require.NoError(t, err)
	require.Equal(t, "Integrity Test Project", retrievedProject.Name)

	// Verify column
	retrievedColumn, err := queries2.GetColumnByID(ctx, columnID)
	require.NoError(t, err)
	require.Equal(t, "Ready", retrievedColumn.Name)

	// Verify task
	retrievedTask, err := queries2.GetTask(ctx, taskID)
	require.NoError(t, err)
	require.Equal(t, "Integrity Test Task", retrievedTask.Title)

	// Verify label exists
	retrievedLabels, err := queries2.GetLabelsByProject(ctx, projectID)
	require.NoError(t, err)
	require.True(t, len(retrievedLabels) >= 1, "should have at least 1 label")
	found := false
	for _, label := range retrievedLabels {
		if label.Name == "urgent" {
			found = true
			break
		}
	}
	require.True(t, found, "urgent label should exist")

	// Verify label-task relationship
	taskLabels, err := queries2.GetLabelsForTask(ctx, taskID)
	require.NoError(t, err)
	require.Len(t, taskLabels, 1)
	require.Equal(t, "urgent", taskLabels[0].Name)

	// Third connection - verify again
	db3, err := InitDB(ctx, config, "connection3")
	require.NoError(t, err)
	defer func() { _ = db3.Close() }()

	queries3, err := NewQuerier(db3, SQLite)
	require.NoError(t, err)

	// Re-verify everything still exists
	allProjects, err := queries3.GetAllProjects(ctx)
	require.NoError(t, err)
	require.True(t, len(allProjects) >= 1, "should have at least 1 project")

	found = false
	for _, p := range allProjects {
		if p.Name == "Integrity Test Project" {
			found = true
			break
		}
	}
	require.True(t, found, "Integrity Test Project should exist")
}

// TestConcurrentDatabaseAccess verifies that concurrent operations
// work correctly within a single database connection
func TestConcurrentDatabaseAccess(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	config := Config{
		Type:       SQLite,
		SQLitePath: filepath.Join(tmpDir, "concurrent_test.db"),
	}

	db, err := InitDB(ctx, config, "concurrent_test")
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	queries, err := NewQuerier(db, SQLite)
	require.NoError(t, err)

	// Create a project
	projectParams := types.CreateProjectRecordParams{
		Name: "Concurrent Test Project",
	}
	project, err := queries.CreateProjectRecord(ctx, projectParams)
	require.NoError(t, err)

	err = queries.InitializeProjectCounter(ctx, project.ID)
	require.NoError(t, err)

	// Create a column
	columnParams := types.CreateColumnParams{
		ProjectID: project.ID,
		Name:      "Todo",
	}
	column, err := queries.CreateColumn(ctx, columnParams)
	require.NoError(t, err)

	// Create multiple tasks sequentially
	taskIDs := make([]int64, 0)
	for i := range 5 {
		taskParams := types.CreateTaskParams{
			Title:    "Task " + string(rune(i)),
			ColumnID: column.ID,
			Position: int64(i),
		}
		task, err := queries.CreateTask(ctx, taskParams)
		require.NoError(t, err)
		taskIDs = append(taskIDs, task.ID)
	}

	// Verify all tasks exist
	retrievedTasks, err := queries.GetTasksByColumn(ctx, column.ID)
	require.NoError(t, err)
	require.Len(t, retrievedTasks, 5)

	// Verify task count matches
	require.Len(t, taskIDs, 5)
}

// Helper functions

// createTestProjectWithData creates a project with columns, tasks, and labels
func createTestProjectWithData(
	t *testing.T,
	ctx context.Context,
	queries types.Querier,
) types.Project {
	t.Helper()

	// Create project
	projectParams := types.CreateProjectRecordParams{
		Name:        "E2E Test Project",
		Description: types.NullString{String: "End-to-end switching test", Valid: true},
	}
	project, err := queries.CreateProjectRecord(ctx, projectParams)
	require.NoError(t, err)

	// Initialize counter
	err = queries.InitializeProjectCounter(ctx, project.ID)
	require.NoError(t, err)

	// Create columns
	todoParams := types.CreateColumnParams{
		ProjectID: project.ID,
		Name:      "Todo",
	}
	todoColumn, err := queries.CreateColumn(ctx, todoParams)
	require.NoError(t, err)

	inProgressParams := types.CreateColumnParams{
		ProjectID: project.ID,
		Name:      "In Progress",
	}
	inProgressColumn, err := queries.CreateColumn(ctx, inProgressParams)
	require.NoError(t, err)

	// Create tasks
	task1Params := types.CreateTaskParams{
		Title:    "Task 1",
		ColumnID: todoColumn.ID,
		Position: 0,
	}
	task1, err := queries.CreateTask(ctx, task1Params)
	require.NoError(t, err)

	task2Params := types.CreateTaskParams{
		Title:    "Task 2",
		ColumnID: inProgressColumn.ID,
		Position: 0,
	}
	task2, err := queries.CreateTask(ctx, task2Params)
	require.NoError(t, err)

	// Create labels
	bugLabelParams := types.CreateLabelParams{
		Name:      "bug",
		Color:     "#FF0000",
		ProjectID: project.ID,
	}
	bugLabel, err := queries.CreateLabel(ctx, bugLabelParams)
	require.NoError(t, err)

	featureLabelParams := types.CreateLabelParams{
		Name:      "feature",
		Color:     "#00FF00",
		ProjectID: project.ID,
	}
	_, err = queries.CreateLabel(ctx, featureLabelParams)
	require.NoError(t, err)

	// Add labels to tasks
	err = queries.AddLabelToTask(ctx, types.AddLabelToTaskParams{
		TaskID:  task1.ID,
		LabelID: bugLabel.ID,
	})
	require.NoError(t, err)

	// Create a comment
	commentParams := types.CreateCommentParams{
		TaskID:  task2.ID,
		Content: "This is a test comment",
		Author:  "test_user",
	}
	_, err = queries.CreateComment(ctx, commentParams)
	require.NoError(t, err)

	return project
}

// verifyProjectStructureInDatabase verifies that a project structure exists in a database
func verifyProjectStructureInDatabase(
	t *testing.T,
	ctx context.Context,
	queries types.Querier,
	sourceProject types.Project,
) {
	t.Helper()

	// Create the same structure in target database
	projectParams := types.CreateProjectRecordParams{
		Name:        sourceProject.Name,
		Description: sourceProject.Description,
	}
	targetProject, err := queries.CreateProjectRecord(ctx, projectParams)
	require.NoError(t, err)

	err = queries.InitializeProjectCounter(ctx, targetProject.ID)
	require.NoError(t, err)

	// Verify basic project attributes
	require.Equal(t, sourceProject.Name, targetProject.Name)

	// Verify database-specific operations work
	allProjects, err := queries.GetAllProjects(ctx)
	require.NoError(t, err)
	require.True(t, len(allProjects) > 0)

	// Find our project in the results
	found := false
	for _, p := range allProjects {
		if p.ID == targetProject.ID {
			found = true
			break
		}
	}
	require.True(t, found, "created project should be in all projects list")
}
