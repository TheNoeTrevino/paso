package gh

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/config/colors"
)

func TestValidateIssueNumber_Valid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected int
	}{
		{"1", 1},
		{"42", 42},
		{"101", 101},
		{"9999", 9999},
	}

	for _, tc := range tests {
		n, err := ValidateIssueNumber(tc.input)
		require.NoError(t, err)
		assert.Equal(t, tc.expected, n)
	}
}

func TestValidateIssueNumber_Invalid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input       string
		errContains string
	}{
		{"", "invalid issue number"},
		{"abc", "invalid issue number"},
		{"0", "must be a positive integer"},
		{"-1", "must be a positive integer"},
		{"1.5", "invalid issue number"},
	}

	for _, tc := range tests {
		_, err := ValidateIssueNumber(tc.input)
		require.Error(t, err)
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
		Title:         "feature: cool thing",
		Description:   "Some description",
		Project:       "Test Project",
		IssueNumber:   101,
		CommentsCount: 3,
	}

	output := result.PrettyPrint(*colors.Default())

	assert.Contains(t, output, "GitHub issue imported")
	assert.Contains(t, output, "42")
	assert.Contains(t, output, "#101")
	assert.Contains(t, output, "feature: cool thing")
	assert.Contains(t, output, "Test Project")
	assert.Contains(t, output, "3 imported")
}

func TestImportResult_PrettyPrint_NoDescription(t *testing.T) {
	t.Parallel()

	result := &importResult{
		TaskID:        1,
		Title:         "No desc",
		Project:       "Proj",
		IssueNumber:   5,
		CommentsCount: 0,
	}

	output := result.PrettyPrint(*colors.Default())

	assert.Contains(t, output, "No desc")
	assert.Contains(t, output, "0 imported")
	assert.NotContains(t, output, "Description")
}
