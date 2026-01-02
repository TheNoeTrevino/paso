package cli

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/models"
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

// ParseTaskType maps a type string to its ID.
// Returns an error with a helpful message if the type is not recognized.
// This uses a map lookup with the ok pattern to detect invalid types.
func ParseTaskType(typeStr string) (int, error) {
	types := map[string]int{
		"task":    models.TaskTypeTask,
		"feature": models.TaskTypeFeature,
	}

	id, ok := types[strings.ToLower(typeStr)]
	if !ok {
		return 0, fmt.Errorf("invalid type '%s' (must be: task, feature)", typeStr)
	}
	return id, nil
}

// ParsePriority maps a priority string to its ID.
// Returns an error with a helpful message if the priority is not recognized.
// This uses a map lookup with the ok pattern to detect invalid priorities.
func ParsePriority(priority string) (int, error) {
	priorities := map[string]int{
		"trivial":  models.PriorityTrivial,
		"low":      models.PriorityLow,
		"medium":   models.PriorityMedium,
		"high":     models.PriorityHigh,
		"critical": models.PriorityCritical,
	}

	id, ok := priorities[strings.ToLower(priority)]
	if !ok {
		return 0, fmt.Errorf("invalid priority '%s' (must be: trivial, low, medium, high, critical)", priority)
	}
	return id, nil
}

// FindColumnByName finds a column by name (case-insensitive)
// Returns the column and nil error if found, nil and error if not found
func FindColumnByName(columns []*models.Column, name string) (*models.Column, error) {
	for _, col := range columns {
		if strings.EqualFold(col.Name, name) {
			return col, nil
		}
	}
	return nil, fmt.Errorf("column '%s' not found", name)
}

// FormatAvailableColumns formats column list for error messages
func FormatAvailableColumns(columns []*models.Column) string {
	names := make([]string, len(columns))
	for i, col := range columns {
		names[i] = col.Name
	}
	return strings.Join(names, ", ")
}

// GetCurrentColumnName finds the column name for a given column ID
func GetCurrentColumnName(columns []*models.Column, columnID int) string {
	for _, col := range columns {
		if col.ID == columnID {
			return col.Name
		}
	}
	return "Unknown"
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
	projects, err := cliInstance.App.ProjectService.GetAllProjects(ctx)
	if err != nil {
		return nil, err
	}

	for _, project := range projects {
		labels, err := cliInstance.App.LabelService.GetLabelsByProject(ctx, project.ID)
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

// GetProjectID returns the project ID from flag or environment variable
// Precedence: --project flag > PASO_PROJECT env var > error
func GetProjectID(cmd *cobra.Command) (int, error) {
	// Check if --project flag was explicitly set
	projectFlag := cmd.Flags().Lookup("project")
	if projectFlag != nil && projectFlag.Changed {
		return cmd.Flags().GetInt("project")
	}

	// Fall back to PASO_PROJECT environment variable
	if envProject := os.Getenv("PASO_PROJECT"); envProject != "" {
		var projectID int
		if _, err := fmt.Sscanf(envProject, "%d", &projectID); err == nil {
			return projectID, nil
		}
	}

	return 0, fmt.Errorf("no project specified: use --project flag or set with 'eval $(paso use project <project-id>)'")
}
