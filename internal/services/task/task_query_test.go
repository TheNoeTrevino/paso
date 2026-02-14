package task

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/database"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/testing/fixtures"
)

func TestGetTaskSummariesByProjectFiltered(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)
		// Create tasks with different titles
		_, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
			Title:    "Fix bug in login",
			ColumnID: env.ColumnID,
			Position: 0,
		})
		require.NoError(t, err)

		_, err = env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
			Title:    "Add new feature",
			ColumnID: env.ColumnID,
			Position: 1,
		})
		require.NoError(t, err)

		_, err = env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
			Title:    "Fix bug in signup",
			ColumnID: env.ColumnID,
			Position: 2,
		})
		require.NoError(t, err)

		// Filter by "bug"
		results, err := env.Svc.GetTaskSummariesByProjectFiltered(env.Ctx, env.ProjectID, "bug")
		require.NoError(t, err, "Operation failed")

		// Should return 2 tasks with "bug" in title
		totalTasks := 0
		for _, tasks := range results {
			totalTasks += len(tasks)
		}

		assert.Equal(t, 2, totalTasks)
	})
}

func TestGetTaskSummariesByProjectFiltered_NoResults(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)
		// Create task
		_, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
			Title:    "Test Task",
			ColumnID: env.ColumnID,
			Position: 0,
		})
		require.NoError(t, err)

		// Filter by non-existent term
		results, err := env.Svc.GetTaskSummariesByProjectFiltered(env.Ctx, env.ProjectID, "nonexistent")
		require.NoError(t, err, "Operation failed")

		// Should return empty map
		totalTasks := 0
		for _, tasks := range results {
			totalTasks += len(tasks)
		}

		assert.Equal(t, 0, totalTasks)
	})
}

func TestGetTaskSummariesByProjectFiltered_EmptyQuery(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)
		// Create tasks
		_, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
			Title:    "Task 1",
			ColumnID: env.ColumnID,
			Position: 0,
		})
		require.NoError(t, err)

		_, err = env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
			Title:    "Task 2",
			ColumnID: env.ColumnID,
			Position: 1,
		})
		require.NoError(t, err)

		// Filter with empty query (should return all)
		results, err := env.Svc.GetTaskSummariesByProjectFiltered(env.Ctx, env.ProjectID, "")
		require.NoError(t, err, "Operation failed")

		// Should return all tasks
		totalTasks := 0
		for _, tasks := range results {
			totalTasks += len(tasks)
		}

		assert.Equal(t, 2, totalTasks)
	})
}

func TestGetTaskReferencesForProject(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)
		// Create tasks
		_, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
			Title:    "Task 1",
			ColumnID: env.ColumnID,
			Position: 0,
		})
		require.NoError(t, err)

		_, err = env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
			Title:    "Task 2",
			ColumnID: env.ColumnID,
			Position: 1,
		})
		require.NoError(t, err)

		// Get task references
		refs, err := env.Svc.GetTaskReferencesForProject(env.Ctx, env.ProjectID)
		require.NoError(t, err, "Operation failed")

		assert.Len(t, refs, 2)
	})
}

func TestGetTaskReferencesForProject_EmptyProject(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)
		// Get task references for empty project
		refs, err := env.Svc.GetTaskReferencesForProject(env.Ctx, env.ProjectID)
		require.NoError(t, err, "Operation failed")

		assert.Len(t, refs, 0)
	})
}

func TestGetReadyTaskSummariesByProject_OnlyReadyColumn(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)
		// Create three columns: Todo (ready), In Progress, Done
		todoCol := fixtures.CreateColumnWithFlags(t, env.DB, env.Dialect, env.ProjectID, "Todo", true, false, false)
		inProgressCol := fixtures.CreateTestColumn(t, env.DB, env.Dialect, env.ProjectID, "In Progress")
		doneCol := fixtures.CreateTestColumn(t, env.DB, env.Dialect, env.ProjectID, "Done")

		// Create tasks in each column
		task1 := fixtures.CreateTestTask(t, env.DB, env.Dialect, todoCol, "Task in Todo")
		task2 := fixtures.CreateTestTask(t, env.DB, env.Dialect, inProgressCol, "Task in In Progress")
		task3 := fixtures.CreateTestTask(t, env.DB, env.Dialect, doneCol, "Task in Done")

		// Get ready tasks
		readyTasks, err := env.Svc.GetReadyTaskSummariesByProject(env.Ctx, env.ProjectID)
		require.NoError(t, err, "Operation failed")

		require.Len(t, readyTasks, 1)

		// Verify only task from Todo column is returned
		assert.Equal(t, task1, readyTasks[0].ID)

		// Verify tasks from other columns are not included
		for _, task := range readyTasks {
			assert.False(t, task.ID == task2 || task.ID == task3, "Unexpected task ID %d in ready tasks (should only be from Todo column)", task.ID)
		}
	})
}

