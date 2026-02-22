package task

import (
	"fmt"
	"strconv"

	"github.com/thenoetrevino/paso/internal/cli/styles"
	"github.com/thenoetrevino/paso/internal/config/colors"
)

const (
	maxCommentLength      = 1000
	commentTruncateLength = 60
)

// CommentInput represents the parsed input for creating a comment
type CommentInput struct {
	TaskID  int
	Message string
	Author  string
}

// CommentResult represents the output of a successful comment operation
type CommentResult struct {
	CommentID    int
	TaskID       int
	Message      string
	Author       string
	CreatedAt    string // Formatted time string
	TaskTitle    string
	TaskNumber int
	ProjectName  string
}

// ValidateCommentMessage validates the comment message length
func ValidateCommentMessage(message string) error {
	if len(message) > maxCommentLength {
		return fmt.Errorf("message exceeds %d character limit", maxCommentLength)
	}
	return nil
}

// FormatCommentQuiet generates minimal output (comment ID only)
func FormatCommentQuiet(commentID int) string {
	return fmt.Sprintf("%d\n", commentID)
}

// FormatCommentJSON generates the JSON output structure
func FormatCommentJSON(result *CommentResult) map[string]any {
	return map[string]any{
		"success": true,
		"comment": map[string]any{
			"id":         result.CommentID,
			"task_id":    result.TaskID,
			"message":    result.Message,
			"author":     result.Author,
			"created_at": result.CreatedAt,
		},
		"task": map[string]any{
			"id":            result.TaskID,
			"title":         result.TaskTitle,
			"task_number": result.TaskNumber,
			"project":       result.ProjectName,
		},
	}
}

// FormatCommentHuman generates the human-readable output with details
func FormatCommentHuman(result *CommentResult, colorScheme colors.ColorScheme) string {
	details := []styles.Detail{
		{Key: "Task", Value: fmt.Sprintf("#%d (%s)", result.TaskNumber, result.TaskTitle)},
		{Key: "Project", Value: result.ProjectName},
		{Key: "Message", Value: styles.TruncateString(result.Message, commentTruncateLength)},
		{Key: "Comment ID", Value: strconv.Itoa(result.CommentID)},
	}
	return styles.RenderSuccessWithDetails("Comment added", details, colorScheme)
}
