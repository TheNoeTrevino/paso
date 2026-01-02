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
		if len(m.Forms.Comment.Items) > 0 {
			m.Forms.Comment.MoveCursorDown(len(m.Forms.Comment.Items) - 1)
		}
		return m, nil

	case "k", "up":
		// Move cursor up
		m.Forms.Comment.MoveCursorUp()
		return m, nil

	case "enter":
		// Open form to edit the selected comment
		if len(m.Forms.Comment.Items) > 0 && m.Forms.Comment.Cursor < len(m.Forms.Comment.Items) {
			comment := m.Forms.Comment.Items[m.Forms.Comment.Cursor].Comment
			m.Forms.Form.FormCommentMessage = comment.Message
			m.Forms.Form.EditingCommentID = comment.ID
			m.Forms.Form.CommentForm = huhforms.CreateCommentForm(&m.Forms.Form.FormCommentMessage, true).
				WithTheme(huhforms.CreatePasoTheme(m.Config.ColorScheme))
			m.Forms.Form.SnapshotCommentFormInitialValues()
			m.UIState.SetMode(state.CommentFormMode)
			return m, m.Forms.Form.CommentForm.Init()
		}
		return m, nil

	case "ctrl+n", "n":
		// Open form to create a new comment
		m.Forms.Form.FormCommentMessage = ""
		m.Forms.Form.EditingCommentID = 0
		m.Forms.Form.CommentForm = huhforms.CreateCommentForm(&m.Forms.Form.FormCommentMessage, false).
			WithTheme(huhforms.CreatePasoTheme(m.Config.ColorScheme))
		m.Forms.Form.SnapshotCommentFormInitialValues()
		m.UIState.SetMode(state.CommentFormMode)
		return m, m.Forms.Form.CommentForm.Init()

	case "delete", "d":
		// Delete the selected comment
		if len(m.Forms.Comment.Items) > 0 && m.Forms.Comment.Cursor < len(m.Forms.Comment.Items) {
			ctx, cancel := m.DBContext()
			defer cancel()

			commentID := m.Forms.Comment.Items[m.Forms.Comment.Cursor].Comment.ID
			err := m.App.TaskService.DeleteComment(ctx, commentID)
			if err != nil {
				slog.Error("failed to deleting comment", "error", err)
				m.UI.Notification.Add(state.LevelError, "Failed to delete comment")
				return m, nil
			}

			// Reload comments from database
			taskID := m.Forms.Form.EditingTaskID
			comments, err := m.App.TaskService.GetCommentsByTask(ctx, taskID)
			if err != nil {
				slog.Error("failed to reloading comments", "error", err)
				m.UI.Notification.Add(state.LevelError, "Failed to reload comments")
				return m, nil
			}

			// Update form state and comment state
			m.Forms.Form.FormComments = comments
			m.Forms.Comment.Items = convertToCommentItems(comments)

			// Adjust cursor if needed
			if m.Forms.Comment.Cursor >= len(m.Forms.Comment.Items) && len(m.Forms.Comment.Items) > 0 {
				m.Forms.Comment.Cursor = len(m.Forms.Comment.Items) - 1
			}
			if len(m.Forms.Comment.Items) == 0 {
				m.Forms.Comment.Cursor = 0
			}

			m.UI.Notification.Add(state.LevelInfo, "Comment deleted")
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
