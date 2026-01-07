package cli

import (
	"context"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/testutil"
	testutilcli "github.com/thenoetrevino/paso/internal/testutil/cli"
)

// ============================================================================
// Color Validation Tests
// ============================================================================

func TestValidateColorHex_Valid(t *testing.T) {
	tests := []string{
		"#FF0000", // Red
		"#00FF00", // Green
		"#0000FF", // Blue
		"#FFFFFF", // White
		"#000000", // Black
		"#FF5733", // Random color
		"#ff5733", // Lowercase (should work)
		"#AbCdEf", // Mixed case
	}

	for _, color := range tests {
		t.Run(color, func(t *testing.T) {
			err := ValidateColorHex(color)
			assert.NoError(t, err, "Color should be valid: %s", color)
		})
	}
}

func TestValidateColorHex_Invalid(t *testing.T) {
	tests := []struct {
		color       string
		description string
	}{
		{"FF0000", "missing # prefix"},
		{"#FFF", "too short (3 chars)"},
		{"#FF00000", "too long (7 chars)"},
		{"#GGGGGG", "invalid hex characters"},
		{"#FF00G0", "one invalid character"},
		{"#FF 000", "contains space"},
		{"", "empty string"},
		{"#", "only # symbol"},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			err := ValidateColorHex(tt.color)
			if err == nil {
				t.Errorf("Expected %s to be invalid (%s), but got no error", tt.color, tt.description)
			}
		})
	}
}

// ============================================================================
// Priority Parsing Tests
// ============================================================================

func TestParsePriority_Valid(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"trivial", 1},
		{"low", 2},
		{"medium", 3},
		{"high", 4},
		{"critical", 5},
		// Test case insensitivity
		{"TRIVIAL", 1},
		{"Low", 2},
		{"MeDiUm", 3},
		{"HIGH", 4},
		{"Critical", 5},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ParsePriority(tt.input)
			assert.NoError(t, err, "ParsePriority should not return error for: %s", tt.input)
			if result != tt.expected {
				t.Errorf("Expected %d for '%s', got %d", tt.expected, tt.input, result)
			}
		})
	}
}

func TestParsePriority_Invalid(t *testing.T) {
	tests := []string{
		"invalid",
		"normal",
		"urgent",
		"",
		"123",
		"trivial ",
		" low",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, err := ParsePriority(input)
			if err == nil {
				t.Errorf("Expected error for invalid priority '%s', got nil", input)
			}
		})
	}
}

// ============================================================================
// Task Type Parsing Tests
// ============================================================================

func TestParseTaskType_Valid(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"task", 1},
		{"feature", 2},
		// Test case insensitivity
		{"TASK", 1},
		{"Feature", 2},
		{"TaSk", 1},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ParseTaskType(tt.input)
			assert.NoError(t, err, "ParseTaskType should not return error for: %s", tt.input)
			if result != tt.expected {
				t.Errorf("Expected %d for '%s', got %d", tt.expected, tt.input, result)
			}
		})
	}
}

func TestParseTaskType_Invalid(t *testing.T) {
	tests := []string{
		"bug",
		"story",
		"epic",
		"",
		"123",
		"task ",
		" feature",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, err := ParseTaskType(input)
			if err == nil {
				t.Errorf("Expected error for invalid type '%s', got nil", input)
			}
		})
	}
}

// ============================================================================
// GetLabelByID Tests
// ============================================================================