func TestGetReadyTaskSummariesByProject_ExcludesBlockedTasks(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)
		// Create ready column
		readyCol := fixtures.CreateColumnWithFlags(t, env.DB, env.Dialect, env.ProjectID, "Todo", true, false, false)

		// Create two tasks
		task1 := fixtures.CreateTestTask(t, env.DB, env.Dialect, readyCol, "Unblocked Task")
		task2 := fixtures.CreateTestTask(t, env.DB, env.Dialect, readyCol, "Blocked Task")

		// Create a blocker relationship (task2 is blocked by task1)
		// relation_type_id = 2 is "Blocked By/Blocker" with is_blocking = 1
		_, err := env.DB.ExecContext(env.Ctx,
			fmt.Sprintf("INSERT INTO task_subtasks (parent_id, child_id, relation_type_id) VALUES (%s, %s, %s)",
				d.Placeholder(1), d.Placeholder(2), d.Placeholder(3)),
			task2, task1, 2)
		require.NoError(t, err, "failed to create blocking relationship")

		// Get ready tasks
		readyTasks, err := env.Svc.GetReadyTaskSummariesByProject(env.Ctx, env.ProjectID)
		require.NoError(t, err, "Operation failed")

		// Should only get task1 (task2 is blocked)
		require.Len(t, readyTasks, 1)

		assert.Equal(t, task1, readyTasks[0].ID)
	})
}

func TestGetReadyTaskSummariesByProject_EmptyWhenNoReadyColumn(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)
		// Create columns but none are ready
		col1 := fixtures.CreateTestColumn(t, env.DB, env.Dialect, env.ProjectID, "Todo")
		fixtures.CreateTestColumn(t, env.DB, env.Dialect, env.ProjectID, "Done")

		// Create tasks
		fixtures.CreateTestTask(t, env.DB, env.Dialect, col1, "Task 1")

		// Get ready tasks
		readyTasks, err := env.Svc.GetReadyTaskSummariesByProject(env.Ctx, env.ProjectID)
		require.NoError(t, err, "Operation failed")

		assert.Len(t, readyTasks, 0)
	})
}

func TestGetReadyTaskSummariesByProject_EmptyReadyColumn(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)
		// Create ready column with no tasks
		fixtures.CreateColumnWithFlags(t, env.DB, env.Dialect, env.ProjectID, "Todo", true, false, false)

		// Get ready tasks
		readyTasks, err := env.Svc.GetReadyTaskSummariesByProject(env.Ctx, env.ProjectID)
		require.NoError(t, err, "Operation failed")

		assert.Len(t, readyTasks, 0)
	})
}

func TestGetReadyTaskSummariesByProject_IncludesTaskDetails(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)
		// Create ready column
		readyCol := fixtures.CreateColumnWithFlags(t, env.DB, env.Dialect, env.ProjectID, "Todo", true, false, false)

		// Create label
		labelID := fixtures.CreateTestLabel(t, env.DB, env.Dialect, env.ProjectID, "bug", fixtures.DefaultTestLabelColor)

		// Create task with label and high priority (priority_id = 4)
		taskID := fixtures.CreateTestTask(t, env.DB, env.Dialect, readyCol, "Important Bug Fix")
		_, err := env.DB.ExecContext(env.Ctx,
			fmt.Sprintf("UPDATE tasks SET priority_id = 4 WHERE id = %s", d.Placeholder(1)), taskID)
		require.NoError(t, err, "failed to update task priority")

		// Attach label
		_, err = env.DB.ExecContext(env.Ctx,
			fmt.Sprintf("INSERT INTO task_labels (task_id, label_id) VALUES (%s, %s)", d.Placeholder(1), d.Placeholder(2)), taskID, labelID)
		require.NoError(t, err, "failed to attach label")

		// Get ready tasks
		readyTasks, err := env.Svc.GetReadyTaskSummariesByProject(env.Ctx, env.ProjectID)
		require.NoError(t, err, "Operation failed")

		require.Len(t, readyTasks, 1)

		task := readyTasks[0]

		// Verify task details
		assert.Equal(t, "Important Bug Fix", task.Title)
		assert.Equal(t, "high", task.PriorityDescription)

		require.Len(t, task.Labels, 1)

		assert.Equal(t, "bug", task.Labels[0].Name)
	})
}

