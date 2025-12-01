package database

import (
	"context"

	"github.com/thenoetrevino/paso/internal/models"
)

// ProjectReader defines read operations for projects.
type ProjectReader interface {
	GetAllProjects(ctx context.Context) ([]*models.Project, error)
	GetProjectByID(ctx context.Context, id int) (*models.Project, error)
}

// ProjectWriter defines write operations for projects.
type ProjectWriter interface {
	CreateProject(ctx context.Context, name, description string) (*models.Project, error)
	UpdateProject(ctx context.Context, id int, name, description string) error
	DeleteProject(ctx context.Context, id int) error
}

// ProjectRepository combines all project-related operations.
type ProjectRepository interface {
	ProjectReader
	ProjectWriter
}
