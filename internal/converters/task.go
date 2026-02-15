// Package converters provides type-safe conversion between
// database models (from SQLC) and domain models.
//
// All conversions handle:
// - NULL database values (sql.Null* types)
// - Type coercions (int64 from database to int in domain)
// - Relationship parsing (GROUP_CONCAT labels)
//
// Conversion failures are explicit - never silent type coercions.
//
// Example usage:
//
//	// Converting a single task
//	task := converters.TaskToModel(dbTask)
//
//	// Converting labels
//	labels := converters.LabelsToModels(dbLabels)
//
//	// Parsing GROUP_CONCAT results
//	labels := converters.ParseLabelsFromConcatenated(ids, names, colors)
package converters

import (
	"fmt"
	"strings"

	"github.com/thenoetrevino/paso/internal/database/types"
	"github.com/thenoetrevino/paso/internal/models"
)

// labelSeparator is used to separate concatenated label fields in queries
// Using CHAR(31) as a delimiter since it's a control character unlikely in text
const labelSeparator = string(rune(31))

// TaskToModel converts a types.Task (SQLC database model) to models.Task (domain model).
//
// Handles NULL values for optional fields:
// - description (types.NullString)
// - created_at, updated_at (types.NullTime)
//
// Type conversions:
// - All ID fields: int64 → int
func TaskToModel(t types.Task) *models.Task {
	task := &models.Task{
		ID:         int(t.ID),
		Title:      t.Title,
		ColumnID:   int(t.ColumnID),
		Position:   int(t.Position),
		TypeID:     int(t.TypeID),
		PriorityID: int(t.PriorityID),
	}

	if t.Description.Valid {
		task.Description = t.Description.String
	}
	if t.CreatedAt.Valid {
		task.CreatedAt = t.CreatedAt.Time
	}
	if t.UpdatedAt.Valid {
		task.UpdatedAt = t.UpdatedAt.Time
	}
	if t.DueDate.Valid {
		task.DueDate = &t.DueDate.Time
	}

	return task
}

// ParentTasksToReferences converts parent task rows to TaskReference slice
func ParentTasksToReferences(rows []types.GetParentTasksRow) []*models.TaskReference {
	result := make([]*models.TaskReference, 0, len(rows))
	for _, row := range rows {
		ref := &models.TaskReference{
			ID:             int(row.ID),
			Title:          row.Title,
			ProjectName:    row.Name,
			RelationTypeID: int(row.ID_2),
			RelationLabel:  row.PToCLabel,
			RelationColor:  row.Color,
			IsBlocking:     row.IsBlocking,
		}
		if row.TicketNumber.Valid {
			ref.TicketNumber = int(row.TicketNumber.Int64)
		}
		result = append(result, ref)
	}
	return result
}

// ChildTasksToReferences converts child task rows to TaskReference slice
func ChildTasksToReferences(rows []types.GetChildTasksRow) []*models.TaskReference {
	result := make([]*models.TaskReference, 0, len(rows))
	for _, row := range rows {
		ref := &models.TaskReference{
			ID:             int(row.ID),
			Title:          row.Title,
			ProjectName:    row.Name,
			RelationTypeID: int(row.ID_2),
			RelationLabel:  row.CToPLabel,
			RelationColor:  row.Color,
			IsBlocking:     row.IsBlocking,
		}
		if row.TicketNumber.Valid {
			ref.TicketNumber = int(row.TicketNumber.Int64)
		}
		result = append(result, ref)
	}
	return result
}

// CommentsToModels converts types.TaskComment slice to models.Comment slice
func CommentsToModels(comments []types.TaskComment) []*models.Comment {
	result := make([]*models.Comment, 0, len(comments))
	for _, c := range comments {
		result = append(result, &models.Comment{
			ID:        int(c.ID),
			TaskID:    int(c.TaskID),
			Message:   c.Content,
			Author:    c.Author,
			CreatedAt: c.CreatedAt.Time,
		})
	}
	return result
}

// TaskSummaryFromRowToModel converts a task summary row to models.TaskSummary
func TaskSummaryFromRowToModel(row types.GetTaskSummariesByProjectRow) *models.TaskSummary {
	summary := &models.TaskSummary{
		ID:        int(row.ID),
		Title:     row.Title,
		ColumnID:  int(row.ColumnID),
		Position:  int(row.Position),
		IsBlocked: row.IsBlocked,
		Labels:    ParseLabelsFromConcatenated(row.LabelIds, row.LabelNames, row.LabelColors),
	}

	if row.TypeDescription.Valid {
		summary.TypeDescription = row.TypeDescription.String
	}
	if row.PriorityDescription.Valid {
		summary.PriorityDescription = row.PriorityDescription.String
	}
	if row.PriorityColor.Valid {
		summary.PriorityColor = row.PriorityColor.String
	}
	if row.AssigneeID.Valid {
		id := int(row.AssigneeID.Int64)
		summary.AssigneeID = &id
	}
	if row.AssigneeName.Valid {
		name := row.AssigneeName.String
		summary.AssigneeName = &name
	}
	if row.Estimate.Valid {
		estimate := row.Estimate.String
		summary.Estimate = &estimate
	}
	if row.DueDate.Valid {
		summary.DueDate = &row.DueDate.Time
	}

	return summary
}

