package postgres

import (
	"database/sql"

	"github.com/thenoetrevino/paso/internal/database/generated_postgres"
	"github.com/thenoetrevino/paso/internal/database/types"
)

// ============================================================================
// FROM generated_postgres TO types (for return values)
// ============================================================================

func fromGeneratedTask(g generated_postgres.Task) types.Task {
	return types.Task{
		ID:           g.ID,
		Title:        g.Title,
		Description:  types.FromSQLNullString(g.Description),
		ColumnID:     g.ColumnID,
		Position:     g.Position,
		TicketNumber: types.FromSQLNullInt64(g.TicketNumber),
		TypeID:       g.TypeID,
		PriorityID:   g.PriorityID,
		CreatedAt:    types.FromSQLNullTime(g.CreatedAt),
		UpdatedAt:    types.FromSQLNullTime(g.UpdatedAt),
	}
}

func fromGeneratedProject(g generated_postgres.Project) types.Project {
	return types.Project{
		ID:          g.ID,
		Name:        g.Name,
		Description: types.FromSQLNullString(g.Description),
		CreatedAt:   types.FromSQLNullTime(g.CreatedAt),
		UpdatedAt:   types.FromSQLNullTime(g.UpdatedAt),
	}
}

func fromGeneratedColumn(g generated_postgres.Column) types.Column {
	return types.Column{
		ID:                   g.ID,
		Name:                 g.Name,
		PrevID:               types.NullInt64FromInterface(g.PrevID),
		NextID:               types.NullInt64FromInterface(g.NextID),
		ProjectID:            g.ProjectID,
		HoldsReadyTasks:      g.HoldsReadyTasks,
		HoldsCompletedTasks:  g.HoldsCompletedTasks,
		HoldsInProgressTasks: g.HoldsInProgressTasks,
	}
}

func fromGeneratedLabel(g generated_postgres.Label) types.Label {
	return types.Label{
		ID:        g.ID,
		Name:      g.Name,
		Color:     g.Color,
		ProjectID: g.ProjectID,
	}
}

func fromGeneratedTaskComment(g generated_postgres.TaskComment) types.TaskComment {
	return types.TaskComment{
		ID:        g.ID,
		TaskID:    g.TaskID,
		Content:   g.Content,
		Author:    g.Author,
		CreatedAt: types.FromSQLNullTime(g.CreatedAt),
		UpdatedAt: types.FromSQLNullTime(g.UpdatedAt),
	}
}

func fromGeneratedPriority(g generated_postgres.Priority) types.Priority {
	return types.Priority{
		ID:          g.ID,
		Description: g.Description,
		Color:       g.Color,
	}
}

func fromGeneratedRelationType(g generated_postgres.RelationType) types.RelationType {
	return types.RelationType{
		ID:         g.ID,
		PToCLabel:  g.PToCLabel,
		CToPLabel:  g.CToPLabel,
		Color:      g.Color,
		IsBlocking: g.IsBlocking,
	}
}

func fromGeneratedType(g generated_postgres.Type) types.Type {
	return types.Type{
		ID:          g.ID,
		Description: g.Description,
	}
}

// ============================================================================
// FROM generated_postgres Row types TO types Row types (for query results)
// ============================================================================

func fromGeneratedGetChildTasksRow(g generated_postgres.GetChildTasksRow) types.GetChildTasksRow {
	return types.GetChildTasksRow{
		ID:           g.ID,
		TicketNumber: types.FromSQLNullInt64(g.TicketNumber),
		Title:        g.Title,
		Name:         g.Name,
		ID_2:         g.ID_2,
		CToPLabel:    g.CToPLabel,
		Color:        g.Color,
		IsBlocking:   g.IsBlocking,
	}
}

