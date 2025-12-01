package database

import (
	"context"
	"database/sql"

	"github.com/thenoetrevino/paso/internal/models"
)

// Repository provides a unified interface to all data operations.
// It composes domain-specific repositories using struct embedding.
type Repository struct {
	*ProjectRepo
	*ColumnRepo
	*TaskRepo
	*LabelRepo
}

// NewRepository creates a new Repository instance wrapping the given database connection.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		ProjectRepo: &ProjectRepo{db: db},
		ColumnRepo:  &ColumnRepo{db: db},
		TaskRepo:    &TaskRepo{db: db},
		LabelRepo:   &LabelRepo{db: db},
	}
}

// Wrapper methods for ProjectRepo to maintain existing API
func (r *Repository) CreateProject(ctx context.Context, name, description string) (*models.Project, error) {
	return r.ProjectRepo.Create(ctx, name, description)
}

func (r *Repository) GetAllProjects(ctx context.Context) ([]*models.Project, error) {
	return r.ProjectRepo.GetAll(ctx)
}

func (r *Repository) GetProjectByID(ctx context.Context, id int) (*models.Project, error) {
	return r.ProjectRepo.GetByID(ctx, id)
}

func (r *Repository) UpdateProject(ctx context.Context, id int, name, description string) error {
	return r.ProjectRepo.Update(ctx, id, name, description)
}

func (r *Repository) DeleteProject(ctx context.Context, id int) error {
	return r.ProjectRepo.Delete(ctx, id)
}

// Wrapper methods for ColumnRepo to maintain existing API
func (r *Repository) CreateColumn(ctx context.Context, name string, projectID int, afterID *int) (*models.Column, error) {
	return r.ColumnRepo.Create(ctx, name, projectID, afterID)
}

func (r *Repository) GetColumnsByProject(ctx context.Context, projectID int) ([]*models.Column, error) {
	return r.ColumnRepo.GetByProject(ctx, projectID)
}

func (r *Repository) GetColumnByID(ctx context.Context, id int) (*models.Column, error) {
	return r.ColumnRepo.GetByID(ctx, id)
}

func (r *Repository) UpdateColumnName(ctx context.Context, id int, name string) error {
	return r.ColumnRepo.UpdateName(ctx, id, name)
}

func (r *Repository) DeleteColumn(ctx context.Context, id int) error {
	return r.ColumnRepo.Delete(ctx, id)
}

// Wrapper methods for TaskRepo to maintain existing API
func (r *Repository) CreateTask(ctx context.Context, title, description string, columnID, position int) (*models.Task, error) {
	return r.TaskRepo.Create(ctx, title, description, columnID, position)
}

func (r *Repository) GetTaskSummariesByColumn(ctx context.Context, columnID int) ([]*models.TaskSummary, error) {
	return r.TaskRepo.GetSummariesByColumn(ctx, columnID)
}

func (r *Repository) GetTaskDetail(ctx context.Context, id int) (*models.TaskDetail, error) {
	return r.TaskRepo.GetDetail(ctx, id)
}

func (r *Repository) GetTaskCountByColumn(ctx context.Context, columnID int) (int, error) {
	return r.TaskRepo.GetCountByColumn(ctx, columnID)
}

func (r *Repository) UpdateTask(ctx context.Context, id int, title, description string) error {
	return r.TaskRepo.Update(ctx, id, title, description)
}

func (r *Repository) MoveTaskToNextColumn(ctx context.Context, taskID int) error {
	return r.TaskRepo.MoveToNextColumn(ctx, taskID)
}

func (r *Repository) MoveTaskToPrevColumn(ctx context.Context, taskID int) error {
	return r.TaskRepo.MoveToPrevColumn(ctx, taskID)
}

func (r *Repository) DeleteTask(ctx context.Context, id int) error {
	return r.TaskRepo.Delete(ctx, id)
}

// Wrapper methods for LabelRepo to maintain existing API
func (r *Repository) CreateLabel(ctx context.Context, projectID int, name, color string) (*models.Label, error) {
	return r.LabelRepo.Create(ctx, projectID, name, color)
}

func (r *Repository) GetLabelsByProject(ctx context.Context, projectID int) ([]*models.Label, error) {
	return r.LabelRepo.GetByProject(ctx, projectID)
}

func (r *Repository) GetLabelsForTask(ctx context.Context, taskID int) ([]*models.Label, error) {
	return r.LabelRepo.GetForTask(ctx, taskID)
}

func (r *Repository) UpdateLabel(ctx context.Context, id int, name, color string) error {
	return r.LabelRepo.Update(ctx, id, name, color)
}

func (r *Repository) DeleteLabel(ctx context.Context, id int) error {
	return r.LabelRepo.Delete(ctx, id)
}

func (r *Repository) AddLabelToTask(ctx context.Context, taskID, labelID int) error {
	return r.LabelRepo.AddToTask(ctx, taskID, labelID)
}

func (r *Repository) RemoveLabelFromTask(ctx context.Context, taskID, labelID int) error {
	return r.LabelRepo.RemoveFromTask(ctx, taskID, labelID)
}

func (r *Repository) SetTaskLabels(ctx context.Context, taskID int, labelIDs []int) error {
	return r.LabelRepo.SetForTask(ctx, taskID, labelIDs)
}
