package task

import "fmt"

const (
	RelationTypeParentChild = 1
	RelationTypeBlocking    = 2
	RelationTypeRelated     = 3
)

// LinkInput represents the parsed input for linking tasks
type LinkInput struct {
	ParentID       int
	ChildID        int
	RelationTypeID int
	RelationName   string
}

// LinkResult represents the output of a successful link operation
type LinkResult struct {
	ParentID       int
	ChildID        int
	RelationTypeID int
	RelationName   string
}

// ValidateLinkFlags validates mutually exclusive relationship type flags
func ValidateLinkFlags(blocker, related bool) error {
	if blocker && related {
		return fmt.Errorf("cannot specify both --blocker and --related flags")
	}
	return nil
}

// DetermineRelationType determines the relation type ID and name based on flags
func DetermineRelationType(blocker, related bool) (int, string) {
	if blocker {
		return RelationTypeBlocking, "blocking"
	}
	if related {
		return RelationTypeRelated, "related"
	}
	return RelationTypeParentChild, "parent-child"
}

// FormatLinkOutput generates the human-readable output message
func FormatLinkOutput(result *LinkResult) string {
	switch result.RelationTypeID {
	case RelationTypeBlocking:
		return fmt.Sprintf("Created blocking relationship: task %d is blocked by task %d", result.ParentID, result.ChildID)
	case RelationTypeRelated:
		return fmt.Sprintf("Created related relationship between task %d and task %d", result.ParentID, result.ChildID)
	default:
		return fmt.Sprintf("Linked task %d as child of task %d", result.ChildID, result.ParentID)
	}
}

// FormatLinkJSON generates the JSON output structure
func FormatLinkJSON(result *LinkResult) map[string]any {
	return map[string]any{
		"success":          true,
		"parent_id":        result.ParentID,
		"child_id":         result.ChildID,
		"relation_type_id": result.RelationTypeID,
		"relation_type":    result.RelationName,
	}
}
