package cli

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
