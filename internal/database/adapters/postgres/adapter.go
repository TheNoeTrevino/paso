package postgres

import (
	"context"

	"github.com/thenoetrevino/paso/internal/database/generated_postgres"
	"github.com/thenoetrevino/paso/internal/database/types"
)

// Adapter wraps the generated PostgreSQL queries and implements the types.Querier interface
// by converting between database-agnostic types and PostgreSQL-specific generated types.
type Adapter struct {
	queries *generated_postgres.Queries
}

// New creates a new PostgreSQL adapter that implements types.Querier.
func New(db types.DBTX) *Adapter {
	return &Adapter{
		queries: generated_postgres.New(db),
	}
}

// Compile-time check that Adapter implements types.Querier interface
var _ types.Querier = (*Adapter)(nil)

// ============================================================================
// Label Operations
// ============================================================================

func (a *Adapter) AddLabelToTask(ctx context.Context, arg types.AddLabelToTaskParams) error {
	return a.queries.AddLabelToTask(ctx, toGeneratedAddLabelToTaskParams(arg))
}

func (a *Adapter) CreateLabel(ctx context.Context, arg types.CreateLabelParams) (types.Label, error) {
	result, err := a.queries.CreateLabel(ctx, toGeneratedCreateLabelParams(arg))
	if err != nil {
		return types.Label{}, err
	}
	return fromGeneratedLabel(result), nil
}

func (a *Adapter) DeleteAllLabelsFromTask(ctx context.Context, taskID int64) error {
	return a.queries.DeleteAllLabelsFromTask(ctx, taskID)
}

func (a *Adapter) DeleteLabel(ctx context.Context, id int64) error {
	return a.queries.DeleteLabel(ctx, id)
}

func (a *Adapter) GetLabelByID(ctx context.Context, id int64) (types.Label, error) {
	result, err := a.queries.GetLabelByID(ctx, id)
	if err != nil {
		return types.Label{}, err
	}
	return fromGeneratedLabel(result), nil
}

func (a *Adapter) GetLabelsByProject(ctx context.Context, projectID int64) ([]types.Label, error) {
	results, err := a.queries.GetLabelsByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return types.ConvertSlice(results, fromGeneratedLabel), nil
}

func (a *Adapter) GetLabelsForTask(ctx context.Context, taskID int64) ([]types.Label, error) {
	results, err := a.queries.GetLabelsForTask(ctx, taskID)
	if err != nil {
		return nil, err
	}
	return types.ConvertSlice(results, fromGeneratedLabel), nil
}

func (a *Adapter) GetTaskLabels(ctx context.Context, taskID int64) ([]types.Label, error) {
	results, err := a.queries.GetTaskLabels(ctx, taskID)
	if err != nil {
		return nil, err
	}
	return types.ConvertSlice(results, fromGeneratedLabel), nil
}

func (a *Adapter) InsertTaskLabel(ctx context.Context, arg types.InsertTaskLabelParams) error {
	return a.queries.InsertTaskLabel(ctx, toGeneratedInsertTaskLabelParams(arg))
}

func (a *Adapter) RemoveLabelFromTask(ctx context.Context, arg types.RemoveLabelFromTaskParams) error {
	return a.queries.RemoveLabelFromTask(ctx, toGeneratedRemoveLabelFromTaskParams(arg))
}

func (a *Adapter) UpdateLabel(ctx context.Context, arg types.UpdateLabelParams) error {
	return a.queries.UpdateLabel(ctx, toGeneratedUpdateLabelParams(arg))
}

// ============================================================================
// Column Operations
// ============================================================================

func (a *Adapter) ClearCompletedColumnByProject(ctx context.Context, projectID int64) error {
	return a.queries.ClearCompletedColumnByProject(ctx, projectID)
}

func (a *Adapter) ClearInProgressColumnByProject(ctx context.Context, projectID int64) error {
	return a.queries.ClearInProgressColumnByProject(ctx, projectID)
}

func (a *Adapter) ClearReadyColumnByProject(ctx context.Context, projectID int64) error {
	return a.queries.ClearReadyColumnByProject(ctx, projectID)
}

func (a *Adapter) ColumnExists(ctx context.Context, id int64) (int64, error) {
	return a.queries.ColumnExists(ctx, id)
}

func (a *Adapter) CreateColumn(ctx context.Context, arg types.CreateColumnParams) (types.Column, error) {
	result, err := a.queries.CreateColumn(ctx, toGeneratedCreateColumnParams(arg))
	if err != nil {
		return types.Column{}, err
	}
	return fromGeneratedColumn(result), nil
}

func (a *Adapter) DeleteColumn(ctx context.Context, id int64) error {
	return a.queries.DeleteColumn(ctx, id)
}

func (a *Adapter) DeleteColumnsByProject(ctx context.Context, projectID int64) error {
	return a.queries.DeleteColumnsByProject(ctx, projectID)
}

