package tui

import (
	"context"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/testing/fixtures"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

func createTestActivityItems(count int, activityType models.ActivityType) []models.ActivityItem {
	items := make([]models.ActivityItem, count)
	baseTime := time.Now()
	for i := 0; i < count; i++ {
		items[i] = models.ActivityItem{
			ID:        i + 1,
			TaskID:    1,
			Type:      activityType,
			Content:   "Test content " + string(rune('A'+i)),
			Author:    "testuser",
			CreatedAt: baseTime.Add(-time.Duration(i) * time.Hour),
		}
	}
	return items
}

func createMixedActivityItems() []models.ActivityItem {
	baseTime := time.Now()
	return []models.ActivityItem{
		{
			ID:        1,
			TaskID:    1,
			Type:      models.ActivityTypeComment,
			Content:   "This is a comment",
			Author:    "user1",
			CreatedAt: baseTime,
		},
		{
			ID:        2,
			TaskID:    1,
			Type:      models.ActivityTypeEvent,
			Content:   "Task created",
			Author:    "system",
			CreatedAt: baseTime.Add(-1 * time.Hour),
		},
		{
			ID:        3,
			TaskID:    1,
			Type:      models.ActivityTypeComment,
			Content:   "Another comment",
			Author:    "user2",
			CreatedAt: baseTime.Add(-2 * time.Hour),
		},
		{
			ID:        4,
			TaskID:    1,
			Type:      models.ActivityTypeEvent,
			Content:   "Priority changed to high",
			Author:    "system",
			CreatedAt: baseTime.Add(-3 * time.Hour),
		},
	}
}

func TestActivityView_NavigateUp(t *testing.T) {
	t.Parallel()
	m, _ := SetupTestModelWithDB(t)

	m.UIState.Mode = state.CommentsViewMode
	m.Forms.Comment.SetActivities(createTestActivityItems(5, models.ActivityTypeComment))
	m.Forms.Comment.Cursor = 2

	msg := tea.KeyPressMsg(tea.Key{Code: tea.KeyUp})
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)

	assert.Equal(t, 1, m.Forms.Comment.Cursor, "Cursor should move up by 1")
}

func TestActivityView_NavigateUpWithK(t *testing.T) {
	t.Parallel()
	m, _ := SetupTestModelWithDB(t)

	m.UIState.Mode = state.CommentsViewMode
	m.Forms.Comment.SetActivities(createTestActivityItems(5, models.ActivityTypeComment))
	m.Forms.Comment.Cursor = 3

	msg := tea.KeyPressMsg(tea.Key{Text: "k", Code: 'k'})
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)

	assert.Equal(t, 2, m.Forms.Comment.Cursor, "Cursor should move up by 1 with 'k'")
}

func TestActivityView_NavigateDown(t *testing.T) {
	t.Parallel()
	m, _ := SetupTestModelWithDB(t)

	m.UIState.Mode = state.CommentsViewMode
	m.Forms.Comment.SetActivities(createTestActivityItems(5, models.ActivityTypeComment))
	m.Forms.Comment.Cursor = 1

	msg := tea.KeyPressMsg(tea.Key{Code: tea.KeyDown})
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)

	assert.Equal(t, 2, m.Forms.Comment.Cursor, "Cursor should move down by 1")
}

func TestActivityView_NavigateDownWithJ(t *testing.T) {
	t.Parallel()
	m, _ := SetupTestModelWithDB(t)

	m.UIState.Mode = state.CommentsViewMode
	m.Forms.Comment.SetActivities(createTestActivityItems(5, models.ActivityTypeComment))
	m.Forms.Comment.Cursor = 0

	msg := tea.KeyPressMsg(tea.Key{Text: "j", Code: 'j'})
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)

	assert.Equal(t, 1, m.Forms.Comment.Cursor, "Cursor should move down by 1 with 'j'")
}

