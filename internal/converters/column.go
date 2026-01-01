package converters

import (
	"github.com/thenoetrevino/paso/internal/database"
	"github.com/thenoetrevino/paso/internal/database/generated"
	"github.com/thenoetrevino/paso/internal/models"
)

// ColumnToModel converts a generated.Column to models.Column
func ColumnToModel(c generated.Column) *models.Column {
	return &models.Column{
		ID:                   int(c.ID),
		Name:                 c.Name,
		ProjectID:            int(c.ProjectID),
		PrevID:               database.InterfaceToIntPtr(c.PrevID),
		NextID:               database.InterfaceToIntPtr(c.NextID),
		HoldsReadyTasks:      c.HoldsReadyTasks,
		HoldsCompletedTasks:  c.HoldsCompletedTasks,
		HoldsInProgressTasks: c.HoldsInProgressTasks,
	}
}

// ColumnFromIDRowToModel converts a generated.GetColumnByIDRow to models.Column
func ColumnFromIDRowToModel(r generated.GetColumnByIDRow) *models.Column {
	return &models.Column{
		ID:                   int(r.ID),
		Name:                 r.Name,
		ProjectID:            int(r.ProjectID),
		PrevID:               database.InterfaceToIntPtr(r.PrevID),
		NextID:               database.InterfaceToIntPtr(r.NextID),
		HoldsReadyTasks:      r.HoldsReadyTasks,
		HoldsCompletedTasks:  r.HoldsCompletedTasks,
		HoldsInProgressTasks: r.HoldsInProgressTasks,
	}
}

// ColumnsFromRowsToModels converts a slice of generated.GetColumnsByProjectRow to a slice of models.Column
func ColumnsFromRowsToModels(rows []generated.GetColumnsByProjectRow) []*models.Column {
	result := make([]*models.Column, len(rows))
	for i, r := range rows {
		result[i] = &models.Column{
			ID:                   int(r.ID),
			Name:                 r.Name,
			ProjectID:            int(r.ProjectID),
			PrevID:               database.InterfaceToIntPtr(r.PrevID),
			NextID:               database.InterfaceToIntPtr(r.NextID),
			HoldsReadyTasks:      r.HoldsReadyTasks,
			HoldsCompletedTasks:  r.HoldsCompletedTasks,
			HoldsInProgressTasks: r.HoldsInProgressTasks,
		}
	}
	return result
}
