package db

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/config"
)

func ConnectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "connect",
		Short: "Add a database connection",
		Long: `Add a new database connection to paso.

Examples:
  # Add a remote PostgreSQL database
  paso db connect --remote "postgresql://user:pass@host:5432/dbname" --name cloud

  # Add a local SQLite database
  paso db connect --local ~/.paso/tasks.db --name local
`,
		RunE: runConnect,
	}

	cmd.Flags().String("remote", "", "Remote PostgreSQL connection URL")
	cmd.Flags().String("local", "", "Local SQLite database path")
	cmd.Flags().String("name", "", "Name for this connection (required)")
	if err := cmd.MarkFlagRequired("name"); err != nil {
		fmt.Fprintf(os.Stderr, "failed to mark flag as required: %v\n", err)
	}

	cmd.Flags().Bool("json", false, "Output in JSON format")
	cmd.Flags().Bool("quiet", false, "Minimal output")

	return cmd
}

func runConnect(cmd *cobra.Command, args []string) error {
	remote, _ := cmd.Flags().GetString("remote")
	local, _ := cmd.Flags().GetString("local")
	name, _ := cmd.Flags().GetString("name")
	jsonOutput, _ := cmd.Flags().GetBool("json")
	quiet, _ := cmd.Flags().GetBool("quiet")

	if remote == "" && local == "" {
		return fmt.Errorf("either --remote or --local must be provided")
	}
	if remote != "" && local != "" {
		return fmt.Errorf("cannot specify both --remote and --local")
	}

	var connStr, dbType string
	if remote != "" {
		connStr = remote
		dbType = "postgres"
	} else {
		connStr = local
		dbType = "sqlite"
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := cfg.AddDatabase(name, connStr, dbType); err != nil {
		return fmt.Errorf("failed to add database: %w", err)
	}

	if quiet {
		fmt.Println(name)
		return nil
	}

	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(map[string]any{
			"success": true,
			"name":    name,
			"type":    dbType,
			"message": fmt.Sprintf("Database '%s' added successfully", name),
		})
	}

	fmt.Printf("Database '%s' added successfully\n", name)
	return nil
}
