package converters

import (
	"github.com/thenoetrevino/paso/internal/database"
	"github.com/thenoetrevino/paso/internal/database/generated"
	"github.com/thenoetrevino/paso/internal/models"
)

// ColumnToModel converts a generated.Column (SQLC database model) to models.Column (domain model).
//
// Handles nullable foreign keys for linked list structure:
// - prev_id (interface{} → *int) - Previous column in workflow
// - next_id (interface{} → *int) - Next column in workflow
//
// Special flags indicate column purpose:
// - HoldsReadyTasks: Column contains tasks ready to work on
// - HoldsCompletedTasks: Column contains finished tasks
// - HoldsInProgressTasks: Column contains tasks currently being worked on
//
// Type conversions:
// - All ID fields: int64 → int
// - Nullable IDs: interface{} → *int using database.AnyToIntPtr()
func ColumnToModel(c generated.Column) *models.Column {
	return &models.Column{
		ID:                   int(c.ID),
		Name:                 c.Name,
		ProjectID:            int(c.ProjectID),
		PrevID:               database.AnyToIntPtr(c.PrevID),
		NextID:               database.AnyToIntPtr(c.NextID),
		HoldsReadyTasks:      c.HoldsReadyTasks,
		HoldsCompletedTasks:  c.HoldsCompletedTasks,
		HoldsInProgressTasks: c.HoldsInProgressTasks,
	}
}

// ColumnFromIDRowToModel converts a generated.GetColumnByIDRow to models.Column.
// This handles the specific row type returned by the GetColumnByID query.
// See ColumnToModel for field conversion details.
func ColumnFromIDRowToModel(r generated.GetColumnByIDRow) *models.Column {
	return &models.Column{
		ID:                   int(r.ID),
		Name:                 r.Name,
		ProjectID:            int(r.ProjectID),
		PrevID:               database.AnyToIntPtr(r.PrevID),
		NextID:               database.AnyToIntPtr(r.NextID),
		HoldsReadyTasks:      r.HoldsReadyTasks,
		HoldsCompletedTasks:  r.HoldsCompletedTasks,
		HoldsInProgressTasks: r.HoldsInProgressTasks,
	}
}

// ColumnsFromRowsToModels converts a slice of generated.GetColumnsByProjectRow to a slice of models.Column.
// This handles the specific row type returned by the GetColumnsByProject query.
// See ColumnToModel for field conversion details.
func ColumnsFromRowsToModels(rows []generated.GetColumnsByProjectRow) []*models.Column {
	result := make([]*models.Column, len(rows))
	for i, r := range rows {
		result[i] = &models.Column{
			ID:                   int(r.ID),
			Name:                 r.Name,
			ProjectID:            int(r.ProjectID),
			PrevID:               database.AnyToIntPtr(r.PrevID),
			NextID:               database.AnyToIntPtr(r.NextID),
			HoldsReadyTasks:      r.HoldsReadyTasks,
			HoldsCompletedTasks:  r.HoldsCompletedTasks,
			HoldsInProgressTasks: r.HoldsInProgressTasks,
		}
	}
	return result
}
