package setup

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func readJSON(t *testing.T, path string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	var result map[string]any
	require.NoError(t, json.Unmarshal(data, &result))
	return result
}

func writeJSON(t *testing.T, path string, v any) {
	t.Helper()
	data, err := json.MarshalIndent(v, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, data, 0o644))
}

func chdir(t *testing.T, dir string) {
	t.Helper()
	orig, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(orig) })
}

func TestInstallClaude_Project(t *testing.T) {
	tmpDir := t.TempDir()
	chdir(t, tmpDir)

	require.NoError(t, os.MkdirAll(".claude", 0o755))

	err := InstallClaude(true)
	require.NoError(t, err)

	settingsPath := filepath.Join(tmpDir, ".claude", "settings.local.json")
	require.FileExists(t, settingsPath)

	settings := readJSON(t, settingsPath)
	hooks, ok := settings["hooks"].(map[string]any)
	require.True(t, ok, "expected hooks key in settings")

	for _, event := range []string{"SessionStart", "PreCompact"} {
		eventHooks, ok := hooks[event].([]any)
		require.True(t, ok, "expected %s key in hooks", event)
		require.Len(t, eventHooks, 1)

		hookMap := eventHooks[0].(map[string]any)
		commands := hookMap["hooks"].([]any)
		require.Len(t, commands, 1)
		assert.Equal(t, "paso tutorial", commands[0].(map[string]any)["command"])
	}
}

func TestInstallClaude_Global(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	err := InstallClaude(false)
	require.NoError(t, err)

	settingsPath := filepath.Join(tmpHome, ".claude", "settings.json")
	require.FileExists(t, settingsPath)

	settings := readJSON(t, settingsPath)
	hooks, ok := settings["hooks"].(map[string]any)
	require.True(t, ok)

	for _, event := range []string{"SessionStart", "PreCompact"} {
		eventHooks, ok := hooks[event].([]any)
		require.True(t, ok)
		require.Len(t, eventHooks, 1)

		hookMap := eventHooks[0].(map[string]any)
		commands := hookMap["hooks"].([]any)
		require.Len(t, commands, 1)
		assert.Equal(t, "paso tutorial", commands[0].(map[string]any)["command"])
	}
}

func TestInstallClaude_Idempotent(t *testing.T) {
	tmpDir := t.TempDir()
	chdir(t, tmpDir)
	require.NoError(t, os.MkdirAll(".claude", 0o755))

	require.NoError(t, InstallClaude(true))
	require.NoError(t, InstallClaude(true))

	settings := readJSON(t, filepath.Join(tmpDir, ".claude", "settings.local.json"))
	hooks := settings["hooks"].(map[string]any)
	for _, event := range []string{"SessionStart", "PreCompact"} {
		eventHooks := hooks[event].([]any)
		assert.Len(t, eventHooks, 1, "duplicate hooks created for %s", event)
	}
}

func TestInstallClaude_PreservesExistingSettings(t *testing.T) {
	tmpDir := t.TempDir()
	chdir(t, tmpDir)

	existing := map[string]any{
		"permissions": map[string]any{"allow": []any{"Read"}},
	}
	writeJSON(t, filepath.Join(tmpDir, ".claude", "settings.local.json"), existing)

	require.NoError(t, InstallClaude(true))

	settings := readJSON(t, filepath.Join(tmpDir, ".claude", "settings.local.json"))
	_, ok := settings["permissions"]
	assert.True(t, ok, "existing permissions key should be preserved")
	_, ok = settings["hooks"]
	assert.True(t, ok, "hooks key should be added")
}

func TestCheckClaude_Installed(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	require.NoError(t, InstallClaude(false))

	err := CheckClaude()
	assert.NoError(t, err)
}

func TestCheckClaude_NotInstalled(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	tmpProject := t.TempDir()
	chdir(t, tmpProject)

	err := CheckClaude()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no hooks installed")
}

func TestRemoveClaude_Project(t *testing.T) {
	tmpDir := t.TempDir()
	chdir(t, tmpDir)
	require.NoError(t, os.MkdirAll(".claude", 0o755))

	require.NoError(t, InstallClaude(true))

	settingsPath := filepath.Join(tmpDir, ".claude", "settings.local.json")
	assert.True(t, hasPasoHooks(settingsPath))

	require.NoError(t, RemoveClaude(true))

	assert.False(t, hasPasoHooks(settingsPath), "hooks should be removed")
}

