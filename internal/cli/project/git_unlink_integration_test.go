package project

import (
	"context"
	"database/sql"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thenoetrevino/paso/internal/testutil/cli"
)

func TestGitUnlink_PositionalArgument(t *testing.T) {
	db, app := cli.SetupCLITest(t)
	defer func() {
		err := db.Close()
		assert.NoError(t, err)
	}()

	t.Run("Unlink project with positional project ID", func(t *testing.T) {
		result, err := db.ExecContext(context.Background(),
			"INSERT INTO projects (name, description, git_branch) VALUES (?, ?, ?)",
			"Test Project", "Test Description", "test-branch")
		assert.NoError(t, err)
		projectID, _ := result.LastInsertId()

		_, err = db.ExecContext(context.Background(),
			"INSERT INTO columns (project_id, name) VALUES (?, 'Todo'), (?, 'In Progress'), (?, 'Done')",
			projectID, projectID, projectID)
		assert.NoError(t, err)

		cmd := GitUnlinkCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			strconv.Itoa(int(projectID)),
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
	db, app := cli.SetupCLITest(t)
	defer func() {
		err := db.Close()
		assert.NoError(t, err)
	}()

	t.Run("Unlink project with --id flag", func(t *testing.T) {
		result, err := db.ExecContext(context.Background(),
			"INSERT INTO projects (name, description, git_branch) VALUES (?, ?, ?)",
			"Flag Test Project", "Test Description", "flag-test-branch")
		assert.NoError(t, err)
		projectID, _ := result.LastInsertId()

		_, err = db.ExecContext(context.Background(),
			"INSERT INTO columns (project_id, name) VALUES (?, 'Todo'), (?, 'In Progress'), (?, 'Done')",
			projectID, projectID, projectID)
		assert.NoError(t, err)

		cmd := GitUnlinkCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", strconv.Itoa(int(projectID)),
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
	db, app := cli.SetupCLITest(t)
	defer func() {
		err := db.Close()
		assert.NoError(t, err)
	}()

	t.Run("Invalid project ID returns error", func(t *testing.T) {
		cmd := GitUnlinkCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"invalid-id",
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be a number")
	})

	t.Run("Too many arguments returns error", func(t *testing.T) {
		cmd := GitUnlinkCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"1",
			"2",
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "at most 1 arg(s)")
	})
}

func TestGitUnlink_AlreadyUnlinked(t *testing.T) {
	db, app := cli.SetupCLITest(t)
	defer func() {
		err := db.Close()
		assert.NoError(t, err)
	}()

	t.Run("Unlinking project with no branch shows warning", func(t *testing.T) {
		result, err := db.ExecContext(context.Background(),
			"INSERT INTO projects (name, description) VALUES (?, ?)",
			"Unlinked Project", "Already unlinked")
		assert.NoError(t, err)
		projectID, _ := result.LastInsertId()

		_, err = db.ExecContext(context.Background(),
			"INSERT INTO columns (project_id, name) VALUES (?, 'Todo'), (?, 'In Progress'), (?, 'Done')",
			projectID, projectID, projectID)
		assert.NoError(t, err)

		cmd := GitUnlinkCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			strconv.Itoa(int(projectID)),
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "no branch linked")
	})
}

func TestGitUnlink_JSONOutput(t *testing.T) {
	db, app := cli.SetupCLITest(t)
	defer func() {
		err := db.Close()
		assert.NoError(t, err)
	}()

	t.Run("JSON output contains expected fields", func(t *testing.T) {
		result, err := db.ExecContext(context.Background(),
			"INSERT INTO projects (name, description, git_branch) VALUES (?, ?, ?)",
			"JSON Test Project", "Test", "json-test-branch")
		assert.NoError(t, err)
		projectID, _ := result.LastInsertId()

		_, err = db.ExecContext(context.Background(),
			"INSERT INTO columns (project_id, name) VALUES (?, 'Todo'), (?, 'In Progress'), (?, 'Done')",
			projectID, projectID, projectID)
		assert.NoError(t, err)

		cmd := GitUnlinkCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			strconv.Itoa(int(projectID)),
			"--json",
		})

		assert.NoError(t, err)
		assert.Contains(t, output, `"success":true`)
		assert.Contains(t, output, `"json-test-branch"`)
		assert.Contains(t, output, `"unlinked_branch"`)
	})
}
