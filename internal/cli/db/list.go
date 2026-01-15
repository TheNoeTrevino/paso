package db

import (
	"encoding/json"
	"fmt"
	"os"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
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

			baseStyle := lipgloss.NewStyle().Padding(0, 1)
			headerStyle := baseStyle.Bold(true).Foreground(lipgloss.Color("252"))
			activeStyle := baseStyle.Foreground(lipgloss.Color("#01BE85")).Background(lipgloss.Color("#00432F"))

			var rows [][]string
			for _, db := range cfg.Databases {
				active := ""
				if db.Name == cfg.ActiveDatabase {
					active = "✓"
				}
				rows = append(rows, []string{db.Name, db.Type, active})
			}

			t := table.New().
				Border(lipgloss.RoundedBorder()).
				BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("238"))).
				Headers("NAME", "TYPE", "ACTIVE").
				Rows(rows...).
				StyleFunc(func(row, col int) lipgloss.Style {
					if row == table.HeaderRow {
						return headerStyle
					}

					isActive := rows[row][2] == "✓"
					if isActive {
						return activeStyle
					}

					even := row%2 == 0

					if col == 1 {
						dbType := rows[row][col]
						var typeColor string
						switch dbType {
						case "postgres":
							if even {
								typeColor = "#439F8E"
							} else {
								typeColor = "#00E2C7"
							}
						case "sqlite":
							if even {
								typeColor = "#59B980"
							} else {
								typeColor = "#75FBAB"
							}
						}
						if typeColor != "" {
							return baseStyle.Foreground(lipgloss.Color(typeColor))
						}
					}

					if even {
						return baseStyle.Foreground(lipgloss.Color("245"))
					}
					return baseStyle.Foreground(lipgloss.Color("252"))
				})

			fmt.Println(t)

			return nil
		},
	}

	cmd.Flags().Bool("json", false, "Output in JSON format")

	return cmd
}
