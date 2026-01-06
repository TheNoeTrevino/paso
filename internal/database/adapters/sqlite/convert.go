package sqlite

import (
	"database/sql"

	"github.com/thenoetrevino/paso/internal/database/generated_sqlite"
	"github.com/thenoetrevino/paso/internal/database/types"
)

// ============================================================================
// FROM generated_sqlite TO types (for return values)
// ============================================================================

func fromGeneratedTask(g generated_sqlite.Task) types.Task {
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

func fromGeneratedProject(g generated_sqlite.Project) types.Project {
	return types.Project{
		ID:          g.ID,
		Name:        g.Name,
		Description: types.FromSQLNullString(g.Description),
		CreatedAt:   types.FromSQLNullTime(g.CreatedAt),
		UpdatedAt:   types.FromSQLNullTime(g.UpdatedAt),
	}
}

func fromGeneratedColumn(g generated_sqlite.Column) types.Column {
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

func fromGeneratedLabel(g generated_sqlite.Label) types.Label {
	return types.Label{
		ID:        g.ID,
		Name:      g.Name,
		Color:     g.Color,
		ProjectID: g.ProjectID,
	}
}

func fromGeneratedTaskComment(g generated_sqlite.TaskComment) types.TaskComment {
	return types.TaskComment{
		ID:        g.ID,
		TaskID:    g.TaskID,
		Content:   g.Content,
		Author:    g.Author,
		CreatedAt: types.FromSQLNullTime(g.CreatedAt),
		UpdatedAt: types.FromSQLNullTime(g.UpdatedAt),
	}
}

func fromGeneratedPriority(g generated_sqlite.Priority) types.Priority {
	return types.Priority{
		ID:          g.ID,
		Description: g.Description,
		Color:       g.Color,
	}
}

func fromGeneratedRelationType(g generated_sqlite.RelationType) types.RelationType {
	return types.RelationType{
		ID:         g.ID,
		PToCLabel:  g.PToCLabel,
		CToPLabel:  g.CToPLabel,
		Color:      g.Color,
		IsBlocking: g.IsBlocking,
	}
}

func fromGeneratedType(g generated_sqlite.Type) types.Type {
	return types.Type{
		ID:          g.ID,
		Description: g.Description,
	}
}

// ============================================================================
// FROM generated_sqlite Row types TO types Row types (for query results)
// ============================================================================

func fromGeneratedGetChildTasksRow(g generated_sqlite.GetChildTasksRow) types.GetChildTasksRow {
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

func fromGeneratedGetInProgressTaskDetailsRow(g generated_sqlite.GetInProgressTaskDetailsRow) types.GetInProgressTaskDetailsRow {
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
		IsBlocked:           g.IsBlocked != 0,
	}
}

func fromGeneratedGetInProgressTasksByProjectRow(g generated_sqlite.GetInProgressTasksByProjectRow) types.GetInProgressTasksByProjectRow {
	return types.GetInProgressTasksByProjectRow{
		ID:           g.ID,
		TicketNumber: types.FromSQLNullInt64(g.TicketNumber),
		Title:        g.Title,
		Description:  types.FromSQLNullString(g.Description),
		ColumnName:   g.ColumnName,
		ProjectName:  g.ProjectName,
	}
}

func fromGeneratedGetParentTasksRow(g generated_sqlite.GetParentTasksRow) types.GetParentTasksRow {
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

func fromGeneratedGetReadyTaskSummariesByProjectRow(g generated_sqlite.GetReadyTaskSummariesByProjectRow) types.GetReadyTaskSummariesByProjectRow {
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
		IsBlocked:           g.IsBlocked != 0,
	}
}

func fromGeneratedGetTaskRow(g generated_sqlite.GetTaskRow) types.GetTaskRow {
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

func fromGeneratedGetTaskAboveRow(g generated_sqlite.GetTaskAboveRow) types.GetTaskAboveRow {
	return types.GetTaskAboveRow{
		ID:       g.ID,
		Position: g.Position,
	}
}

func fromGeneratedGetTaskBelowRow(g generated_sqlite.GetTaskBelowRow) types.GetTaskBelowRow {
	return types.GetTaskBelowRow{
		ID:       g.ID,
		Position: g.Position,
	}
}

func fromGeneratedGetTaskDetailRow(g generated_sqlite.GetTaskDetailRow) types.GetTaskDetailRow {
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
		IsBlocked:           g.IsBlocked != 0,
	}
}

func fromGeneratedGetTaskPositionRow(g generated_sqlite.GetTaskPositionRow) types.GetTaskPositionRow {
	return types.GetTaskPositionRow{
		ColumnID: g.ColumnID,
		Position: g.Position,
	}
}

func fromGeneratedGetTaskReferencesForProjectRow(g generated_sqlite.GetTaskReferencesForProjectRow) types.GetTaskReferencesForProjectRow {
	return types.GetTaskReferencesForProjectRow{
		ID:           g.ID,
		TicketNumber: types.FromSQLNullInt64(g.TicketNumber),
		Title:        g.Title,
		Name:         g.Name,
	}
}

func fromGeneratedGetTaskRelationsForProjectRow(g generated_sqlite.GetTaskRelationsForProjectRow) types.GetTaskRelationsForProjectRow {
	return types.GetTaskRelationsForProjectRow{
		ParentID:      g.ParentID,
		ChildID:       g.ChildID,
		RelationLabel: g.RelationLabel,
		RelationColor: g.RelationColor,
		IsBlocking:    g.IsBlocking,
	}
}

func fromGeneratedGetTaskSummariesByColumnRow(g generated_sqlite.GetTaskSummariesByColumnRow) types.GetTaskSummariesByColumnRow {
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

func fromGeneratedGetTaskSummariesByProjectRow(g generated_sqlite.GetTaskSummariesByProjectRow) types.GetTaskSummariesByProjectRow {
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
		IsBlocked:           g.IsBlocked != 0,
	}
}

func fromGeneratedGetTaskSummariesByProjectFilteredRow(g generated_sqlite.GetTaskSummariesByProjectFilteredRow) types.GetTaskSummariesByProjectFilteredRow {
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
		IsBlocked:           g.IsBlocked != 0,
	}
}

func fromGeneratedGetTasksByColumnRow(g generated_sqlite.GetTasksByColumnRow) types.GetTasksByColumnRow {
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

func fromGeneratedGetTasksForTreeRow(g generated_sqlite.GetTasksForTreeRow) types.GetTasksForTreeRow {
	return types.GetTasksForTreeRow{
		ID:           g.ID,
		TicketNumber: types.FromSQLNullInt64(g.TicketNumber),
		Title:        g.Title,
		ColumnName:   g.ColumnName,
		ProjectName:  g.ProjectName,
	}
}

func fromGeneratedGetColumnByIDRow(g generated_sqlite.GetColumnByIDRow) types.GetColumnByIDRow {
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

func fromGeneratedGetColumnLinkedListInfoRow(g generated_sqlite.GetColumnLinkedListInfoRow) types.GetColumnLinkedListInfoRow {
	return types.GetColumnLinkedListInfoRow{
		PrevID:    types.NullInt64FromInterface(g.PrevID),
		NextID:    types.NullInt64FromInterface(g.NextID),
		ProjectID: g.ProjectID,
	}
}

func fromGeneratedGetColumnsByProjectRow(g generated_sqlite.GetColumnsByProjectRow) types.GetColumnsByProjectRow {
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

func fromGeneratedGetCompletedColumnByProjectRow(g generated_sqlite.GetCompletedColumnByProjectRow) types.GetCompletedColumnByProjectRow {
	return types.GetCompletedColumnByProjectRow{
		ID:                  g.ID,
		Name:                g.Name,
		ProjectID:           g.ProjectID,
		PrevID:              types.NullInt64FromInterface(g.PrevID),
		NextID:              types.NullInt64FromInterface(g.NextID),
		HoldsCompletedTasks: g.HoldsCompletedTasks,
	}
}

func fromGeneratedGetInProgressColumnByProjectRow(g generated_sqlite.GetInProgressColumnByProjectRow) types.GetInProgressColumnByProjectRow {
	return types.GetInProgressColumnByProjectRow{
		ID:                   g.ID,
		Name:                 g.Name,
		ProjectID:            g.ProjectID,
		PrevID:               types.NullInt64FromInterface(g.PrevID),
		NextID:               types.NullInt64FromInterface(g.NextID),
		HoldsInProgressTasks: g.HoldsInProgressTasks,
	}
}

func fromGeneratedGetReadyColumnByProjectRow(g generated_sqlite.GetReadyColumnByProjectRow) types.GetReadyColumnByProjectRow {
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
// TO generated_sqlite FROM types (for parameters)
// ============================================================================

func toGeneratedCreateTaskParams(t types.CreateTaskParams) generated_sqlite.CreateTaskParams {
	return generated_sqlite.CreateTaskParams{
		Title:        t.Title,
		Description:  t.Description.ToSQLNullString(),
		ColumnID:     t.ColumnID,
		Position:     t.Position,
		TicketNumber: t.TicketNumber.ToSQLNullInt64(),
	}
}

func toGeneratedAddSubtaskParams(t types.AddSubtaskParams) generated_sqlite.AddSubtaskParams {
	return generated_sqlite.AddSubtaskParams{
		ParentID: t.ParentID,
		ChildID:  t.ChildID,
	}
}

func toGeneratedAddSubtaskWithRelationTypeParams(t types.AddSubtaskWithRelationTypeParams) generated_sqlite.AddSubtaskWithRelationTypeParams {
	return generated_sqlite.AddSubtaskWithRelationTypeParams{
		ParentID:       t.ParentID,
		ChildID:        t.ChildID,
		RelationTypeID: t.RelationTypeID,
	}
}

func toGeneratedGetTaskAboveParams(t types.GetTaskAboveParams) generated_sqlite.GetTaskAboveParams {
	return generated_sqlite.GetTaskAboveParams{
		ColumnID: t.ColumnID,
		Position: t.Position,
	}
}

func toGeneratedGetTaskBelowParams(t types.GetTaskBelowParams) generated_sqlite.GetTaskBelowParams {
	return generated_sqlite.GetTaskBelowParams{
		ColumnID: t.ColumnID,
		Position: t.Position,
	}
}

func toGeneratedGetTaskSummariesByProjectFilteredParams(t types.GetTaskSummariesByProjectFilteredParams) generated_sqlite.GetTaskSummariesByProjectFilteredParams {
	return generated_sqlite.GetTaskSummariesByProjectFilteredParams{
		ProjectID: t.ProjectID,
		Title:     t.Title,
	}
}

func toGeneratedMoveTaskToColumnParams(t types.MoveTaskToColumnParams) generated_sqlite.MoveTaskToColumnParams {
	return generated_sqlite.MoveTaskToColumnParams{
		ColumnID: t.ColumnID,
		Position: t.Position,
		ID:       t.ID,
	}
}

func toGeneratedRemoveSubtaskParams(t types.RemoveSubtaskParams) generated_sqlite.RemoveSubtaskParams {
	return generated_sqlite.RemoveSubtaskParams{
		ParentID: t.ParentID,
		ChildID:  t.ChildID,
	}
}

func toGeneratedSetTaskPositionParams(t types.SetTaskPositionParams) generated_sqlite.SetTaskPositionParams {
	return generated_sqlite.SetTaskPositionParams{
		Position: t.Position,
		ID:       t.ID,
	}
}

func toGeneratedUpdateTaskParams(t types.UpdateTaskParams) generated_sqlite.UpdateTaskParams {
	return generated_sqlite.UpdateTaskParams{
		Title:       t.Title,
		Description: t.Description.ToSQLNullString(),
		ID:          t.ID,
	}
}

func toGeneratedUpdateTaskPriorityParams(t types.UpdateTaskPriorityParams) generated_sqlite.UpdateTaskPriorityParams {
	return generated_sqlite.UpdateTaskPriorityParams{
		PriorityID: t.PriorityID,
		ID:         t.ID,
	}
}

func toGeneratedUpdateTaskTypeParams(t types.UpdateTaskTypeParams) generated_sqlite.UpdateTaskTypeParams {
	return generated_sqlite.UpdateTaskTypeParams{
		TypeID: t.TypeID,
		ID:     t.ID,
	}
}

func toGeneratedCreateProjectRecordParams(t types.CreateProjectRecordParams) generated_sqlite.CreateProjectRecordParams {
	return generated_sqlite.CreateProjectRecordParams{
		Name:        t.Name,
		Description: t.Description.ToSQLNullString(),
	}
}

func toGeneratedUpdateProjectParams(t types.UpdateProjectParams) generated_sqlite.UpdateProjectParams {
	return generated_sqlite.UpdateProjectParams{
		Name:        t.Name,
		Description: t.Description.ToSQLNullString(),
		ID:          t.ID,
	}
}

func toGeneratedCreateColumnParams(t types.CreateColumnParams) generated_sqlite.CreateColumnParams {
	return generated_sqlite.CreateColumnParams{
		Name:                 t.Name,
		ProjectID:            t.ProjectID,
		PrevID:               t.PrevID.ToInterface(),
		NextID:               t.NextID.ToInterface(),
		HoldsReadyTasks:      t.HoldsReadyTasks,
		HoldsCompletedTasks:  t.HoldsCompletedTasks,
		HoldsInProgressTasks: t.HoldsInProgressTasks,
	}
}

func toGeneratedUpdateColumnHoldsCompletedTasksParams(t types.UpdateColumnHoldsCompletedTasksParams) generated_sqlite.UpdateColumnHoldsCompletedTasksParams {
	return generated_sqlite.UpdateColumnHoldsCompletedTasksParams{
		HoldsCompletedTasks: t.HoldsCompletedTasks,
		ID:                  t.ID,
	}
}

func toGeneratedUpdateColumnHoldsInProgressTasksParams(t types.UpdateColumnHoldsInProgressTasksParams) generated_sqlite.UpdateColumnHoldsInProgressTasksParams {
	return generated_sqlite.UpdateColumnHoldsInProgressTasksParams{
		HoldsInProgressTasks: t.HoldsInProgressTasks,
		ID:                   t.ID,
	}
}

func toGeneratedUpdateColumnHoldsReadyTasksParams(t types.UpdateColumnHoldsReadyTasksParams) generated_sqlite.UpdateColumnHoldsReadyTasksParams {
	return generated_sqlite.UpdateColumnHoldsReadyTasksParams{
		HoldsReadyTasks: t.HoldsReadyTasks,
		ID:              t.ID,
	}
}

func toGeneratedUpdateColumnNameParams(t types.UpdateColumnNameParams) generated_sqlite.UpdateColumnNameParams {
	return generated_sqlite.UpdateColumnNameParams{
		Name: t.Name,
		ID:   t.ID,
	}
}

func toGeneratedUpdateColumnNextIDParams(t types.UpdateColumnNextIDParams) generated_sqlite.UpdateColumnNextIDParams {
	return generated_sqlite.UpdateColumnNextIDParams{
		NextID: t.NextID.ToInterface(),
		ID:     t.ID,
	}
}

func toGeneratedUpdateColumnPrevIDParams(t types.UpdateColumnPrevIDParams) generated_sqlite.UpdateColumnPrevIDParams {
	return generated_sqlite.UpdateColumnPrevIDParams{
		PrevID: t.PrevID.ToInterface(),
		ID:     t.ID,
	}
}

func toGeneratedCreateLabelParams(t types.CreateLabelParams) generated_sqlite.CreateLabelParams {
	return generated_sqlite.CreateLabelParams{
		Name:      t.Name,
		Color:     t.Color,
		ProjectID: t.ProjectID,
	}
}

func toGeneratedAddLabelToTaskParams(t types.AddLabelToTaskParams) generated_sqlite.AddLabelToTaskParams {
	return generated_sqlite.AddLabelToTaskParams{
		TaskID:  t.TaskID,
		LabelID: t.LabelID,
	}
}

func toGeneratedInsertTaskLabelParams(t types.InsertTaskLabelParams) generated_sqlite.InsertTaskLabelParams {
	return generated_sqlite.InsertTaskLabelParams{
		TaskID:  t.TaskID,
		LabelID: t.LabelID,
	}
}

func toGeneratedRemoveLabelFromTaskParams(t types.RemoveLabelFromTaskParams) generated_sqlite.RemoveLabelFromTaskParams {
	return generated_sqlite.RemoveLabelFromTaskParams{
		TaskID:  t.TaskID,
		LabelID: t.LabelID,
	}
}

func toGeneratedUpdateLabelParams(t types.UpdateLabelParams) generated_sqlite.UpdateLabelParams {
	return generated_sqlite.UpdateLabelParams{
		Name:  t.Name,
		Color: t.Color,
		ID:    t.ID,
	}
}

func toGeneratedCreateCommentParams(t types.CreateCommentParams) generated_sqlite.CreateCommentParams {
	return generated_sqlite.CreateCommentParams{
		TaskID:  t.TaskID,
		Content: t.Content,
		Author:  t.Author,
	}
}

func toGeneratedUpdateCommentParams(t types.UpdateCommentParams) generated_sqlite.UpdateCommentParams {
	return generated_sqlite.UpdateCommentParams{
		Content: t.Content,
		ID:      t.ID,
	}
}

// Suppress unused import warning
var _ = sql.ErrNoRows
