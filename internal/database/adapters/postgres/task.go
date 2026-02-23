package postgres

import (
	"context"

	"github.com/thenoetrevino/paso/internal/database/types"
)

func (a *Adapter) AddSubtask(ctx context.Context, arg types.AddSubtaskParams) error {
	return a.queries.AddSubtask(ctx, toGeneratedAddSubtaskParams(arg))
}

func (a *Adapter) AddSubtaskWithRelationType(ctx context.Context, arg types.AddSubtaskWithRelationTypeParams) error {
	return a.queries.AddSubtaskWithRelationType(ctx, toGeneratedAddSubtaskWithRelationTypeParams(arg))
}

func (a *Adapter) CreateTask(ctx context.Context, arg types.CreateTaskParams) (types.Task, error) {
	result, err := a.queries.CreateTask(ctx, toGeneratedCreateTaskParams(arg))
	if err != nil {
		return types.Task{}, err
	}
	return fromGeneratedTask(result), nil
}

func (a *Adapter) DeleteTask(ctx context.Context, id int64) error {
	return a.queries.DeleteTask(ctx, id)
}

func (a *Adapter) DeleteTasksByColumn(ctx context.Context, columnID int64) error {
	return a.queries.DeleteTasksByColumn(ctx, columnID)
}

func (a *Adapter) DeleteTasksByProject(ctx context.Context, projectID int64) error {
	return a.queries.DeleteTasksByProject(ctx, projectID)
}

func (a *Adapter) GetAllPriorities(ctx context.Context) ([]types.Priority, error) {
	results, err := a.queries.GetAllPriorities(ctx)
	if err != nil {
		return nil, err
	}
	return types.ConvertSlice(results, fromGeneratedPriority), nil
}

func (a *Adapter) GetAllRelationTypes(ctx context.Context) ([]types.RelationType, error) {
	results, err := a.queries.GetAllRelationTypes(ctx)
	if err != nil {
		return nil, err
	}
	return types.ConvertSlice(results, fromGeneratedRelationType), nil
}

func (a *Adapter) GetAllTypes(ctx context.Context) ([]types.Type, error) {
	results, err := a.queries.GetAllTypes(ctx)
	if err != nil {
		return nil, err
	}
	return types.ConvertSlice(results, fromGeneratedType), nil
}

func (a *Adapter) GetChildTasks(ctx context.Context, parentID int64) ([]types.GetChildTasksRow, error) {
	results, err := a.queries.GetChildTasks(ctx, parentID)
	if err != nil {
		return nil, err
	}
	return types.ConvertSlice(results, fromGeneratedGetChildTasksRow), nil
}

func (a *Adapter) GetInProgressTaskDetails(ctx context.Context, id int64) ([]types.GetInProgressTaskDetailsRow, error) {
	results, err := a.queries.GetInProgressTaskDetails(ctx, id)
	if err != nil {
		return nil, err
	}
	return types.ConvertSlice(results, fromGeneratedGetInProgressTaskDetailsRow), nil
}

func (a *Adapter) GetInProgressTasksByProject(ctx context.Context, id int64) ([]types.GetInProgressTasksByProjectRow, error) {
	results, err := a.queries.GetInProgressTasksByProject(ctx, id)
	if err != nil {
		return nil, err
	}
	return types.ConvertSlice(results, fromGeneratedGetInProgressTasksByProjectRow), nil
}

func (a *Adapter) GetParentTasks(ctx context.Context, childID int64) ([]types.GetParentTasksRow, error) {
	results, err := a.queries.GetParentTasks(ctx, childID)
	if err != nil {
		return nil, err
	}
	return types.ConvertSlice(results, fromGeneratedGetParentTasksRow), nil
}

func (a *Adapter) GetReadyTaskSummariesByProject(ctx context.Context, projectID int64) ([]types.GetReadyTaskSummariesByProjectRow, error) {
	results, err := a.queries.GetReadyTaskSummariesByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return types.ConvertSlice(results, fromGeneratedGetReadyTaskSummariesByProjectRow), nil
}

func (a *Adapter) GetTask(ctx context.Context, id int64) (types.GetTaskRow, error) {
	result, err := a.queries.GetTask(ctx, id)
	if err != nil {
		return types.GetTaskRow{}, err
	}
	return fromGeneratedGetTaskRow(result), nil
}

func (a *Adapter) GetTaskTypeAndPriorityIDs(ctx context.Context, id int64) (types.GetTaskTypeAndPriorityIDsRow, error) {
	result, err := a.queries.GetTaskTypeAndPriorityIDs(ctx, id)
	if err != nil {
		return types.GetTaskTypeAndPriorityIDsRow{}, err
	}
	return types.GetTaskTypeAndPriorityIDsRow{
		TypeID:     result.TypeID,
		PriorityID: result.PriorityID,
	}, nil
}

func (a *Adapter) GetTaskAbove(ctx context.Context, arg types.GetTaskAboveParams) (types.GetTaskAboveRow, error) {
	result, err := a.queries.GetTaskAbove(ctx, toGeneratedGetTaskAboveParams(arg))
	if err != nil {
		return types.GetTaskAboveRow{}, err
	}
	return fromGeneratedGetTaskAboveRow(result), nil
}

func (a *Adapter) GetTaskBelow(ctx context.Context, arg types.GetTaskBelowParams) (types.GetTaskBelowRow, error) {
	result, err := a.queries.GetTaskBelow(ctx, toGeneratedGetTaskBelowParams(arg))
	if err != nil {
		return types.GetTaskBelowRow{}, err
	}
	return fromGeneratedGetTaskBelowRow(result), nil
}

func (a *Adapter) GetTaskCountByColumn(ctx context.Context, columnID int64) (int64, error) {
	return a.queries.GetTaskCountByColumn(ctx, columnID)
}

