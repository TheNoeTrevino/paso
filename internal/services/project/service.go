package project

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/thenoetrevino/paso/internal/database"
	"github.com/thenoetrevino/paso/internal/database/types"
	"github.com/thenoetrevino/paso/internal/events"
	"github.com/thenoetrevino/paso/internal/git"
	"github.com/thenoetrevino/paso/internal/models"
)

// GitChecker defines the interface for checking git branch existence
type GitChecker interface {
	BranchExists(ctx context.Context, branchName string) (bool, error)
}

// realGitChecker implements GitChecker using the actual git package
type realGitChecker struct{}

func (r *realGitChecker) BranchExists(ctx context.Context, branchName string) (bool, error) {
	return git.BranchExists(ctx, branchName)
}

// Service defines all project-related business operations
type Service interface {
	// Read operations
	GetAllProjects(ctx context.Context) ([]*models.Project, error)
	GetProjectByID(ctx context.Context, id int) (*models.Project, error)
	GetProjectByGitBranch(ctx context.Context, gitBranch string) (*models.Project, error)
	GetTaskCount(ctx context.Context, projectID int) (int, error)

	// Write operations
	CreateProject(ctx context.Context, req CreateProjectRequest) (*models.Project, error)
	UpdateProject(ctx context.Context, req UpdateProjectRequest) error
	DeleteProject(ctx context.Context, id int, force bool) error
}

// CreateProjectRequest encapsulates data for creating a project
type CreateProjectRequest struct {
	Name        string
	Description string
	GitBranch   string
}

// UpdateProjectRequest encapsulates data for updating a project
type UpdateProjectRequest struct {
	ID          int
	Name        *string
	Description *string
	GitBranch   *string
}

// service implements Service interface using database.Querier abstraction
type service struct {
	db          *sql.DB
	dbType      database.DatabaseType
	queries     database.Querier
	eventClient events.EventPublisher
	gitChecker  GitChecker
}

// NewService creates a new project service with database-agnostic queries.
// If gitChecker is nil, a real git checker will be used.
func NewService(db *sql.DB, dbType database.DatabaseType, eventClient events.EventPublisher, gitChecker GitChecker) (Service, error) {
	queries, err := database.NewQuerier(db, dbType)
	if err != nil {
		return nil, fmt.Errorf("failed to create project service: %w", err)
	}

	if gitChecker == nil {
		gitChecker = &realGitChecker{}
	}

	return &service{
		db:          db,
		dbType:      dbType,
		queries:     queries,
		eventClient: eventClient,
		gitChecker:  gitChecker,
	}, nil
}

// GetAllProjects retrieves all projects
func (s *service) GetAllProjects(ctx context.Context) ([]*models.Project, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	projects, err := s.queries.GetAllProjects(ctx)
	if err != nil {
		return nil, err
	}
	return toProjectModels(projects), nil
}

// GetProjectByID retrieves a specific project
func (s *service) GetProjectByID(ctx context.Context, id int) (*models.Project, error) {
	if err := validateProjectID(id); err != nil {
		return nil, err
	}
	project, err := s.queries.GetProjectByID(ctx, int64(id))
	if err != nil {
		return nil, err
	}
	return toProjectModel(project), nil
}

// GetProjectByGitBranch retrieves a project by its git branch
// Returns (nil, nil) if not found or if gitBranch is empty
func (s *service) GetProjectByGitBranch(ctx context.Context, gitBranch string) (*models.Project, error) {
	if gitBranch == "" {
		return nil, nil
	}
	project, err := s.queries.GetProjectByGitBranch(ctx, gitBranch)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found is not an error
		}
		return nil, err
	}
	return toProjectModel(project), nil
}

