package tui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// SnapshotHelper manages snapshot file operations for TUI testing
type SnapshotHelper struct {
	t           *testing.T
	updateMode  bool
	snapshotDir string
}

// NewSnapshotHelper creates a new snapshot helper
func NewSnapshotHelper(t *testing.T) *SnapshotHelper {
	return &SnapshotHelper{
		t:           t,
		updateMode:  os.Getenv("UPDATE_SNAPSHOTS") == "1",
		snapshotDir: "testdata/snapshots",
	}
}

// Compare compares rendered output against a golden file
// If UPDATE_SNAPSHOTS=1, updates the golden file instead of comparing
func (sh *SnapshotHelper) Compare(name, output string) {
	sh.t.Helper()

	// Ensure snapshot directory exists
	err := os.MkdirAll(sh.snapshotDir, 0755)
	require.NoError(sh.t, err, "Failed to create snapshot directory")

	snapshotPath := filepath.Join(sh.snapshotDir, name+".golden")

	if sh.updateMode {
		// Write/update golden file
		err := os.WriteFile(snapshotPath, []byte(output), 0644)
		require.NoError(sh.t, err, "Failed to write snapshot file")
		sh.t.Logf("Updated snapshot: %s", snapshotPath)
		return
	}

	// Compare against existing golden file
	golden, err := os.ReadFile(snapshotPath)
	if err != nil {
		if os.IsNotExist(err) {
			sh.t.Logf("Snapshot file does not exist, creating: %s\nRun UPDATE_SNAPSHOTS=1 to create/update", snapshotPath)
			// Write initial snapshot for first run
			err := os.WriteFile(snapshotPath, []byte(output), 0644)
			require.NoError(sh.t, err, "Failed to write initial snapshot")
			return
		}
		require.NoError(sh.t, err, "Failed to read snapshot file")
	}

	// Compare output
	assert.Equal(sh.t, string(golden), output, "Snapshot mismatch for %s\n\nRun UPDATE_SNAPSHOTS=1 to update", name)
}

// WriteSnapshot writes a snapshot file directly
func (sh *SnapshotHelper) WriteSnapshot(name, content string) {
	sh.t.Helper()

	err := os.MkdirAll(sh.snapshotDir, 0755)
	require.NoError(sh.t, err, "Failed to create snapshot directory")

	snapshotPath := filepath.Join(sh.snapshotDir, name+".golden")
	err = os.WriteFile(snapshotPath, []byte(content), 0644)
	require.NoError(sh.t, err, "Failed to write snapshot")
}

// ReadSnapshot reads an existing snapshot file
func (sh *SnapshotHelper) ReadSnapshot(name string) (string, error) {
	snapshotPath := filepath.Join(sh.snapshotDir, name+".golden")
	content, err := os.ReadFile(snapshotPath)
	return string(content), err
}
