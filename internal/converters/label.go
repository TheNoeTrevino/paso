package converters

import (
	"github.com/thenoetrevino/paso/internal/database/generated"
	"github.com/thenoetrevino/paso/internal/models"
)

// LabelToModel converts a generated.Label (SQLC database model) to models.Label (domain model).
//
// Labels are simple entities with no NULL fields, but still require type conversion:
// - ID fields: int64 → int
//
// Example usage:
//
//	dbLabel, _ := queries.GetLabelByID(ctx, labelID)
//	label := converters.LabelToModel(dbLabel)
func LabelToModel(l generated.Label) *models.Label {
	return &models.Label{
		ID:        int(l.ID),
		Name:      l.Name,
		Color:     l.Color,
		ProjectID: int(l.ProjectID),
	}
}

// LabelsToModels converts a slice of generated.Label to a slice of models.Label.
// Preserves order of labels from database query.
//
// Returns empty slice (not nil) for nil or empty input to maintain consistent API.
//
// Example usage:
//
//	dbLabels, _ := queries.GetLabelsByProject(ctx, projectID)
//	labels := converters.LabelsToModels(dbLabels)
func LabelsToModels(labels []generated.Label) []*models.Label {
	result := make([]*models.Label, len(labels))
	for i, l := range labels {
		result[i] = LabelToModel(l)
	}
	return result
}