func TestActivityView_CursorBoundsAtTop(t *testing.T) {
	t.Parallel()
	m, _ := SetupTestModelWithDB(t)

	m.UIState.Mode = state.CommentsViewMode
	m.Forms.Comment.SetActivities(createTestActivityItems(5, models.ActivityTypeComment))
	m.Forms.Comment.Cursor = 0

	msg := tea.KeyPressMsg(tea.Key{Code: tea.KeyUp})
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)

	assert.Equal(t, 0, m.Forms.Comment.Cursor, "Cursor should not go below 0")
}

func TestActivityView_CursorBoundsAtBottom(t *testing.T) {
	t.Parallel()
	m, _ := SetupTestModelWithDB(t)

	m.UIState.Mode = state.CommentsViewMode
	activities := createTestActivityItems(5, models.ActivityTypeComment)
	m.Forms.Comment.SetActivities(activities)
	m.Forms.Comment.Cursor = len(activities) - 1

	msg := tea.KeyPressMsg(tea.Key{Code: tea.KeyDown})
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)

	assert.Equal(t, len(activities)-1, m.Forms.Comment.Cursor, "Cursor should not exceed item count")
}

func TestActivityView_EditEventShowsWarning(t *testing.T) {
	t.Parallel()
	m, _ := SetupTestModelWithDB(t)

	m.UIState.Mode = state.CommentsViewMode
	m.Forms.Comment.SetActivities(createTestActivityItems(3, models.ActivityTypeEvent))
	m.Forms.Comment.Cursor = 0

	initialMode := m.UIState.Mode

	msg := tea.KeyPressMsg(tea.Key{Text: "e", Code: 'e'})
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)

	assert.Equal(t, initialMode, m.UIState.Mode, "Mode should not change when editing an event")
	assert.NotEqual(t, state.CommentFormMode, m.UIState.Mode, "Should not enter form mode for events")
}

func TestActivityView_DeleteEventShowsWarning(t *testing.T) {
	t.Parallel()
	m, _ := SetupTestModelWithDB(t)

	m.UIState.Mode = state.CommentsViewMode
	m.Forms.Comment.SetActivities(createTestActivityItems(3, models.ActivityTypeEvent))
	m.Forms.Comment.Cursor = 0

	initialMode := m.UIState.Mode

	msg := tea.KeyPressMsg(tea.Key{Text: "d", Code: 'd'})
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)

	assert.Equal(t, initialMode, m.UIState.Mode, "Mode should not change when deleting an event")
	assert.NotEqual(t, state.DeleteConfirmMode, m.UIState.Mode, "Should not enter delete mode for events")
}

func TestActivityView_EditCommentEntersFormMode(t *testing.T) {
	t.Parallel()
	m, _ := SetupTestModelWithDB(t)

	m.UIState.Mode = state.CommentsViewMode
	m.Forms.Comment.SetActivities(createTestActivityItems(3, models.ActivityTypeComment))
	m.Forms.Comment.Cursor = 0

	msg := tea.KeyPressMsg(tea.Key{Text: "e", Code: 'e'})
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)

	assert.Equal(t, state.CommentFormMode, m.UIState.Mode, "Should enter form mode when editing a comment")
}

func TestActivityView_DeleteCommentEntersDeleteMode(t *testing.T) {
	t.Parallel()
	m, _ := SetupTestModelWithDB(t)

	m.UIState.Mode = state.CommentsViewMode
	m.Forms.Comment.SetActivities(createTestActivityItems(3, models.ActivityTypeComment))
	m.Forms.Comment.Cursor = 0

	msg := tea.KeyPressMsg(tea.Key{Text: "d", Code: 'd'})
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)

	assert.Equal(t, state.DeleteConfirmMode, m.UIState.Mode, "Should enter delete confirmation mode for comments")
}

func TestActivityView_MixedItems_EditEventBlocked(t *testing.T) {
	t.Parallel()
	m, _ := SetupTestModelWithDB(t)

	m.UIState.Mode = state.CommentsViewMode
	m.Forms.Comment.SetActivities(createMixedActivityItems())
	m.Forms.Comment.Cursor = 1

	selectedActivity := m.Forms.Comment.GetSelectedActivity()
	require.NotNil(t, selectedActivity)
	require.Equal(t, models.ActivityTypeEvent, selectedActivity.Type, "Should have selected an event")

	msg := tea.KeyPressMsg(tea.Key{Text: "e", Code: 'e'})
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)

	assert.Equal(t, state.CommentsViewMode, m.UIState.Mode, "Should remain in comments view mode")
}

