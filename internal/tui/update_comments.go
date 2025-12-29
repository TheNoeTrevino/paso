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
	m.NoteState.MoveCursorUp()
	return m, nil
}

// handleCommentsViewDown moves cursor down in comments list
func (m Model) handleCommentsViewDown() (tea.Model, tea.Cmd) {
	maxIdx := len(m.NoteState.Items) - 1
	m.NoteState.MoveCursorDown(maxIdx)
	return m, nil
}

// handleCommentsViewEdit opens the comment form to edit the selected comment
func (m Model) handleCommentsViewEdit() (tea.Model, tea.Cmd) {
	selectedComment := m.NoteState.GetSelectedComment()
	if selectedComment == nil {
		m.NotificationState.Add(state.LevelError, "No comment selected")
		return m, nil
	}

	// Set up form state for editing
	m.FormState.FormCommentMessage = selectedComment.Message
	m.FormState.EditingCommentID = selectedComment.ID
	m.FormState.CommentFormReturnMode = state.CommentsViewMode

	// Create comment form
	isEdit := true
	m.FormState.CommentForm = huhforms.CreateCommentForm(
		&m.FormState.FormCommentMessage,
		isEdit,
	).WithTheme(huhforms.CreatePasoTheme(m.Config.ColorScheme))
	m.FormState.SnapshotCommentFormInitialValues()

	// Switch to note form mode
	m.UiState.SetMode(state.NoteFormMode)

	return m, m.FormState.CommentForm.Init()
}

// handleCommentsViewAdd opens the comment form to create a new comment
func (m Model) handleCommentsViewAdd() (tea.Model, tea.Cmd) {
	// Set up form state for creating
	m.FormState.FormCommentMessage = ""
	m.FormState.EditingCommentID = 0
	m.FormState.CommentFormReturnMode = state.CommentsViewMode

	// Create comment form
	isEdit := false
	m.FormState.CommentForm = huhforms.CreateCommentForm(
		&m.FormState.FormCommentMessage,
		isEdit,
	).WithTheme(huhforms.CreatePasoTheme(m.Config.ColorScheme))
	m.FormState.SnapshotCommentFormInitialValues()

	// Switch to note form mode
	m.UiState.SetMode(state.NoteFormMode)

	return m, m.FormState.CommentForm.Init()
}

// handleCommentsViewDelete shows confirmation dialog for deleting the selected comment
func (m Model) handleCommentsViewDelete() (tea.Model, tea.Cmd) {
	selectedComment := m.NoteState.GetSelectedComment()
	if selectedComment == nil {
		m.NotificationState.Add(state.LevelError, "No comment selected")
		return m, nil
	}

	// Store comment ID for deletion
	m.FormState.EditingCommentID = selectedComment.ID

	// Show delete confirmation
	// We'll use the same DeleteConfirmMode and handle comment deletion there
	m.UiState.SetMode(state.DeleteConfirmMode)

	return m, nil
}

// handleCommentsViewClose closes the comments view and returns to ticket form
func (m Model) handleCommentsViewClose() (tea.Model, tea.Cmd) {
	// Return to ticket form mode
	m.UiState.SetMode(state.TicketFormMode)
	return m, nil
}
