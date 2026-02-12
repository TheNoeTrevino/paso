package project_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/cli/project"
	"github.com/thenoetrevino/paso/internal/models"
	cli "github.com/thenoetrevino/paso/internal/testing/cli"
	"github.com/thenoetrevino/paso/internal/testing/mocks"
)

func TestCreateProject(t *testing.T) {
	t.Parallel()

	t.Run("create project", func(t *testing.T) {
		t.Parallel()
		mockProject := mocks.NewMockProjectService()
		mockProject.CreateProjectResult = &models.Project{
			ID:        1,
			Name:      "Backend API",
			CreatedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		}

		output, err := cli.ExecuteCLICommandWithMocks(t, cli.MockServices{
			ProjectService: mockProject,
			GitDetector:    mocks.NewMockGitDetector(),
		}, project.ProjectCmd(), []string{"create", "-t", "Backend API"})

		require.NoError(t, err)
		assert.Contains(t, output, "1")
		assert.Contains(t, output, "Backend API")
		assert.True(t, mockProject.HasCall("CreateProject"))
		assert.Equal(t, 1, mockProject.CallCount("CreateProject"))
	})

	t.Run("create project with --json", func(t *testing.T) {
		t.Parallel()
		mockProject := mocks.NewMockProjectService()
		mockProject.CreateProjectResult = &models.Project{
			ID:        2,
			Name:      "Frontend App",
			CreatedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		}

		output, err := cli.ExecuteCLICommandWithMocks(t, cli.MockServices{
			ProjectService: mockProject,
			GitDetector:    mocks.NewMockGitDetector(),
		}, project.ProjectCmd(), []string{"create", "-t", "Frontend App", "-j"})

		require.NoError(t, err)

		var result map[string]any
		require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(output)), &result),
			"output should be valid JSON: %s", output)
		assert.Equal(t, true, result["success"])
		data, ok := result["data"].(map[string]any)
		require.True(t, ok, "data should be a map")
		assert.Equal(t, float64(2), data["ID"])
		assert.Equal(t, "Frontend App", data["Name"])
	})

	t.Run("create project with --quiet", func(t *testing.T) {
		t.Parallel()
		mockProject := mocks.NewMockProjectService()
		mockProject.CreateProjectResult = &models.Project{
			ID:        3,
			Name:      "Quiet Project",
			CreatedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		}

		output, err := cli.ExecuteCLICommandWithMocks(t, cli.MockServices{
			ProjectService: mockProject,
			GitDetector:    mocks.NewMockGitDetector(),
		}, project.ProjectCmd(), []string{"create", "-t", "Quiet Project", "-q"})

		require.NoError(t, err)
		assert.Equal(t, "3", strings.TrimSpace(output))
	})

	t.Run("create project missing title", func(t *testing.T) {
		t.Parallel()
		mockProject := mocks.NewMockProjectService()

		_, err := cli.ExecuteCLICommandWithMocks(t, cli.MockServices{
			ProjectService: mockProject,
			GitDetector:    mocks.NewMockGitDetector(),
		}, project.ProjectCmd(), []string{"create"})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "required")
		assert.Equal(t, 0, mockProject.CallCount("CreateProject"))
	})
}

