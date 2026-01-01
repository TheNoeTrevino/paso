package modelops

import (
	"log/slog"

	"github.com/thenoetrevino/paso/internal/tui"
	"github.com/thenoetrevino/paso/internal/tui/renderers"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// InitParentPickerForForm initializes the parent picker for use in task form mode.
// In edit mode: loads existing parent relationships from FormState.
// In create mode: starts with empty selection (relationships applied after task creation).
//
// Returns false if there's no current project.
func InitParentPickerForForm(m *tui.Model) bool {
	project := GetCurrentProject(m)
	if project == nil {
		return false
	}

	// Get all task references for the entire project
	ctx, cancel := m.DBContext()
	defer cancel()
	allTasks, err := m.App.TaskService.GetTaskReferencesForProject(ctx, project.ID)
	if err != nil {
		slog.Error("Error loading project tasks", "error", err)
		return false
	}

	// Build map of currently selected parent task IDs and their relation types from form state
	parentTaskMap := make(map[int]int) // map[taskID]relationTypeID
	for _, parentRef := range m.FormState.FormParentRefs {
		parentTaskMap[parentRef.ID] = parentRef.RelationTypeID
	}

	// Build map of child task IDs to exclude (prevent circular dependencies)
	// A task that is already a child cannot become a parent
	childTaskMap := make(map[int]bool)
	for _, childID := range m.FormState.FormChildIDs {
		childTaskMap[childID] = true
	}

	// Build picker items from all tasks, excluding current task (if editing)
	// and excluding tasks that are already children (circular dependency prevention)
	items := make([]state.TaskPickerItem, 0, len(allTasks))
	for _, task := range allTasks {
		// In edit mode, exclude the task being edited
		if m.FormState.EditingTaskID != 0 && task.ID == m.FormState.EditingTaskID {
			continue
		}

		// Exclude tasks that are already children (prevents circular dependencies)
		if childTaskMap[task.ID] {
			continue
		}

		relationTypeID, isSelected := parentTaskMap[task.ID]

		items = append(items, state.TaskPickerItem{
			TaskRef:        task,
			Selected:       isSelected,
			RelationTypeID: relationTypeID,
		})
	}

	// Initialize ParentPickerState
	m.ParentPickerState.Items = items
	m.ParentPickerState.TaskID = m.FormState.EditingTaskID // 0 for create mode
	m.ParentPickerState.Cursor = 0
	m.ParentPickerState.Filter = ""
	m.ParentPickerState.PickerType = "parent"
	m.ParentPickerState.ReturnMode = state.TicketFormMode

	return true
}

// InitChildPickerForForm initializes the child picker for use in task form mode.
// In edit mode: loads existing child relationships from FormState.
// In create mode: starts with empty selection (relationships applied after task creation).
//
// Returns false if there's no current project.
func InitChildPickerForForm(m *tui.Model) bool {
	project := GetCurrentProject(m)
	if project == nil {
		return false
	}

	// Get all task references for the entire project
	ctx, cancel := m.DBContext()
	defer cancel()
	allTasks, err := m.App.TaskService.GetTaskReferencesForProject(ctx, project.ID)
	if err != nil {
		slog.Error("Error loading project tasks", "error", err)
		return false
	}

	// Build map of currently selected child task IDs and their relation types from form state
	childTaskMap := make(map[int]int) // map[taskID]relationTypeID
	for _, childRef := range m.FormState.FormChildRefs {
		childTaskMap[childRef.ID] = childRef.RelationTypeID
	}

	// Build map of parent task IDs to exclude (prevent circular dependencies)
	// A task that is already a parent cannot become a child
	parentTaskMap := make(map[int]bool)
	for _, parentID := range m.FormState.FormParentIDs {
		parentTaskMap[parentID] = true
	}

	// Build picker items from all tasks, excluding current task (if editing)
	// and excluding tasks that are already parents (circular dependency prevention)
	items := make([]state.TaskPickerItem, 0, len(allTasks))
	for _, task := range allTasks {
		// In edit mode, exclude the task being edited
		if m.FormState.EditingTaskID != 0 && task.ID == m.FormState.EditingTaskID {
			continue
		}
		if parentTaskMap[task.ID] {
			continue
		}

		relationTypeID, isSelected := childTaskMap[task.ID]

		items = append(items, state.TaskPickerItem{
			TaskRef:        task,
			Selected:       isSelected,
			RelationTypeID: relationTypeID,
		})
	}

	// Initialize ChildPickerState
	m.ChildPickerState.Items = items
	m.ChildPickerState.TaskID = m.FormState.EditingTaskID // 0 for create mode
	m.ChildPickerState.Cursor = 0
	m.ChildPickerState.Filter = ""
	m.ChildPickerState.PickerType = "child"
	m.ChildPickerState.ReturnMode = state.TicketFormMode

	return true
}

// InitLabelPickerForForm initializes the label picker for use in task form mode.
// In edit mode: loads existing label selections from FormState.
// In create mode: starts with empty selection (labels applied on form submission).
//
// Returns false if there's no current project.
func InitLabelPickerForForm(m *tui.Model) bool {
	project := GetCurrentProject(m)
	if project == nil {
		return false
	}

	// Build map of currently selected label IDs from form state
	labelIDMap := make(map[int]bool)
	for _, labelID := range m.FormState.FormLabelIDs {
		labelIDMap[labelID] = true
	}

	// Build picker items from all available labels
	var items []state.LabelPickerItem
	for _, label := range m.AppState.Labels() {
		items = append(items, state.LabelPickerItem{
			Label:    label,
			Selected: labelIDMap[label.ID],
		})
	}

	// Initialize LabelPickerState
	m.LabelPickerState.Items = items
	m.LabelPickerState.TaskID = m.FormState.EditingTaskID // 0 for create mode
	m.LabelPickerState.Cursor = 0
	m.LabelPickerState.Filter = ""
	m.LabelPickerState.ReturnMode = state.TicketFormMode

	return true
}

// GetFilteredLabelPickerItems returns label picker items filtered by the current filter text
func GetFilteredLabelPickerItems(m *tui.Model) []state.LabelPickerItem {
	// Delegate to LabelPickerState which now owns this logic
	return m.LabelPickerState.GetFilteredItems()
}

// InitPriorityPickerForForm initializes the priority picker for use in task form mode.
// Loads the current priority from the form state.
func InitPriorityPickerForForm(m *tui.Model) bool {
	// Get current priority ID from form state
	// If we're editing a task, get it from the task detail
	// Otherwise, default to medium (id=3)
	currentPriorityID := 3 // Default to medium

	// If editing an existing task, we need to get the current priority from database
	if m.FormState.EditingTaskID != 0 {
		ctx, cancel := m.DBContext()
		defer cancel()

		taskDetail, err := m.App.TaskService.GetTaskDetail(ctx, m.FormState.EditingTaskID)
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
	m.PriorityPickerState.SetSelectedPriorityID(currentPriorityID)
	// Set cursor to match the selected priority (adjust for 0-indexing)
	m.PriorityPickerState.SetCursor(currentPriorityID - 1)
	m.PriorityPickerState.ReturnMode = state.TicketFormMode

	return true
}

// InitTypePickerForForm initializes the type picker for use in task form mode.
// Loads the current type from the form state.
func InitTypePickerForForm(m *tui.Model) bool {
	// Get current type ID from form state
	// If we're editing a task, get it from the task detail
	// Otherwise, default to task (id=1)
	currentTypeID := 1 // Default to task

	// If editing an existing task, we need to get the current type from database
	if m.FormState.EditingTaskID != 0 {
		ctx, cancel := m.DBContext()
		defer cancel()

		taskDetail, err := m.App.TaskService.GetTaskDetail(ctx, m.FormState.EditingTaskID)
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
	m.TypePickerState.SetSelectedTypeID(currentTypeID)
	// Set cursor to match the selected type (adjust for 0-indexing)
	m.TypePickerState.SetCursor(currentTypeID - 1)
	m.TypePickerState.ReturnMode = state.TicketFormMode

	return true
}
