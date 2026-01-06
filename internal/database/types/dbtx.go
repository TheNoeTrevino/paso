// Package types provides database-agnostic interfaces and types
package types

import (
	"context"
	"database/sql"
)

// DBTX is a database transaction interface that abstracts over *sql.DB and *sql.Tx.
// This allows adapters to work with both regular database connections and transactions
// using the same code. Both adapters (SQLite and PostgreSQL) use this shared interface
// to ensure consistency across database backends.
type DBTX interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}
