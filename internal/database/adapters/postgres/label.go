package postgres

import (
	"context"

	"github.com/thenoetrevino/paso/internal/database/types"
)

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
