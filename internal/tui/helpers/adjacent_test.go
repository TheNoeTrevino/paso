package helpers

import (
	"testing"

	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui/commands"
)

func TestGetAdjacentTaskIDs_EmptyColumns(t *testing.T) {
	result := GetAdjacentTaskIDs(
		[]*models.Column{},
		map[int][]*models.TaskSummary{},
		0,
		0,
	)

	want := commands.AdjacentTasks{}
	if result != want {
		t.Errorf("GetAdjacentTaskIDs with empty columns = %+v, want zero-value struct", result)
	}
}

func TestGetAdjacentTaskIDs_NilColumns(t *testing.T) {
	result := GetAdjacentTaskIDs(
		nil,
		map[int][]*models.TaskSummary{},
		0,
		0,
	)

	want := commands.AdjacentTasks{}
	if result != want {
		t.Errorf("GetAdjacentTaskIDs with nil columns = %+v, want zero-value struct", result)
	}
}

func TestGetAdjacentTaskIDs_SingleColumnNoTasks(t *testing.T) {
	columns := []*models.Column{{ID: 1, Name: "Todo"}}
	tasks := map[int][]*models.TaskSummary{
		1: {},
	}

	result := GetAdjacentTaskIDs(columns, tasks, 0, 0)

	if result.Current != 0 {
		t.Errorf("Current = %d, want 0 (no tasks in column)", result.Current)
	}
}

func TestGetAdjacentTaskIDs_SelectedColumnOutOfBounds(t *testing.T) {
	columns := []*models.Column{
		{ID: 1, Name: "Todo"},
		{ID: 2, Name: "Done"},
	}
	tasks := map[int][]*models.TaskSummary{
		1: {{ID: 10, Title: "Task 1"}},
		2: {{ID: 20, Title: "Task 2"}},
	}

	tests := []struct {
		name        string
		selectedCol int
	}{
		{"negative index", -1},
		{"index equals length", 2},
		{"index way out of bounds", 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetAdjacentTaskIDs(columns, tasks, tt.selectedCol, 0)
			want := commands.AdjacentTasks{}
			if result != want {
				t.Errorf("GetAdjacentTaskIDs with selectedCol=%d = %+v, want zero-value struct", tt.selectedCol, result)
			}
		})
	}
}

func TestGetAdjacentTaskIDs_SelectedTaskOutOfBounds(t *testing.T) {
	columns := []*models.Column{{ID: 1, Name: "Todo"}}
	tasks := map[int][]*models.TaskSummary{
		1: {
			{ID: 10, Title: "Task 1"},
			{ID: 20, Title: "Task 2"},
		},
	}

	tests := []struct {
		name         string
		selectedTask int
	}{
		{"negative index", -1},
		{"index equals length", 2},
		{"index way out of bounds", 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetAdjacentTaskIDs(columns, tasks, 0, tt.selectedTask)
			want := commands.AdjacentTasks{}
			if result != want {
				t.Errorf("GetAdjacentTaskIDs with selectedTask=%d = %+v, want zero-value struct", tt.selectedTask, result)
			}
		})
	}
}

func TestGetAdjacentTaskIDs_SingleColumnSingleTask(t *testing.T) {
	columns := []*models.Column{{ID: 1, Name: "Todo"}}
	tasks := map[int][]*models.TaskSummary{
		1: {{ID: 42, Title: "Only Task"}},
	}

	result := GetAdjacentTaskIDs(columns, tasks, 0, 0)

	if result.Current != 42 {
		t.Errorf("Current = %d, want 42", result.Current)
	}
	if result.Above != 0 {
		t.Errorf("Above = %d, want 0 (no task above)", result.Above)
	}
	if result.Below != 0 {
		t.Errorf("Below = %d, want 0 (no task below)", result.Below)
	}
	if result.Left != 0 {
		t.Errorf("Left = %d, want 0 (no column to left)", result.Left)
	}
	if result.Right != 0 {
		t.Errorf("Right = %d, want 0 (no column to right)", result.Right)
	}
}

