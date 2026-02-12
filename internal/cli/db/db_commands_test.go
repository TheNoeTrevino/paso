package db_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/cli/db"
	cli "github.com/thenoetrevino/paso/internal/testing/cli"
)

func isolateConfig(t *testing.T) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
}

// addLocalDB adds a database to the isolated config.
// It ensures config is isolated to prevent polluting the actual user config directory.
func addLocalDB(t *testing.T, name, path string) {
	t.Helper()
	if os.Getenv("XDG_CONFIG_HOME") == "" {
		isolateConfig(t)
	}
	cmd := db.AddCmd()
	cli.SetupCobraCommand(cmd, []string{path, "--name", name, "--local"})
	_, err := cli.ExecuteCommand(t, cmd)
	require.NoError(t, err)
}

func TestListCmd_Empty(t *testing.T) {
	isolateConfig(t)

	cmd := db.ListCmd()
	cli.SetupCobraCommand(cmd, []string{})
	output, err := cli.ExecuteCommand(t, cmd)

	require.NoError(t, err)
	assert.Contains(t, output, "No databases configured")
}

func TestListCmd_WithEntries(t *testing.T) {
	isolateConfig(t)

	tmpDB := filepath.Join(t.TempDir(), "test.db")
	addLocalDB(t, "mydb", tmpDB)

	cmd := db.ListCmd()
	cli.SetupCobraCommand(cmd, []string{})
	output, err := cli.ExecuteCommand(t, cmd)

	require.NoError(t, err)
	assert.Contains(t, output, "mydb")
}

func TestListCmd_JSON_Empty(t *testing.T) {
	isolateConfig(t)

	cmd := db.ListCmd()
	cli.SetupCobraCommand(cmd, []string{"--json"})
	output, err := cli.ExecuteCommand(t, cmd)

	require.NoError(t, err)
	var result map[string]any
	require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(output)), &result))

	databases, ok := result["databases"].([]any)
	require.True(t, ok)
	assert.Empty(t, databases)
}

func TestListCmd_JSON_WithEntries(t *testing.T) {
	isolateConfig(t)

	tmpDB := filepath.Join(t.TempDir(), "first.db")
	addLocalDB(t, "first", tmpDB)

	tmpDB2 := filepath.Join(t.TempDir(), "second.db")
	addLocalDB(t, "second", tmpDB2)

	cmd := db.ListCmd()
	cli.SetupCobraCommand(cmd, []string{"--json"})
	output, err := cli.ExecuteCommand(t, cmd)

	require.NoError(t, err)
	var result map[string]any
	require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(output)), &result))

	databases, ok := result["databases"].([]any)
	require.True(t, ok)
	assert.Len(t, databases, 2)

	names := make([]string, len(databases))
	for i, d := range databases {
		entry := d.(map[string]any)
		names[i] = entry["name"].(string)
		assert.Equal(t, "sqlite", entry["type"])
	}
	assert.Contains(t, names, "first")
	assert.Contains(t, names, "second")
}

func TestAddCmd_Local(t *testing.T) {
	isolateConfig(t)

	tmpDB := filepath.Join(t.TempDir(), "local.db")

	cmd := db.AddCmd()
	cli.SetupCobraCommand(cmd, []string{tmpDB, "--name", "localdb", "--local"})
	output, err := cli.ExecuteCommand(t, cmd)

	require.NoError(t, err)
	assert.Contains(t, output, "localdb")

	// Verify it appears in list
	listCmd := db.ListCmd()
	cli.SetupCobraCommand(listCmd, []string{"--json"})
	listOutput, err := cli.ExecuteCommand(t, listCmd)
	require.NoError(t, err)

	var result map[string]any
	require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(listOutput)), &result))
	databases := result["databases"].([]any)
	assert.Len(t, databases, 1)
	entry := databases[0].(map[string]any)
	assert.Equal(t, "localdb", entry["name"])
	assert.Equal(t, "sqlite", entry["type"])
}

func TestAddCmd_RemoteBadConnection(t *testing.T) {
	isolateConfig(t)

	cmd := db.AddCmd()
	cli.SetupCobraCommand(cmd, []string{
		"postgres://invalid:invalid@localhost:59999/nonexistent",
		"--name", "badremote",
		"--remote",
	})
	output, err := cli.ExecuteCommand(t, cmd)

	// The command succeeds — it saves the config even when the connection test fails.
	// The failure warning goes to stderr (not captured by ExecuteCommand).
	require.NoError(t, err)
	assert.Contains(t, output, "badremote")
	assert.Contains(t, output, "saved")

	// Verify the database was actually persisted
	listCmd := db.ListCmd()
	cli.SetupCobraCommand(listCmd, []string{"--json"})
	listOutput, err := cli.ExecuteCommand(t, listCmd)
	require.NoError(t, err)

	var result map[string]any
	require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(listOutput)), &result))
	databases := result["databases"].([]any)
	require.Len(t, databases, 1)
	entry := databases[0].(map[string]any)
	assert.Equal(t, "badremote", entry["name"])
	assert.Equal(t, "postgres", entry["type"])
}

