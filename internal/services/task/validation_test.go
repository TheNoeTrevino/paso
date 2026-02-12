package task

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateEstimate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		estimate *string
		wantErr  bool
		errType  error
		errMsg   string
	}{
		// Valid cases
		{
			name:     "nil estimate is allowed",
			estimate: nil,
			wantErr:  false,
		},
		{
			name:     "empty string is allowed",
			estimate: strPtr(""),
			wantErr:  false,
		},
		{
			name:     "whitespace only is allowed",
			estimate: strPtr("   "),
			wantErr:  false,
		},
		{
			name:     "valid single hour",
			estimate: strPtr("1h"),
			wantErr:  false,
		},
		{
			name:     "valid single day",
			estimate: strPtr("2d"),
			wantErr:  false,
		},
		{
			name:     "valid single week",
			estimate: strPtr("3w"),
			wantErr:  false,
		},
		{
			name:     "valid single month",
			estimate: strPtr("4m"),
			wantErr:  false,
		},
		{
			name:     "valid multiple units - weeks and days",
			estimate: strPtr("1w2d"),
			wantErr:  false,
		},
		{
			name:     "valid multiple units - weeks days hours",
			estimate: strPtr("1w2d3h"),
			wantErr:  false,
		},
		{
			name:     "valid multiple units - all units",
			estimate: strPtr("1w2d3h4m"),
			wantErr:  false,
		},
		{
			name:     "valid multi-digit numbers",
			estimate: strPtr("10h"),
			wantErr:  false,
		},
		{
			name:     "valid complex estimate",
			estimate: strPtr("2w5d8h"),
			wantErr:  false,
		},

		// Invalid cases - wrong units
		{
			name:     "invalid unit x",
			estimate: strPtr("1x"),
			wantErr:  true,
			errType:  ErrInvalidEstimateFormat,
			errMsg:   "no valid components",
		},
		{
			name:     "invalid unit abc",
			estimate: strPtr("abc"),
			wantErr:  true,
			errType:  ErrInvalidEstimateFormat,
			errMsg:   "no valid components",
		},
		{
			name:     "invalid unit y",
			estimate: strPtr("5y"),
			wantErr:  true,
			errType:  ErrInvalidEstimateFormat,
			errMsg:   "no valid components",
		},
		{
			name:     "invalid unit s",
			estimate: strPtr("10s"),
			wantErr:  true,
			errType:  ErrInvalidEstimateFormat,
			errMsg:   "no valid components",
		},

		// Invalid cases - wrong format
		{
			name:     "unit before number",
			estimate: strPtr("d1"),
			wantErr:  true,
			errType:  ErrInvalidEstimateFormat,
			errMsg:   "no valid components",
		},
		{
			name:     "number without unit",
			estimate: strPtr("1"),
			wantErr:  true,
			errType:  ErrInvalidEstimateFormat,
			errMsg:   "no valid components",
		},
		{
			name:     "multiple numbers without units",
			estimate: strPtr("12"),
			wantErr:  true,
			errType:  ErrInvalidEstimateFormat,
			errMsg:   "no valid components",
		},

		// Invalid cases - duplicate units
		{
			name:     "duplicate week units",
			estimate: strPtr("1w2w"),
			wantErr:  true,
			errType:  ErrDuplicateEstimateUnit,
			errMsg:   "'w' appears multiple times",
		},
		{
			name:     "duplicate day units",
			estimate: strPtr("1d2d"),
			wantErr:  true,
			errType:  ErrDuplicateEstimateUnit,
			errMsg:   "'d' appears multiple times",
		},
		{
			name:     "duplicate hour units",
			estimate: strPtr("1h2h"),
			wantErr:  true,
			errType:  ErrDuplicateEstimateUnit,
			errMsg:   "'h' appears multiple times",
		},
		{
			name:     "duplicate month units",
			estimate: strPtr("1m2m"),
			wantErr:  true,
			errType:  ErrDuplicateEstimateUnit,
			errMsg:   "'m' appears multiple times",
		},

		// Invalid cases - mixed issues
		{
			name:     "valid unit then invalid unit",
			estimate: strPtr("1w2x"),
			wantErr:  true,
			errType:  ErrInvalidEstimateFormat,
			errMsg:   "invalid characters",
		},
		{
			name:     "space in estimate",
			estimate: strPtr("1w 2d"),
			wantErr:  true,
			errType:  ErrInvalidEstimateFormat,
			errMsg:   "invalid characters",
		},
		{
			name:     "special characters",
			estimate: strPtr("1w-2d"),
			wantErr:  true,
			errType:  ErrInvalidEstimateFormat,
			errMsg:   "invalid characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateEstimate(tt.estimate)

			if tt.wantErr {
				assert.Error(t, err, "Expected error but got none")
				if tt.errType != nil {
					assert.True(t, errors.Is(err, tt.errType), "Expected error type %v, got %v", tt.errType, err)
				}
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg, "Error message should contain expected text")
				}
			} else {
				assert.NoError(t, err, "Expected no error but got: %v", err)
			}
		})
	}
}

// strPtr is a helper function to create a pointer to a string
func strPtr(s string) *string {
	return &s
}
