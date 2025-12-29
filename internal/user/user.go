package user

import (
	"os"
	"os/user"
)

// GetCurrentUsername returns the current system username.
// It tries multiple methods with fallbacks:
// 1. user.Current() - most reliable, gets username from OS
// 2. USER environment variable - fallback for restricted environments
// 3. "unknown" - final fallback to ensure a non-empty value
func GetCurrentUsername() string {
	// Try to get current user from OS
	currentUser, err := user.Current()
	if err != nil {
		// Fallback to USER environment variable
		username := os.Getenv("USER")
		if username == "" {
			// Final fallback
			return "unknown"
		}
		return username
	}
	return currentUser.Username
}