func TestGetTaskActivities_RetrievesBothCommentsAndEvents(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env, mock := setupTestEnvWithEventMock(t, db, d, dbType)
		taskID := fixtures.CreateTestTask(t, env.DB, env.Dialect, env.ColumnID, "Test Task")
		// Create mock event service with predefined events
		mock.GetEventsByTaskResult = []models.TaskEvent{
			{
				ID:        1,
				TaskID:    taskID,
				Content:   "Task created",
				Author:    "system",
				CreatedAt: time.Now().Add(-2 * time.Hour),
			},
			{
				ID:        2,
				TaskID:    taskID,
				Content:   "Task moved from Todo to Done",
				Author:    "system",
				CreatedAt: time.Now().Add(-1 * time.Hour),
			},
		}
		// Create some comments directly in the database
		fixtures.CreateTestComment(t, env.DB, env.Dialect, taskID, "First comment", "user1")
		fixtures.CreateTestComment(t, env.DB, env.Dialect, taskID, "Second comment", "user2")
		activities, err := env.Svc.GetTaskActivities(env.Ctx, taskID)
		require.NoError(t, err)

		// Should have 4 activities total: 2 events + 2 comments
		require.Len(t, activities, 4)

		// Count events and comments
		eventCount := 0
		commentCount := 0
		for _, activity := range activities {
			switch activity.Type {
			case models.ActivityTypeEvent:
				eventCount++
			case models.ActivityTypeComment:
				commentCount++
			}
		}

		assert.Equal(t, 2, eventCount)
		assert.Equal(t, 2, commentCount)

		// Verify mock was called
		require.True(t, mock.HasCallWithId("GetEventsByTask", taskID))
	})
}

func TestGetTaskActivities_SortedByTimestampDescending(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env, mock := setupTestEnvWithEventMock(t, db, d, dbType)
		taskID := fixtures.CreateTestTask(t, env.DB, env.Dialect, env.ColumnID, "Test Task")
		// Create mock event service with events at specific times
		now := time.Now()
		mock.GetEventsByTaskResult = []models.TaskEvent{
			{
				ID:        1,
				TaskID:    taskID,
				Content:   "Oldest event",
				Author:    "system",
				CreatedAt: now.Add(-3 * time.Hour),
			},
			{
				ID:        2,
				TaskID:    taskID,
				Content:   "Newest event",
				Author:    "system",
				CreatedAt: now.Add(-30 * time.Minute),
			},
		}
		// Create comments with specific timestamps using Go time values
		twoHoursAgo := now.Add(-2 * time.Hour)
		oneHourAgo := now.Add(-1 * time.Hour)
		_, err := env.DB.ExecContext(env.Ctx,
			fmt.Sprintf("INSERT INTO task_comments (task_id, content, author, created_at) VALUES (%s, 'Middle comment', 'user1', %s), (%s, 'Recent comment', 'user2', %s)",
				d.Placeholder(1), d.Placeholder(2), d.Placeholder(3), d.Placeholder(4)),
			taskID, twoHoursAgo, taskID, oneHourAgo)
		require.NoError(t, err, "failed to create test comments")
		activities, err := env.Svc.GetTaskActivities(env.Ctx, taskID)
		require.NoError(t, err)

		require.Len(t, activities, 4)

		// Verify activities are sorted by CreatedAt descending (newest first)
		for i := 1; i < len(activities); i++ {
			assert.False(t, activities[i-1].CreatedAt.Before(activities[i].CreatedAt),
				"Activities not sorted correctly at index %d", i)
		}
	})
}

func TestGetTaskActivities_TaskWithNoActivities(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env, mock := setupTestEnvWithEventMock(t, db, d, dbType)
		taskID := fixtures.CreateTestTask(t, env.DB, env.Dialect, env.ColumnID, "Test Task")
		// Create mock with no events
		mock.GetEventsByTaskResult = []models.TaskEvent{}
		activities, err := env.Svc.GetTaskActivities(env.Ctx, taskID)
		require.NoError(t, err)

		assert.Len(t, activities, 0)
	})
}