func TestGetAdjacentTaskIDs_AtFirstColumn(t *testing.T) {
	columns := []*models.Column{
		{ID: 1, Name: "Todo"},
		{ID: 2, Name: "In Progress"},
		{ID: 3, Name: "Done"},
	}
	tasks := map[int][]*models.TaskSummary{
		1: {{ID: 10, Title: "Task 1"}},
		2: {{ID: 20, Title: "Task 2"}},
		3: {{ID: 30, Title: "Task 3"}},
	}

	result := GetAdjacentTaskIDs(columns, tasks, 0, 0)

	if result.Left != 0 {
		t.Errorf("Left at first column = %d, want 0", result.Left)
	}
	if result.Right != 20 {
		t.Errorf("Right at first column = %d, want 20", result.Right)
	}
	if result.Current != 10 {
		t.Errorf("Current = %d, want 10", result.Current)
	}
}

func TestGetAdjacentTaskIDs_AtLastColumn(t *testing.T) {
	columns := []*models.Column{
		{ID: 1, Name: "Todo"},
		{ID: 2, Name: "In Progress"},
		{ID: 3, Name: "Done"},
	}
	tasks := map[int][]*models.TaskSummary{
		1: {{ID: 10, Title: "Task 1"}},
		2: {{ID: 20, Title: "Task 2"}},
		3: {{ID: 30, Title: "Task 3"}},
	}

	result := GetAdjacentTaskIDs(columns, tasks, 2, 0)

	if result.Right != 0 {
		t.Errorf("Right at last column = %d, want 0", result.Right)
	}
	if result.Left != 20 {
		t.Errorf("Left at last column = %d, want 20", result.Left)
	}
	if result.Current != 30 {
		t.Errorf("Current = %d, want 30", result.Current)
	}
}

func TestGetAdjacentTaskIDs_AtFirstTaskInColumn(t *testing.T) {
	columns := []*models.Column{{ID: 1, Name: "Todo"}}
	tasks := map[int][]*models.TaskSummary{
		1: {
			{ID: 10, Title: "Task 1"},
			{ID: 20, Title: "Task 2"},
			{ID: 30, Title: "Task 3"},
		},
	}

	result := GetAdjacentTaskIDs(columns, tasks, 0, 0)

	if result.Above != 0 {
		t.Errorf("Above at first task = %d, want 0", result.Above)
	}
	if result.Below != 20 {
		t.Errorf("Below at first task = %d, want 20", result.Below)
	}
	if result.Current != 10 {
		t.Errorf("Current = %d, want 10", result.Current)
	}
}

func TestGetAdjacentTaskIDs_AtLastTaskInColumn(t *testing.T) {
	columns := []*models.Column{{ID: 1, Name: "Todo"}}
	tasks := map[int][]*models.TaskSummary{
		1: {
			{ID: 10, Title: "Task 1"},
			{ID: 20, Title: "Task 2"},
			{ID: 30, Title: "Task 3"},
		},
	}

	result := GetAdjacentTaskIDs(columns, tasks, 0, 2)

	if result.Below != 0 {
		t.Errorf("Below at last task = %d, want 0", result.Below)
	}
	if result.Above != 20 {
		t.Errorf("Above at last task = %d, want 20", result.Above)
	}
	if result.Current != 30 {
		t.Errorf("Current = %d, want 30", result.Current)
	}
}

func TestGetAdjacentTaskIDs_MiddlePosition(t *testing.T) {
	columns := []*models.Column{
		{ID: 1, Name: "Todo"},
		{ID: 2, Name: "In Progress"},
		{ID: 3, Name: "Done"},
	}
	tasks := map[int][]*models.TaskSummary{
		1: {{ID: 10, Title: "Task 1"}},
		2: {
			{ID: 20, Title: "Task 2.1"},
			{ID: 21, Title: "Task 2.2"},
			{ID: 22, Title: "Task 2.3"},
		},
		3: {{ID: 30, Title: "Task 3"}},
	}

	result := GetAdjacentTaskIDs(columns, tasks, 1, 1)

	if result.Current != 21 {
		t.Errorf("Current = %d, want 21", result.Current)
	}
	if result.Above != 20 {
		t.Errorf("Above = %d, want 20", result.Above)
	}
	if result.Below != 22 {
		t.Errorf("Below = %d, want 22", result.Below)
	}
	if result.Left != 10 {
		t.Errorf("Left = %d, want 10", result.Left)
	}
	if result.Right != 30 {
		t.Errorf("Right = %d, want 30", result.Right)
	}
}

