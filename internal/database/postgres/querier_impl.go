package postgres

import (
	"context"

	"github.com/thenoetrevino/paso/internal/database/types"
)

const createTaskEvent = `INSERT INTO task_events (task_id, content, author) VALUES ($1, $2, $3) RETURNING *`
const getEventsByTask = `SELECT id, task_id, content, author, created_at FROM task_events WHERE task_id = $1 ORDER BY created_at DESC`
const deleteEventsByTask = `DELETE FROM task_events WHERE task_id = $1`

type Queries struct {
	db types.DBTX
}

func New(db types.DBTX) *Queries {
	return &Queries{db: db}
}

func (q *Queries) CreateTaskEvent(ctx context.Context, arg types.CreateTaskEventParams) (types.TaskEvent, error) {
	row := q.db.QueryRowContext(ctx, createTaskEvent, arg.TaskID, arg.Content, arg.Author)
	var e types.TaskEvent
	err := row.Scan(&e.ID, &e.TaskID, &e.Content, &e.Author, &e.CreatedAt)
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
		if err := rows.Scan(&e.ID, &e.TaskID, &e.Content, &e.Author, &e.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

func (q *Queries) DeleteEventsByTask(ctx context.Context, taskID int64) error {
	_, err := q.db.ExecContext(ctx, deleteEventsByTask, taskID)
	return err
}
