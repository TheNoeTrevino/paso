package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/thenoetrevino/paso/internal/events"
)

// withTx executes a function within a database transaction.
// It automatically handles begin, rollback on error, and commit on success.
func withTx(ctx context.Context, db *sql.DB, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("failed to rollback transaction: %v", err)
		}
	}()

	if err := fn(tx); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// sendEvent sends a database change event notification if the client is available.
// Errors are logged but not returned (fire-and-forget pattern).
func sendEvent(eventClient events.EventPublisher, projectID int) {
	if eventClient != nil {
		if err := eventClient.SendEvent(events.Event{
			Type:      events.EventDatabaseChanged,
			ProjectID: projectID,
		}); err != nil {
			log.Printf("failed to send event for project %d: %v", projectID, err)
		}
	}
}

// getProjectIDFromTable retrieves the project_id for an entity in a given table.
// Common pattern used before sending event notifications.
func getProjectIDFromTable(ctx context.Context, db *sql.DB, table string, entityID int) (int, error) {
	var projectID int
	query := fmt.Sprintf("SELECT project_id FROM %s WHERE id = ?", table)
	err := db.QueryRowContext(ctx, query, entityID).Scan(&projectID)
	if err != nil {
		return 0, fmt.Errorf("failed to get project_id from %s for entity %d: %w", table, entityID, err)
	}
	return projectID, nil
}

// nullInt64ToPtr converts sql.NullInt64 to *int.
// Returns nil if the value is not valid.
func nullInt64ToPtr(nv sql.NullInt64) *int {
	if nv.Valid {
		val := int(nv.Int64)
		return &val
	}
	return nil
}

// NullStringToString converts sql.NullString to string.
// Returns empty string if the value is not valid.
func NullStringToString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

// NullTimeToTime converts sql.NullTime to time.Time.
// Returns zero time if the value is not valid.
func NullTimeToTime(nt sql.NullTime) time.Time {
	if nt.Valid {
		return nt.Time
	}
	return time.Time{}
}

// InterfaceToIntPtr converts interface{} to *int.
// Used for converting SQLC query results to pointer types.
func InterfaceToIntPtr(v interface{}) *int {
	if v == nil {
		return nil
	}
	switch val := v.(type) {
	case int64:
		intVal := int(val)
		return &intVal
	case int:
		return &val
	default:
		return nil
	}
}
