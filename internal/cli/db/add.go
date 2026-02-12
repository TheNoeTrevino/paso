package db

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
	"github.com/thenoetrevino/paso/internal/config"
	"github.com/thenoetrevino/paso/internal/database"
)

func AddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <connection-string>",
		Short: "Add a database connection",
		Long: `Add a new database connection to paso.

The connection string is provided as a positional argument.
Use -r/--remote for PostgreSQL or -l/--local for SQLite.

If a connection with the same name already exists, it will be updated.

Examples:
  # Add a remote PostgreSQL database
  paso db add "postgresql://user:pass@host:5432/dbname" -n cloud -r

  # Add a local SQLite database
  paso db add ~/.paso/tasks.db -n local -l
`,
		Args: cobra.ExactArgs(1),
		RunE: runAdd,
	}

	cmd.Flags().StringP("name", "n", "", "Name for this connection (required)")
	if err := cmd.MarkFlagRequired("name"); err != nil {
		fmt.Fprintf(os.Stderr, "failed to mark flag as required: %v\n", err)
	}

	cmd.Flags().BoolP("remote", "r", false, "Treat as PostgreSQL connection")
	cmd.Flags().BoolP("local", "l", false, "Treat as SQLite connection")

	cmd.Flags().Bool("json", false, "Output in JSON format")
	cmd.Flags().Bool("quiet", false, "Minimal output")

	return cmd
}

func runAdd(cmd *cobra.Command, args []string) error {
	connStr := args[0]
	name, _ := cmd.Flags().GetString("name")
	remote, _ := cmd.Flags().GetBool("remote")
	local, _ := cmd.Flags().GetBool("local")
	jsonOutput, _ := cmd.Flags().GetBool("json")
	quiet, _ := cmd.Flags().GetBool("quiet")

	if remote && local {
		return fmt.Errorf("cannot specify both --remote and --local")
	}
	if !remote && !local {
		return fmt.Errorf("either --remote or --local must be specified")
	}

	var dbType string
	if remote {
		dbType = "postgres"
	} else {
		dbType = "sqlite"
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := cfg.AddDatabase(name, connStr, dbType); err != nil {
		return fmt.Errorf("failed to add database: %w", err)
	}

	pingErr := database.TestConnection(connStr, database.DatabaseType(dbType))
	connected := pingErr == nil

	if quiet {
		fmt.Println(name)
		return nil
	}

	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(map[string]any{
			"success":   true,
			"name":      name,
			"type":      dbType,
			"connected": connected,
			"message":   formatAddMessage(name, connected),
		})
	}

	if connected {
		cli.PrintSuccessf("Database '%s' added successfully", name)
	} else {
		cli.PrintSuccessf("Database '%s' saved", name)
		fmt.Fprintf(os.Stderr, "⚠ Connection test failed: %v\n", pingErr)
	}

	return nil
}

func formatAddMessage(name string, connected bool) string {
	if connected {
		return fmt.Sprintf("Database '%s' added successfully", name)
	}
	return fmt.Sprintf("Database '%s' saved (connection test failed)", name)
}
