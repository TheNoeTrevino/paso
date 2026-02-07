package assignee

import (
	"fmt"
	"strings"
)

// validateAssigneeID validates that an assignee ID is positive
func validateAssigneeID(id int) error {
	if id <= 0 {
		return fmt.Errorf("%w: %d", ErrInvalidAssigneeID, id)
	}
	return nil
}

// validateName validates that an assignee name is not empty and within length limits
func validateName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("%w: name cannot be empty", ErrInvalidAssigneeName)
	}
	if len(name) > 255 {
		return fmt.Errorf("%w: name cannot exceed 255 characters", ErrInvalidAssigneeName)
	}
	return nil
}
