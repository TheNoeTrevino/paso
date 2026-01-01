package converters

import (
	"fmt"
	"strings"

	"github.com/thenoetrevino/paso/internal/database/generated"
	"github.com/thenoetrevino/paso/internal/models"
)

// TaskToModel converts a generated.Task to models.Task
func TaskToModel(t generated.Task) *models.Task {
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

	return task
}

// ParentTasksToReferences converts parent task rows to TaskReference slice
func ParentTasksToReferences(rows []generated.GetParentTasksRow) []*models.TaskReference {
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
func ChildTasksToReferences(rows []generated.GetChildTasksRow) []*models.TaskReference {
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

// CommentsToModels converts generated.TaskComment slice to models.Comment slice
func CommentsToModels(comments []generated.TaskComment) []*models.Comment {
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
func TaskSummaryFromRowToModel(row generated.GetTaskSummariesByProjectRow) *models.TaskSummary {
	summary := &models.TaskSummary{
		ID:        int(row.ID),
		Title:     row.Title,
		ColumnID:  int(row.ColumnID),
		Position:  int(row.Position),
		IsBlocked: row.IsBlocked > 0,
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

	return summary
}

// ReadyTaskSummaryFromRowToModel converts a ready task summary row to models.TaskSummary
func ReadyTaskSummaryFromRowToModel(row generated.GetReadyTaskSummariesByProjectRow) *models.TaskSummary {
	summary := &models.TaskSummary{
		ID:        int(row.ID),
		Title:     row.Title,
		ColumnID:  int(row.ColumnID),
		Position:  int(row.Position),
		IsBlocked: row.IsBlocked > 0,
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

	return summary
}

// FilteredTaskSummaryFromRowToModel converts a filtered task summary row to models.TaskSummary
func FilteredTaskSummaryFromRowToModel(row generated.GetTaskSummariesByProjectFilteredRow) *models.TaskSummary {
	summary := &models.TaskSummary{
		ID:        int(row.ID),
		Title:     row.Title,
		ColumnID:  int(row.ColumnID),
		Position:  int(row.Position),
		IsBlocked: row.IsBlocked > 0,
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

	return summary
}

// ParseLabelsFromConcatenated parses GROUP_CONCAT label data into Label slice
func ParseLabelsFromConcatenated(ids, names, colors string) []*models.Label {
	if ids == "" || names == "" || colors == "" {
		return []*models.Label{}
	}

	idParts := strings.Split(ids, string(rune(31))) // CHAR(31) separator
	nameParts := strings.Split(names, string(rune(31)))
	colorParts := strings.Split(colors, string(rune(31)))

	if len(idParts) != len(nameParts) || len(idParts) != len(colorParts) {
		return []*models.Label{}
	}

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
