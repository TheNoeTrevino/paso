package fixtures

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// Dialect abstracts SQL differences between database engines for test helpers.
type Dialect interface {
	// Placeholder returns the bind parameter for position n (1-indexed).
	// SQLite returns "?" (ignoring n), Postgres returns "$1", "$2", etc.
	Placeholder(n int) string

	// InsertReturningID executes an INSERT and returns the new row's ID.
	// SQLite uses ExecContext + LastInsertId, Postgres uses QueryRowContext
	// with a RETURNING id clause.
	InsertReturningID(ctx context.Context, db *sql.DB, query string, args ...any) (int, error)
}

type sqliteDialect struct{}

func (sqliteDialect) Placeholder(_ int) string { return "?" }

func (sqliteDialect) InsertReturningID(ctx context.Context, db *sql.DB, query string, args ...any) (int, error) {
	result, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

type postgresDialect struct{}

func (postgresDialect) Placeholder(n int) string { return fmt.Sprintf("$%d", n) }

func (postgresDialect) InsertReturningID(ctx context.Context, db *sql.DB, query string, args ...any) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if !strings.HasSuffix(strings.TrimSpace(query), "RETURNING id") {
		query = strings.TrimSpace(query) + " RETURNING id"
	}
	var id int
	err := db.QueryRowContext(ctx, query, args...).Scan(&id)
	return id, err
}

// SQLiteDialect returns a Dialect for SQLite databases.
func SQLiteDialect() Dialect { return sqliteDialect{} }

// PostgresDialect returns a Dialect for PostgreSQL databases.
func PostgresDialect() Dialect { return postgresDialect{} }
