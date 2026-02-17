package gh_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ghcli "github.com/thenoetrevino/paso/internal/cli/gh"
	"github.com/thenoetrevino/paso/internal/github"
	"github.com/thenoetrevino/paso/internal/models"
	cli "github.com/thenoetrevino/paso/internal/testing/cli"
	"github.com/thenoetrevino/paso/internal/testing/mocks"
)

func setupImportMocks() (*mocks.MockTaskService, *mocks.MockProjectService, *mocks.MockColumnService, *mocks.MockGitHubFetcher, cli.MockServices) {
	mockTask := mocks.NewMockTaskService()
	mockProject := mocks.NewMockProjectService()
	mockColumn := mocks.NewMockColumnService()
	mockGH := mocks.NewMockGitHubFetcher()

	services := cli.MockServices{
		TaskService:    mockTask,
		ProjectService: mockProject,
		ColumnService:  mockColumn,
		GitHubFetcher:  mockGH,
	}

	return mockTask, mockProject, mockColumn, mockGH, services
}

func setDefaultImportResults(mockTask *mocks.MockTaskService, mockProject *mocks.MockProjectService, mockColumn *mocks.MockColumnService, mockGH *mocks.MockGitHubFetcher) {
	mockProject.GetProjectByIDResult = &models.Project{ID: 1, Name: "Test Project"}
	mockColumn.GetColumnsByProjectResult = []*models.Column{
		{ID: 10, Name: "Todo", ProjectID: 1},
	}
	mockTask.CreateTaskResult = &models.Task{
		ID:        42,
		Title:     "feature: github issues import",
		CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	mockTask.CreateCommentResult = &models.Comment{
		ID:        1,
		TaskID:    42,
		Message:   "comment",
		Author:    "user",
		CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	mockGH.FetchIssueResult = &github.Issue{
		Number: 101,
		Title:  "feature: github issues import",
		Body:   "Import issues from GitHub",
		Comments: []github.Comment{
			{
				Author:    "TheNoeTrevino",
				Body:      "First comment",
				CreatedAt: time.Date(2026, 2, 15, 9, 56, 15, 0, time.UTC),
			},
		},
	}
}

func TestImport_BasicImport(t *testing.T) {
	t.Parallel()

	mockTask, mockProject, mockColumn, mockGH, services := setupImportMocks()
	setDefaultImportResults(mockTask, mockProject, mockColumn, mockGH)

	output, err := cli.ExecuteCLICommandWithMocks(t, services, ghcli.GhCmd(),
		[]string{"import", "101", "-p", "1"})

	require.NoError(t, err)
	assert.Contains(t, output, "GitHub issue imported")
	assert.Contains(t, output, "42")
	assert.Contains(t, output, "#101")
	assert.True(t, mockGH.HasCall("FetchIssue"))
	assert.Equal(t, 1, mockTask.CallCount("CreateTask"))
	assert.Equal(t, 1, mockTask.CallCount("CreateComment"))
}

func TestImport_WithMultipleComments(t *testing.T) {
	t.Parallel()

	mockTask, mockProject, mockColumn, mockGH, services := setupImportMocks()
	setDefaultImportResults(mockTask, mockProject, mockColumn, mockGH)

	mockGH.FetchIssueResult = &github.Issue{
		Number: 50,
		Title:  "Multi-comment issue",
		Body:   "Description",
		Comments: []github.Comment{
			{Author: "user1", Body: "Comment 1"},
			{Author: "user2", Body: "Comment 2"},
			{Author: "user3", Body: "Comment 3"},
		},
	}
	mockTask.CreateTaskResult = &models.Task{
		ID:    42,
		Title: "Multi-comment issue",
	}

	output, err := cli.ExecuteCLICommandWithMocks(t, services, ghcli.GhCmd(),
		[]string{"import", "50", "-p", "1"})

	require.NoError(t, err)
	assert.Contains(t, output, "3 imported")
	assert.Equal(t, 3, mockTask.CallCount("CreateComment"))

	calls := mockTask.GetCalls()
	commentAuthors := []string{}
	for _, c := range calls {
		if c.Method == "CreateComment" {
			commentAuthors = append(commentAuthors, c.Args["author"].(string))
		}
	}
	assert.Equal(t, []string{"user1", "user2", "user3"}, commentAuthors)
}

func TestImport_NoComments(t *testing.T) {
	t.Parallel()

	mockTask, mockProject, mockColumn, mockGH, services := setupImportMocks()
	setDefaultImportResults(mockTask, mockProject, mockColumn, mockGH)

	mockGH.FetchIssueResult = &github.Issue{
		Number:   10,
		Title:    "No comments issue",
		Body:     "Just a body",
		Comments: []github.Comment{},
	}
	mockTask.CreateTaskResult = &models.Task{
		ID:    42,
		Title: "No comments issue",
	}

	output, err := cli.ExecuteCLICommandWithMocks(t, services, ghcli.GhCmd(),
		[]string{"import", "10", "-p", "1"})

	require.NoError(t, err)
	assert.Contains(t, output, "0 imported")
	assert.Equal(t, 1, mockTask.CallCount("CreateTask"))
	assert.Equal(t, 0, mockTask.CallCount("CreateComment"))
}

func TestImport_QuietMode(t *testing.T) {
	t.Parallel()

	mockTask, mockProject, mockColumn, mockGH, services := setupImportMocks()
	setDefaultImportResults(mockTask, mockProject, mockColumn, mockGH)

	output, err := cli.ExecuteCLICommandWithMocks(t, services, ghcli.GhCmd(),
		[]string{"import", "101", "-p", "1", "-q"})

	require.NoError(t, err)
	assert.Equal(t, "42", strings.TrimSpace(output))
}

func TestImport_JSONMode(t *testing.T) {
	t.Parallel()

	mockTask, mockProject, mockColumn, mockGH, services := setupImportMocks()
	setDefaultImportResults(mockTask, mockProject, mockColumn, mockGH)

	output, err := cli.ExecuteCLICommandWithMocks(t, services, ghcli.GhCmd(),
		[]string{"import", "101", "-p", "1", "-j"})

	require.NoError(t, err)

	var result map[string]any
	require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(output)), &result),
		"output should be valid JSON: %s", output)

	assert.Equal(t, true, result["success"])
	data, ok := result["data"].(map[string]any)
	require.True(t, ok, "data should be a map")
	assert.Equal(t, float64(42), data["TaskID"])
	assert.Equal(t, float64(101), data["IssueNumber"])
}

