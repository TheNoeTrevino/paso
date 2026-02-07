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
	assert.Equal(t, 42, projectID)
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
	require.Error(t, err)

	// Check error message
	assert.Equal(t, "no project specified: use --project flag or create a project associated with this branch", err.Error())
}
