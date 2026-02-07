package cli

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/testutil"
	testutilcli "github.com/thenoetrevino/paso/internal/testutil/cli"
)

func TestGetLabelByID_Found(t *testing.T) {
	db, appInstance := testutilcli.SetupCLITest(t)

	ctx := context.Background()

	// Create test data
	projectID := testutil.CreateTestProject(t, db, "Test Project")
	labelID := testutil.CreateTestLabel(t, db, projectID, "Bug", "#FF0000")

	// Create CLI instance
	cliInstance := &CLI{
		ctx: ctx,
		App: appInstance,
	}

	// Test finding the label
	label, err := GetLabelByID(ctx, cliInstance, labelID)
	require.NoError(t, err, "Should find label successfully")

	if label.ID != labelID {
		t.Errorf("Expected label ID %d, got %d", labelID, label.ID)
	}
	if label.Name != "Bug" {
		t.Errorf("Expected label name 'Bug', got '%s'", label.Name)
	}
	if label.Color != "#FF0000" {
		t.Errorf("Expected label color '#FF0000', got '%s'", label.Color)
	}
	if label.ProjectID != projectID {
		t.Errorf("Expected project ID %d, got %d", projectID, label.ProjectID)
	}
}

func TestGetLabelByID_Found_MultipleProjects(t *testing.T) {
	db, appInstance := testutilcli.SetupCLITest(t)

	ctx := context.Background()

	// Create multiple projects with labels
	project1ID := testutil.CreateTestProject(t, db, "Project 1")
	testutil.CreateTestLabel(t, db, project1ID, "Label 1", "#FF0000")

	project2ID := testutil.CreateTestProject(t, db, "Project 2")
	label2ID := testutil.CreateTestLabel(t, db, project2ID, "Label 2", "#00FF00")

	// Create CLI instance
	cliInstance := &CLI{
		ctx: ctx,
		App: appInstance,
	}

	// Test finding label from second project
	label, err := GetLabelByID(ctx, cliInstance, label2ID)
	require.NoError(t, err, "Should find label successfully")

	if label.Name != "Label 2" {
		t.Errorf("Expected label name 'Label 2', got '%s'", label.Name)
	}
	if label.ProjectID != project2ID {
		t.Errorf("Expected project ID %d, got %d", project2ID, label.ProjectID)
	}
}

func TestGetLabelByID_NotFound(t *testing.T) {
	_, appInstance := testutilcli.SetupCLITest(t)

	ctx := context.Background()

	// Create CLI instance
	cliInstance := &CLI{
		ctx: ctx,
		App: appInstance,
	}

	// Try to find non-existent label
	_, err := GetLabelByID(ctx, cliInstance, 9999)
	if err == nil {
		t.Fatal("Expected error for non-existent label, got nil")
	}

	// Check error message
	expectedMsg := "label 9999 not found"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestGetLabelByID_EmptyDatabase(t *testing.T) {
	_, appInstance := testutilcli.SetupCLITest(t)

	ctx := context.Background()

	// Create CLI instance (no projects or labels created)
	cliInstance := &CLI{
		ctx: ctx,
		App: appInstance,
	}

	// Try to find label in empty database
	_, err := GetLabelByID(ctx, cliInstance, 1)
	if err == nil {
		t.Fatal("Expected error for label in empty database, got nil")
	}
}
