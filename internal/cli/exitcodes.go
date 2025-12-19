package cli

// Exit codes for CLI commands.
// These codes follow Unix conventions and provide consistent error reporting
// across all CLI commands.
const (
	// ExitSuccess indicates the command completed successfully.
	// Use for: Normal, successful command execution.
	ExitSuccess = 0

	// ExitError indicates a general error occurred.
	// Use for: Database errors, network errors, unexpected failures,
	// or any error that doesn't fit the specific categories below.
	ExitError = 1

	// ExitUsage indicates incorrect command usage.
	// Use for: Missing required flags, invalid flag combinations,
	// or when the user needs to provide different arguments.
	ExitUsage = 2

	// ExitNotFound indicates a requested resource was not found.
	// Use for: Task not found, project not found, column not found,
	// label not found, or any case where a resource ID or name doesn't exist.
	ExitNotFound = 3

	// ExitDataErr indicates invalid or malformed data.
	// Use for: Invalid JSON input, corrupted data, or data that cannot be processed.
	ExitDataErr = 4

	// ExitValidation indicates a validation error.
	// Use for: Invalid priority values, invalid type values, invalid status,
	// or any case where input fails validation rules.
	ExitValidation = 5
)
