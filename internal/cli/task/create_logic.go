package task

import (
	"fmt"
	"io"
	"strings"
)

// CreateInput represents the parsed input for creating a task
type CreateInput struct {
	Title       string
	Description string
	Type        string
	Priority    string
	ProjectID   int
	ColumnName  string
	Assignee    string
	Estimate    string
	ParentID    int
	BlockedByID int
	BlocksID    int
}

// ValidateCreateTitle validates the task title
func ValidateCreateTitle(title string) error {
	if strings.TrimSpace(title) == "" {
		return fmt.Errorf("failed to validate input: title is required")
	}
	return nil
}

// ReadDescription reads description from reader or returns the value directly
// Returns error if reader is provided but reading fails
func ReadDescription(description string, reader io.Reader) (string, error) {
	if description != "-" {
		return description, nil
	}

	if reader == nil {
		return "", fmt.Errorf("failed to read description: no reader provided")
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("failed to read stdin: %w", err)
	}

	return string(data), nil
}

// BuildCreateDescription is a convenience function that reads description from a string value
// This matches the pattern in create.go where stdin is handled
func BuildCreateDescription(description string, stdinAvailable bool, reader io.Reader) (string, error) {
	if description != "-" {
		return description, nil
	}

	if !stdinAvailable {
		return "", fmt.Errorf("failed to read description: stdin not available")
	}

	return ReadDescription(description, reader)
}
