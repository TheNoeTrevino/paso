package tui

import (
	"log/slog"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// ============================================================================
// INPUT MODE HANDLERS
// ============================================================================

// handleInputMode handles text input for column creation/editing.
func (m Model) handleInputMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		return m.handleInputConfirm()
	case "esc":
		// Check for changes before closing
		shouldConfirm := false
		contextMsg := ""

		if m.uiState.Mode() == state.AddColumnMode {
			// AddColumnMode: confirm if user typed anything
			shouldConfirm = !m.inputState.IsEmpty()
			contextMsg = "Discard column?"
		} else if m.uiState.Mode() == state.EditColumnMode {
			// EditColumnMode: confirm if text changed from original
			shouldConfirm = m.inputState.HasInputChanges()
			contextMsg = "Discard changes?"
		}

		if shouldConfirm {
			m.uiState.SetDiscardContext(&state.DiscardContext{
				SourceMode: m.uiState.Mode(),
				Message:    contextMsg,
			})
			m.uiState.SetMode(state.DiscardConfirmMode)
			return m, nil
		}

		// No changes - immediate close
		return m.handleInputCancel()
	case "backspace", "ctrl+h":
		m.inputState.Backspace()
		return m, nil
	default:
		key := msg.String()
		if len(key) == 1 {
			m.inputState.AppendChar(rune(key[0]))
		}
		return m, nil
	}
}

// handleInputConfirm processes the input and creates/renames column.
func (m Model) handleInputConfirm() (tea.Model, tea.Cmd) {
	if strings.TrimSpace(m.inputState.Buffer) == "" {
		m.uiState.SetMode(state.NormalMode)
		m.inputState.Clear()
		return m, nil
	}

	if m.uiState.Mode() == state.AddColumnMode {
		return m.createColumn()
	}
	return m.renameColumn()
}

// handleInputCancel cancels the input operation.
func (m Model) handleInputCancel() (tea.Model, tea.Cmd) {
	m.uiState.SetMode(state.NormalMode)
	m.inputState.Clear()
	return m, nil
}

// createColumn creates a new column with the input buffer as name.
func (m Model) createColumn() (tea.Model, tea.Cmd) {
	var afterColumnID *int
	if len(m.appState.Columns()) > 0 {
		currentCol := m.getCurrentColumn()
		if currentCol != nil {
			afterColumnID = &currentCol.ID
		}
	}

	projectID := 0
	if project := m.getCurrentProject(); project != nil {
		projectID = project.ID
	}

	ctx, cancel := m.dbContext()
	defer cancel()
	column, err := m.repo.CreateColumn(ctx, strings.TrimSpace(m.inputState.Buffer), projectID, afterColumnID)
	if err != nil {
		slog.Error("Error creating column", "error", err)
		m.notificationState.Add(state.LevelError, "Failed to create column")
	} else {
		columns, err := m.repo.GetColumnsByProject(ctx, projectID)
		if err != nil {
			slog.Error("Error reloading columns", "error", err)
			m.notificationState.Add(state.LevelError, "Failed to reload columns")
		}
		m.appState.SetColumns(columns)
		m.appState.Tasks()[column.ID] = []*models.TaskSummary{}
		if afterColumnID != nil {
			m.uiState.SetSelectedColumn(m.uiState.SelectedColumn() + 1)
		}
	}

	m.uiState.SetMode(state.NormalMode)
	m.inputState.Clear()
	return m, nil
}

// renameColumn renames the current column with the input buffer.
func (m Model) renameColumn() (tea.Model, tea.Cmd) {
	column := m.getCurrentColumn()
	if column != nil {
		ctx, cancel := m.dbContext()
		defer cancel()
		err := m.repo.UpdateColumnName(ctx, column.ID, strings.TrimSpace(m.inputState.Buffer))
		if err != nil {
			slog.Error("Error updating column", "error", err)
			m.notificationState.Add(state.LevelError, "Failed to rename column")
		} else {
			column.Name = strings.TrimSpace(m.inputState.Buffer)
		}
	}

	m.uiState.SetMode(state.NormalMode)
	m.inputState.Clear()
	return m, nil
}
