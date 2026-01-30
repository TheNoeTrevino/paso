package label

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
)

// ListCmd returns the label list subcommand
func ListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List labels in a project",
		Long: `List all labels in a project.

Examples:
  # Human-readable list (shorthand)
  paso label list -p 1

  # JSON output for agents
  paso label list -p 1 -j

  # Quiet mode (one ID per line)
  paso label list -p 1 -q

  # Long-form flags also supported
  paso label list --project=1 --json
`,
		RunE: runList,
	}

	// Flags
	cmd.Flags().IntP("project", "p", 0, "Project ID (uses git branch association if not specified)")

	// Agent-friendly flags
	cmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	cmd.Flags().BoolP("quiet", "q", false, "Minimal output (IDs only)")

	return cmd
}

func runList(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	jsonOutput, _ := cmd.Flags().GetBool("json")
	quietMode, _ := cmd.Flags().GetBool("quiet")

	formatter := &cli.OutputFormatter{JSON: jsonOutput, Quiet: quietMode}

	// Initialize CLI first
	cliInstance, err := cli.GetCLIFromContext(ctx)
	if err != nil {
		if fmtErr := formatter.Error("INITIALIZATION_ERROR", err.Error()); fmtErr != nil {
			slog.Error("failed to formatting error message", "error", fmtErr)
		}
		return err
	}
	defer func() {
		if err := cliInstance.Close(); err != nil {
			slog.Error("failed to closing CLI", "error", err)
		}
	}()

	// Get project ID from flag or git branch
	labelProject, err := cli.GetProjectIDWithCLI(cmd, cliInstance)
	if err != nil {
		if fmtErr := formatter.ErrorWithSuggestion("NO_PROJECT",
			err.Error(),
			"Use --project flag or create a project associated with this git branch"); fmtErr != nil {
			slog.Error("failed to formatting error message", "error", fmtErr)
		}
		os.Exit(cli.ExitUsage)
	}

	// Validate project exists
	project, err := cliInstance.App.ProjectService.GetProjectByID(ctx, labelProject)
	if err != nil {
		if fmtErr := formatter.Error("PROJECT_NOT_FOUND", fmt.Sprintf("project %d not found", labelProject)); fmtErr != nil {
			slog.Error("failed to formatting error message", "error", fmtErr)
		}
		os.Exit(cli.ExitNotFound)
	}

	// Get labels
	labels, err := cliInstance.App.LabelService.GetLabelsByProject(ctx, labelProject)
	if err != nil {
		if fmtErr := formatter.Error("LABEL_FETCH_ERROR", err.Error()); fmtErr != nil {
			slog.Error("failed to formatting error message", "error", fmtErr)
		}
		return err
	}

	// Output based on mode
	if quietMode {
		for _, lbl := range labels {
			fmt.Printf("%d\n", lbl.ID)
		}
		return nil
	}

	if jsonOutput {
		labelList := make([]map[string]any, len(labels))
		for i, lbl := range labels {
			labelList[i] = map[string]any{
				"id":         lbl.ID,
				"name":       lbl.Name,
				"color":      lbl.Color,
				"project_id": lbl.ProjectID,
			}
		}
		return json.NewEncoder(os.Stdout).Encode(map[string]any{
			"success": true,
			"labels":  labelList,
		})
	}

	// Human-readable output
	if len(labels) == 0 {
		fmt.Printf("No labels found in project '%s'\n", project.Name)
		return nil
	}

	fmt.Printf("Labels in project '%s':\n", project.Name)
	fmt.Printf("  %-4s %-20s %s\n", "ID", "Name", "Color")
	fmt.Println("  " + strings.Repeat("-", 50))
	for _, lbl := range labels {
		fmt.Printf("  %-4d %-20s %s\n", lbl.ID, lbl.Name, lbl.Color)
	}
	return nil
}
