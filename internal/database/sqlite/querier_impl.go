package sqlite

import (
	"context"
	"database/sql"

	"github.com/thenoetrevino/paso/internal/database/types"
)

const createTaskEvent = `INSERT INTO task_events (task_id, content, author) VALUES (?, ?, ?) RETURNING *`
const getEventsByTask = `SELECT id, task_id, content, author, created_at FROM task_events WHERE task_id = ? ORDER BY created_at DESC`
const deleteEventsByTask = `DELETE FROM task_events WHERE task_id = ?`

type Queries struct {
	db types.DBTX
}

func New(db types.DBTX) *Queries {
	return &Queries{db: db}
}

func (q *Queries) CreateTaskEvent(ctx context.Context, arg types.CreateTaskEventParams) (types.TaskEvent, error) {
	row := q.db.QueryRowContext(ctx, createTaskEvent, arg.TaskID, arg.Content, arg.Author)
	var e types.TaskEvent
	var createdAt sql.NullTime
	err := row.Scan(&e.ID, &e.TaskID, &e.Content, &e.Author, &createdAt)
	e.CreatedAt = types.FromSQLNullTime(createdAt)
	return e, err
}

func (q *Queries) GetEventsByTask(ctx context.Context, taskID int64) ([]types.TaskEvent, error) {
	rows, err := q.db.QueryContext(ctx, getEventsByTask, taskID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var events []types.TaskEvent
	for rows.Next() {
		var e types.TaskEvent
		var createdAt sql.NullTime
		if err := rows.Scan(&e.ID, &e.TaskID, &e.Content, &e.Author, &createdAt); err != nil {
			return nil, err
		}
		e.CreatedAt = types.FromSQLNullTime(createdAt)
		events = append(events, e)
	}
	return events, rows.Err()
}

func (q *Queries) DeleteEventsByTask(ctx context.Context, taskID int64) error {
	_, err := q.db.ExecContext(ctx, deleteEventsByTask, taskID)
	return err
}
