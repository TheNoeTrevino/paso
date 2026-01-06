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

	var projectID int
	if len(args) > 0 {
		var err error
		projectID, err = strconv.Atoi(args[0])
		if err != nil {
			projectID = 0
		}
	} else {
		// TODO: chore and handle the never existent error
		projectID, _ = cmd.Flags().GetInt("project-id")
	}

	jsonOutput, _ := cmd.Flags().GetBool("json")
	quietMode, _ := cmd.Flags().GetBool("quiet")

	formatter := &cli.OutputFormatter{JSON: jsonOutput, Quiet: quietMode}

	if projectID <= 0 {
		if fmtErr := formatter.ErrorWithSuggestion("INVALID_PROJECT_ID",
			"project ID must be a positive integer",
			"Usage: paso project tree <project-id> or paso project tree --project-id=<id>"); fmtErr != nil {
			slog.Error("failed to formatting error message", "error", fmtErr)
		}
		os.Exit(cli.ExitUsage)
	}

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

	for _, root := range tree {
		markBlockingChains(root)
	}

	// Output in appropriate format
	if quietMode {
		outputQuietTree(tree, "", true)
		return nil
	}

	if jsonOutput {
		return outputJSONTree(projectID, tree)
	}

	return outputStyledTree(tree)
}

// markBlockingChains marks nodes that are part of a blocking chain
// Returns true if this node or any descendant is a blocker
func markBlockingChains(node *models.TaskTreeNode) bool {
	hasBlockerInSubtree := false

	for _, child := range node.Children {
		if markBlockingChains(child) {
			hasBlockerInSubtree = true
		}
		if child.IsBlocking {
			hasBlockerInSubtree = true
		}
	}

	if hasBlockerInSubtree || node.IsBlocking {
		node.InBlockingPath = true
	}

	return hasBlockerInSubtree || node.IsBlocking
}

// hasIncompleteChild returns true if any direct child is not completed
func hasIncompleteChild(node *models.TaskTreeNode) bool {
	for _, child := range node.Children {
		if !child.IsCompleted {
			return true
		}
	}
	return false
}

// outputQuietTree outputs the tree in quiet mode (IDs with relation labels)
func outputQuietTree(nodes []*models.TaskTreeNode, prefix string, isRoot bool) {
	for i, node := range nodes {
		isLast := i == len(nodes)-1

		if isRoot {
			fmt.Printf("%d\n", node.ID)
			outputQuietTree(node.Children, "", false)
		} else {
			connector := styles.TreeBranch
			if isLast {
				connector = styles.TreeLastBranch
			}
			fmt.Printf("%s%s%d %s\n", prefix, connector, node.TicketNumber, node.RelationLabel)

			childPrefix := prefix
			if isLast {
				childPrefix += styles.TreeSpace
			} else {
				childPrefix += styles.TreeVertical
			}
			outputQuietTree(node.Children, childPrefix, false)
		}
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
	cfg, err := config.Load()
	if err != nil {
		cfg = &config.Config{
			ColorScheme: config.DefaultColorScheme(),
		}
	}

	styles.Init(cfg.ColorScheme)

	var output strings.Builder
	renderTreeNodes(&output, tree, nil, true, cfg.ColorScheme)

	fmt.Print(output.String())
	return nil
}

// ancestorState tracks whether an ancestor was last and in blocking path
type ancestorState struct {
	isLast         bool
	inBlockingPath bool
}

// renderPrefix renders the accumulated prefix from ancestor states. will dim if shouldDim
func renderPrefix(ancestors []ancestorState, shouldDim bool, colors colors.ColorScheme) string {
	var prefix strings.Builder
	for _, a := range ancestors {
		if a.isLast {
			prefix.WriteString(styles.TreeSpace)
		} else {
			prefix.WriteString(styles.RenderTreeVertical(a.inBlockingPath, shouldDim, colors))
		}
	}
	return prefix.String()
}

// recursive
func renderTreeNodes(output *strings.Builder, nodes []*models.TaskTreeNode, ancestors []ancestorState, isRoot bool, colors colors.ColorScheme) {
	for i, node := range nodes {
		isLast := i == len(nodes)-1
		shouldDim := node.IsCompleted && !hasIncompleteChild(node)

		if isRoot {
			line := styles.RenderTreeRootTask(node.TicketNumber, node.Title, node.ColumnName, shouldDim, colors)
			output.WriteString(line + "\n")
			renderTreeNodes(output, node.Children, nil, false, colors)
			continue
		}

		prefix := renderPrefix(ancestors, shouldDim, colors)
		line := styles.RenderTreeChildLine(prefix, isLast, node, shouldDim, colors)
		output.WriteString(line + "\n")

		childAncestors := append(ancestors, ancestorState{
			isLast:         isLast,
			inBlockingPath: node.InBlockingPath,
		})
		renderTreeNodes(output, node.Children, childAncestors, false, colors)
	}
}
