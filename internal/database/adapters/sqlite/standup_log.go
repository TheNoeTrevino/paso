package sqlite

import (
	"context"

	"github.com/thenoetrevino/paso/internal/database/types"
)

func (a *Adapter) CreateStandupLog(ctx context.Context, arg types.CreateStandupLogParams) (types.StandupLog, error) {
	result, err := a.queries.CreateStandupLog(ctx, toGeneratedCreateStandupLogParams(arg))
	if err != nil {
		return types.StandupLog{}, err
	}
	return fromGeneratedStandupLog(result), nil
}

func (a *Adapter) GetStandupLog(ctx context.Context, id int64) (types.StandupLog, error) {
	result, err := a.queries.GetStandupLog(ctx, id)
	if err != nil {
		return types.StandupLog{}, err
	}
	return fromGeneratedStandupLog(result), nil
}

func (a *Adapter) GetStandupLogsByProject(ctx context.Context, projectID int64) ([]types.StandupLog, error) {
	results, err := a.queries.GetStandupLogsByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return types.ConvertSlice(results, fromGeneratedStandupLog), nil
}

func (a *Adapter) GetStandupLogsByProjectAndDateRange(ctx context.Context, arg types.GetStandupLogsByProjectAndDateRangeParams) ([]types.StandupLog, error) {
	results, err := a.queries.GetStandupLogsByProjectAndDateRange(ctx, toGeneratedGetStandupLogsByProjectAndDateRangeParams(arg))
	if err != nil {
		return nil, err
	}
	return types.ConvertSlice(results, fromGeneratedStandupLog), nil
}

func (a *Adapter) DeleteStandupLog(ctx context.Context, id int64) error {
	return a.queries.DeleteStandupLog(ctx, id)
}
