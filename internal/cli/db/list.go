package db

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/config"
)

type databaseOutput struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Active bool   `json:"active"`
}

func ListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List configured databases",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			jsonOutput, _ := cmd.Flags().GetBool("json")

			if len(cfg.Databases) == 0 {
				if jsonOutput {
					return json.NewEncoder(os.Stdout).Encode(map[string]any{
						"databases": []databaseOutput{},
					})
				}
				fmt.Println("No databases configured")
				fmt.Println("Use 'paso db connect' to add a database")
				return nil
			}

			if jsonOutput {
				databases := make([]databaseOutput, 0, len(cfg.Databases))
				for _, db := range cfg.Databases {
					databases = append(databases, databaseOutput{
						Name:   db.Name,
						Type:   db.Type,
						Active: db.Name == cfg.ActiveDatabase,
					})
				}
				return json.NewEncoder(os.Stdout).Encode(map[string]any{
					"databases": databases,
				})
			}

			fmt.Printf("%-15s %-12s %s\n", "NAME", "TYPE", "ACTIVE")

			for _, db := range cfg.Databases {
				active := ""
				if db.Name == cfg.ActiveDatabase {
					active = "*"
				}
				fmt.Printf("%-15s %-12s %s\n", db.Name, db.Type, active)
			}

			return nil
		},
	}

	cmd.Flags().Bool("json", false, "Output in JSON format")

	return cmd
}
