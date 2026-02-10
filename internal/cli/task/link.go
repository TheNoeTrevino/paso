package task

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
	"github.com/thenoetrevino/paso/internal/cli/styles"
)

// LinkCmd returns the task link subcommand
func LinkCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "link",
		Short: "Link tasks with relationships",
		Long: `Create a relationship between two tasks.

Relationship Types:
  (default)    Parent-Child: Non-blocking hierarchical relationship
  --blocker/-b Blocked By/Blocker: Blocking relationship (parent blocked by child)
  --related/-R Related To: Non-blocking associative relationship

The --blocker/-b and --related/-R flags are mutually exclusive. If neither is specified,
a parent-child relationship is created.

Examples:
  # Parent-child relationship (shorthand)
  paso task link -P 5 -C 3

  # Blocking relationship (task 5 blocked by task 3)
  paso task link -P 5 -C 3 -b

  # Related relationship
  paso task link -P 5 -C 3 -R

  # Long-form flags also supported
  paso task link --parent=5 --child=3 --blocker
`,
		RunE: runLink,
	}

	// Required flags
	cmd.Flags().IntP("parent", "P", 0, "Parent task ID (required)")
	if err := cmd.MarkFlagRequired("parent"); err != nil {
		slog.Error("failed to mark flag as required", "error", err)
	}

	cmd.Flags().IntP("child", "C", 0, "Child task ID (required)")
	if err := cmd.MarkFlagRequired("child"); err != nil {
		slog.Error("failed to mark flag as required", "error", err)
	}

	// Agent-friendly flags
	cmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	cmd.Flags().BoolP("quiet", "q", false, "Minimal output")

	// Relationship type flags (mutually exclusive)
	cmd.Flags().BoolP("blocker", "b", false, "Create blocking relationship (Blocked By/Blocker)")
	cmd.Flags().BoolP("related", "R", false, "Create related relationship (Related To)")

	return cmd
}

func runLink(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	parentID, _ := cmd.Flags().GetInt("parent")
	childID, _ := cmd.Flags().GetInt("child")
	blocker, _ := cmd.Flags().GetBool("blocker")
	related, _ := cmd.Flags().GetBool("related")
	jsonOutput, _ := cmd.Flags().GetBool("json")
	quietMode, _ := cmd.Flags().GetBool("quiet")

	formatter := &cli.OutputFormatter{JSON: jsonOutput, Quiet: quietMode}

	// Validate mutually exclusive flags
	if blocker && related {
		if fmtErr := formatter.Error("INVALID_FLAGS",
			"cannot specify both --blocker and --related flags"); fmtErr != nil {
			slog.Error("failed to format error message", "error", fmtErr)
		}
		os.Exit(cli.ExitUsage)
	}

	// Determine relation type ID
	relationTypeID := 1 // Default: Parent/Child
	relationTypeName := "parent-child"

	if blocker {
		relationTypeID = 2 // Blocked By/Blocker
		relationTypeName = "blocking"
	} else if related {
		relationTypeID = 3 // Related To
		relationTypeName = "related"
	}

	// Initialize CLI
	cliInstance, err := cli.GetCLIFromContext(ctx)
	if err != nil {
		if fmtErr := formatter.Error("INITIALIZATION_ERROR", err.Error()); fmtErr != nil {
			slog.Error("failed to format error message", "error", fmtErr)
		}
		os.Exit(cli.ExitError)
	}
	defer func() {
		if err := cliInstance.Close(); err != nil {
			slog.Error("failed to close CLI", "error", err)
		}
	}()

	// Create the relationship with specific type
	if err := cliInstance.App.TaskService.AddChildRelation(ctx, parentID, childID, relationTypeID); err != nil {
		if fmtErr := formatter.Error("LINK_ERROR", err.Error()); fmtErr != nil {
			slog.Error("failed to format error message", "error", fmtErr)
		}
		os.Exit(cli.ExitError)
	}

	// Output success
	if quietMode {
		return nil
	}

	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(map[string]any{
			"success":          true,
			"parent_id":        parentID,
			"child_id":         childID,
			"relation_type_id": relationTypeID,
			"relation_type":    relationTypeName,
		})
	}

	// Human-readable output with relationship type
	colors := cli.GetColorScheme()
	var message string
	switch relationTypeID {
	case 2:
		message = fmt.Sprintf("Created blocking relationship: task %d is blocked by task %d", parentID, childID)
	case 3:
		message = fmt.Sprintf("Created related relationship between task %d and task %d", parentID, childID)
	default:
		message = fmt.Sprintf("Linked task %d as child of task %d", childID, parentID)
	}
	fmt.Print(styles.RenderSuccess(message, colors))
	return nil
}
