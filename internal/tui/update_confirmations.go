package tui

import (
	"log/slog"

	tea "charm.land/bubbletea/v2"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// ============================================================================
// CONFIRMATION HANDLERS (Inlined from deleted confirmation.go)
// ============================================================================

// handleDeleteConfirm handles task or comment deletion confirmation.
// Inlined from confirmation.go (deleted to reduce duplication)
func (m Model) handleDeleteConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		// Check if we're deleting a comment or task
		if m.FormState.EditingCommentID != 0 {
			return m.confirmDeleteComment()
		}
		return m.confirmDeleteTask()
	case "n", "N", "esc":
		// Return to appropriate mode
		if m.FormState.EditingCommentID != 0 {
			m.FormState.EditingCommentID = 0
			m.UIState.SetMode(state.CommentsViewMode)
		} else {
			m.UIState.SetMode(state.NormalMode)
		}
		return m, nil
	}
	return m, nil
}

// confirmDeleteTask performs the actual task deletion.
// Inlined from confirmation.go (deleted to reduce duplication)
func (m Model) confirmDeleteTask() (tea.Model, tea.Cmd) {
	task := m.getCurrentTask()
	if task != nil {
		ctx, cancel := m.DBContext()
		defer cancel()
		err := m.App.TaskService.DeleteTask(ctx, task.ID)
		if err != nil {
			slog.Error("Error deleting task", "error", err)
			m.NotificationState.Add(state.LevelError, "Failed to delete task")
		} else {
			m.removeCurrentTask()
		}
	}
	m.UIState.SetMode(state.NormalMode)
	return m, nil
}

// confirmDeleteComment performs the actual comment deletion.
func (m Model) confirmDeleteComment() (tea.Model, tea.Cmd) {
	commentID := m.FormState.EditingCommentID
	taskID := m.CommentState.TaskID

	if commentID != 0 && taskID != 0 {
		ctx, cancel := m.DBContext()
		defer cancel()
		err := m.App.TaskService.DeleteComment(ctx, commentID)
		if err != nil {
			slog.Error("Error deleting comment", "error", err)
			m.NotificationState.Add(state.LevelError, "Failed to delete comment")
		} else {
			// Remove from local state
			m.CommentState.DeleteSelected()
			m.NotificationState.Add(state.LevelInfo, "Comment deleted")

			// Refresh comments in form state
			comments, err := m.App.TaskService.GetCommentsByTask(ctx, taskID)
			if err == nil {
				m.FormState.FormComments = comments
				m.CommentState.SetComments(comments)
			}
		}
	}

	m.FormState.EditingCommentID = 0
	m.UIState.SetMode(state.CommentsViewMode)
	return m, nil
}

// handleDiscardConfirm handles discard confirmation for forms and inputs.
// Inlined from confirmation.go (deleted to reduce duplication)
func (m Model) handleDiscardConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	ctx := m.UIState.DiscardContext()
	if ctx == nil {
		// Safety: if context is missing, return to normal mode
		m.UIState.SetMode(state.NormalMode)
		return m, nil
	}

	switch msg.String() {
	case "y", "Y":
		// User confirmed discard - clear form/input and return to normal mode
		return m.confirmDiscard()

	case "n", "N", "esc":
		// User cancelled - return to source mode without clearing
		m.UIState.SetMode(ctx.SourceMode)
		m.UIState.ClearDiscardContext()
		return m, nil
	}

	return m, nil
}

// confirmDiscard performs the actual discard operation based on context.
// Inlined from confirmation.go (deleted to reduce duplication)
func (m Model) confirmDiscard() (tea.Model, tea.Cmd) {
	ctx := m.UIState.DiscardContext()
	if ctx == nil {
		m.UIState.SetMode(state.NormalMode)
		return m, nil
	}

	// Clear the appropriate form/input based on source mode
	switch ctx.SourceMode {
	case state.TicketFormMode:
		m.FormState.ClearTaskForm()

	case state.ProjectFormMode:
		m.FormState.ClearProjectForm()

	case state.AddColumnFormMode, state.EditColumnFormMode:
		m.FormState.ClearColumnForm()

	case state.CommentFormMode:
		m.FormState.ClearCommentForm()
		// Return to comment list instead of normal mode
		m.UIState.SetMode(state.CommentEditMode)
		m.UIState.ClearDiscardContext()
		return m, tea.ClearScreen
	}

	// Always return to normal mode after discard
	m.UIState.SetMode(state.NormalMode)
	m.UIState.ClearDiscardContext()

	return m, tea.ClearScreen
}

// handleDeleteColumnConfirm handles column deletion confirmation.
// Inlined from confirmation.go (deleted to reduce duplication)
func (m Model) handleDeleteColumnConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		return m.confirmDeleteColumn()
	case "n", "N", "esc":
		m.UIState.SetMode(state.NormalMode)
		return m, nil
	}
	return m, nil
}

// confirmDeleteColumn performs the actual column deletion.
// Inlined from confirmation.go (deleted to reduce duplication)
func (m Model) confirmDeleteColumn() (tea.Model, tea.Cmd) {
	column := m.getCurrentColumn()
	if column != nil {
		ctx, cancel := m.DBContext()
		defer cancel()
		err := m.App.ColumnService.DeleteColumn(ctx, column.ID)
		if err != nil {
			slog.Error("Error deleting column", "error", err)
			m.NotificationState.Add(state.LevelError, "Failed to delete column")
		} else {
			delete(m.AppState.Tasks(), column.ID)
			m.removeCurrentColumn()
		}
	}
	m.UIState.SetMode(state.NormalMode)
	return m, nil
}