func (a *Adapter) GetTaskDetail(ctx context.Context, id int64) (types.GetTaskDetailRow, error) {
	result, err := a.queries.GetTaskDetail(ctx, id)
	if err != nil {
		return types.GetTaskDetailRow{}, err
	}
	return fromGeneratedGetTaskDetailRow(result), nil
}

func (a *Adapter) GetTaskPosition(ctx context.Context, id int64) (types.GetTaskPositionRow, error) {
	result, err := a.queries.GetTaskPosition(ctx, id)
	if err != nil {
		return types.GetTaskPositionRow{}, err
	}
	return fromGeneratedGetTaskPositionRow(result), nil
}

func (a *Adapter) GetTaskReferencesForProject(ctx context.Context, id int64) ([]types.GetTaskReferencesForProjectRow, error) {
	results, err := a.queries.GetTaskReferencesForProject(ctx, id)
	if err != nil {
		return nil, err
	}
	return types.ConvertSlice(results, fromGeneratedGetTaskReferencesForProjectRow), nil
}

func (a *Adapter) GetTaskRelationsForProject(ctx context.Context, projectID int64) ([]types.GetTaskRelationsForProjectRow, error) {
	results, err := a.queries.GetTaskRelationsForProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return types.ConvertSlice(results, fromGeneratedGetTaskRelationsForProjectRow), nil
}

func (a *Adapter) GetTaskSummariesByColumn(ctx context.Context, columnID int64) ([]types.GetTaskSummariesByColumnRow, error) {
	results, err := a.queries.GetTaskSummariesByColumn(ctx, columnID)
	if err != nil {
		return nil, err
	}
	return types.ConvertSlice(results, fromGeneratedGetTaskSummariesByColumnRow), nil
}

func (a *Adapter) GetTaskSummariesByProject(ctx context.Context, projectID int64) ([]types.GetTaskSummariesByProjectRow, error) {
	results, err := a.queries.GetTaskSummariesByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return types.ConvertSlice(results, fromGeneratedGetTaskSummariesByProjectRow), nil
}

func (a *Adapter) GetTaskSummariesWithFilters(ctx context.Context, arg types.GetTaskSummariesWithFiltersParams) ([]types.GetTaskSummariesWithFiltersRow, error) {
	results, err := a.queries.GetTaskSummariesWithFilters(ctx, toGeneratedGetTaskSummariesWithFiltersParams(arg))
	if err != nil {
		return nil, err
	}
	return types.ConvertSlice(results, fromGeneratedGetTaskSummariesWithFiltersRow), nil
}

func (a *Adapter) GetTasksByColumn(ctx context.Context, columnID int64) ([]types.GetTasksByColumnRow, error) {
	results, err := a.queries.GetTasksByColumn(ctx, columnID)
	if err != nil {
		return nil, err
	}
	return types.ConvertSlice(results, fromGeneratedGetTasksByColumnRow), nil
}

func (a *Adapter) GetTasksForTree(ctx context.Context, id int64) ([]types.GetTasksForTreeRow, error) {
	results, err := a.queries.GetTasksForTree(ctx, id)
	if err != nil {
		return nil, err
	}
	return types.ConvertSlice(results, fromGeneratedGetTasksForTreeRow), nil
}

func (a *Adapter) MoveTaskToColumn(ctx context.Context, arg types.MoveTaskToColumnParams) error {
	return a.queries.MoveTaskToColumn(ctx, toGeneratedMoveTaskToColumnParams(arg))
}

func (a *Adapter) RemoveSubtask(ctx context.Context, arg types.RemoveSubtaskParams) error {
	return a.queries.RemoveSubtask(ctx, toGeneratedRemoveSubtaskParams(arg))
}

func (a *Adapter) SetTaskPosition(ctx context.Context, arg types.SetTaskPositionParams) error {
	return a.queries.SetTaskPosition(ctx, toGeneratedSetTaskPositionParams(arg))
}

func (a *Adapter) SetTaskPositionTemporary(ctx context.Context, id int64) error {
	return a.queries.SetTaskPositionTemporary(ctx, id)
}

func (a *Adapter) UpdateTask(ctx context.Context, arg types.UpdateTaskParams) error {
	return a.queries.UpdateTask(ctx, toGeneratedUpdateTaskParams(arg))
}

func (a *Adapter) UpdateTaskPriority(ctx context.Context, arg types.UpdateTaskPriorityParams) error {
	return a.queries.UpdateTaskPriority(ctx, toGeneratedUpdateTaskPriorityParams(arg))
}

func (a *Adapter) UpdateTaskAssignee(ctx context.Context, arg types.UpdateTaskAssigneeParams) error {
	return a.queries.UpdateTaskAssignee(ctx, toGeneratedUpdateTaskAssigneeParams(arg))
}

func (a *Adapter) UpdateTaskEstimate(ctx context.Context, arg types.UpdateTaskEstimateParams) error {
	return a.queries.UpdateTaskEstimate(ctx, toGeneratedUpdateTaskEstimateParams(arg))
}

func (a *Adapter) UpdateTaskDueDate(ctx context.Context, arg types.UpdateTaskDueDateParams) error {
	return a.queries.UpdateTaskDueDate(ctx, toGeneratedUpdateTaskDueDateParams(arg))
}

func (a *Adapter) UpdateTaskArchived(ctx context.Context, arg types.UpdateTaskArchivedParams) error {
	return a.queries.UpdateTaskArchived(ctx, toGeneratedUpdateTaskArchivedParams(arg))
}

func (a *Adapter) UpdateTaskType(ctx context.Context, arg types.UpdateTaskTypeParams) error {
	return a.queries.UpdateTaskType(ctx, toGeneratedUpdateTaskTypeParams(arg))
}
