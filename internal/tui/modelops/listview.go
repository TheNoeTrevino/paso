package modelops

import (
	"sort"
	"strings"

	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui"
	"github.com/thenoetrevino/paso/internal/tui/renderers"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// BuildListViewRows creates a list of all tasks across all columns for list view.
// Returns rows sorted according to current sort settings.
func BuildListViewRows(m *tui.Model) []renderers.ListViewRow {
	var rows []renderers.ListViewRow
	for _, col := range m.AppState.Columns() {
		tasks := m.AppState.Tasks()[col.ID]
		for _, task := range tasks {
			rows = append(rows, renderers.ListViewRow{
				Task:       task,
				ColumnName: col.Name,
				ColumnID:   col.ID,
			})
		}
	}

	// Apply sorting
	SortListViewRows(m, rows)
	return rows
}

// SortListViewRows sorts the rows based on current sort settings.
// Modifies rows in place.
func SortListViewRows(m *tui.Model, rows []renderers.ListViewRow) {
	if m.ListViewState.SortField() == state.SortNone {
		return
	}

	sort.Slice(rows, func(i, j int) bool {
		var cmp int
		switch m.ListViewState.SortField() {
		case state.SortByTitle:
			cmp = strings.Compare(rows[i].Task.Title, rows[j].Task.Title)
		case state.SortByStatus:
			cmp = strings.Compare(rows[i].ColumnName, rows[j].ColumnName)
		default:
			return false
		}

		if m.ListViewState.SortOrder() == state.SortDesc {
			cmp = -cmp
		}
		return cmp < 0
	})
}

// SyncKanbanToListSelection maps the current kanban selection to a list row index.
// This should be called when switching from kanban to list view.
func SyncKanbanToListSelection(m *tui.Model) {
	rows := BuildListViewRows(m)
	if len(rows) == 0 {
		m.ListViewState.SetSelectedRow(0)
		return
	}

	// Find the task that matches the current kanban selection
	currentTask := GetCurrentTask(m)
	if currentTask == nil {
		m.ListViewState.SetSelectedRow(0)
		return
	}

	for i, row := range rows {
		if row.Task.ID == currentTask.ID {
			m.ListViewState.SetSelectedRow(i)
			return
		}
	}
	m.ListViewState.SetSelectedRow(0)
}

// SyncListToKanbanSelection maps the current list row to kanban column/task selection.
// This should be called when switching from list to kanban view.
func SyncListToKanbanSelection(m *tui.Model) {
	rows := BuildListViewRows(m)
	if len(rows) == 0 {
		return
	}

	selectedRow := m.ListViewState.SelectedRow()
	if selectedRow >= len(rows) {
		selectedRow = len(rows) - 1
	}
	if selectedRow < 0 {
		return
	}

	selectedTask := rows[selectedRow].Task

	// Find the column and task position in kanban view
	for colIdx, col := range m.AppState.Columns() {
		tasks := m.AppState.Tasks()[col.ID]
		for taskIdx, task := range tasks {
			if task.ID == selectedTask.ID {
				m.UiState.SetSelectedColumn(colIdx)
				m.UiState.SetSelectedTask(taskIdx)
				m.UiState.EnsureSelectionVisible(colIdx)
				return
			}
		}
	}
}

// GetTaskFromListRow returns the task at the given list row index.
// Returns nil if the index is out of bounds or no tasks exist.
func GetTaskFromListRow(m *tui.Model, rowIdx int) *models.TaskSummary {
	rows := BuildListViewRows(m)
	if rowIdx < 0 || rowIdx >= len(rows) {
		return nil
	}
	return rows[rowIdx].Task
}

// GetSelectedListTask returns the currently selected task in list view.
// This is a convenience method that uses GetTaskFromListRow with the current selection.
func GetSelectedListTask(m *tui.Model) *models.TaskSummary {
	return GetTaskFromListRow(m, m.ListViewState.SelectedRow())
}
