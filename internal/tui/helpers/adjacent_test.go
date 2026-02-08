package helpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
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

	assert.Equal(t, commands.AdjacentTasks{}, result)
}

func TestGetAdjacentTaskIDs_NilColumns(t *testing.T) {
	result := GetAdjacentTaskIDs(
		nil,
		map[int][]*models.TaskSummary{},
		0,
		0,
	)

	assert.Equal(t, commands.AdjacentTasks{}, result)
}

func TestGetAdjacentTaskIDs_SingleColumnNoTasks(t *testing.T) {
	columns := []*models.Column{{ID: 1, Name: "Todo"}}
	tasks := map[int][]*models.TaskSummary{
		1: {},
	}

	result := GetAdjacentTaskIDs(columns, tasks, 0, 0)

	assert.Equal(t, 0, result.Current)
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
			assert.Equal(t, commands.AdjacentTasks{}, result)
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
			assert.Equal(t, commands.AdjacentTasks{}, result)
		})
	}
}

func TestGetAdjacentTaskIDs_SingleColumnSingleTask(t *testing.T) {
	columns := []*models.Column{{ID: 1, Name: "Todo"}}
	tasks := map[int][]*models.TaskSummary{
		1: {{ID: 42, Title: "Only Task"}},
	}

	result := GetAdjacentTaskIDs(columns, tasks, 0, 0)

	assert.Equal(t, 42, result.Current)
	assert.Equal(t, 0, result.Above)
	assert.Equal(t, 0, result.Below)
	assert.Equal(t, 0, result.Left)
	assert.Equal(t, 0, result.Right)
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

	assert.Equal(t, 0, result.Left)
	assert.Equal(t, 20, result.Right)
	assert.Equal(t, 10, result.Current)
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

	assert.Equal(t, 0, result.Right)
	assert.Equal(t, 20, result.Left)
	assert.Equal(t, 30, result.Current)
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

	assert.Equal(t, 0, result.Above)
	assert.Equal(t, 20, result.Below)
	assert.Equal(t, 10, result.Current)
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

	assert.Equal(t, 0, result.Below)
	assert.Equal(t, 20, result.Above)
	assert.Equal(t, 30, result.Current)
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

	assert.Equal(t, 21, result.Current)
	assert.Equal(t, 20, result.Above)
	assert.Equal(t, 22, result.Below)
	assert.Equal(t, 10, result.Left)
	assert.Equal(t, 30, result.Right)
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

	assert.Equal(t, 0, result.Left)
	assert.Equal(t, 20, result.Current)
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

	assert.Equal(t, 0, result.Right)
	assert.Equal(t, 10, result.Current)
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

	assert.Equal(t, 20, result.Right)
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

	assert.Equal(t, 21, result.Right)
}

func TestGetAdjacentTaskIDs_NilTasksMap(t *testing.T) {
	columns := []*models.Column{{ID: 1, Name: "Todo"}}

	result := GetAdjacentTaskIDs(columns, nil, 0, 0)

	assert.Equal(t, commands.AdjacentTasks{}, result)
}

func TestGetAdjacentTaskIDs_MissingColumnInTasksMap(t *testing.T) {
	columns := []*models.Column{{ID: 1, Name: "Todo"}}
	tasks := map[int][]*models.TaskSummary{}

	result := GetAdjacentTaskIDs(columns, tasks, 0, 0)

	assert.Equal(t, 0, result.Current)
}
