package converters

import (
	"github.com/thenoetrevino/paso/internal/database/types"
	"github.com/thenoetrevino/paso/internal/models"
)

// AssigneeToModel converts a types.Assignee (SQLC database model) to models.Assignee (domain model).
//
// Assignees require type conversion:
// - ID fields: int64 → int
// - Timestamps: NullTime → time.Time (using Value() method)
//
// Example usage:
//
//	dbAssignee, _ := queries.GetAssigneeByID(ctx, assigneeID)
//	assignee := converters.AssigneeToModel(dbAssignee)
func AssigneeToModel(a types.Assignee) *models.Assignee {
	return &models.Assignee{
		ID:        int(a.ID),
		Name:      a.Name,
		CreatedAt: a.CreatedAt.Time,
		UpdatedAt: a.UpdatedAt.Time,
	}
}

// AssigneesToModels converts a slice of types.Assignee to a slice of models.Assignee.
// Preserves order of assignees from database query.
//
// Returns empty slice (not nil) for nil or empty input to maintain consistent API.
//
// Example usage:
//
//	dbAssignees, _ := queries.ListAssignees(ctx)
//	assignees := converters.AssigneesToModels(dbAssignees)
func AssigneesToModels(assignees []types.Assignee) []*models.Assignee {
	result := make([]*models.Assignee, len(assignees))
	for i, a := range assignees {
		result[i] = AssigneeToModel(a)
	}
	return result
}
