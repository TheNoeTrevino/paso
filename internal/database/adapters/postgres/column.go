package postgres

import (
	"context"

	"github.com/thenoetrevino/paso/internal/database/types"
)

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
