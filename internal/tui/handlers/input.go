package handlers

import (
	"log/slog"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui/modelops"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// ============================================================================
// INPUT MODE HANDLERS
// ============================================================================

// HandleInputMode handles text input for column creation/editing.
func (w *Wrapper) HandleInputMode(msg tea.KeyMsg) (*Wrapper, tea.Cmd) {
	switch msg.String() {
	case "enter":
		return w.handleInputConfirm()
	case "esc":
		// Check for changes before closing
		shouldConfirm := false
		contextMsg := ""

		if w.UiState.Mode() == state.AddColumnMode {
			// AddColumnMode: confirm if user typed anything
			shouldConfirm = !w.InputState.IsEmpty()
			contextMsg = "Discard column?"
		} else if w.UiState.Mode() == state.EditColumnMode {
			// EditColumnMode: confirm if text changed from original
			shouldConfirm = w.InputState.HasInputChanges()
			contextMsg = "Discard changes?"
		}

		if shouldConfirm {
			w.UiState.SetDiscardContext(&state.DiscardContext{
				SourceMode: w.UiState.Mode(),
				Message:    contextMsg,
			})
			w.UiState.SetMode(state.DiscardConfirmMode)
			return w, nil
		}

		// No changes - immediate close
		return w.handleInputCancel()
	case "backspace", "ctrl+h":
		w.InputState.Backspace()
		return w, nil
	default:
		key := msg.String()
		if len(key) == 1 {
			w.InputState.AppendChar(rune(key[0]))
		}
		return w, nil
	}
}

// handleInputConfirm processes the input and creates/renames column.
func (w *Wrapper) handleInputConfirm() (*Wrapper, tea.Cmd) {
	if strings.TrimSpace(w.InputState.Buffer) == "" {
		w.UiState.SetMode(state.NormalMode)
		w.InputState.Clear()
		return w, nil
	}

	if w.UiState.Mode() == state.AddColumnMode {
		return w.createColumn()
	}
	return w.renameColumn()
}

// handleInputCancel cancels the input operation.
func (w *Wrapper) handleInputCancel() (*Wrapper, tea.Cmd) {
	w.UiState.SetMode(state.NormalMode)
	w.InputState.Clear()
	return w, nil
}

// createColumn creates a new column with the input buffer as name.
func (w *Wrapper) createColumn() (*Wrapper, tea.Cmd) {
	ops := modelops.New(w.Model)

	var afterColumnID *int
	if len(w.AppState.Columns()) > 0 {
		currentCol := ops.GetCurrentColumn()
		if currentCol != nil {
			afterColumnID = &currentCol.ID
		}
	}

	projectID := 0
	if project := ops.GetCurrentProject(); project != nil {
		projectID = project.ID
	}

	ctx, cancel := w.DbContext()
	defer cancel()
	column, err := w.Repo.CreateColumn(ctx, strings.TrimSpace(w.InputState.Buffer), projectID, afterColumnID)
	if err != nil {
		slog.Error("Error creating column", "error", err)
		w.NotificationState.Add(state.LevelError, "Failed to create column")
	} else {
		columns, err := w.Repo.GetColumnsByProject(ctx, projectID)
		if err != nil {
			slog.Error("Error reloading columns", "error", err)
			w.NotificationState.Add(state.LevelError, "Failed to reload columns")
		}
		w.AppState.SetColumns(columns)
		w.AppState.Tasks()[column.ID] = []*models.TaskSummary{}
		if afterColumnID != nil {
			w.UiState.SetSelectedColumn(w.UiState.SelectedColumn() + 1)
		}
	}

	w.UiState.SetMode(state.NormalMode)
	w.InputState.Clear()
	return w, nil
}

// renameColumn renames the current column with the input buffer.
func (w *Wrapper) renameColumn() (*Wrapper, tea.Cmd) {
	ops := modelops.New(w.Model)
	column := ops.GetCurrentColumn()
	if column != nil {
		ctx, cancel := w.DbContext()
		defer cancel()
		err := w.Repo.UpdateColumnName(ctx, column.ID, strings.TrimSpace(w.InputState.Buffer))
		if err != nil {
			slog.Error("Error updating column", "error", err)
			w.NotificationState.Add(state.LevelError, "Failed to rename column")
		} else {
			column.Name = strings.TrimSpace(w.InputState.Buffer)
		}
	}

	w.UiState.SetMode(state.NormalMode)
	w.InputState.Clear()
	return w, nil
}
