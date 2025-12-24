package label

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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
  # Human-readable list
  paso label list --project=1

  # JSON output for agents
  paso label list --project=1 --json

  # Quiet mode (one ID per line)
  paso label list --project=1 --quiet
`,
		RunE: runList,
	}

	// Required flags
	cmd.Flags().Int("project", 0, "Project ID (required)")
	if err := cmd.MarkFlagRequired("project"); err != nil {
		log.Printf("Error marking flag as required: %v", err)
	}

	// Agent-friendly flags
	cmd.Flags().Bool("json", false, "Output in JSON format")
	cmd.Flags().Bool("quiet", false, "Minimal output (IDs only)")

	return cmd
}

func runList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	labelProject, _ := cmd.Flags().GetInt("project")
	jsonOutput, _ := cmd.Flags().GetBool("json")
	quietMode, _ := cmd.Flags().GetBool("quiet")

	formatter := &cli.OutputFormatter{JSON: jsonOutput, Quiet: quietMode}

	// Initialize CLI
	cliInstance, err := cli.NewCLI(ctx)
	if err != nil {
		if fmtErr := formatter.Error("INITIALIZATION_ERROR", err.Error()); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		return err
	}
	defer func() {
		if err := cliInstance.Close(); err != nil {
			log.Printf("Error closing CLI: %v", err)
		}
	}()

	// Validate project exists
	project, err := cliInstance.Repo.GetProjectByID(ctx, labelProject)
	if err != nil {
		if fmtErr := formatter.Error("PROJECT_NOT_FOUND", fmt.Sprintf("project %d not found", labelProject)); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		os.Exit(cli.ExitNotFound)
	}

	// Get labels
	labels, err := cliInstance.Repo.GetLabelsByProject(ctx, labelProject)
	if err != nil {
		if fmtErr := formatter.Error("LABEL_FETCH_ERROR", err.Error()); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
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
		labelList := make([]map[string]interface{}, len(labels))
		for i, lbl := range labels {
			labelList[i] = map[string]interface{}{
				"id":         lbl.ID,
				"name":       lbl.Name,
				"color":      lbl.Color,
				"project_id": lbl.ProjectID,
			}
		}
		return json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
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