func fromGeneratedGetInProgressTaskDetailsRow(g generated_postgres.GetInProgressTaskDetailsRow) types.GetInProgressTaskDetailsRow {
	return types.GetInProgressTaskDetailsRow{
		ID:                  g.ID,
		TicketNumber:        types.FromSQLNullInt64(g.TicketNumber),
		Title:               g.Title,
		Description:         types.FromSQLNullString(g.Description),
		ColumnID:            g.ColumnID,
		Position:            g.Position,
		CreatedAt:           types.FromSQLNullTime(g.CreatedAt),
		UpdatedAt:           types.FromSQLNullTime(g.UpdatedAt),
		ColumnName:          g.ColumnName,
		ProjectName:         g.ProjectName,
		TypeDescription:     types.FromSQLNullString(g.TypeDescription),
		PriorityDescription: types.FromSQLNullString(g.PriorityDescription),
		PriorityColor:       types.FromSQLNullString(g.PriorityColor),
		LabelIds:            g.LabelIds,
		LabelNames:          g.LabelNames,
		LabelColors:         g.LabelColors,
		IsBlocked:           g.IsBlocked,
	}
}

func fromGeneratedGetInProgressTasksByProjectRow(g generated_postgres.GetInProgressTasksByProjectRow) types.GetInProgressTasksByProjectRow {
	return types.GetInProgressTasksByProjectRow{
		ID:           g.ID,
		TicketNumber: types.FromSQLNullInt64(g.TicketNumber),
		Title:        g.Title,
		Description:  types.FromSQLNullString(g.Description),
		ColumnName:   g.ColumnName,
		ProjectName:  g.ProjectName,
	}
}

func fromGeneratedGetParentTasksRow(g generated_postgres.GetParentTasksRow) types.GetParentTasksRow {
	return types.GetParentTasksRow{
		ID:           g.ID,
		TicketNumber: types.FromSQLNullInt64(g.TicketNumber),
		Title:        g.Title,
		Name:         g.Name,
		ID_2:         g.ID_2,
		PToCLabel:    g.PToCLabel,
		Color:        g.Color,
		IsBlocking:   g.IsBlocking,
	}
}

func fromGeneratedGetReadyTaskSummariesByProjectRow(g generated_postgres.GetReadyTaskSummariesByProjectRow) types.GetReadyTaskSummariesByProjectRow {
	return types.GetReadyTaskSummariesByProjectRow{
		ID:                  g.ID,
		Title:               g.Title,
		ColumnID:            g.ColumnID,
		Position:            g.Position,
		TypeDescription:     types.FromSQLNullString(g.TypeDescription),
		PriorityDescription: types.FromSQLNullString(g.PriorityDescription),
		PriorityColor:       types.FromSQLNullString(g.PriorityColor),
		LabelIds:            g.LabelIds,
		LabelNames:          g.LabelNames,
		LabelColors:         g.LabelColors,
		IsBlocked:           g.IsBlocked,
	}
}

func fromGeneratedGetTaskRow(g generated_postgres.GetTaskRow) types.GetTaskRow {
	return types.GetTaskRow{
		ID:          g.ID,
		Title:       g.Title,
		Description: types.FromSQLNullString(g.Description),
		ColumnID:    g.ColumnID,
		Position:    g.Position,
		CreatedAt:   types.FromSQLNullTime(g.CreatedAt),
		UpdatedAt:   types.FromSQLNullTime(g.UpdatedAt),
	}
}

func fromGeneratedGetTaskAboveRow(g generated_postgres.GetTaskAboveRow) types.GetTaskAboveRow {
	return types.GetTaskAboveRow{
		ID:       g.ID,
		Position: g.Position,
	}
}

func fromGeneratedGetTaskBelowRow(g generated_postgres.GetTaskBelowRow) types.GetTaskBelowRow {
	return types.GetTaskBelowRow{
		ID:       g.ID,
		Position: g.Position,
	}
}

func fromGeneratedGetTaskDetailRow(g generated_postgres.GetTaskDetailRow) types.GetTaskDetailRow {
	return types.GetTaskDetailRow{
		ID:                  g.ID,
		Title:               g.Title,
		Description:         types.FromSQLNullString(g.Description),
		ColumnID:            g.ColumnID,
		Position:            g.Position,
		TicketNumber:        types.FromSQLNullInt64(g.TicketNumber),
		CreatedAt:           types.FromSQLNullTime(g.CreatedAt),
		UpdatedAt:           types.FromSQLNullTime(g.UpdatedAt),
		TypeDescription:     types.FromSQLNullString(g.TypeDescription),
		PriorityDescription: types.FromSQLNullString(g.PriorityDescription),
		PriorityColor:       types.FromSQLNullString(g.PriorityColor),
		ColumnName:          g.ColumnName,
		ProjectName:         g.ProjectName,
		IsBlocked:           g.IsBlocked,
	}
}

