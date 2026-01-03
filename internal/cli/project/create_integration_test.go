package project

import (
	"context"
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

		// Check for default columns (standard columns created by service)
		// Note: The service implementation creates Todo, In Progress, Done
		assert.Contains(t, columns, "Todo")
		assert.Contains(t, columns, "In Progress")
		assert.Contains(t, columns, "Done")
	})
}
