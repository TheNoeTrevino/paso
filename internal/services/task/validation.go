package task

import (
	"fmt"
	"strings"
	"unicode"
)

// validateCreateTaskRequest validates all fields for creating a task.
// All required fields must pass validation.
func validateCreateTaskRequest(req CreateTaskRequest) error {
	if err := validateTitle(req.Title); err != nil {
		return err
	}
	if err := validateColumnID(req.ColumnID); err != nil {
		return err
	}
	if err := validatePosition(req.Position); err != nil {
		return err
	}
	if err := validatePriorityID(req.PriorityID); err != nil {
		return err
	}
	if err := validateTypeID(req.TypeID); err != nil {
		return err
	}
	return nil
}

// validateUpdateTaskRequest validates fields for updating a task.
// Only provided fields (non-nil pointers) are validated.
func validateUpdateTaskRequest(req UpdateTaskRequest) error {
	if err := validateTaskID(req.TaskID); err != nil {
		return err
	}
	if req.Title != nil {
		if err := validateTitle(*req.Title); err != nil {
			return err
		}
	}
	if req.PriorityID != nil {
		if err := validatePriorityID(*req.PriorityID); err != nil {
			return err
		}
	}
	if req.TypeID != nil {
		if err := validateTypeID(*req.TypeID); err != nil {
			return err
		}
	}
	return nil
}

// validateCreateCommentRequest validates all fields for creating a comment.
func validateCreateCommentRequest(req CreateCommentRequest) error {
	if err := validateTaskID(req.TaskID); err != nil {
		return err
	}
	return validateCommentMessage(req.Message)
}

// validateUpdateCommentRequest validates fields for updating a comment.
func validateUpdateCommentRequest(req UpdateCommentRequest) error {
	if err := validateCommentID(req.CommentID); err != nil {
		return err
	}
	return validateCommentMessage(req.Message)
}

// validateTaskID validates a task ID.
func validateTaskID(id int) error {
	if id <= 0 {
		return ErrInvalidTaskID
	}
	return nil
}

// validateColumnID validates a column ID.
func validateColumnID(id int) error {
	if id <= 0 {
		return ErrInvalidColumnID
	}
	return nil
}

// validateProjectID validates a project ID.
func validateProjectID(id int) error {
	if id <= 0 {
		return ErrInvalidProjectID
	}
	return nil
}

// validateLabelID validates a label ID.
func validateLabelID(id int) error {
	if id <= 0 {
		return ErrInvalidLabelID
	}
	return nil
}

// validateCommentID validates a comment ID.
func validateCommentID(id int) error {
	if id <= 0 {
		return ErrInvalidCommentID
	}
	return nil
}

// validatePriorityID validates a priority ID.
// Note: 0 is allowed as it means "use default priority".
func validatePriorityID(id int) error {
	if id < 0 {
		return ErrInvalidPriority
	}
	return nil
}

// validateTypeID validates a type ID.
// Note: 0 is allowed as it means "use default type".
func validateTypeID(id int) error {
	if id < 0 {
		return ErrInvalidType
	}
	return nil
}

// validateTitle validates a task title.
func validateTitle(title string) error {
	if title == "" {
		return ErrEmptyTitle
	}
	if len(title) > 255 {
		return ErrTitleTooLong
	}
	return nil
}

// validatePosition validates a task position.
func validatePosition(position int) error {
	if position < 0 {
		return ErrInvalidPosition
	}
	return nil
}

// validateCommentMessage validates a comment message.
func validateCommentMessage(message string) error {
	if message == "" {
		return ErrEmptyCommentMessage
	}
	if len(message) > 1000 {
		return ErrCommentMessageTooLong
	}
	return nil
}

// ValidateEstimate validates a time estimate string.
// Valid formats: "1h", "2d", "3w", "4m", "1w2d3h", etc.
// Invalid formats: "1x", "abc", "d1", "1", "1ww", "1d2d", empty strings
// nil or empty estimates are allowed (estimates are optional).
func ValidateEstimate(estimate *string) error {
	// nil or empty estimates are allowed (optional field)
	if estimate == nil || *estimate == "" {
		return nil
	}

	est := strings.TrimSpace(*estimate)
	if est == "" {
		return nil
	}

	// Track which units we've seen to prevent duplicates
	seenUnits := make(map[rune]bool)
	validUnits := map[rune]bool{
		'h': true, // hours
		'd': true, // days
		'w': true, // weeks
		'm': true, // months
	}

	i := 0
	foundAtLeastOneSegment := false

	for i < len(est) {
		// Must start with a digit
		if !unicode.IsDigit(rune(est[i])) {
			return fmt.Errorf("%w: expected digit at position %d, got '%c'", ErrInvalidEstimateFormat, i+1, est[i])
		}

		// Consume all digits
		start := i
		for i < len(est) && unicode.IsDigit(rune(est[i])) {
			i++
		}

		// Must be followed by a unit letter
		if i >= len(est) {
			return fmt.Errorf("%w: missing unit after number '%s'", ErrInvalidEstimateFormat, est[start:i])
		}

		unit := rune(est[i])

		// Validate the unit is one of h, d, w, m
		if !validUnits[unit] {
			return fmt.Errorf("%w: '%c' is not valid (use h, d, w, or m)", ErrInvalidEstimateUnit, unit)
		}

		// Check for duplicate units
		if seenUnits[unit] {
			return fmt.Errorf("%w: '%c' appears multiple times", ErrDuplicateEstimateUnit, unit)
		}
		seenUnits[unit] = true

		i++
		foundAtLeastOneSegment = true
	}

	if !foundAtLeastOneSegment {
		return ErrInvalidEstimateFormat
	}

	return nil
}
