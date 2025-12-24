package events

import (
	"errors"
	"os"
	"syscall"
)

// ErrorCode represents daemon-related error types.
type ErrorCode int

const (
	ErrSocketNotFound ErrorCode = iota
	ErrSocketPermission
	ErrDaemonNotRunning
	ErrConnectionRefused
)

// DaemonError represents a structured daemon error with context.
type DaemonError struct {
	Code    ErrorCode
	Message string
	Hint    string
}

// Error implements the error interface.
func (e *DaemonError) Error() string {
	if e.Hint != "" {
		return e.Message + ". " + e.Hint
	}
	return e.Message
}

// ClassifyDaemonError maps common errors to structured DaemonError types.
func ClassifyDaemonError(err error) *DaemonError {
	if err == nil {
		return nil
	}

	if os.IsNotExist(err) {
		return &DaemonError{
			Code:    ErrSocketNotFound,
			Message: "Socket file not found",
			Hint:    "Start daemon: systemctl --user start paso",
		}
	}

	if os.IsPermission(err) {
		return &DaemonError{
			Code:    ErrSocketPermission,
			Message: "Permission denied",
			Hint:    "Check ~/.paso/ permissions: chmod 700 ~/.paso/",
		}
	}

	var errno syscall.Errno
	if errors.As(err, &errno) && errno == syscall.ECONNREFUSED {
		return &DaemonError{
			Code:    ErrConnectionRefused,
			Message: "Connection refused",
			Hint:    "Daemon may be crashed. Restart: systemctl --user restart paso",
		}
	}

	return &DaemonError{
		Code:    ErrDaemonNotRunning,
		Message: "Daemon not running",
		Hint:    "Start daemon: systemctl --user start paso",
	}
}
