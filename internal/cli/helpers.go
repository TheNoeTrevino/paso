package cli

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

// ValidateColorHex validates that a color string is in valid hex format #RRGGBB
func ValidateColorHex(color string) error {
	matched, err := regexp.MatchString(`^#[0-9A-Fa-f]{6}$`, color)
	if err != nil {
		return fmt.Errorf("error validating color: %w", err)
	}
	if !matched {
		return fmt.Errorf("color must be in hex format #RRGGBB (e.g., #FF0000), got: %s", color)
	}
	return nil
}

// ParseTaskType maps a type string to its ID
func ParseTaskType(typeStr string) (int, error) {
	types := map[string]int{
		"task":    1,
		"feature": 2,
	}

	id, ok := types[strings.ToLower(typeStr)]
	if !ok {
		return 0, fmt.Errorf("invalid type '%s' (must be: task, feature)", typeStr)
	}
	return id, nil
}

// ParsePriority maps a priority string to its ID
func ParsePriority(priority string) (int, error) {
	priorities := map[string]int{
		"trivial":  1,
		"low":      2,
		"medium":   3,
		"high":     4,
		"critical": 5,
	}

	id, ok := priorities[strings.ToLower(priority)]
	if !ok {
		return 0, fmt.Errorf("invalid priority '%s' (must be: trivial, low, medium, high, critical)", priority)
	}
	return id, nil
}

// GetLabelByID is a helper function to get a single label by ID
// Since the database layer doesn't have GetLabelByID, we iterate through all projects
func GetLabelByID(ctx context.Context, cliInstance *CLI, labelID int) (*struct {
	ID        int
	Name      string
	Color     string
	ProjectID int
}, error,
) {
	// Get all projects to search for the label
	projects, err := cliInstance.Repo().GetAllProjects(ctx)
	if err != nil {
		return nil, err
	}

	for _, project := range projects {
		labels, err := cliInstance.Repo().GetLabelsByProject(ctx, project.ID)
		if err != nil {
			continue
		}
		for _, lbl := range labels {
			if lbl.ID == labelID {
				return &struct {
					ID        int
					Name      string
					Color     string
					ProjectID int
				}{
					ID:        lbl.ID,
					Name:      lbl.Name,
					Color:     lbl.Color,
					ProjectID: lbl.ProjectID,
				}, nil
			}
		}
	}

	return nil, fmt.Errorf("label %d not found", labelID)
}