func TestRemoveClaude_Global(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	require.NoError(t, InstallClaude(false))

	settingsPath := filepath.Join(tmpHome, ".claude", "settings.json")
	assert.True(t, hasPasoHooks(settingsPath))

	require.NoError(t, RemoveClaude(false))

	assert.False(t, hasPasoHooks(settingsPath))
}

func TestRemoveClaude_NoSettingsFile(t *testing.T) {
	tmpDir := t.TempDir()
	chdir(t, tmpDir)

	err := RemoveClaude(true)
	assert.NoError(t, err, "removing from nonexistent file should not error")
}

func TestInstallOpenCode_Project(t *testing.T) {
	tmpDir := t.TempDir()
	chdir(t, tmpDir)

	err := InstallOpenCode(true)
	require.NoError(t, err)

	configPath := filepath.Join(tmpDir, "opencode.json")
	require.FileExists(t, configPath)

	config := readJSON(t, configPath)
	plugins, ok := config["plugin"].([]any)
	require.True(t, ok, "expected plugin key in config")
	require.Len(t, plugins, 1)
	assert.Equal(t, "opencode-paso", plugins[0])
}

func TestInstallOpenCode_Global(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	err := InstallOpenCode(false)
	require.NoError(t, err)

	configPath := filepath.Join(tmpHome, ".config", "opencode", "opencode.json")
	require.FileExists(t, configPath)

	config := readJSON(t, configPath)
	plugins, ok := config["plugin"].([]any)
	require.True(t, ok)
	require.Len(t, plugins, 1)
	assert.Equal(t, "opencode-paso", plugins[0])
}

func TestInstallOpenCode_Idempotent(t *testing.T) {
	tmpDir := t.TempDir()
	chdir(t, tmpDir)

	require.NoError(t, InstallOpenCode(true))
	require.NoError(t, InstallOpenCode(true))

	config := readJSON(t, filepath.Join(tmpDir, "opencode.json"))
	plugins := config["plugin"].([]any)
	assert.Len(t, plugins, 1, "duplicate plugin entries should not be created")
}

func TestInstallOpenCode_PreservesExistingConfig(t *testing.T) {
	tmpDir := t.TempDir()
	chdir(t, tmpDir)

	existing := map[string]any{
		"theme": "dark",
	}
	writeJSON(t, filepath.Join(tmpDir, "opencode.json"), existing)

	require.NoError(t, InstallOpenCode(true))

	config := readJSON(t, filepath.Join(tmpDir, "opencode.json"))
	assert.Equal(t, "dark", config["theme"], "existing config keys should be preserved")
	plugins, ok := config["plugin"].([]any)
	require.True(t, ok)
	assert.Contains(t, plugins, "opencode-paso")
}

func TestCheckOpenCode_Installed(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	require.NoError(t, InstallOpenCode(false))

	err := CheckOpenCode()
	assert.NoError(t, err)
}

func TestCheckOpenCode_NotInstalled(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	tmpProject := t.TempDir()
	chdir(t, tmpProject)

	err := CheckOpenCode()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "plugin not installed")
}

func TestRemoveOpenCode_Project(t *testing.T) {
	tmpDir := t.TempDir()
	chdir(t, tmpDir)

	require.NoError(t, InstallOpenCode(true))

	configPath := filepath.Join(tmpDir, "opencode.json")
	assert.True(t, hasPasoPlugin(configPath))

	require.NoError(t, RemoveOpenCode(true))

	assert.False(t, hasPasoPlugin(configPath), "plugin should be removed")
}

func TestRemoveOpenCode_Global(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	require.NoError(t, InstallOpenCode(false))

	configPath := filepath.Join(tmpHome, ".config", "opencode", "opencode.json")
	assert.True(t, hasPasoPlugin(configPath))

	require.NoError(t, RemoveOpenCode(false))

	assert.False(t, hasPasoPlugin(configPath))
}

func TestRemoveOpenCode_NoConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	chdir(t, tmpDir)

	err := RemoveOpenCode(true)
	assert.NoError(t, err, "removing from nonexistent file should not error")
}

func TestRemoveOpenCode_PreservesOtherPlugins(t *testing.T) {
	tmpDir := t.TempDir()
	chdir(t, tmpDir)

	existing := map[string]any{
		"plugin": []any{"other-plugin", "opencode-paso", "another-plugin"},
	}
	writeJSON(t, filepath.Join(tmpDir, "opencode.json"), existing)

	require.NoError(t, RemoveOpenCode(true))

	config := readJSON(t, filepath.Join(tmpDir, "opencode.json"))
	plugins := config["plugin"].([]any)
	assert.Len(t, plugins, 2)
	assert.Contains(t, plugins, "other-plugin")
	assert.Contains(t, plugins, "another-plugin")
	assert.NotContains(t, plugins, "opencode-paso")
}

