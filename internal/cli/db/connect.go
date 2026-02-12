package db

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
	"github.com/thenoetrevino/paso/internal/cli/styles"
	"github.com/thenoetrevino/paso/internal/config"
	"github.com/thenoetrevino/paso/internal/database"
)

func ConnectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "connect <name>",
		Short: "Switch to a saved database connection",
		Long: `Switch the active database to a previously saved connection and verify connectivity.

Examples:
  # Connect to a saved database
  paso db connect cloud

  # Connect with JSON output
  paso db connect local --json
`,
		Args: cobra.ExactArgs(1),
		RunE: runConnect,
	}

	cmd.Flags().Bool("json", false, "Output in JSON format")
	cmd.Flags().Bool("quiet", false, "Minimal output")

	return cmd
}

func runConnect(cmd *cobra.Command, args []string) error {
	name := args[0]
	jsonOutput, _ := cmd.Flags().GetBool("json")
	quiet, _ := cmd.Flags().GetBool("quiet")

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	db := cfg.GetDatabase(name)
	if db == nil {
		return fmt.Errorf("database '%s' not found. Use 'paso db list' to see available databases", name)
	}

	if err := cfg.SetActiveDatabase(name); err != nil {
		return fmt.Errorf("failed to switch database: %w", err)
	}

	pingErr := database.TestConnection(db.ConnectionString, database.DatabaseType(db.Type))
	connected := pingErr == nil

	if quiet {
		fmt.Println(name)
		return nil
	}

	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(map[string]any{
			"success":   true,
			"name":      name,
			"type":      db.Type,
			"connected": connected,
			"message":   formatConnectMessage(name, connected),
		})
	}

	colorScheme := cli.GetColorScheme()
	if connected {
		fmt.Print(styles.RenderSuccess(fmt.Sprintf("Connected to '%s'", name), colorScheme))
	} else {
		fmt.Print(styles.RenderSuccess(fmt.Sprintf("Switched to '%s'", name), colorScheme))
		fmt.Fprintf(os.Stderr, "⚠ Connection test failed: %v\n", pingErr)
	}

	return nil
}

func formatConnectMessage(name string, connected bool) string {
	if connected {
		return fmt.Sprintf("Connected to '%s'", name)
	}
	return fmt.Sprintf("Switched to '%s' (connection test failed)", name)
}
