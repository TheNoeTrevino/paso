package postgres

import (
	"github.com/thenoetrevino/paso/internal/database/generated_postgres"
	"github.com/thenoetrevino/paso/internal/database/types"
)

func toGeneratedCreateTaskParams(t types.CreateTaskParams) generated_postgres.CreateTaskParams {
	return generated_postgres.CreateTaskParams{
		Title:        t.Title,
		Description:  t.Description.ToSQLNullString(),
		ColumnID:     t.ColumnID,
		Position:     t.Position,
		TicketNumber: t.TicketNumber.ToSQLNullInt64(),
		AssigneeID:   t.AssigneeID.ToSQLNullInt32(),
		Estimate:     t.Estimate.ToSQLNullString(),
		DueDate:      t.DueDate.ToSQLNullTime(),
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

func toGeneratedGetTaskSummariesWithFiltersParams(t types.GetTaskSummariesWithFiltersParams) generated_postgres.GetTaskSummariesWithFiltersParams {
	return generated_postgres.GetTaskSummariesWithFiltersParams{
		ProjectID:   t.ProjectID,
		TitleFilter: t.TitleFilter.ToSQLNullString(),
		PriorityID:  t.PriorityID.ToSQLNullInt64(),
		TypeID:      t.TypeID.ToSQLNullInt64(),
		AssigneeID:  t.AssigneeID.ToSQLNullInt64(),
		LabelIdsCsv: t.LabelIdsCsv,
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

func toGeneratedUpdateTaskAssigneeParams(t types.UpdateTaskAssigneeParams) generated_postgres.UpdateTaskAssigneeParams {
	return generated_postgres.UpdateTaskAssigneeParams{
		AssigneeID: t.AssigneeID.ToSQLNullInt32(),
		ID:         t.ID,
	}
}

func toGeneratedUpdateTaskEstimateParams(t types.UpdateTaskEstimateParams) generated_postgres.UpdateTaskEstimateParams {
	return generated_postgres.UpdateTaskEstimateParams{
		Estimate: t.Estimate.ToSQLNullString(),
		ID:       t.ID,
	}
}

func toGeneratedUpdateTaskDueDateParams(t types.UpdateTaskDueDateParams) generated_postgres.UpdateTaskDueDateParams {
	return generated_postgres.UpdateTaskDueDateParams{
		DueDate: t.DueDate.ToSQLNullTime(),
		ID:      t.ID,
	}
}

func toGeneratedCreateProjectRecordParams(t types.CreateProjectRecordParams) generated_postgres.CreateProjectRecordParams {
	return generated_postgres.CreateProjectRecordParams{
		Name:        t.Name,
		Description: t.Description.ToSQLNullString(),
		GitBranch:   t.GitBranch.ToSQLNullString(),
	}
}

func toGeneratedUpdateProjectParams(t types.UpdateProjectParams) generated_postgres.UpdateProjectParams {
	return generated_postgres.UpdateProjectParams{
		Name:        t.Name,
		Description: t.Description.ToSQLNullString(),
		GitBranch:   t.GitBranch.ToSQLNullString(),
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

func toGeneratedUpsertLabelParams(t types.UpsertLabelParams) generated_postgres.UpsertLabelParams {
	return generated_postgres.UpsertLabelParams{
		Name:      t.Name,
		Color:     t.Color,
		ProjectID: t.ProjectID,
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