func TestGetTaskActivities_TaskWithOnlyComments(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env, mock := setupTestEnvWithEventMock(t, db, d, dbType)
		taskID := fixtures.CreateTestTask(t, env.DB, env.Dialect, env.ColumnID, "Test Task")
		// Create mock with no events
		mock.GetEventsByTaskResult = []models.TaskEvent{}
		// Create comments
		fixtures.CreateTestComment(t, env.DB, env.Dialect, taskID, "Comment 1", "user1")
		fixtures.CreateTestComment(t, env.DB, env.Dialect, taskID, "Comment 2", "user2")
		fixtures.CreateTestComment(t, env.DB, env.Dialect, taskID, "Comment 3", "user3")
		activities, err := env.Svc.GetTaskActivities(env.Ctx, taskID)
		require.NoError(t, err)

		require.Len(t, activities, 3)

		// Verify all are comments
		for _, activity := range activities {
			assert.Equal(t, models.ActivityTypeComment, activity.Type)
		}
	})
}

func TestGetTaskActivities_TaskWithOnlyEvents(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env, mock := setupTestEnvWithEventMock(t, db, d, dbType)
		taskID := fixtures.CreateTestTask(t, env.DB, env.Dialect, env.ColumnID, "Test Task")
		// Create mock with events
		mock.GetEventsByTaskResult = []models.TaskEvent{
			{
				ID:        1,
				TaskID:    taskID,
				Content:   "Task created",
				Author:    "system",
				CreatedAt: time.Now().Add(-2 * time.Hour),
			},
			{
				ID:        2,
				TaskID:    taskID,
				Content:   "Label 'bug' added",
				Author:    "user1",
				CreatedAt: time.Now().Add(-1 * time.Hour),
			},
		}
		activities, err := env.Svc.GetTaskActivities(env.Ctx, taskID)
		require.NoError(t, err)

		require.Len(t, activities, 2)

		// Verify all are events
		for _, activity := range activities {
			assert.Equal(t, models.ActivityTypeEvent, activity.Type)
		}
	})
}

func TestGetTaskActivities_InvalidTaskID(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)
		_, err := env.Svc.GetTaskActivities(env.Ctx, 0)

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidTaskID)
	})
}

func TestGetTaskActivities_NegativeTaskID(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)
		_, err := env.Svc.GetTaskActivities(env.Ctx, -1)

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidTaskID)
	})
}

func TestGetTaskActivities_EventServiceError(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env, mock := setupTestEnvWithEventMock(t, db, d, dbType)
		taskID := fixtures.CreateTestTask(t, env.DB, env.Dialect, env.ColumnID, "Test Task")
		// Create mock that returns an error
		mock.GetEventsByTaskErr = errors.New("event service unavailable")
		_, err := env.Svc.GetTaskActivities(env.Ctx, taskID)

		require.Error(t, err)
		assert.ErrorContains(t, err, "failed to get task events")
	})
}

func TestGetTaskActivities_VerifiesActivityItemFields(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env, mock := setupTestEnvWithEventMock(t, db, d, dbType)
		taskID := fixtures.CreateTestTask(t, env.DB, env.Dialect, env.ColumnID, "Test Task")
		// Create mock with a specific event
		eventTime := time.Now().Add(-1 * time.Hour)
		mock.GetEventsByTaskResult = []models.TaskEvent{
			{
				ID:        42,
				TaskID:    taskID,
				Content:   "Task moved to Done",
				Author:    "testauthor",
				CreatedAt: eventTime,
			},
		}
		// Create a comment
		commentID := fixtures.CreateTestComment(t, env.DB, env.Dialect, taskID, "Test comment content", "commentauthor")
		activities, err := env.Svc.GetTaskActivities(env.Ctx, taskID)
		require.NoError(t, err)

		require.Len(t, activities, 2)

		// Find the event activity
		var eventActivity *models.ActivityItem
		var commentActivity *models.ActivityItem
		for i := range activities {
			if activities[i].Type == models.ActivityTypeEvent {
				eventActivity = &activities[i]
			} else {
				commentActivity = &activities[i]
			}
		}

		// Verify event activity fields
		require.NotNil(t, eventActivity, "Event activity not found")
		assert.Equal(t, 42, eventActivity.ID)
		assert.Equal(t, taskID, eventActivity.TaskID)
		assert.Equal(t, "Task moved to Done", eventActivity.Content)
		assert.Equal(t, "testauthor", eventActivity.Author)

		// Verify comment activity fields
		require.NotNil(t, commentActivity, "Comment activity not found")
		assert.Equal(t, commentID, commentActivity.ID)
		assert.Equal(t, taskID, commentActivity.TaskID)
		assert.Equal(t, "Test comment content", commentActivity.Content)
		assert.Equal(t, "commentauthor", commentActivity.Author)
	})
}

