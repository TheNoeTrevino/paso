package install

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmbeddedContent_NotEmpty(t *testing.T) {
	t.Run("pasoSkillContent is populated", func(t *testing.T) {
		assert.NotEmpty(t, pasoSkillContent, "pasoSkillContent should not be empty")
	})

	t.Run("pasoDelegateContent is populated", func(t *testing.T) {
		assert.NotEmpty(t, pasoDelegateContent, "pasoDelegateContent should not be empty")
	})

	t.Run("pasoSetupContent is populated", func(t *testing.T) {
		assert.NotEmpty(t, pasoSetupContent, "pasoSetupContent should not be empty")
	})
}

func TestInstallSelected(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	allItems := allKeys()
	allTargetKeys := make([]string, len(targets))
	for i, tgt := range targets {
		allTargetKeys[i] = tgt.key
	}

	err := installSelected(allItems, allTargetKeys)
	require.NoError(t, err)

	for _, tgt := range targets {
		baseDir, err := tgt.baseDir()
		require.NoError(t, err)

		for _, item := range allInstallables() {
			dest := item.pathFunc(baseDir)
			assert.True(t, FileExists(dest), "expected file to exist: %s", dest)

			content, err := os.ReadFile(dest)
			require.NoError(t, err)
			assert.Equal(t, item.content, content, "content mismatch for %s at %s", item.key, dest)
		}
	}
}

func TestInstallSelected_SingleSkillSingleTarget(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	err := installSelected([]string{"skill:paso"}, []string{"opencode"})
	require.NoError(t, err)

	opencodeBase := filepath.Join(tmpHome, ".config", "opencode")
	skillPath := filepath.Join(opencodeBase, "skills", "paso", "SKILL.md")
	assert.True(t, FileExists(skillPath), "expected skill file to exist")

	content, err := os.ReadFile(skillPath)
	require.NoError(t, err)
	assert.Equal(t, pasoSkillContent, content)

	// Commands should not be installed
	delegatePath := filepath.Join(opencodeBase, "commands", "paso-delegate.md")
	assert.False(t, FileExists(delegatePath), "delegate command should not be installed")

	// Claude target should not have anything
	claudeBase := filepath.Join(tmpHome, ".claude")
	assert.False(t, DirExists(claudeBase), "claude directory should not exist")
}

func TestInstallSelected_Idempotent(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	allItems := allKeys()
	allTargetKeys := []string{"opencode"}

	err := installSelected(allItems, allTargetKeys)
	require.NoError(t, err)

	// Install again — should succeed without error
	err = installSelected(allItems, allTargetKeys)
	require.NoError(t, err)

	opencodeBase := filepath.Join(tmpHome, ".config", "opencode")
	for _, item := range allInstallables() {
		dest := item.pathFunc(opencodeBase)
		content, err := os.ReadFile(dest)
		require.NoError(t, err)
		assert.Equal(t, item.content, content, "content should still match after reinstall for %s", item.key)
	}
}

func TestInstallAll(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	err := installAll()
	require.NoError(t, err)

	for _, tgt := range targets {
		baseDir, err := tgt.baseDir()
		require.NoError(t, err)

		for _, item := range allInstallables() {
			dest := item.pathFunc(baseDir)
			assert.True(t, FileExists(dest), "expected file to exist after installAll: %s", dest)

			content, err := os.ReadFile(dest)
			require.NoError(t, err)
			assert.Equal(t, item.content, content, "content mismatch for %s", item.key)
		}
	}
}

func TestRunCheck(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Install everything first
	err := installAll()
	require.NoError(t, err)

	// runCheck should succeed
	err = runCheck()
	require.NoError(t, err)
}

func TestRunCheck_NothingInstalled(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// runCheck should still succeed even with nothing installed
	err := runCheck()
	require.NoError(t, err)
}

func TestRunRemove(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Install everything
	err := installAll()
	require.NoError(t, err)

	// Verify files exist before removal
	for _, tgt := range targets {
		baseDir, err := tgt.baseDir()
		require.NoError(t, err)
		for _, item := range allInstallables() {
			dest := item.pathFunc(baseDir)
			assert.True(t, FileExists(dest), "expected file to exist before removal: %s", dest)
		}
	}

	// Remove everything
	err = runRemove()
	require.NoError(t, err)

	// Verify all files are gone
	for _, tgt := range targets {
		baseDir, err := tgt.baseDir()
		require.NoError(t, err)
		for _, item := range allInstallables() {
			dest := item.pathFunc(baseDir)
			assert.False(t, FileExists(dest), "expected file to be removed: %s", dest)
		}
	}
}

func TestRunRemove_NothingInstalled(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Remove with nothing installed should succeed
	err := runRemove()
	require.NoError(t, err)
}

func TestRunRemove_Idempotent(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	err := installAll()
	require.NoError(t, err)

	err = runRemove()
	require.NoError(t, err)

	// Removing again should still succeed
	err = runRemove()
	require.NoError(t, err)
}

