package task

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateLinkFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		blocker       bool
		related       bool
		expectedError string
	}{
		{
			name:          "both flags set",
			blocker:       true,
			related:       true,
			expectedError: "cannot specify both --blocker and --related flags",
		},
		{
			name:          "only blocker flag",
			blocker:       true,
			related:       false,
			expectedError: "",
		},
		{
			name:          "only related flag",
			blocker:       false,
			related:       true,
			expectedError: "",
		},
		{
			name:          "neither flag",
			blocker:       false,
			related:       false,
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateLinkFlags(tt.blocker, tt.related)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDetermineRelationType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		blocker          bool
		related          bool
		expectedTypeID   int
		expectedTypeName string
	}{
		{
			name:             "blocker flag",
			blocker:          true,
			related:          false,
			expectedTypeID:   2,
			expectedTypeName: "blocking",
		},
		{
			name:             "related flag",
			blocker:          false,
			related:          true,
			expectedTypeID:   3,
			expectedTypeName: "related",
		},
		{
			name:             "neither flag (default parent-child)",
			blocker:          false,
			related:          false,
			expectedTypeID:   1,
			expectedTypeName: "parent-child",
		},
		{
			name:             "both flags (blocker takes precedence)",
			blocker:          true,
			related:          true,
			expectedTypeID:   2,
			expectedTypeName: "blocking",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			typeID, typeName := DetermineRelationType(tt.blocker, tt.related)

			assert.Equal(t, tt.expectedTypeID, typeID)
			assert.Equal(t, tt.expectedTypeName, typeName)
		})
	}
}

func TestFormatLinkOutput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		result   *LinkResult
		expected string
	}{
		{
			name: "parent-child relationship",
			result: &LinkResult{
				ParentID:       5,
				ChildID:        3,
				RelationTypeID: 1,
				RelationName:   "parent-child",
			},
			expected: "Linked task 3 as child of task 5",
		},
		{
			name: "blocking relationship",
			result: &LinkResult{
				ParentID:       10,
				ChildID:        7,
				RelationTypeID: 2,
				RelationName:   "blocking",
			},
			expected: "Created blocking relationship: task 10 is blocked by task 7",
		},
		{
			name: "related relationship",
			result: &LinkResult{
				ParentID:       42,
				ChildID:        99,
				RelationTypeID: 3,
				RelationName:   "related",
			},
			expected: "Created related relationship between task 42 and task 99",
		},
		{
			name: "unknown relationship type defaults to parent-child",
			result: &LinkResult{
				ParentID:       1,
				ChildID:        2,
				RelationTypeID: 999,
				RelationName:   "unknown",
			},
			expected: "Linked task 2 as child of task 1",
		},
		{
			name: "zero task IDs",
			result: &LinkResult{
				ParentID:       0,
				ChildID:        0,
				RelationTypeID: 1,
				RelationName:   "parent-child",
			},
			expected: "Linked task 0 as child of task 0",
		},
		{
			name: "negative task IDs with blocking",
			result: &LinkResult{
				ParentID:       -1,
				ChildID:        -2,
				RelationTypeID: 2,
				RelationName:   "blocking",
			},
			expected: "Created blocking relationship: task -1 is blocked by task -2",
		},
		{
			name: "large task IDs with related",
			result: &LinkResult{
				ParentID:       999999,
				ChildID:        888888,
				RelationTypeID: 3,
				RelationName:   "related",
			},
			expected: "Created related relationship between task 999999 and task 888888",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output := FormatLinkOutput(tt.result)
			assert.Equal(t, tt.expected, output)
		})
	}
}

func TestFormatLinkJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		result   *LinkResult
		expected map[string]any
	}{
		{
			name: "parent-child relationship",
			result: &LinkResult{
				ParentID:       5,
				ChildID:        3,
				RelationTypeID: 1,
				RelationName:   "parent-child",
			},
			expected: map[string]any{
				"success":          true,
				"parent_id":        5,
				"child_id":         3,
				"relation_type_id": 1,
				"relation_type":    "parent-child",
			},
		},
		{
			name: "blocking relationship",
			result: &LinkResult{
				ParentID:       10,
				ChildID:        7,
				RelationTypeID: 2,
				RelationName:   "blocking",
			},
			expected: map[string]any{
				"success":          true,
				"parent_id":        10,
				"child_id":         7,
				"relation_type_id": 2,
				"relation_type":    "blocking",
			},
		},
		{
			name: "related relationship",
			result: &LinkResult{
				ParentID:       42,
				ChildID:        99,
				RelationTypeID: 3,
				RelationName:   "related",
			},
			expected: map[string]any{
				"success":          true,
				"parent_id":        42,
				"child_id":         99,
				"relation_type_id": 3,
				"relation_type":    "related",
			},
		},
		{
			name: "zero task IDs",
			result: &LinkResult{
				ParentID:       0,
				ChildID:        0,
				RelationTypeID: 1,
				RelationName:   "parent-child",
			},
			expected: map[string]any{
				"success":          true,
				"parent_id":        0,
				"child_id":         0,
				"relation_type_id": 1,
				"relation_type":    "parent-child",
			},
		},
		{
			name: "negative task IDs",
			result: &LinkResult{
				ParentID:       -1,
				ChildID:        -2,
				RelationTypeID: 2,
				RelationName:   "blocking",
			},
			expected: map[string]any{
				"success":          true,
				"parent_id":        -1,
				"child_id":         -2,
				"relation_type_id": 2,
				"relation_type":    "blocking",
			},
		},
		{
			name: "large task IDs",
			result: &LinkResult{
				ParentID:       999999,
				ChildID:        888888,
				RelationTypeID: 3,
				RelationName:   "related",
			},
			expected: map[string]any{
				"success":          true,
				"parent_id":        999999,
				"child_id":         888888,
				"relation_type_id": 3,
				"relation_type":    "related",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output := FormatLinkJSON(tt.result)
			assert.Equal(t, tt.expected, output)
		})
	}
}
