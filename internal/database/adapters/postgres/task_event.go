package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/thenoetrevino/paso/internal/database/types"
)

const createTaskEventSQL = `INSERT INTO task_events (task_id, content, author) VALUES ($1, $2, $3) RETURNING id, task_id, content, author, created_at`
const getEventsByTaskSQL = `SELECT id, task_id, content, author, created_at FROM task_events WHERE task_id = $1 ORDER BY created_at DESC`
const deleteEventsByTaskSQL = `DELETE FROM task_events WHERE task_id = $1`

func (a *Adapter) CreateTaskEvent(ctx context.Context, arg types.CreateTaskEventParams) (types.TaskEvent, error) {
	row := a.db.QueryRowContext(ctx, createTaskEventSQL, arg.TaskID, arg.Content, arg.Author)
	var e types.TaskEvent
	var createdAt sql.NullTime
	err := row.Scan(&e.ID, &e.TaskID, &e.Content, &e.Author, &createdAt)
	if err != nil {
		return types.TaskEvent{}, fmt.Errorf("create task event for task %d: %w", arg.TaskID, err)
	}
	e.CreatedAt = types.FromSQLNullTime(createdAt)
	return e, nil
}

func (a *Adapter) GetEventsByTask(ctx context.Context, taskID int64) ([]types.TaskEvent, error) {
	rows, err := a.db.QueryContext(ctx, getEventsByTaskSQL, taskID)
	if err != nil {
		return nil, fmt.Errorf("get events for task %d: %w", taskID, err)
	}
	defer func() { _ = rows.Close() }()
	var events []types.TaskEvent
	for rows.Next() {
		var e types.TaskEvent
		var createdAt sql.NullTime
		if err := rows.Scan(&e.ID, &e.TaskID, &e.Content, &e.Author, &createdAt); err != nil {
			return nil, fmt.Errorf("scan task event for task %d: %w", taskID, err)
		}
		e.CreatedAt = types.FromSQLNullTime(createdAt)
		events = append(events, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate events for task %d: %w", taskID, err)
	}
	return events, nil
}

func (a *Adapter) DeleteEventsByTask(ctx context.Context, taskID int64) error {
	_, err := a.db.ExecContext(ctx, deleteEventsByTaskSQL, taskID)
	if err != nil {
		return fmt.Errorf("delete events for task %d: %w", taskID, err)
	}
	return nil
}