func TestActivityView_MixedItems_EditCommentAllowed(t *testing.T) {
	t.Parallel()
	m, _ := SetupTestModelWithDB(t)

	m.UIState.Mode = state.CommentsViewMode
	m.Forms.Comment.SetActivities(createMixedActivityItems())
	m.Forms.Comment.Cursor = 0

	selectedActivity := m.Forms.Comment.GetSelectedActivity()
	require.NotNil(t, selectedActivity)
	require.Equal(t, models.ActivityTypeComment, selectedActivity.Type, "Should have selected a comment")

	msg := tea.KeyPressMsg(tea.Key{Text: "e", Code: 'e'})
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)

	assert.Equal(t, state.CommentFormMode, m.UIState.Mode, "Should enter form mode for comments")
}

func TestActivityView_AddNewComment(t *testing.T) {
	t.Parallel()
	m, _ := SetupTestModelWithDB(t)

	m.UIState.Mode = state.CommentsViewMode
	m.Forms.Comment.SetActivities(createMixedActivityItems())

	msg := tea.KeyPressMsg(tea.Key{Text: "a", Code: 'a'})
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)

	assert.Equal(t, state.CommentFormMode, m.UIState.Mode, "Should enter form mode for adding comment")
	assert.Equal(t, 0, m.Forms.Form.EditingCommentID, "EditingCommentID should be 0 for new comment")
}

func TestActivityView_EscapeClosesView(t *testing.T) {
	t.Parallel()
	m, _ := SetupTestModelWithDB(t)

	m.UIState.Mode = state.CommentsViewMode
	m.Forms.Comment.SetActivities(createMixedActivityItems())

	msg := tea.KeyPressMsg(tea.Key{Code: tea.KeyEscape})
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)

	assert.Equal(t, state.TicketFormMode, m.UIState.Mode, "Escape should return to task form mode")
}

func TestActivityView_QClosesView(t *testing.T) {
	t.Parallel()
	m, _ := SetupTestModelWithDB(t)

	m.UIState.Mode = state.CommentsViewMode
	m.Forms.Comment.SetActivities(createMixedActivityItems())

	msg := tea.KeyPressMsg(tea.Key{Text: "q", Code: 'q'})
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)

	assert.Equal(t, state.TicketFormMode, m.UIState.Mode, "'q' should return to task form mode")
}

func TestActivityView_EnterAlsoEdits(t *testing.T) {
	t.Parallel()
	m, _ := SetupTestModelWithDB(t)

	m.UIState.Mode = state.CommentsViewMode
	m.Forms.Comment.SetActivities(createTestActivityItems(3, models.ActivityTypeComment))
	m.Forms.Comment.Cursor = 0

	msg := tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter})
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)

	assert.Equal(t, state.CommentFormMode, m.UIState.Mode, "Enter should also edit selected comment")
}

func TestActivityView_EnterOnEventBlocked(t *testing.T) {
	t.Parallel()
	m, _ := SetupTestModelWithDB(t)

	m.UIState.Mode = state.CommentsViewMode
	m.Forms.Comment.SetActivities(createTestActivityItems(3, models.ActivityTypeEvent))
	m.Forms.Comment.Cursor = 0

	msg := tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter})
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)

	assert.Equal(t, state.CommentsViewMode, m.UIState.Mode, "Enter on event should not enter form mode")
}

func TestActivityView_EmptyList(t *testing.T) {
	t.Parallel()
	m, _ := SetupTestModelWithDB(t)

	m.UIState.Mode = state.CommentsViewMode
	m.Forms.Comment.SetActivities([]models.ActivityItem{})

	msg := tea.KeyPressMsg(tea.Key{Code: tea.KeyDown})
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)

	assert.Equal(t, state.CommentsViewMode, m.UIState.Mode, "Should stay in mode with empty list")
	assert.Equal(t, 0, m.Forms.Comment.Cursor, "Cursor should be 0 on empty list")
}

