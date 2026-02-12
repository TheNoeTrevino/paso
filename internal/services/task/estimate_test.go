package task

import (
	"testing"
)

func TestParseEstimate(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      []EstimateComponent
		wantError bool
	}{
		{
			name:  "single hour component",
			input: "4h",
			want: []EstimateComponent{
				{Quantity: 4, Unit: "h"},
			},
			wantError: false,
		},
		{
			name:  "single day component",
			input: "2d",
			want: []EstimateComponent{
				{Quantity: 2, Unit: "d"},
			},
			wantError: false,
		},
		{
			name:  "single week component",
			input: "1w",
			want: []EstimateComponent{
				{Quantity: 1, Unit: "w"},
			},
			wantError: false,
		},
		{
			name:  "single month component",
			input: "3m",
			want: []EstimateComponent{
				{Quantity: 3, Unit: "m"},
			},
			wantError: false,
		},
		{
			name:  "week and day combination",
			input: "1w2d",
			want: []EstimateComponent{
				{Quantity: 1, Unit: "w"},
				{Quantity: 2, Unit: "d"},
			},
			wantError: false,
		},
		{
			name:  "week day and hour combination",
			input: "1w2d3h",
			want: []EstimateComponent{
				{Quantity: 1, Unit: "w"},
				{Quantity: 2, Unit: "d"},
				{Quantity: 3, Unit: "h"},
			},
			wantError: false,
		},
		{
			name:  "all units combination",
			input: "1w2d3h4m",
			want: []EstimateComponent{
				{Quantity: 1, Unit: "w"},
				{Quantity: 2, Unit: "d"},
				{Quantity: 3, Unit: "h"},
				{Quantity: 4, Unit: "m"},
			},
			wantError: false,
		},
		{
			name:  "day and hour combination",
			input: "5d6h",
			want: []EstimateComponent{
				{Quantity: 5, Unit: "d"},
				{Quantity: 6, Unit: "h"},
			},
			wantError: false,
		},
		{
			name:  "large quantities",
			input: "100w200d",
			want: []EstimateComponent{
				{Quantity: 100, Unit: "w"},
				{Quantity: 200, Unit: "d"},
			},
			wantError: false,
		},
		{
			name:      "empty string",
			input:     "",
			want:      nil,
			wantError: true,
		},
		{
			name:      "invalid characters",
			input:     "1w2d3x",
			want:      nil,
			wantError: true,
		},
		{
			name:      "no number before unit",
			input:     "w2d",
			want:      nil,
			wantError: true,
		},
		{
			name:      "number without unit",
			input:     "1w2",
			want:      nil,
			wantError: true,
		},
		{
			name:      "duplicate unit",
			input:     "1w2w",
			want:      nil,
			wantError: true,
		},
		{
			name:      "spaces in string",
			input:     "1w 2d",
			want:      nil,
			wantError: true,
		},
		{
			name:      "zero quantity",
			input:     "0d",
			want:      nil,
			wantError: true,
		},
		{
			name:      "negative quantity",
			input:     "-1d",
			want:      nil,
			wantError: true,
		},
		{
			name:      "uppercase units",
			input:     "1W2D",
			want:      nil,
			wantError: true,
		},
		{
			name:      "random text",
			input:     "hello",
			want:      nil,
			wantError: true,
		},
		{
			name:      "units only",
			input:     "wdhm",
			want:      nil,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseEstimate(tt.input)

			if tt.wantError {
				if err == nil {
					t.Errorf("ParseEstimate(%q) expected error, got nil", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseEstimate(%q) unexpected error: %v", tt.input, err)
				return
			}

			if len(got) != len(tt.want) {
				t.Errorf("ParseEstimate(%q) returned %d components, want %d", tt.input, len(got), len(tt.want))
				return
			}

			for i := range got {
				if got[i].Quantity != tt.want[i].Quantity {
					t.Errorf("ParseEstimate(%q) component %d: got quantity %d, want %d", tt.input, i, got[i].Quantity, tt.want[i].Quantity)
				}
				if got[i].Unit != tt.want[i].Unit {
					t.Errorf("ParseEstimate(%q) component %d: got unit %q, want %q", tt.input, i, got[i].Unit, tt.want[i].Unit)
				}
			}
		})
	}
}

func TestParseEstimate_EdgeCases(t *testing.T) {
	t.Run("very large number", func(t *testing.T) {
		components, err := ParseEstimate("999999w")
		if err != nil {
			t.Errorf("ParseEstimate with large number failed: %v", err)
		}
		if len(components) != 1 {
			t.Errorf("Expected 1 component, got %d", len(components))
		}
		if components[0].Quantity != 999999 {
			t.Errorf("Expected quantity 999999, got %d", components[0].Quantity)
		}
	})

	t.Run("all units in different order", func(t *testing.T) {
		// Order: months, hours, days, weeks (alphabetically by unit)
		components, err := ParseEstimate("4m3h2d1w")
		if err != nil {
			t.Errorf("ParseEstimate with mixed order failed: %v", err)
		}
		if len(components) != 4 {
			t.Errorf("Expected 4 components, got %d", len(components))
		}
		// Verify all components are present
		expectedUnits := map[string]int{"m": 4, "h": 3, "d": 2, "w": 1}
		for _, comp := range components {
			expectedQty, ok := expectedUnits[comp.Unit]
			if !ok {
				t.Errorf("Unexpected unit %q", comp.Unit)
			}
			if comp.Quantity != expectedQty {
				t.Errorf("For unit %q, expected quantity %d, got %d", comp.Unit, expectedQty, comp.Quantity)
			}
		}
	})
}
