// Package snapshots provides a shared golden file comparison helper for all snapshot tests.
//
// All snapshot tests (CLI help, CLI styles, CLI command output, TUI rendering)
// use this helper for consistent golden file management.
//
// To update snapshots:
//
//	UPDATE_SNAPSHOTS=1 go test ./internal/testing/snapshots/...
package snapshots

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper manages golden file comparison for snapshot testing.
type Helper struct {
	t           *testing.T
	updateMode  bool
	snapshotDir string
}

// NewHelper creates a snapshot Helper rooted at the given directory.
// Golden files are stored as <dir>/<name>.golden.
// When UPDATE_SNAPSHOTS=1 is set, golden files are written instead of compared.
func NewHelper(t *testing.T, dir string) *Helper {
	t.Helper()
	return &Helper{
		t:           t,
		updateMode:  os.Getenv("UPDATE_SNAPSHOTS") == "1",
		snapshotDir: dir,
	}
}

// Compare compares rendered output against a golden file.
// If UPDATE_SNAPSHOTS=1, the golden file is written/updated instead of compared.
// If the golden file does not exist on first run, it is created automatically.
func (h *Helper) Compare(name, output string) {
	h.t.Helper()

	err := os.MkdirAll(h.snapshotDir, 0755)
	require.NoError(h.t, err, "failed to create snapshot directory")

	goldenPath := filepath.Join(h.snapshotDir, name+".golden")

	if h.updateMode {
		err := os.WriteFile(goldenPath, []byte(output), 0644)
		require.NoError(h.t, err, "failed to write snapshot file")
		h.t.Logf("Updated snapshot: %s", goldenPath)
		return
	}

	golden, err := os.ReadFile(goldenPath)
	if err != nil {
		if os.IsNotExist(err) {
			h.t.Logf("Snapshot does not exist, creating: %s\nRun UPDATE_SNAPSHOTS=1 to update", goldenPath)
			err := os.WriteFile(goldenPath, []byte(output), 0644)
			require.NoError(h.t, err, "failed to write initial snapshot")
			return
		}
		require.NoError(h.t, err, "failed to read snapshot file")
	}

	assert.Equal(h.t, string(golden), output,
		"snapshot mismatch for %s\n\nRun UPDATE_SNAPSHOTS=1 go test ./internal/testing/snapshots/... to update", name)
}

// Read reads an existing golden file and returns its content.
func (h *Helper) Read(name string) (string, error) {
	goldenPath := filepath.Join(h.snapshotDir, name+".golden")
	content, err := os.ReadFile(goldenPath)
	return string(content), err
}
