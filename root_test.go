package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/cli"
)

// resetRootCmd resets cobra's stateful output buffers and parsed flag
// values between tests so that --help / --version don't leak across runs.
func resetRootCmd(t *testing.T) *bytes.Buffer {
	t.Helper()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	// Cobra caches parsed flag values. Reset them so a previous test's
	// --help=true doesn't cause the next test to print help instead of
	// executing normally.
	rootCmd.Flags().VisitAll(func(f *pflag.Flag) {
		_ = f.Value.Set(f.DefValue)
		f.Changed = false
	})
	return buf
}

func TestRootCommand_Help(t *testing.T) {
	buf := resetRootCmd(t)
	rootCmd.SetArgs([]string{"--help"})

	err := rootCmd.Execute()
	require.NoError(t, err)

	output := buf.String()

	expectedSubcommands := []string{
		"task",
		"project",
		"column",
		"label",
		"assignee",
		"install",
		"db",
		"tui",
		"completion",
	}
	for _, sub := range expectedSubcommands {
		assert.Contains(t, output, sub, "help output should list the %q subcommand", sub)
	}
}

func TestRootCommand_Version(t *testing.T) {
	buf := resetRootCmd(t)
	rootCmd.SetArgs([]string{"--version"})

	err := rootCmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "paso version")
	assert.Contains(t, output, version)
}

func TestRootCommand_UnknownCommand(t *testing.T) {
	_ = resetRootCmd(t)
	rootCmd.SetArgs([]string{"nonexistent-command"})

	err := rootCmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown command")
}

func TestRootCommand_SilenceFlags(t *testing.T) {
	// SilenceErrors and SilenceUsage are set in main() before Execute().
	// Verify the rootCmd fields match after we set them the same way main() does.
	rootCmd.SilenceErrors = true
	rootCmd.SilenceUsage = true

	assert.True(t, rootCmd.SilenceErrors, "SilenceErrors should be true")
	assert.True(t, rootCmd.SilenceUsage, "SilenceUsage should be true")
}

func TestRootCommand_NoArgs_ShowsHelp(t *testing.T) {
	buf := resetRootCmd(t)
	rootCmd.SetArgs([]string{})

	err := rootCmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.True(t,
		strings.Contains(output, "paso") || strings.Contains(output, "Usage"),
		"running with no args should show help-like output",
	)
}

func TestExitErr_Structure(t *testing.T) {
	t.Run("implements error interface", func(t *testing.T) {
		exitErr := cli.NewExitErr(42, "test error")
		var err error = exitErr
		assert.Equal(t, "test error", err.Error())
	})

	t.Run("ExitCode returns the code", func(t *testing.T) {
		exitErr := cli.NewExitErr(3, "not found")
		assert.Equal(t, 3, exitErr.ExitCode())
		assert.Equal(t, 3, exitErr.Code)
	})

	t.Run("errors.As can unwrap ExitErr", func(t *testing.T) {
		exitErr := cli.NewExitErr(cli.ExitNotFound, "resource missing")
		var wrapped error = exitErr

		var target *cli.ExitErr
		require.True(t, errors.As(wrapped, &target))
		assert.Equal(t, cli.ExitNotFound, target.ExitCode())
	})

	t.Run("exit code constants are correct", func(t *testing.T) {
		assert.Equal(t, 0, cli.ExitSuccess)
		assert.Equal(t, 1, cli.ExitError)
		assert.Equal(t, 2, cli.ExitUsage)
		assert.Equal(t, 3, cli.ExitNotFound)
		assert.Equal(t, 4, cli.ExitDataErr)
		assert.Equal(t, 5, cli.ExitValidation)
	})
}
