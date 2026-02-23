package sqlite

import (
	"context"
	"database/sql"

	"github.com/thenoetrevino/paso/internal/database/types"
)

// sqliteDatetimeFormat is the RFC3339 UTC format that the modernc.org/sqlite driver
// uses internally when storing and comparing datetime values. Using this format for
// query parameters ensures correct string comparison with values stored via
// DEFAULT CURRENT_TIMESTAMP (which the driver also normalizes to this format on read).
const sqliteDatetimeFormat = "2006-01-02T15:04:05Z"

func (a *Adapter) CreateStandupLog(ctx context.Context, arg types.CreateStandupLogParams) (types.StandupLog, error) {
	result, err := a.queries.CreateStandupLog(ctx, toGeneratedCreateStandupLogParams(arg))
	if err != nil {
		return types.StandupLog{}, err
	}
	return fromGeneratedStandupLog(result), nil
}

func (a *Adapter) GetStandupLog(ctx context.Context, id int64) (types.StandupLog, error) {
	result, err := a.queries.GetStandupLog(ctx, id)
	if err != nil {
		return types.StandupLog{}, err
	}
	return fromGeneratedStandupLog(result), nil
}

func (a *Adapter) GetStandupLogsByProject(ctx context.Context, projectID int64) ([]types.StandupLog, error) {
	results, err := a.queries.GetStandupLogsByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return types.ConvertSlice(results, fromGeneratedStandupLog), nil
}

// GetStandupLogsByProjectAndDateRange overrides the generated query to work around a
// modernc.org/sqlite driver incompatibility: the driver serializes sql.NullTime params
// as Go's time.Time.String() format ("2006-01-02 17:13:02 +0000 UTC"), which does not
// compare correctly with datetime values stored by DEFAULT CURRENT_TIMESTAMP (normalized
// to RFC3339 "2006-01-02T15:04:05Z" by the driver on read). Using pre-formatted RFC3339
// strings as query parameters ensures consistent string comparison.
func (a *Adapter) GetStandupLogsByProjectAndDateRange(ctx context.Context, arg types.GetStandupLogsByProjectAndDateRangeParams) ([]types.StandupLog, error) {
	// Use datetime() on both sides to normalize the comparison. The modernc.org/sqlite
	// driver applies internal type conversion to datetime-affinity columns that causes
	// plain string comparison with text parameters to fail even when the strings are
	// identical. Wrapping both sides with datetime() forces consistent comparison.
	const query = `
		select id, project_id, content, created_at
		from standup_logs
		where project_id = ?
		  and datetime(created_at) >= datetime(?)
		  and datetime(created_at) < datetime(?)
		order by created_at desc`

	since := arg.Since.UTC().Format(sqliteDatetimeFormat)
	until := arg.Until.UTC().Format(sqliteDatetimeFormat)

	rows, err := a.db.QueryContext(ctx, query, arg.ProjectID, since, until)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var items []types.StandupLog
	for rows.Next() {
		var id, projectID int64
		var content string
		var createdAt sql.NullTime
		if err := rows.Scan(&id, &projectID, &content, &createdAt); err != nil {
			return nil, err
		}
		items = append(items, types.StandupLog{
			ID:        id,
			ProjectID: projectID,
			Content:   content,
			CreatedAt: types.FromSQLNullTime(createdAt),
		})
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (a *Adapter) DeleteStandupLog(ctx context.Context, id int64) error {
	return a.queries.DeleteStandupLog(ctx, id)
}
