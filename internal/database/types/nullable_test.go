package types

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromSQLNullInt64(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    sql.NullInt64
		expected NullInt64
	}{
		{
			name:     "valid value",
			input:    sql.NullInt64{Int64: 42, Valid: true},
			expected: NullInt64{Int64: 42, Valid: true},
		},
		{
			name:     "null value",
			input:    sql.NullInt64{Int64: 0, Valid: false},
			expected: NullInt64{Int64: 0, Valid: false},
		},
		{
			name:     "zero valid value",
			input:    sql.NullInt64{Int64: 0, Valid: true},
			expected: NullInt64{Int64: 0, Valid: true},
		},
		{
			name:     "negative value",
			input:    sql.NullInt64{Int64: -100, Valid: true},
			expected: NullInt64{Int64: -100, Valid: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := FromSQLNullInt64(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFromSQLNullInt32(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    sql.NullInt32
		expected NullInt64
	}{
		{
			name:     "valid value widens to int64",
			input:    sql.NullInt32{Int32: 42, Valid: true},
			expected: NullInt64{Int64: 42, Valid: true},
		},
		{
			name:     "null value",
			input:    sql.NullInt32{Int32: 0, Valid: false},
			expected: NullInt64{Int64: 0, Valid: false},
		},
		{
			name:     "negative value",
			input:    sql.NullInt32{Int32: -10, Valid: true},
			expected: NullInt64{Int64: -10, Valid: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := FromSQLNullInt32(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNullInt64FromInterface(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    any
		expected NullInt64
	}{
		{
			name:     "nil returns invalid",
			input:    nil,
			expected: NullInt64{Valid: false},
		},
		{
			name:     "int64 value",
			input:    int64(42),
			expected: NullInt64{Int64: 42, Valid: true},
		},
		{
			name:     "int32 value widens",
			input:    int32(99),
			expected: NullInt64{Int64: 99, Valid: true},
		},
		{
			name:     "int value widens",
			input:    int(7),
			expected: NullInt64{Int64: 7, Valid: true},
		},
		{
			name:     "unsupported type returns invalid",
			input:    "not a number",
			expected: NullInt64{Valid: false},
		},
		{
			name:     "float64 is unsupported",
			input:    float64(3.14),
			expected: NullInt64{Valid: false},
		},
		{
			name:     "bool is unsupported",
			input:    true,
			expected: NullInt64{Valid: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := NullInt64FromInterface(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNullInt64_ToSQLNullInt64(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    NullInt64
		expected sql.NullInt64
	}{
		{
			name:     "valid value",
			input:    NullInt64{Int64: 42, Valid: true},
			expected: sql.NullInt64{Int64: 42, Valid: true},
		},
		{
			name:     "invalid value",
			input:    NullInt64{Valid: false},
			expected: sql.NullInt64{Valid: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.input.ToSQLNullInt64()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNullInt64_ToSQLNullInt32(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    NullInt64
		expected sql.NullInt32
	}{
		{
			name:     "valid value narrows to int32",
			input:    NullInt64{Int64: 42, Valid: true},
			expected: sql.NullInt32{Int32: 42, Valid: true},
		},
		{
			name:     "invalid value",
			input:    NullInt64{Valid: false},
			expected: sql.NullInt32{Valid: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.input.ToSQLNullInt32()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNullInt64_ToInterface(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    NullInt64
		expected any
	}{
		{
			name:     "valid returns int64 value",
			input:    NullInt64{Int64: 42, Valid: true},
			expected: int64(42),
		},
		{
			name:     "invalid returns nil",
			input:    NullInt64{Valid: false},
			expected: nil,
		},
		{
			name:     "valid zero returns int64 zero",
			input:    NullInt64{Int64: 0, Valid: true},
			expected: int64(0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.input.ToInterface()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNullInt64_RoundTrip(t *testing.T) {
	t.Parallel()

	original := NullInt64{Int64: 123, Valid: true}
	roundTripped := FromSQLNullInt64(original.ToSQLNullInt64())
	assert.Equal(t, original, roundTripped)

	null := NullInt64{Valid: false}
	roundTrippedNull := FromSQLNullInt64(null.ToSQLNullInt64())
	assert.Equal(t, null, roundTrippedNull)
}

func TestFromSQLNullString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    sql.NullString
		expected NullString
	}{
		{
			name:     "valid string",
			input:    sql.NullString{String: "hello", Valid: true},
			expected: NullString{String: "hello", Valid: true},
		},
		{
			name:     "null string",
			input:    sql.NullString{Valid: false},
			expected: NullString{Valid: false},
		},
		{
			name:     "empty valid string",
			input:    sql.NullString{String: "", Valid: true},
			expected: NullString{String: "", Valid: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := FromSQLNullString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNullStringFromInterface(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    any
		expected NullString
	}{
		{
			name:     "nil returns invalid",
			input:    nil,
			expected: NullString{Valid: false},
		},
		{
			name:     "string value",
			input:    "hello",
			expected: NullString{String: "hello", Valid: true},
		},
		{
			name:     "byte slice value",
			input:    []byte("world"),
			expected: NullString{String: "world", Valid: true},
		},
		{
			name:     "unsupported type returns invalid",
			input:    42,
			expected: NullString{Valid: false},
		},
		{
			name:     "empty string is valid",
			input:    "",
			expected: NullString{String: "", Valid: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := NullStringFromInterface(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNullString_ToSQLNullString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    NullString
		expected sql.NullString
	}{
		{
			name:     "valid string",
			input:    NullString{String: "test", Valid: true},
			expected: sql.NullString{String: "test", Valid: true},
		},
		{
			name:     "invalid string",
			input:    NullString{Valid: false},
			expected: sql.NullString{Valid: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.input.ToSQLNullString()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNullString_RoundTrip(t *testing.T) {
	t.Parallel()

	original := NullString{String: "round trip", Valid: true}
	roundTripped := FromSQLNullString(original.ToSQLNullString())
	assert.Equal(t, original, roundTripped)
}

func TestFromSQLNullTime(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name     string
		input    sql.NullTime
		expected NullTime
	}{
		{
			name:     "valid time",
			input:    sql.NullTime{Time: now, Valid: true},
			expected: NullTime{Time: now, Valid: true},
		},
		{
			name:     "null time",
			input:    sql.NullTime{Valid: false},
			expected: NullTime{Valid: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := FromSQLNullTime(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNullTime_ToSQLNullTime(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name     string
		input    NullTime
		expected sql.NullTime
	}{
		{
			name:     "valid time",
			input:    NullTime{Time: now, Valid: true},
			expected: sql.NullTime{Time: now, Valid: true},
		},
		{
			name:     "invalid time",
			input:    NullTime{Valid: false},
			expected: sql.NullTime{Valid: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.input.ToSQLNullTime()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNullTime_RoundTrip(t *testing.T) {
	t.Parallel()

	now := time.Now()
	original := NullTime{Time: now, Valid: true}
	roundTripped := FromSQLNullTime(original.ToSQLNullTime())
	assert.Equal(t, original, roundTripped)
}

func TestConvertSlice(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    []int
		expected []string
	}{
		{
			name:     "nil input returns nil",
			input:    nil,
			expected: nil,
		},
		{
			name:     "empty slice returns empty slice",
			input:    []int{},
			expected: []string{},
		},
		{
			name:     "single element",
			input:    []int{1},
			expected: []string{"1"},
		},
		{
			name:     "multiple elements",
			input:    []int{1, 2, 3},
			expected: []string{"1", "2", "3"},
		},
	}

	converter := func(i int) string {
		return string(rune('0' + i))
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := ConvertSlice(tt.input, converter)

			if tt.input == nil {
				require.Nil(t, result)
			} else {
				require.Len(t, result, len(tt.expected))
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
