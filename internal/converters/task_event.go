package converters

import (
	"github.com/thenoetrevino/paso/internal/database/types"
	"github.com/thenoetrevino/paso/internal/models"
)

// TaskEventToModel converts a database TaskEvent to a model TaskEvent.
func TaskEventToModel(e types.TaskEvent) *models.TaskEvent {
	return &models.TaskEvent{
		ID:        int(e.ID),
		TaskID:    int(e.TaskID),
		Content:   e.Content,
		Author:    e.Author,
		CreatedAt: e.CreatedAt.Time,
	}
}

// TaskEventsToModels converts a slice of database TaskEvents to model TaskEvents.
func TaskEventsToModels(events []types.TaskEvent) []models.TaskEvent {
	result := make([]models.TaskEvent, len(events))
	for i, e := range events {
		result[i] = *TaskEventToModel(e)
	}
	return result
}
