package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

// LabelCmd returns the label parent command
func LabelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "label",
		Short: "Manage labels",
	}

	cmd.AddCommand(labelCreateCmd())
	cmd.AddCommand(labelListCmd())
	cmd.AddCommand(labelUpdateCmd())
	cmd.AddCommand(labelDeleteCmd())
	cmd.AddCommand(labelAttachCmd())
	cmd.AddCommand(labelDetachCmd())

	return cmd
}

// labelCreateCmd returns the label create subcommand
func labelCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new label",
		Long: `Create a new label with a name and color.

Examples:
  # Create label (human-readable output)
  paso label create --name="bug" --color="#FF0000" --project=1

  # JSON output for agents
  paso label create --name="bug" --color="#FF0000" --project=1 --json

  # Quiet mode for bash capture
  LABEL_ID=$(paso label create --name="bug" --color="#FF0000" --project=1 --quiet)
`,
		RunE: runLabelCreate,
	}

	// Required flags
	cmd.Flags().String("name", "", "Label name (required)")
	if err := cmd.MarkFlagRequired("name"); err != nil {
		log.Printf("Error marking flag as required: %v", err)
	}

	cmd.Flags().String("color", "", "Label color in hex format #RRGGBB (required)")
	if err := cmd.MarkFlagRequired("color"); err != nil {
		log.Printf("Error marking flag as required: %v", err)
	}

	cmd.Flags().Int("project", 0, "Project ID (required)")
	if err := cmd.MarkFlagRequired("project"); err != nil {
		log.Printf("Error marking flag as required: %v", err)
	}

	// Agent-friendly flags
	cmd.Flags().Bool("json", false, "Output in JSON format")
	cmd.Flags().Bool("quiet", false, "Minimal output (ID only)")

	return cmd
}

func runLabelCreate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	formatter := &OutputFormatter{JSON: jsonOutput, Quiet: quietMode}

	// Initialize CLI
	cli, err := NewCLI(ctx)
	if err != nil {
		if fmtErr := formatter.Error("INITIALIZATION_ERROR", err.Error()); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		return err
	}
	defer func() {
		if err := cli.Close(); err != nil {
			log.Printf("Error closing CLI: %v", err)
		}
	}()

	// Validate color format
	if err := validateColorHex(labelColor); err != nil {
		if fmtErr := formatter.Error("INVALID_COLOR", err.Error()); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		os.Exit(5) // Exit code 5 = validation error
	}

	// Validate project exists
	project, err := cli.Repo.GetProjectByID(ctx, labelProject)
	if err != nil {
		if fmtErr := formatter.Error("PROJECT_NOT_FOUND", fmt.Sprintf("project %d not found", labelProject)); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		os.Exit(3) // Exit code 3 = not found
	}

	// Create label
	label, err := cli.Repo.CreateLabel(ctx, labelProject, labelName, labelColor)
	if err != nil {
		if fmtErr := formatter.Error("LABEL_CREATE_ERROR", err.Error()); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		return err
	}

	// Output based on mode
	if quietMode {
		fmt.Printf("%d\n", label.ID)
		return nil
	}

	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
			"success": true,
			"label": map[string]interface{}{
				"id":         label.ID,
				"name":       label.Name,
				"color":      label.Color,
				"project_id": label.ProjectID,
			},
		})
	}

	// Human-readable output
	fmt.Printf("✓ Label '%s' created successfully (ID: %d)\n", labelName, label.ID)
	fmt.Printf("  Project: %s\n", project.Name)
	fmt.Printf("  Color: %s\n", labelColor)
	return nil
}

// labelListCmd returns the label list subcommand
func labelListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List labels in a project",
		Long: `List all labels in a project.

Examples:
  # Human-readable list
  paso label list --project=1

  # JSON output for agents
  paso label list --project=1 --json

  # Quiet mode (one ID per line)
  paso label list --project=1 --quiet
`,
		RunE: runLabelList,
	}

	// Required flags
	cmd.Flags().IntVar(&labelProject, "project", 0, "Project ID (required)")
	if err := cmd.MarkFlagRequired("project"); err != nil {
		log.Printf("Error marking flag as required: %v", err)
	}

	// Agent-friendly flags
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	cmd.Flags().BoolVar(&quietMode, "quiet", false, "Minimal output (IDs only)")

	return cmd
}

