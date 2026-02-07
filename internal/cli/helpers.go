package cli

import (
	"context"
	"fmt"
	"log/slog"
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

// GetProjectID returns the project ID using precedence:
// 1. --project flag (if explicitly set)
// 2. Git branch association (if in git repo with associated project)
// 3. Error (no project found)
//
// DEPRECATED: Use GetProjectIDWithCLI instead.
// This function creates its own CLI instance which wastes resources.
func GetProjectID(cmd *cobra.Command) (int, error) {
	// Check if --project flag was explicitly set
	projectFlag := cmd.Flags().Lookup("project")
	if projectFlag != nil && projectFlag.Changed {
		return cmd.Flags().GetInt("project")
	}

	// Try git branch detection
	ctx := cmd.Context()
	if ctx == nil {
		return 0, fmt.Errorf("no project specified: use --project flag or create a project associated with this branch")
	}

	projectID, err := tryGitBranchDetection(ctx)
	if err == nil {
		return projectID, nil
	}

	return 0, fmt.Errorf("no project specified: use --project flag or create a project associated with this branch")
}

// GetProjectIDWithCLI returns the project ID using precedence:
// 1. --project flag (if explicitly set)
// 2. Git branch association (if in git repo with associated project)
// 3. Error (no project found)
//
// The caller must provide a CLI instance (does not create its own).
// This is more efficient than GetProjectID as it reuses the existing CLI instance.
func GetProjectIDWithCLI(cmd *cobra.Command, cliInstance *CLI) (int, error) {
	// Check if --project flag was explicitly set
	projectFlag := cmd.Flags().Lookup("project")
	if projectFlag != nil && projectFlag.Changed {
		return cmd.Flags().GetInt("project")
	}

	// Try git branch detection
	ctx := cmd.Context()
	if ctx == nil {
		return 0, fmt.Errorf("no project specified: use --project flag or create a project associated with this branch")
	}

	projectID, err := tryGitBranchDetectionWithCLI(ctx, cliInstance)
	if err == nil {
		return projectID, nil
	}

	return 0, fmt.Errorf("no project specified: use --project flag or create a project associated with this branch")
}

// tryGitBranchDetection attempts to find a project by git branch association
// Returns (0, error) if detection fails or no project found
//
// DEPRECATED: Use tryGitBranchDetectionWithCLI instead.
// This function creates its own CLI instance which wastes resources.
func tryGitBranchDetection(ctx context.Context) (int, error) {
	cliInstance, err := GetCLIFromContext(ctx)
	if err != nil {
		return 0, err
	}
	defer func() {
		if closeErr := cliInstance.Close(); closeErr != nil {
			slog.Warn("failed to close CLI instance during git branch detection",
				"error", closeErr,
				"function", "tryGitBranchDetection")
		}
	}()

	gitInfo := cliInstance.App.GitDetector.DetectGitInfo(ctx)

	if !gitInfo.IsValidForAssociation() {
		return 0, fmt.Errorf("not in valid git repository state")
	}

	project, err := cliInstance.App.ProjectService.GetProjectByGitBranch(ctx, gitInfo.CurrentBranch)
	if err != nil {
		slog.Debug("failed to lookup project by git branch",
			"branch", gitInfo.CurrentBranch,
			"error", err)
		return 0, err
	}

	if project == nil {
		return 0, fmt.Errorf("no project associated with branch: %s", gitInfo.CurrentBranch)
	}

	return project.ID, nil
}

// tryGitBranchDetectionWithCLI attempts to find a project by git branch association.
// Accepts an existing CLI instance instead of creating a new one.
// Returns (0, error) if detection fails or no project found.
func tryGitBranchDetectionWithCLI(ctx context.Context, cliInstance *CLI) (int, error) {
	gitInfo := cliInstance.App.GitDetector.DetectGitInfo(ctx)

	if !gitInfo.IsValidForAssociation() {
		return 0, fmt.Errorf("not in valid git repository state")
	}

	project, err := cliInstance.App.ProjectService.GetProjectByGitBranch(ctx, gitInfo.CurrentBranch)
	if err != nil {
		slog.Debug("failed to lookup project by git branch",
			"branch", gitInfo.CurrentBranch,
			"error", err)
		return 0, err
	}

	if project == nil {
		return 0, fmt.Errorf("no project associated with branch: %s", gitInfo.CurrentBranch)
	}

	return project.ID, nil
}
