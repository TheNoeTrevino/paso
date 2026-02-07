package project

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thenoetrevino/paso/internal/testutil/cli"
)

// TestGitLink_Positive tests are skipped because runGitLink calls
// git.BranchExists which requires the branch to actually exist in the
// local git repository. We cannot create real git branches in the test
// environment without significant setup. The service layer tests in
// internal/services/project use a mock GitChecker to test this logic.
func TestGitLink_Positive(t *testing.T) {
	t.Skip("Skipping: git-link validates branch existence via real git commands")
}

func TestGitLink_Negative(t *testing.T) {
	_, app := cli.SetupCLITest(t)

	t.Run("Non-existent project ID", func(t *testing.T) {
		cmd := GitLinkCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", "99999",
			"--branch", "feature/test",
		})
		assert.Error(t, err)
	})
}

func TestGitUnlink_Positive(t *testing.T) {
	db, app := cli.SetupCLITest(t)

	// Tests that require a pre-linked branch are skipped because git-link
	// validates branch existence via real git commands. We test the
	// no-branch-linked paths which don't depend on git.

	t.Run("Unlink project with pre-linked branch", func(t *testing.T) {
		t.Skip("Skipping: requires git-link which validates branch existence via real git commands")
	})

	t.Run("Unlink project with no branch linked", func(t *testing.T) {
		projectID := cli.CreateTestProject(t, db, "No Branch Project")

		cmd := GitUnlinkCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", projectID),
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "no branch linked")
	})

	t.Run("Unlink project with no branch linked JSON", func(t *testing.T) {
		projectID := cli.CreateTestProject(t, db, "No Branch JSON Project")

		cmd := GitUnlinkCmd()
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

func TestGitUnlink_Negative(t *testing.T) {
	_, app := cli.SetupCLITest(t)

	t.Run("Non-existent project ID", func(t *testing.T) {
		cmd := GitUnlinkCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", "99999",
		})
		assert.Error(t, err)
	})
}
