package database

import (
	"context"

	"github.com/thenoetrevino/paso/internal/models"
)

// ColumnReader defines read operations for columns.
type ColumnReader interface {
	GetColumnsByProject(ctx context.Context, projectID int) ([]*models.Column, error)
	GetColumnByID(ctx context.Context, id int) (*models.Column, error)
}

// ColumnWriter defines write operations for columns.
type ColumnWriter interface {
	CreateColumn(ctx context.Context, name string, projectID int, afterID *int) (*models.Column, error)
	UpdateColumnName(ctx context.Context, id int, name string) error
	DeleteColumn(ctx context.Context, id int) error
}

// ColumnRepository combines all column-related operations.
type ColumnRepository interface {
	ColumnReader
	ColumnWriter
}