func TestAtomicWriteFile(t *testing.T) {
	t.Run("writes file with correct content", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "test.txt")
		data := []byte("hello world")

		err := atomicWriteFile(path, data)
		require.NoError(t, err)

		content, err := os.ReadFile(path)
		require.NoError(t, err)
		assert.Equal(t, data, content)
	})

	t.Run("sets permissions to 0644", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "perms.txt")

		err := atomicWriteFile(path, []byte("content"))
		require.NoError(t, err)

		info, err := os.Stat(path)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0644), info.Mode().Perm())
	})

	t.Run("overwrites existing file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "overwrite.txt")

		err := atomicWriteFile(path, []byte("original"))
		require.NoError(t, err)

		err = atomicWriteFile(path, []byte("updated"))
		require.NoError(t, err)

		content, err := os.ReadFile(path)
		require.NoError(t, err)
		assert.Equal(t, []byte("updated"), content)
	})

	t.Run("handles empty content", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "empty.txt")

		err := atomicWriteFile(path, []byte{})
		require.NoError(t, err)

		content, err := os.ReadFile(path)
		require.NoError(t, err)
		assert.Empty(t, content)
	})

	t.Run("fails for nonexistent directory", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "no", "such", "dir", "file.txt")

		err := atomicWriteFile(path, []byte("data"))
		assert.Error(t, err)
	})

	t.Run("no leftover temp files on success", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "clean.txt")

		err := atomicWriteFile(path, []byte("data"))
		require.NoError(t, err)

		entries, err := os.ReadDir(dir)
		require.NoError(t, err)
		assert.Len(t, entries, 1, "should only have the target file, no leftover temp files")
		assert.Equal(t, "clean.txt", entries[0].Name())
	})
}

func TestDirExists(t *testing.T) {
	t.Run("returns true for existing directory", func(t *testing.T) {
		dir := t.TempDir()
		assert.True(t, DirExists(dir))
	})

	t.Run("returns false for nonexistent path", func(t *testing.T) {
		assert.False(t, DirExists("/nonexistent/path/that/does/not/exist"))
	})

	t.Run("returns false for a file", func(t *testing.T) {
		dir := t.TempDir()
		filePath := filepath.Join(dir, "file.txt")
		require.NoError(t, os.WriteFile(filePath, []byte("data"), 0644))

		assert.False(t, DirExists(filePath))
	})
}

func TestFileExists(t *testing.T) {
	t.Run("returns true for existing file", func(t *testing.T) {
		dir := t.TempDir()
		filePath := filepath.Join(dir, "file.txt")
		require.NoError(t, os.WriteFile(filePath, []byte("data"), 0644))

		assert.True(t, FileExists(filePath))
	})

	t.Run("returns false for nonexistent path", func(t *testing.T) {
		assert.False(t, FileExists("/nonexistent/file.txt"))
	})

	t.Run("returns false for a directory", func(t *testing.T) {
		dir := t.TempDir()
		assert.False(t, FileExists(dir))
	})
}

func TestEnsureDir(t *testing.T) {
	t.Run("creates directory that does not exist", func(t *testing.T) {
		base := t.TempDir()
		newDir := filepath.Join(base, "a", "b", "c")

		err := EnsureDir(newDir, 0755)
		require.NoError(t, err)
		assert.True(t, DirExists(newDir))
	})

	t.Run("succeeds for already existing directory", func(t *testing.T) {
		dir := t.TempDir()

		err := EnsureDir(dir, 0755)
		require.NoError(t, err)
		assert.True(t, DirExists(dir))
	})

	t.Run("creates with correct permissions", func(t *testing.T) {
		base := t.TempDir()
		newDir := filepath.Join(base, "permdir")

		err := EnsureDir(newDir, 0700)
		require.NoError(t, err)

		info, err := os.Stat(newDir)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0700), info.Mode().Perm())
	})
}

func TestAllInstallables(t *testing.T) {
	items := allInstallables()
	assert.Len(t, items, len(skills)+len(commands))

	keys := make(map[string]bool)
	for _, item := range items {
		keys[item.key] = true
	}
	assert.True(t, keys["skill:paso"])
	assert.True(t, keys["cmd:paso-delegate"])
	assert.True(t, keys["cmd:paso-setup"])
}

func TestAllKeys(t *testing.T) {
	keys := allKeys()
	assert.Len(t, keys, len(skills)+len(commands))
	assert.Contains(t, keys, "skill:paso")
	assert.Contains(t, keys, "cmd:paso-delegate")
	assert.Contains(t, keys, "cmd:paso-setup")
}

func TestResolveItems(t *testing.T) {
	t.Run("resolves subset of items", func(t *testing.T) {
		items := resolveItems([]string{"skill:paso"})
		require.Len(t, items, 1)
		assert.Equal(t, "skill:paso", items[0].key)
	})

	t.Run("resolves all items", func(t *testing.T) {
		items := resolveItems(allKeys())
		assert.Len(t, items, len(skills)+len(commands))
	})

	t.Run("returns empty for unknown keys", func(t *testing.T) {
		items := resolveItems([]string{"nonexistent:key"})
		assert.Empty(t, items)
	})

	t.Run("returns empty for empty input", func(t *testing.T) {
		items := resolveItems([]string{})
		assert.Empty(t, items)
	})
}

func TestResolveTargets(t *testing.T) {
	t.Run("resolves single target", func(t *testing.T) {
		resolved := resolveTargets([]string{"opencode"})
		require.Len(t, resolved, 1)
		assert.Equal(t, "opencode", resolved[0].key)
	})

	t.Run("resolves all targets", func(t *testing.T) {
		keys := make([]string, len(targets))
		for i, tgt := range targets {
			keys[i] = tgt.key
		}
		resolved := resolveTargets(keys)
		assert.Len(t, resolved, len(targets))
	})

	t.Run("returns empty for unknown keys", func(t *testing.T) {
		resolved := resolveTargets([]string{"nonexistent"})
		assert.Empty(t, resolved)
	})
}
