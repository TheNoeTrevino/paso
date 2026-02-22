package converters

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/thenoetrevino/paso/internal/database/types"
)

func TestStandupLogToModel(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 2, 22, 10, 30, 0, 0, time.UTC)

	dbLog := types.StandupLog{
		ID:        1,
		ProjectID: 5,
		Content:   "Fixed the auth bug",
		CreatedAt: types.NullTime{Time: now, Valid: true},
	}

	model := StandupLogToModel(dbLog)

	assert.NotNil(t, model)
	assert.Equal(t, 1, model.ID)
	assert.Equal(t, 5, model.ProjectID)
	assert.Equal(t, "Fixed the auth bug", model.Content)
	assert.Equal(t, now, model.CreatedAt)
}

func TestStandupLogsToModels(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 2, 22, 10, 30, 0, 0, time.UTC)

	dbLogs := []types.StandupLog{
		{
			ID:        1,
			ProjectID: 5,
			Content:   "First log",
			CreatedAt: types.NullTime{Time: now, Valid: true},
		},
		{
			ID:        2,
			ProjectID: 5,
			Content:   "Second log",
			CreatedAt: types.NullTime{Time: now.Add(time.Hour), Valid: true},
		},
	}

	models := StandupLogsToModels(dbLogs)

	assert.Len(t, models, 2)
	assert.Equal(t, 1, models[0].ID)
	assert.Equal(t, "First log", models[0].Content)
	assert.Equal(t, 2, models[1].ID)
	assert.Equal(t, "Second log", models[1].Content)
}

func TestStandupLogsToModels_Empty(t *testing.T) {
	t.Parallel()

	models := StandupLogsToModels([]types.StandupLog{})
	assert.Empty(t, models)
	assert.NotNil(t, models)
}
