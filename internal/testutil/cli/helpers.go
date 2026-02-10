package cli

import (
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/testutil"
)

// exitCoder is satisfied by *cli.ExitErr without importing the cli package,
// which would create an import cycle when tests in internal/cli use this helper.
type exitCoder interface {
	error
	ExitCode() int
}

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
func ParseJSON(t *testing.T, output string) map[string]any {
	t.Helper()

	var result map[string]any
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

// AssertExitError asserts that an error implements the exitCoder interface (i.e. is a *cli.ExitErr)
// with the expected exit code. It fails the test immediately if err is nil or does not implement
// exitCoder (using require), then checks the exit code (using assert) so remaining assertions
// in the caller can still run.
func AssertExitError(t *testing.T, err error, expectedCode int) {
	t.Helper()
	require.Error(t, err, "expected an error but got nil")
	ec, ok := err.(exitCoder)
	if !ok {
		// errors.As unwraps; try that too
		var target exitCoder
		ok = assert.ErrorAs(t, err, &target, "expected error to implement exitCoder (ExitCode() int), got %T: %v", err, err)
		if !ok {
			return
		}
		ec = target
	}
	assert.Equal(t, expectedCode, ec.ExitCode(), "expected exit code %d but got %d; error: %s", expectedCode, ec.ExitCode(), err)
}
