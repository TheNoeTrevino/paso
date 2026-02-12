package db

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
	"github.com/thenoetrevino/paso/internal/config"
)

func RemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove a saved database connection",
		Long: `Remove a saved database connection by name.

Requires confirmation unless --force or --quiet is specified.

Examples:
  # Remove with confirmation prompt
  paso db remove cloud

  # Remove without confirmation
  paso db remove cloud --force
`,
		Args: cobra.ExactArgs(1),
		RunE: runRemove,
	}

	cmd.Flags().Bool("force", false, "Skip confirmation")
	cmd.Flags().Bool("json", false, "Output in JSON format")
	cmd.Flags().Bool("quiet", false, "Minimal output")

	return cmd
}

func runRemove(cmd *cobra.Command, args []string) error {
	name := args[0]
	force, _ := cmd.Flags().GetBool("force")
	jsonOutput, _ := cmd.Flags().GetBool("json")
	quiet, _ := cmd.Flags().GetBool("quiet")

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.GetDatabase(name) == nil {
		return fmt.Errorf("database '%s' not found", name)
	}

	if !force && !quiet {
		fmt.Printf("Delete connection '%s'? (y/N): ", name)
		var response string
		if _, err := fmt.Scanln(&response); err != nil {
			fmt.Println("Cancelled (failed to read input)")
			return nil
		}
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Println("Cancelled")
			return nil
		}
	}

	if err := cfg.RemoveDatabase(name); err != nil {
		return fmt.Errorf("failed to remove database: %w", err)
	}

	if quiet {
		return nil
	}

	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(map[string]any{
			"success": true,
			"name":    name,
			"message": fmt.Sprintf("Database '%s' removed", name),
		})
	}

	cli.PrintSuccessf("Database '%s' removed", name)
	return nil
}
