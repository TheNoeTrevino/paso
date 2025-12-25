package cli

import (
	"context"
	"testing"

	"github.com/thenoetrevino/paso/internal/testutil"
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
	db, appInstance := testutil.SetupCLITest(t)
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
	db, appInstance := testutil.SetupCLITest(t)
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
	db, appInstance := testutil.SetupCLITest(t)
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
	db, appInstance := testutil.SetupCLITest(t)
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
