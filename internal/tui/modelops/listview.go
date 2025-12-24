package modelops

import (
	"sort"
	"strings"

	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui/renderers"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// BuildListViewRows creates a list of all tasks across all columns for list view.
// Returns rows sorted according to current sort settings.
func (w *Wrapper) BuildListViewRows() []renderers.ListViewRow {
	var rows []renderers.ListViewRow
	for _, col := range w.AppState.Columns() {
		tasks := w.AppState.Tasks()[col.ID]
		for _, task := range tasks {
			rows = append(rows, renderers.ListViewRow{
				Task:       task,
				ColumnName: col.Name,
				ColumnID:   col.ID,
			})
		}
	}

	// Apply sorting
	w.SortListViewRows(rows)
	return rows
}

// SortListViewRows sorts the rows based on current sort settings.
// Modifies rows in place.
func (w *Wrapper) SortListViewRows(rows []renderers.ListViewRow) {
	if w.ListViewState.SortField() == state.SortNone {
		return
	}

	sort.Slice(rows, func(i, j int) bool {
		var cmp int
		switch w.ListViewState.SortField() {
		case state.SortByTitle:
			cmp = strings.Compare(rows[i].Task.Title, rows[j].Task.Title)
		case state.SortByStatus:
			cmp = strings.Compare(rows[i].ColumnName, rows[j].ColumnName)
		default:
			return false
		}

		if w.ListViewState.SortOrder() == state.SortDesc {
			cmp = -cmp
		}
		return cmp < 0
	})
}

// SyncKanbanToListSelection maps the current kanban selection to a list row index.
// This should be called when switching from kanban to list view.
func (w *Wrapper) SyncKanbanToListSelection() {
	rows := w.BuildListViewRows()
	if len(rows) == 0 {
		w.ListViewState.SetSelectedRow(0)
		return
	}

	// Find the task that matches the current kanban selection
	currentTask := w.GetCurrentTask()
	if currentTask == nil {
		w.ListViewState.SetSelectedRow(0)
		return
	}

	for i, row := range rows {
		if row.Task.ID == currentTask.ID {
			w.ListViewState.SetSelectedRow(i)
			return
		}
	}
	w.ListViewState.SetSelectedRow(0)
}

// SyncListToKanbanSelection maps the current list row to kanban column/task selection.
// This should be called when switching from list to kanban view.
func (w *Wrapper) SyncListToKanbanSelection() {
	rows := w.BuildListViewRows()
	if len(rows) == 0 {
		return
	}

	selectedRow := w.ListViewState.SelectedRow()
	if selectedRow >= len(rows) {
		selectedRow = len(rows) - 1
	}
	if selectedRow < 0 {
		return
	}

	selectedTask := rows[selectedRow].Task

	// Find the column and task position in kanban view
	for colIdx, col := range w.AppState.Columns() {
		tasks := w.AppState.Tasks()[col.ID]
		for taskIdx, task := range tasks {
			if task.ID == selectedTask.ID {
				w.UiState.SetSelectedColumn(colIdx)
				w.UiState.SetSelectedTask(taskIdx)
				w.UiState.EnsureSelectionVisible(colIdx)
				return
			}
		}
	}
}

// GetTaskFromListRow returns the task at the given list row index.
// Returns nil if the index is out of bounds or no tasks exist.
func (w *Wrapper) GetTaskFromListRow(rowIdx int) *models.TaskSummary {
	rows := w.BuildListViewRows()
	if rowIdx < 0 || rowIdx >= len(rows) {
		return nil
	}
	return rows[rowIdx].Task
}

// GetSelectedListTask returns the currently selected task in list view.
// This is a convenience method that uses GetTaskFromListRow with the current selection.
func (w *Wrapper) GetSelectedListTask() *models.TaskSummary {
	return w.GetTaskFromListRow(w.ListViewState.SelectedRow())
}
