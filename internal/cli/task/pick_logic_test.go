package task

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thenoetrevino/paso/internal/models"
)

func TestTaskSummaryPickerLabel(t *testing.T) {
	tests := []struct {
		name string
		task *models.TaskSummary
		want string
	}{
		{
			name: "task with priority and type",
			task: &models.TaskSummary{
				ID:                  14,
				Title:               "Fix login bug",
				PriorityDescription: "high",
				TypeDescription:     "bug",
			},
			want: "Fix login bug (#14) — high, bug",
		},
		{
			name: "task with only priority",
			task: &models.TaskSummary{
				ID:                  5,
				Title:               "Update docs",
				PriorityDescription: "low",
			},
			want: "Update docs (#5) — low",
		},
		{
			name: "task with only type",
			task: &models.TaskSummary{
				ID:              22,
				Title:           "Add feature",
				TypeDescription: "feature",
			},
			want: "Add feature (#22) — feature",
		},
		{
			name: "task with no priority or type",
			task: &models.TaskSummary{
				ID:    1,
				Title: "Simple task",
			},
			want: "Simple task (#1)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.task.PickerLabel()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTaskSummaryPickerValue(t *testing.T) {
	task := &models.TaskSummary{ID: 42}
	assert.Equal(t, "42", task.PickerValue())
}
