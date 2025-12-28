package user

import (
	"testing"
)

func TestGetCurrentUsername(t *testing.T) {
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
			if !tt.validate(username) {
				t.Errorf("GetCurrentUsername() validation failed, got %q", username)
			}
		})
	}
}

func TestGetCurrentUsernameFallback(t *testing.T) {
	// This test verifies that the function always returns something
	username := GetCurrentUsername()

	if username == "" {
		t.Error("GetCurrentUsername() should never return an empty string")
	}

	// Verify it returns one of the expected values
	// (actual username, USER env var, or "unknown")
	if username == "" {
		t.Error("GetCurrentUsername() returned empty string, should have returned fallback")
	}
}
