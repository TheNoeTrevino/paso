package assignee_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/cli/assignee"
	cli "github.com/thenoetrevino/paso/internal/testing/cli"
	"github.com/thenoetrevino/paso/internal/user"
)

func TestSetAssignee(t *testing.T) {
	_, app := cli.SetupCLITest(t)

	t.Run("set assignee default output", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", tmpDir)

		cmd := assignee.SetCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{"--name", "alice"})
		require.NoError(t, err)
		assert.Contains(t, output, "alice")
	})

	t.Run("set assignee quiet output", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", tmpDir)

		cmd := assignee.SetCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{"--name", "bob", "--quiet"})
		require.NoError(t, err)
		assert.Empty(t, strings.TrimSpace(output))
	})

	t.Run("set assignee json output", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", tmpDir)

		cmd := assignee.SetCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{"--name", "charlie", "--json"})
		require.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal([]byte(output), &result)
		require.NoError(t, err)
		assert.True(t, result["success"].(bool))
		assert.Equal(t, "charlie", result["active_assignee"])
	})
}

func TestSetAssignee_Errors(t *testing.T) {
	_, app := cli.SetupCLITest(t)

	t.Run("missing name flag", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", tmpDir)

		cmd := assignee.SetCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "required flag")
	})
}

func TestWhoAmI(t *testing.T) {
	_, app := cli.SetupCLITest(t)

	t.Run("whoami with default system username", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", tmpDir)

		expectedUsername := user.GetCurrentUsername()

		cmd := assignee.WhoAmICmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{})
		require.NoError(t, err)
		assert.Contains(t, output, expectedUsername)
	})

	t.Run("whoami after set", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", tmpDir)

		// First set an assignee
		setCmd := assignee.SetCmd()
		_, err := cli.ExecuteCLICommand(t, app, setCmd, []string{"--name", "dave"})
		require.NoError(t, err)

		// Now whoami should return the set assignee
		whoamiCmd := assignee.WhoAmICmd()
		output, err := cli.ExecuteCLICommand(t, app, whoamiCmd, []string{})
		require.NoError(t, err)
		assert.Contains(t, output, "dave")
	})

	t.Run("whoami json output", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", tmpDir)

		// Set assignee first so we get a predictable value
		setCmd := assignee.SetCmd()
		_, err := cli.ExecuteCLICommand(t, app, setCmd, []string{"--name", "eve"})
		require.NoError(t, err)

		whoamiCmd := assignee.WhoAmICmd()
		output, err := cli.ExecuteCLICommand(t, app, whoamiCmd, []string{"--json"})
		require.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal([]byte(output), &result)
		require.NoError(t, err)
		assert.True(t, result["success"].(bool))
		assert.Equal(t, "eve", result["active_assignee"])
	})

	t.Run("whoami quiet output", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", tmpDir)

		// Set assignee first so we get a predictable value
		setCmd := assignee.SetCmd()
		_, err := cli.ExecuteCLICommand(t, app, setCmd, []string{"--name", "frank"})
		require.NoError(t, err)

		whoamiCmd := assignee.WhoAmICmd()
		output, err := cli.ExecuteCLICommand(t, app, whoamiCmd, []string{"--quiet"})
		require.NoError(t, err)
		assert.Equal(t, "frank", strings.TrimSpace(output))
	})

	t.Run("whoami json with default username", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", tmpDir)

		expectedUsername := user.GetCurrentUsername()

		whoamiCmd := assignee.WhoAmICmd()
		output, err := cli.ExecuteCLICommand(t, app, whoamiCmd, []string{"--json"})
		require.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal([]byte(output), &result)
		require.NoError(t, err)
		assert.True(t, result["success"].(bool))
		assert.Equal(t, expectedUsername, result["active_assignee"])
	})

	t.Run("whoami quiet with default username", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", tmpDir)

		expectedUsername := user.GetCurrentUsername()

		whoamiCmd := assignee.WhoAmICmd()
		output, err := cli.ExecuteCLICommand(t, app, whoamiCmd, []string{"--quiet"})
		require.NoError(t, err)
		assert.Equal(t, expectedUsername, strings.TrimSpace(output))
	})
}
