package testutil

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/app"
)

// ContextKey is a custom type for context keys to avoid collisions
type ContextKey string

const TestAppKey ContextKey = "testApp"

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

// ExecuteCLICommand executes a CLI command with a test app instance
// This properly injects the app context so commands can access the test database
// Note: The cliInstance will be created by GetCLIFromContext in the CLI package
func ExecuteCLICommand(t *testing.T, testApp *app.App, cmd *cobra.Command, args []string) (string, error) {
	t.Helper()

	if testApp == nil {
		t.Fatal("testApp cannot be nil - SetupCLITest must be called first")
	}

	ctx := context.Background()
	return ExecuteCLICommandWithContext(t, ctx, testApp, cmd, args)
}

// ExecuteCLICommandWithContext executes a CLI command with a specific context and test app
func ExecuteCLICommandWithContext(t *testing.T, ctx context.Context, testApp *app.App, cmd *cobra.Command, args []string) (string, error) {
	t.Helper()

	if testApp == nil {
		t.Fatal("testApp cannot be nil - SetupCLITest must be called first")
	}

	// Set command args
	cmd.SetArgs(args)

	// Create a wrapper context that will be recognized by GetCLIFromContext in CLI package
	// We pass the app instance through the context
	ctxWithApp := context.WithValue(ctx, TestAppKey, testApp)

	// Set the context on the command
	cmd.SetContext(ctxWithApp)

	// Disable usage output on error for cleaner test output
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	// Capture output and execute
	var output string
	var executeErr error

	output = CaptureOutput(t, func() {
		executeErr = cmd.ExecuteContext(ctxWithApp)
	})

	return output, executeErr
}
