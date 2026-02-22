package jira

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/config/colors"
)

func TestValidateIssueKey_Valid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected string
	}{
		{"PROJ-1", "PROJ-1"},
		{"PROJ-123", "PROJ-123"},
		{"AB-42", "AB-42"},
		{"MY_PROJ-9999", "MY_PROJ-9999"},
		{"A2B-100", "A2B-100"},
	}

	for _, tc := range tests {
		key, err := ValidateIssueKey(tc.input)
		require.NoError(t, err)
		assert.Equal(t, tc.expected, key)
	}
}

func TestValidateIssueKey_Invalid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input       string
		errContains string
	}{
		{"", "invalid issue key"},
		{"123", "invalid issue key"},
		{"proj-1", "invalid issue key"},
		{"PROJ", "invalid issue key"},
		{"PROJ-", "invalid issue key"},
		{"-123", "invalid issue key"},
		{"PROJ 123", "invalid issue key"},
		{"PROJ-abc", "invalid issue key"},
	}

	for _, tc := range tests {
		_, err := ValidateIssueKey(tc.input)
		require.Error(t, err, "input: %q", tc.input)
		assert.Contains(t, err.Error(), tc.errContains)
	}
}

func TestImportResult_GetID(t *testing.T) {
	t.Parallel()

	result := &importResult{TaskID: 42}
	assert.Equal(t, 42, result.GetID())
}

func TestImportResult_PrettyPrint(t *testing.T) {
	t.Parallel()

	result := &importResult{
		TaskID:        42,
		Title:         "Implement login page",
		Description:   "Some description",
		Project:       "Test Project",
		IssueKey:      "PROJ-123",
		CommentsCount: 3,
	}

	output := result.PrettyPrint(*colors.Default())

	assert.Contains(t, output, "Jira issue imported")
	assert.Contains(t, output, "42")
	assert.Contains(t, output, "PROJ-123")
	assert.Contains(t, output, "Implement login page")
	assert.Contains(t, output, "Test Project")
	assert.Contains(t, output, "3 imported")
}

func TestImportResult_PrettyPrint_NoDescription(t *testing.T) {
	t.Parallel()

	result := &importResult{
		TaskID:        1,
		Title:         "No desc",
		Project:       "Proj",
		IssueKey:      "AB-5",
		CommentsCount: 0,
	}

	output := result.PrettyPrint(*colors.Default())

	assert.Contains(t, output, "No desc")
	assert.Contains(t, output, "0 imported")
	assert.NotContains(t, output, "Description")
}
