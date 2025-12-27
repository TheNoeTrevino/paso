package project

import (
	"testing"

	"github.com/thenoetrevino/paso/internal/models"
)

func TestMarkBlockingChains(t *testing.T) {
	tests := []struct {
		name     string
		tree     *models.TaskTreeNode
		expected map[int]bool // map of task IDs to expected InBlockingPath value
	}{
		{
			name: "simple blocker - marks parent as in blocking path",
			tree: &models.TaskTreeNode{
				ID:           1,
				TicketNumber: 1,
				Title:        "Parent Task",
				Children: []*models.TaskTreeNode{
					{
						ID:           2,
						TicketNumber: 2,
						Title:        "Blocking Child",
						IsBlocking:   true,
						Children:     []*models.TaskTreeNode{},
					},
				},
			},
			expected: map[int]bool{
				1: true, // Parent should be marked because child is a blocker
				2: true, // Child should be marked because it is a blocker
			},
		},
		{
			name: "no blockers - nothing marked",
			tree: &models.TaskTreeNode{
				ID:           1,
				TicketNumber: 1,
				Title:        "Parent Task",
				Children: []*models.TaskTreeNode{
					{
						ID:           2,
						TicketNumber: 2,
						Title:        "Normal Child",
						IsBlocking:   false,
						Children:     []*models.TaskTreeNode{},
					},
				},
			},
			expected: map[int]bool{
				1: false,
				2: false,
			},
		},
		{
			name: "deep blocker - marks entire chain",
			tree: &models.TaskTreeNode{
				ID:           1,
				TicketNumber: 1,
				Title:        "Root Task",
				Children: []*models.TaskTreeNode{
					{
						ID:           2,
						TicketNumber: 2,
						Title:        "Mid Task",
						IsBlocking:   false,
						Children: []*models.TaskTreeNode{
							{
								ID:           3,
								TicketNumber: 3,
								Title:        "Deep Blocker",
								IsBlocking:   true,
								Children:     []*models.TaskTreeNode{},
							},
						},
					},
				},
			},
			expected: map[int]bool{
				1: true, // Root should be marked (descendant is blocker)
				2: true, // Mid should be marked (child is blocker)
				3: true, // Deep should be marked (is blocker)
			},
		},
		{
			name: "multiple children with one blocker",
			tree: &models.TaskTreeNode{
				ID:           1,
				TicketNumber: 1,
				Title:        "Parent Task",
				Children: []*models.TaskTreeNode{
					{
						ID:           2,
						TicketNumber: 2,
						Title:        "Normal Child 1",
						IsBlocking:   false,
						Children:     []*models.TaskTreeNode{},
					},
					{
						ID:           3,
						TicketNumber: 3,
						Title:        "Blocking Child",
						IsBlocking:   true,
						Children:     []*models.TaskTreeNode{},
					},
					{
						ID:           4,
						TicketNumber: 4,
						Title:        "Normal Child 2",
						IsBlocking:   false,
						Children:     []*models.TaskTreeNode{},
					},
				},
			},
			expected: map[int]bool{
				1: true,  // Parent marked because one child is blocker
				2: false, // Normal child not marked
				3: true,  // Blocking child marked
				4: false, // Normal child not marked
			},
		},
		{
			name: "mixed chain - blocker in one branch only",
			tree: &models.TaskTreeNode{
				ID:           1,
				TicketNumber: 1,
				Title:        "Root Task",
				Children: []*models.TaskTreeNode{
					{
						ID:           2,
						TicketNumber: 2,
						Title:        "Branch A",
						IsBlocking:   false,
						Children: []*models.TaskTreeNode{
							{
								ID:           3,
								TicketNumber: 3,
								Title:        "Branch A Child (blocker)",
								IsBlocking:   true,
								Children:     []*models.TaskTreeNode{},
							},
						},
					},
					{
						ID:           4,
						TicketNumber: 4,
						Title:        "Branch B",
						IsBlocking:   false,
						Children: []*models.TaskTreeNode{
							{
								ID:           5,
								TicketNumber: 5,
								Title:        "Branch B Child (normal)",
								IsBlocking:   false,
								Children:     []*models.TaskTreeNode{},
							},
						},
					},
				},
			},
			expected: map[int]bool{
				1: true,  // Root marked (descendant is blocker)
				2: true,  // Branch A marked (child is blocker)
				3: true,  // Branch A child marked (is blocker)
				4: false, // Branch B not marked (no blockers)
				5: false, // Branch B child not marked
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run markBlockingChains
			markBlockingChains(tt.tree)

			// Check results recursively
			checkNode := func(node *models.TaskTreeNode) {
				expected, exists := tt.expected[node.ID]
				if !exists {
					t.Errorf("Node ID %d not found in expected results", node.ID)
					return
				}
				if node.InBlockingPath != expected {
					t.Errorf("Node ID %d: expected InBlockingPath=%v, got %v",
						node.ID, expected, node.InBlockingPath)
				}
			}

			// Walk the tree
			var walk func(*models.TaskTreeNode)
			walk = func(node *models.TaskTreeNode) {
				checkNode(node)
				for _, child := range node.Children {
					walk(child)
				}
			}
			walk(tt.tree)
		})
	}
}

func TestMarkBlockingChainsEmptyTree(t *testing.T) {
	// Test with a node that has no children
	node := &models.TaskTreeNode{
		ID:           1,
		TicketNumber: 1,
		Title:        "Solo Task",
		IsBlocking:   false,
		Children:     []*models.TaskTreeNode{},
	}

	result := markBlockingChains(node)

	if result {
		t.Errorf("Expected markBlockingChains to return false for non-blocking solo task, got true")
	}
	if node.InBlockingPath {
		t.Errorf("Expected InBlockingPath to be false for non-blocking solo task, got true")
	}
}

func TestMarkBlockingChainsSelfBlocking(t *testing.T) {
	// Test with a root node that is itself blocking (edge case)
	node := &models.TaskTreeNode{
		ID:           1,
		TicketNumber: 1,
		Title:        "Self Blocking Task",
		IsBlocking:   true,
		Children:     []*models.TaskTreeNode{},
	}

	result := markBlockingChains(node)

	if !result {
		t.Errorf("Expected markBlockingChains to return true for self-blocking task, got false")
	}
	if !node.InBlockingPath {
		t.Errorf("Expected InBlockingPath to be true for self-blocking task, got false")
	}
}
