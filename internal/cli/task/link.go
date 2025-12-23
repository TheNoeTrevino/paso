package task

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
)

// LinkCmd returns the task link subcommand
func LinkCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "link",
		Short: "Link tasks with relationships",
		Long: `Create a relationship between two tasks.

Relationship Types:
  (default)  Parent-Child: Non-blocking hierarchical relationship
  --blocker  Blocked By/Blocker: Blocking relationship (parent blocked by child)
  --related  Related To: Non-blocking associative relationship

The --blocker and --related flags are mutually exclusive. If neither is specified,
a parent-child relationship is created.

Examples:
  # Parent-child relationship (default)
  paso task link --parent=5 --child=3

  # Blocking relationship (task 5 blocked by task 3)
  paso task link --parent=5 --child=3 --blocker

  # Related relationship
  paso task link --parent=5 --child=3 --related
`,
		RunE: runLink,
	}

	// Required flags
	cmd.Flags().Int("parent", 0, "Parent task ID (required)")
	if err := cmd.MarkFlagRequired("parent"); err != nil {
		log.Printf("Error marking flag as required: %v", err)
	}

	cmd.Flags().Int("child", 0, "Child task ID (required)")
	if err := cmd.MarkFlagRequired("child"); err != nil {
		log.Printf("Error marking flag as required: %v", err)
	}

	// Agent-friendly flags
	cmd.Flags().Bool("json", false, "Output in JSON format")
	cmd.Flags().Bool("quiet", false, "Minimal output")

	// Relationship type flags (mutually exclusive)
	cmd.Flags().Bool("blocker", false, "Create blocking relationship (Blocked By/Blocker)")
	cmd.Flags().Bool("related", false, "Create related relationship (Related To)")

	return cmd
}

func runLink(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

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
			log.Printf("Error formatting error message: %v", fmtErr)
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

	// Create the relationship with specific type
	if err := cliInstance.Repo.AddSubtaskWithRelationType(ctx, parentID, childID, relationTypeID); err != nil {
		if fmtErr := formatter.Error("LINK_ERROR", err.Error()); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		return err
	}

	// Output success
	if quietMode {
		return nil
	}

	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
			"success":          true,
			"parent_id":        parentID,
			"child_id":         childID,
			"relation_type_id": relationTypeID,
			"relation_type":    relationTypeName,
		})
	}

	// Human-readable output with relationship type
	switch relationTypeID {
	case 2:
		fmt.Printf("✓ Created blocking relationship: task %d is blocked by task %d\n", parentID, childID)
	case 3:
		fmt.Printf("✓ Created related relationship between task %d and task %d\n", parentID, childID)
	default:
		fmt.Printf("✓ Linked task %d as child of task %d\n", childID, parentID)
	}

	return nil
}
