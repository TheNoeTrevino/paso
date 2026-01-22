package postgres

import (
	"context"

	"github.com/thenoetrevino/paso/internal/database/types"
)

const createTaskEventSQL = `INSERT INTO task_events (task_id, content, author) VALUES ($1, $2, $3) RETURNING *`
const getEventsByTaskSQL = `SELECT id, task_id, content, author, created_at FROM task_events WHERE task_id = $1 ORDER BY created_at DESC`
const deleteEventsByTaskSQL = `DELETE FROM task_events WHERE task_id = $1`

func (a *Adapter) CreateTaskEvent(ctx context.Context, arg types.CreateTaskEventParams) (types.TaskEvent, error) {
	row := a.db.QueryRowContext(ctx, createTaskEventSQL, arg.TaskID, arg.Content, arg.Author)
	var e types.TaskEvent
	err := row.Scan(&e.ID, &e.TaskID, &e.Content, &e.Author, &e.CreatedAt)
	return e, err
}

func (a *Adapter) GetEventsByTask(ctx context.Context, taskID int64) ([]types.TaskEvent, error) {
	rows, err := a.db.QueryContext(ctx, getEventsByTaskSQL, taskID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var events []types.TaskEvent
	for rows.Next() {
		var e types.TaskEvent
		if err := rows.Scan(&e.ID, &e.TaskID, &e.Content, &e.Author, &e.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

func (a *Adapter) DeleteEventsByTask(ctx context.Context, taskID int64) error {
	_, err := a.db.ExecContext(ctx, deleteEventsByTaskSQL, taskID)
	return err
}