// GetTaskCount returns the number of tasks in a project
func (s *service) GetTaskCount(ctx context.Context, projectID int) (int, error) {
	if err := validateProjectID(projectID); err != nil {
		return 0, err
	}
	count, err := s.queries.GetProjectTaskCount(ctx, int64(projectID))
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

// CreateProject creates a new project with validation
func (s *service) CreateProject(ctx context.Context, req CreateProjectRequest) (*models.Project, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := validateCreateProjectRequest(req); err != nil {
		return nil, err
	}

	var project types.Project

	sanitizedBranch := ""
	if req.GitBranch != "" {
		if err := git.ValidateBranchName(ctx, req.GitBranch); err != nil {
			return nil, fmt.Errorf("invalid git branch name: %w", err)
		}

		var sanitizeErr error
		sanitizedBranch, sanitizeErr = git.SanitizeBranchName(req.GitBranch)
		if sanitizeErr != nil {
			return nil, fmt.Errorf("invalid branch name: %w", sanitizeErr)
		}

		gitInfo := git.DetectGitInfo(ctx)
		if gitInfo.IsRepo {
			exists, err := s.gitChecker.BranchExists(ctx, sanitizedBranch)
			if err != nil {
				return nil, fmt.Errorf("failed to check branch existence: %w", err)
			}
			if !exists {
				return nil, ErrBranchDoesNotExist
			}
		}
	}

	// Use WithTx helper for transaction management
	err := database.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		qtx := database.MustNewQuerier(tx, s.dbType)

		// Create project record
		var projErr error
		project, projErr = qtx.CreateProjectRecord(ctx, types.CreateProjectRecordParams{
			Name:        req.Name,
			Description: types.NullString{String: req.Description, Valid: req.Description != ""},
			GitBranch:   types.NullString{String: sanitizedBranch, Valid: sanitizedBranch != ""},
		})
		if projErr != nil {
			// Check for unique constraint violation on git_branch
			if database.IsUniqueViolation(projErr) {
				return ErrGitBranchAlreadyAssociated
			}
			return fmt.Errorf("failed to create project: %w", projErr)
		}

		// Initialize project counter (for task ticket numbers)
		if err := qtx.InitializeProjectCounter(ctx, project.ID); err != nil {
			return fmt.Errorf("failed to initialize project counter: %w", err)
		}

		// Create default columns (Todo, In Progress, Done)
		if err := database.CreateDefaultColumns(ctx, qtx, project.ID); err != nil {
			return fmt.Errorf("failed to create default columns: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Publish event after successful commit
	s.publishProjectEvent(ctx, int(project.ID))

	return toProjectModel(project), nil
}

// UpdateProject updates an existing project
func (s *service) UpdateProject(ctx context.Context, req UpdateProjectRequest) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := validateUpdateProjectRequest(req); err != nil {
		return err
	}

	// Get existing project to fill in missing fields
	existing, err := s.queries.GetProjectByID(ctx, int64(req.ID))
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	// merge the two values
	name := existing.Name
	if req.Name != nil {
		name = *req.Name
	}

	description := existing.Description
	if req.Description != nil {
		description = types.NullString{String: *req.Description, Valid: *req.Description != ""}
	}

	gitBranch := existing.GitBranch
	if req.GitBranch != nil {
		if *req.GitBranch != "" {
			if err := git.ValidateBranchName(ctx, *req.GitBranch); err != nil {
				return fmt.Errorf("invalid git branch name: %w", err)
			}

			sanitized, err := git.SanitizeBranchName(*req.GitBranch)
			if err != nil {
				return fmt.Errorf("invalid branch name: %w", err)
			}

			gitInfo := git.DetectGitInfo(ctx)
			if gitInfo.IsRepo {
				exists, err := s.gitChecker.BranchExists(ctx, sanitized)
				if err != nil {
					return fmt.Errorf("failed to check branch existence: %w", err)
				}
				if !exists {
					return ErrBranchDoesNotExist
				}
			}

			gitBranch = types.NullString{String: sanitized, Valid: true}
		} else {
			gitBranch = types.NullString{String: "", Valid: false}
		}
	}

	// Update project
	if err := s.queries.UpdateProject(ctx, types.UpdateProjectParams{
		ID:          int64(req.ID),
		Name:        name,
		Description: description,
		GitBranch:   gitBranch,
	}); err != nil {
		// Check for unique constraint violation on git_branch
		if database.IsUniqueViolation(err) {
			return ErrGitBranchAlreadyAssociated
		}
		return fmt.Errorf("failed to update project: %w", err)
	}

	// Publish event
	s.publishProjectEvent(ctx, req.ID)

	return nil
}

// DeleteProject deletes a project (business rule: must not have tasks unless force=true)
func (s *service) DeleteProject(ctx context.Context, id int, force bool) error {
	if err := validateProjectID(id); err != nil {
		return err
	}

	// Business rule: Check if project has tasks (unless force is enabled)
	if !force {
		taskCount, err := s.queries.GetProjectTaskCount(ctx, int64(id))
		if err != nil {
			return fmt.Errorf("failed to check project tasks: %w", err)
		}
		if taskCount > 0 {
			return ErrProjectHasTasks
		}
	}

	// Use WithTx helper for transaction management
	err := database.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		qtx := database.MustNewQuerier(tx, s.dbType)

		// Delete tasks first
		if err := qtx.DeleteTasksByProject(ctx, int64(id)); err != nil {
			return fmt.Errorf("failed to delete tasks: %w", err)
		}

		// Delete columns
		if err := qtx.DeleteColumnsByProject(ctx, int64(id)); err != nil {
			return fmt.Errorf("failed to delete columns: %w", err)
		}

		// Delete project counter
		if err := qtx.DeleteProjectCounter(ctx, int64(id)); err != nil {
			return fmt.Errorf("failed to delete counter: %w", err)
		}

		// Delete project
		if err := qtx.DeleteProject(ctx, int64(id)); err != nil {
			return fmt.Errorf("failed to delete project: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	// Publish event after successful deletion
	s.publishProjectEvent(ctx, id)

	return nil
}

// publishProjectEvent publishes a project event with retry logic
func (s *service) publishProjectEvent(ctx context.Context, projectID int) {
	if s.eventClient == nil {
		return
	}

	// Publish with retry (3 attempts with exponential backoff)
	// Non-blocking: errors are logged but don't affect the operation
	_ = events.PublishWithRetry(s.eventClient, events.Event{
		Type:      events.EventDatabaseChanged,
		ProjectID: projectID,
	}, 3)
}

func toProjectModel(p types.Project) *models.Project {
	return &models.Project{
		ID:          int(p.ID),
		Name:        p.Name,
		Description: database.NullStringToString(p.Description.ToSQLNullString()),
		GitBranch:   database.NullStringToString(p.GitBranch.ToSQLNullString()),
		CreatedAt:   database.NullTimeToTime(p.CreatedAt.ToSQLNullTime()),
		UpdatedAt:   database.NullTimeToTime(p.UpdatedAt.ToSQLNullTime()),
	}
}

func toProjectModels(projects []types.Project) []*models.Project {
	result := make([]*models.Project, len(projects))
	for i, p := range projects {
		result[i] = toProjectModel(p)
	}
	return result
}