func TestImport_FetchError(t *testing.T) {
	t.Parallel()

	mockTask, mockProject, mockColumn, mockGH, services := setupImportMocks()
	setDefaultImportResults(mockTask, mockProject, mockColumn, mockGH)

	mockGH.FetchIssueResult = nil
	mockGH.FetchIssueErr = fmt.Errorf("gh CLI failed: issue not found")

	_, err := cli.ExecuteCLICommandWithMocks(t, services, ghcli.GhCmd(),
		[]string{"import", "999", "-p", "1"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch GitHub issue")
	assert.Equal(t, 0, mockTask.CallCount("CreateTask"))
}

func TestImport_MissingIssueNumber(t *testing.T) {
	t.Parallel()

	_, _, _, _, services := setupImportMocks()

	_, err := cli.ExecuteCLICommandWithMocks(t, services, ghcli.GhCmd(),
		[]string{"import", "-p", "1"})

	require.Error(t, err)
}

func TestImport_InvalidIssueNumber(t *testing.T) {
	t.Parallel()

	_, _, _, _, services := setupImportMocks()

	_, err := cli.ExecuteCLICommandWithMocks(t, services, ghcli.GhCmd(),
		[]string{"import", "abc", "-p", "1"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid issue number")
}

func TestImport_ProjectNotFound(t *testing.T) {
	t.Parallel()

	_, mockProject, _, _, services := setupImportMocks()
	mockProject.GetProjectByIDErr = fmt.Errorf("not found")

	_, err := cli.ExecuteCLICommandWithMocks(t, services, ghcli.GhCmd(),
		[]string{"import", "101", "-p", "999"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to find project")
}
