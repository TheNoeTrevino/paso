package testutil

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/spf13/cobra"
)

// CaptureOutput captures stdout during function execution
func CaptureOutput(t *testing.T, fn func()) string {
	t.Helper()

	// Save original stdout
	oldStdout := os.Stdout

	// Create pipe to capture output
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}

	// Replace stdout with pipe writer
	os.Stdout = w

	// Channel to collect output
	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		outC <- buf.String()
	}()

	// Execute function
	fn()

	// Close writer and restore stdout
	_ = w.Close()
	os.Stdout = oldStdout

	// Get captured output
	return <-outC
}

// ExecuteCommand runs a cobra command and captures its output
func ExecuteCommand(t *testing.T, cmd *cobra.Command) (string, error) {
	t.Helper()

	// Capture stdout
	var output string
	var executeErr error

	output = CaptureOutput(t, func() {
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