func TestGetTaskActivities_LargeNumberOfActivities(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env, mock := setupTestEnvWithEventMock(t, db, d, dbType)
		taskID := fixtures.CreateTestTask(t, env.DB, env.Dialect, env.ColumnID, "Test Task")
		// Create mock with many events
		events := make([]models.TaskEvent, 50)
		for i := 0; i < 50; i++ {
			events[i] = models.TaskEvent{
				ID:        i + 1,
				TaskID:    taskID,
				Content:   "Event " + string(rune('A'+i%26)),
				Author:    "system",
				CreatedAt: time.Now().Add(-time.Duration(i) * time.Minute),
			}
		}
		mock.GetEventsByTaskResult = events
		// Create many comments
		for i := 0; i < 50; i++ {
			fixtures.CreateTestComment(t, env.DB, env.Dialect, taskID, "Comment "+string(rune('A'+i%26)), "user")
		}
		activities, err := env.Svc.GetTaskActivities(env.Ctx, taskID)
		require.NoError(t, err)

		require.Len(t, activities, 100)

		// Verify sorting is maintained with large number of activities
		for i := 1; i < len(activities); i++ {
			assert.False(t, activities[i-1].CreatedAt.Before(activities[i].CreatedAt),
				"Activities not sorted correctly at index %d", i)
		}
	})
}

func TestGetInProgressTasksByProject(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)
		// Setup: Create project with in-progress column
		inProgressCol := fixtures.CreateColumnWithFlags(t, env.DB, env.Dialect, env.ProjectID, "In Progress", false, false, true)
		// Create multiple in-progress tasks with labels
		label1ID := fixtures.CreateTestLabel(t, env.DB, env.Dialect, env.ProjectID, "urgent", fixtures.DefaultTestLabelColor)
		label2ID := fixtures.CreateTestLabel(t, env.DB, env.Dialect, env.ProjectID, "review", fixtures.DefaultTestLabelColor)

		task1, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
			Title:    "Task 1",
			ColumnID: inProgressCol,
			Position: 0,
		})
		require.NoError(t, err)

		task2, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
			Title:    "Task 2",
			ColumnID: inProgressCol,
			Position: 1,
		})
		require.NoError(t, err)

		// Attach labels to tasks
		err = env.Svc.AttachLabel(env.Ctx, task1.ID, label1ID)
		require.NoError(t, err)
		err = env.Svc.AttachLabel(env.Ctx, task2.ID, label2ID)
		require.NoError(t, err)

		// Get in-progress tasks
		tasks, err := env.Svc.GetInProgressTasksByProject(env.Ctx, env.ProjectID)
		require.NoError(t, err)

		// Verify results
		require.Len(t, tasks, 2)

		// Check first task
		assert.Equal(t, "Task 1", tasks[0].Title)
		assert.Len(t, tasks[0].Labels, 1)
		assert.Equal(t, "urgent", tasks[0].Labels[0].Name)

		// Check second task
		assert.Equal(t, "Task 2", tasks[1].Title)
		assert.Len(t, tasks[1].Labels, 1)
		assert.Equal(t, "review", tasks[1].Labels[0].Name)
	})
}

func TestGetInProgressTasksByProject_InvalidProjectID(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)
		// Test with invalid project ID
		_, err := env.Svc.GetInProgressTasksByProject(env.Ctx, -1)
		assert.Error(t, err)

		_, err = env.Svc.GetInProgressTasksByProject(env.Ctx, 0)
		assert.Error(t, err)
	})
}

func TestGetInProgressTasksByProject_EmptyProject(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)
		// Project has no in-progress column
		tasks, err := env.Svc.GetInProgressTasksByProject(env.Ctx, env.ProjectID)
		require.NoError(t, err, "Operation failed")

		// Should return empty slice
		assert.Len(t, tasks, 0)
	})
}
