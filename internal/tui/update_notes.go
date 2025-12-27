package tui

import (
	"log/slog"

	tea "charm.land/bubbletea/v2"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui/huhforms"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// updateNotes handles keyboard input in the note editing mode.
// This function processes navigation (up/down), opening forms for editing, and note deletion.
func (m Model) updateNotes(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Return to ticket form mode
		m.UiState.SetMode(state.TicketFormMode)
		return m, nil

	case "j", "down":
		// Move cursor down
		if len(m.NoteState.Items) > 0 {
			m.NoteState.MoveCursorDown(len(m.NoteState.Items) - 1)
		}
		return m, nil

	case "k", "up":
		// Move cursor up
		m.NoteState.MoveCursorUp()
		return m, nil

	case "enter":
		// Open form to edit the selected note
		if len(m.NoteState.Items) > 0 && m.NoteState.Cursor < len(m.NoteState.Items) {
			comment := m.NoteState.Items[m.NoteState.Cursor].Comment
			m.FormState.FormCommentMessage = comment.Message
			m.FormState.EditingCommentID = comment.ID
			m.FormState.CommentForm = huhforms.CreateCommentForm(&m.FormState.FormCommentMessage, true).
				WithTheme(huhforms.CreatePasoTheme(m.Config.ColorScheme))
			m.FormState.SnapshotCommentFormInitialValues()
			m.UiState.SetMode(state.NoteFormMode)
			return m, m.FormState.CommentForm.Init()
		}
		return m, nil

	case "ctrl+n", "n":
		// Open form to create a new note
		m.FormState.FormCommentMessage = ""
		m.FormState.EditingCommentID = 0
		m.FormState.CommentForm = huhforms.CreateCommentForm(&m.FormState.FormCommentMessage, false).
			WithTheme(huhforms.CreatePasoTheme(m.Config.ColorScheme))
		m.FormState.SnapshotCommentFormInitialValues()
		m.UiState.SetMode(state.NoteFormMode)
		return m, m.FormState.CommentForm.Init()

	case "delete", "d":
		// Delete the selected note
		if len(m.NoteState.Items) > 0 && m.NoteState.Cursor < len(m.NoteState.Items) {
			ctx, cancel := m.DbContext()
			defer cancel()

			commentID := m.NoteState.Items[m.NoteState.Cursor].Comment.ID
			err := m.App.TaskService.DeleteComment(ctx, commentID)
			if err != nil {
				slog.Error("Error deleting note", "error", err)
				m.NotificationState.Add(state.LevelError, "Failed to delete note")
				return m, nil
			}

			// Reload comments from database
			taskID := m.FormState.EditingTaskID
			comments, err := m.App.TaskService.GetCommentsByTask(ctx, taskID)
			if err != nil {
				slog.Error("Error reloading notes", "error", err)
				m.NotificationState.Add(state.LevelError, "Failed to reload notes")
				return m, nil
			}

			// Update form state and note state
			m.FormState.FormComments = comments
			m.NoteState.Items = convertToNoteItems(comments)

			// Adjust cursor if needed
			if m.NoteState.Cursor >= len(m.NoteState.Items) && len(m.NoteState.Items) > 0 {
				m.NoteState.Cursor = len(m.NoteState.Items) - 1
			}
			if len(m.NoteState.Items) == 0 {
				m.NoteState.Cursor = 0
			}

			m.NotificationState.Add(state.LevelInfo, "Note deleted")
		}
		return m, nil
	}

	return m, nil
}

// convertToNoteItems converts a slice of Comment models to NoteItem for display
func convertToNoteItems(comments []*models.Comment) []state.NoteItem {
	items := make([]state.NoteItem, len(comments))
	for i, comment := range comments {
		items[i] = state.NoteItem{
			Comment: comment,
		}
	}
	return items
}
