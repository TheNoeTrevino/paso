package label_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thenoetrevino/paso/internal/cli/label"
	"github.com/thenoetrevino/paso/internal/testing/cli"
)

func TestCreateLabel_Errors(t *testing.T) {
	t.Parallel()
	// Setup test DB and App
	db, app := cli.SetupCLITest(t)

	// Create test project
	projectID := cli.CreateTestProject(t, db, "Test Project")

	// Task 31: Test Duplicate Labels
	t.Run("create duplicate label", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		cmd := label.CreateCmd()

		// 1. Create first label
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", fmt.Sprintf("%d", projectID),
			"--name", "bug",
			"--color", "#FF0000",
			"--quiet",
		})
		assert.NoError(t, err)

		// 2. Attempt to create duplicate label
		cmd = label.CreateCmd()
		_, err = cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", fmt.Sprintf("%d", projectID),
			"--name", "bug",
			"--color", "#00FF00", // Even with different color, name should be unique per project
			"--quiet",
		})

		if assert.Error(t, err) {
			// Check for duplicate error message (actual message depends on DB constraint)
			assert.Contains(t, err.Error(), "label creation error")
		}
	})

	t.Run("create label with invalid color", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		cmd := label.CreateCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", fmt.Sprintf("%d", projectID),
			"--name", "bad-color",
			"--color", "invalid-color",
			"--quiet",
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid color")
	})
}
