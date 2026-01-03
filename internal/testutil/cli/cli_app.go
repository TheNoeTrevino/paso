package cli

import (
	"bytes"
	"context"
	"io"
	"os"
	"sync"
	"testing"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/app"
	"github.com/thenoetrevino/paso/internal/testutil"
)

// stdoutStderrMutex protects concurrent access to os.Stdout and os.Stderr
// This is necessary because tests can run in parallel and all modify global stdio
var stdoutStderrMutex sync.Mutex

// CaptureOutputFunc captures stdout and stderr during function execution
func CaptureOutputFunc(t *testing.T, fn func()) string {
	t.Helper()

	// Lock to prevent concurrent modification of global stdout/stderr
	stdoutStderrMutex.Lock()
	defer stdoutStderrMutex.Unlock()

	// Save original stdout and stderr
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	// Create pipes to capture output
	rOut, wOut, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create stdout pipe: %v", err)
	}
	rErr, wErr, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create stderr pipe: %v", err)
	}

	// Replace stdout and stderr with pipe writers
	os.Stdout = wOut
	os.Stderr = wErr

	// Use buffered channels and WaitGroup to ensure proper synchronization
	var wg sync.WaitGroup
	outC := make(chan string, 1)
	errC := make(chan string, 1)

	// Goroutine to read stdout
	wg.Add(1)
	go func() {
		defer wg.Done()
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, rOut)
		outC <- buf.String()
	}()

	// Goroutine to read stderr
	wg.Add(1)
	go func() {
		defer wg.Done()
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, rErr)
		errC <- buf.String()
	}()

	// Execute function
	fn()

	// Close writers - this signals EOF to the goroutines
	_ = wOut.Close()
	_ = wErr.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	// Wait for both goroutines to finish reading before collecting output
	wg.Wait()

	// Get captured output - both channels should have data now
	stdoutText := <-outC
	stderrText := <-errC

	// Return combined output (stderr first since errors are more important)
	if stderrText != "" {
		return stderrText + stdoutText
	}
	return stdoutText
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
	ctxWithApp := context.WithValue(ctx, testutil.TestAppKey, testApp)

	// Set the context on the command
	cmd.SetContext(ctxWithApp)

	// Disable usage output on error for cleaner test output
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	// Capture output and execute
	var output string
	var executeErr error

	output = CaptureOutputFunc(t, func() {
		executeErr = cmd.ExecuteContext(ctxWithApp)
	})

	return output, executeErr
}