func TestActivityView_EditWithNoSelection(t *testing.T) {
	t.Parallel()
	m, _ := SetupTestModelWithDB(t)

	m.UIState.Mode = state.CommentsViewMode
	m.Forms.Comment.SetActivities([]models.ActivityItem{})
	m.Forms.Comment.Cursor = 0

	initialMode := m.UIState.Mode

	msg := tea.KeyPressMsg(tea.Key{Text: "e", Code: 'e'})
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)

	assert.Equal(t, initialMode, m.UIState.Mode, "Should not change mode when nothing selected")
}

func TestActivityView_DeleteWithNoSelection(t *testing.T) {
	t.Parallel()
	m, _ := SetupTestModelWithDB(t)

	m.UIState.Mode = state.CommentsViewMode
	m.Forms.Comment.SetActivities([]models.ActivityItem{})
	m.Forms.Comment.Cursor = 0

	initialMode := m.UIState.Mode

	msg := tea.KeyPressMsg(tea.Key{Text: "d", Code: 'd'})
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)

	assert.Equal(t, initialMode, m.UIState.Mode, "Should not change mode when nothing selected")
}

func TestActivityView_NavigationSequence(t *testing.T) {
	t.Parallel()
	m, _ := SetupTestModelWithDB(t)

	m.UIState.Mode = state.CommentsViewMode
	activities := createTestActivityItems(5, models.ActivityTypeComment)
	m.Forms.Comment.SetActivities(activities)
	m.Forms.Comment.Cursor = 0

	msg := tea.KeyPressMsg(tea.Key{Code: tea.KeyDown})
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)
	assert.Equal(t, 1, m.Forms.Comment.Cursor)

	msg = tea.KeyPressMsg(tea.Key{Code: tea.KeyDown})
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(Model)
	assert.Equal(t, 2, m.Forms.Comment.Cursor)

	msg = tea.KeyPressMsg(tea.Key{Code: tea.KeyUp})
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(Model)
	assert.Equal(t, 1, m.Forms.Comment.Cursor)

	msg = tea.KeyPressMsg(tea.Key{Code: tea.KeyUp})
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(Model)
	assert.Equal(t, 0, m.Forms.Comment.Cursor)
}

func TestMergeActivities_SortedByCreatedAtDescending(t *testing.T) {
	t.Parallel()
	baseTime := time.Now()

	events := []models.TaskEvent{
		{ID: 1, TaskID: 1, Content: "Event 1", Author: "system", CreatedAt: baseTime.Add(-1 * time.Hour)},
		{ID: 2, TaskID: 1, Content: "Event 2", Author: "system", CreatedAt: baseTime.Add(-3 * time.Hour)},
	}

	comments := []models.Comment{
		{ID: 1, TaskID: 1, Message: "Comment 1", Author: "user", CreatedAt: baseTime, UpdatedAt: baseTime},
		{ID: 2, TaskID: 1, Message: "Comment 2", Author: "user", CreatedAt: baseTime.Add(-2 * time.Hour), UpdatedAt: baseTime.Add(-2 * time.Hour)},
	}

	merged := models.MergeActivities(events, comments)

	require.Len(t, merged, 4, "Should have 4 merged activities")

	for i := 0; i < len(merged)-1; i++ {
		assert.True(t, merged[i].CreatedAt.After(merged[i+1].CreatedAt) || merged[i].CreatedAt.Equal(merged[i+1].CreatedAt),
			"Activities should be sorted newest first")
	}

	assert.Equal(t, models.ActivityTypeComment, merged[0].Type, "Newest item should be Comment 1")
	assert.Equal(t, models.ActivityTypeEvent, merged[1].Type, "Second item should be Event 1")
	assert.Equal(t, models.ActivityTypeComment, merged[2].Type, "Third item should be Comment 2")
	assert.Equal(t, models.ActivityTypeEvent, merged[3].Type, "Fourth item should be Event 2")
}

