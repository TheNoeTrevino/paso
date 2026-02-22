// Package github provides types and functionality for fetching GitHub issues.
package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"time"
)

// Issue represents a GitHub issue with its comments.
type Issue struct {
	Number   int
	Title    string
	Body     string
	Comments []Comment
}

// Comment represents a single comment on a GitHub issue.
type Comment struct {
	Author    string
	Body      string
	CreatedAt time.Time
}

// lookPath is the function used to find executables on PATH.
// Exported as a variable so tests can override it.
var lookPath = exec.LookPath

// CheckInstalled verifies that the gh CLI is available on the system PATH.
func CheckInstalled() error {
	_, err := lookPath("gh")
	if err != nil {
		return fmt.Errorf("gh CLI is not installed: install it from https://cli.github.com")
	}
	return nil
}

// IssueFetcher defines the interface for fetching GitHub issues.
type IssueFetcher interface {
	CheckInstalled() error
	FetchIssue(ctx context.Context, issueNumber int) (*Issue, error)
}

// ghIssueFetcher is the real implementation that shells out to the gh CLI.
type ghIssueFetcher struct{}

// NewIssueFetcher returns a real IssueFetcher backed by the gh CLI.
func NewIssueFetcher() IssueFetcher {
	return &ghIssueFetcher{}
}

// CheckInstalled verifies that the gh CLI is available on the system PATH.
func (f *ghIssueFetcher) CheckInstalled() error {
	return CheckInstalled()
}

// FetchIssue fetches a GitHub issue by number using the gh CLI.
func (f *ghIssueFetcher) FetchIssue(ctx context.Context, issueNumber int) (*Issue, error) {
	output, err := runGH(ctx, issueNumber)
	if err != nil {
		return nil, err
	}
	return ParseIssueJSON(output)
}

// runGH executes the gh CLI command and returns the raw JSON output.
func runGH(ctx context.Context, issueNumber int) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "gh", "issue", "view",
		strconv.Itoa(issueNumber),
		"--json", "number,title,body,comments",
	)

	output, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return nil, fmt.Errorf("gh CLI failed: %s", string(exitErr.Stderr))
		}
		var execErr *exec.Error
		if errors.As(err, &execErr) {
			return nil, fmt.Errorf("gh CLI not found: install it from https://cli.github.com")
		}
		return nil, fmt.Errorf("failed to run gh CLI: %w", err)
	}

	return output, nil
}

// ghIssueJSON matches the JSON structure returned by gh issue view.
type ghIssueJSON struct {
	Number   int             `json:"number"`
	Title    string          `json:"title"`
	Body     string          `json:"body"`
	Comments []ghCommentJSON `json:"comments"`
}

type ghCommentJSON struct {
	Author    ghAuthorJSON `json:"author"`
	Body      string       `json:"body"`
	CreatedAt string       `json:"createdAt"`
}

type ghAuthorJSON struct {
	Login string `json:"login"`
}

// ParseIssueJSON parses raw gh CLI JSON output into an Issue.
// Exported for testing without requiring the gh CLI.
func ParseIssueJSON(data []byte) (*Issue, error) {
	var raw ghIssueJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse GitHub issue JSON: %w", err)
	}

	issue := &Issue{
		Number:   raw.Number,
		Title:    raw.Title,
		Body:     raw.Body,
		Comments: make([]Comment, 0, len(raw.Comments)),
	}

	for _, c := range raw.Comments {
		createdAt, err := time.Parse(time.RFC3339, c.CreatedAt)
		if err != nil {
			createdAt = time.Time{}
		}

		issue.Comments = append(issue.Comments, Comment{
			Author:    c.Author.Login,
			Body:      c.Body,
			CreatedAt: createdAt,
		})
	}

	return issue, nil
}
