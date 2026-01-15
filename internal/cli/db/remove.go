package db

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/config"
)

func RemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove a saved database connection",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			if cfg.GetDatabase(name) == nil {
				return fmt.Errorf("database '%s' not found", name)
			}

			if err := cfg.RemoveDatabase(name); err != nil {
				return err
			}

			fmt.Printf("Database '%s' removed\n", name)
			return nil
		},
	}

	return cmd
}
