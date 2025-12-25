package tui

import (
	"context"
	"time"

	"github.com/thenoetrevino/paso/internal/config"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// createTestProjects creates a slice of test projects for testing.
func createTestProjects(count int) []*models.Project {
	projects := make([]*models.Project, count)
	for i := 0; i < count; i++ {
		projects[i] = &models.Project{
			ID:          i + 1,
			Name:        "Project " + string(rune('A'+i)),
			Description: "Test project description",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
	}
	return projects
}

// createTestColumns creates test columns for a given project ID.
func createTestColumns(projectID int, count int) []*models.Column {
	columns := make([]*models.Column, count)
	for i := 0; i < count; i++ {
		var prevID, nextID *int
		if i > 0 {
			prev := (projectID * 100) + i - 1
			prevID = &prev
		}
		if i < count-1 {
			next := (projectID * 100) + i + 1
			nextID = &next
		}

		columns[i] = &models.Column{
			ID:        (projectID * 100) + i,
			Name:      "Column " + string(rune('A'+i)),
			ProjectID: projectID,
			PrevID:    prevID,
			NextID:    nextID,
		}
	}
	return columns
}

// createTestTasks creates test tasks for a given column ID.
func createTestTasks(columnID int, count int) []*models.TaskSummary {
	tasks := make([]*models.TaskSummary, count)
	for i := 0; i < count; i++ {
		tasks[i] = &models.TaskSummary{
			ID:                  (columnID * 100) + i + 1,
			Title:               "Task " + string(rune('A'+i)),
			Labels:              []*models.Label{},
			TypeDescription:     "Feature",
			PriorityDescription: "Medium",
			PriorityColor:       "#FFA500",
			ColumnID:            columnID,
			Position:            i,
			IsBlocked:           false,
		}
	}
	return tasks
}

// createTestTasksMap creates a map of column ID to tasks for testing.
func createTestTasksMap(columns []*models.Column, tasksPerColumn int) map[int][]*models.TaskSummary {
	tasksMap := make(map[int][]*models.TaskSummary)
	for _, col := range columns {
		tasksMap[col.ID] = createTestTasks(col.ID, tasksPerColumn)
	}
	return tasksMap
}

// createTestModel creates a basic Model with minimal test data.
// This is useful for tests that only need basic state without a full setup.
func createTestModel() Model {
	cfg := &config.Config{
		KeyMappings: config.DefaultKeyMappings(),
	}

	return Model{
		Ctx:               context.Background(),
		App:               nil, // No app needed for pure state tests
		Config:            cfg,
		AppState:          state.NewAppState(nil, 0, nil, nil, nil),
		UiState:           state.NewUIState(),
		InputState:        state.NewInputState(),
		FormState:         state.NewFormState(),
		LabelPickerState:  state.NewLabelPickerState(),
		ParentPickerState: state.NewTaskPickerState(),
		ChildPickerState:  state.NewTaskPickerState(),
		NotificationState: state.NewNotificationState(),
		SearchState:       state.NewSearchState(),
		ListViewState:     state.NewListViewState(),
		StatusPickerState: state.NewStatusPickerState(),
	}
}

// createTestModelWithProjects creates a Model with multiple projects.
// The first project is set as the active project.
// Each project has the specified number of columns and tasks.
func createTestModelWithProjects(numProjects int, columnsPerProject int, tasksPerColumn int) Model {
	projects := createTestProjects(numProjects)

	// Create columns and tasks for the first project (active project)
	columns := createTestColumns(1, columnsPerProject)
	tasks := createTestTasksMap(columns, tasksPerColumn)

	cfg := &config.Config{
		KeyMappings: config.DefaultKeyMappings(),
	}

	return Model{
		Ctx:               context.Background(),
		App:               nil, // No app needed for subscription tests
		Config:            cfg,
		AppState:          state.NewAppState(projects, 0, columns, tasks, nil),
		UiState:           state.NewUIState(),
		InputState:        state.NewInputState(),
		FormState:         state.NewFormState(),
		LabelPickerState:  state.NewLabelPickerState(),
		ParentPickerState: state.NewTaskPickerState(),
		ChildPickerState:  state.NewTaskPickerState(),
		NotificationState: state.NewNotificationState(),
		SearchState:       state.NewSearchState(),
		ListViewState:     state.NewListViewState(),
		StatusPickerState: state.NewStatusPickerState(),
	}
}
