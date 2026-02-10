package project

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/tree"
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
are highlighted in red to show the blocking chain.

Examples:
  # Using positional argument
  paso project tree 1

  # Using shorthand flag
  paso project tree -p 1

  # JSON output
  paso project tree 1 -j

  # Long-form flags also supported
  paso project tree --project=1 --json`,
		Args: cobra.MaximumNArgs(1),
		RunE: runTree,
	}

	// Flags
	cmd.Flags().IntP("project", "p", 0, "Project ID (uses git branch association if not specified)")
	cmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	cmd.Flags().BoolP("quiet", "q", false, "Minimal output (IDs with relation labels in tree order)")

	return cmd
}

func runTree(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	jsonOutput, _ := cmd.Flags().GetBool("json")
	quietMode, _ := cmd.Flags().GetBool("quiet")

	formatter := &cli.OutputFormatter{JSON: jsonOutput, Quiet: quietMode}

	cliInstance, err := cli.GetCLIFromContext(ctx)
	if err != nil {
		return formatter.Error(cli.ExitError, "INITIALIZATION_ERROR", err.Error())
	}
	defer func() {
		if err := cliInstance.Close(); err != nil {
			slog.Error("failed to close CLI", "error", err)
		}
	}()

	var projectID int
	if len(args) > 0 {
		projectID, err = strconv.Atoi(args[0])
		if err != nil || projectID <= 0 {
			return formatter.ErrorWithSuggestion(cli.ExitUsage, "INVALID_PROJECT_ID",
				"project ID must be a positive integer",
				"Usage: paso project tree <project-id> or paso project tree -p <id>")
		}
	} else {
		projectID, err = cli.GetProjectIDWithCLI(cmd, cliInstance)
		if err != nil {
			return formatter.ErrorWithSuggestion(cli.ExitUsage, "NO_PROJECT",
				err.Error(),
				"Use --project flag or create a project associated with this git branch")
		}
	}

	tree, err := cliInstance.App.TaskService.GetTaskTreeByProject(ctx, projectID)
	if err != nil {
		return formatter.Error(cli.ExitError, "TREE_FETCH_ERROR", err.Error())
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

// taskNode wraps a TaskTreeNode for lipgloss tree rendering
type taskNode struct {
	node   *models.TaskTreeNode
	colors colors.ColorScheme
	isRoot bool
}

func (t taskNode) String() string {
	shouldDim := t.node.IsCompleted && !hasIncompleteChild(t.node)

	if t.isRoot {
		color := t.colors.Title
		if shouldDim {
			color = styles.DimColor(color, styles.CompletedDimIntensity)
		}
		return lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(color)).
			Render(fmt.Sprintf("%d: %s - %s", t.node.TicketNumber, t.node.Title, t.node.ColumnName))
	}

	textColor := t.colors.Normal
	if t.node.IsBlocking {
		textColor = t.colors.ErrorFg
	}
	if shouldDim {
		textColor = styles.DimColor(textColor, styles.CompletedDimIntensity)
	}

	ticketStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(textColor))
	relationStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(textColor)).
		Bold(t.node.IsBlocking)
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(textColor))

	ticketNum := ticketStyle.Render(fmt.Sprintf("%d: ", t.node.TicketNumber))
	relationChip := relationStyle.Render(t.node.RelationLabel)
	titleInfo := titleStyle.Render(fmt.Sprintf(" %s - %s", t.node.Title, t.node.ColumnName))

	return fmt.Sprintf("%s%s%s", ticketNum, relationChip, titleInfo)
}

// buildTaskTrees converts a slice of TaskTreeNodes to lipgloss trees (one per root)
func buildTaskTrees(nodes []*models.TaskTreeNode, clrs colors.ColorScheme) []*tree.Tree {
	trees := make([]*tree.Tree, 0, len(nodes))

	for _, node := range nodes {
		rootNode := taskNode{node: node, colors: clrs, isRoot: true}
		t := tree.Root(rootNode)
		addChildren(t, node.Children, clrs)
		trees = append(trees, t)
	}

	return trees
}

func addChildren(parent *tree.Tree, children []*models.TaskTreeNode, clrs colors.ColorScheme) {
	for _, child := range children {
		childNode := taskNode{node: child, colors: clrs, isRoot: false}
		if len(child.Children) > 0 {
			subtree := tree.Root(childNode)
			addChildren(subtree, child.Children, clrs)
			parent.Child(subtree)
		} else {
			parent.Child(childNode)
		}
	}
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

func outputStyledTree(taskTree []*models.TaskTreeNode) error {
	cfg, err := config.Load()
	if err != nil {
		cfg = &config.Config{
			ColorScheme: config.DefaultColorScheme(),
		}
	}

	styles.Init(cfg.ColorScheme)

	enumeratorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(cfg.ColorScheme.Subtle))
	styledIndenter := makeStyledIndenter(cfg.ColorScheme.Subtle)
	trees := buildTaskTrees(taskTree, cfg.ColorScheme)

	for _, t := range trees {
		t.Enumerator(tree.RoundedEnumerator).
			EnumeratorStyle(enumeratorStyle).
			Indenter(styledIndenter)
		fmt.Println(t)
	}
	return nil
}

func makeStyledIndenter(color string) tree.Indenter {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
	styledVertical := style.Render("│") + "  "
	spaces := "   "

	return func(children tree.Children, index int) string {
		if children.Length()-1 == index {
			return spaces
		}
		return styledVertical
	}
}
