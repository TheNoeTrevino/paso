package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/thenoetrevino/paso/internal/database/types"
)

const createStandupLogSQL = `INSERT INTO standup_logs (project_id, content) VALUES (?, ?) RETURNING id, project_id, content, created_at`
const getStandupLogSQL = `SELECT id, project_id, content, created_at FROM standup_logs WHERE id = ?`
const getStandupLogsByProjectSQL = `SELECT id, project_id, content, created_at FROM standup_logs WHERE project_id = ? ORDER BY created_at DESC`
const deleteStandupLogSQL = `DELETE FROM standup_logs WHERE id = ?`

func (a *Adapter) CreateStandupLog(ctx context.Context, arg types.CreateStandupLogParams) (types.StandupLog, error) {
	row := a.db.QueryRowContext(ctx, createStandupLogSQL, arg.ProjectID, arg.Content)
	var s types.StandupLog
	var createdAt sql.NullTime
	err := row.Scan(&s.ID, &s.ProjectID, &s.Content, &createdAt)
	if err != nil {
		return types.StandupLog{}, fmt.Errorf("create standup log for project %d: %w", arg.ProjectID, err)
	}
	s.CreatedAt = types.FromSQLNullTime(createdAt)
	return s, nil
}

func (a *Adapter) GetStandupLog(ctx context.Context, id int64) (types.StandupLog, error) {
	row := a.db.QueryRowContext(ctx, getStandupLogSQL, id)
	var s types.StandupLog
	var createdAt sql.NullTime
	err := row.Scan(&s.ID, &s.ProjectID, &s.Content, &createdAt)
	if err != nil {
		return types.StandupLog{}, fmt.Errorf("get standup log %d: %w", id, err)
	}
	s.CreatedAt = types.FromSQLNullTime(createdAt)
	return s, nil
}

func (a *Adapter) GetStandupLogsByProject(ctx context.Context, projectID int64) ([]types.StandupLog, error) {
	rows, err := a.db.QueryContext(ctx, getStandupLogsByProjectSQL, projectID)
	if err != nil {
		return nil, fmt.Errorf("get standup logs for project %d: %w", projectID, err)
	}
	defer func() { _ = rows.Close() }()
	var logs []types.StandupLog
	for rows.Next() {
		var s types.StandupLog
		var createdAt sql.NullTime
		if err := rows.Scan(&s.ID, &s.ProjectID, &s.Content, &createdAt); err != nil {
			return nil, fmt.Errorf("scan standup log for project %d: %w", projectID, err)
		}
		s.CreatedAt = types.FromSQLNullTime(createdAt)
		logs = append(logs, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate standup logs for project %d: %w", projectID, err)
	}
	return logs, nil
}

func (a *Adapter) DeleteStandupLog(ctx context.Context, id int64) error {
	_, err := a.db.ExecContext(ctx, deleteStandupLogSQL, id)
	if err != nil {
		return fmt.Errorf("delete standup log %d: %w", id, err)
	}
	return nil
}
