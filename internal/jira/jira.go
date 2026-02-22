// Package jira provides types and functionality for fetching Jira issues.
package jira

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"time"
)

// Issue represents a Jira issue with its comments.
type Issue struct {
	Key      string
	Title    string
	Body     string
	Comments []Comment
}

// Comment represents a single comment on a Jira issue.
type Comment struct {
	Author    string
	Body      string
	CreatedAt time.Time
}

// lookPath is the function used to find executables on PATH.
// Exported as a variable so tests can override it.
var lookPath = exec.LookPath

// CheckInstalled verifies that the jira CLI is available on the system PATH.
func CheckInstalled() error {
	_, err := lookPath("jira")
	if err != nil {
		return fmt.Errorf("jira CLI is not installed: install it from https://github.com/ankitpokhrel/jira-cli")
	}
	return nil
}

// IssueFetcher defines the interface for fetching Jira issues.
type IssueFetcher interface {
	FetchIssue(ctx context.Context, issueKey string) (*Issue, error)
}

// jiraIssueFetcher is the real implementation that shells out to the jira CLI.
type jiraIssueFetcher struct{}

// NewIssueFetcher returns a real IssueFetcher backed by the jira CLI.
func NewIssueFetcher() IssueFetcher {
	return &jiraIssueFetcher{}
}

// FetchIssue fetches a Jira issue by key using the jira CLI.
func (f *jiraIssueFetcher) FetchIssue(ctx context.Context, issueKey string) (*Issue, error) {
	output, err := runJira(ctx, issueKey)
	if err != nil {
		return nil, err
	}
	return ParseIssueJSON(output)
}

// runJira executes the jira CLI command and returns the raw JSON output.
func runJira(ctx context.Context, issueKey string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "jira", "issue", "view", issueKey, "--raw")

	output, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return nil, fmt.Errorf("jira CLI failed: %s", string(exitErr.Stderr))
		}
		var execErr *exec.Error
		if errors.As(err, &execErr) {
			return nil, fmt.Errorf("jira CLI not found: install it from https://github.com/ankitpokhrel/jira-cli")
		}
		return nil, fmt.Errorf("failed to run jira CLI: %w", err)
	}

	return output, nil
}

// jiraIssueJSON matches the top-level structure of the Jira REST API issue response.
type jiraIssueJSON struct {
	Key    string         `json:"key"`
	Fields jiraFieldsJSON `json:"fields"`
}

type jiraFieldsJSON struct {
	Summary     string              `json:"summary"`
	Description string              `json:"description"`
	Comment     *jiraCommentWrapper `json:"comment"`
}

// jiraCommentWrapper wraps the comment array in the Jira API response.
// The Jira REST API nests comments under fields.comment.comments.
type jiraCommentWrapper struct {
	Comments []jiraCommentJSON `json:"comments"`
}

type jiraCommentJSON struct {
	Body    string         `json:"body"`
	Author  jiraAuthorJSON `json:"author"`
	Created string         `json:"created"`
}

type jiraAuthorJSON struct {
	DisplayName string `json:"displayName"`
}

// ParseIssueJSON parses raw jira CLI JSON output into an Issue.
// Exported for testing without requiring the jira CLI.
func ParseIssueJSON(data []byte) (*Issue, error) {
	var raw jiraIssueJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse Jira issue JSON: %w", err)
	}

	issue := &Issue{
		Key:      raw.Key,
		Title:    raw.Fields.Summary,
		Body:     raw.Fields.Description,
		Comments: make([]Comment, 0),
	}

	if raw.Fields.Comment == nil {
		return issue, nil
	}

	issue.Comments = make([]Comment, 0, len(raw.Fields.Comment.Comments))

	for _, c := range raw.Fields.Comment.Comments {
		createdAt, err := time.Parse("2006-01-02T15:04:05.000-0700", c.Created)
		if err != nil {
			createdAt = time.Time{}
		}

		issue.Comments = append(issue.Comments, Comment{
			Author:    c.Author.DisplayName,
			Body:      c.Body,
			CreatedAt: createdAt,
		})
	}

	return issue, nil
}