func runLabelList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	formatter := &OutputFormatter{JSON: jsonOutput, Quiet: quietMode}

	// Initialize CLI
	cli, err := NewCLI(ctx)
	if err != nil {
		if fmtErr := formatter.Error("INITIALIZATION_ERROR", err.Error()); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		return err
	}
	defer func() {
		if err := cli.Close(); err != nil {
			log.Printf("Error closing CLI: %v", err)
		}
	}()

	// Validate project exists
	project, err := cli.Repo.GetProjectByID(ctx, labelProject)
	if err != nil {
		if fmtErr := formatter.Error("PROJECT_NOT_FOUND", fmt.Sprintf("project %d not found", labelProject)); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		os.Exit(3)
	}

	// Get labels
	labels, err := cli.Repo.GetLabelsByProject(ctx, labelProject)
	if err != nil {
		if fmtErr := formatter.Error("LABEL_FETCH_ERROR", err.Error()); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		return err
	}

	// Output based on mode
	if quietMode {
		for _, lbl := range labels {
			fmt.Printf("%d\n", lbl.ID)
		}
		return nil
	}

	if jsonOutput {
		labelList := make([]map[string]interface{}, len(labels))
		for i, lbl := range labels {
			labelList[i] = map[string]interface{}{
				"id":         lbl.ID,
				"name":       lbl.Name,
				"color":      lbl.Color,
				"project_id": lbl.ProjectID,
			}
		}
		return json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
			"success": true,
			"labels":  labelList,
		})
	}

	// Human-readable output
	if len(labels) == 0 {
		fmt.Printf("No labels found in project '%s'\n", project.Name)
		return nil
	}

	fmt.Printf("Labels in project '%s':\n", project.Name)
	fmt.Printf("  %-4s %-20s %s\n", "ID", "Name", "Color")
	fmt.Println("  " + strings.Repeat("-", 50))
	for _, lbl := range labels {
		fmt.Printf("  %-4d %-20s %s\n", lbl.ID, lbl.Name, lbl.Color)
	}
	return nil
}

// labelUpdateCmd returns the label update subcommand
func labelUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a label",
		Long: `Update a label's name and/or color.

Examples:
  # Update both name and color
  paso label update --id=1 --name="critical-bug" --color="#FF0000"

  # Update only name (keeps existing color)
  paso label update --id=1 --name="critical-bug"

  # Update only color (keeps existing name)
  paso label update --id=1 --color="#FF0000"

  # JSON output
  paso label update --id=1 --name="urgent" --json
`,
		RunE: runLabelUpdate,
	}

	// Required flags
	cmd.Flags().IntVar(&labelID, "id", 0, "Label ID (required)")
	if err := cmd.MarkFlagRequired("id"); err != nil {
		log.Printf("Error marking flag as required: %v", err)
	}

	// Optional flags (at least one required)
	cmd.Flags().StringVar(&labelName, "name", "", "New label name")
	cmd.Flags().StringVar(&labelColor, "color", "", "New label color in hex format #RRGGBB")

	// Agent-friendly flags
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	cmd.Flags().BoolVar(&quietMode, "quiet", false, "Minimal output")

	return cmd
}

func runLabelUpdate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	formatter := &OutputFormatter{JSON: jsonOutput, Quiet: quietMode}

	// Initialize CLI
	cli, err := NewCLI(ctx)
	if err != nil {
		if fmtErr := formatter.Error("INITIALIZATION_ERROR", err.Error()); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		return err
	}
	defer func() {
		if err := cli.Close(); err != nil {
			log.Printf("Error closing CLI: %v", err)
		}
	}()

	// Check if at least one update flag is provided
	nameProvided := cmd.Flags().Changed("name")
	colorProvided := cmd.Flags().Changed("color")

	if !nameProvided && !colorProvided {
		if fmtErr := formatter.Error("MISSING_FLAGS", "at least one of --name or --color must be provided"); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		os.Exit(5)
	}

	// Get existing label to fetch current values
	currentLabel, err := getLabelByID(ctx, cli, labelID)
	if err != nil {
		if fmtErr := formatter.Error("LABEL_NOT_FOUND", err.Error()); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		os.Exit(3)
	}

	// Use existing values if not provided
	newName := currentLabel.Name
	if nameProvided {
		newName = labelName
	}

	newColor := currentLabel.Color
	if colorProvided {
		// Validate color format
		if err := validateColorHex(labelColor); err != nil {
			if fmtErr := formatter.Error("INVALID_COLOR", err.Error()); fmtErr != nil {
				log.Printf("Error formatting error message: %v", fmtErr)
			}
			os.Exit(5)
		}
		newColor = labelColor
	}

	// Update label
	if err := cli.Repo.UpdateLabel(ctx, labelID, newName, newColor); err != nil {
		if fmtErr := formatter.Error("UPDATE_ERROR", err.Error()); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		return err
	}

	// Output based on mode
	if quietMode {
		return nil
	}

	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
			"success": true,
			"label": map[string]interface{}{
				"id":        labelID,
				"name":      newName,
				"color":     newColor,
				"old_name":  currentLabel.Name,
				"old_color": currentLabel.Color,
			},
		})
	}

	// Human-readable output
	fmt.Printf("✓ Label %d updated successfully\n", labelID)
	if nameProvided {
		fmt.Printf("  Name: '%s' → '%s'\n", currentLabel.Name, newName)
	}
	if colorProvided {
		fmt.Printf("  Color: %s → %s\n", currentLabel.Color, newColor)
	}
	return nil
}

