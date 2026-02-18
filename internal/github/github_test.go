package github

import (
	"fmt"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseIssueJSON_WithComments(t *testing.T) {
	t.Parallel()

	data := []byte(`{
		"number": 101,
		"title": "feature: github issues import",
		"body": "### Description\n\nImport issues from GitHub",
		"comments": [
			{
				"author": {"login": "TheNoeTrevino"},
				"body": "First comment",
				"createdAt": "2026-02-15T09:56:15Z"
			},
			{
				"author": {"login": "contributor"},
				"body": "Second comment",
				"createdAt": "2026-02-16T10:00:00Z"
			}
		]
	}`)

	issue, err := ParseIssueJSON(data)

	require.NoError(t, err)
	assert.Equal(t, 101, issue.Number)
	assert.Equal(t, "feature: github issues import", issue.Title)
	assert.Equal(t, "### Description\n\nImport issues from GitHub", issue.Body)
	require.Len(t, issue.Comments, 2)

	assert.Equal(t, "TheNoeTrevino", issue.Comments[0].Author)
	assert.Equal(t, "First comment", issue.Comments[0].Body)
	assert.Equal(t, time.Date(2026, 2, 15, 9, 56, 15, 0, time.UTC), issue.Comments[0].CreatedAt)

	assert.Equal(t, "contributor", issue.Comments[1].Author)
	assert.Equal(t, "Second comment", issue.Comments[1].Body)
}

func TestParseIssueJSON_NoComments(t *testing.T) {
	t.Parallel()

	data := []byte(`{
		"number": 42,
		"title": "Simple issue",
		"body": "Just a description",
		"comments": []
	}`)

	issue, err := ParseIssueJSON(data)

	require.NoError(t, err)
	assert.Equal(t, 42, issue.Number)
	assert.Equal(t, "Simple issue", issue.Title)
	assert.Equal(t, "Just a description", issue.Body)
	assert.Empty(t, issue.Comments)
}

func TestParseIssueJSON_EmptyBody(t *testing.T) {
	t.Parallel()

	data := []byte(`{
		"number": 1,
		"title": "No body issue",
		"body": "",
		"comments": []
	}`)

	issue, err := ParseIssueJSON(data)

	require.NoError(t, err)
	assert.Equal(t, "No body issue", issue.Title)
	assert.Empty(t, issue.Body)
}

func TestParseIssueJSON_InvalidJSON(t *testing.T) {
	t.Parallel()

	data := []byte(`{not valid json}`)

	issue, err := ParseIssueJSON(data)

	require.Error(t, err)
	assert.Nil(t, issue)
	assert.Contains(t, err.Error(), "failed to parse GitHub issue JSON")
}

func TestParseIssueJSON_MalformedCreatedAt(t *testing.T) {
	t.Parallel()

	data := []byte(`{
		"number": 5,
		"title": "Bad date",
		"body": "",
		"comments": [
			{
				"author": {"login": "user"},
				"body": "comment with bad date",
				"createdAt": "not-a-date"
			}
		]
	}`)

	issue, err := ParseIssueJSON(data)

	require.NoError(t, err)
	require.Len(t, issue.Comments, 1)
	assert.Equal(t, "comment with bad date", issue.Comments[0].Body)
	assert.True(t, issue.Comments[0].CreatedAt.IsZero(), "malformed date should result in zero time")
}

func TestParseIssueJSON_NullComments(t *testing.T) {
	t.Parallel()

	data := []byte(`{
		"number": 10,
		"title": "Null comments",
		"body": "body",
		"comments": null
	}`)

	issue, err := ParseIssueJSON(data)

	require.NoError(t, err)
	assert.Empty(t, issue.Comments)
}

func TestCheckInstalled_GHExists(t *testing.T) {
	original := lookPath
	t.Cleanup(func() { lookPath = original })

	lookPath = exec.LookPath // real lookup — gh is installed in dev/CI

	err := CheckInstalled()
	require.NoError(t, err)
}

func TestCheckInstalled_GHMissing(t *testing.T) {
	original := lookPath
	t.Cleanup(func() { lookPath = original })

	lookPath = func(file string) (string, error) {
		return "", fmt.Errorf("executable file not found in $PATH")
	}

	err := CheckInstalled()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "gh CLI is not installed")
	assert.Contains(t, err.Error(), "https://cli.github.com")
}
