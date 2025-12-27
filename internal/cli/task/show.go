package task

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
	"github.com/thenoetrevino/paso/internal/config"
	"github.com/thenoetrevino/paso/internal/models"
)

// ShowCmd returns the task show subcommand
func ShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show [id]",
		Short: "Show task details",
		Long:  "Display all details of a task including description, relationships, labels, and metadata.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runShow,
	}

	// Flags
	cmd.Flags().Int("id", 0, "Task ID (can also be provided as positional argument)")
	cmd.Flags().Bool("json", false, "Output in JSON format")
	cmd.Flags().Bool("quiet", false, "Minimal output (ID only)")

	return cmd
}

func runShow(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Parse task ID from positional arg or flag
	var taskID int
	if len(args) > 0 {
		fmt.Sscanf(args[0], "%d", &taskID)
	} else {
		taskID, _ = cmd.Flags().GetInt("id")
	}

	jsonOutput, _ := cmd.Flags().GetBool("json")
	quietMode, _ := cmd.Flags().GetBool("quiet")

	formatter := &cli.OutputFormatter{JSON: jsonOutput, Quiet: quietMode}

	// Validate task ID
	if taskID <= 0 {
		if fmtErr := formatter.ErrorWithSuggestion("INVALID_TASK_ID",
			"task ID must be a positive integer",
			"Usage: paso task show <id> or paso task show --id=<id>"); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		os.Exit(cli.ExitUsage)
		return nil
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

	// Get task details
	task, err := cliInstance.App.TaskService.GetTaskDetail(ctx, taskID)
	if err != nil {
		if fmtErr := formatter.Error("TASK_NOT_FOUND", fmt.Sprintf("task %d not found", taskID)); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		os.Exit(cli.ExitNotFound)
		return nil
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
	return json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
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
			"position":     task.Position,
			"is_blocked":   task.IsBlocked,
			"labels":       task.Labels,
			"parent_tasks": task.ParentTasks,
			"child_tasks":  task.ChildTasks,
			"created_at":   task.CreatedAt,
			"updated_at":   task.UpdatedAt,
		},
	})
}

func outputHuman(task *models.TaskDetail, colors config.ColorScheme) error {
	var content strings.Builder

	// Define styles
	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(colors.Accent)).
		Padding(1, 2).
		Width(80)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(colors.Title))

	subtitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(colors.Subtle))

	labelStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(colors.Accent))

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(colors.Normal))

	blockedStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(colors.ErrorFg)).
		Background(lipgloss.Color(colors.ErrorBg)).
		Padding(0, 1)

	sectionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(colors.Accent)).
		Bold(true).
		MarginTop(1)

	// Header with ticket ID
	ticketID := fmt.Sprintf("%s-%d", task.ProjectName, task.TicketNumber)
	header := titleStyle.Render(ticketID + ": " + task.Title)
	content.WriteString(header)
	content.WriteString("\n\n")

	// Blocked indicator
	if task.IsBlocked {
		blocked := blockedStyle.Render("BLOCKED")
		content.WriteString(blocked)
		content.WriteString("\n\n")
	}

	// Description
	if task.Description != "" {
		content.WriteString(sectionStyle.Render("Description"))
		content.WriteString("\n")
		// Indent each line
		for _, line := range strings.Split(task.Description, "\n") {
			content.WriteString("  " + valueStyle.Render(line) + "\n")
		}
		content.WriteString("\n")
	}

	// Metadata row
	metaLine := fmt.Sprintf("%s %s  %s %s",
		labelStyle.Render("Type:"),
		valueStyle.Render(task.TypeDescription),
		labelStyle.Render("Priority:"),
		lipgloss.NewStyle().
			Foreground(lipgloss.Color(task.PriorityColor)).
			Bold(true).
			Render(task.PriorityDescription),
	)
	content.WriteString(metaLine + "\n")

	// Column
	content.WriteString(fmt.Sprintf("%s %s\n",
		labelStyle.Render("Column:"),
		valueStyle.Render(task.ColumnName),
	))

	// Timestamps
	if !task.CreatedAt.IsZero() {
		content.WriteString(fmt.Sprintf("%s %s\n",
			labelStyle.Render("Created:"),
			subtitleStyle.Render(task.CreatedAt.Format("Jan 2, 2006 3:04 PM")),
		))
	}
	if !task.UpdatedAt.IsZero() {
		content.WriteString(fmt.Sprintf("%s %s\n",
			labelStyle.Render("Updated:"),
			subtitleStyle.Render(task.UpdatedAt.Format("Jan 2, 2006 3:04 PM")),
		))
	}

	// Labels
	if len(task.Labels) > 0 {
		content.WriteString("\n")
		content.WriteString(sectionStyle.Render("Labels"))
		content.WriteString("\n")
		var labelChips []string
		for _, label := range task.Labels {
			chip := lipgloss.NewStyle().
				Foreground(lipgloss.Color(label.Color)).
				Bold(true).
				Render("[" + label.Name + "]")
			labelChips = append(labelChips, chip)
		}
		content.WriteString("  " + strings.Join(labelChips, " ") + "\n")
	}

	// Relationships
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
		content.WriteString(sectionStyle.Render("Blocked By"))
		content.WriteString("\n")
		for _, child := range blockingChildren {
			bulletStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color(child.RelationColor))
			taskRef := fmt.Sprintf("%s-%d - %s", child.ProjectName, child.TicketNumber, child.Title)
			content.WriteString("  " + bulletStyle.Render("• "+taskRef) + "\n")
		}
	}

	// Blocking section
	if len(blockingParents) > 0 {
		content.WriteString("\n")
		content.WriteString(sectionStyle.Render("Blocking"))
		content.WriteString("\n")
		for _, parent := range blockingParents {
			bulletStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color(parent.RelationColor))
			taskRef := fmt.Sprintf("%s-%d - %s", parent.ProjectName, parent.TicketNumber, parent.Title)
			content.WriteString("  " + bulletStyle.Render("• "+taskRef) + "\n")
		}
	}

	// Parent Tasks
	if len(nonBlockingParents) > 0 {
		content.WriteString("\n")
		content.WriteString(sectionStyle.Render("Parent Tasks"))
		content.WriteString("\n")
		for _, parent := range nonBlockingParents {
			bulletStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color(parent.RelationColor))
			taskRef := fmt.Sprintf("%s-%d - %s - %s",
				parent.ProjectName, parent.TicketNumber, parent.RelationLabel, parent.Title)
			content.WriteString("  " + bulletStyle.Render("• "+taskRef) + "\n")
		}
	}

	// Child Tasks
	if len(nonBlockingChildren) > 0 {
		content.WriteString("\n")
		content.WriteString(sectionStyle.Render("Child Tasks"))
		content.WriteString("\n")
		for _, child := range nonBlockingChildren {
			bulletStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color(child.RelationColor))
			taskRef := fmt.Sprintf("%s-%d - %s - %s",
				child.ProjectName, child.TicketNumber, child.RelationLabel, child.Title)
			content.WriteString("  " + bulletStyle.Render("• "+taskRef) + "\n")
		}
	}

	// Render the card
	fmt.Println(cardStyle.Render(content.String()))

	return nil
}
