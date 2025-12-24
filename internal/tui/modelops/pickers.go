package modelops

import (
	"log/slog"

	"github.com/thenoetrevino/paso/internal/tui/renderers"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// InitParentPickerForForm initializes the parent picker for use in TicketFormMode.
// In edit mode: loads existing parent relationships from FormState.
// In create mode: starts with empty selection (relationships applied after task creation).
//
// Returns false if there's no current project.
func (w *Wrapper) InitParentPickerForForm() bool {
	project := w.GetCurrentProject()
	if project == nil {
		return false
	}

	// Get all task references for the entire project
	ctx, cancel := w.DbContext()
	defer cancel()
	allTasks, err := w.Repo.GetTaskReferencesForProject(ctx, project.ID)
	if err != nil {
		slog.Error("Error loading project tasks", "error", err)
		return false
	}

	// Build map of currently selected parent task IDs from form state
	parentTaskMap := make(map[int]bool)
	for _, parentID := range w.FormState.FormParentIDs {
		parentTaskMap[parentID] = true
	}

	// Build picker items from all tasks, excluding current task (if editing)
	items := make([]state.TaskPickerItem, 0, len(allTasks))
	for _, task := range allTasks {
		// In edit mode, exclude the task being edited
		if w.FormState.EditingTaskID != 0 && task.ID == w.FormState.EditingTaskID {
			continue
		}
		items = append(items, state.TaskPickerItem{
			TaskRef:  task,
			Selected: parentTaskMap[task.ID],
		})
	}

	// Initialize ParentPickerState
	w.ParentPickerState.Items = items
	w.ParentPickerState.TaskID = w.FormState.EditingTaskID // 0 for create mode
	w.ParentPickerState.Cursor = 0
	w.ParentPickerState.Filter = ""
	w.ParentPickerState.PickerType = "parent"
	w.ParentPickerState.ReturnMode = state.TicketFormMode

	return true
}

// InitChildPickerForForm initializes the child picker for use in TicketFormMode.
// In edit mode: loads existing child relationships from FormState.
// In create mode: starts with empty selection (relationships applied after task creation).
//
// Returns false if there's no current project.
func (w *Wrapper) InitChildPickerForForm() bool {
	project := w.GetCurrentProject()
	if project == nil {
		return false
	}

	// Get all task references for the entire project
	ctx, cancel := w.DbContext()
	defer cancel()
	allTasks, err := w.Repo.GetTaskReferencesForProject(ctx, project.ID)
	if err != nil {
		slog.Error("Error loading project tasks", "error", err)
		return false
	}

	// Build map of currently selected child task IDs from form state
	childTaskMap := make(map[int]bool)
	for _, childID := range w.FormState.FormChildIDs {
		childTaskMap[childID] = true
	}

	// Build picker items from all tasks, excluding current task (if editing)
	items := make([]state.TaskPickerItem, 0, len(allTasks))
	for _, task := range allTasks {
		// In edit mode, exclude the task being edited
		if w.FormState.EditingTaskID != 0 && task.ID == w.FormState.EditingTaskID {
			continue
		}
		items = append(items, state.TaskPickerItem{
			TaskRef:  task,
			Selected: childTaskMap[task.ID],
		})
	}

	// Initialize ChildPickerState
	w.ChildPickerState.Items = items
	w.ChildPickerState.TaskID = w.FormState.EditingTaskID // 0 for create mode
	w.ChildPickerState.Cursor = 0
	w.ChildPickerState.Filter = ""
	w.ChildPickerState.PickerType = "child"
	w.ChildPickerState.ReturnMode = state.TicketFormMode

	return true
}

// InitLabelPickerForForm initializes the label picker for use in TicketFormMode.
// In edit mode: loads existing label selections from FormState.
// In create mode: starts with empty selection (labels applied on form submission).
//
// Returns false if there's no current project.
func (w *Wrapper) InitLabelPickerForForm() bool {
	project := w.GetCurrentProject()
	if project == nil {
		return false
	}

	// Build map of currently selected label IDs from form state
	labelIDMap := make(map[int]bool)
	for _, labelID := range w.FormState.FormLabelIDs {
		labelIDMap[labelID] = true
	}

	// Build picker items from all available labels
	var items []state.LabelPickerItem
	for _, label := range w.AppState.Labels() {
		items = append(items, state.LabelPickerItem{
			Label:    label,
			Selected: labelIDMap[label.ID],
		})
	}

	// Initialize LabelPickerState
	w.LabelPickerState.Items = items
	w.LabelPickerState.TaskID = w.FormState.EditingTaskID // 0 for create mode
	w.LabelPickerState.Cursor = 0
	w.LabelPickerState.Filter = ""
	w.LabelPickerState.ReturnMode = state.TicketFormMode

	return true
}

// GetFilteredLabelPickerItems returns label picker items filtered by the current filter text
func (w *Wrapper) GetFilteredLabelPickerItems() []state.LabelPickerItem {
	// Delegate to LabelPickerState which now owns this logic
	return w.LabelPickerState.GetFilteredItems()
}

// InitPriorityPickerForForm initializes the priority picker for use in TicketFormMode.
// Loads the current priority from the form state.
func (w *Wrapper) InitPriorityPickerForForm() bool {
	// Get current priority ID from form state
	// If we're editing a task, get it from the task detail
	// Otherwise, default to medium (id=3)
	currentPriorityID := 3 // Default to medium

	// If editing an existing task, we need to get the current priority from database
	if w.FormState.EditingTaskID != 0 {
		ctx, cancel := w.DbContext()
		defer cancel()

		taskDetail, err := w.Repo.GetTaskDetail(ctx, w.FormState.EditingTaskID)
		if err != nil {
			slog.Error("Error loading task detail for priority picker", "error", err)
			return false
		}

		// Find the priority ID from the priority description
		// We need to match it against our priority options
		priorities := renderers.GetPriorityOptions()
		for _, p := range priorities {
			if p.Description == taskDetail.PriorityDescription {
				currentPriorityID = p.ID
				break
			}
		}
	}

	// Initialize PriorityPickerState
	w.PriorityPickerState.SetSelectedPriorityID(currentPriorityID)
	// Set cursor to match the selected priority (adjust for 0-indexing)
	w.PriorityPickerState.SetCursor(currentPriorityID - 1)
	w.PriorityPickerState.SetReturnMode(state.TicketFormMode)

	return true
}

// InitTypePickerForForm initializes the type picker for use in TicketFormMode.
// Loads the current type from the form state.
func (w *Wrapper) InitTypePickerForForm() bool {
	// Get current type ID from form state
	// If we're editing a task, get it from the task detail
	// Otherwise, default to task (id=1)
	currentTypeID := 1 // Default to task

	// If editing an existing task, we need to get the current type from database
	if w.FormState.EditingTaskID != 0 {
		ctx, cancel := w.DbContext()
		defer cancel()

		taskDetail, err := w.Repo.GetTaskDetail(ctx, w.FormState.EditingTaskID)
		if err != nil {
			slog.Error("Error loading task detail for type picker", "error", err)
			return false
		}

		// Find the type ID from the type description
		// We need to match it against our type options
		types := renderers.GetTypeOptions()
		for _, t := range types {
			if t.Description == taskDetail.TypeDescription {
				currentTypeID = t.ID
				break
			}
		}
	}

	// Initialize TypePickerState
	w.TypePickerState.SetSelectedTypeID(currentTypeID)
	// Set cursor to match the selected type (adjust for 0-indexing)
	w.TypePickerState.SetCursor(currentTypeID - 1)
	w.TypePickerState.SetReturnMode(state.TicketFormMode)

	return true
}
