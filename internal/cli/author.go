package cli

import (
	"github.com/thenoetrevino/paso/internal/config"
	"github.com/thenoetrevino/paso/internal/user"
)

// ResolveDefaultAuthor returns the active assignee from config,
// falling back to the current OS username.
func ResolveDefaultAuthor() string {
	cfg, err := config.Load()
	if err == nil {
		if name := cfg.GetActiveAssignee(); name != "" {
			return name
		}
	}
	return user.GetCurrentUsername()
}