func TestMergeActivities_EmptyInputs(t *testing.T) {
	t.Parallel()
	t.Run("empty events", func(t *testing.T) {
		comments := []models.Comment{
			{ID: 1, TaskID: 1, Message: "Comment", Author: "user", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		}
		merged := models.MergeActivities(nil, comments)
		assert.Len(t, merged, 1, "Should have 1 activity from comments")
	})

	t.Run("empty comments", func(t *testing.T) {
		events := []models.TaskEvent{
			{ID: 1, TaskID: 1, Content: "Event", Author: "system", CreatedAt: time.Now()},
		}
		merged := models.MergeActivities(events, nil)
		assert.Len(t, merged, 1, "Should have 1 activity from events")
	})

	t.Run("both empty", func(t *testing.T) {
		merged := models.MergeActivities(nil, nil)
		assert.Len(t, merged, 0, "Should have 0 activities")
	})
}

func TestCommentState_SetActivities(t *testing.T) {
	t.Parallel()
	cs := state.NewCommentState()

	activities := createTestActivityItems(5, models.ActivityTypeComment)
	cs.SetActivities(activities)

	assert.Len(t, cs.Items, 5, "Should have 5 items")
	assert.Equal(t, 0, cs.Cursor, "Cursor should start at 0")
}

func TestCommentState_SetActivities_ResetsCursor(t *testing.T) {
	t.Parallel()
	cs := state.NewCommentState()

	activities := createTestActivityItems(5, models.ActivityTypeComment)
	cs.SetActivities(activities)
	cs.Cursor = 4

	newActivities := createTestActivityItems(2, models.ActivityTypeComment)
	cs.SetActivities(newActivities)

	assert.Equal(t, 1, cs.Cursor, "Cursor should be reset to max valid index")
}

func TestCommentState_GetSelectedActivity(t *testing.T) {
	t.Parallel()
	cs := state.NewCommentState()

	activities := createTestActivityItems(3, models.ActivityTypeComment)
	cs.SetActivities(activities)
	cs.Cursor = 1

	selected := cs.GetSelectedActivity()
	require.NotNil(t, selected)
	assert.Equal(t, 2, selected.ID, "Should return activity at cursor position")
}

func TestCommentState_GetSelectedActivity_Empty(t *testing.T) {
	t.Parallel()
	cs := state.NewCommentState()

	selected := cs.GetSelectedActivity()
	assert.Nil(t, selected, "Should return nil when no activities")
}

func TestCommentState_MoveCursorUp(t *testing.T) {
	t.Parallel()
	cs := state.NewCommentState()
	cs.SetActivities(createTestActivityItems(5, models.ActivityTypeComment))
	cs.Cursor = 2

	moved := cs.MoveCursorUp()
	assert.True(t, moved, "Should return true when cursor moved")
	assert.Equal(t, 1, cs.Cursor)
}

func TestCommentState_MoveCursorUp_AtTop(t *testing.T) {
	t.Parallel()
	cs := state.NewCommentState()
	cs.SetActivities(createTestActivityItems(5, models.ActivityTypeComment))
	cs.Cursor = 0

	moved := cs.MoveCursorUp()
	assert.False(t, moved, "Should return false when at top")
	assert.Equal(t, 0, cs.Cursor)
}

func TestCommentState_MoveCursorDown(t *testing.T) {
	t.Parallel()
	cs := state.NewCommentState()
	cs.SetActivities(createTestActivityItems(5, models.ActivityTypeComment))
	cs.Cursor = 2

	moved := cs.MoveCursorDown(4)
	assert.True(t, moved, "Should return true when cursor moved")
	assert.Equal(t, 3, cs.Cursor)
}

func TestCommentState_MoveCursorDown_AtBottom(t *testing.T) {
	t.Parallel()
	cs := state.NewCommentState()
	cs.SetActivities(createTestActivityItems(5, models.ActivityTypeComment))
	cs.Cursor = 4

	moved := cs.MoveCursorDown(4)
	assert.False(t, moved, "Should return false when at bottom")
	assert.Equal(t, 4, cs.Cursor)
}

func TestCommentState_EnsureCommentVisible_ScrollUp(t *testing.T) {
	t.Parallel()
	cs := state.NewCommentState()
	cs.SetActivities(createTestActivityItems(10, models.ActivityTypeComment))
	cs.ScrollOffset = 5
	cs.Cursor = 3

	cs.EnsureCommentVisible(5)

	assert.Equal(t, 3, cs.ScrollOffset, "ScrollOffset should be adjusted to show cursor")
}

func TestCommentState_EnsureCommentVisible_ScrollDown(t *testing.T) {
	t.Parallel()
	cs := state.NewCommentState()
	cs.SetActivities(createTestActivityItems(10, models.ActivityTypeComment))
	cs.ScrollOffset = 0
	cs.Cursor = 7

	cs.EnsureCommentVisible(5)

	assert.Equal(t, 3, cs.ScrollOffset, "ScrollOffset should be adjusted to show cursor")
}

func TestCommentState_DeleteSelected(t *testing.T) {
	t.Parallel()
	cs := state.NewCommentState()
	cs.SetActivities(createTestActivityItems(5, models.ActivityTypeComment))
	cs.Cursor = 2

	cs.DeleteSelected()

	assert.Len(t, cs.Items, 4, "Should have 4 items after delete")
	assert.Equal(t, 2, cs.Cursor, "Cursor should stay at same position")
}

func TestCommentState_DeleteSelected_AtEnd(t *testing.T) {
	t.Parallel()
	cs := state.NewCommentState()
	cs.SetActivities(createTestActivityItems(5, models.ActivityTypeComment))
	cs.Cursor = 4

	cs.DeleteSelected()

	assert.Len(t, cs.Items, 4, "Should have 4 items after delete")
	assert.Equal(t, 3, cs.Cursor, "Cursor should adjust when at end")
}

func TestCommentState_DeleteSelected_LastItem(t *testing.T) {
	t.Parallel()
	cs := state.NewCommentState()
	cs.SetActivities(createTestActivityItems(1, models.ActivityTypeComment))
	cs.Cursor = 0

	cs.DeleteSelected()

	assert.Len(t, cs.Items, 0, "Should have 0 items after delete")
	assert.Equal(t, 0, cs.Cursor, "Cursor should be 0 on empty list")
}

func TestCommentState_Clear(t *testing.T) {
	t.Parallel()
	cs := state.NewCommentState()
	cs.SetActivities(createTestActivityItems(5, models.ActivityTypeComment))
	cs.Cursor = 3
	cs.TaskID = 42
	cs.ScrollOffset = 2

	cs.Clear()

	assert.Len(t, cs.Items, 0, "Items should be empty")
	assert.Equal(t, 0, cs.Cursor, "Cursor should be 0")
	assert.Equal(t, 0, cs.TaskID, "TaskID should be 0")
	assert.Equal(t, 0, cs.ScrollOffset, "ScrollOffset should be 0")
}

func TestCommentState_IsEmpty(t *testing.T) {
	t.Parallel()
	cs := state.NewCommentState()

	assert.True(t, cs.IsEmpty(), "Should be empty initially")

	cs.SetActivities(createTestActivityItems(1, models.ActivityTypeComment))
	assert.False(t, cs.IsEmpty(), "Should not be empty after setting activities")
}

func TestActivityItem_FromEvent(t *testing.T) {
	t.Parallel()
	event := models.TaskEvent{
		ID:        1,
		TaskID:    42,
		Content:   "Task created",
		Author:    "system",
		CreatedAt: time.Now(),
	}

	item := models.NewActivityItemFromEvent(event)

	assert.Equal(t, event.ID, item.ID)
	assert.Equal(t, event.TaskID, item.TaskID)
	assert.Equal(t, models.ActivityTypeEvent, item.Type)
	assert.Equal(t, event.Content, item.Content)
	assert.Equal(t, event.Author, item.Author)
	assert.Equal(t, event.CreatedAt, item.CreatedAt)
	assert.Nil(t, item.UpdatedAt, "Events should not have UpdatedAt")
}

func TestActivityItem_FromComment(t *testing.T) {
	t.Parallel()
	now := time.Now()
	later := now.Add(time.Hour)
	comment := models.Comment{
		ID:        1,
		TaskID:    42,
		Message:   "This is a comment",
		Author:    "user",
		CreatedAt: now,
		UpdatedAt: later,
	}

	item := models.NewActivityItemFromComment(comment)

	assert.Equal(t, comment.ID, item.ID)
	assert.Equal(t, comment.TaskID, item.TaskID)
	assert.Equal(t, models.ActivityTypeComment, item.Type)
	assert.Equal(t, comment.Message, item.Content)
	assert.Equal(t, comment.Author, item.Author)
	assert.Equal(t, comment.CreatedAt, item.CreatedAt)
	require.NotNil(t, item.UpdatedAt)
	assert.Equal(t, later, *item.UpdatedAt)
}

func TestActivityType_String(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "Event", models.ActivityTypeEvent.String())
	assert.Equal(t, "Comment", models.ActivityTypeComment.String())
	assert.Equal(t, "Unknown", models.ActivityType(99).String())
}

