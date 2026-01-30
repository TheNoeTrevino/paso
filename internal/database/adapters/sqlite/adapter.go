package sqlite

import (
	"github.com/thenoetrevino/paso/internal/database/generated_sqlite"
	"github.com/thenoetrevino/paso/internal/database/types"
)

// Adapter wraps the generated SQLite queries and implements the types.Querier interface
// by converting between database-agnostic types and SQLite-specific generated types.
type Adapter struct {
	queries *generated_sqlite.Queries
	db      types.DBTX
}

// New creates a new SQLite adapter that implements types.Querier.
func New(db types.DBTX) *Adapter {
	return &Adapter{
		queries: generated_sqlite.New(db),
		db:      db,
	}
}

// Compile-time check that Adapter implements types.Querier interface
var _ types.Querier = (*Adapter)(nil)