func TestGetLabelByID_Found(t *testing.T) {
	db, appInstance := testutilcli.SetupCLITest(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	// Create test data
	projectID := testutil.CreateTestProject(t, db, "Test Project")
	labelID := testutil.CreateTestLabel(t, db, projectID, "Bug", "#FF0000")

	// Create CLI instance
	cliInstance := &CLI{
		ctx: ctx,
		App: appInstance,
	}

	// Test finding the label
	label, err := GetLabelByID(ctx, cliInstance, labelID)
	require.NoError(t, err, "Should find label successfully")

	if label.ID != labelID {
		t.Errorf("Expected label ID %d, got %d", labelID, label.ID)
	}
	if label.Name != "Bug" {
		t.Errorf("Expected label name 'Bug', got '%s'", label.Name)
	}
	if label.Color != "#FF0000" {
		t.Errorf("Expected label color '#FF0000', got '%s'", label.Color)
	}
	if label.ProjectID != projectID {
		t.Errorf("Expected project ID %d, got %d", projectID, label.ProjectID)
	}
}

func TestGetLabelByID_Found_MultipleProjects(t *testing.T) {
	db, appInstance := testutilcli.SetupCLITest(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	// Create multiple projects with labels
	project1ID := testutil.CreateTestProject(t, db, "Project 1")
	testutil.CreateTestLabel(t, db, project1ID, "Label 1", "#FF0000")

	project2ID := testutil.CreateTestProject(t, db, "Project 2")
	label2ID := testutil.CreateTestLabel(t, db, project2ID, "Label 2", "#00FF00")

	// Create CLI instance
	cliInstance := &CLI{
		ctx: ctx,
		App: appInstance,
	}

	// Test finding label from second project
	label, err := GetLabelByID(ctx, cliInstance, label2ID)
	require.NoError(t, err, "Should find label successfully")

	if label.Name != "Label 2" {
		t.Errorf("Expected label name 'Label 2', got '%s'", label.Name)
	}
	if label.ProjectID != project2ID {
		t.Errorf("Expected project ID %d, got %d", project2ID, label.ProjectID)
	}
}

func TestGetLabelByID_NotFound(t *testing.T) {
	db, appInstance := testutilcli.SetupCLITest(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	// Create CLI instance
	cliInstance := &CLI{
		ctx: ctx,
		App: appInstance,
	}

	// Try to find non-existent label
	_, err := GetLabelByID(ctx, cliInstance, 9999)
	if err == nil {
		t.Fatal("Expected error for non-existent label, got nil")
	}

	// Check error message
	expectedMsg := "label 9999 not found"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestGetLabelByID_EmptyDatabase(t *testing.T) {
	db, appInstance := testutilcli.SetupCLITest(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	// Create CLI instance (no projects or labels created)
	cliInstance := &CLI{
		ctx: ctx,
		App: appInstance,
	}

	// Try to find label in empty database
	_, err := GetLabelByID(ctx, cliInstance, 1)
	if err == nil {
		t.Fatal("Expected error for label in empty database, got nil")
	}
}

// ============================================================================
// FindColumnByName Tests
// ============================================================================

func TestFindColumnByName_Found(t *testing.T) {
	columns := []*models.Column{
		{ID: 1, Name: "Todo", ProjectID: 1},
		{ID: 2, Name: "In Progress", ProjectID: 1},
		{ID: 3, Name: "Done", ProjectID: 1},
	}

	tests := []struct {
		name       string
		searchName string
		expectedID int
	}{
		{"exact match", "Todo", 1},
		{"exact match with spaces", "In Progress", 2},
		{"lowercase", "todo", 1},
		{"uppercase", "TODO", 1},
		{"mixed case", "ToDo", 1},
		{"lowercase with spaces", "in progress", 2},
		{"uppercase with spaces", "IN PROGRESS", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			col, err := FindColumnByName(columns, tt.searchName)
			assert.NoError(t, err, "FindColumnByName should not return error for: %s", tt.searchName)
			if col != nil && col.ID != tt.expectedID {
				t.Errorf("Expected column ID %d, got %d", tt.expectedID, col.ID)
			}
		})
	}
}

func TestFindColumnByName_NotFound(t *testing.T) {
	columns := []*models.Column{
		{ID: 1, Name: "Todo", ProjectID: 1},
		{ID: 2, Name: "In Progress", ProjectID: 1},
		{ID: 3, Name: "Done", ProjectID: 1},
	}

	tests := []string{
		"Nonexistent",
		"Doing",
		"",
		"Tod", // partial match should not work
		"Todoo",
	}

	for _, searchName := range tests {
		t.Run(searchName, func(t *testing.T) {
			_, err := FindColumnByName(columns, searchName)
			if err == nil {
				t.Errorf("Expected error for '%s', got nil", searchName)
			}
		})
	}
}

func TestFindColumnByName_EmptyList(t *testing.T) {
	columns := []*models.Column{}

	_, err := FindColumnByName(columns, "Todo")
	if err == nil {
		t.Error("Expected error for empty column list, got nil")
	}
}

// ============================================================================
// FormatAvailableColumns Tests
// ============================================================================

func TestFormatAvailableColumns(t *testing.T) {
	tests := []struct {
		name     string
		columns  []*models.Column
		expected string
	}{
		{
			name: "multiple columns",
			columns: []*models.Column{
				{ID: 1, Name: "Todo"},
				{ID: 2, Name: "In Progress"},
				{ID: 3, Name: "Done"},
			},
			expected: "Todo, In Progress, Done",
		},
		{
			name: "single column",
			columns: []*models.Column{
				{ID: 1, Name: "Backlog"},
			},
			expected: "Backlog",
		},
		{
			name:     "empty list",
			columns:  []*models.Column{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatAvailableColumns(tt.columns)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// ============================================================================
// GetCurrentColumnName Tests
// ============================================================================

func TestGetCurrentColumnName(t *testing.T) {
	columns := []*models.Column{
		{ID: 1, Name: "Todo", ProjectID: 1},
		{ID: 2, Name: "In Progress", ProjectID: 1},
		{ID: 3, Name: "Done", ProjectID: 1},
	}

	tests := []struct {
		name     string
		columnID int
		expected string
	}{
		{"first column", 1, "Todo"},
		{"middle column", 2, "In Progress"},
		{"last column", 3, "Done"},
		{"non-existent column", 999, "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetCurrentColumnName(columns, tt.columnID)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestGetCurrentColumnName_EmptyList(t *testing.T) {
	columns := []*models.Column{}

	result := GetCurrentColumnName(columns, 1)
	if result != "Unknown" {
		t.Errorf("Expected 'Unknown' for empty list, got '%s'", result)
	}
}

// ============================================================================
// GetProjectID Tests
// ============================================================================

func TestGetProjectID_FlagSet(t *testing.T) {
	// Create a command with the --project flag
	cmd := &cobra.Command{
		Use: "test",
		Run: func(cmd *cobra.Command, args []string) {},
	}
	cmd.Flags().Int("project", 0, "Project ID")

	// Set the flag value
	err := cmd.Flags().Set("project", "42")
	require.NoError(t, err, "Failed to set project flag")

	// Test getting the project ID
	projectID, err := GetProjectID(cmd)
	assert.NoError(t, err, "GetProjectID should not return error")
	if projectID != 42 {
		t.Errorf("Expected project ID 42, got %d", projectID)
	}
}

// TestGetProjectID_EnvVarSet removed - PASO_PROJECT env var no longer supported
// Git branch association has replaced the environment variable approach

// TestGetProjectID_FlagTakesPrecedence removed - PASO_PROJECT env var no longer supported
// See TestGetProjectID_FlagTakesPrecedenceOverGitBranch for git branch precedence tests

func TestGetProjectID_NeitherSet(t *testing.T) {
	// Create a command without setting the flag
	cmd := &cobra.Command{
		Use: "test",
		Run: func(cmd *cobra.Command, args []string) {},
	}
	cmd.Flags().Int("project", 0, "Project ID")

	// Test that we get an error
	_, err := GetProjectID(cmd)
	if err == nil {
		t.Error("Expected error when flag not set and not in git repo, got nil")
	}

	// Check error message
	expectedMsg := "no project specified: use --project flag or create a project associated with this branch"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestGetProjectID_InvalidEnvVar removed - PASO_PROJECT env var no longer supported
// TestGetProjectID_NoProjectFlag removed - PASO_PROJECT env var no longer supported
// TestGetProjectID_ZeroValueFlag removed - PASO_PROJECT env var no longer supported

// ============================================================================
// NEW GIT BRANCH DETECTION TESTS (TDD RED PHASE)
// ============================================================================

// These tests verify the new behavior where GetProjectID uses git branch detection
// instead of environment variables

func TestGetProjectID_GitBranchDetection_NotImplementedYet(t *testing.T) {
	// This test will fail until git branch detection is implemented
	// It verifies the precedence: --project flag > git branch > error

	t.Skip("Skipping until git branch detection is implemented in GetProjectID")

	// The new implementation should:
	// 1. Check if --project flag is set (explicitly changed)
	// 2. If not, detect git branch and lookup project
	// 3. If no git branch or project not found, return error

	// This test would look something like:
	// db, app := testutilcli.SetupCLITest(t)
	// defer db.Close()
	//
	// Create project with git branch in DB
	// result, _ := db.ExecContext(ctx, "INSERT INTO projects (name, git_branch) VALUES (?, ?)", "Test", "main")
	// projectID, _ := result.LastInsertId()
	//
	// Create command without --project flag
	// cmd := &cobra.Command{Use: "test"}
	// cmd.Flags().Int("project", 0, "Project ID")
	//
	// Set context with CLI instance
	// ctx := context.WithValue(context.Background(), TestAppKey, cliInstance)
	// cmd.SetContext(ctx)
	//
	// Call GetProjectID - should detect git branch and return project ID
	// id, err := GetProjectID(cmd)
	// assert.NoError(t, err)
	// assert.Equal(t, int(projectID), id)
}

func TestGetProjectID_FlagTakesPrecedenceOverGitBranch(t *testing.T) {
	// This test verifies that --project flag takes precedence over git branch detection
	// When the flag is explicitly set, git detection should not run

	t.Skip("Skipping until git branch detection is implemented")

	// Test setup:
	// 1. Create two projects in DB:
	//    - Project 1 (id=1) with git_branch="main"
	//    - Project 2 (id=2) with git_branch="feature/other"
	// 2. Simulate being on "main" branch (would detect Project 1)
	// 3. Set --project flag to 2
	// 4. GetProjectID should return 2 (from flag), not 1 (from git)

	// Expected behavior:
	// projectID, err := GetProjectID(cmd)
	// assert.NoError(t, err)
	// assert.Equal(t, 2, projectID, "Flag should take precedence over git branch")
}

func TestGetProjectID_GitBranchFound(t *testing.T) {
	// This test verifies that GetProjectID can find a project by git branch
	// when no --project flag is set

	t.Skip("Skipping until git branch detection is implemented")

	// Test setup:
	// 1. Create project with git_branch="feature/test"
	// 2. Simulate being on "feature/test" branch
	// 3. Don't set --project flag
	// 4. GetProjectID should find and return the project

	// Expected behavior:
	// projectID, err := GetProjectID(cmd)
	// assert.NoError(t, err)
	// assert.Equal(t, expectedID, projectID)
}

func TestGetProjectID_GitBranchNotAssociated(t *testing.T) {
	// This test verifies error when on a git branch with no associated project

	t.Skip("Skipping until git branch detection is implemented")

	// Test setup:
	// 1. Create project with git_branch="main"
	// 2. Simulate being on "feature/unassociated" branch
	// 3. Don't set --project flag
	// 4. GetProjectID should return error

	// Expected behavior:
	// _, err := GetProjectID(cmd)
	// assert.Error(t, err)
	// assert.Contains(t, err.Error(), "feature/unassociated")
	// assert.Contains(t, err.Error(), "no project associated")
}

func TestGetProjectID_NotInGitRepo(t *testing.T) {
	// This test verifies error when not in a git repository

	t.Skip("Skipping until git branch detection is implemented")

	// Test setup:
	// 1. Change to a non-git directory
	// 2. Don't set --project flag
	// 3. GetProjectID should return error

	// Expected behavior:
	// _, err := GetProjectID(cmd)
	// assert.Error(t, err)
	// assert.Contains(t, err.Error(), "no project specified")
	// assert.Contains(t, err.Error(), "--project flag")
}

func TestGetProjectID_DetachedHead(t *testing.T) {
	// This test verifies error when in detached HEAD state

	t.Skip("Skipping until git branch detection is implemented")

	// Test setup:
	// 1. Simulate detached HEAD state (IsDetached=true)
	// 2. Don't set --project flag
	// 3. GetProjectID should return error

	// Expected behavior:
	// _, err := GetProjectID(cmd)
	// assert.Error(t, err)
	// assert.Contains(t, err.Error(), "no project specified")
}

func TestGetProjectID_EmptyRepo(t *testing.T) {
	// This test verifies error when in a git repo with no commits

	t.Skip("Skipping until git branch detection is implemented")

	// Test setup:
	// 1. Simulate empty repo (HasCommits=false)
	// 2. Don't set --project flag
	// 3. GetProjectID should return error

	// Expected behavior:
	// _, err := GetProjectID(cmd)
	// assert.Error(t, err)
}

func TestGetProjectID_FlagNotChanged(t *testing.T) {
	// This test verifies that GetProjectID uses git detection
	// when --project flag exists but was not explicitly set (Changed=false)

	t.Skip("Skipping until git branch detection is implemented")

	// Test setup:
	// 1. Create project with git_branch="main"
	// 2. Create command with --project flag (but don't set it)
	// 3. Flag should have Changed=false
	// 4. GetProjectID should fall through to git detection

	// Expected behavior:
	// cmd := &cobra.Command{Use: "test"}
	// cmd.Flags().Int("project", 0, "Project ID")
	// // Don't call cmd.Flags().Set("project", ...)
	//
	// projectID, err := GetProjectID(cmd)
	// assert.NoError(t, err)
	// assert.Equal(t, expectedID, projectID)
}

func TestGetProjectID_BranchWithSlashes(t *testing.T) {
	// This test verifies that git branches with slashes work correctly

	t.Skip("Skipping until git branch detection is implemented")

	// Test setup:
	// 1. Create project with git_branch="feature/auth/user-login"
	// 2. Simulate being on that branch
	// 3. GetProjectID should find the project

	// Expected behavior:
	// projectID, err := GetProjectID(cmd)
	// assert.NoError(t, err)
	// assert.Equal(t, expectedID, projectID)
}

func TestGetProjectID_ErrorMessage(t *testing.T) {
	// This test verifies that error messages are helpful and include the branch name

	t.Skip("Skipping until git branch detection is implemented")

	// Test setup:
	// 1. Simulate being on "feature/my-feature" with no associated project
	// 2. Error message should mention the branch name
	// 3. Error message should suggest using --project flag

	// Expected behavior:
	// _, err := GetProjectID(cmd)
	// assert.Error(t, err)
	// assert.Contains(t, err.Error(), "feature/my-feature")
	// assert.Contains(t, err.Error(), "--project flag")
}

func TestGetProjectID_CLICleanup(t *testing.T) {
	// This test verifies that GetProjectID properly cleans up CLI instance

	t.Skip("Skipping until git branch detection is implemented")

	// Test that the CLI instance created in GetProjectID is properly closed
	// This prevents resource leaks

	// The implementation should use defer to close the CLI instance:
	// defer func() {
	//     if closeErr := cliInstance.Close(); closeErr != nil {
	//         slog.Warn("failed to close CLI in GetProjectID", "error", closeErr)
	//     }
	// }()
}

// ============================================================================
// INTEGRATION TESTS WITH MULTIPLE SCENARIOS
// ============================================================================

func TestGetProjectID_Scenarios_TableDriven(t *testing.T) {
	// This is a comprehensive table-driven test for all GetProjectID scenarios

	t.Skip("Skipping until git branch detection is implemented")

	tests := []struct {
		name          string
		setupFunc     func(t *testing.T) (*cobra.Command, context.Context)
		expectError   bool
		expectedID    int
		errorContains string
		description   string
	}{
		{
			name:        "flag_set_explicitly",
			expectError: false,
			expectedID:  42,
			description: "When --project flag is set, use it",
		},
		{
			name:        "git_branch_found",
			expectError: false,
			expectedID:  10,
			description: "When on git branch with project, find it",
		},
		{
			name:          "git_branch_not_found",
			expectError:   true,
			errorContains: "no project associated",
			description:   "When on git branch without project, error",
		},
		{
			name:          "not_in_git_repo",
			expectError:   true,
			errorContains: "no project specified",
			description:   "When not in git repo, error",
		},
		{
			name:          "detached_head",
			expectError:   true,
			errorContains: "no project specified",
			description:   "When in detached HEAD, error",
		},
		{
			name:          "empty_repo",
			expectError:   true,
			errorContains: "no project specified",
			description:   "When in empty repo, error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test
			// cmd, ctx := tt.setupFunc(t)

			// Execute
			// projectID, err := GetProjectID(cmd)

			// Verify
			// if tt.expectError {
			//     assert.Error(t, err)
			//     if tt.errorContains != "" {
			//         assert.Contains(t, err.Error(), tt.errorContains)
			//     }
			// } else {
			//     assert.NoError(t, err)
			//     assert.Equal(t, tt.expectedID, projectID)
			// }
		})
	}
}
