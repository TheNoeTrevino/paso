package sqlite

import (
	"context"

	"github.com/thenoetrevino/paso/internal/database/types"
)

func (a *Adapter) CreateTaskEvent(ctx context.Context, arg types.CreateTaskEventParams) (types.TaskEvent, error) {
	result, err := a.queries.CreateTaskEvent(ctx, toGeneratedCreateTaskEventParams(arg))
	if err != nil {
		return types.TaskEvent{}, err
	}
	return fromGeneratedTaskEvent(result), nil
}

func (a *Adapter) GetEventsByTask(ctx context.Context, taskID int64) ([]types.TaskEvent, error) {
	results, err := a.queries.GetEventsByTask(ctx, taskID)
	if err != nil {
		return nil, err
	}
	return types.ConvertSlice(results, fromGeneratedTaskEvent), nil
}

func (a *Adapter) DeleteEventsByTask(ctx context.Context, taskID int64) error {
	return a.queries.DeleteEventsByTask(ctx, taskID)
}