func TestActivityView_DataLoading_WithRealDB(t *testing.T) {
	t.Parallel()
	m, db := SetupTestModelWithDB(t)
	ctx := context.Background()

	projectID := m.AppState.GetCurrentProjectID()
	columns, err := m.App.ColumnService.GetColumnsByProject(ctx, projectID)
	require.NoError(t, err)
	require.NotEmpty(t, columns)

	taskID := fixtures.CreateTestTask(t, db, fixtures.SQLiteDialect(), columns[0].ID, "Test Task")

	_, err = db.ExecContext(ctx, `
		INSERT INTO task_comments (task_id, content, author) 
		VALUES (?, 'Test comment 1', 'user1'), (?, 'Test comment 2', 'user2')
	`, taskID, taskID)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO task_events (task_id, content, author)
		VALUES (?, 'Task created', 'system')
	`, taskID)
	require.NoError(t, err)

	var comments []models.Comment
	rows, err := db.QueryContext(ctx, `
		SELECT id, task_id, content, author, created_at, updated_at 
		FROM task_comments WHERE task_id = ?
	`, taskID)
	require.NoError(t, err)

	for rows.Next() {
		var c models.Comment
		err := rows.Scan(&c.ID, &c.TaskID, &c.Message, &c.Author, &c.CreatedAt, &c.UpdatedAt)
		require.NoError(t, err)
		comments = append(comments, c)
	}
	require.NoError(t, rows.Err())
	require.NoError(t, rows.Close())

	var events []models.TaskEvent
	rows2, err := db.QueryContext(ctx, `
		SELECT id, task_id, content, author, created_at 
		FROM task_events WHERE task_id = ?
	`, taskID)
	require.NoError(t, err)

	for rows2.Next() {
		var e models.TaskEvent
		err := rows2.Scan(&e.ID, &e.TaskID, &e.Content, &e.Author, &e.CreatedAt)
		require.NoError(t, err)
		events = append(events, e)
	}
	require.NoError(t, rows2.Err())
	require.NoError(t, rows2.Close())

	activities := models.MergeActivities(events, comments)
	m.Forms.Comment.SetActivities(activities)

	assert.Len(t, m.Forms.Comment.Items, 3, "Should have 3 activity items (2 comments + 1 event)")

	for i := 0; i < len(m.Forms.Comment.Items)-1; i++ {
		item1 := m.Forms.Comment.Items[i].Activity
		item2 := m.Forms.Comment.Items[i+1].Activity
		assert.True(t, item1.CreatedAt.After(item2.CreatedAt) || item1.CreatedAt.Equal(item2.CreatedAt),
			"Activities should be sorted newest first")
	}
}
