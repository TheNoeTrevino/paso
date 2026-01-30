package task

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
