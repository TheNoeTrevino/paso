package column

// validateCreateColumnRequest validates all fields for creating a column.
// All required fields must pass validation.
func validateCreateColumnRequest(req CreateColumnRequest) error {
	if err := validateColumnName(req.Name); err != nil {
		return err
	}
	if err := validateProjectID(req.ProjectID); err != nil {
		return err
	}
	if req.AfterID != nil {
		if err := validateColumnID(*req.AfterID); err != nil {
			return err
		}
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

// validateColumnName validates a column name.
func validateColumnName(name string) error {
	if name == "" {
		return ErrEmptyName
	}
	if len(name) > 50 {
		return ErrNameTooLong
	}
	return nil
}
