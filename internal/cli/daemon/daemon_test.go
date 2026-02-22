package daemon

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/cli"
)

func TestDaemonCmd_SubcommandRegistration(t *testing.T) {
	t.Parallel()

	cmd := DaemonCmd()

	assert.Equal(t, "daemon", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)

	subcommands := cmd.Commands()
	require.Len(t, subcommands, 4)

	names := make(map[string]bool, len(subcommands))
	for _, sub := range subcommands {
		names[sub.Use] = true
	}

	assert.True(t, names["start"], "missing subcommand: start")
	assert.True(t, names["setup"], "missing subcommand: setup")
	assert.True(t, names["status"], "missing subcommand: status")
	assert.True(t, names["stop"], "missing subcommand: stop")
}

func TestDaemonCmd_SubcommandMetadata(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "start"},
		{name: "setup"},
		{name: "status"},
		{name: "stop"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cmd := DaemonCmd()
			var found bool
			for _, sub := range cmd.Commands() {
				if sub.Use == tt.name {
					found = true
					assert.NotEmpty(t, sub.Short)
					assert.NotEmpty(t, sub.Long)
					assert.NotNil(t, sub.RunE)
					break
				}
			}
			require.True(t, found, "subcommand %q not found", tt.name)
		})
	}
}

func TestDaemonCmd_HasPersistentPreRunE(t *testing.T) {
	t.Parallel()

	cmd := DaemonCmd()
	assert.NotNil(t, cmd.PersistentPreRunE, "daemon command should have a PersistentPreRunE")
}

func TestCheckOS_Linux(t *testing.T) {
	t.Parallel()

	err := checkOS("linux")
	assert.NoError(t, err)
}

func TestCheckOS_UnsupportedSystems(t *testing.T) {
	t.Parallel()

	unsupported := []string{"darwin", "windows", "freebsd", "openbsd", "plan9"}

	for _, goos := range unsupported {
		t.Run(goos, func(t *testing.T) {
			t.Parallel()

			err := checkOS(goos)
			require.Error(t, err)

			var exitErr *cli.ExitErr
			require.True(t, errors.As(err, &exitErr))
			assert.Equal(t, cli.ExitError, exitErr.Code)
			assert.Contains(t, exitErr.Message, goos)
			assert.Contains(t, exitErr.Message, "paso daemon")
			assert.Contains(t, exitErr.Message, "Supported: linux")
		})
	}
}

func TestCheckOS_EmptyString(t *testing.T) {
	t.Parallel()

	err := checkOS("")
	require.Error(t, err)

	var exitErr *cli.ExitErr
	require.True(t, errors.As(err, &exitErr))
	assert.Equal(t, cli.ExitError, exitErr.Code)
}

func TestSetupCmd_RemoveFlag(t *testing.T) {
	t.Parallel()

	cmd := SetupCmd()

	flag := cmd.Flags().Lookup("remove")
	require.NotNil(t, flag, "--remove flag should exist")
	assert.Equal(t, "false", flag.DefValue)
}
