package assignee

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
	"github.com/thenoetrevino/paso/internal/cli/styles"
)

// DeleteCmd returns the assignee delete subcommand
func DeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete an assignee",
		Long: `Delete an assignee by ID (requires confirmation unless --force/-f or --quiet/-q).
Tasks assigned to this assignee will have their assignee set to NULL.

Examples:
  # Delete with confirmation
  paso assignee delete -i 1

  # Skip confirmation
  paso assignee delete -i 1 -f

  # Quiet mode (no confirmation)
  paso assignee delete -i 1 -q
`,
		RunE: runDelete,
	}

	cmd.Flags().IntP("id", "i", 0, "Assignee ID (required)")
	if err := cmd.MarkFlagRequired("id"); err != nil {
		slog.Error("failed to marking flag as required", "error", err)
	}

	cmd.Flags().BoolP("force", "f", false, "Skip confirmation")
	cmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	cmd.Flags().BoolP("quiet", "q", false, "Minimal output")

	return cmd
}

func runDelete(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	assigneeID, _ := cmd.Flags().GetInt("id")
	force, _ := cmd.Flags().GetBool("force")
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

	assignee, err := cliInstance.App.AssigneeService.GetByID(ctx, assigneeID)
	if err != nil {
		assignee = nil
	}

	if !force && !quietMode {
		var assigneeName string
		if assignee != nil {
			assigneeName = assignee.Name
		} else {
			assigneeName = "unknown"
		}
		fmt.Printf("Delete assignee #%d (%s)? Tasks will be unassigned. (y/N): ", assigneeID, assigneeName)
		var response string
		_, err := fmt.Scanln(&response)
		if err != nil {
			slog.Error("failed to reading user input", "error", err)
			fmt.Println("Cancelled (failed to read input)")
			return nil
		}
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Println("Cancelled")
			return nil
		}
	}

	if err := cliInstance.App.AssigneeService.Delete(ctx, assigneeID); err != nil {
		return formatter.Error(cli.ExitError, "DELETE_ERROR", err.Error())
	}

	if quietMode {
		return nil
	}

	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(map[string]any{
			"success":     true,
			"assignee_id": assigneeID,
		})
	}

	colorScheme := cli.GetColorScheme()
	fmt.Print(styles.RenderSuccess(fmt.Sprintf("Assignee %d deleted successfully", assigneeID), colorScheme))
	return nil
}
