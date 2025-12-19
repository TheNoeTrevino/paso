package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// ColumnCmd returns the column parent command
func ColumnCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "column",
		Short: "Manage columns",
	}

	cmd.AddCommand(columnCreateCmd())
	cmd.AddCommand(columnListCmd())
	cmd.AddCommand(columnUpdateCmd())
	cmd.AddCommand(columnDeleteCmd())

	return cmd
}

// columnCreateCmd returns the column create subcommand
func columnCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new column",
		Long: `Create a new column in a project.

Examples:
  # Create column at end (human-readable output)
  paso column create --name="Review" --project=1

  # JSON output for agents
  paso column create --name="Review" --project=1 --json

  # Quiet mode for bash capture
  COLUMN_ID=$(paso column create --name="Review" --project=1 --quiet)

  # Create column after specific column
  paso column create --name="Done" --project=1 --after=3
`,
		RunE: runColumnCreate,
	}

	// Required flags
	cmd.Flags().String("name", "", "Column name (required)")
	if err := cmd.MarkFlagRequired("name"); err != nil {
		log.Printf("Error marking flag as required: %v", err)
	}

	cmd.Flags().Int("project", 0, "Project ID (required)")
	if err := cmd.MarkFlagRequired("project"); err != nil {
		log.Printf("Error marking flag as required: %v", err)
	}

	// Optional flags
	cmd.Flags().Int("after", 0, "Insert after column ID (0 = append to end)")

	// Agent-friendly flags
	cmd.Flags().Bool("json", false, "Output in JSON format")
	cmd.Flags().Bool("quiet", false, "Minimal output (ID only)")

	return cmd
}

func runColumnCreate(cmd *cobra.Command, args []string) error {
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
	project, err := cli.Repo.GetProjectByID(ctx, columnProject)
	if err != nil {
		if fmtErr := formatter.Error("PROJECT_NOT_FOUND", fmt.Sprintf("project %d not found", columnProject)); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		os.Exit(ExitNotFound)
	}

	// Validate after column if specified
	var afterID *int
	if columnAfter > 0 {
		afterCol, err := cli.Repo.GetColumnByID(ctx, columnAfter)
		if err != nil {
			if fmtErr := formatter.Error("COLUMN_NOT_FOUND", fmt.Sprintf("column %d not found", columnAfter)); fmtErr != nil {
				log.Printf("Error formatting error message: %v", fmtErr)
			}
			os.Exit(ExitNotFound)
		}
		// Verify column belongs to same project
		if afterCol.ProjectID != columnProject {
			if fmtErr := formatter.Error("INVALID_COLUMN", fmt.Sprintf("column %d does not belong to project %d", columnAfter, columnProject)); fmtErr != nil {
				log.Printf("Error formatting error message: %v", fmtErr)
			}
			os.Exit(ExitValidation)
		}
		afterID = &columnAfter
	}

	// Create column
	column, err := cli.Repo.CreateColumn(ctx, columnName, columnProject, afterID)
	if err != nil {
		if fmtErr := formatter.Error("COLUMN_CREATE_ERROR", err.Error()); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		return err
	}

	// Output based on mode
	if quietMode {
		fmt.Printf("%d\n", column.ID)
		return nil
	}

	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
			"success": true,
			"column": map[string]interface{}{
				"id":         column.ID,
				"name":       column.Name,
				"project_id": column.ProjectID,
			},
		})
	}

	// Human-readable output
	fmt.Printf("✓ Column '%s' created successfully (ID: %d)\n", columnName, column.ID)
	fmt.Printf("  Project: %s\n", project.Name)
	return nil
}

// columnListCmd returns the column list subcommand
func columnListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List columns in a project",
		Long: `List all columns in a project (in order).

Examples:
  # Human-readable list
  paso column list --project=1

  # JSON output for agents
  paso column list --project=1 --json

  # Quiet mode (one ID per line)
  paso column list --project=1 --quiet
`,
		RunE: runColumnList,
	}

	// Required flags
	cmd.Flags().Int("project", 0, "Project ID (required)")
	if err := cmd.MarkFlagRequired("project"); err != nil {
		log.Printf("Error marking flag as required: %v", err)
	}

	// Agent-friendly flags
	cmd.Flags().Bool("json", false, "Output in JSON format")
	cmd.Flags().Bool("quiet", false, "Minimal output (IDs only)")

	return cmd
}

