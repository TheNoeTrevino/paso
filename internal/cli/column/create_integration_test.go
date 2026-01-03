package column

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thenoetrevino/paso/internal/testutil/cli"
)

func TestCreateColumn_Integration(t *testing.T) {
	// Setup test DB and App
	db, app := cli.SetupCLITest(t)
	defer func() {
		_ = db.Close()
	}()

	// Create test project with default columns (Todo, In Progress, Done)
	projectID := cli.CreateTestProject(t, db, "Test Project")

	tests := []struct {
		name          string
		flags         []string
		expectedError bool
		expectedInDB  bool
		expectedName  string
		expectedReady bool
		expectedDone  bool
		verifyOutput  func(t *testing.T, output string)
	}{
		{
			name: "Create column with basic flags",
			flags: []string{
				"--name", "Custom Column",
				"--project", fmt.Sprintf("%d", projectID),
			},
			expectedError: false,
			expectedInDB:  true,
			expectedName:  "Custom Column",
			expectedReady: false,
			expectedDone:  false,
			verifyOutput: func(t *testing.T, output string) {
				// Human-readable output should contain success message
				assert.Contains(t, output, "Column 'Custom Column' created successfully")
				assert.Contains(t, output, "Test Project")
			},
		},
		{
			name: "Create column with ready flag",
			flags: []string{
				"--name", "Ready Column",
				"--project", fmt.Sprintf("%d", projectID),
				"--ready",
			},
			expectedError: false,
			expectedInDB:  true,
			expectedName:  "Ready Column",
			expectedReady: true,
			expectedDone:  false,
			verifyOutput: func(t *testing.T, output string) {
				// Human-readable output should contain success message
				assert.Contains(t, output, "Column 'Ready Column' created successfully")
			},
		},
		{
			name: "Create column with completed flag",
			flags: []string{
				"--name", "Completed Column",
				"--project", fmt.Sprintf("%d", projectID),
				"--completed",
			},
			expectedError: false,
			expectedInDB:  true,
			expectedName:  "Completed Column",
			expectedReady: false,
			expectedDone:  true,
			verifyOutput: func(t *testing.T, output string) {
				// Human-readable output should contain success message
				assert.Contains(t, output, "Column 'Completed Column' created successfully")
			},
		},
		{
			name: "Create column with JSON output",
			flags: []string{
				"--name", "JSON Column",
				"--project", fmt.Sprintf("%d", projectID),
				"--json",
			},
			expectedError: false,
			expectedInDB:  true,
			expectedName:  "JSON Column",
			expectedReady: false,
			expectedDone:  false,
			verifyOutput: func(t *testing.T, output string) {
				// Parse JSON output
				var result map[string]interface{}
				err := json.Unmarshal([]byte(output), &result)
				assert.NoError(t, err, "Output should be valid JSON")

				// Verify JSON structure
				assert.True(t, result["success"].(bool), "success should be true")

				column := result["column"].(map[string]interface{})
				assert.Equal(t, "JSON Column", column["name"])
				assert.Equal(t, float64(projectID), column["project_id"])

				// Verify ID is numeric
				columnID := column["id"]
				assert.NotNil(t, columnID, "column ID should be present in JSON")
			},
		},
		{
			name: "Create column with quiet mode",
			flags: []string{
				"--name", "Quiet Column",
				"--project", fmt.Sprintf("%d", projectID),
				"--quiet",
			},
			expectedError: false,
			expectedInDB:  true,
			expectedName:  "Quiet Column",
			expectedReady: false,
			expectedDone:  false,
			verifyOutput: func(t *testing.T, output string) {
				// Quiet mode should return only the ID
				trimmed := strings.TrimSpace(output)
				// Verify output is just a number
				assert.Regexp(t, regexp.MustCompile(`^\d+$`), trimmed,
					"Quiet mode should return only column ID")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CreateCmd()

			output, err := cli.ExecuteCLICommand(t, app, cmd, tt.flags)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// If output verification is defined, run it
			if tt.verifyOutput != nil {
				tt.verifyOutput(t, output)
			}

			// Verify column was created in database
			if tt.expectedInDB {
				var dbName string
				var holdsReady bool
				var holdsCompleted bool
				var columnID int

				err := db.QueryRowContext(context.Background(), `
					SELECT id, name, holds_ready_tasks, holds_completed_tasks
					FROM columns
					WHERE project_id = ? AND name = ?
				`, projectID, tt.expectedName).Scan(&columnID, &dbName, &holdsReady, &holdsCompleted)

				assert.NoError(t, err, "Column should exist in database")
				assert.Equal(t, tt.expectedName, dbName)
				assert.Equal(t, tt.expectedReady, holdsReady)
				assert.Equal(t, tt.expectedDone, holdsCompleted)

				// Verify the ID from output matches the database ID (for non-JSON, non-quiet modes)
				if strings.Contains(strings.Join(tt.flags, " "), "--quiet") {
					trimmed := strings.TrimSpace(output)
					var outputID int
					_, err := fmt.Sscanf(trimmed, "%d", &outputID)
					assert.NoError(t, err)
					assert.Equal(t, columnID, outputID)
				}
			}
		})
	}
}

