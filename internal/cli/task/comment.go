package task

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
	taskservice "github.com/thenoetrevino/paso/internal/services/task"
	userutil "github.com/thenoetrevino/paso/internal/user"
)

// CommentCmd returns the task comment subcommand
func CommentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "comment",
		Short: "Add a comment to a task",
		Long: `Add a comment to a task.

Comments are limited to 1000 characters and are displayed in the task detail view.

Examples:
  # Add a comment to task #42
  paso task comment --id=42 --message="Need to follow up with team"

  # Add a longer comment
  paso task comment --id=42 --message="Blocked by API changes in PR #123"

  # JSON output for agents
  paso task comment --id=42 --message="Investigation complete" --json

  # Quiet mode for bash capture
  COMMENT_ID=$(paso task comment --id=42 --message="Fixed" --quiet)
`,
		RunE: runComment,
	}

	// Required flags
	cmd.Flags().Int("id", 0, "Task ID (required)")
	if err := cmd.MarkFlagRequired("id"); err != nil {
		log.Printf("Error marking flag as required: %v", err)
	}

	cmd.Flags().String("message", "", "Comment message (required, max 1000 chars)")
	if err := cmd.MarkFlagRequired("message"); err != nil {
		log.Printf("Error marking flag as required: %v", err)
	}

	cmd.Flags().String("author", "", "Comment author (defaults to current user)")

	// Agent-friendly flags
	cmd.Flags().Bool("json", false, "Output in JSON format")
	cmd.Flags().Bool("quiet", false, "Minimal output (comment ID only)")

	return cmd
}

func runComment(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	taskID, _ := cmd.Flags().GetInt("id")
	message, _ := cmd.Flags().GetString("message")
	author, _ := cmd.Flags().GetString("author")
	jsonOutput, _ := cmd.Flags().GetBool("json")
	quietMode, _ := cmd.Flags().GetBool("quiet")

	// Default author to current user if not provided
	if author == "" {
		author = userutil.GetCurrentUsername()
	}

	formatter := &cli.OutputFormatter{JSON: jsonOutput, Quiet: quietMode}

	// Validate message length before initializing CLI
	if len(message) > 1000 {
		if fmtErr := formatter.Error("MESSAGE_TOO_LONG",
			"message exceeds 1000 character limit"); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		os.Exit(cli.ExitValidation)
	}

	// Initialize CLI
	cliInstance, err := cli.GetCLIFromContext(ctx)
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

	// Validate task exists
	taskDetail, err := cliInstance.App.TaskService.GetTaskDetail(ctx, taskID)
	if err != nil {
		if fmtErr := formatter.Error("TASK_FETCH_ERROR", err.Error()); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		return err
	}

	// Create comment
	comment, err := cliInstance.App.TaskService.CreateComment(ctx, taskservice.CreateCommentRequest{
		TaskID:  taskID,
		Message: message,
		Author:  author,
	})
	if err != nil {
		if fmtErr := formatter.Error("COMMENT_CREATE_ERROR", err.Error()); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		return err
	}

	// Output based on mode (JSON/Quiet/Human)
	if quietMode {
		fmt.Printf("%d\n", comment.ID)
		return nil
	}

	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
			"success": true,
			"comment": map[string]interface{}{
				"id":         comment.ID,
				"task_id":    comment.TaskID,
				"message":    comment.Message,
				"author":     comment.Author,
				"created_at": comment.CreatedAt,
			},
			"task": map[string]interface{}{
				"id":            taskDetail.ID,
				"title":         taskDetail.Title,
				"ticket_number": taskDetail.TicketNumber,
				"project":       taskDetail.ProjectName,
			},
		})
	}

	// Human-readable output
	fmt.Printf("✓ Comment added to task #%d (%s)\n", taskDetail.TicketNumber, taskDetail.Title)
	fmt.Printf("  Project: %s\n", taskDetail.ProjectName)
	fmt.Printf("  Message: %s\n", message)
	fmt.Printf("  Comment ID: %d\n", comment.ID)

	return nil
}
