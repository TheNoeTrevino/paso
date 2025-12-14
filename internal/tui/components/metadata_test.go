package components_test

import (
	"testing"

	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui/components"
)

func TestRenderLabelChip(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		label    *models.Label
		selected bool
		want     string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := components.RenderLabelChip(tt.label, tt.selected)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("RenderLabelChip() = %v, want %v", got, tt.want)
			}
		})
	}
}
