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
	switch msg.String() {
	case "esc":
		// Return to ticket form mode
		m.UIState.SetMode(state.TicketFormMode)
		return m, nil

	case "j", "down":
		// Move cursor down
		if len(m.CommentState.Items) > 0 {
			m.CommentState.MoveCursorDown(len(m.CommentState.Items) - 1)
		}
		return m, nil

	case "k", "up":
		// Move cursor up
		m.CommentState.MoveCursorUp()
		return m, nil

	case "enter":
		// Open form to edit the selected comment
		if len(m.CommentState.Items) > 0 && m.CommentState.Cursor < len(m.CommentState.Items) {
			comment := m.CommentState.Items[m.CommentState.Cursor].Comment
			m.FormState.FormCommentMessage = comment.Message
			m.FormState.EditingCommentID = comment.ID
			m.FormState.CommentForm = huhforms.CreateCommentForm(&m.FormState.FormCommentMessage, true).
				WithTheme(huhforms.CreatePasoTheme(m.Config.ColorScheme))
			m.FormState.SnapshotCommentFormInitialValues()
			m.UIState.SetMode(state.CommentFormMode)
			return m, m.FormState.CommentForm.Init()
		}
		return m, nil

	case "ctrl+n", "n":
		// Open form to create a new comment
		m.FormState.FormCommentMessage = ""
		m.FormState.EditingCommentID = 0
		m.FormState.CommentForm = huhforms.CreateCommentForm(&m.FormState.FormCommentMessage, false).
			WithTheme(huhforms.CreatePasoTheme(m.Config.ColorScheme))
		m.FormState.SnapshotCommentFormInitialValues()
		m.UIState.SetMode(state.CommentFormMode)
		return m, m.FormState.CommentForm.Init()

	case "delete", "d":
		// Delete the selected comment
		if len(m.CommentState.Items) > 0 && m.CommentState.Cursor < len(m.CommentState.Items) {
			ctx, cancel := m.DBContext()
			defer cancel()

			commentID := m.CommentState.Items[m.CommentState.Cursor].Comment.ID
			err := m.App.TaskService.DeleteComment(ctx, commentID)
			if err != nil {
				slog.Error("Error deleting comment", "error", err)
				m.NotificationState.Add(state.LevelError, "Failed to delete comment")
				return m, nil
			}

			// Reload comments from database
			taskID := m.FormState.EditingTaskID
			comments, err := m.App.TaskService.GetCommentsByTask(ctx, taskID)
			if err != nil {
				slog.Error("Error reloading comments", "error", err)
				m.NotificationState.Add(state.LevelError, "Failed to reload comments")
				return m, nil
			}

			// Update form state and comment state
			m.FormState.FormComments = comments
			m.CommentState.Items = convertToCommentItems(comments)

			// Adjust cursor if needed
			if m.CommentState.Cursor >= len(m.CommentState.Items) && len(m.CommentState.Items) > 0 {
				m.CommentState.Cursor = len(m.CommentState.Items) - 1
			}
			if len(m.CommentState.Items) == 0 {
				m.CommentState.Cursor = 0
			}

			m.NotificationState.Add(state.LevelInfo, "Comment deleted")
		}
		return m, nil
	}

	return m, nil
}

// convertToCommentItems converts a slice of Comment models to CommentItem for display
func convertToCommentItems(comments []*models.Comment) []state.CommentItem {
	items := make([]state.CommentItem, len(comments))
	for i, comment := range comments {
		items[i] = state.CommentItem{
			Comment: comment,
		}
	}
	return items
}
