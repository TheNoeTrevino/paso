package project_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thenoetrevino/paso/internal/cli/project"
	"github.com/thenoetrevino/paso/internal/git"
	"github.com/thenoetrevino/paso/internal/testing/cli"
)

func TestGitLink(t *testing.T) {
	t.Parallel()
	db, app, mockGit := cli.SetupCLITestWithGit(t)
	ctx := context.Background()

	// Configure mock: simulate being in a git repo
	mockGit.Info = git.GitInfo{
		IsRepo:        true,
		CurrentBranch: "feature/current",
		IsDetached:    false,
		HasCommits:    true,
	}

	t.Run("link project to explicit branch by ID", func(t *testing.T) {
		t.Parallel()
		projectID := cli.CreateTestProject(t, db, "Link Explicit Branch")

		// Register the branch in the mock
		mockGit.Branches["feature/explicit"] = true

		cmd := project.GitLinkCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", projectID),
			"--branch", "feature/explicit",
			"--quiet",
		})

		assert.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("%d", projectID), strings.TrimSpace(output))

		// Verify branch was linked in DB
		var gitBranch sql.NullString
		err = db.QueryRowContext(ctx,
			"SELECT git_branch FROM projects WHERE id = ?", projectID).Scan(&gitBranch)
		assert.NoError(t, err)
		assert.True(t, gitBranch.Valid)
		assert.Equal(t, "feature/explicit", gitBranch.String)
	})

	t.Run("link project using current branch auto-detection", func(t *testing.T) {
		t.Parallel()
		projectID := cli.CreateTestProject(t, db, "Link Current Branch")

		// The mock's current branch is "feature/current"
		mockGit.Branches["feature/current"] = true

		cmd := project.GitLinkCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", projectID),
			"--quiet",
		})

		assert.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("%d", projectID), strings.TrimSpace(output))

		// Verify the auto-detected branch was linked
		var gitBranch sql.NullString
		err = db.QueryRowContext(ctx,
			"SELECT git_branch FROM projects WHERE id = ?", projectID).Scan(&gitBranch)
		assert.NoError(t, err)
		assert.True(t, gitBranch.Valid)
		assert.Equal(t, "feature/current", gitBranch.String)
	})

	t.Run("link project JSON output", func(t *testing.T) {
		t.Parallel()
		projectID := cli.CreateTestProject(t, db, "Link JSON Project")

		mockGit.Branches["feature/json-branch"] = true

		cmd := project.GitLinkCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", projectID),
			"--branch", "feature/json-branch",
			"--json",
		})

		assert.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)
		assert.True(t, result["success"].(bool))

		project := result["project"].(map[string]any)
		assert.Equal(t, "Link JSON Project", project["name"])
		assert.Equal(t, "feature/json-branch", project["branch"])
	})

	t.Run("force transfer branch from another project", func(t *testing.T) {
		t.Parallel()
		// Create two projects; link the first to the branch
		firstProjectID := cli.CreateTestProject(t, db, "First Owner")
		secondProjectID := cli.CreateTestProject(t, db, "New Owner")

		mockGit.Branches["feature/contested"] = true

		// Link first project to the branch
		cmd := project.GitLinkCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", firstProjectID),
			"--branch", "feature/contested",
			"--quiet",
		})
		assert.NoError(t, err)

		// Now force-link the second project to the same branch
		cmd = project.GitLinkCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", secondProjectID),
			"--branch", "feature/contested",
			"--force",
			"--quiet",
		})
		assert.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("%d", secondProjectID), strings.TrimSpace(output))

		// Verify first project lost the branch
		var firstBranch sql.NullString
		err = db.QueryRowContext(ctx,
			"SELECT git_branch FROM projects WHERE id = ?", firstProjectID).Scan(&firstBranch)
		assert.NoError(t, err)
		assert.False(t, firstBranch.Valid, "First project should have lost the branch")

		// Verify second project now owns it
		var secondBranch sql.NullString
		err = db.QueryRowContext(ctx,
			"SELECT git_branch FROM projects WHERE id = ?", secondProjectID).Scan(&secondBranch)
		assert.NoError(t, err)
		assert.True(t, secondBranch.Valid)
		assert.Equal(t, "feature/contested", secondBranch.String)
	})
}

func TestGitLink_Errors(t *testing.T) {
	t.Parallel()
	_, app := cli.SetupCLITest(t)

	t.Run("non-existent project ID", func(t *testing.T) {
		cmd := project.GitLinkCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", "999999",
			"--branch", "feature/test",
			"--quiet",
		})
		cli.AssertExitError(t, err, 3)
		assert.Contains(t, err.Error(), "project with ID 999999 not found")
	})
}

func TestGitUnlink(t *testing.T) {
	t.Parallel()
	db, app, mockGit := cli.SetupCLITestWithGit(t)
	ctx := context.Background()

	// Configure mock: simulate being in a git repo
	mockGit.Info = git.GitInfo{
		IsRepo:        true,
		CurrentBranch: "feature/unlink-test",
		IsDetached:    false,
		HasCommits:    true,
	}

	t.Run("unlink project with pre-linked branch", func(t *testing.T) {
		t.Parallel()
		projectID := cli.CreateTestProject(t, db, "Linked For Unlink")

		mockGit.Branches["feature/to-unlink"] = true

		// Link the project to a branch first
		cmd := project.GitLinkCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", projectID),
			"--branch", "feature/to-unlink",
			"--quiet",
		})
		assert.NoError(t, err)

		// Verify the branch is linked
		var gitBranch sql.NullString
		err = db.QueryRowContext(ctx,
			"SELECT git_branch FROM projects WHERE id = ?", projectID).Scan(&gitBranch)
		assert.NoError(t, err)
		assert.True(t, gitBranch.Valid)
		assert.Equal(t, "feature/to-unlink", gitBranch.String)

		// Now unlink
		cmd = project.GitUnlinkCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", projectID),
		})
		assert.NoError(t, err)
		assert.Contains(t, output, "Unlinked branch")
		assert.Contains(t, output, "feature/to-unlink")

		// Verify the branch was removed
		err = db.QueryRowContext(ctx,
			"SELECT git_branch FROM projects WHERE id = ?", projectID).Scan(&gitBranch)
		assert.NoError(t, err)
		assert.False(t, gitBranch.Valid, "Branch should be NULL after unlink")
	})

	t.Run("unlink project with no branch linked", func(t *testing.T) {
		t.Parallel()
		projectID := cli.CreateTestProject(t, db, "No Branch Project")

		cmd := project.GitUnlinkCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", projectID),
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "no branch linked")
	})

	t.Run("unlink project with no branch linked JSON", func(t *testing.T) {
		t.Parallel()
		projectID := cli.CreateTestProject(t, db, "No Branch JSON Project")

		cmd := project.GitUnlinkCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", projectID),
			"--json",
		})

		assert.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)
		assert.True(t, result["success"].(bool))
		assert.True(t, result["warning"].(bool))
	})
}

func TestGitUnlink_Errors(t *testing.T) {
	t.Parallel()
	_, app := cli.SetupCLITest(t)

	t.Run("non-existent project ID", func(t *testing.T) {
		cmd := project.GitUnlinkCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", "999999",
			"--quiet",
		})
		cli.AssertExitError(t, err, 3)
		assert.Contains(t, err.Error(), "project with ID 999999 not found")
	})
}
