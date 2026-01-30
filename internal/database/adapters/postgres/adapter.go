package postgres

import (
	"github.com/thenoetrevino/paso/internal/database/generated_postgres"
	"github.com/thenoetrevino/paso/internal/database/types"
)

// Adapter wraps the generated PostgreSQL queries and implements the types.Querier interface
// by converting between database-agnostic types and PostgreSQL-specific generated types.
type Adapter struct {
	queries *generated_postgres.Queries
	db      types.DBTX
}

// New creates a new PostgreSQL adapter that implements types.Querier.
func New(db types.DBTX) *Adapter {
	return &Adapter{
		queries: generated_postgres.New(db),
		db:      db,
	}
}

// Compile-time check that Adapter implements types.Querier interface
var _ types.Querier = (*Adapter)(nil)
