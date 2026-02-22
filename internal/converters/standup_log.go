package converters

import (
	"github.com/thenoetrevino/paso/internal/database/types"
	"github.com/thenoetrevino/paso/internal/models"
)

// StandupLogToModel converts a database StandupLog to a model StandupLog.
func StandupLogToModel(s types.StandupLog) *models.StandupLog {
	return &models.StandupLog{
		ID:        int(s.ID),
		ProjectID: int(s.ProjectID),
		Content:   s.Content,
		CreatedAt: s.CreatedAt.Time,
	}
}

// StandupLogsToModels converts a slice of database StandupLogs to model StandupLogs.
func StandupLogsToModels(logs []types.StandupLog) []models.StandupLog {
	result := make([]models.StandupLog, len(logs))
	for i, s := range logs {
		result[i] = *StandupLogToModel(s)
	}
	return result
}
