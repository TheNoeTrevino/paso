package gh

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
	"github.com/thenoetrevino/paso/internal/cli/handler"
	"github.com/thenoetrevino/paso/internal/config"
	"github.com/thenoetrevino/paso/internal/github"
	"github.com/thenoetrevino/paso/internal/models"
	taskservice "github.com/thenoetrevino/paso/internal/services/task"
	"github.com/thenoetrevino/paso/internal/spinner"
	"github.com/thenoetrevino/paso/internal/user"
)

// ImportCmd returns the gh import subcommand.
func ImportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import <issue-number>",
		Short: "Import a GitHub issue as a paso task",
		Long: `Import a GitHub issue as a paso task, including its title, body, and comments.

Requires the GitHub CLI (gh) to be installed and authenticated.

Examples:
  # Import issue #101 into the current project
  paso gh import 101

  # Import into a specific project
  paso gh import 101 -p 1

  # Import into a specific column
  paso gh import 101 -p 1 -c "In Progress"

  # JSON output for agents
  paso gh import 101 -j

  # Quiet mode for bash capture
  TASK_ID=$(paso gh import 101 -q)
`,
		Args: cobra.ExactArgs(1),
		RunE: handler.Command(&importHandler{}, parseImportFlags),
	}

	cmd.Flags().IntP("project", "p", 0, "Project ID (uses git branch association if not specified)")
	cmd.Flags().StringP("column", "c", "", "Column name (defaults to first column)")

	cmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	cmd.Flags().BoolP("quiet", "q", false, "Minimal output (task ID only)")

	return cmd
}

type importHandler struct{}

func (h *importHandler) Execute(ctx context.Context, args *handler.Arguments) (any, error) {
	if len(args.Args) == 0 {
		return nil, fmt.Errorf("issue number is required")
	}

	issueNumber, err := ValidateIssueNumber(args.Args[0])
	if err != nil {
		return nil, err
	}

	if err := github.CheckInstalled(); err != nil {
		return nil, err
	}

	if !args.GetBool("json") && !args.GetBool("quiet") {
		sp := spinner.New(fmt.Sprintf("Importing GitHub issue #%d...", issueNumber))
		sp.Start()
		defer sp.Stop()
	}

	columnName := args.GetString("column", "")

	cliInstance, err := cli.GetCLIFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize CLI: %w", err)
	}
	defer func() {
		if err := cliInstance.Close(); err != nil {
			slog.Error("failed to close CLI", "error", err)
		}
	}()

	cmd := args.GetCmd()
	projectID, err := cli.GetProjectIDWithCLI(cmd, cliInstance)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: no project specified: use --project flag or create a project associated with this branch")
	}

	project, err := cliInstance.App.ProjectService.GetProjectByID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to find project: project %d not found", projectID)
	}

	columns, err := cliInstance.App.ColumnService.GetColumnsByProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch columns: %w", err)
	}
	if len(columns) == 0 {
		return nil, fmt.Errorf("failed to import issue: project has no columns")
	}

	var targetColumnID int
	if columnName == "" {
		targetColumnID = columns[0].ID
	} else {
		col, err := cli.FindColumnByName(columns, columnName)
		if err != nil {
			return nil, fmt.Errorf("failed to find column: column '%s' not found", columnName)
		}
		targetColumnID = col.ID
	}

	issue, err := cliInstance.App.GitHubFetcher.FetchIssue(ctx, issueNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch GitHub issue: %w", err)
	}

	task, err := cliInstance.App.TaskService.CreateTask(ctx, taskservice.CreateTaskRequest{
		Title:       issue.Title,
		Description: issue.Body,
		ColumnID:    targetColumnID,
		Position:    models.DefaultTaskPosition,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	defaultAuthor := resolveDefaultAuthor()

	commentsImported := 0
	for _, comment := range issue.Comments {
		author := comment.Author
		if author == "" {
			author = defaultAuthor
		}
		_, err := cliInstance.App.TaskService.CreateComment(ctx, taskservice.CreateCommentRequest{
			TaskID:  task.ID,
			Message: comment.Body,
			Author:  author,
		})
		if err != nil {
			slog.Warn("failed to import comment", "issue", issueNumber, "error", err)
			continue
		}
		commentsImported++
	}

	return &importResult{
		TaskID:        task.ID,
		Title:         task.Title,
		Description:   task.Description,
		Project:       project.Name,
		IssueNumber:   issueNumber,
		CommentsCount: commentsImported,
	}, nil
}

func resolveDefaultAuthor() string {
	cfg, err := config.Load()
	if err == nil {
		if name := cfg.GetActiveAssignee(); name != "" {
			return name
		}
	}
	return user.GetCurrentUsername()
}

func parseImportFlags(_ *cobra.Command) error {
	return nil
}
