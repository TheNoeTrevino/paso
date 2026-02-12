package project_test

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thenoetrevino/paso/internal/cli/project"
	"github.com/thenoetrevino/paso/internal/git"
	"github.com/thenoetrevino/paso/internal/testing/cli"
)

func TestCreateProject_Integration(t *testing.T) {
	t.Parallel()
	// Setup test DB and App
	db, app := cli.SetupCLITest(t)
	ctx := context.Background()

	t.Run("create project with title only", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		cmd := project.CreateCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--title", "New Project",
			"--quiet",
		})

		assert.NoError(t, err)

		projectIDStr := strings.TrimSpace(output)
		assert.Regexp(t, `^\d+$`, projectIDStr)

		// Verify project exists in DB
		var name string
		err = db.QueryRowContext(ctx,
			"SELECT name FROM projects WHERE id = ?", projectIDStr).Scan(&name)
		assert.NoError(t, err)
		assert.Equal(t, "New Project", name)
	})

	t.Run("create project with description", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		cmd := project.CreateCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--title", "Detailed Project",
			"--description", "This is a detailed project",
			"--quiet",
		})

		assert.NoError(t, err)

		projectIDStr := strings.TrimSpace(output)

		var name, description string
		err = db.QueryRowContext(ctx,
			"SELECT name, description FROM projects WHERE id = ?", projectIDStr).Scan(&name, &description)
		assert.NoError(t, err)
		assert.Equal(t, "Detailed Project", name)
		assert.Equal(t, "This is a detailed project", description)
	})

	t.Run("create project creates default columns", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		cmd := project.CreateCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--title", "Project With Columns",
			"--quiet",
		})

		assert.NoError(t, err)

		projectIDStr := strings.TrimSpace(output)

		// Verify default columns exist
		rows, err := db.QueryContext(ctx,
			"SELECT name FROM columns WHERE project_id = ? ORDER BY id", projectIDStr)
		assert.NoError(t, err)
		defer func() {
			err := rows.Close()
			assert.NoError(t, err)
		}()

		var columns []string
		for rows.Next() {
			var name string
			err := rows.Scan(&name)
			assert.NoError(t, err)
			columns = append(columns, name)
		}
		assert.NoError(t, rows.Err())

		// Check for default columns (standard columns created by service)
		// Note: The service implementation creates Todo, In Progress, Done
		assert.Contains(t, columns, "Todo")
		assert.Contains(t, columns, "In Progress")
		assert.Contains(t, columns, "Done")
	})
}

func TestCreateProject_InGitRepo(t *testing.T) {
	t.Parallel()
	// This test verifies that creating a project in a git repo associates it with the current branch

	// Setup test DB and App with mock git detector
	db, app, mockGit := cli.SetupCLITestWithGit(t)

	t.Run("create project in git repo associates with branch", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		// Configure mock: simulate being in a git repo on branch "feature/test-branch"
		mockGit.Info = git.GitInfo{
			IsRepo:        true,
			CurrentBranch: "feature/test-branch",
			IsDetached:    false,
			HasCommits:    true,
		}
		mockGit.Branches["feature/test-branch"] = true

		cmd := project.CreateCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--title", "Git Project",
			"--description", "Created in git repo",
			"--quiet",
		})

		assert.NoError(t, err)

		projectIDStr := strings.TrimSpace(output)
		assert.Regexp(t, `^\d+$`, projectIDStr)

		// Verify project has git_branch set in DB
		var name string
		var gitBranch sql.NullString
		err = db.QueryRowContext(context.Background(),
			"SELECT name, git_branch FROM projects WHERE id = ?", projectIDStr).Scan(&name, &gitBranch)
		assert.NoError(t, err)
		assert.Equal(t, "Git Project", name)
		assert.True(t, gitBranch.Valid, "Git branch should be set")
		assert.Equal(t, "feature/test-branch", gitBranch.String)
	})
}

func TestCreateProject_OutsideGitRepo(t *testing.T) {
	t.Parallel()
	// This test verifies that creating a project outside a git repo works fine
	// and doesn't associate with any branch

	// Setup test DB and App with mock git detector
	db, app, mockGit := cli.SetupCLITestWithGit(t)

	t.Run("create project outside git repo has no branch", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		// Configure mock: simulate NOT being in a git repo
		mockGit.Info = git.GitInfo{
			IsRepo: false,
		}

		cmd := project.CreateCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--title", "Non-Git Project",
			"--quiet",
		})

		assert.NoError(t, err)

		projectIDStr := strings.TrimSpace(output)

		// Verify project has NULL/empty git_branch
		var name string
		var gitBranch sql.NullString
		err = db.QueryRowContext(context.Background(),
			"SELECT name, git_branch FROM projects WHERE id = ?", projectIDStr).Scan(&name, &gitBranch)
		assert.NoError(t, err)
		assert.Equal(t, "Non-Git Project", name)
		assert.False(t, gitBranch.Valid, "Git branch should be NULL for non-git projects")
	})
}

