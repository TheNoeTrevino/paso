package task

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
	"github.com/thenoetrevino/paso/internal/cli/styles"
	"github.com/thenoetrevino/paso/internal/config"
	"github.com/thenoetrevino/paso/internal/config/colors"
	"github.com/thenoetrevino/paso/internal/models"
)

// ShowCmd returns the task show subcommand
func ShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show [id]",
		Short: "Show task details",
		Long: `Display all details of a task including description, relationships, labels, and metadata.

Examples:
  paso task show 42
  paso task show -i 42
  paso task show -i 42 -j
  paso task show --id=42 --json
`,
		Args: cobra.MaximumNArgs(1),
		RunE: runShow,
	}

	// Flags
	cmd.Flags().IntP("id", "i", 0, "Task ID (can also be provided as positional argument)")
	cmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	cmd.Flags().BoolP("quiet", "q", false, "Minimal output (ID only)")

	return cmd
}

func runShow(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	// Parse task ID from positional arg or flag
	var taskID int
	if len(args) > 0 {
		if _, err := fmt.Sscanf(args[0], "%d", &taskID); err != nil {
			taskID = 0 // Invalid input, will be caught by validation below
		}
	} else {
		taskID, _ = cmd.Flags().GetInt("id")
	}

	jsonOutput, _ := cmd.Flags().GetBool("json")
	quietMode, _ := cmd.Flags().GetBool("quiet")

	formatter := &cli.OutputFormatter{JSON: jsonOutput, Quiet: quietMode}

	// Validate task ID
	if taskID <= 0 {
		return formatter.ErrorWithSuggestion(cli.ExitUsage, "INVALID_TASK_ID",
			"task ID must be a positive integer",
			"Usage: paso task show <id> or paso task show --id=<id>")
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

	// Get task details
	task, err := cliInstance.App.TaskService.GetTaskDetail(ctx, taskID)
	if err != nil {
		return formatter.Error(cli.ExitNotFound, "TASK_NOT_FOUND", fmt.Sprintf("task %d not found", taskID))
	}

	// Output in appropriate format
	if quietMode {
		fmt.Printf("%d\n", task.ID)
		return nil
	}

	if jsonOutput {
		return outputJSON(task)
	}

	// Load config for color scheme
	cfg, err := config.Load()
	if err != nil {
		// Fallback to default colors if config fails to load
		cfg = &config.Config{
			ColorScheme: config.DefaultColorScheme(),
		}
	}

	// Human-readable output with lipgloss
	return outputHuman(task, cfg.ColorScheme)
}

func outputJSON(task *models.TaskDetail) error {
	// Convert comments to a serializable format
	comments := make([]map[string]any, 0, len(task.Comments))
	for _, c := range task.Comments {
		comments = append(comments, map[string]any{
			"id":         c.ID,
			"content":    c.Message,
			"author":     c.Author,
			"created_at": c.CreatedAt,
			"updated_at": c.UpdatedAt,
		})
	}

	return json.NewEncoder(os.Stdout).Encode(map[string]any{
		"success": true,
		"task": map[string]any{
			"id":            task.ID,
			"ticket_number": task.TicketNumber,
			"project_name":  task.ProjectName,
			"title":         task.Title,
			"description":   task.Description,
			"type":          task.TypeDescription,
			"priority": map[string]string{
				"name":  task.PriorityDescription,
				"color": task.PriorityColor,
			},
			"column": map[string]any{
				"id":   task.ColumnID,
				"name": task.ColumnName,
			},
			"assignee":     task.AssigneeName,
			"estimate":     task.Estimate,
			"position":     task.Position,
			"is_blocked":   task.IsBlocked,
			"labels":       task.Labels,
			"parent_tasks": task.ParentTasks,
			"child_tasks":  task.ChildTasks,
			"comments":     comments,
			"created_at":   task.CreatedAt,
			"updated_at":   task.UpdatedAt,
		},
	})
}

func outputHuman(task *models.TaskDetail, colors colors.ColorScheme) error {
	// Initialize styles with the color scheme
	styles.Init(colors)

	var content strings.Builder

	// Header with ticket ID
	ticketID := fmt.Sprintf("%s-%d", task.ProjectName, task.TicketNumber)
	header := styles.TitleStyle.Render(ticketID + ": " + task.Title)
	content.WriteString(header)
	content.WriteString("\n\n")

	// Blocked indicator
	if task.IsBlocked {
		blocked := styles.BlockedStyle.Render("BLOCKED")
		content.WriteString(blocked)
		content.WriteString("\n\n")
	}

	// Description
	if task.Description != "" {
		content.WriteString(styles.SectionStyle.Render("Description"))
		content.WriteString("\n")
		// Indent each line
		for _, line := range strings.Split(task.Description, "\n") {
			content.WriteString("  " + styles.ValueStyle.Render(line) + "\n")
		}
		content.WriteString("\n")
	}

	// Metadata row
	metaLine := fmt.Sprintf("%s %s  %s %s",
		styles.LabelStyle.Render("Type:"),
		styles.ValueStyle.Render(task.TypeDescription),
		styles.LabelStyle.Render("Priority:"),
		styles.BoldColoredText(task.PriorityDescription, task.PriorityColor),
	)
	content.WriteString(metaLine + "\n")

	// Column
	content.WriteString(fmt.Sprintf("%s %s\n",
		styles.LabelStyle.Render("Column:"),
		styles.ValueStyle.Render(task.ColumnName),
	))

	// Assignee
	assigneeDisplay := "None"
	if task.AssigneeName != nil && *task.AssigneeName != "" {
		assigneeDisplay = *task.AssigneeName
	}
	content.WriteString(fmt.Sprintf("%s %s\n",
		styles.LabelStyle.Render("Assignee:"),
		styles.ValueStyle.Render(assigneeDisplay),
	))

	// Estimate
	estimateDisplay := "None"
	if task.Estimate != nil && *task.Estimate != "" {
		estimateDisplay = *task.Estimate
	}
	content.WriteString(fmt.Sprintf("%s %s\n",
		styles.LabelStyle.Render("Estimate:"),
		styles.ValueStyle.Render(estimateDisplay),
	))

	// Timestamps
	if !task.CreatedAt.IsZero() {
		content.WriteString(fmt.Sprintf("%s %s\n",
			styles.LabelStyle.Render("Created:"),
			styles.SubtitleStyle.Render(task.CreatedAt.Format("Jan 2, 2006 3:04 PM")),
		))
	}
	if !task.UpdatedAt.IsZero() {
		content.WriteString(fmt.Sprintf("%s %s\n",
			styles.LabelStyle.Render("Updated:"),
			styles.SubtitleStyle.Render(task.UpdatedAt.Format("Jan 2, 2006 3:04 PM")),
		))
	}

	// Labels
	if len(task.Labels) > 0 {
		content.WriteString("\n")
		content.WriteString(styles.SectionStyle.Render("Labels"))
		content.WriteString("\n  ")
		var labelChips []string
		for _, label := range task.Labels {
			labelChips = append(labelChips, styles.RenderLabelChip(label))
		}
		content.WriteString(strings.Join(labelChips, " ") + "\n")
	}

	// Organize relationships
	var blockingChildren []*models.TaskReference
	var blockingParents []*models.TaskReference
	var nonBlockingParents []*models.TaskReference
	var nonBlockingChildren []*models.TaskReference

	for _, parent := range task.ParentTasks {
		if parent.IsBlocking {
			blockingParents = append(blockingParents, parent)
		} else {
			nonBlockingParents = append(nonBlockingParents, parent)
		}
	}

	for _, child := range task.ChildTasks {
		if child.IsBlocking {
			blockingChildren = append(blockingChildren, child)
		} else {
			nonBlockingChildren = append(nonBlockingChildren, child)
		}
	}

	// Blocked By section
	if len(blockingChildren) > 0 {
		content.WriteString("\n")
		content.WriteString(styles.SectionStyle.Render("Blocked By"))
		content.WriteString("\n")
		for _, child := range blockingChildren {
			content.WriteString("  " + styles.RenderTaskReference(child) + "\n")
		}
	}

	// Blocking section
	if len(blockingParents) > 0 {
		content.WriteString("\n")
		content.WriteString(styles.SectionStyle.Render("Blocking"))
		content.WriteString("\n")
		for _, parent := range blockingParents {
			content.WriteString("  " + styles.RenderTaskReference(parent) + "\n")
		}
	}

	// Parent Tasks
	if len(nonBlockingParents) > 0 {
		content.WriteString("\n")
		content.WriteString(styles.SectionStyle.Render("Parent Tasks"))
		content.WriteString("\n")
		for _, parent := range nonBlockingParents {
			content.WriteString("  " + styles.RenderTaskReferenceWithLabel(parent) + "\n")
		}
	}

	// Child Tasks
	if len(nonBlockingChildren) > 0 {
		content.WriteString("\n")
		content.WriteString(styles.SectionStyle.Render("Child Tasks"))
		content.WriteString("\n")
		for _, child := range nonBlockingChildren {
			content.WriteString("  " + styles.RenderTaskReferenceWithLabel(child) + "\n")
		}
	}

	// Comments
	if len(task.Comments) > 0 {
		content.WriteString("\n")
		header := fmt.Sprintf("Comments (%d)", len(task.Comments))
		content.WriteString(styles.SectionStyle.Render(header))
		content.WriteString("\n")
		for _, comment := range task.Comments {
			// Author and timestamp
			timestamp := comment.CreatedAt.Format("Jan 2, 2006 3:04 PM")
			meta := fmt.Sprintf("[%s - %s]", comment.Author, timestamp)
			content.WriteString("  " + styles.SubtitleStyle.Render(meta) + "\n")
			// Comment content (indented)
			for _, line := range strings.Split(comment.Message, "\n") {
				content.WriteString("    " + styles.ValueStyle.Render(line) + "\n")
			}
			content.WriteString("\n")
		}
	}

	// Render the card
	fmt.Println(styles.RenderCard(content.String()))

	return nil
}
