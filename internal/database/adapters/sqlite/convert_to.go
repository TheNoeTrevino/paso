package sqlite

import (
	"database/sql"

	"github.com/thenoetrevino/paso/internal/database/generated_sqlite"
	"github.com/thenoetrevino/paso/internal/database/types"
)

func toGeneratedCreateTaskParams(t types.CreateTaskParams) generated_sqlite.CreateTaskParams {
	return generated_sqlite.CreateTaskParams{
		Title:        t.Title,
		Description:  t.Description.ToSQLNullString(),
		ColumnID:     t.ColumnID,
		Position:     t.Position,
		TicketNumber: t.TicketNumber.ToSQLNullInt64(),
		AssigneeID:   t.AssigneeID.ToInterface(),
		Estimate:     t.Estimate.ToInterface(),
		DueDate:      t.DueDate.ToInterface(),
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

func toGeneratedGetTaskSummariesWithFiltersParams(t types.GetTaskSummariesWithFiltersParams) generated_sqlite.GetTaskSummariesWithFiltersParams {
	var showArchived int64
	if t.ShowArchived {
		showArchived = int64(1)
	} else {
		showArchived = int64(0)
	}

	return generated_sqlite.GetTaskSummariesWithFiltersParams{
		ProjectID:    t.ProjectID,
		TitleFilter:  t.TitleFilter.ToInterface(),
		PriorityID:   t.PriorityID.ToInterface(),
		TypeID:       t.TypeID.ToInterface(),
		AssigneeID:   t.AssigneeID.ToInterface(),
		LabelIdsCsv:  t.LabelIdsCsv,
		ShowArchived: showArchived,
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

func toGeneratedUpdateTaskAssigneeParams(t types.UpdateTaskAssigneeParams) generated_sqlite.UpdateTaskAssigneeParams {
	return generated_sqlite.UpdateTaskAssigneeParams{
		AssigneeID: t.AssigneeID.ToInterface(),
		ID:         t.ID,
	}
}

func toGeneratedUpdateTaskEstimateParams(t types.UpdateTaskEstimateParams) generated_sqlite.UpdateTaskEstimateParams {
	return generated_sqlite.UpdateTaskEstimateParams{
		Estimate: t.Estimate.ToSQLNullString(),
		ID:       t.ID,
	}
}

func toGeneratedUpdateTaskDueDateParams(t types.UpdateTaskDueDateParams) generated_sqlite.UpdateTaskDueDateParams {
	return generated_sqlite.UpdateTaskDueDateParams{
		DueDate: t.DueDate.ToInterface(),
		ID:      t.ID,
	}
}

func toGeneratedCreateProjectRecordParams(t types.CreateProjectRecordParams) generated_sqlite.CreateProjectRecordParams {
	return generated_sqlite.CreateProjectRecordParams{
		Name:        t.Name,
		Description: t.Description.ToSQLNullString(),
		GitBranch:   t.GitBranch.ToSQLNullString(),
	}
}

func toGeneratedUpdateProjectParams(t types.UpdateProjectParams) generated_sqlite.UpdateProjectParams {
	return generated_sqlite.UpdateProjectParams{
		Name:        t.Name,
		Description: t.Description.ToSQLNullString(),
		GitBranch:   t.GitBranch.ToSQLNullString(),
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

func toGeneratedUpsertLabelParams(t types.UpsertLabelParams) generated_sqlite.UpsertLabelParams {
	return generated_sqlite.UpsertLabelParams{
		Name:      t.Name,
		Color:     t.Color,
		ProjectID: t.ProjectID,
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

func toGeneratedCreateTaskEventParams(t types.CreateTaskEventParams) generated_sqlite.CreateTaskEventParams {
	return generated_sqlite.CreateTaskEventParams{
		TaskID:  t.TaskID,
		Content: t.Content,
		Author:  t.Author,
	}
}

func toGeneratedCreateStandupLogParams(t types.CreateStandupLogParams) generated_sqlite.CreateStandupLogParams {
	return generated_sqlite.CreateStandupLogParams{
		ProjectID: t.ProjectID,
		Content:   t.Content,
	}
}

// Suppress unused import warning
var _ = sql.ErrNoRows
