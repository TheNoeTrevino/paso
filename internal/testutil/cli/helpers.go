package cli

import (
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/testutil"
)

// ExecuteCommand runs a cobra command and captures its output
func ExecuteCommand(t *testing.T, cmd *cobra.Command) (string, error) {
	t.Helper()

	// Capture stdout
	var output string
	var executeErr error

	output = testutil.CaptureOutput(t, func() {
		executeErr = cmd.Execute()
	})

	return output, executeErr
}

// ParseJSON parses JSON output from CLI commands
func ParseJSON(t *testing.T, output string) map[string]interface{} {
	t.Helper()

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, output)
	}

	return result
}

// SetupCobraCommand sets up a cobra command with args for testing
func SetupCobraCommand(cmd *cobra.Command, args []string) {
	cmd.SetArgs(args)
	// Disable usage output on error for cleaner test output
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
}
