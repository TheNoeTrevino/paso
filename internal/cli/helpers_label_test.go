package cli

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/testutil"
	testutilcli "github.com/thenoetrevino/paso/internal/testutil/cli"
)

func TestGetLabelByID_Found(t *testing.T) {
	db, appInstance := testutilcli.SetupCLITest(t)

	ctx := context.Background()

	// Create test data
	projectID := testutil.CreateTestProject(t, db, testutil.SQLiteDialect(), "Test Project")
	labelID := testutil.CreateTestLabel(t, db, testutil.SQLiteDialect(), projectID, "Bug", "#FF0000")

	// Create CLI instance
	cliInstance := &CLI{
		ctx: ctx,
		App: appInstance,
	}

	// Test finding the label
	label, err := GetLabelByID(ctx, cliInstance, labelID)
	require.NoError(t, err, "Should find label successfully")

	assert.Equal(t, labelID, label.ID)
	assert.Equal(t, "Bug", label.Name)
	assert.Equal(t, "#FF0000", label.Color)
	assert.Equal(t, projectID, label.ProjectID)
}

func TestGetLabelByID_Found_MultipleProjects(t *testing.T) {
	db, appInstance := testutilcli.SetupCLITest(t)

	ctx := context.Background()

	// Create multiple projects with labels
	project1ID := testutil.CreateTestProject(t, db, testutil.SQLiteDialect(), "Project 1")
	testutil.CreateTestLabel(t, db, testutil.SQLiteDialect(), project1ID, "Label 1", "#FF0000")

	project2ID := testutil.CreateTestProject(t, db, testutil.SQLiteDialect(), "Project 2")
	label2ID := testutil.CreateTestLabel(t, db, testutil.SQLiteDialect(), project2ID, "Label 2", "#00FF00")

	// Create CLI instance
	cliInstance := &CLI{
		ctx: ctx,
		App: appInstance,
	}

	// Test finding label from second project
	label, err := GetLabelByID(ctx, cliInstance, label2ID)
	require.NoError(t, err, "Should find label successfully")

	assert.Equal(t, "Label 2", label.Name)
	assert.Equal(t, project2ID, label.ProjectID)
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
	require.Error(t, err)

	// Check error message
	assert.ErrorContains(t, err, "label 9999 not found")
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
	require.Error(t, err)
}
