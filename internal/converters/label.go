package converters

import (
	"github.com/thenoetrevino/paso/internal/database/generated"
	"github.com/thenoetrevino/paso/internal/models"
)

// LabelToModel converts a generated.Label to models.Label
func LabelToModel(l generated.Label) *models.Label {
	return &models.Label{
		ID:        int(l.ID),
		Name:      l.Name,
		Color:     l.Color,
		ProjectID: int(l.ProjectID),
	}
}

// LabelsToModels converts a slice of generated.Label to a slice of models.Label
func LabelsToModels(labels []generated.Label) []*models.Label {
	result := make([]*models.Label, len(labels))
	for i, l := range labels {
		result[i] = LabelToModel(l)
	}
	return result
}
