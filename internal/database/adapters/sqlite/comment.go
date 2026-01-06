package sqlite

import (
	"context"

	"github.com/thenoetrevino/paso/internal/database/types"
)

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
