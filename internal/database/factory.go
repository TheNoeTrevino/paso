// Package database provides the Querier interface factory for database-agnostic access
package database

import (
	"database/sql"
	"fmt"

	"github.com/thenoetrevino/paso/internal/database/adapters/postgres"
	"github.com/thenoetrevino/paso/internal/database/adapters/sqlite"
	"github.com/thenoetrevino/paso/internal/database/types"
)

// Querier is the database-agnostic interface for all database operations.
// This is an alias to types.Querier, allowing services to depend on database.Querier
// without needing to import the types package directly.
type Querier = types.Querier

// NewQuerier returns the appropriate Querier implementation based on the database type.
//
// This factory function is the entry point to the database abstraction layer. It allows
// services throughout the application to remain completely database-agnostic by depending
// on the Querier interface rather than concrete database implementations.
//
// Parameters:
//   - db: Either *sql.DB (for normal queries) or *sql.Tx (for transactional queries)
//   - dbType: SQLite or PostgreSQL (determines which adapter to use)
//
// Returns:
//   - For SQLite: Returns sqlite.Adapter which implements Querier
//   - For PostgreSQL: Returns postgres.Adapter which implements Querier
//   - Returns an error for unknown database types or invalid db parameter types
//
// Both adapters wrap their respective SQLC-generated code and convert between
// database-specific types and the database-agnostic types.* types.
//
// Example usage:
//
//	db, _ := sql.Open("sqlite3", "paso.db")
//	queries, err := NewQuerier(db, SQLite)
//	if err != nil {
//	    return err
//	}
//	tasks, _ := queries.GetTasksByColumn(ctx, columnID)
//
// Or with PostgreSQL:
//
//	db, _ := sql.Open("postgres", connString)
//	queries, err := NewQuerier(db, PostgreSQL)
//	if err != nil {
//	    return err
//	}
//	tasks, _ := queries.GetTasksByColumn(ctx, columnID)  // Same interface!
func NewQuerier(db any, dbType DatabaseType) (Querier, error) {
	// Handle both *sql.DB and *sql.Tx types
	switch v := db.(type) {
	case *sql.DB:
		switch dbType {
		case SQLite:
			return sqlite.New(v), nil
		case PostgreSQL:
			return postgres.New(v), nil
		default:
			return nil, fmt.Errorf("unknown database type: %s", dbType)
		}
	case *sql.Tx:
		switch dbType {
		case SQLite:
			return sqlite.New(v), nil
		case PostgreSQL:
			return postgres.New(v), nil
		default:
			return nil, fmt.Errorf("unknown database type: %s", dbType)
		}
	default:
		return nil, fmt.Errorf("NewQuerier requires *sql.DB or *sql.Tx, got %T", db)
	}
}

// MustNewQuerier is like NewQuerier but panics on error.
// Use this only in contexts where the database type has already been validated
// (e.g., inside transaction callbacks where dbType was validated at service creation).
func MustNewQuerier(db any, dbType DatabaseType) Querier {
	q, err := NewQuerier(db, dbType)
	if err != nil {
		panic(err)
	}
	return q
}
