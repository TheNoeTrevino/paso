package models

import (
	"errors"
	"testing"
)

// ============================================================================
// Error Tests
// ============================================================================

func TestErrors_Defined(t *testing.T) {
	// Test that all error variables are defined and not nil
	if ErrAlreadyFirstTask == nil {
		t.Error("ErrAlreadyFirstTask should not be nil")
	}
	if ErrAlreadyLastTask == nil {
		t.Error("ErrAlreadyLastTask should not be nil")
	}
	if ErrAlreadyLastColumn == nil {
		t.Error("ErrAlreadyLastColumn should not be nil")
	}
	if ErrAlreadyFirstColumn == nil {
		t.Error("ErrAlreadyFirstColumn should not be nil")
	}
}

func TestErrors_Messages(t *testing.T) {
	tests := []struct {
		err             error
		expectedMessage string
	}{
		{ErrAlreadyFirstTask, "task is already at the top of the column"},
		{ErrAlreadyLastTask, "task is already at the bottom of the column"},
		{ErrAlreadyLastColumn, "task is already in the last column"},
		{ErrAlreadyFirstColumn, "task is already in the first column"},
	}

	for _, tt := range tests {
		if tt.err.Error() != tt.expectedMessage {
			t.Errorf("Expected error message '%s', got '%s'", tt.expectedMessage, tt.err.Error())
		}
	}
}

func TestErrors_Unique(t *testing.T) {
	// Ensure each error is distinct
	if errors.Is(ErrAlreadyFirstTask, ErrAlreadyLastTask) {
		t.Error("ErrAlreadyFirstTask should not equal ErrAlreadyLastTask")
	}
	if errors.Is(ErrAlreadyFirstColumn, ErrAlreadyLastColumn) {
		t.Error("ErrAlreadyFirstColumn should not equal ErrAlreadyLastColumn")
	}
}

// ============================================================================
// Struct Tests
// ============================================================================

func TestPriority_Creation(t *testing.T) {
	p := Priority{
		ID:          4,
		Description: "high",
		Color:       "#FF0000",
	}

	if p.ID != 4 {
		t.Errorf("Expected ID 4, got %d", p.ID)
	}
	if p.Description != "high" {
		t.Errorf("Expected description 'high', got '%s'", p.Description)
	}
	if p.Color != "#FF0000" {
		t.Errorf("Expected color '#FF0000', got '%s'", p.Color)
	}
}

func TestType_Creation(t *testing.T) {
	typ := Type{
		ID:          2,
		Description: "feature",
	}

	if typ.ID != 2 {
		t.Errorf("Expected ID 2, got %d", typ.ID)
	}
	if typ.Description != "feature" {
		t.Errorf("Expected description 'feature', got '%s'", typ.Description)
	}
}

func TestRelationType_Creation(t *testing.T) {
	rt := RelationType{
		ID:         2,
		PToCLabel:  "Blocked By",
		CToPLabel:  "Blocker",
		Color:      "#EF4444",
		IsBlocking: true,
	}

	if rt.ID != 2 {
		t.Errorf("Expected ID 2, got %d", rt.ID)
	}
	if rt.PToCLabel != "Blocked By" {
		t.Errorf("Expected PToCLabel 'Blocked By', got '%s'", rt.PToCLabel)
	}
	if rt.CToPLabel != "Blocker" {
		t.Errorf("Expected CToPLabel 'Blocker', got '%s'", rt.CToPLabel)
	}
	if rt.Color != "#EF4444" {
		t.Errorf("Expected color '#EF4444', got '%s'", rt.Color)
	}
	if !rt.IsBlocking {
		t.Error("Expected IsBlocking to be true")
	}
}

func TestRelationType_NonBlocking(t *testing.T) {
	rt := RelationType{
		ID:         1,
		PToCLabel:  "Parent",
		CToPLabel:  "Child",
		Color:      "#6B7280",
		IsBlocking: false,
	}

	if rt.IsBlocking {
		t.Error("Expected IsBlocking to be false")
	}
}
