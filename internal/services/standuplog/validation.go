package standuplog

func validateCreateParams(projectID int, content string) error {
	if projectID <= 0 {
		return ErrInvalidProjectID
	}
	if content == "" {
		return ErrEmptyContent
	}
	return nil
}
