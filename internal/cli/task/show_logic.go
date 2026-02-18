package task

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/thenoetrevino/paso/internal/cli/styles"
	"github.com/thenoetrevino/paso/internal/config/colors"
	"github.com/thenoetrevino/paso/internal/models"
)

// ShowInput represents the parsed input for showing a task
type ShowInput struct {
	TaskID int
}

// ShowResult represents the output of a successful show operation
type ShowResult struct {
	TaskDetail *models.TaskDetail
}

// TaskRelationships organizes tasks into blocking and non-blocking categories
type TaskRelationships struct {
	BlockingChildren    []*models.TaskReference
	BlockingParents     []*models.TaskReference
	NonBlockingParents  []*models.TaskReference
	NonBlockingChildren []*models.TaskReference
}

// ParseShowTaskID parses task ID from positional arguments or flags
func ParseShowTaskID(args []string, flagID int) (*ShowInput, error) {
	var taskID int

	if len(args) > 0 {
		var err error
		taskID, err = strconv.Atoi(args[0])
		if err != nil {
			return nil, fmt.Errorf("task ID must be a positive integer")
		}
	} else {
		taskID = flagID
	}

	if taskID <= 0 {
		return nil, fmt.Errorf("task ID must be a positive integer")
	}

	return &ShowInput{TaskID: taskID}, nil
}

// FormatShowJSON generates the JSON output structure
func FormatShowJSON(task *models.TaskDetail) map[string]any {
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

	return map[string]any{
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
	}
}

// OrganizeTaskRelationships categorizes parent and child tasks by blocking status
func OrganizeTaskRelationships(task *models.TaskDetail) *TaskRelationships {
	relationships := &TaskRelationships{
		BlockingChildren:    make([]*models.TaskReference, 0),
		BlockingParents:     make([]*models.TaskReference, 0),
		NonBlockingParents:  make([]*models.TaskReference, 0),
		NonBlockingChildren: make([]*models.TaskReference, 0),
	}

	for _, parent := range task.ParentTasks {
		if parent.IsBlocking {
			relationships.BlockingParents = append(relationships.BlockingParents, parent)
		} else {
			relationships.NonBlockingParents = append(relationships.NonBlockingParents, parent)
		}
	}

	for _, child := range task.ChildTasks {
		if child.IsBlocking {
			relationships.BlockingChildren = append(relationships.BlockingChildren, child)
		} else {
			relationships.NonBlockingChildren = append(relationships.NonBlockingChildren, child)
		}
	}

	return relationships
}

// FormatShowHuman generates the human-readable output
func FormatShowHuman(task *models.TaskDetail, colorScheme colors.ColorScheme) string {
	// Initialize styles with the color scheme
	styles.Init(colorScheme)

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
		WriteIndentedLines(&content, task.Description, "  ", styles.ValueStyle)
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
	fmt.Fprintf(&content, "%s %s\n",
		styles.LabelStyle.Render("Column:"),
		styles.ValueStyle.Render(task.ColumnName),
	)

	// Assignee
	fmt.Fprintf(&content, "%s %s\n",
		styles.LabelStyle.Render("Assignee:"),
		styles.ValueStyle.Render(DisplayOrDefault(task.AssigneeName, "None")),
	)

	// Estimate
	fmt.Fprintf(&content, "%s %s\n",
		styles.LabelStyle.Render("Estimate:"),
		styles.ValueStyle.Render(DisplayOrDefault(task.Estimate, "None")),
	)

	// Timestamps
	if !task.CreatedAt.IsZero() {
		fmt.Fprintf(&content, "%s %s\n",
			styles.LabelStyle.Render("Created:"),
			styles.SubtitleStyle.Render(task.CreatedAt.Format("Jan 2, 2006 3:04 PM")),
		)
	}
	if !task.UpdatedAt.IsZero() {
		fmt.Fprintf(&content, "%s %s\n",
			styles.LabelStyle.Render("Updated:"),
			styles.SubtitleStyle.Render(task.UpdatedAt.Format("Jan 2, 2006 3:04 PM")),
		)
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
	relationships := OrganizeTaskRelationships(task)

	// Blocked By section
	if len(relationships.BlockingChildren) > 0 {
		content.WriteString("\n")
		content.WriteString(styles.SectionStyle.Render("Blocked By"))
		content.WriteString("\n")
		for _, child := range relationships.BlockingChildren {
			content.WriteString("  " + styles.RenderTaskReference(child) + "\n")
		}
	}

	// Blocking section
	if len(relationships.BlockingParents) > 0 {
		content.WriteString("\n")
		content.WriteString(styles.SectionStyle.Render("Blocking"))
		content.WriteString("\n")
		for _, parent := range relationships.BlockingParents {
			content.WriteString("  " + styles.RenderTaskReference(parent) + "\n")
		}
	}

	// Parent Tasks
	if len(relationships.NonBlockingParents) > 0 {
		content.WriteString("\n")
		content.WriteString(styles.SectionStyle.Render("Parent Tasks"))
		content.WriteString("\n")
		for _, parent := range relationships.NonBlockingParents {
			content.WriteString("  " + styles.RenderTaskReferenceWithLabel(parent) + "\n")
		}
	}

	// Child Tasks
	if len(relationships.NonBlockingChildren) > 0 {
		content.WriteString("\n")
		content.WriteString(styles.SectionStyle.Render("Child Tasks"))
		content.WriteString("\n")
		for _, child := range relationships.NonBlockingChildren {
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
			WriteIndentedLines(&content, comment.Message, "    ", styles.ValueStyle)
			content.WriteString("\n")
		}
	}

	// Render the card
	return styles.RenderCard(content.String())
}
