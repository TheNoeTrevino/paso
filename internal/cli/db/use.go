package db

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/config"
)

func UseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "use <name>",
		Short: "Switch active database",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			if cfg.GetDatabase(name) == nil {
				return fmt.Errorf("database '%s' not found. Use 'paso db list' to see available databases", name)
			}

			if err := cfg.SetActiveDatabase(name); err != nil {
				return err
			}

			fmt.Printf("Switched to database '%s'\n", name)
			return nil
		},
	}

	return cmd
}
