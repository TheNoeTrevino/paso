package task

import (
	"encoding/json"
	"fmt"
	"log/slog"
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
  # Add a comment using shorthand flags
  paso task comment -i 42 -m "Need to follow up with team"

  # With author
  paso task comment -i 42 -m "Blocked by API changes in PR #123" -a "noe"

  # JSON output for agents
  paso task comment -i 42 -m "Investigation complete" -j

  # Quiet mode for bash capture
  COMMENT_ID=$(paso task comment -i 42 -m "Fixed" -q)

  # Long-form flags also supported
  paso task comment --id=42 --message="Fixed" --quiet
`,
		RunE: runComment,
	}

	// Required flags
	cmd.Flags().IntP("id", "i", 0, "Task ID (required)")
	if err := cmd.MarkFlagRequired("id"); err != nil {
		slog.Error("failed to mark flag as required", "error", err)
	}

	cmd.Flags().StringP("message", "m", "", "Comment message (required, max 1000 chars)")
	if err := cmd.MarkFlagRequired("message"); err != nil {
		slog.Error("failed to mark flag as required", "error", err)
	}

	cmd.Flags().StringP("author", "a", "", "Comment author (defaults to current user)")

	// Agent-friendly flags
	cmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	cmd.Flags().BoolP("quiet", "q", false, "Minimal output (comment ID only)")

	return cmd
}

func runComment(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

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
	if err := ValidateCommentMessage(message); err != nil {
		return formatter.Error(cli.ExitValidation, "MESSAGE_TOO_LONG", err.Error())
	}

	// Initialize CLI
	cliInstance, err := cli.GetCLIFromContext(ctx)
	if err != nil {
		return formatter.Error(cli.ExitError, "INITIALIZATION_ERROR", err.Error())
	}
	defer func() {
		if err := cliInstance.Close(); err != nil {
			slog.Error("failed to close CLI", "error", err)
		}
	}()

	// Validate task exists
	taskDetail, err := cliInstance.App.TaskService.GetTaskDetail(ctx, taskID)
	if err != nil {
		return formatter.Error(cli.ExitError, "TASK_FETCH_ERROR", err.Error())
	}

	// Create comment
	comment, err := cliInstance.App.TaskService.CreateComment(ctx, taskservice.CreateCommentRequest{
		TaskID:  taskID,
		Message: message,
		Author:  author,
	})
	if err != nil {
		return formatter.Error(cli.ExitError, "COMMENT_CREATE_ERROR", err.Error())
	}

	// Build result
	result := &CommentResult{
		CommentID:   comment.ID,
		TaskID:      comment.TaskID,
		Message:     comment.Message,
		Author:      comment.Author,
		CreatedAt:   comment.CreatedAt.Format("2006-01-02 15:04:05"),
		TaskTitle:   taskDetail.Title,
		TaskNumber:  taskDetail.TaskNumber,
		ProjectName: taskDetail.ProjectName,
	}

	// Output based on mode (JSON/Quiet/Human)
	if quietMode {
		fmt.Print(FormatCommentQuiet(result.CommentID))
		return nil
	}

	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(FormatCommentJSON(result))
	}

	// Human-readable output
	fmt.Print(FormatCommentHuman(result, cli.GetColorScheme()))
	return nil
}
