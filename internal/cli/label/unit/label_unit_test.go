package label_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/cli/label"
	"github.com/thenoetrevino/paso/internal/models"
	cli "github.com/thenoetrevino/paso/internal/testing/cli"
	"github.com/thenoetrevino/paso/internal/testing/mocks"
)

func TestCreateLabel(t *testing.T) {
	t.Parallel()

	mockProject := mocks.NewMockProjectService()
	mockLabel := mocks.NewMockLabelService()

	mockProject.GetProjectByIDResult = &models.Project{ID: 1, Name: "Test Project"}
	mockLabel.CreateLabelResult = &models.Label{ID: 5, Name: "bug", Color: "#ff0000", ProjectID: 1}

	output, err := cli.ExecuteCLICommandWithMocks(t, cli.MockServices{
		ProjectService: mockProject,
		LabelService:   mockLabel,
	}, label.LabelCmd(), []string{"create", "-n", "bug", "-c", "#ff0000", "-p", "1"})

	require.NoError(t, err)
	assert.Contains(t, output, "bug")
	assert.Contains(t, output, "5")
	assert.True(t, mockProject.HasCall("GetProjectByID"))
	assert.Equal(t, 1, mockLabel.CallCount("CreateLabel"))
}

func TestCreateLabel_JSON(t *testing.T) {
	t.Parallel()

	mockProject := mocks.NewMockProjectService()
	mockLabel := mocks.NewMockLabelService()

	mockProject.GetProjectByIDResult = &models.Project{ID: 1, Name: "Test Project"}
	mockLabel.CreateLabelResult = &models.Label{ID: 5, Name: "bug", Color: "#ff0000", ProjectID: 1}

	output, err := cli.ExecuteCLICommandWithMocks(t, cli.MockServices{
		ProjectService: mockProject,
		LabelService:   mockLabel,
	}, label.LabelCmd(), []string{"create", "-n", "bug", "-c", "#ff0000", "-p", "1", "-j"})

	require.NoError(t, err)

	var result map[string]any
	require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(output)), &result), "output should be valid JSON: %s", output)
	assert.Equal(t, true, result["success"])
	data, ok := result["data"].(map[string]any)
	require.True(t, ok, "data should be a map")
	assert.Equal(t, float64(5), data["ID"])
	assert.Equal(t, "bug", data["Name"])
}

func TestCreateLabel_MissingName(t *testing.T) {
	t.Parallel()

	mockProject := mocks.NewMockProjectService()
	mockLabel := mocks.NewMockLabelService()

	_, err := cli.ExecuteCLICommandWithMocks(t, cli.MockServices{
		ProjectService: mockProject,
		LabelService:   mockLabel,
	}, label.LabelCmd(), []string{"create", "-c", "#ff0000", "-p", "1"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "required")
}

func TestListLabel(t *testing.T) {
	t.Parallel()

	mockProject := mocks.NewMockProjectService()
	mockLabel := mocks.NewMockLabelService()

	mockProject.GetProjectByIDResult = &models.Project{ID: 1, Name: "Test Project"}
	mockLabel.GetLabelsByProjectResult = []*models.Label{
		{ID: 5, Name: "bug", Color: "#ff0000", ProjectID: 1},
		{ID: 6, Name: "feature", Color: "#00ff00", ProjectID: 1},
	}

	output, err := cli.ExecuteCLICommandWithMocks(t, cli.MockServices{
		ProjectService: mockProject,
		LabelService:   mockLabel,
	}, label.LabelCmd(), []string{"list", "1"})

	require.NoError(t, err)
	assert.Contains(t, output, "bug")
	assert.Contains(t, output, "feature")
	assert.True(t, mockProject.HasCall("GetProjectByID"))
	assert.Equal(t, 1, mockLabel.CallCount("GetLabelsByProject"))
}

func TestListLabel_JSON(t *testing.T) {
	t.Parallel()

	mockProject := mocks.NewMockProjectService()
	mockLabel := mocks.NewMockLabelService()

	mockProject.GetProjectByIDResult = &models.Project{ID: 1, Name: "Test Project"}
	mockLabel.GetLabelsByProjectResult = []*models.Label{
		{ID: 5, Name: "bug", Color: "#ff0000", ProjectID: 1},
	}

	output, err := cli.ExecuteCLICommandWithMocks(t, cli.MockServices{
		ProjectService: mockProject,
		LabelService:   mockLabel,
	}, label.LabelCmd(), []string{"list", "1", "-j"})

	require.NoError(t, err)

	var result map[string]any
	require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(output)), &result), "output should be valid JSON: %s", output)
	assert.Equal(t, true, result["success"])
	labels, ok := result["labels"].([]any)
	require.True(t, ok, "labels should be an array")
	require.Len(t, labels, 1)
	lbl := labels[0].(map[string]any)
	assert.Equal(t, float64(5), lbl["id"])
	assert.Equal(t, "bug", lbl["name"])
}

