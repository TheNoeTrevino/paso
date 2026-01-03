package project

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
	"github.com/thenoetrevino/paso/internal/cli/styles"
	"github.com/thenoetrevino/paso/internal/config"
	"github.com/thenoetrevino/paso/internal/config/colors"
	"github.com/thenoetrevino/paso/internal/models"
)

// TreeCmd returns the project tree subcommand
func TreeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tree [project-id]",
		Short: "Display tasks in a tree structure",
		Long: `Display all tasks in a project as a hierarchical tree structure.
Subtasks are indented under their parent tasks. Blocking relationships
are highlighted in red to show the blocking chain.`,
		Args: cobra.MaximumNArgs(1),
		RunE: runTree,
	}

	// Flags
	cmd.Flags().Int("project-id", 0, "Project ID (can also be provided as positional argument)")
	cmd.Flags().Bool("json", false, "Output in JSON format")
	cmd.Flags().Bool("quiet", false, "Minimal output (IDs with relation labels in tree order)")

	return cmd
}

func runTree(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	// Parse project ID from positional arg or flag
	var projectID int
	if len(args) > 0 {
		var err error
		projectID, err = strconv.Atoi(args[0])
		if err != nil {
			projectID = 0 // Invalid input, will be caught by validation below
		}
	} else {
		projectID, _ = cmd.Flags().GetInt("project-id")
	}

	jsonOutput, _ := cmd.Flags().GetBool("json")
	quietMode, _ := cmd.Flags().GetBool("quiet")

	formatter := &cli.OutputFormatter{JSON: jsonOutput, Quiet: quietMode}

	// Validate project ID
	if projectID <= 0 {
		if fmtErr := formatter.ErrorWithSuggestion("INVALID_PROJECT_ID",
			"project ID must be a positive integer",
			"Usage: paso project tree <project-id> or paso project tree --project-id=<id>"); fmtErr != nil {
			slog.Error("failed to formatting error message", "error", fmtErr)
		}
		os.Exit(cli.ExitUsage)
	}

	// Initialize CLI
	cliInstance, err := cli.GetCLIFromContext(ctx)
	if err != nil {
		if fmtErr := formatter.Error("INITIALIZATION_ERROR", err.Error()); fmtErr != nil {
			slog.Error("failed to formatting error message", "error", fmtErr)
		}
		return err
	}
	defer func() {
		if err := cliInstance.Close(); err != nil {
			slog.Error("failed to closing CLI", "error", err)
		}
	}()

	// Get task tree
	tree, err := cliInstance.App.TaskService.GetTaskTreeByProject(ctx, projectID)
	if err != nil {
		if fmtErr := formatter.Error("TREE_FETCH_ERROR", err.Error()); fmtErr != nil {
			slog.Error("failed to formatting error message", "error", fmtErr)
		}
		return err
	}

	// Handle empty tree
	if len(tree) == 0 {
		if quietMode {
			return nil
		}
		if jsonOutput {
			return json.NewEncoder(os.Stdout).Encode(map[string]any{
				"success":    true,
				"project_id": projectID,
				"tree":       []any{},
			})
		}
		fmt.Println("No tasks found")
		return nil
	}

	// Mark blocking chains recursively
	for _, root := range tree {
		markBlockingChains(root)
	}

	// Output in appropriate format
	if quietMode {
		outputQuietTree(tree, 0)
		return nil
	}

	if jsonOutput {
		return outputJSONTree(projectID, tree)
	}

	// Human-readable output with lipgloss styling
	return outputStyledTree(tree)
}

// markBlockingChains marks nodes that are part of a blocking chain
// Returns true if this node or any descendant is a blocker
func markBlockingChains(node *models.TaskTreeNode) bool {
	hasBlockerInSubtree := false

	// Recursively check children
	for _, child := range node.Children {
		if markBlockingChains(child) {
			hasBlockerInSubtree = true
		}
		if child.IsBlocking {
			hasBlockerInSubtree = true
		}
	}

	// Mark this node as being in a blocking path if any descendant is a blocker
	// or if this node itself is a blocking relationship
	if hasBlockerInSubtree || node.IsBlocking {
		node.InBlockingPath = true
	}

	return hasBlockerInSubtree || node.IsBlocking
}

// outputQuietTree outputs the tree in quiet mode (IDs with relation labels)
func outputQuietTree(nodes []*models.TaskTreeNode, depth int) {
	indent := strings.Repeat("  ", depth)
	for _, node := range nodes {
		if depth == 0 {
			// Root node - just ID
			fmt.Printf("%d\n", node.ID)
		} else {
			// Child node - ID with relation label
			fmt.Printf("%s%d %s\n", indent, node.ID, node.RelationLabel)
		}
		outputQuietTree(node.Children, depth+1)
	}
}

// treeNodeJSON represents a node in JSON output
type treeNodeJSON struct {
	ID           int             `json:"id"`
	TicketNumber int             `json:"ticket_number"`
	Title        string          `json:"title"`
	ColumnName   string          `json:"column_name"`
	RelationType string          `json:"relation_type,omitempty"`
	IsBlocking   bool            `json:"is_blocking,omitempty"`
	Children     []*treeNodeJSON `json:"children,omitempty"`
}

func convertToJSONTree(nodes []*models.TaskTreeNode) []*treeNodeJSON {
	result := make([]*treeNodeJSON, 0, len(nodes))
	for _, node := range nodes {
		jsonNode := &treeNodeJSON{
			ID:           node.ID,
			TicketNumber: node.TicketNumber,
			Title:        node.Title,
			ColumnName:   node.ColumnName,
			RelationType: node.RelationLabel,
			IsBlocking:   node.IsBlocking,
			Children:     convertToJSONTree(node.Children),
		}
		result = append(result, jsonNode)
	}
	return result
}

func outputJSONTree(projectID int, tree []*models.TaskTreeNode) error {
	return json.NewEncoder(os.Stdout).Encode(map[string]any{
		"success":    true,
		"project_id": projectID,
		"tree":       convertToJSONTree(tree),
	})
}

func outputStyledTree(tree []*models.TaskTreeNode) error {
	// Load config for color scheme
	cfg, err := config.Load()
	if err != nil {
		// Fallback to default colors if config fails to load
		cfg = &config.Config{
			ColorScheme: config.DefaultColorScheme(),
		}
	}

	// Initialize styles
	styles.Init(cfg.ColorScheme)

	var output strings.Builder
	renderTreeNodes(&output, tree, 0, cfg.ColorScheme)

	fmt.Print(output.String())
	return nil
}

func renderTreeNodes(output *strings.Builder, nodes []*models.TaskTreeNode, depth int, colors colors.ColorScheme) {
	for _, node := range nodes {
		indent := strings.Repeat("  ", depth)

		if depth == 0 {
			// Root node - render with title style
			line := styles.RenderTreeRootTask(node.ProjectName, node.TicketNumber, node.Title, node.ColumnName, colors)
			output.WriteString(line + "\n")
		} else {
			// Child node - render with tree connector and relation chip
			line := styles.RenderTreeChildLine(indent, node, colors)
			output.WriteString(line + "\n")
		}

		// Recursively render children
		renderTreeNodes(output, node.Children, depth+1, colors)
	}
}
