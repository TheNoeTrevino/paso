package jira

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
	"github.com/thenoetrevino/paso/internal/cli/handler"
	"github.com/thenoetrevino/paso/internal/jira"
	"github.com/thenoetrevino/paso/internal/models"
	taskservice "github.com/thenoetrevino/paso/internal/services/task"
	"github.com/thenoetrevino/paso/internal/spinner"
)

// ImportCmd returns the jira import subcommand.
func ImportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import <issue-key>",
		Short: "Import a Jira issue as a paso task",
		Long: `Import a Jira issue as a paso task, including its title, description, and comments.

Requires the jira CLI (ankitpokhrel/jira-cli) to be installed and authenticated.

Examples:
  # Import issue PROJ-123 into the current project
  paso jira import PROJ-123

  # Import into a specific project
  paso jira import PROJ-123 -p 1

  # Import into a specific column
  paso jira import PROJ-123 -p 1 -c "In Progress"

  # JSON output for agents
  paso jira import PROJ-123 -j

  # Quiet mode for bash capture
  TASK_ID=$(paso jira import PROJ-123 -q)
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
		return nil, fmt.Errorf("issue key is required")
	}

	issueKey, err := ValidateIssueKey(args.Args[0])
	if err != nil {
		return nil, err
	}

	if err := jira.CheckInstalled(); err != nil {
		return nil, err
	}

	if !args.GetBool("json") && !args.GetBool("quiet") {
		sp := spinner.New(fmt.Sprintf("Importing Jira issue %s...", issueKey))
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

	issue, err := cliInstance.App.JiraFetcher.FetchIssue(ctx, issueKey)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Jira issue: %w", err)
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

	defaultAuthor := cli.ResolveDefaultAuthor()

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
			slog.Warn("failed to import comment", "issue", issueKey, "error", err)
			continue
		}
		commentsImported++
	}

	return &importResult{
		TaskID:        task.ID,
		Title:         task.Title,
		Description:   task.Description,
		Project:       project.Name,
		IssueKey:      issueKey,
		CommentsCount: commentsImported,
	}, nil
}

func parseImportFlags(_ *cobra.Command) error {
	return nil
}