func TestUpdateLabel_Name(t *testing.T) {
	t.Parallel()

	mockProject := mocks.NewMockProjectService()
	mockLabel := mocks.NewMockLabelService()

	mockProject.GetAllProjectsResult = []*models.Project{{ID: 1, Name: "Test Project"}}
	mockLabel.GetLabelsByProjectResult = []*models.Label{
		{ID: 5, Name: "bug", Color: "#ff0000", ProjectID: 1},
	}

	output, err := cli.ExecuteCLICommandWithMocks(t, cli.MockServices{
		ProjectService: mockProject,
		LabelService:   mockLabel,
	}, label.LabelCmd(), []string{"update", "5", "-n", "critical-bug"})

	require.NoError(t, err)
	assert.Contains(t, output, "updated")
	assert.True(t, mockProject.HasCall("GetAllProjects"))
	assert.True(t, mockLabel.HasCall("GetLabelsByProject"))
	assert.Equal(t, 1, mockLabel.CallCount("UpdateLabel"))
}

func TestDeleteLabel_Force(t *testing.T) {
	t.Parallel()

	mockProject := mocks.NewMockProjectService()
	mockLabel := mocks.NewMockLabelService()

	mockProject.GetAllProjectsResult = []*models.Project{{ID: 1, Name: "Test Project"}}
	mockLabel.GetLabelsByProjectResult = []*models.Label{
		{ID: 5, Name: "bug", Color: "#ff0000", ProjectID: 1},
	}

	output, err := cli.ExecuteCLICommandWithMocks(t, cli.MockServices{
		ProjectService: mockProject,
		LabelService:   mockLabel,
	}, label.LabelCmd(), []string{"delete", "5", "-f"})

	require.NoError(t, err)
	assert.Contains(t, output, "deleted")
	assert.True(t, mockProject.HasCall("GetAllProjects"))
	assert.True(t, mockLabel.HasCall("GetLabelsByProject"))
	assert.Equal(t, 1, mockLabel.CallCount("DeleteLabel"))
}

func TestAttachLabel(t *testing.T) {
	t.Parallel()

	mockProject := mocks.NewMockProjectService()
	mockLabel := mocks.NewMockLabelService()
	mockTask := mocks.NewMockTaskService()
	mockColumn := mocks.NewMockColumnService()

	mockTask.GetTaskDetailResult = &models.TaskDetail{ID: 10, Title: "Test Task", ColumnID: 20}
	mockColumn.GetColumnByIDResult = &models.Column{ID: 20, ProjectID: 1}
	mockProject.GetAllProjectsResult = []*models.Project{{ID: 1, Name: "Test Project"}}
	mockLabel.GetLabelsByProjectResult = []*models.Label{
		{ID: 5, Name: "bug", Color: "#ff0000", ProjectID: 1},
	}

	output, err := cli.ExecuteCLICommandWithMocks(t, cli.MockServices{
		TaskService:    mockTask,
		ProjectService: mockProject,
		LabelService:   mockLabel,
		ColumnService:  mockColumn,
	}, label.LabelCmd(), []string{"attach", "-t", "10", "-l", "5"})

	require.NoError(t, err)
	assert.Contains(t, output, "bug")
	assert.Contains(t, output, "10")
	assert.True(t, mockTask.HasCall("GetTaskDetail", 10))
	assert.True(t, mockColumn.HasCall("GetColumnByID"))
	assert.True(t, mockTask.HasCall("AttachLabel", 10))
}

func TestDetachLabel(t *testing.T) {
	t.Parallel()

	mockTask := mocks.NewMockTaskService()

	output, err := cli.ExecuteCLICommandWithMocks(t, cli.MockServices{
		TaskService: mockTask,
	}, label.LabelCmd(), []string{"detach", "-t", "10", "-l", "5"})

	require.NoError(t, err)
	assert.Contains(t, output, "10")
	assert.True(t, mockTask.HasCall("DetachLabel", 10))
}

func TestLabelNotFound(t *testing.T) {
	t.Parallel()

	mockProject := mocks.NewMockProjectService()
	mockLabel := mocks.NewMockLabelService()

	mockProject.GetAllProjectsResult = []*models.Project{{ID: 1, Name: "Test Project"}}
	mockLabel.GetLabelsByProjectResult = []*models.Label{} // no labels

	_, err := cli.ExecuteCLICommandWithMocks(t, cli.MockServices{
		ProjectService: mockProject,
		LabelService:   mockLabel,
	}, label.LabelCmd(), []string{"update", "999", "-n", "nope"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), fmt.Sprintf("label %d not found", 999))
}
