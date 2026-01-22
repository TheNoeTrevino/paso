package taskevent

// validateEventParams validates common event parameters
func validateEventParams(taskID int, content string) error {
	if taskID <= 0 {
		return ErrInvalidTaskID
	}
	if content == "" {
		return ErrEmptyContent
	}
	return nil
}
