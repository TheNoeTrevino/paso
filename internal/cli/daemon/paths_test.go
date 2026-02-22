package daemon

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetPasoDir_WithHomeEnv(t *testing.T) {
	tests := []struct {
		name     string
		home     string
		expected string
	}{
		{
			name:     "standard home directory",
			home:     "/home/testuser",
			expected: "/home/testuser/.paso",
		},
		{
			name:     "tmp fake home directory",
			home:     "/tmp/fakehome",
			expected: "/tmp/fakehome/.paso",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("HOME", tt.home)

			result, err := getPasoDir()

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetPasoDir_FallsBackToUserHomeDir(t *testing.T) {
	t.Setenv("HOME", "")

	// os.UserHomeDir() may fail when HOME is unset depending on the OS;
	// skip in that case since the fallback path is platform-dependent.
	if _, err := os.UserHomeDir(); err != nil {
		t.Skip("os.UserHomeDir() unavailable without HOME set on this platform")
	}

	result, err := getPasoDir()

	require.NoError(t, err)
	assert.True(t, strings.HasSuffix(result, "/.paso"), "expected path to end with /.paso, got: %s", result)
	assert.True(t, strings.HasPrefix(result, "/"), "expected absolute path starting with /, got: %s", result)
}

func TestGetSocketPath_ReturnsCorrectPath(t *testing.T) {
	t.Setenv("HOME", "/home/testuser")

	result, err := getSocketPath()

	require.NoError(t, err)
	assert.Equal(t, "/home/testuser/.paso/paso.sock", result)
}

func TestGetSystemdUserDir_WithXDGConfigHome(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/custom/config")

	result, err := getSystemdUserDir()

	require.NoError(t, err)
	assert.Equal(t, "/custom/config/systemd/user", result)
}

func TestGetSystemdUserDir_DefaultPath(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")

	result, err := getSystemdUserDir()

	require.NoError(t, err)
	assert.True(t, strings.HasSuffix(result, "/.config/systemd/user"), "expected path to end with /.config/systemd/user, got: %s", result)
}
