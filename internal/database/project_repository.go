package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/thenoetrevino/paso/internal/models"
)

// ProjectRepo handles all project-related database operations.
type ProjectRepo struct {
	db *sql.DB
}

// CreateProject creates a new project with default columns (Todo, In Progress, Done)
func (r *ProjectRepo) CreateProject(ctx context.Context, name, description string) (*models.Project, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction for project creation: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("failed to rollback transaction: %v", err)
		}
	}()

	// Insert the project
	result, err := tx.ExecContext(ctx,
		`INSERT INTO projects (name, description) VALUES (?, ?)`,
		name, description,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert project '%s': %w", name, err)
	}

	projectID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get project ID after insert: %w", err)
	}

	// Initialize the project counter
	_, err = tx.ExecContext(ctx,
		`INSERT INTO project_counters (project_id, next_ticket_number) VALUES (?, 1)`,
		projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize project counter for project %d: %w", projectID, err)
	}

	// Create default columns for the project (as a linked list)
	defaultColumns := []string{"Todo", "In Progress", "Done"}
	var prevColID *int

	for _, colName := range defaultColumns {
		var colResult sql.Result
		if prevColID == nil {
			// First column: no prev_id
			colResult, err = tx.ExecContext(ctx,
				`INSERT INTO columns (name, project_id, prev_id, next_id) VALUES (?, ?, NULL, NULL)`,
				colName, projectID,
			)
		} else {
			// Subsequent columns: set prev_id
			colResult, err = tx.ExecContext(ctx,
				`INSERT INTO columns (name, project_id, prev_id, next_id) VALUES (?, ?, ?, NULL)`,
				colName, projectID, *prevColID,
			)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to create default column '%s' for project %d: %w", colName, projectID, err)
		}

		colID, err := colResult.LastInsertId()
		if err != nil {
			return nil, fmt.Errorf("failed to get column ID for '%s': %w", colName, err)
		}
		colIDInt := int(colID)

		// Update the previous column's next_id
		if prevColID != nil {
			_, err = tx.ExecContext(ctx, `UPDATE columns SET next_id = ? WHERE id = ?`, colIDInt, *prevColID)
			if err != nil {
				return nil, fmt.Errorf("failed to link columns for project %d: %w", projectID, err)
			}
		}

		prevColID = &colIDInt
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit project creation transaction: %w", err)
	}

	// Retrieve the created project
	return r.GetProjectByID(ctx, int(projectID))
}

// GetProjectByID retrieves a project by its ID
func (r *ProjectRepo) GetProjectByID(ctx context.Context, id int) (*models.Project, error) {
	project := &models.Project{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, name, description, created_at, updated_at FROM projects WHERE id = ?`,
		id,
	).Scan(&project.ID, &project.Name, &project.Description, &project.CreatedAt, &project.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get project %d: %w", id, err)
	}
	return project, nil
}

// GetAllProjects retrieves all projects ordered by ID
func (r *ProjectRepo) GetAllProjects(ctx context.Context) ([]*models.Project, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, description, created_at, updated_at FROM projects ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("failed to query all projects: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("failed to close rows: %v", err)
		}
	}()

	projects := make([]*models.Project, 0, 10)
	for rows.Next() {
		project := &models.Project{}
		if err := rows.Scan(&project.ID, &project.Name, &project.Description, &project.CreatedAt, &project.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan project row: %w", err)
		}
		projects = append(projects, project)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating project rows: %w", err)
	}
	return projects, nil
}

// UpdateProject updates a project's name and description
func (r *ProjectRepo) UpdateProject(ctx context.Context, id int, name, description string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE projects SET name = ?, description = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		name, description, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update project %d: %w", id, err)
	}
	return nil
}

// DeleteProject removes a project and all its columns and tasks (cascade)
func (r *ProjectRepo) DeleteProject(ctx context.Context, id int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction for project %d deletion: %w", id, err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("failed to rollback transaction: %v", err)
		}
	}()

	// Delete all tasks for columns in this project (using subquery)
	_, err = tx.ExecContext(ctx, `DELETE FROM tasks WHERE column_id IN (SELECT id FROM columns WHERE project_id = ?)`, id)
	if err != nil {
		return fmt.Errorf("failed to delete tasks for project %d: %w", id, err)
	}

	// Delete all columns for this project
	_, err = tx.ExecContext(ctx, `DELETE FROM columns WHERE project_id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete columns for project %d: %w", id, err)
	}

	// Delete the project counter
	_, err = tx.ExecContext(ctx, `DELETE FROM project_counters WHERE project_id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete counter for project %d: %w", id, err)
	}

	// Delete the project
	_, err = tx.ExecContext(ctx, `DELETE FROM projects WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete project %d: %w", id, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit project %d deletion: %w", id, err)
	}
	return nil
}

// GetTaskCount returns the total number of tasks in a project
func (r *ProjectRepo) GetTaskCount(ctx context.Context, projectID int) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM tasks t
		JOIN columns c ON t.column_id = c.id
		WHERE c.project_id = ?
	`, projectID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get task count for project %d: %w", projectID, err)
	}
	return count, nil
}

// GetNextTicketNumber returns and increments the next ticket number for a project
func (r *ProjectRepo) GetNextTicketNumber(ctx context.Context, projectID int) (int, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction for ticket number: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("failed to rollback transaction: %v", err)
		}
	}()

	// Get current counter
	var ticketNumber int
	err = tx.QueryRowContext(ctx,
		`SELECT next_ticket_number FROM project_counters WHERE project_id = ?`,
		projectID,
	).Scan(&ticketNumber)
	if err != nil {
		return 0, fmt.Errorf("failed to get next ticket number for project %d: %w", projectID, err)
	}

	// Increment counter
	_, err = tx.ExecContext(ctx,
		`UPDATE project_counters SET next_ticket_number = next_ticket_number + 1 WHERE project_id = ?`,
		projectID,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to increment ticket counter for project %d: %w", projectID, err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit ticket number transaction: %w", err)
	}

	return ticketNumber, nil
}
