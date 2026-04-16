package tui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui/huhforms"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// handleCommentsViewInput processes input in comments view mode
func (m Model) handleCommentsViewInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	km := m.Config.KeyMappings

	switch key {
	case km.Navigation.MoveUp, "up":
		return m.handleCommentsViewUp()
	case km.Navigation.MoveDown, "down":
		return m.handleCommentsViewDown()
	case "enter", km.Tasks.EditTask:
		return m.handleCommentsViewEdit()
	case km.Tasks.AddTask:
		return m.handleCommentsViewAdd()
	case km.Tasks.DeleteTask:
		return m.handleCommentsViewDelete()
	case "esc", km.General.Quit:
		return m.handleCommentsViewClose()
	}

	return m, nil
}

// handleCommentsViewUp moves cursor up in comments list
func (m Model) handleCommentsViewUp() (tea.Model, tea.Cmd) {
	moved := m.Forms.Comment.MoveCursorUp()
	if moved {
		availableHeight := commentsViewAvailableHeight(m.UIState.Height)
		_, _, maxVisible := calculateVisibleCommentRange(
			m.Forms.Comment.ScrollOffset,
			len(m.Forms.Comment.Items),
			availableHeight,
		)
		m.Forms.Comment.EnsureCommentVisible(maxVisible)
	}
	return m, nil
}

// handleCommentsViewDown moves cursor down in comments list
func (m Model) handleCommentsViewDown() (tea.Model, tea.Cmd) {
	maxIdx := len(m.Forms.Comment.Items) - 1
	moved := m.Forms.Comment.MoveCursorDown(maxIdx)
	if moved {
		availableHeight := commentsViewAvailableHeight(m.UIState.Height)
		_, _, maxVisible := calculateVisibleCommentRange(
			m.Forms.Comment.ScrollOffset,
			len(m.Forms.Comment.Items),
			availableHeight,
		)
		m.Forms.Comment.EnsureCommentVisible(maxVisible)
	}
	return m, nil
}

// handleCommentsViewEdit opens the comment form to edit the selected activity
func (m Model) handleCommentsViewEdit() (tea.Model, tea.Cmd) {
	selectedActivity := m.Forms.Comment.GetSelectedActivity()
	if selectedActivity == nil {
		return m, m.addNotification(state.LevelError, "No item selected")
	}

	// Events are read-only and cannot be edited
	if selectedActivity.Type == models.ActivityTypeEvent {
		return m, m.addNotification(state.LevelWarning, "Events cannot be edited")
	}

	// Set up form state for editing
	m.Forms.Form.FormCommentMessage = selectedActivity.Content
	m.Forms.Form.EditingCommentID = selectedActivity.ID
	m.Forms.Form.CommentFormReturnMode = state.CommentsViewMode

	// Create comment form
	isEdit := true
	m.Forms.Form.CommentForm = huhforms.CreateCommentForm(
		&m.Forms.Form.FormCommentMessage,
		isEdit,
	).WithTheme(huhforms.CreatePasoTheme(m.Config.ColorScheme))
	m.Forms.Form.SnapshotCommentFormInitialValues()

	// Switch to comment form mode
	m.UIState.Mode = state.CommentFormMode

	return m, m.Forms.Form.CommentForm.Init()
}

// handleCommentsViewAdd opens the comment form to create a new comment
func (m Model) handleCommentsViewAdd() (tea.Model, tea.Cmd) {
	// Set up form state for creating
	m.Forms.Form.FormCommentMessage = ""
	m.Forms.Form.EditingCommentID = 0
	m.Forms.Form.CommentFormReturnMode = state.CommentsViewMode

	// Create comment form
	isEdit := false
	m.Forms.Form.CommentForm = huhforms.CreateCommentForm(
		&m.Forms.Form.FormCommentMessage,
		isEdit,
	).WithTheme(huhforms.CreatePasoTheme(m.Config.ColorScheme))
	m.Forms.Form.SnapshotCommentFormInitialValues()

	// Switch to comment form mode
	m.UIState.Mode = state.CommentFormMode

	return m, m.Forms.Form.CommentForm.Init()
}

// handleCommentsViewDelete shows confirmation dialog for deleting the selected activity
func (m Model) handleCommentsViewDelete() (tea.Model, tea.Cmd) {
	selectedActivity := m.Forms.Comment.GetSelectedActivity()
	if selectedActivity == nil {
		return m, m.addNotification(state.LevelError, "No item selected")
	}

	// Events are read-only and cannot be deleted
	if selectedActivity.Type == models.ActivityTypeEvent {
		return m, m.addNotification(state.LevelWarning, "Events cannot be deleted")
	}

	// Store comment ID for deletion
	m.Forms.Form.EditingCommentID = selectedActivity.ID

	// Show delete confirmation
	// We'll use the same DeleteConfirmMode and handle comment deletion there
	m.UIState.Mode = state.DeleteConfirmMode

	return m, nil
}

// handleCommentsViewClose closes the comments view and returns to task form
func (m Model) handleCommentsViewClose() (tea.Model, tea.Cmd) {
	// Return to task form mode
	m.UIState.Mode = state.TicketFormMode
	return m, nil
}