func runColumnList(cmd *cobra.Command, args []string) error {
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
	project, err := cli.Repo.GetProjectByID(ctx, columnProject)
	if err != nil {
		if fmtErr := formatter.Error("PROJECT_NOT_FOUND", fmt.Sprintf("project %d not found", columnProject)); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		os.Exit(ExitNotFound)
	}

	// Get columns
	columns, err := cli.Repo.GetColumnsByProject(ctx, columnProject)
	if err != nil {
		if fmtErr := formatter.Error("COLUMN_FETCH_ERROR", err.Error()); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		return err
	}

	// Output based on mode
	if quietMode {
		for _, col := range columns {
			fmt.Printf("%d\n", col.ID)
		}
		return nil
	}

	if jsonOutput {
		columnList := make([]map[string]interface{}, len(columns))
		for i, col := range columns {
			columnList[i] = map[string]interface{}{
				"id":         col.ID,
				"name":       col.Name,
				"project_id": col.ProjectID,
			}
		}
		return json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
			"success": true,
			"columns": columnList,
		})
	}

	// Human-readable output
	if len(columns) == 0 {
		fmt.Printf("No columns found in project '%s'\n", project.Name)
		return nil
	}

	fmt.Printf("Columns in project '%s':\n", project.Name)
	for i, col := range columns {
		fmt.Printf("  %d. %s (ID: %d)\n", i+1, col.Name, col.ID)
	}
	return nil
}

// columnUpdateCmd returns the column update subcommand
func columnUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a column",
		Long: `Update a column's name.

Examples:
  # Update column name (human-readable output)
  paso column update --id=1 --name="Completed"

  # JSON output for agents
  paso column update --id=1 --name="Completed" --json

  # Quiet mode
  paso column update --id=1 --name="Completed" --quiet
`,
		RunE: runColumnUpdate,
	}

	// Required flags
	cmd.Flags().Int("id", 0, "Column ID (required)")
	if err := cmd.MarkFlagRequired("id"); err != nil {
		log.Printf("Error marking flag as required: %v", err)
	}

	cmd.Flags().String("name", "", "New column name (required)")
	if err := cmd.MarkFlagRequired("name"); err != nil {
		log.Printf("Error marking flag as required: %v", err)
	}

	// Agent-friendly flags
	cmd.Flags().Bool("json", false, "Output in JSON format")
	cmd.Flags().Bool("quiet", false, "Minimal output")

	return cmd
}

func runColumnUpdate(cmd *cobra.Command, args []string) error {
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

	// Validate column exists
	column, err := cli.Repo.GetColumnByID(ctx, columnID)
	if err != nil {
		if fmtErr := formatter.Error("COLUMN_NOT_FOUND", fmt.Sprintf("column %d not found", columnID)); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		os.Exit(ExitNotFound)
	}

	oldName := column.Name

	// Update column
	if err := cli.Repo.UpdateColumnName(ctx, columnID, columnName); err != nil {
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
			"column": map[string]interface{}{
				"id":       columnID,
				"name":     columnName,
				"old_name": oldName,
			},
		})
	}

	// Human-readable output
	fmt.Printf("✓ Column %d updated successfully\n", columnID)
	fmt.Printf("  '%s' → '%s'\n", oldName, columnName)
	return nil
}

// columnDeleteCmd returns the column delete subcommand
func columnDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a column",
		Long: `Delete a column by ID (requires confirmation unless --force or --quiet).

Warning: Deleting a column will move all tasks in that column to the project's first column.

Examples:
  # Delete with confirmation
  paso column delete --id=1

  # Skip confirmation
  paso column delete --id=1 --force

  # Quiet mode (no confirmation)
  paso column delete --id=1 --quiet
`,
		RunE: runColumnDelete,
	}

	// Required flags
	cmd.Flags().Int("id", 0, "Column ID (required)")
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

func runColumnDelete(cmd *cobra.Command, args []string) error {
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

	// Get column details for confirmation
	column, err := cli.Repo.GetColumnByID(ctx, columnID)
	if err != nil {
		if fmtErr := formatter.Error("COLUMN_NOT_FOUND", fmt.Sprintf("column %d not found", columnID)); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		os.Exit(ExitNotFound)
	}

	// Ask for confirmation unless force or quiet mode
	if !force && !quietMode {
		fmt.Println("⚠ Warning: Deleting column will move all tasks to the project's first column")
		fmt.Printf("Delete column #%d: '%s'? (y/N): ", columnID, column.Name)
		var response string
		if _, err := fmt.Scanln(&response); err != nil {
			log.Printf("Error reading user input: %v", err)
		}
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Println("Cancelled")
			return nil
		}
	}

	// Delete the column
	if err := cli.Repo.DeleteColumn(ctx, columnID); err != nil {
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
			"success":   true,
			"column_id": columnID,
		})
	}

	fmt.Printf("✓ Column %d deleted successfully\n", columnID)
	return nil
}
