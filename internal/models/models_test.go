package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPriority_Creation(t *testing.T) {
	t.Parallel()
	p := Priority{
		ID:          4,
		Description: "high",
		Color:       "#FF0000",
	}

	assert.Equal(t, 4, p.ID)
	assert.Equal(t, "high", p.Description)
	assert.Equal(t, "#FF0000", p.Color)
}

func TestType_Creation(t *testing.T) {
	t.Parallel()
	typ := Type{
		ID:          2,
		Description: "feature",
	}

	assert.Equal(t, 2, typ.ID)
	assert.Equal(t, "feature", typ.Description)
}

func TestRelationType_Creation(t *testing.T) {
	t.Parallel()
	rt := RelationType{
		ID:         2,
		PToCLabel:  "Blocked By",
		CToPLabel:  "Blocker",
		Color:      "#EF4444",
		IsBlocking: true,
	}

	assert.Equal(t, 2, rt.ID)
	assert.Equal(t, "Blocked By", rt.PToCLabel)
	assert.Equal(t, "Blocker", rt.CToPLabel)
	assert.Equal(t, "#EF4444", rt.Color)
	assert.True(t, rt.IsBlocking)
}

func TestRelationType_NonBlocking(t *testing.T) {
	t.Parallel()
	rt := RelationType{
		ID:         1,
		PToCLabel:  "Parent",
		CToPLabel:  "Child",
		Color:      "#6B7280",
		IsBlocking: false,
	}

	assert.False(t, rt.IsBlocking)
}
