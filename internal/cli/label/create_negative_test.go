package label

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thenoetrevino/paso/internal/testutil/cli"
)

func TestCreateLabel_Negative(t *testing.T) {
	// Setup test DB and App
	db, app := cli.SetupCLITest(t)
	defer db.Close()

	// Create test project
	projectID := cli.CreateTestProject(t, db, "Test Project")

	// Task 31: Test Duplicate Labels
	t.Run("Create duplicate label", func(t *testing.T) {
		cmd := CreateCmd()

		// 1. Create first label
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", fmt.Sprintf("%d", projectID),
			"--name", "bug",
			"--color", "#FF0000",
			"--quiet",
		})
		assert.NoError(t, err)

		// 2. Attempt to create duplicate label
		cmd = CreateCmd()
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

	t.Run("Create label with invalid color", func(t *testing.T) {
		cmd := CreateCmd()

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