// labelDeleteCmd returns the label delete subcommand
func labelDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a label",
		Long: `Delete a label by ID (requires confirmation unless --force or --quiet).

Examples:
  # Delete with confirmation
  paso label delete --id=1

  # Skip confirmation
  paso label delete --id=1 --force

  # Quiet mode (no confirmation)
  paso label delete --id=1 --quiet
`,
		RunE: runLabelDelete,
	}

	// Required flags
	cmd.Flags().Int("id", 0, "Label ID (required)")
	if err := cmd.MarkFlagRequired("id"); err != nil {
		log.Printf("Error marking flag as required: %v", err)
	}

	// Optional flags
	cmd.Flags().Bool("force", false, "Skip confirmation")

	// Agent-friendly flags
	cmd.Flags().Bool("json", false, "Output in JSON format")
	cmd.Flags().Bool("quiet", false, "Minimal output")

	return cmd
}

func runLabelDelete(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	formatter := &OutputFormatter{JSON: jsonOutput, Quiet: quietMode}

	// Initialize CLI
	cli, err := NewCLI(ctx)
	if err != nil {
		if fmtErr := formatter.Error("INITIALIZATION_ERROR", err.Error()); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		return err
	}
	defer func() {
		if err := cli.Close(); err != nil {
			log.Printf("Error closing CLI: %v", err)
		}
	}()

	// Get label details for confirmation
	label, err := getLabelByID(ctx, cli, labelID)
	if err != nil {
		if fmtErr := formatter.Error("LABEL_NOT_FOUND", err.Error()); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		os.Exit(3)
	}

	// Ask for confirmation unless force or quiet mode
	if !force && !quietMode {
		fmt.Printf("Delete label #%d: '%s'? (y/N): ", labelID, label.Name)
		var response string
		if _, err := fmt.Scanln(&response); err != nil {
			log.Printf("Error reading user input: %v", err)
		}
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Println("Cancelled")
			return nil
		}
	}

	// Delete the label
	if err := cli.Repo.DeleteLabel(ctx, labelID); err != nil {
		if fmtErr := formatter.Error("DELETE_ERROR", err.Error()); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		return err
	}

	// Output success
	if quietMode {
		return nil
	}

	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
			"success":  true,
			"label_id": labelID,
		})
	}

	fmt.Printf("✓ Label %d deleted successfully\n", labelID)
	return nil
}

// labelAttachCmd returns the label attach subcommand
func labelAttachCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "attach",
		Short: "Attach a label to a task",
		Long: `Attach a label to a task by their IDs.

Examples:
  # Attach label to task
  paso label attach --task=5 --label=2

  # JSON output
  paso label attach --task=5 --label=2 --json

  # Quiet mode
  paso label attach --task=5 --label=2 --quiet
`,
		RunE: runLabelAttach,
	}

	// Required flags
	cmd.Flags().IntVar(&taskID, "task", 0, "Task ID (required)")
	if err := cmd.MarkFlagRequired("task"); err != nil {
		log.Printf("Error marking flag as required: %v", err)
	}

	cmd.Flags().IntVar(&labelID, "label", 0, "Label ID (required)")
	if err := cmd.MarkFlagRequired("label"); err != nil {
		log.Printf("Error marking flag as required: %v", err)
	}

	// Agent-friendly flags
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	cmd.Flags().BoolVar(&quietMode, "quiet", false, "Minimal output")

	return cmd
}