func (a *Adapter) GetColumnByID(ctx context.Context, id int64) (types.GetColumnByIDRow, error) {
	result, err := a.queries.GetColumnByID(ctx, id)
	if err != nil {
		return types.GetColumnByIDRow{}, err
	}
	return fromGeneratedGetColumnByIDRow(result), nil
}

func (a *Adapter) GetColumnLinkedListInfo(ctx context.Context, id int64) (types.GetColumnLinkedListInfoRow, error) {
	result, err := a.queries.GetColumnLinkedListInfo(ctx, id)
	if err != nil {
		return types.GetColumnLinkedListInfoRow{}, err
	}
	return fromGeneratedGetColumnLinkedListInfoRow(result), nil
}

func (a *Adapter) GetColumnNextID(ctx context.Context, id int64) (types.NullInt64, error) {
	result, err := a.queries.GetColumnNextID(ctx, id)
	if err != nil {
		return types.NullInt64{}, err
	}
	return types.FromSQLNullInt64(result), nil
}

func (a *Adapter) GetColumnsByProject(ctx context.Context, projectID int64) ([]types.GetColumnsByProjectRow, error) {
	results, err := a.queries.GetColumnsByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return types.ConvertSlice(results, fromGeneratedGetColumnsByProjectRow), nil
}

func (a *Adapter) GetCompletedColumnByProject(ctx context.Context, projectID int64) (types.GetCompletedColumnByProjectRow, error) {
	result, err := a.queries.GetCompletedColumnByProject(ctx, projectID)
	if err != nil {
		return types.GetCompletedColumnByProjectRow{}, err
	}
	return fromGeneratedGetCompletedColumnByProjectRow(result), nil
}

func (a *Adapter) GetInProgressColumnByProject(ctx context.Context, projectID int64) (types.GetInProgressColumnByProjectRow, error) {
	result, err := a.queries.GetInProgressColumnByProject(ctx, projectID)
	if err != nil {
		return types.GetInProgressColumnByProjectRow{}, err
	}
	return fromGeneratedGetInProgressColumnByProjectRow(result), nil
}

func (a *Adapter) GetNextColumnID(ctx context.Context, id int64) (types.NullInt64, error) {
	result, err := a.queries.GetNextColumnID(ctx, id)
	if err != nil {
		return types.NullInt64{}, err
	}
	return types.FromSQLNullInt64(result), nil
}

func (a *Adapter) GetPrevColumnID(ctx context.Context, id int64) (types.NullInt64, error) {
	result, err := a.queries.GetPrevColumnID(ctx, id)
	if err != nil {
		return types.NullInt64{}, err
	}
	return types.FromSQLNullInt64(result), nil
}

func (a *Adapter) GetReadyColumnByProject(ctx context.Context, projectID int64) (types.GetReadyColumnByProjectRow, error) {
	result, err := a.queries.GetReadyColumnByProject(ctx, projectID)
	if err != nil {
		return types.GetReadyColumnByProjectRow{}, err
	}
	return fromGeneratedGetReadyColumnByProjectRow(result), nil
}

func (a *Adapter) GetTailColumnForProject(ctx context.Context, projectID int64) (int64, error) {
	return a.queries.GetTailColumnForProject(ctx, projectID)
}

func (a *Adapter) UpdateColumnHoldsCompletedTasks(ctx context.Context, arg types.UpdateColumnHoldsCompletedTasksParams) error {
	return a.queries.UpdateColumnHoldsCompletedTasks(ctx, toGeneratedUpdateColumnHoldsCompletedTasksParams(arg))
}

func (a *Adapter) UpdateColumnHoldsInProgressTasks(ctx context.Context, arg types.UpdateColumnHoldsInProgressTasksParams) error {
	return a.queries.UpdateColumnHoldsInProgressTasks(ctx, toGeneratedUpdateColumnHoldsInProgressTasksParams(arg))
}

func (a *Adapter) UpdateColumnHoldsReadyTasks(ctx context.Context, arg types.UpdateColumnHoldsReadyTasksParams) error {
	return a.queries.UpdateColumnHoldsReadyTasks(ctx, toGeneratedUpdateColumnHoldsReadyTasksParams(arg))
}

func (a *Adapter) UpdateColumnName(ctx context.Context, arg types.UpdateColumnNameParams) error {
	return a.queries.UpdateColumnName(ctx, toGeneratedUpdateColumnNameParams(arg))
}

func (a *Adapter) UpdateColumnNextID(ctx context.Context, arg types.UpdateColumnNextIDParams) error {
	return a.queries.UpdateColumnNextID(ctx, toGeneratedUpdateColumnNextIDParams(arg))
}

func (a *Adapter) UpdateColumnPrevID(ctx context.Context, arg types.UpdateColumnPrevIDParams) error {
	return a.queries.UpdateColumnPrevID(ctx, toGeneratedUpdateColumnPrevIDParams(arg))
}

// ============================================================================
// Project Operations
// ============================================================================

