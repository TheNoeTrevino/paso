package setup

import (
	"github.com/spf13/cobra"
)

// SetupCmd returns the setup command
func SetupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Setup integrations with AI tools",
		Long:  `Configure paso to work with AI coding assistants like Claude Code.`,
	}

	cmd.AddCommand(ClaudeCmd())

	return cmd
}
