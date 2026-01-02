package cli

import (
	"context"
	"os"
	"testing"

	"github.com/spf13/cobra"
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
			if err != nil {
				t.Errorf("Expected %s to be valid, got error: %v", color, err)
			}
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
			if err != nil {
				t.Errorf("Expected no error for '%s', got: %v", tt.input, err)
			}
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
			if err != nil {
				t.Errorf("Expected no error for '%s', got: %v", tt.input, err)
			}
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
	if err != nil {
		t.Fatalf("Expected to find label, got error: %v", err)
	}

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
	if err != nil {
		t.Fatalf("Expected to find label, got error: %v", err)
	}

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
			if err != nil {
				t.Errorf("Expected to find column '%s', got error: %v", tt.searchName, err)
			}
			if col.ID != tt.expectedID {
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
	if err != nil {
		t.Fatalf("Failed to set project flag: %v", err)
	}

	// Test getting the project ID
	projectID, err := GetProjectID(cmd)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if projectID != 42 {
		t.Errorf("Expected project ID 42, got %d", projectID)
	}
}

func TestGetProjectID_EnvVarSet(t *testing.T) {
	// Save and restore original env var
	originalEnv := os.Getenv("PASO_PROJECT")
	defer func() {
		if originalEnv != "" {
			os.Setenv("PASO_PROJECT", originalEnv)
		} else {
			os.Unsetenv("PASO_PROJECT")
		}
	}()

	// Set environment variable
	os.Setenv("PASO_PROJECT", "123")

	// Create a command without setting the flag
	cmd := &cobra.Command{
		Use: "test",
		Run: func(cmd *cobra.Command, args []string) {},
	}
	cmd.Flags().Int("project", 0, "Project ID")

	// Test getting the project ID from env var
	projectID, err := GetProjectID(cmd)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if projectID != 123 {
		t.Errorf("Expected project ID 123, got %d", projectID)
	}
}

func TestGetProjectID_FlagTakesPrecedence(t *testing.T) {
	// Save and restore original env var
	originalEnv := os.Getenv("PASO_PROJECT")
	defer func() {
		if originalEnv != "" {
			os.Setenv("PASO_PROJECT", originalEnv)
		} else {
			os.Unsetenv("PASO_PROJECT")
		}
	}()

	// Set environment variable
	os.Setenv("PASO_PROJECT", "100")

	// Create a command and set the flag
	cmd := &cobra.Command{
		Use: "test",
		Run: func(cmd *cobra.Command, args []string) {},
	}
	cmd.Flags().Int("project", 0, "Project ID")
	err := cmd.Flags().Set("project", "200")
	if err != nil {
		t.Fatalf("Failed to set project flag: %v", err)
	}

	// Test that flag takes precedence over env var
	projectID, err := GetProjectID(cmd)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if projectID != 200 {
		t.Errorf("Expected project ID 200 (from flag), got %d", projectID)
	}
}

func TestGetProjectID_NeitherSet(t *testing.T) {
	// Save and restore original env var
	originalEnv := os.Getenv("PASO_PROJECT")
	defer func() {
		if originalEnv != "" {
			os.Setenv("PASO_PROJECT", originalEnv)
		} else {
			os.Unsetenv("PASO_PROJECT")
		}
	}()

	// Ensure env var is not set
	os.Unsetenv("PASO_PROJECT")

	// Create a command without setting the flag
	cmd := &cobra.Command{
		Use: "test",
		Run: func(cmd *cobra.Command, args []string) {},
	}
	cmd.Flags().Int("project", 0, "Project ID")

	// Test that we get an error
	_, err := GetProjectID(cmd)
	if err == nil {
		t.Error("Expected error when neither flag nor env var is set, got nil")
	}

	// Check error message
	expectedMsg := "no project specified: use --project flag or set with 'eval $(paso use project <project-id>)'"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestGetProjectID_InvalidEnvVar(t *testing.T) {
	// Save and restore original env var
	originalEnv := os.Getenv("PASO_PROJECT")
	defer func() {
		if originalEnv != "" {
			os.Setenv("PASO_PROJECT", originalEnv)
		} else {
			os.Unsetenv("PASO_PROJECT")
		}
	}()

	// Set invalid environment variable (non-numeric)
	os.Setenv("PASO_PROJECT", "invalid")

	// Create a command without setting the flag
	cmd := &cobra.Command{
		Use: "test",
		Run: func(cmd *cobra.Command, args []string) {},
	}
	cmd.Flags().Int("project", 0, "Project ID")

	// Test that we get an error for invalid env var
	_, err := GetProjectID(cmd)
	if err == nil {
		t.Error("Expected error for invalid env var format, got nil")
	}
}

func TestGetProjectID_NoProjectFlag(t *testing.T) {
	// Save and restore original env var
	originalEnv := os.Getenv("PASO_PROJECT")
	defer func() {
		if originalEnv != "" {
			os.Setenv("PASO_PROJECT", originalEnv)
		} else {
			os.Unsetenv("PASO_PROJECT")
		}
	}()

	// Set environment variable
	os.Setenv("PASO_PROJECT", "456")

	// Create a command WITHOUT the --project flag
	cmd := &cobra.Command{
		Use: "test",
		Run: func(cmd *cobra.Command, args []string) {},
	}

	// Test that env var still works when flag doesn't exist
	projectID, err := GetProjectID(cmd)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if projectID != 456 {
		t.Errorf("Expected project ID 456, got %d", projectID)
	}
}

func TestGetProjectID_ZeroValueFlag(t *testing.T) {
	// Save and restore original env var
	originalEnv := os.Getenv("PASO_PROJECT")
	defer func() {
		if originalEnv != "" {
			os.Setenv("PASO_PROJECT", originalEnv)
		} else {
			os.Unsetenv("PASO_PROJECT")
		}
	}()

	// Set environment variable
	os.Setenv("PASO_PROJECT", "789")

	// Create a command with the --project flag but don't set it
	cmd := &cobra.Command{
		Use: "test",
		Run: func(cmd *cobra.Command, args []string) {},
	}
	cmd.Flags().Int("project", 0, "Project ID")

	// Test that env var is used when flag is not changed (even if it's 0)
	projectID, err := GetProjectID(cmd)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if projectID != 789 {
		t.Errorf("Expected project ID 789 (from env var), got %d", projectID)
	}
}