func fromGeneratedGetTaskPositionRow(g generated_postgres.GetTaskPositionRow) types.GetTaskPositionRow {
	return types.GetTaskPositionRow{
		ColumnID: g.ColumnID,
		Position: g.Position,
	}
}

func fromGeneratedGetTaskReferencesForProjectRow(g generated_postgres.GetTaskReferencesForProjectRow) types.GetTaskReferencesForProjectRow {
	return types.GetTaskReferencesForProjectRow{
		ID:           g.ID,
		TicketNumber: types.FromSQLNullInt64(g.TicketNumber),
		Title:        g.Title,
		Name:         g.Name,
	}
}

func fromGeneratedGetTaskRelationsForProjectRow(g generated_postgres.GetTaskRelationsForProjectRow) types.GetTaskRelationsForProjectRow {
	return types.GetTaskRelationsForProjectRow{
		ParentID:      g.ParentID,
		ChildID:       g.ChildID,
		RelationLabel: g.RelationLabel,
		RelationColor: g.RelationColor,
		IsBlocking:    g.IsBlocking,
	}
}

func fromGeneratedGetTaskSummariesByColumnRow(g generated_postgres.GetTaskSummariesByColumnRow) types.GetTaskSummariesByColumnRow {
	return types.GetTaskSummariesByColumnRow{
		ID:                  g.ID,
		Title:               g.Title,
		ColumnID:            g.ColumnID,
		Position:            g.Position,
		TypeDescription:     types.FromSQLNullString(g.TypeDescription),
		PriorityDescription: types.FromSQLNullString(g.PriorityDescription),
		PriorityColor:       types.FromSQLNullString(g.PriorityColor),
		LabelIds:            g.LabelIds,
		LabelNames:          g.LabelNames,
		LabelColors:         g.LabelColors,
	}
}

func fromGeneratedGetTaskSummariesByProjectRow(g generated_postgres.GetTaskSummariesByProjectRow) types.GetTaskSummariesByProjectRow {
	return types.GetTaskSummariesByProjectRow{
		ID:                  g.ID,
		Title:               g.Title,
		ColumnID:            g.ColumnID,
		Position:            g.Position,
		TypeDescription:     types.FromSQLNullString(g.TypeDescription),
		PriorityDescription: types.FromSQLNullString(g.PriorityDescription),
		PriorityColor:       types.FromSQLNullString(g.PriorityColor),
		LabelIds:            g.LabelIds,
		LabelNames:          g.LabelNames,
		LabelColors:         g.LabelColors,
		IsBlocked:           g.IsBlocked,
	}
}

func fromGeneratedGetTaskSummariesByProjectFilteredRow(g generated_postgres.GetTaskSummariesByProjectFilteredRow) types.GetTaskSummariesByProjectFilteredRow {
	return types.GetTaskSummariesByProjectFilteredRow{
		ID:                  g.ID,
		Title:               g.Title,
		ColumnID:            g.ColumnID,
		Position:            g.Position,
		TypeDescription:     types.FromSQLNullString(g.TypeDescription),
		PriorityDescription: types.FromSQLNullString(g.PriorityDescription),
		PriorityColor:       types.FromSQLNullString(g.PriorityColor),
		LabelIds:            g.LabelIds,
		LabelNames:          g.LabelNames,
		LabelColors:         g.LabelColors,
		IsBlocked:           g.IsBlocked,
	}
}

func fromGeneratedGetTasksByColumnRow(g generated_postgres.GetTasksByColumnRow) types.GetTasksByColumnRow {
	return types.GetTasksByColumnRow{
		ID:          g.ID,
		Title:       g.Title,
		Description: types.FromSQLNullString(g.Description),
		ColumnID:    g.ColumnID,
		Position:    g.Position,
		CreatedAt:   types.FromSQLNullTime(g.CreatedAt),
		UpdatedAt:   types.FromSQLNullTime(g.UpdatedAt),
	}
}