func (a *Adapter) CreateProjectRecord(ctx context.Context, arg types.CreateProjectRecordParams) (types.Project, error) {
	result, err := a.queries.CreateProjectRecord(ctx, toGeneratedCreateProjectRecordParams(arg))
	if err != nil {
		return types.Project{}, err
	}
	return fromGeneratedProject(result), nil
}

func (a *Adapter) DeleteProject(ctx context.Context, id int64) error {
	return a.queries.DeleteProject(ctx, id)
}

func (a *Adapter) DeleteProjectCounter(ctx context.Context, projectID int64) error {
	return a.queries.DeleteProjectCounter(ctx, projectID)
}

func (a *Adapter) GetAllProjects(ctx context.Context) ([]types.Project, error) {
	results, err := a.queries.GetAllProjects(ctx)
	if err != nil {
		return nil, err
	}
	return types.ConvertSlice(results, fromGeneratedProject), nil
}

func (a *Adapter) GetNextTicketNumber(ctx context.Context, projectID int64) (types.NullInt64, error) {
	result, err := a.queries.GetNextTicketNumber(ctx, projectID)
	if err != nil {
		return types.NullInt64{}, err
	}
	return types.FromSQLNullInt64(result), nil
}

func (a *Adapter) GetProjectByID(ctx context.Context, id int64) (types.Project, error) {
	result, err := a.queries.GetProjectByID(ctx, id)
	if err != nil {
		return types.Project{}, err
	}
	return fromGeneratedProject(result), nil
}

func (a *Adapter) GetProjectIDFromColumn(ctx context.Context, id int64) (int64, error) {
	return a.queries.GetProjectIDFromColumn(ctx, id)
}

func (a *Adapter) GetProjectIDFromTask(ctx context.Context, id int64) (int64, error) {
	return a.queries.GetProjectIDFromTask(ctx, id)
}

func (a *Adapter) GetProjectTaskCount(ctx context.Context, projectID int64) (int64, error) {
	return a.queries.GetProjectTaskCount(ctx, projectID)
}

func (a *Adapter) IncrementTicketNumber(ctx context.Context, projectID int64) error {
	return a.queries.IncrementTicketNumber(ctx, projectID)
}

func (a *Adapter) InitializeProjectCounter(ctx context.Context, projectID int64) error {
	return a.queries.InitializeProjectCounter(ctx, projectID)
}

func (a *Adapter) UpdateProject(ctx context.Context, arg types.UpdateProjectParams) error {
	return a.queries.UpdateProject(ctx, toGeneratedUpdateProjectParams(arg))
}

// ============================================================================
// Task Operations
// ============================================================================

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

func (a *Adapter) GetTaskSummariesByProjectFiltered(ctx context.Context, arg types.GetTaskSummariesByProjectFilteredParams) ([]types.GetTaskSummariesByProjectFilteredRow, error) {
	results, err := a.queries.GetTaskSummariesByProjectFiltered(ctx, toGeneratedGetTaskSummariesByProjectFilteredParams(arg))
	if err != nil {
		return nil, err
	}
	return types.ConvertSlice(results, fromGeneratedGetTaskSummariesByProjectFilteredRow), nil
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

func (a *Adapter) UpdateTaskType(ctx context.Context, arg types.UpdateTaskTypeParams) error {
	return a.queries.UpdateTaskType(ctx, toGeneratedUpdateTaskTypeParams(arg))
}

// ============================================================================
// Comment Operations
// ============================================================================

func (a *Adapter) CreateComment(ctx context.Context, arg types.CreateCommentParams) (types.TaskComment, error) {
	result, err := a.queries.CreateComment(ctx, toGeneratedCreateCommentParams(arg))
	if err != nil {
		return types.TaskComment{}, err
	}
	return fromGeneratedTaskComment(result), nil
}

func (a *Adapter) DeleteComment(ctx context.Context, id int64) error {
	return a.queries.DeleteComment(ctx, id)
}

func (a *Adapter) GetComment(ctx context.Context, id int64) (types.TaskComment, error) {
	result, err := a.queries.GetComment(ctx, id)
	if err != nil {
		return types.TaskComment{}, err
	}
	return fromGeneratedTaskComment(result), nil
}

func (a *Adapter) GetCommentCountByTask(ctx context.Context, taskID int64) (int64, error) {
	return a.queries.GetCommentCountByTask(ctx, taskID)
}

func (a *Adapter) GetCommentsByTask(ctx context.Context, taskID int64) ([]types.TaskComment, error) {
	results, err := a.queries.GetCommentsByTask(ctx, taskID)
	if err != nil {
		return nil, err
	}
	return types.ConvertSlice(results, fromGeneratedTaskComment), nil
}

func (a *Adapter) UpdateComment(ctx context.Context, arg types.UpdateCommentParams) error {
	return a.queries.UpdateComment(ctx, toGeneratedUpdateCommentParams(arg))
}
