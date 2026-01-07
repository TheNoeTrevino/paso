package project

// validateCreateProjectRequest validates all fields for creating a project.
// All required fields must pass validation.
func validateCreateProjectRequest(req CreateProjectRequest) error {
	return validateProjectName(req.Name)
}

// validateUpdateProjectRequest validates fields for updating a project.
// Only provided fields (non-nil pointers) are validated.
func validateUpdateProjectRequest(req UpdateProjectRequest) error {
	if err := validateProjectID(req.ID); err != nil {
		return err
	}
	if req.Name != nil {
		if err := validateProjectName(*req.Name); err != nil {
			return err
		}
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

// validateProjectName validates a project name.
func validateProjectName(name string) error {
	if name == "" {
		return ErrEmptyName
	}
	if len(name) > 100 {
		return ErrNameTooLong
	}
	return nil
}
