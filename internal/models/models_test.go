package models

import (
	"testing"
)

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