func TestAddCmd_MissingNameFlag(t *testing.T) {
	isolateConfig(t)

	cmd := db.AddCmd()
	cli.SetupCobraCommand(cmd, []string{"some-connection-string", "--local"})
	_, err := cli.ExecuteCommand(t, cmd)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required flag")
}

func TestAddCmd_MissingTypeFlag(t *testing.T) {
	isolateConfig(t)

	cmd := db.AddCmd()
	cli.SetupCobraCommand(cmd, []string{"some-connection-string", "--name", "test"})
	_, err := cli.ExecuteCommand(t, cmd)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "either --remote or --local must be specified")
}

func TestAddCmd_BothTypeFlags(t *testing.T) {
	isolateConfig(t)

	cmd := db.AddCmd()
	cli.SetupCobraCommand(cmd, []string{"some-connection-string", "--name", "test", "--remote", "--local"})
	_, err := cli.ExecuteCommand(t, cmd)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot specify both --remote and --local")
}

func TestConnectCmd(t *testing.T) {
	isolateConfig(t)

	tmpDB := filepath.Join(t.TempDir(), "connect.db")
	addLocalDB(t, "mydb", tmpDB)

	cmd := db.ConnectCmd()
	cli.SetupCobraCommand(cmd, []string{"mydb"})
	output, err := cli.ExecuteCommand(t, cmd)

	require.NoError(t, err)
	assert.Contains(t, output, "mydb")

	// Verify it's now the active database via list --json
	listCmd := db.ListCmd()
	cli.SetupCobraCommand(listCmd, []string{"--json"})
	listOutput, err := cli.ExecuteCommand(t, listCmd)
	require.NoError(t, err)

	var result map[string]any
	require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(listOutput)), &result))
	databases := result["databases"].([]any)
	require.Len(t, databases, 1)
	entry := databases[0].(map[string]any)
	assert.True(t, entry["active"].(bool))
}

func TestConnectCmd_NonExistent(t *testing.T) {
	isolateConfig(t)

	cmd := db.ConnectCmd()
	cli.SetupCobraCommand(cmd, []string{"ghost"})
	_, err := cli.ExecuteCommand(t, cmd)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ghost")
	assert.Contains(t, err.Error(), "not found")
}

func TestRemoveCmd_Force(t *testing.T) {
	isolateConfig(t)

	tmpDB := filepath.Join(t.TempDir(), "remove.db")
	addLocalDB(t, "goner", tmpDB)

	cmd := db.RemoveCmd()
	cli.SetupCobraCommand(cmd, []string{"goner", "--force"})
	output, err := cli.ExecuteCommand(t, cmd)

	require.NoError(t, err)
	assert.Contains(t, output, "goner")
	assert.Contains(t, output, "removed")

	// Verify it's gone
	listCmd := db.ListCmd()
	cli.SetupCobraCommand(listCmd, []string{"--json"})
	listOutput, err := cli.ExecuteCommand(t, listCmd)
	require.NoError(t, err)

	var result map[string]any
	require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(listOutput)), &result))
	databases := result["databases"].([]any)
	assert.Empty(t, databases)
}

func TestRemoveCmd_Quiet(t *testing.T) {
	isolateConfig(t)

	tmpDB := filepath.Join(t.TempDir(), "remove-quiet.db")
	addLocalDB(t, "quietdb", tmpDB)

	cmd := db.RemoveCmd()
	cli.SetupCobraCommand(cmd, []string{"quietdb", "--quiet"})
	output, err := cli.ExecuteCommand(t, cmd)

	require.NoError(t, err)
	assert.Empty(t, strings.TrimSpace(output))

	// Verify it's gone
	listCmd := db.ListCmd()
	cli.SetupCobraCommand(listCmd, []string{"--json"})
	listOutput, err := cli.ExecuteCommand(t, listCmd)
	require.NoError(t, err)

	var result map[string]any
	require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(listOutput)), &result))
	databases := result["databases"].([]any)
	assert.Empty(t, databases)
}

func TestRemoveCmd_NonExistent(t *testing.T) {
	isolateConfig(t)

	cmd := db.RemoveCmd()
	cli.SetupCobraCommand(cmd, []string{"phantom", "--force"})
	_, err := cli.ExecuteCommand(t, cmd)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "phantom")
	assert.Contains(t, err.Error(), "not found")
}

func TestRemoveCmd_JSON(t *testing.T) {
	isolateConfig(t)

	tmpDB := filepath.Join(t.TempDir(), "json-remove.db")
	addLocalDB(t, "jsondb", tmpDB)

	cmd := db.RemoveCmd()
	cli.SetupCobraCommand(cmd, []string{"jsondb", "--force", "--json"})
	output, err := cli.ExecuteCommand(t, cmd)

	require.NoError(t, err)
	var result map[string]any
	require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(output)), &result))
	assert.True(t, result["success"].(bool))
	assert.Equal(t, "jsondb", result["name"])
}
