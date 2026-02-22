package jira

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
		"key": "PROJ-123",
		"fields": {
			"summary": "Implement login page",
			"description": "We need a login page with email and password fields.",
			"comment": {
				"comments": [
					{
						"author": {"displayName": "Noe Trevino"},
						"body": "Started working on this",
						"created": "2026-02-15T09:56:15.000-0600"
					},
					{
						"author": {"displayName": "Jane Doe"},
						"body": "Looks good so far",
						"created": "2026-02-16T10:00:00.000+0000"
					}
				]
			}
		}
	}`)

	issue, err := ParseIssueJSON(data)

	require.NoError(t, err)
	assert.Equal(t, "PROJ-123", issue.Key)
	assert.Equal(t, "Implement login page", issue.Title)
	assert.Equal(t, "We need a login page with email and password fields.", issue.Body)
	require.Len(t, issue.Comments, 2)

	assert.Equal(t, "Noe Trevino", issue.Comments[0].Author)
	assert.Equal(t, "Started working on this", issue.Comments[0].Body)

	expectedUTC := time.Date(2026, 2, 15, 15, 56, 15, 0, time.UTC)
	assert.Equal(t, expectedUTC, issue.Comments[0].CreatedAt.UTC())

	assert.Equal(t, "Jane Doe", issue.Comments[1].Author)
	assert.Equal(t, "Looks good so far", issue.Comments[1].Body)
}

func TestParseIssueJSON_NoComments(t *testing.T) {
	t.Parallel()

	data := []byte(`{
		"key": "PROJ-42",
		"fields": {
			"summary": "Simple issue",
			"description": "Just a description",
			"comment": {
				"comments": []
			}
		}
	}`)

	issue, err := ParseIssueJSON(data)

	require.NoError(t, err)
	assert.Equal(t, "PROJ-42", issue.Key)
	assert.Equal(t, "Simple issue", issue.Title)
	assert.Equal(t, "Just a description", issue.Body)
	assert.Empty(t, issue.Comments)
}

func TestParseIssueJSON_NullCommentField(t *testing.T) {
	t.Parallel()

	data := []byte(`{
		"key": "PROJ-10",
		"fields": {
			"summary": "No comment field",
			"description": "body text",
			"comment": null
		}
	}`)

	issue, err := ParseIssueJSON(data)

	require.NoError(t, err)
	assert.Equal(t, "PROJ-10", issue.Key)
	assert.Empty(t, issue.Comments)
}

func TestParseIssueJSON_MissingCommentField(t *testing.T) {
	t.Parallel()

	data := []byte(`{
		"key": "PROJ-11",
		"fields": {
			"summary": "Missing comment key",
			"description": "body"
		}
	}`)

	issue, err := ParseIssueJSON(data)

	require.NoError(t, err)
	assert.Equal(t, "PROJ-11", issue.Key)
	assert.Empty(t, issue.Comments)
}

func TestParseIssueJSON_EmptyDescription(t *testing.T) {
	t.Parallel()

	data := []byte(`{
		"key": "PROJ-1",
		"fields": {
			"summary": "No body issue",
			"description": "",
			"comment": {"comments": []}
		}
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
	assert.Contains(t, err.Error(), "failed to parse Jira issue JSON")
}

func TestParseIssueJSON_MalformedCreatedAt(t *testing.T) {
	t.Parallel()

	data := []byte(`{
		"key": "PROJ-5",
		"fields": {
			"summary": "Bad date",
			"description": "",
			"comment": {
				"comments": [
					{
						"author": {"displayName": "user"},
						"body": "comment with bad date",
						"created": "not-a-date"
					}
				]
			}
		}
	}`)

	issue, err := ParseIssueJSON(data)

	require.NoError(t, err)
	require.Len(t, issue.Comments, 1)
	assert.Equal(t, "comment with bad date", issue.Comments[0].Body)
	assert.True(t, issue.Comments[0].CreatedAt.IsZero(), "malformed date should result in zero time")
}

func TestParseIssueJSON_UTCZuluTimestamp(t *testing.T) {
	t.Parallel()

	data := []byte(`{
		"key": "PROJ-99",
		"fields": {
			"summary": "Zulu time issue",
			"description": "body",
			"comment": {
				"comments": [
					{
						"author": {"displayName": "user"},
						"body": "zulu timestamp",
						"created": "2026-02-16T10:00:00Z"
					}
				]
			}
		}
	}`)

	issue, err := ParseIssueJSON(data)

	require.NoError(t, err)
	require.Len(t, issue.Comments, 1)
	expected := time.Date(2026, 2, 16, 10, 0, 0, 0, time.UTC)
	assert.Equal(t, expected, issue.Comments[0].CreatedAt.UTC())
}

func TestCheckInstalled_JiraExists(t *testing.T) {
	original := lookPath
	t.Cleanup(func() { lookPath = original })

	lookPath = func(file string) (string, error) {
		return "/usr/local/bin/jira", nil
	}

	err := CheckInstalled()
	require.NoError(t, err)
}

func TestCheckInstalled_JiraMissing(t *testing.T) {
	original := lookPath
	t.Cleanup(func() { lookPath = original })

	lookPath = func(_ string) (string, error) {
		return "", fmt.Errorf("executable file not found in $PATH")
	}

	err := CheckInstalled()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "jira CLI is not installed")
	assert.Contains(t, err.Error(), "https://github.com/ankitpokhrel/jira-cli")
}

func TestCheckInstalled_UsesJiraExecutable(t *testing.T) {
	original := lookPath
	t.Cleanup(func() { lookPath = original })

	var requestedFile string
	lookPath = func(file string) (string, error) {
		requestedFile = file
		return "", &exec.Error{Name: file, Err: exec.ErrNotFound}
	}

	_ = CheckInstalled()
	assert.Equal(t, "jira", requestedFile)
}
