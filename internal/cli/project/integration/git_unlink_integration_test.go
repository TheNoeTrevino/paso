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

func TestGitUnlink_PositionalArgument(t *testing.T) {
	t.Parallel()
	db, app := cli.SetupCLITest(t)
	defer func() {
		err := db.Close()
		assert.NoError(t, err)
	}()

	t.Run("unlink project with positional project ID", func(t *testing.T) {
		projectID := cli.CreateTestProjectWithBranch(t, db, "Test Project", "test-branch")

		cmd := project.GitUnlinkCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			strconv.Itoa(projectID),
			"--quiet",
		})

		assert.NoError(t, err)
		projectIDStr := strings.TrimSpace(output)
		assert.Regexp(t, `^\d+$`, projectIDStr)

		var gitBranch sql.NullString
		err = db.QueryRowContext(context.Background(),
			"SELECT git_branch FROM projects WHERE id = ?", projectID).Scan(&gitBranch)
		assert.NoError(t, err)
		assert.False(t, gitBranch.Valid, "Git branch should be NULL after unlinking")
	})
}

func TestGitUnlink_FlagArgument(t *testing.T) {
	t.Parallel()
	db, app := cli.SetupCLITest(t)
	defer func() {
		err := db.Close()
		assert.NoError(t, err)
	}()

	t.Run("unlink project with --id flag", func(t *testing.T) {
		projectID := cli.CreateTestProjectWithBranch(t, db, "Flag Test Project", "flag-test-branch")

		cmd := project.GitUnlinkCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", strconv.Itoa(projectID),
			"--quiet",
		})

		assert.NoError(t, err)
		projectIDStr := strings.TrimSpace(output)
		assert.Regexp(t, `^\d+$`, projectIDStr)

		var gitBranch sql.NullString
		err = db.QueryRowContext(context.Background(),
			"SELECT git_branch FROM projects WHERE id = ?", projectID).Scan(&gitBranch)
		assert.NoError(t, err)
		assert.False(t, gitBranch.Valid, "Git branch should be NULL after unlinking")
	})
}

func TestGitUnlink_ErrorCases(t *testing.T) {
	t.Parallel()
	db, app := cli.SetupCLITest(t)
	defer func() {
		err := db.Close()
		assert.NoError(t, err)
	}()

	t.Run("invalid project ID returns error", func(t *testing.T) {
		cmd := project.GitUnlinkCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"invalid-id",
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be a number")
	})

	t.Run("too many arguments returns error", func(t *testing.T) {
		cmd := project.GitUnlinkCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"1",
			"2",
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "at most 1 arg(s)")
	})
}

func TestGitUnlink_AlreadyUnlinked(t *testing.T) {
	t.Parallel()
	db, app := cli.SetupCLITest(t)
	defer func() {
		err := db.Close()
		assert.NoError(t, err)
	}()

	t.Run("unlinking project with no branch shows warning", func(t *testing.T) {
		projectID := cli.CreateTestProject(t, db, "Unlinked Project")

		cmd := project.GitUnlinkCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			strconv.Itoa(projectID),
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "no branch linked")
	})
}

func TestGitUnlink_JSONOutput(t *testing.T) {
	t.Parallel()
	db, app := cli.SetupCLITest(t)
	defer func() {
		err := db.Close()
		assert.NoError(t, err)
	}()

	t.Run("jSON output contains expected fields", func(t *testing.T) {
		projectID := cli.CreateTestProjectWithBranch(t, db, "JSON Test Project", "json-test-branch")

		cmd := project.GitUnlinkCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			strconv.Itoa(projectID),
			"--json",
		})

		assert.NoError(t, err)
		assert.Contains(t, output, `"success":true`)
		assert.Contains(t, output, `"json-test-branch"`)
		assert.Contains(t, output, `"unlinked_branch"`)
	})
}
