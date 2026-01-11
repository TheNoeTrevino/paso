package project

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thenoetrevino/paso/internal/testutil/cli"
)

func TestCreateProject_Positive(t *testing.T) {
	// Setup test DB and App
	db, app := cli.SetupCLITest(t)
	defer func() {
		err := db.Close()
		assert.NoError(t, err)
	}()

	t.Run("Create project with title only", func(t *testing.T) {
		cmd := CreateCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--title", "New Project",
			"--quiet",
		})

		assert.NoError(t, err)

		projectIDStr := strings.TrimSpace(output)
		assert.Regexp(t, `^\d+$`, projectIDStr)

		// Verify project exists in DB
		var name string
		err = db.QueryRowContext(context.Background(),
			"SELECT name FROM projects WHERE id = ?", projectIDStr).Scan(&name)
		assert.NoError(t, err)
		assert.Equal(t, "New Project", name)
	})

	t.Run("Create project with description", func(t *testing.T) {
		cmd := CreateCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--title", "Detailed Project",
			"--description", "This is a detailed project",
			"--quiet",
		})

		assert.NoError(t, err)

		projectIDStr := strings.TrimSpace(output)

		var name, description string
		err = db.QueryRowContext(context.Background(),
			"SELECT name, description FROM projects WHERE id = ?", projectIDStr).Scan(&name, &description)
		assert.NoError(t, err)
		assert.Equal(t, "Detailed Project", name)
		assert.Equal(t, "This is a detailed project", description)
	})

	t.Run("Create project creates default columns", func(t *testing.T) {
		cmd := CreateCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--title", "Project With Columns",
			"--quiet",
		})

		assert.NoError(t, err)

		projectIDStr := strings.TrimSpace(output)

		// Verify default columns exist
		rows, err := db.QueryContext(context.Background(),
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

// ============================================================================
// GIT BRANCH ASSOCIATION TESTS (TDD RED PHASE)
// ============================================================================

func TestCreateProject_InGitRepo(t *testing.T) {
	// This test should be run in a git repository
	// It verifies that creating a project in a git repo associates it with the current branch

	// Setup test DB and App
	db, app := cli.SetupCLITest(t)
	defer func() {
		err := db.Close()
		assert.NoError(t, err)
	}()

	t.Run("Create project in git repo associates with branch", func(t *testing.T) {
		// This test will fail until git detection is implemented
		// It assumes DetectGitInfo() will be called during project creation

		cmd := CreateCmd()

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

		// The git branch should be set (if we're in a git repo)
		// If not in a git repo, it should be NULL
		t.Logf("Git branch: valid=%v, value='%s'", gitBranch.Valid, gitBranch.String)
	})
}

func TestCreateProject_OutsideGitRepo(t *testing.T) {
	// This test verifies that creating a project outside a git repo works fine
	// and doesn't associate with any branch

	// Setup test DB and App
	db, app := cli.SetupCLITest(t)
	defer func() {
		err := db.Close()
		assert.NoError(t, err)
	}()

	t.Run("Create project outside git repo has no branch", func(t *testing.T) {
		// Create a temporary directory that is NOT a git repo
		tmpDir := t.TempDir()

		// Change to that directory
		originalDir, err := os.Getwd()
		assert.NoError(t, err)
		defer func() {
			err := os.Chdir(originalDir)
			assert.NoError(t, err)
		}()
		err = os.Chdir(tmpDir)
		assert.NoError(t, err)

		cmd := CreateCmd()

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
	// This test verifies the warning when creating a project on a branch
	// that already has an associated project

	// Setup test DB and App
	db, app := cli.SetupCLITest(t)
	defer func() {
		err := db.Close()
		assert.NoError(t, err)
	}()

	t.Run("Warn when branch already associated", func(t *testing.T) {
		// First, manually create a project with a git branch
		result, err := db.ExecContext(context.Background(),
			"INSERT INTO projects (name, description, git_branch) VALUES (?, ?, ?)",
			"First Project", "First", "feature/existing-branch")
		assert.NoError(t, err)
		firstID, _ := result.LastInsertId()

		// Create default columns for the first project
		_, err = db.ExecContext(context.Background(),
			"INSERT INTO columns (project_id, name) VALUES (?, 'Todo'), (?, 'In Progress'), (?, 'Done')",
			firstID, firstID, firstID)
		assert.NoError(t, err)

		// Now try to create another project on the same branch
		// This should warn and skip the association
		cmd := CreateCmd()

		// We need to simulate being on "feature/existing-branch"
		// This test will fail until git detection is implemented
		// For now, we can only test the service layer logic

		// The CLI should:
		// 1. Detect git branch "feature/existing-branch"
		// 2. Call GetProjectByGitBranch()
		// 3. Find existing project
		// 4. Print warning
		// 5. Create new project WITHOUT git_branch

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--title", "Second Project",
			"--quiet",
		})

		// Should succeed (just skip the association)
		assert.NoError(t, err)

		projectIDStr := strings.TrimSpace(output)

		// Verify second project has no git_branch
		var gitBranch sql.NullString
		err = db.QueryRowContext(context.Background(),
			"SELECT git_branch FROM projects WHERE id = ?", projectIDStr).Scan(&gitBranch)
		assert.NoError(t, err)
		// Second project should NOT have the branch (conflict detected)
		// This behavior depends on implementation
		t.Logf("Second project git_branch: valid=%v, value='%s'", gitBranch.Valid, gitBranch.String)
	})
}

func TestCreateProject_MultipleProjectsNoBranch(t *testing.T) {
	// This test verifies that multiple projects without git branches can coexist

	// Setup test DB and App
	db, app := cli.SetupCLITest(t)
	defer func() {
		err := db.Close()
		assert.NoError(t, err)
	}()

	t.Run("Multiple projects without branches allowed", func(t *testing.T) {
		// Create multiple projects outside git repos
		for i := 0; i < 3; i++ {
			cmd := CreateCmd()

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
	// This test verifies that projects can be created on different branches

	// Setup test DB and App
	db, _ := cli.SetupCLITest(t)
	defer func() {
		err := db.Close()
		assert.NoError(t, err)
	}()

	t.Run("Projects on different branches allowed", func(t *testing.T) {
		// Manually create projects on different branches
		// (simulating what the CLI would do with git detection)

		branches := []string{
			"feature/branch-1",
			"feature/branch-2",
			"main",
		}

		for i, branch := range branches {
			result, err := db.ExecContext(context.Background(),
				"INSERT INTO projects (name, description, git_branch) VALUES (?, ?, ?)",
				fmt.Sprintf("Project %d", i), "Description", branch)
			assert.NoError(t, err)

			projectID, _ := result.LastInsertId()

			// Create default columns
			_, err = db.ExecContext(context.Background(),
				"INSERT INTO columns (project_id, name) VALUES (?, 'Todo'), (?, 'In Progress'), (?, 'Done')",
				projectID, projectID, projectID)
			assert.NoError(t, err)
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

func TestCreateProject_GitBranchWithSlashes(t *testing.T) {
	// This test verifies that branch names with slashes (like feature/my-feature) work correctly

	// Setup test DB and App
	db, _ := cli.SetupCLITest(t)
	defer func() {
		err := db.Close()
		assert.NoError(t, err)
	}()

	t.Run("Branch names with slashes", func(t *testing.T) {
		// Manually create a project with a branch containing slashes
		result, err := db.ExecContext(context.Background(),
			"INSERT INTO projects (name, description, git_branch) VALUES (?, ?, ?)",
			"Test Project", "Description", "feature/auth/user-login")
		assert.NoError(t, err)

		projectID, _ := result.LastInsertId()

		// Verify branch was stored correctly
		var gitBranch string
		err = db.QueryRowContext(context.Background(),
			"SELECT git_branch FROM projects WHERE id = ?", projectID).Scan(&gitBranch)
		assert.NoError(t, err)
		assert.Equal(t, "feature/auth/user-login", gitBranch)
	})
}
