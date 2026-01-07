package label

// validateCreateLabelRequest validates all fields for creating a label.
// All fields are required and must pass validation.
func validateCreateLabelRequest(req CreateLabelRequest) error {
	if err := validateProjectID(req.ProjectID); err != nil {
		return err
	}
	if err := validateLabelName(req.Name); err != nil {
		return err
	}
	return validateHexColor(req.Color)
}

// validateUpdateLabelRequest validates fields for updating a label.
// Only provided fields (non-nil pointers) are validated.
func validateUpdateLabelRequest(req UpdateLabelRequest) error {
	if err := validateLabelID(req.ID); err != nil {
		return err
	}
	if req.Name != nil {
		if err := validateLabelName(*req.Name); err != nil {
			return err
		}
	}
	if req.Color != nil {
		if err := validateHexColor(*req.Color); err != nil {
			return err
		}
	}
	return nil
}

// ============================================================================
// FIELD-LEVEL VALIDATORS
// ============================================================================

// validateLabelID validates a label ID.
func validateLabelID(id int) error {
	if id <= 0 {
		return ErrInvalidLabelID
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

// validateTaskID validates a task ID.
func validateTaskID(id int) error {
	if id <= 0 {
		return ErrInvalidTaskID
	}
	return nil
}

// validateLabelName validates a label name.
func validateLabelName(name string) error {
	if name == "" {
		return ErrEmptyName
	}
	if len(name) > 50 {
		return ErrNameTooLong
	}
	return nil
}

// validateHexColor validates a hex color format.
func validateHexColor(color string) error {
	if !hexColorRegex.MatchString(color) {
		return ErrInvalidColor
	}
	return nil
}
