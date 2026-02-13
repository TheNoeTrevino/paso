package types

import (
	"database/sql"
	"time"
)

// NullInt64 is a database-agnostic nullable integer type.
// It unifies SQLite's any and PostgreSQL's sql.NullInt64 representations.
type NullInt64 struct {
	Int64 int64
	Valid bool // Valid is true if Int64 is not NULL
}

// FromSQLNullInt64 converts sql.NullInt64 to types.NullInt64
func FromSQLNullInt64(n sql.NullInt64) NullInt64 {
	return NullInt64{Int64: n.Int64, Valid: n.Valid}
}

// FromSQLNullInt32 converts sql.NullInt32 to types.NullInt64
func FromSQLNullInt32(n sql.NullInt32) NullInt64 {
	return NullInt64{Int64: int64(n.Int32), Valid: n.Valid}
}

// NullInt64FromInterface converts any (SQLite representation) to types.NullInt64.
// Handles int64, int32, and int types that databases may return for nullable integer columns.
func NullInt64FromInterface(v any) NullInt64 {
	if v == nil {
		return NullInt64{Valid: false}
	}
	switch i := v.(type) {
	case int64:
		return NullInt64{Int64: i, Valid: true}
	case int32:
		return NullInt64{Int64: int64(i), Valid: true}
	case int:
		return NullInt64{Int64: int64(i), Valid: true}
	default:
		return NullInt64{Valid: false}
	}
}

// ToSQLNullInt64 converts types.NullInt64 to sql.NullInt64
func (n NullInt64) ToSQLNullInt64() sql.NullInt64 {
	return sql.NullInt64{Int64: n.Int64, Valid: n.Valid}
}

// ToSQLNullInt32 converts types.NullInt64 to sql.NullInt32
func (n NullInt64) ToSQLNullInt32() sql.NullInt32 {
	return sql.NullInt32{Int32: int32(n.Int64), Valid: n.Valid}
}

// ToInterface converts types.NullInt64 to any (for SQLite)
func (n NullInt64) ToInterface() any {
	if !n.Valid {
		return nil
	}
	return n.Int64
}

// NullString is a database-agnostic nullable string type.
type NullString struct {
	String string
	Valid  bool // Valid is true if String is not NULL
}

// FromSQLNullString converts sql.NullString to types.NullString
func FromSQLNullString(n sql.NullString) NullString {
	return NullString{String: n.String, Valid: n.Valid}
}

// NullStringFromInterface converts any (SQLite representation) to types.NullString.
// Handles string types that databases may return for nullable string columns.
func NullStringFromInterface(v any) NullString {
	if v == nil {
		return NullString{Valid: false}
	}
	switch s := v.(type) {
	case string:
		return NullString{String: s, Valid: true}
	case []byte:
		return NullString{String: string(s), Valid: true}
	default:
		return NullString{Valid: false}
	}
}

// ToSQLNullString converts types.NullString to sql.NullString
func (n NullString) ToSQLNullString() sql.NullString {
	return sql.NullString{String: n.String, Valid: n.Valid}
}

// ToInterface converts types.NullString to any (for SQLite)
func (n NullString) ToInterface() any {
	if !n.Valid {
		return nil
	}
	return n.String
}

// NullTime is a database-agnostic nullable time type.
type NullTime struct {
	Time  time.Time
	Valid bool // Valid is true if Time is not NULL
}

// FromSQLNullTime converts sql.NullTime to types.NullTime
func FromSQLNullTime(n sql.NullTime) NullTime {
	return NullTime{Time: n.Time, Valid: n.Valid}
}

// ToSQLNullTime converts types.NullTime to sql.NullTime
func (n NullTime) ToSQLNullTime() sql.NullTime {
	return sql.NullTime{Time: n.Time, Valid: n.Valid}
}

// ToInterface converts types.NullTime to any (for SQLite)
func (n NullTime) ToInterface() any {
	if !n.Valid {
		return nil
	}
	return n.Time
}

// ConvertSlice converts a slice of one type to another using the provided converter function.
// This generic helper eliminates boilerplate slice conversion functions in adapters.
//
// Example usage:
//
//	projects, err := a.q.GetAllProjects(ctx)
//	return types.ConvertSlice(projects, fromGeneratedProject), err
func ConvertSlice[From, To any](items []From, convert func(From) To) []To {
	if items == nil {
		return nil
	}
	result := make([]To, len(items))
	for i, item := range items {
		result[i] = convert(item)
	}
	return result
}
