package assignee

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
	"github.com/thenoetrevino/paso/internal/config"
	"github.com/thenoetrevino/paso/internal/user"
)

// WhoAmICmd returns the assignee whoami subcommand
func WhoAmICmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "whoami",
		Short: "Show the active assignee",
		Long: `Show the active assignee for the current session.
Falls back to the system username if no active assignee is set.

Examples:
  # Show active assignee
  paso assignee whoami

  # JSON output
  paso assignee whoami -j

  # Quiet mode (name only)
  paso assignee whoami -q
`,
		RunE: runWhoAmI,
	}

	cmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	cmd.Flags().BoolP("quiet", "q", false, "Minimal output (name only)")

	return cmd
}

func runWhoAmI(cmd *cobra.Command, args []string) error {
	jsonOutput, _ := cmd.Flags().GetBool("json")
	quietMode, _ := cmd.Flags().GetBool("quiet")

	formatter := &cli.OutputFormatter{JSON: jsonOutput, Quiet: quietMode}

	cfg, err := config.Load()
	if err != nil {
		if fmtErr := formatter.Error("CONFIG_LOAD_ERROR", err.Error()); fmtErr != nil {
			slog.Error("failed to formatting error message", "error", fmtErr)
		}
		return err
	}

	activeAssignee := cfg.GetActiveAssignee()
	if activeAssignee == "" {
		activeAssignee = user.GetCurrentUsername()
	}

	if quietMode {
		fmt.Println(activeAssignee)
		return nil
	}

	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(map[string]any{
			"success":         true,
			"active_assignee": activeAssignee,
		})
	}

	fmt.Printf("Active assignee: %s\n", activeAssignee)
	return nil
}
