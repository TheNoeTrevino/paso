package fixtures

import (
	"os"
	"testing"
)

// ChdirTemp changes the working directory to dir and registers a
// t.Cleanup to restore the original directory when the test finishes.
func ChdirTemp(tb testing.TB, dir string) {
	tb.Helper()
	original, err := os.Getwd()
	if err != nil {
		tb.Fatalf("failed to get working directory: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		tb.Fatalf("failed to chdir to %s: %v", dir, err)
	}
	tb.Cleanup(func() {
		_ = os.Chdir(original)
	})
}
