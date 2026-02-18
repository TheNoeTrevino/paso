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
	"github.com/thenoetrevino/paso/internal/appcontext"
	"github.com/thenoetrevino/paso/internal/git"
	"github.com/thenoetrevino/paso/internal/github"
	assigneeservice "github.com/thenoetrevino/paso/internal/services/assignee"
	columnservice "github.com/thenoetrevino/paso/internal/services/column"
	labelservice "github.com/thenoetrevino/paso/internal/services/label"
	projectservice "github.com/thenoetrevino/paso/internal/services/project"
	taskservice "github.com/thenoetrevino/paso/internal/services/task"
)

// stdoutStderrMutex protects concurrent access to os.Stdout and os.Stderr.
// This is necessary because tests can run in parallel and all modify global stdio.
var stdoutStderrMutex sync.Mutex

// CaptureOutputFunc captures stdout and stderr during function execution.
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
	wg.Go(func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, rOut)
		outC <- buf.String()
	})

	// Goroutine to read stderr
	wg.Go(func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, rErr)
		errC <- buf.String()
	})

	// Execute function
	fn()

	// Close writers — this signals EOF to the goroutines
	_ = wOut.Close()
	_ = wErr.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	// Wait for both goroutines to finish reading before collecting output
	wg.Wait()

	// Get captured output — both channels should have data now
	stdoutText := <-outC
	stderrText := <-errC

	// Return combined output (stderr first since errors are more important)
	if stderrText != "" {
		return stderrText + stdoutText
	}
	return stdoutText
}

// MockServices holds mock service implementations for CLI unit testing.
// Pass only the services your test needs; unused fields default to nil.
type MockServices struct {
	TaskService     taskservice.Service
	ProjectService  projectservice.Service
	ColumnService   columnservice.Service
	LabelService    labelservice.Service
	AssigneeService assigneeservice.Service
	GitDetector     git.Detector
	GitHubFetcher   github.IssueFetcher
}

// ExecuteCLICommandWithMocks executes a CLI command with mock services injected.
// Unlike ExecuteCLICommand which requires a real database, this builds an *app.App
// backed entirely by mock services for fast, isolated unit testing.
func ExecuteCLICommandWithMocks(t *testing.T, services MockServices, cmd *cobra.Command, args []string) (string, error) {
	t.Helper()

	testApp := &app.App{
		TaskService:     services.TaskService,
		ProjectService:  services.ProjectService,
		ColumnService:   services.ColumnService,
		LabelService:    services.LabelService,
		AssigneeService: services.AssigneeService,
		GitDetector:     services.GitDetector,
		GitHubFetcher:   services.GitHubFetcher,
	}

	return ExecuteCLICommand(t, testApp, cmd, args)
}

// ExecuteCLICommand executes a CLI command with a test app instance.
// This properly injects the app context so commands can access the test database.
// Note: The cliInstance will be created by GetCLIFromContext in the CLI package.
func ExecuteCLICommand(t *testing.T, testApp *app.App, cmd *cobra.Command, args []string) (string, error) {
	t.Helper()

	if testApp == nil {
		t.Fatal("testApp cannot be nil - SetupCLITest must be called first")
	}

	ctx := context.Background()
	return ExecuteCLICommandWithContext(t, ctx, testApp, cmd, args)
}

// ExecuteCLICommandWithContext executes a CLI command with a specific context and test app.
func ExecuteCLICommandWithContext(t *testing.T, ctx context.Context, testApp *app.App, cmd *cobra.Command, args []string) (string, error) {
	t.Helper()

	if testApp == nil {
		t.Fatal("testApp cannot be nil - SetupCLITest must be called first")
	}

	// Set command args
	cmd.SetArgs(args)

	// Create a wrapper context that will be recognized by GetCLIFromContext in CLI package.
	// We pass the app instance through the context.
	ctxWithApp := context.WithValue(ctx, appcontext.AppKey, testApp)

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
