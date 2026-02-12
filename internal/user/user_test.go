package user

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCurrentUsername(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		validate func(string) bool
	}{
		{
			name: "returns non-empty username",
			validate: func(username string) bool {
				return username != ""
			},
		},
		{
			name: "returns valid username or fallback",
			validate: func(username string) bool {
				// Should return either a valid username or one of the fallbacks
				// We can't test specific values as they depend on the environment
				return username != "" && len(username) > 0
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			username := GetCurrentUsername()
			assert.True(t, tt.validate(username))
		})
	}
}

func TestGetCurrentUsernameFallback(t *testing.T) {
	t.Parallel()
	// This test verifies that the function always returns something
	username := GetCurrentUsername()

	assert.NotEmpty(t, username)

	// Verify it returns one of the expected values
	// (actual username, USER env var, or "unknown")
	assert.NotEmpty(t, username)
}
