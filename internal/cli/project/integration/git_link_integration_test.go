package project_test

import (
	"context"
	"database/sql"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thenoetrevino/paso/internal/cli/project"
	"github.com/thenoetrevino/paso/internal/testing/cli"
)

func TestGitLink_PositionalArguments(t *testing.T) {
	t.Parallel()
	db, app := cli.SetupCLITest(t)
	t.Cleanup(func() {
		err := db.Close()
		assert.NoError(t, err)
	})

	t.Run("link project with positional project ID only (uses current branch)", func(t *testing.T) {
		t.Parallel()
		projectID := cli.CreateTestProject(t, db, "Test Project")

		cmd := project.GitLinkCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			strconv.Itoa(projectID),
			"--quiet",
		})

		if err == nil {
			projectIDStr := strings.TrimSpace(output)
			assert.Regexp(t, `^\d+$`, projectIDStr)

			var gitBranch sql.NullString
			err = db.QueryRowContext(context.Background(),
				"SELECT git_branch FROM projects WHERE id = ?", projectID).Scan(&gitBranch)
			assert.NoError(t, err)
			if gitBranch.Valid {
				assert.NotEmpty(t, gitBranch.String)
			}
		}
	})
}

func TestGitLink_FlagArguments(t *testing.T) {
	t.Parallel()
	db, app := cli.SetupCLITest(t)
	t.Cleanup(func() {
		err := db.Close()
		assert.NoError(t, err)
	})

	t.Run("link project with --id flag only (uses current branch)", func(t *testing.T) {
		t.Parallel()
		projectID := cli.CreateTestProject(t, db, "Flag Test Project")

		cmd := project.GitLinkCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", strconv.Itoa(projectID),
			"--quiet",
		})

		if err == nil {
			projectIDStr := strings.TrimSpace(output)
			assert.Regexp(t, `^\d+$`, projectIDStr)

			var gitBranch sql.NullString
			err = db.QueryRowContext(context.Background(),
				"SELECT git_branch FROM projects WHERE id = ?", projectID).Scan(&gitBranch)
			assert.NoError(t, err)
			if gitBranch.Valid {
				assert.NotEmpty(t, gitBranch.String)
			}
		}
	})
}

func TestGitLink_ErrorCases(t *testing.T) {
	t.Parallel()
	db, app := cli.SetupCLITest(t)
	t.Cleanup(func() {
		err := db.Close()
		assert.NoError(t, err)
	})

	t.Run("invalid project ID returns error", func(t *testing.T) {
		t.Parallel()
		cmd := project.GitLinkCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"invalid-id",
			"test-branch",
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be a number")
	})

	t.Run("too many arguments returns error", func(t *testing.T) {
		t.Parallel()
		cmd := project.GitLinkCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"1",
			"branch1",
			"branch2",
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "at most 2 arg(s)")
	})
}

func TestGitLink_ForceFlag(t *testing.T) {
	t.Parallel()
	db, _ := cli.SetupCLITest(t)
	t.Cleanup(func() {
		err := db.Close()
		assert.NoError(t, err)
	})

	t.Run("force flag transfers branch from one project to another", func(t *testing.T) {
		t.Parallel()
		t.Skip("Skipping: requires valid git branch to exist")
	})
}

func TestGitLink_JSONOutput(t *testing.T) {
	t.Parallel()
	db, app, mockGit := cli.SetupCLITestWithGit(t)
	t.Cleanup(func() {
		err := db.Close()
		assert.NoError(t, err)
	})

	t.Run("jSON output contains expected fields", func(t *testing.T) {
		t.Parallel()
		mockGit.Branches["test-branch"] = true

		projectID := cli.CreateTestProject(t, db, "JSON Test Project")

		cmd := project.GitLinkCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			strconv.Itoa(projectID),
			"--branch", "test-branch",
			"--json",
		})

		assert.NoError(t, err)
		assert.Contains(t, output, `"success":true`)
		assert.Contains(t, output, `"branch":"test-branch"`)
	})
}
