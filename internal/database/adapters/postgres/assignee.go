package postgres

import (
	"context"

	"github.com/thenoetrevino/paso/internal/database/types"
)

func (a *Adapter) CreateAssignee(ctx context.Context, name string) (types.Assignee, error) {
	result, err := a.queries.CreateAssignee(ctx, name)
	if err != nil {
		return types.Assignee{}, err
	}
	return fromGeneratedAssignee(result), nil
}

func (a *Adapter) GetAssigneeByID(ctx context.Context, id int64) (types.Assignee, error) {
	result, err := a.queries.GetAssigneeByID(ctx, int32(id))
	if err != nil {
		return types.Assignee{}, err
	}
	return fromGeneratedAssignee(result), nil
}

func (a *Adapter) GetAssigneeByName(ctx context.Context, name string) (types.Assignee, error) {
	result, err := a.queries.GetAssigneeByName(ctx, name)
	if err != nil {
		return types.Assignee{}, err
	}
	return fromGeneratedAssignee(result), nil
}

func (a *Adapter) ListAssignees(ctx context.Context) ([]types.Assignee, error) {
	results, err := a.queries.ListAssignees(ctx)
	if err != nil {
		return nil, err
	}
	return types.ConvertSlice(results, fromGeneratedAssignee), nil
}

func (a *Adapter) DeleteAssignee(ctx context.Context, id int64) (int64, error) {
	return a.queries.DeleteAssignee(ctx, int32(id))
}