func TestCreateColumn_ErrorCases(t *testing.T) {
	// Setup test DB and App
	db, app := cli.SetupCLITest(t)
	defer func() {
		_ = db.Close()
	}()

	// Create test project
	projectID := cli.CreateTestProject(t, db, "Test Project")

	errorTests := []struct {
		name          string
		flags         []string
		expectedError bool
	}{
		{
			name: "Missing required name flag",
			flags: []string{
				"--project", fmt.Sprintf("%d", projectID),
			},
			expectedError: true,
		},
		{
			name: "Invalid project ID",
			flags: []string{
				"--name", "Invalid Project Column",
				"--project", "99999",
			},
			expectedError: true,
		},
		{
			name: "Invalid after column ID",
			flags: []string{
				"--name", "Column with Invalid After",
				"--project", fmt.Sprintf("%d", projectID),
				"--after", "99999",
			},
			expectedError: true,
		},
	}

	for _, tt := range errorTests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CreateCmd()

			_, err := cli.ExecuteCLICommand(t, app, cmd, tt.flags)

			if tt.expectedError {
				assert.Error(t, err, fmt.Sprintf("Test case '%s' should have resulted in an error", tt.name))
			}
		})
	}
}

func TestCreateColumn_FlagCombinations(t *testing.T) {
	db, app := cli.SetupCLITest(t)
	defer func() {
		_ = db.Close()
	}()

	projectID := cli.CreateTestProject(t, db, "Combo Test Project")

	t.Run("Create column with ready and quiet flags", func(t *testing.T) {
		cmd := CreateCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--name", "Ready and Quiet",
			"--project", fmt.Sprintf("%d", projectID),
			"--ready",
			"--quiet",
		})

		assert.NoError(t, err)

		// Quiet mode should return only the ID
		trimmed := strings.TrimSpace(output)
		assert.Regexp(t, regexp.MustCompile(`^\d+$`), trimmed)

		// Verify in database
		var columnID int
		var holdsReady bool
		err = db.QueryRowContext(context.Background(), `
			SELECT id, holds_ready_tasks
			FROM columns
			WHERE project_id = ? AND name = 'Ready and Quiet'
		`, projectID).Scan(&columnID, &holdsReady)

		assert.NoError(t, err)
		assert.True(t, holdsReady)
	})

	t.Run("Create column with completed and JSON flags", func(t *testing.T) {
		cmd := CreateCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--name", "Completed and JSON",
			"--project", fmt.Sprintf("%d", projectID),
			"--completed",
			"--json",
		})

		assert.NoError(t, err)

		// Parse JSON output
		var result map[string]interface{}
		err = json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)

		// Verify database state
		var holdsCompleted bool
		err = db.QueryRowContext(context.Background(), `
			SELECT holds_completed_tasks
			FROM columns
			WHERE project_id = ? AND name = 'Completed and JSON'
		`, projectID).Scan(&holdsCompleted)

		assert.NoError(t, err)
		assert.True(t, holdsCompleted)
	})

	t.Run("Create multiple columns in same project", func(t *testing.T) {
		cmd := CreateCmd()

		for i := 1; i <= 3; i++ {
			output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
				"--name", fmt.Sprintf("Sequential Column %d", i),
				"--project", fmt.Sprintf("%d", projectID),
				"--quiet",
			})

			assert.NoError(t, err)
			trimmed := strings.TrimSpace(output)
			assert.Regexp(t, regexp.MustCompile(`^\d+$`), trimmed)
		}

		// Verify all columns exist in database
		var count int
		err := db.QueryRowContext(context.Background(), `
			SELECT COUNT(*)
			FROM columns
			WHERE project_id = ? AND name LIKE 'Sequential Column %'
		`, projectID).Scan(&count)

		assert.NoError(t, err)
		assert.Equal(t, 3, count)
	})
}