func TestRemoveClaude_PreservesOtherHooks(t *testing.T) {
	tmpDir := t.TempDir()
	chdir(t, tmpDir)

	settings := map[string]any{
		"hooks": map[string]any{
			"SessionStart": []any{
				map[string]any{
					"matcher": "",
					"hooks": []any{
						map[string]any{
							"type":    "command",
							"command": "paso tutorial",
						},
					},
				},
				map[string]any{
					"matcher": "",
					"hooks": []any{
						map[string]any{
							"type":    "command",
							"command": "other-tool init",
						},
					},
				},
			},
		},
	}
	writeJSON(t, filepath.Join(tmpDir, ".claude", "settings.local.json"), settings)

	require.NoError(t, RemoveClaude(true))

	result := readJSON(t, filepath.Join(tmpDir, ".claude", "settings.local.json"))
	hooks := result["hooks"].(map[string]any)
	sessionHooks := hooks["SessionStart"].([]any)
	assert.Len(t, sessionHooks, 1, "should only remove paso hook, keeping other-tool")

	remaining := sessionHooks[0].(map[string]any)
	cmds := remaining["hooks"].([]any)
	assert.Equal(t, "other-tool init", cmds[0].(map[string]any)["command"])
}

func TestHasPasoHooks(t *testing.T) {
	tmpDir := t.TempDir()

	noHooksPath := filepath.Join(tmpDir, "no-hooks.json")
	writeJSON(t, noHooksPath, map[string]any{"hooks": map[string]any{}})
	assert.False(t, hasPasoHooks(noHooksPath))

	assert.False(t, hasPasoHooks(filepath.Join(tmpDir, "nonexistent.json")))

	withHooksPath := filepath.Join(tmpDir, "with-hooks.json")
	writeJSON(t, withHooksPath, map[string]any{
		"hooks": map[string]any{
			"SessionStart": []any{
				map[string]any{
					"matcher": "",
					"hooks": []any{
						map[string]any{"type": "command", "command": "paso tutorial"},
					},
				},
			},
		},
	})
	assert.True(t, hasPasoHooks(withHooksPath))
}

func TestHasPasoPlugin(t *testing.T) {
	tmpDir := t.TempDir()

	noPluginPath := filepath.Join(tmpDir, "no-plugin.json")
	writeJSON(t, noPluginPath, map[string]any{"plugin": []any{}})
	assert.False(t, hasPasoPlugin(noPluginPath))

	assert.False(t, hasPasoPlugin(filepath.Join(tmpDir, "nonexistent.json")))

	withPluginPath := filepath.Join(tmpDir, "with-plugin.json")
	writeJSON(t, withPluginPath, map[string]any{"plugin": []any{"opencode-paso"}})
	assert.True(t, hasPasoPlugin(withPluginPath))
}

func TestAtomicWriteFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.json")

	content := []byte(`{"key": "value"}`)
	require.NoError(t, atomicWriteFile(path, content))

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, content, data)

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
}

func TestEnsureDir(t *testing.T) {
	tmpDir := t.TempDir()
	newDir := filepath.Join(tmpDir, "a", "b", "c")

	require.NoError(t, EnsureDir(newDir, 0o755))
	assert.True(t, DirExists(newDir))

	require.NoError(t, EnsureDir(newDir, 0o755), "calling on existing dir should not error")
}

func TestDirExists(t *testing.T) {
	tmpDir := t.TempDir()
	assert.True(t, DirExists(tmpDir))
	assert.False(t, DirExists(filepath.Join(tmpDir, "nonexistent")))

	f := filepath.Join(tmpDir, "file.txt")
	require.NoError(t, os.WriteFile(f, []byte("hi"), 0o644))
	assert.False(t, DirExists(f), "regular file should not count as dir")
}

func TestFileExists(t *testing.T) {
	tmpDir := t.TempDir()

	f := filepath.Join(tmpDir, "file.txt")
	require.NoError(t, os.WriteFile(f, []byte("hi"), 0o644))
	assert.True(t, FileExists(f))

	assert.False(t, FileExists(filepath.Join(tmpDir, "nonexistent")))
	assert.False(t, FileExists(tmpDir), "directory should not count as file")
}
