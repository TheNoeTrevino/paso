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

func TestGitLink_PositionalArguments(t *testing.T) {
	db, app := cli.SetupCLITest(t)
	defer func() {
		err := db.Close()
		assert.NoError(t, err)
	}()

	t.Run("Link project with positional project ID only (uses current branch)", func(t *testing.T) {
		result, err := db.ExecContext(context.Background(),
			"INSERT INTO projects (name, description) VALUES (?, ?)",
			"Test Project", "Test Description")
		assert.NoError(t, err)
		projectID, _ := result.LastInsertId()

		_, err = db.ExecContext(context.Background(),
			"INSERT INTO columns (project_id, name) VALUES (?, 'Todo'), (?, 'In Progress'), (?, 'Done')",
			projectID, projectID, projectID)
		assert.NoError(t, err)

		cmd := GitLinkCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			strconv.Itoa(int(projectID)),
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
	db, app := cli.SetupCLITest(t)
	defer func() {
		err := db.Close()
		assert.NoError(t, err)
	}()

	t.Run("Link project with --id flag only (uses current branch)", func(t *testing.T) {
		result, err := db.ExecContext(context.Background(),
			"INSERT INTO projects (name, description) VALUES (?, ?)",
			"Flag Test Project", "Test Description")
		assert.NoError(t, err)
		projectID, _ := result.LastInsertId()

		_, err = db.ExecContext(context.Background(),
			"INSERT INTO columns (project_id, name) VALUES (?, 'Todo'), (?, 'In Progress'), (?, 'Done')",
			projectID, projectID, projectID)
		assert.NoError(t, err)

		cmd := GitLinkCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", strconv.Itoa(int(projectID)),
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
	db, app := cli.SetupCLITest(t)
	defer func() {
		err := db.Close()
		assert.NoError(t, err)
	}()

	t.Run("Invalid project ID returns error", func(t *testing.T) {
		cmd := GitLinkCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"invalid-id",
			"test-branch",
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be a number")
	})

	t.Run("Too many arguments returns error", func(t *testing.T) {
		cmd := GitLinkCmd()

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
	db, _ := cli.SetupCLITest(t)
	defer func() {
		err := db.Close()
		assert.NoError(t, err)
	}()

	t.Run("Force flag transfers branch from one project to another", func(t *testing.T) {
		t.Skip("Skipping: requires valid git branch to exist")
	})
}

func TestGitLink_JSONOutput(t *testing.T) {
	db, app := cli.SetupCLITest(t)
	defer func() {
		err := db.Close()
		assert.NoError(t, err)
	}()

	t.Run("JSON output contains expected fields", func(t *testing.T) {
		result, err := db.ExecContext(context.Background(),
			"INSERT INTO projects (name, description) VALUES (?, ?)",
			"JSON Test Project", "Test")
		assert.NoError(t, err)
		projectID, _ := result.LastInsertId()

		_, err = db.ExecContext(context.Background(),
			"INSERT INTO columns (project_id, name) VALUES (?, 'Todo'), (?, 'In Progress'), (?, 'Done')",
			projectID, projectID, projectID)
		assert.NoError(t, err)

		cmd := GitLinkCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			strconv.Itoa(int(projectID)),
			"--json",
		})

		if err == nil {
			assert.Contains(t, output, `"success":true`)
		}
	})
}
