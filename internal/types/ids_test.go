package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIDFromInt(t *testing.T) {
	t.Parallel()

	t.Run("ProjectIDFromInt", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, ProjectID(1), ProjectIDFromInt(1))
		assert.Equal(t, ProjectID(0), ProjectIDFromInt(0))
		assert.Equal(t, ProjectID(-1), ProjectIDFromInt(-1))
	})

	t.Run("ColumnIDFromInt", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, ColumnID(42), ColumnIDFromInt(42))
		assert.Equal(t, ColumnID(0), ColumnIDFromInt(0))
		assert.Equal(t, ColumnID(-5), ColumnIDFromInt(-5))
	})

	t.Run("TaskIDFromInt", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, TaskID(100), TaskIDFromInt(100))
		assert.Equal(t, TaskID(0), TaskIDFromInt(0))
	})

	t.Run("LabelIDFromInt", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, LabelID(7), LabelIDFromInt(7))
		assert.Equal(t, LabelID(0), LabelIDFromInt(0))
	})

	t.Run("TypeIDFromInt", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, TypeID(1), TypeIDFromInt(1))
		assert.Equal(t, TypeID(3), TypeIDFromInt(3))
	})

	t.Run("PriorityIDFromInt", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, PriorityID(5), PriorityIDFromInt(5))
		assert.Equal(t, PriorityID(0), PriorityIDFromInt(0))
	})

	t.Run("RelationTypeIDFromInt", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, RelationTypeID(2), RelationTypeIDFromInt(2))
		assert.Equal(t, RelationTypeID(0), RelationTypeIDFromInt(0))
	})

	t.Run("CommentIDFromInt", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, CommentID(99), CommentIDFromInt(99))
		assert.Equal(t, CommentID(0), CommentIDFromInt(0))
	})
}

func TestIDConstants(t *testing.T) {
	t.Parallel()

	t.Run("task type constants", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, TypeID(1), TaskTypeTask)
		assert.Equal(t, TypeID(2), TaskTypeFeature)
		assert.Equal(t, TypeID(3), TaskTypeBug)
	})

	t.Run("priority constants", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, PriorityID(1), PriorityTrivial)
		assert.Equal(t, PriorityID(2), PriorityLow)
		assert.Equal(t, PriorityID(3), PriorityMedium)
		assert.Equal(t, PriorityID(4), PriorityHigh)
		assert.Equal(t, PriorityID(5), PriorityCritical)
	})

	t.Run("relation type constants", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, RelationTypeID(1), RelationTypeParentChild)
		assert.Equal(t, RelationTypeID(2), RelationTypeBlocking)
		assert.Equal(t, RelationTypeID(3), RelationTypeRelated)
	})
}