// ReadyTaskSummaryFromRowToModel converts a ready task summary row to models.TaskSummary
func ReadyTaskSummaryFromRowToModel(row types.GetReadyTaskSummariesByProjectRow) *models.TaskSummary {
	summary := &models.TaskSummary{
		ID:        int(row.ID),
		Title:     row.Title,
		ColumnID:  int(row.ColumnID),
		Position:  int(row.Position),
		IsBlocked: row.IsBlocked,
		Labels:    ParseLabelsFromConcatenated(row.LabelIds, row.LabelNames, row.LabelColors),
	}

	if row.TypeDescription.Valid {
		summary.TypeDescription = row.TypeDescription.String
	}
	if row.PriorityDescription.Valid {
		summary.PriorityDescription = row.PriorityDescription.String
	}
	if row.PriorityColor.Valid {
		summary.PriorityColor = row.PriorityColor.String
	}
	if row.AssigneeID.Valid {
		id := int(row.AssigneeID.Int64)
		summary.AssigneeID = &id
	}
	if row.AssigneeName.Valid {
		name := row.AssigneeName.String
		summary.AssigneeName = &name
	}
	if row.Estimate.Valid {
		estimate := row.Estimate.String
		summary.Estimate = &estimate
	}
	if row.DueDate.Valid {
		summary.DueDate = &row.DueDate.Time
	}

	return summary
}

// FilteredTaskSummaryFromRowToModel converts a filtered task summary row to models.TaskSummary
func FilteredTaskSummaryFromRowToModel(row types.GetTaskSummariesByProjectFilteredRow) *models.TaskSummary {
	summary := &models.TaskSummary{
		ID:        int(row.ID),
		Title:     row.Title,
		ColumnID:  int(row.ColumnID),
		Position:  int(row.Position),
		IsBlocked: row.IsBlocked,
		Labels:    ParseLabelsFromConcatenated(row.LabelIds, row.LabelNames, row.LabelColors),
	}

	if row.TypeDescription.Valid {
		summary.TypeDescription = row.TypeDescription.String
	}
	if row.PriorityDescription.Valid {
		summary.PriorityDescription = row.PriorityDescription.String
	}
	if row.PriorityColor.Valid {
		summary.PriorityColor = row.PriorityColor.String
	}
	if row.AssigneeID.Valid {
		id := int(row.AssigneeID.Int64)
		summary.AssigneeID = &id
	}
	if row.AssigneeName.Valid {
		name := row.AssigneeName.String
		summary.AssigneeName = &name
	}
	if row.Estimate.Valid {
		estimate := row.Estimate.String
		summary.Estimate = &estimate
	}
	if row.DueDate.Valid {
		summary.DueDate = &row.DueDate.Time
	}

	return summary
}

// WithFiltersTaskSummaryFromRowToModel converts a WithFilters task summary row to models.TaskSummary
func WithFiltersTaskSummaryFromRowToModel(row types.GetTaskSummariesWithFiltersRow) *models.TaskSummary {
	summary := &models.TaskSummary{
		ID:        int(row.ID),
		Title:     row.Title,
		ColumnID:  int(row.ColumnID),
		Position:  int(row.Position),
		IsBlocked: row.IsBlocked,
		Labels:    ParseLabelsFromConcatenated(row.LabelIds, row.LabelNames, row.LabelColors),
	}

	if row.TypeDescription.Valid {
		summary.TypeDescription = row.TypeDescription.String
	}
	if row.PriorityDescription.Valid {
		summary.PriorityDescription = row.PriorityDescription.String
	}
	if row.PriorityColor.Valid {
		summary.PriorityColor = row.PriorityColor.String
	}
	if row.AssigneeID.Valid {
		id := int(row.AssigneeID.Int64)
		summary.AssigneeID = &id
	}
	if row.AssigneeName.Valid {
		name := row.AssigneeName.String
		summary.AssigneeName = &name
	}
	if row.Estimate.Valid {
		estimate := row.Estimate.String
		summary.Estimate = &estimate
	}
	if row.DueDate.Valid {
		summary.DueDate = &row.DueDate.Time
	}

	return summary
}

// ParseLabelsFromConcatenated parses GROUP_CONCAT label data into Label slice
// ParseLabelsFromConcatenated parses GROUP_CONCAT label data from SQL queries.
//
// SQL queries use GROUP_CONCAT to return multiple labels in a single row:
//
//	SELECT
//	  GROUP_CONCAT(l.id, CHAR(31)) as label_ids,
//	  GROUP_CONCAT(l.name, CHAR(31)) as label_names,
//	  GROUP_CONCAT(l.color, CHAR(31)) as label_colors
//
// This function splits the concatenated strings and reconstructs the label objects.
//
// Returns empty slice if:
// - Any input string is empty
// - The number of IDs, names, and colors don't match (data integrity issue)
//
// Example:
//
//	ids    = "1\x1f2\x1f3"
//	names  = "bug\x1ffeature\x1furgent"
//	colors = "#FF0000\x1f#00FF00\x1f#0000FF"
//	result = []*models.Label{{ID:1, Name:"bug", Color:"#FF0000"}, ...}
func ParseLabelsFromConcatenated(ids, names, colors string) []*models.Label {
	if ids == "" || names == "" || colors == "" {
		return []*models.Label{}
	}

	// Use pre-defined separator constant instead of allocating new string each time
	// This optimization reduces allocations when parsing concatenated label fields
	idParts := strings.Split(ids, labelSeparator)
	nameParts := strings.Split(names, labelSeparator)
	colorParts := strings.Split(colors, labelSeparator)

	if len(idParts) != len(nameParts) || len(idParts) != len(colorParts) {
		return []*models.Label{}
	}

	// Pre-allocate labels slice with exact capacity to avoid reallocation
	labels := make([]*models.Label, 0, len(idParts))
	for i := range idParts {
		// Parse ID
		var id int
		_, _ = fmt.Sscanf(idParts[i], "%d", &id)

		labels = append(labels, &models.Label{
			ID:    id,
			Name:  nameParts[i],
			Color: colorParts[i],
		})
	}

	return labels
}