func fromGeneratedGetTasksForTreeRow(g generated_postgres.GetTasksForTreeRow) types.GetTasksForTreeRow {
	return types.GetTasksForTreeRow{
		ID:           g.ID,
		TicketNumber: types.FromSQLNullInt64(g.TicketNumber),
		Title:        g.Title,
		ColumnName:   g.ColumnName,
		ProjectName:  g.ProjectName,
	}
}

func fromGeneratedGetColumnByIDRow(g generated_postgres.GetColumnByIDRow) types.GetColumnByIDRow {
	return types.GetColumnByIDRow{
		ID:                   g.ID,
		Name:                 g.Name,
		ProjectID:            g.ProjectID,
		PrevID:               types.NullInt64FromInterface(g.PrevID),
		NextID:               types.NullInt64FromInterface(g.NextID),
		HoldsReadyTasks:      g.HoldsReadyTasks,
		HoldsCompletedTasks:  g.HoldsCompletedTasks,
		HoldsInProgressTasks: g.HoldsInProgressTasks,
	}
}

func fromGeneratedGetColumnLinkedListInfoRow(g generated_postgres.GetColumnLinkedListInfoRow) types.GetColumnLinkedListInfoRow {
	return types.GetColumnLinkedListInfoRow{
		PrevID:    types.NullInt64FromInterface(g.PrevID),
		NextID:    types.NullInt64FromInterface(g.NextID),
		ProjectID: g.ProjectID,
	}
}

func fromGeneratedGetColumnsByProjectRow(g generated_postgres.GetColumnsByProjectRow) types.GetColumnsByProjectRow {
	return types.GetColumnsByProjectRow{
		ID:                   g.ID,
		Name:                 g.Name,
		ProjectID:            g.ProjectID,
		PrevID:               types.NullInt64FromInterface(g.PrevID),
		NextID:               types.NullInt64FromInterface(g.NextID),
		HoldsReadyTasks:      g.HoldsReadyTasks,
		HoldsCompletedTasks:  g.HoldsCompletedTasks,
		HoldsInProgressTasks: g.HoldsInProgressTasks,
	}
}

func fromGeneratedGetCompletedColumnByProjectRow(g generated_postgres.GetCompletedColumnByProjectRow) types.GetCompletedColumnByProjectRow {
	return types.GetCompletedColumnByProjectRow{
		ID:                  g.ID,
		Name:                g.Name,
		ProjectID:           g.ProjectID,
		PrevID:              types.NullInt64FromInterface(g.PrevID),
		NextID:              types.NullInt64FromInterface(g.NextID),
		HoldsCompletedTasks: g.HoldsCompletedTasks,
	}
}

func fromGeneratedGetInProgressColumnByProjectRow(g generated_postgres.GetInProgressColumnByProjectRow) types.GetInProgressColumnByProjectRow {
	return types.GetInProgressColumnByProjectRow{
		ID:                   g.ID,
		Name:                 g.Name,
		ProjectID:            g.ProjectID,
		PrevID:               types.NullInt64FromInterface(g.PrevID),
		NextID:               types.NullInt64FromInterface(g.NextID),
		HoldsInProgressTasks: g.HoldsInProgressTasks,
	}
}

func fromGeneratedGetReadyColumnByProjectRow(g generated_postgres.GetReadyColumnByProjectRow) types.GetReadyColumnByProjectRow {
	return types.GetReadyColumnByProjectRow{
		ID:              g.ID,
		Name:            g.Name,
		ProjectID:       g.ProjectID,
		PrevID:          types.NullInt64FromInterface(g.PrevID),
		NextID:          types.NullInt64FromInterface(g.NextID),
		HoldsReadyTasks: g.HoldsReadyTasks,
	}
}

// ============================================================================
// TO generated_postgres FROM types (for parameters)
// ============================================================================

func toGeneratedCreateTaskParams(t types.CreateTaskParams) generated_postgres.CreateTaskParams {
	return generated_postgres.CreateTaskParams{
		Title:        t.Title,
		Description:  t.Description.ToSQLNullString(),
		ColumnID:     t.ColumnID,
		Position:     t.Position,
		TicketNumber: t.TicketNumber.ToSQLNullInt64(),
	}
}

