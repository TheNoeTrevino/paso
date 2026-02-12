package assignee

import (
	"encoding/json"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
	"github.com/thenoetrevino/paso/internal/config"
)

// SetCmd returns the assignee set subcommand
func SetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set the active assignee",
		Long: `Set the active assignee for the current session.

Examples:
  # Set active assignee
  paso assignee set -n "john"

  # JSON output
  paso assignee set -n "john" -j

  # Quiet mode
  paso assignee set -n "john" -q
`,
		RunE: runSet,
	}

	cmd.Flags().StringP("name", "n", "", "Assignee name (required)")
	if err := cmd.MarkFlagRequired("name"); err != nil {
		slog.Error("failed to marking flag as required", "error", err)
	}

	cmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	cmd.Flags().BoolP("quiet", "q", false, "Minimal output")

	return cmd
}

func runSet(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	name, _ := cmd.Flags().GetString("name")
	jsonOutput, _ := cmd.Flags().GetBool("json")
	quietMode, _ := cmd.Flags().GetBool("quiet")

	formatter := &cli.OutputFormatter{JSON: jsonOutput, Quiet: quietMode}

	cliInstance, err := cli.GetCLIFromContext(ctx)
	if err != nil {
		return formatter.Error(cli.ExitError, "INITIALIZATION_ERROR", err.Error())
	}
	defer func() {
		if err := cliInstance.Close(); err != nil {
			slog.Error("failed to closing CLI", "error", err)
		}
	}()

	cfg, err := config.Load()
	if err != nil {
		return formatter.Error(cli.ExitError, "CONFIG_LOAD_ERROR", err.Error())
	}

	if err := cfg.SetActiveAssignee(name); err != nil {
		return formatter.Error(cli.ExitError, "SET_ERROR", err.Error())
	}

	if quietMode {
		return nil
	}

	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(map[string]any{
			"success":         true,
			"active_assignee": name,
		})
	}

	cli.PrintSuccessf("Active assignee set to '%s'", name)
	return nil
}
