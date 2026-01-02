package tui

import (
	"os"
	"path/filepath"
	"testing"
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
	if err := os.MkdirAll(sh.snapshotDir, 0755); err != nil {
		sh.t.Fatalf("Failed to create snapshot directory: %v", err)
	}

	snapshotPath := filepath.Join(sh.snapshotDir, name+".golden")

	if sh.updateMode {
		// Write/update golden file
		if err := os.WriteFile(snapshotPath, []byte(output), 0644); err != nil {
			sh.t.Fatalf("Failed to write snapshot file: %v", err)
		}
		sh.t.Logf("Updated snapshot: %s", snapshotPath)
		return
	}

	// Compare against existing golden file
	golden, err := os.ReadFile(snapshotPath)
	if err != nil {
		if os.IsNotExist(err) {
			sh.t.Logf("Snapshot file does not exist, creating: %s\nRun UPDATE_SNAPSHOTS=1 to create/update", snapshotPath)
			// Write initial snapshot for first run
			if err := os.WriteFile(snapshotPath, []byte(output), 0644); err != nil {
				sh.t.Fatalf("Failed to write initial snapshot: %v", err)
			}
			return
		}
		sh.t.Fatalf("Failed to read snapshot file: %v", err)
	}

	// Compare output
	if string(golden) != output {
		sh.t.Errorf("Snapshot mismatch for %s\n\nExpected:\n%s\n\nGot:\n%s\n\nRun UPDATE_SNAPSHOTS=1 to update",
			name, string(golden), output)
	}
}

// WriteSnapshot writes a snapshot file directly
func (sh *SnapshotHelper) WriteSnapshot(name, content string) {
	sh.t.Helper()

	if err := os.MkdirAll(sh.snapshotDir, 0755); err != nil {
		sh.t.Fatalf("Failed to create snapshot directory: %v", err)
	}

	snapshotPath := filepath.Join(sh.snapshotDir, name+".golden")
	if err := os.WriteFile(snapshotPath, []byte(content), 0644); err != nil {
		sh.t.Fatalf("Failed to write snapshot: %v", err)
	}
}

// ReadSnapshot reads an existing snapshot file
func (sh *SnapshotHelper) ReadSnapshot(name string) (string, error) {
	snapshotPath := filepath.Join(sh.snapshotDir, name+".golden")
	content, err := os.ReadFile(snapshotPath)
	return string(content), err
}