func TestListProject(t *testing.T) {
	t.Parallel()

	t.Run("list projects", func(t *testing.T) {
		t.Parallel()
		mockProject := mocks.NewMockProjectService()
		mockProject.GetAllProjectsResult = []*models.Project{
			{ID: 1, Name: "Project Alpha"},
			{ID: 2, Name: "Project Beta"},
		}

		output, err := cli.ExecuteCLICommandWithMocks(t, cli.MockServices{
			ProjectService: mockProject,
			GitDetector:    mocks.NewMockGitDetector(),
		}, project.ProjectCmd(), []string{"list"})

		require.NoError(t, err)
		assert.Contains(t, output, "Project Alpha")
		assert.Contains(t, output, "Project Beta")
		assert.Equal(t, 1, mockProject.CallCount("GetAllProjects"))
	})

	t.Run("list projects with --json", func(t *testing.T) {
		t.Parallel()
		mockProject := mocks.NewMockProjectService()
		mockProject.GetAllProjectsResult = []*models.Project{
			{ID: 1, Name: "Project Alpha"},
			{ID: 2, Name: "Project Beta"},
		}

		output, err := cli.ExecuteCLICommandWithMocks(t, cli.MockServices{
			ProjectService: mockProject,
			GitDetector:    mocks.NewMockGitDetector(),
		}, project.ProjectCmd(), []string{"list", "-j"})

		require.NoError(t, err)

		var result map[string]any
		require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(output)), &result),
			"output should be valid JSON: %s", output)
		assert.Equal(t, true, result["success"])
		projects, ok := result["projects"].([]any)
		require.True(t, ok, "projects should be an array")
		assert.Len(t, projects, 2)
	})

	t.Run("list empty projects", func(t *testing.T) {
		t.Parallel()
		mockProject := mocks.NewMockProjectService()
		mockProject.GetAllProjectsResult = []*models.Project{}

		output, err := cli.ExecuteCLICommandWithMocks(t, cli.MockServices{
			ProjectService: mockProject,
			GitDetector:    mocks.NewMockGitDetector(),
		}, project.ProjectCmd(), []string{"list"})

		require.NoError(t, err)
		assert.Contains(t, output, "No projects found")
	})
}

func TestDeleteProject(t *testing.T) {
	t.Parallel()

	t.Run("delete project with --force", func(t *testing.T) {
		t.Parallel()
		mockProject := mocks.NewMockProjectService()
		mockProject.GetProjectByIDResult = &models.Project{ID: 1, Name: "Doomed Project"}
		mockColumn := mocks.NewMockColumnService()
		mockColumn.GetColumnsByProjectResult = []*models.Column{}
		mockTask := mocks.NewMockTaskService()
		mockTask.GetTaskSummariesByProjectResult = map[int][]*models.TaskSummary{}

		output, err := cli.ExecuteCLICommandWithMocks(t, cli.MockServices{
			ProjectService: mockProject,
			ColumnService:  mockColumn,
			TaskService:    mockTask,
		}, project.ProjectCmd(), []string{"delete", "1", "-f"})

		require.NoError(t, err)
		assert.Contains(t, output, "deleted successfully")
		assert.True(t, mockProject.HasCall("GetProjectByID"))
		assert.True(t, mockProject.HasCall("DeleteProject"))
		assert.Equal(t, 1, mockProject.CallCount("DeleteProject"))
	})

	t.Run("delete project with --json", func(t *testing.T) {
		t.Parallel()
		mockProject := mocks.NewMockProjectService()
		mockProject.GetProjectByIDResult = &models.Project{ID: 5, Name: "JSON Delete"}
		mockColumn := mocks.NewMockColumnService()
		mockColumn.GetColumnsByProjectResult = []*models.Column{}
		mockTask := mocks.NewMockTaskService()
		mockTask.GetTaskSummariesByProjectResult = map[int][]*models.TaskSummary{}

		output, err := cli.ExecuteCLICommandWithMocks(t, cli.MockServices{
			ProjectService: mockProject,
			ColumnService:  mockColumn,
			TaskService:    mockTask,
		}, project.ProjectCmd(), []string{"delete", "5", "-f", "-j"})

		require.NoError(t, err)

		var result map[string]any
		require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(output)), &result),
			"output should be valid JSON: %s", output)
		assert.Equal(t, true, result["success"])
		assert.Equal(t, float64(5), result["project_id"])
	})

	t.Run("delete non-existent project", func(t *testing.T) {
		t.Parallel()
		mockProject := mocks.NewMockProjectService()
		mockProject.GetProjectByIDErr = fmt.Errorf("not found")
		mockColumn := mocks.NewMockColumnService()
		mockTask := mocks.NewMockTaskService()

		_, err := cli.ExecuteCLICommandWithMocks(t, cli.MockServices{
			ProjectService: mockProject,
			ColumnService:  mockColumn,
			TaskService:    mockTask,
		}, project.ProjectCmd(), []string{"delete", "999", "-f"})

		require.Error(t, err)
		assert.True(t, mockProject.HasCall("GetProjectByID"))
		assert.Equal(t, 0, mockProject.CallCount("DeleteProject"),
			"DeleteProject should not be called when project not found")
	})
}