func TestCreateProject_BranchAlreadyAssociated(t *testing.T) {
	t.Parallel()
	// This test verifies the warning when creating a project on a branch
	// that already has an associated project

	// Setup test DB and App with mock git detector
	db, app, mockGit := cli.SetupCLITestWithGit(t)
	ctx := context.Background()

	t.Run("warn when branch already associated", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		// Configure mock: we are on branch "feature/existing-branch" in a git repo
		mockGit.Info = git.GitInfo{
			IsRepo:        true,
			CurrentBranch: "feature/existing-branch",
			IsDetached:    false,
			HasCommits:    true,
		}
		mockGit.Branches["feature/existing-branch"] = true

		// First, create a project that owns this branch via test helper
		// (bypassing the service to avoid the mock interfering with setup)
		cli.CreateTestProjectWithBranch(t, db, "First Project", "feature/existing-branch")

		// Now try to create another project.
		// The CLI will detect we are on "feature/existing-branch" (via mock),
		// attempt to create with that branch, get ErrGitBranchAlreadyAssociated,
		// and retry without the branch.
		cmd := project.CreateCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--title", "Second Project",
			"--quiet",
		})

		// Should succeed (retries without branch association)
		assert.NoError(t, err)

		projectIDStr := strings.TrimSpace(output)
		assert.Regexp(t, `^\d+$`, projectIDStr)

		// Verify second project has no git_branch (conflict was detected, branch skipped)
		var gitBranch sql.NullString
		err = db.QueryRowContext(ctx,
			"SELECT git_branch FROM projects WHERE id = ?", projectIDStr).Scan(&gitBranch)
		assert.NoError(t, err)
		assert.False(t, gitBranch.Valid, "Second project should not have a branch when conflict detected")
	})
}

func TestCreateProject_MultipleProjectsNoBranch(t *testing.T) {
	t.Parallel()
	// This test verifies that multiple projects without git branches can coexist

	// Setup test DB and App
	db, app := cli.SetupCLITest(t)

	t.Run("multiple projects without branches allowed", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		// Create multiple projects outside git repos
		for i := 0; i < 3; i++ {
			cmd := project.CreateCmd()

			output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
				"--title", fmt.Sprintf("Project %d", i),
				"--quiet",
			})

			assert.NoError(t, err)
			assert.Regexp(t, `^\d+$`, strings.TrimSpace(output))
		}

		// Verify all projects exist without git branches
		rows, err := db.QueryContext(context.Background(),
			"SELECT id, name, git_branch FROM projects ORDER BY id")
		assert.NoError(t, err)
		defer func() { _ = rows.Close() }()

		count := 0
		for rows.Next() {
			var id int
			var name string
			var gitBranch sql.NullString
			err := rows.Scan(&id, &name, &gitBranch)
			assert.NoError(t, err)
			count++
		}
		assert.NoError(t, rows.Err())

		assert.GreaterOrEqual(t, count, 3, "Should have at least 3 projects")
	})
}

func TestCreateProject_DifferentBranches(t *testing.T) {
	t.Parallel()
	// This test verifies that projects can be created on different branches

	// Setup test DB and App
	db, _ := cli.SetupCLITest(t)

	t.Run("projects on different branches allowed", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		// Manually create projects on different branches
		// (simulating what the CLI would do with git detection)

		branches := []string{
			"feature/branch-1",
			"feature/branch-2",
			"main",
		}

		for i, branch := range branches {
			cli.CreateTestProjectWithBranch(t, db, fmt.Sprintf("Project %d", i), branch)
		}

		// Verify all projects have unique branches
		rows, err := db.QueryContext(context.Background(),
			"SELECT git_branch FROM projects WHERE git_branch IS NOT NULL ORDER BY id")
		assert.NoError(t, err)
		defer func() { _ = rows.Close() }()

		seenBranches := make(map[string]bool)
		for rows.Next() {
			var gitBranch string
			err := rows.Scan(&gitBranch)
			assert.NoError(t, err)
			assert.False(t, seenBranches[gitBranch], "Branches should be unique")
			seenBranches[gitBranch] = true
		}
		assert.NoError(t, rows.Err())

		assert.Len(t, seenBranches, 3, "Should have 3 unique branches")
	})
}

func TestCreateProject_Integration_Errors(t *testing.T) {
	t.Parallel()
	_, app := cli.SetupCLITest(t)

	t.Run("create project missing title returns error", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		cmd := project.CreateCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{})
		assert.Error(t, err)
	})
}

func TestCreateProject_GitBranchWithSlashes(t *testing.T) {
	t.Parallel()
	// This test verifies that branch names with slashes (like feature/my-feature) work correctly

	// Setup test DB and App
	db, _ := cli.SetupCLITest(t)

	t.Run("branch names with slashes", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		// Create a project with a branch containing slashes
		projectID := cli.CreateTestProjectWithBranch(t, db, "Test Project", "feature/auth/user-login")

		// Verify branch was stored correctly
		var gitBranch string
		err := db.QueryRowContext(context.Background(),
			"SELECT git_branch FROM projects WHERE id = ?", projectID).Scan(&gitBranch)
		assert.NoError(t, err)
		assert.Equal(t, "feature/auth/user-login", gitBranch)
	})
}
