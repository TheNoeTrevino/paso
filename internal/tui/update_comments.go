package tui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/thenoetrevino/paso/internal/tui/huhforms"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// handleCommentsViewInput processes input in comments view mode
func (m Model) handleCommentsViewInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "up", "k":
		return m.handleCommentsViewUp()
	case "down", "j":
		return m.handleCommentsViewDown()
	case "enter", "e":
		return m.handleCommentsViewEdit()
	case "a":
		return m.handleCommentsViewAdd()
	case "d":
		return m.handleCommentsViewDelete()
	case "esc", "q":
		return m.handleCommentsViewClose()
	}

	return m, nil
}

// handleCommentsViewUp moves cursor up in comments list
func (m Model) handleCommentsViewUp() (tea.Model, tea.Cmd) {
	moved := m.Forms.Comment.MoveCursorUp()
	if moved {
		// Calculate max visible for auto-scroll
		// Must match the height calculation in renderCommentsViewContent
		layerHeight := m.UIState.Height() * 8 / 10
		availableHeight := layerHeight - 4 // Reserve for title + help
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
		// Calculate max visible for auto-scroll
		// Must match the height calculation in renderCommentsViewContent
		layerHeight := m.UIState.Height() * 8 / 10
		availableHeight := layerHeight - 4 // Reserve for title + help
		_, _, maxVisible := calculateVisibleCommentRange(
			m.Forms.Comment.ScrollOffset,
			len(m.Forms.Comment.Items),
			availableHeight,
		)
		m.Forms.Comment.EnsureCommentVisible(maxVisible)
	}
	return m, nil
}

// handleCommentsViewEdit opens the comment form to edit the selected comment
func (m Model) handleCommentsViewEdit() (tea.Model, tea.Cmd) {
	selectedComment := m.Forms.Comment.GetSelectedComment()
	if selectedComment == nil {
		m.UI.Notification.Add(state.LevelError, "No comment selected")
		return m, nil
	}

	// Set up form state for editing
	m.Forms.Form.FormCommentMessage = selectedComment.Message
	m.Forms.Form.EditingCommentID = selectedComment.ID
	m.Forms.Form.CommentFormReturnMode = state.CommentsViewMode

	// Create comment form
	isEdit := true
	m.Forms.Form.CommentForm = huhforms.CreateCommentForm(
		&m.Forms.Form.FormCommentMessage,
		isEdit,
	).WithTheme(huhforms.CreatePasoTheme(m.Config.ColorScheme))
	m.Forms.Form.SnapshotCommentFormInitialValues()

	// Switch to comment form mode
	m.UIState.SetMode(state.CommentFormMode)

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
	m.UIState.SetMode(state.CommentFormMode)

	return m, m.Forms.Form.CommentForm.Init()
}

// handleCommentsViewDelete shows confirmation dialog for deleting the selected comment
func (m Model) handleCommentsViewDelete() (tea.Model, tea.Cmd) {
	selectedComment := m.Forms.Comment.GetSelectedComment()
	if selectedComment == nil {
		m.UI.Notification.Add(state.LevelError, "No comment selected")
		return m, nil
	}

	// Store comment ID for deletion
	m.Forms.Form.EditingCommentID = selectedComment.ID

	// Show delete confirmation
	// We'll use the same DeleteConfirmMode and handle comment deletion there
	m.UIState.SetMode(state.DeleteConfirmMode)

	return m, nil
}

// handleCommentsViewClose closes the comments view and returns to task form
func (m Model) handleCommentsViewClose() (tea.Model, tea.Cmd) {
	// Return to task form mode
	m.UIState.SetMode(state.TicketFormMode)
	return m, nil
}