func TestGetAdjacentTaskIDs_LeftColumnEmpty(t *testing.T) {
	columns := []*models.Column{
		{ID: 1, Name: "Todo"},
		{ID: 2, Name: "In Progress"},
	}
	tasks := map[int][]*models.TaskSummary{
		1: {},
		2: {{ID: 20, Title: "Task 2"}},
	}

	result := GetAdjacentTaskIDs(columns, tasks, 1, 0)

	if result.Left != 0 {
		t.Errorf("Left with empty left column = %d, want 0", result.Left)
	}
	if result.Current != 20 {
		t.Errorf("Current = %d, want 20", result.Current)
	}
}

func TestGetAdjacentTaskIDs_RightColumnEmpty(t *testing.T) {
	columns := []*models.Column{
		{ID: 1, Name: "Todo"},
		{ID: 2, Name: "In Progress"},
	}
	tasks := map[int][]*models.TaskSummary{
		1: {{ID: 10, Title: "Task 1"}},
		2: {},
	}

	result := GetAdjacentTaskIDs(columns, tasks, 0, 0)

	if result.Right != 0 {
		t.Errorf("Right with empty right column = %d, want 0", result.Right)
	}
	if result.Current != 10 {
		t.Errorf("Current = %d, want 10", result.Current)
	}
}

func TestGetAdjacentTaskIDs_AdjacentColumnIndexClamping(t *testing.T) {
	columns := []*models.Column{
		{ID: 1, Name: "Todo"},
		{ID: 2, Name: "In Progress"},
	}
	tasks := map[int][]*models.TaskSummary{
		1: {
			{ID: 10, Title: "Task 1.1"},
			{ID: 11, Title: "Task 1.2"},
			{ID: 12, Title: "Task 1.3"},
		},
		2: {{ID: 20, Title: "Task 2.1"}},
	}

	result := GetAdjacentTaskIDs(columns, tasks, 0, 2)

	if result.Right != 20 {
		t.Errorf("Right (index clamped) = %d, want 20 (first task in shorter column)", result.Right)
	}
}

func TestGetAdjacentTaskIDs_AdjacentColumnPreservesIndex(t *testing.T) {
	columns := []*models.Column{
		{ID: 1, Name: "Todo"},
		{ID: 2, Name: "In Progress"},
	}
	tasks := map[int][]*models.TaskSummary{
		1: {
			{ID: 10, Title: "Task 1.1"},
			{ID: 11, Title: "Task 1.2"},
		},
		2: {
			{ID: 20, Title: "Task 2.1"},
			{ID: 21, Title: "Task 2.2"},
			{ID: 22, Title: "Task 2.3"},
		},
	}

	result := GetAdjacentTaskIDs(columns, tasks, 0, 1)

	if result.Right != 21 {
		t.Errorf("Right (same index preserved) = %d, want 21 (same index in right column)", result.Right)
	}
}

func TestGetAdjacentTaskIDs_NilTasksMap(t *testing.T) {
	columns := []*models.Column{{ID: 1, Name: "Todo"}}

	result := GetAdjacentTaskIDs(columns, nil, 0, 0)

	want := commands.AdjacentTasks{}
	if result != want {
		t.Errorf("GetAdjacentTaskIDs with nil tasks map = %+v, want zero-value struct", result)
	}
}

func TestGetAdjacentTaskIDs_MissingColumnInTasksMap(t *testing.T) {
	columns := []*models.Column{{ID: 1, Name: "Todo"}}
	tasks := map[int][]*models.TaskSummary{}

	result := GetAdjacentTaskIDs(columns, tasks, 0, 0)

	if result.Current != 0 {
		t.Errorf("Current with missing column in tasks map = %d, want 0", result.Current)
	}
}
