package sqlite

import (
	"context"

	"github.com/thenoetrevino/paso/internal/database/types"
)

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
	return types.ConvertSlice(results, fromGeneratedGetAllProjectsRow), nil
}

func (a *Adapter) GetNextTaskNumber(ctx context.Context, projectID int64) (types.NullInt64, error) {
	result, err := a.queries.GetNextTaskNumber(ctx, projectID)
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
	return fromGeneratedGetProjectByIDRow(result), nil
}

func (a *Adapter) GetProjectByGitBranch(ctx context.Context, gitBranch string) (types.Project, error) {
	result, err := a.queries.GetProjectByGitBranch(ctx, gitBranch)
	if err != nil {
		return types.Project{}, err
	}
	return fromGeneratedGetProjectByGitBranchRow(result), nil
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

func (a *Adapter) IncrementTaskNumber(ctx context.Context, projectID int64) error {
	return a.queries.IncrementTaskNumber(ctx, projectID)
}

func (a *Adapter) InitializeProjectCounter(ctx context.Context, projectID int64) error {
	return a.queries.InitializeProjectCounter(ctx, projectID)
}

func (a *Adapter) UpdateProject(ctx context.Context, arg types.UpdateProjectParams) error {
	return a.queries.UpdateProject(ctx, toGeneratedUpdateProjectParams(arg))
}
