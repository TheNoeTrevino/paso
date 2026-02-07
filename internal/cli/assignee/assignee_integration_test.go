package assignee

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thenoetrevino/paso/internal/testutil"
	cli "github.com/thenoetrevino/paso/internal/testutil/cli"
)

func TestCreateAssignee_Positive(t *testing.T) {
	db, app := cli.SetupCLITest(t)

	t.Run("create with name default output", func(t *testing.T) {
		cmd := CreateCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{"--name", "alice"})
		assert.NoError(t, err)
		assert.Contains(t, output, "alice")
		assert.Contains(t, output, "ID:")

		// Verify DB state
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM assignees WHERE name = ?", "alice").Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("create with json output", func(t *testing.T) {
		cmd := CreateCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{"--name", "bob", "--json"})
		assert.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)
		assert.True(t, result["success"].(bool))

		data := result["data"].(map[string]any)
		assert.Equal(t, "bob", data["Name"])
		assert.Greater(t, data["ID"].(float64), 0.0)
	})

	t.Run("create with quiet output", func(t *testing.T) {
		cmd := CreateCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{"--name", "charlie", "--quiet"})
		assert.NoError(t, err)

		// Quiet mode should output just the ID
		output = strings.TrimSpace(output)
		assert.Regexp(t, `^\d+$`, output, "Quiet mode should output only the ID")
	})
}

func TestCreateAssignee_Negative(t *testing.T) {
	_, app := cli.SetupCLITest(t)

	t.Run("missing name flag", func(t *testing.T) {
		cmd := CreateCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "required flag(s)")
	})

	t.Run("empty name", func(t *testing.T) {
		cmd := CreateCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{"--name", ""})
		assert.Error(t, err)
	})

	t.Run("duplicate name", func(t *testing.T) {
		// Create first assignee
		cmd := CreateCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{"--name", "duplicate"})
		assert.NoError(t, err)

		// Try to create with same name
		cmd = CreateCmd()
		_, err = cli.ExecuteCLICommand(t, app, cmd, []string{"--name", "duplicate"})
		assert.Error(t, err)
		assert.Contains(t, strings.ToLower(err.Error()), "already exists")
	})

	t.Run("name too long", func(t *testing.T) {
		longName := strings.Repeat("a", 256) // Max is 255
		cmd := CreateCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{"--name", longName})
		assert.Error(t, err)
	})
}

func TestListAssignees_Empty(t *testing.T) {
	_, app := cli.SetupCLITest(t)

	t.Run("default output empty", func(t *testing.T) {
		cmd := ListCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{})
		assert.NoError(t, err)
		assert.Contains(t, output, "No assignees found")
	})

	t.Run("json output empty", func(t *testing.T) {
		cmd := ListCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{"--json"})
		assert.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)
		assert.True(t, result["success"].(bool))

		assignees := result["assignees"].([]any)
		assert.Empty(t, assignees)
	})

	t.Run("quiet output empty", func(t *testing.T) {
		cmd := ListCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{"--quiet"})
		assert.NoError(t, err)
		assert.Empty(t, strings.TrimSpace(output))
	})
}

func TestListAssignees_Populated(t *testing.T) {
	db, app := cli.SetupCLITest(t)

	// Pre-populate with assignees
	testutil.CreateTestAssignee(t, db, "alice")
	testutil.CreateTestAssignee(t, db, "bob")
	testutil.CreateTestAssignee(t, db, "charlie")

	t.Run("default output table", func(t *testing.T) {
		cmd := ListCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{})
		assert.NoError(t, err)
		assert.Contains(t, output, "alice")
		assert.Contains(t, output, "bob")
		assert.Contains(t, output, "charlie")
		assert.Contains(t, output, "ID")
		assert.Contains(t, output, "NAME")
	})

	t.Run("json output", func(t *testing.T) {
		cmd := ListCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{"--json"})
		assert.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)
		assert.True(t, result["success"].(bool))

		assignees := result["assignees"].([]any)
		assert.Len(t, assignees, 3)

		// Check that all names are present
		names := make([]string, 0)
		for _, a := range assignees {
			assignee := a.(map[string]any)
			names = append(names, assignee["name"].(string))
		}
		assert.Contains(t, names, "alice")
		assert.Contains(t, names, "bob")
		assert.Contains(t, names, "charlie")
	})

	t.Run("quiet output ids only", func(t *testing.T) {
		cmd := ListCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{"--quiet"})
		assert.NoError(t, err)

		lines := strings.Split(strings.TrimSpace(output), "\n")
		assert.Len(t, lines, 3)
		for _, line := range lines {
			assert.Regexp(t, `^\d+$`, line, "Each line should be an ID")
		}
	})
}

func TestDeleteAssignee_Positive(t *testing.T) {
	db, app := cli.SetupCLITest(t)

	t.Run("delete with force and json", func(t *testing.T) {
		// Create assignee to delete
		id := testutil.CreateTestAssignee(t, db, "to-delete")

		cmd := DeleteCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", id),
			"--force",
			"--json",
		})
		assert.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)
		assert.True(t, result["success"].(bool))
		assert.Equal(t, float64(id), result["assignee_id"].(float64))

		// Verify DB state - assignee should be gone
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM assignees WHERE id = ?", id).Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("delete with force and quiet", func(t *testing.T) {
		id := testutil.CreateTestAssignee(t, db, "quiet-delete")

		cmd := DeleteCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", id),
			"--force",
			"--quiet",
		})
		assert.NoError(t, err)
		assert.Empty(t, strings.TrimSpace(output))

		// Verify deletion
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM assignees WHERE id = ?", id).Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("delete with force default output", func(t *testing.T) {
		id := testutil.CreateTestAssignee(t, db, "default-delete")

		cmd := DeleteCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", id),
			"--force",
		})
		assert.NoError(t, err)
		assert.Contains(t, output, "deleted successfully")
	})
}

func TestDeleteAssignee_Negative(t *testing.T) {
	_, app := cli.SetupCLITest(t)

	t.Run("nonexistent id", func(t *testing.T) {
		cmd := DeleteCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", "999999",
			"--force",
		})
		assert.Error(t, err)
		assert.Contains(t, strings.ToLower(err.Error()), "not found")
	})

	t.Run("missing id flag", func(t *testing.T) {
		cmd := DeleteCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{"--force"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "required flag")
	})

	t.Run("invalid id format", func(t *testing.T) {
		cmd := DeleteCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", "not-a-number",
			"--force",
		})
		assert.Error(t, err)
	})
}
