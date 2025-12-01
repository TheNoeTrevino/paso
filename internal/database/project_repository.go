package database

import (
	"context"
	"database/sql"

	"github.com/thenoetrevino/paso/internal/models"
)

// ProjectRepo handles all project-related database operations.
type ProjectRepo struct {
	db *sql.DB
}

// Create creates a new project with default columns (Todo, In Progress, Done)
func (r *ProjectRepo) Create(ctx context.Context, name, description string) (*models.Project, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Insert the project
	result, err := tx.ExecContext(ctx,
		`INSERT INTO projects (name, description) VALUES (?, ?)`,
		name, description,
	)
	if err != nil {
		return nil, err
	}

	projectID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Initialize the project counter
	_, err = tx.ExecContext(ctx,
		`INSERT INTO project_counters (project_id, next_ticket_number) VALUES (?, 1)`,
		projectID,
	)
	if err != nil {
		return nil, err
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
			return nil, err
		}

		colID, err := colResult.LastInsertId()
		if err != nil {
			return nil, err
		}
		colIDInt := int(colID)

		// Update the previous column's next_id
		if prevColID != nil {
			_, err = tx.ExecContext(ctx, `UPDATE columns SET next_id = ? WHERE id = ?`, colIDInt, *prevColID)
			if err != nil {
				return nil, err
			}
		}

		prevColID = &colIDInt
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// Retrieve the created project
	return r.GetByID(ctx, int(projectID))
}

// GetByID retrieves a project by its ID
func (r *ProjectRepo) GetByID(ctx context.Context, id int) (*models.Project, error) {
	project := &models.Project{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, name, description, created_at, updated_at FROM projects WHERE id = ?`,
		id,
	).Scan(&project.ID, &project.Name, &project.Description, &project.CreatedAt, &project.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return project, nil
}

// GetAll retrieves all projects ordered by ID
func (r *ProjectRepo) GetAll(ctx context.Context) ([]*models.Project, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, description, created_at, updated_at FROM projects ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []*models.Project
	for rows.Next() {
		project := &models.Project{}
		if err := rows.Scan(&project.ID, &project.Name, &project.Description, &project.CreatedAt, &project.UpdatedAt); err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}

	return projects, rows.Err()
}

// Update updates a project's name and description
func (r *ProjectRepo) Update(ctx context.Context, id int, name, description string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE projects SET name = ?, description = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		name, description, id,
	)
	return err
}

// Delete removes a project and all its columns and tasks (cascade)
func (r *ProjectRepo) Delete(ctx context.Context, id int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Get all columns for this project
	rows, err := tx.QueryContext(ctx, `SELECT id FROM columns WHERE project_id = ?`, id)
	if err != nil {
		return err
	}

	var columnIDs []int
	for rows.Next() {
		var colID int
		if err := rows.Scan(&colID); err != nil {
			rows.Close()
			return err
		}
		columnIDs = append(columnIDs, colID)
	}
	rows.Close()

	// Delete all tasks in those columns
	for _, colID := range columnIDs {
		_, err = tx.ExecContext(ctx, `DELETE FROM tasks WHERE column_id = ?`, colID)
		if err != nil {
			return err
		}
	}

	// Delete all columns for this project
	_, err = tx.ExecContext(ctx, `DELETE FROM columns WHERE project_id = ?`, id)
	if err != nil {
		return err
	}

	// Delete the project counter
	_, err = tx.ExecContext(ctx, `DELETE FROM project_counters WHERE project_id = ?`, id)
	if err != nil {
		return err
	}

	// Delete the project
	_, err = tx.ExecContext(ctx, `DELETE FROM projects WHERE id = ?`, id)
	if err != nil {
		return err
	}

	return tx.Commit()
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
	return count, err
}

// GetNextTicketNumber returns and increments the next ticket number for a project
func (r *ProjectRepo) GetNextTicketNumber(ctx context.Context, projectID int) (int, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// Get current counter
	var ticketNumber int
	err = tx.QueryRowContext(ctx,
		`SELECT next_ticket_number FROM project_counters WHERE project_id = ?`,
		projectID,
	).Scan(&ticketNumber)
	if err != nil {
		return 0, err
	}

	// Increment counter
	_, err = tx.ExecContext(ctx,
		`UPDATE project_counters SET next_ticket_number = next_ticket_number + 1 WHERE project_id = ?`,
		projectID,
	)
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return ticketNumber, nil
}