func toGeneratedAddSubtaskParams(t types.AddSubtaskParams) generated_postgres.AddSubtaskParams {
	return generated_postgres.AddSubtaskParams{
		ParentID: t.ParentID,
		ChildID:  t.ChildID,
	}
}

func toGeneratedAddSubtaskWithRelationTypeParams(t types.AddSubtaskWithRelationTypeParams) generated_postgres.AddSubtaskWithRelationTypeParams {
	return generated_postgres.AddSubtaskWithRelationTypeParams{
		ParentID:       t.ParentID,
		ChildID:        t.ChildID,
		RelationTypeID: t.RelationTypeID,
	}
}

func toGeneratedGetTaskAboveParams(t types.GetTaskAboveParams) generated_postgres.GetTaskAboveParams {
	return generated_postgres.GetTaskAboveParams{
		ColumnID: t.ColumnID,
		Position: t.Position,
	}
}

func toGeneratedGetTaskBelowParams(t types.GetTaskBelowParams) generated_postgres.GetTaskBelowParams {
	return generated_postgres.GetTaskBelowParams{
		ColumnID: t.ColumnID,
		Position: t.Position,
	}
}

func toGeneratedGetTaskSummariesByProjectFilteredParams(t types.GetTaskSummariesByProjectFilteredParams) generated_postgres.GetTaskSummariesByProjectFilteredParams {
	return generated_postgres.GetTaskSummariesByProjectFilteredParams{
		ProjectID: t.ProjectID,
		Title:     t.Title,
	}
}

func toGeneratedMoveTaskToColumnParams(t types.MoveTaskToColumnParams) generated_postgres.MoveTaskToColumnParams {
	return generated_postgres.MoveTaskToColumnParams{
		ColumnID: t.ColumnID,
		Position: t.Position,
		ID:       t.ID,
	}
}

func toGeneratedRemoveSubtaskParams(t types.RemoveSubtaskParams) generated_postgres.RemoveSubtaskParams {
	return generated_postgres.RemoveSubtaskParams{
		ParentID: t.ParentID,
		ChildID:  t.ChildID,
	}
}

func toGeneratedSetTaskPositionParams(t types.SetTaskPositionParams) generated_postgres.SetTaskPositionParams {
	return generated_postgres.SetTaskPositionParams{
		Position: t.Position,
		ID:       t.ID,
	}
}

func toGeneratedUpdateTaskParams(t types.UpdateTaskParams) generated_postgres.UpdateTaskParams {
	return generated_postgres.UpdateTaskParams{
		Title:       t.Title,
		Description: t.Description.ToSQLNullString(),
		ID:          t.ID,
	}
}

func toGeneratedUpdateTaskPriorityParams(t types.UpdateTaskPriorityParams) generated_postgres.UpdateTaskPriorityParams {
	return generated_postgres.UpdateTaskPriorityParams{
		PriorityID: t.PriorityID,
		ID:         t.ID,
	}
}

func toGeneratedUpdateTaskTypeParams(t types.UpdateTaskTypeParams) generated_postgres.UpdateTaskTypeParams {
	return generated_postgres.UpdateTaskTypeParams{
		TypeID: t.TypeID,
		ID:     t.ID,
	}
}

func toGeneratedCreateProjectRecordParams(t types.CreateProjectRecordParams) generated_postgres.CreateProjectRecordParams {
	return generated_postgres.CreateProjectRecordParams{
		Name:        t.Name,
		Description: t.Description.ToSQLNullString(),
	}
}

func toGeneratedUpdateProjectParams(t types.UpdateProjectParams) generated_postgres.UpdateProjectParams {
	return generated_postgres.UpdateProjectParams{
		Name:        t.Name,
		Description: t.Description.ToSQLNullString(),
		ID:          t.ID,
	}
}

func toGeneratedCreateColumnParams(t types.CreateColumnParams) generated_postgres.CreateColumnParams {
	return generated_postgres.CreateColumnParams{
		Name:                 t.Name,
		ProjectID:            t.ProjectID,
		PrevID:               t.PrevID.ToSQLNullInt64(),
		NextID:               t.NextID.ToSQLNullInt64(),
		HoldsReadyTasks:      t.HoldsReadyTasks,
		HoldsCompletedTasks:  t.HoldsCompletedTasks,
		HoldsInProgressTasks: t.HoldsInProgressTasks,
	}
}