func runLabelAttach(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	formatter := &OutputFormatter{JSON: jsonOutput, Quiet: quietMode}

	// Initialize CLI
	cli, err := NewCLI(ctx)
	if err != nil {
		if fmtErr := formatter.Error("INITIALIZATION_ERROR", err.Error()); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		return err
	}
	defer func() {
		if err := cli.Close(); err != nil {
			log.Printf("Error closing CLI: %v", err)
		}
	}()

	// Validate task exists
	task, err := cli.Repo.GetTaskDetail(ctx, taskID)
	if err != nil {
		if fmtErr := formatter.Error("TASK_NOT_FOUND", fmt.Sprintf("task %d not found", taskID)); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		os.Exit(3)
	}

	// Get task's project ID via column
	column, err := cli.Repo.GetColumnByID(ctx, task.ColumnID)
	if err != nil {
		if fmtErr := formatter.Error("COLUMN_FETCH_ERROR", err.Error()); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		return err
	}
	taskProjectID := column.ProjectID

	// Validate label exists
	label, err := getLabelByID(ctx, cli, labelID)
	if err != nil {
		if fmtErr := formatter.Error("LABEL_NOT_FOUND", err.Error()); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		os.Exit(3)
	}

	// Verify task and label belong to same project
	if taskProjectID != label.ProjectID {
		if fmtErr := formatter.Error("PROJECT_MISMATCH", fmt.Sprintf("task %d and label %d do not belong to the same project", taskID, labelID)); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		os.Exit(5)
	}

	// Attach label to task
	if err := cli.Repo.AddLabelToTask(ctx, taskID, labelID); err != nil {
		if fmtErr := formatter.Error("ATTACH_ERROR", err.Error()); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		return err
	}

	// Output success
	if quietMode {
		return nil
	}

	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
			"success":  true,
			"task_id":  taskID,
			"label_id": labelID,
		})
	}

	fmt.Printf("✓ Label '%s' attached to task #%d\n", label.Name, taskID)
	return nil
}

// labelDetachCmd returns the label detach subcommand
func labelDetachCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "detach",
		Short: "Detach a label from a task",
		Long: `Detach a label from a task by their IDs.

Examples:
  # Detach label from task
  paso label detach --task=5 --label=2

  # JSON output
  paso label detach --task=5 --label=2 --json

  # Quiet mode
  paso label detach --task=5 --label=2 --quiet
`,
		RunE: runLabelDetach,
	}

	// Required flags
	cmd.Flags().Int("task", 0, "Task ID (required)")
	if err := cmd.MarkFlagRequired("task"); err != nil {
		log.Printf("Error marking flag as required: %v", err)
	}

	cmd.Flags().Int("label", 0, "Label ID (required)")
	if err := cmd.MarkFlagRequired("label"); err != nil {
		log.Printf("Error marking flag as required: %v", err)
	}

	// Agent-friendly flags
	cmd.Flags().Bool("json", false, "Output in JSON format")
	cmd.Flags().Bool("quiet", false, "Minimal output")

	return cmd
}

func runLabelDetach(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	formatter := &OutputFormatter{JSON: jsonOutput, Quiet: quietMode}

	// Initialize CLI
	cli, err := NewCLI(ctx)
	if err != nil {
		if fmtErr := formatter.Error("INITIALIZATION_ERROR", err.Error()); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		return err
	}
	defer func() {
		if err := cli.Close(); err != nil {
			log.Printf("Error closing CLI: %v", err)
		}
	}()

	// Detach label from task (no validation needed - removing non-existent association is not an error)
	if err := cli.Repo.RemoveLabelFromTask(ctx, taskID, labelID); err != nil {
		if fmtErr := formatter.Error("DETACH_ERROR", err.Error()); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		return err
	}

	// Output success
	if quietMode {
		return nil
	}

	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
			"success":  true,
			"task_id":  taskID,
			"label_id": labelID,
		})
	}

	fmt.Printf("✓ Label #%d detached from task #%d\n", labelID, taskID)
	return nil
}

// validateColorHex validates that a color string is in valid hex format #RRGGBB
func validateColorHex(color string) error {
	matched, err := regexp.MatchString(`^#[0-9A-Fa-f]{6}$`, color)
	if err != nil {
		return fmt.Errorf("error validating color: %w", err)
	}
	if !matched {
		return fmt.Errorf("color must be in hex format #RRGGBB (e.g., #FF0000), got: %s", color)
	}
	return nil
}

// getLabelByID is a helper function to get a single label by ID
// Since the database layer doesn't have GetLabelByID, we iterate through all projects
func getLabelByID(ctx context.Context, cli *CLI, labelID int) (*struct {
	ID        int
	Name      string
	Color     string
	ProjectID int
}, error) {
	// Get all projects to search for the label
	projects, err := cli.Repo.GetAllProjects(ctx)
	if err != nil {
		return nil, err
	}

	for _, project := range projects {
		labels, err := cli.Repo.GetLabelsByProject(ctx, project.ID)
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
