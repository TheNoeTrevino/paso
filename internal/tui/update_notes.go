package tui

import (
	"log/slog"

	tea "charm.land/bubbletea/v2"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui/huhforms"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// updateCommentEdit handles keyboard input in the comment editing mode.
// This function processes navigation (up/down), opening forms for editing, and comment deletion.
func (m Model) updateCommentEdit(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	km := m.Config.KeyMappings

	switch msg.String() {
	case "esc":
		// Return to task form mode
		m.UIState.Mode = state.TaskFormMode
		return m, nil

	case km.Navigation.MoveDown, "down":
		// Move cursor down
		if len(m.Forms.Comment.Items) > 0 {
			m.Forms.Comment.MoveCursorDown(len(m.Forms.Comment.Items) - 1)
		}
		return m, nil

	case km.Navigation.MoveUp, "up":
		// Move cursor up
		m.Forms.Comment.MoveCursorUp()
		return m, nil

	case "enter":
		// Open form to edit the selected activity (only comments are editable)
		if len(m.Forms.Comment.Items) > 0 && m.Forms.Comment.Cursor < len(m.Forms.Comment.Items) {
			activity := m.Forms.Comment.Items[m.Forms.Comment.Cursor].Activity
			// Events are read-only and cannot be edited
			if activity.Type == models.ActivityTypeEvent {
				return m, m.addNotification(state.LevelWarning, "Events cannot be edited")
			}
			m.Forms.Form.FormCommentMessage = activity.Content
			m.Forms.Form.EditingCommentID = activity.ID
			m.Forms.Form.CommentForm = huhforms.CreateCommentForm(&m.Forms.Form.FormCommentMessage, true).
				WithTheme(huhforms.CreatePasoTheme(m.Config.ColorScheme))
			m.Forms.Form.SnapshotCommentFormInitialValues()
			m.UIState.Mode = state.CommentFormMode
			return m, m.Forms.Form.CommentForm.Init()
		}
		return m, nil

	case km.Forms.OpenCommentsView, "n":
		// Open form to create a new comment
		m.Forms.Form.FormCommentMessage = ""
		m.Forms.Form.EditingCommentID = 0
		m.Forms.Form.CommentForm = huhforms.CreateCommentForm(&m.Forms.Form.FormCommentMessage, false).
			WithTheme(huhforms.CreatePasoTheme(m.Config.ColorScheme))
		m.Forms.Form.SnapshotCommentFormInitialValues()
		m.UIState.Mode = state.CommentFormMode
		return m, m.Forms.Form.CommentForm.Init()

	case "delete", km.Tasks.DeleteTask:
		// Delete the selected activity (only comments are deletable)
		if len(m.Forms.Comment.Items) > 0 && m.Forms.Comment.Cursor < len(m.Forms.Comment.Items) {
			activity := m.Forms.Comment.Items[m.Forms.Comment.Cursor].Activity
			// Events are read-only and cannot be deleted
			if activity.Type == models.ActivityTypeEvent {
				return m, m.addNotification(state.LevelWarning, "Events cannot be deleted")
			}

			ctx, cancel := m.DBContext()
			defer cancel()

			commentID := activity.ID
			err := m.App.TaskService.DeleteComment(ctx, commentID)
			if err != nil {
				slog.Error("failed to deleting comment", "error", err)
				return m, m.addNotification(state.LevelError, "Failed to delete comment")
			}

			// Reload comments from database
			taskID := m.Forms.Form.EditingTaskID
			comments, err := m.App.TaskService.GetCommentsByTask(ctx, taskID)
			if err != nil {
				slog.Error("failed to reloading comments", "error", err)
				return m, m.addNotification(state.LevelError, "Failed to reload comments")
			}

			// Update form state
			m.Forms.Form.FormComments = comments
			// Reload activities for comments view and task form preview
			activities, actErr := m.App.TaskService.GetTaskActivities(ctx, taskID)
			if actErr == nil {
				m.Forms.Comment.SetActivities(activities)
				m.Forms.Form.FormActivities = activities
			}

			// Adjust cursor if needed
			if m.Forms.Comment.Cursor >= len(m.Forms.Comment.Items) && len(m.Forms.Comment.Items) > 0 {
				m.Forms.Comment.Cursor = len(m.Forms.Comment.Items) - 1
			}
			if len(m.Forms.Comment.Items) == 0 {
				m.Forms.Comment.Cursor = 0
			}

			notifCmd := m.addNotification(state.LevelInfo, "Comment deleted")
			return m, notifCmd
		}
		return m, nil
	}

	return m, nil
}