func toGeneratedUpdateColumnHoldsCompletedTasksParams(t types.UpdateColumnHoldsCompletedTasksParams) generated_postgres.UpdateColumnHoldsCompletedTasksParams {
	return generated_postgres.UpdateColumnHoldsCompletedTasksParams{
		HoldsCompletedTasks: t.HoldsCompletedTasks,
		ID:                  t.ID,
	}
}

func toGeneratedUpdateColumnHoldsInProgressTasksParams(t types.UpdateColumnHoldsInProgressTasksParams) generated_postgres.UpdateColumnHoldsInProgressTasksParams {
	return generated_postgres.UpdateColumnHoldsInProgressTasksParams{
		HoldsInProgressTasks: t.HoldsInProgressTasks,
		ID:                   t.ID,
	}
}

func toGeneratedUpdateColumnHoldsReadyTasksParams(t types.UpdateColumnHoldsReadyTasksParams) generated_postgres.UpdateColumnHoldsReadyTasksParams {
	return generated_postgres.UpdateColumnHoldsReadyTasksParams{
		HoldsReadyTasks: t.HoldsReadyTasks,
		ID:              t.ID,
	}
}

func toGeneratedUpdateColumnNameParams(t types.UpdateColumnNameParams) generated_postgres.UpdateColumnNameParams {
	return generated_postgres.UpdateColumnNameParams{
		Name: t.Name,
		ID:   t.ID,
	}
}

func toGeneratedUpdateColumnNextIDParams(t types.UpdateColumnNextIDParams) generated_postgres.UpdateColumnNextIDParams {
	return generated_postgres.UpdateColumnNextIDParams{
		NextID: t.NextID.ToSQLNullInt64(),
		ID:     t.ID,
	}
}

func toGeneratedUpdateColumnPrevIDParams(t types.UpdateColumnPrevIDParams) generated_postgres.UpdateColumnPrevIDParams {
	return generated_postgres.UpdateColumnPrevIDParams{
		PrevID: t.PrevID.ToSQLNullInt64(),
		ID:     t.ID,
	}
}

func toGeneratedCreateLabelParams(t types.CreateLabelParams) generated_postgres.CreateLabelParams {
	return generated_postgres.CreateLabelParams{
		Name:      t.Name,
		Color:     t.Color,
		ProjectID: t.ProjectID,
	}
}

func toGeneratedAddLabelToTaskParams(t types.AddLabelToTaskParams) generated_postgres.AddLabelToTaskParams {
	return generated_postgres.AddLabelToTaskParams{
		TaskID:  t.TaskID,
		LabelID: t.LabelID,
	}
}

func toGeneratedInsertTaskLabelParams(t types.InsertTaskLabelParams) generated_postgres.InsertTaskLabelParams {
	return generated_postgres.InsertTaskLabelParams{
		TaskID:  t.TaskID,
		LabelID: t.LabelID,
	}
}

func toGeneratedRemoveLabelFromTaskParams(t types.RemoveLabelFromTaskParams) generated_postgres.RemoveLabelFromTaskParams {
	return generated_postgres.RemoveLabelFromTaskParams{
		TaskID:  t.TaskID,
		LabelID: t.LabelID,
	}
}

func toGeneratedUpdateLabelParams(t types.UpdateLabelParams) generated_postgres.UpdateLabelParams {
	return generated_postgres.UpdateLabelParams{
		Name:  t.Name,
		Color: t.Color,
		ID:    t.ID,
	}
}

func toGeneratedCreateCommentParams(t types.CreateCommentParams) generated_postgres.CreateCommentParams {
	return generated_postgres.CreateCommentParams{
		TaskID:  t.TaskID,
		Content: t.Content,
		Author:  t.Author,
	}
}

func toGeneratedUpdateCommentParams(t types.UpdateCommentParams) generated_postgres.UpdateCommentParams {
	return generated_postgres.UpdateCommentParams{
		Content: t.Content,
		ID:      t.ID,
	}
}

// Suppress unused import warning
var _ = sql.ErrNoRows
