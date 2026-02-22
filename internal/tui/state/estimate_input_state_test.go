package state

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Mock validation functions for testing
var (
	alwaysValid = func(s *string) error {
		return nil
	}
	alwaysInvalid = func(s *string) error {
		return errors.New("invalid")
	}
	validIfNoX = func(s *string) error {
		if s != nil && len(*s) > 0 && (*s)[len(*s)-1] == 'x' {
			return errors.New("cannot end with x")
		}
		return nil
	}
)

func TestEstimateInputState_AppendChar(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		initialBuffer  string
		initialError   string
		char           string
		validate       func(*string) error
		hint           string
		expectedBuffer string
		expectedError  string
	}{
		{
			name:           "append valid character clears error",
			initialBuffer:  "2",
			initialError:   "some error",
			char:           "d",
			validate:       alwaysValid,
			hint:           "invalid format",
			expectedBuffer: "2d",
			expectedError:  "",
		},
		{
			name:           "append invalid character sets error",
			initialBuffer:  "2",
			initialError:   "",
			char:           "x",
			validate:       validIfNoX,
			hint:           "cannot use x",
			expectedBuffer: "2x",
			expectedError:  "cannot use x",
		},
		{
			name:           "append to empty buffer with valid input",
			initialBuffer:  "",
			initialError:   "",
			char:           "1",
			validate:       alwaysValid,
			hint:           "error hint",
			expectedBuffer: "1",
			expectedError:  "",
		},
		{
			name:           "append to empty buffer with invalid input",
			initialBuffer:  "",
			initialError:   "",
			char:           "x",
			validate:       alwaysInvalid,
			hint:           "bad input",
			expectedBuffer: "x",
			expectedError:  "bad input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := &EstimateInputState{
				buffer: tt.initialBuffer,
				error:  tt.initialError,
			}

			s.AppendChar(tt.char, tt.validate, tt.hint)

			assert.Equal(t, tt.expectedBuffer, s.buffer)
			assert.Equal(t, tt.expectedError, s.error)
		})
	}
}

func TestEstimateInputState_Backspace(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		initialBuffer  string
		initialError   string
		validate       func(*string) error
		hint           string
		expectedBuffer string
		expectedError  string
	}{
		{
			name:           "backspace on empty buffer is no-op",
			initialBuffer:  "",
			initialError:   "some error",
			validate:       alwaysValid,
			hint:           "error hint",
			expectedBuffer: "",
			expectedError:  "some error",
		},
		{
			name:           "backspace to empty buffer clears error",
			initialBuffer:  "x",
			initialError:   "bad char",
			validate:       alwaysValid,
			hint:           "error hint",
			expectedBuffer: "",
			expectedError:  "",
		},
		{
			name:           "backspace makes buffer valid clears error",
			initialBuffer:  "2dx",
			initialError:   "cannot end with x",
			validate:       validIfNoX,
			hint:           "cannot use x",
			expectedBuffer: "2d",
			expectedError:  "",
		},
		{
			name:           "backspace leaves buffer invalid keeps error",
			initialBuffer:  "xx",
			initialError:   "",
			validate:       validIfNoX,
			hint:           "cannot use x",
			expectedBuffer: "x",
			expectedError:  "cannot use x",
		},
		{
			name:           "backspace on valid buffer stays valid",
			initialBuffer:  "2d3h",
			initialError:   "",
			validate:       alwaysValid,
			hint:           "error hint",
			expectedBuffer: "2d3",
			expectedError:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := &EstimateInputState{
				buffer: tt.initialBuffer,
				error:  tt.initialError,
			}

			s.Backspace(tt.validate, tt.hint)

			assert.Equal(t, tt.expectedBuffer, s.buffer)
			assert.Equal(t, tt.expectedError, s.error)
		})
	}
}

func TestEstimateInputState_Submit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		initialBuffer  string
		initialError   string
		validate       func(*string) error
		hint           string
		expectedResult string
		expectedOk     bool
		expectedError  string
	}{
		{
			name:           "submit empty buffer succeeds",
			initialBuffer:  "",
			initialError:   "",
			validate:       alwaysValid,
			hint:           "error hint",
			expectedResult: "",
			expectedOk:     true,
			expectedError:  "",
		},
		{
			name:           "submit valid buffer succeeds",
			initialBuffer:  "2d",
			initialError:   "",
			validate:       alwaysValid,
			hint:           "error hint",
			expectedResult: "2d",
			expectedOk:     true,
			expectedError:  "",
		},
		{
			name:           "submit invalid buffer fails and sets error",
			initialBuffer:  "2x",
			initialError:   "",
			validate:       alwaysInvalid,
			hint:           "invalid format",
			expectedResult: "",
			expectedOk:     false,
			expectedError:  "invalid format",
		},
		{
			name:           "submit invalid buffer with existing error updates error",
			initialBuffer:  "bad",
			initialError:   "old error",
			validate:       alwaysInvalid,
			hint:           "new error",
			expectedResult: "",
			expectedOk:     false,
			expectedError:  "new error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := &EstimateInputState{
				buffer: tt.initialBuffer,
				error:  tt.initialError,
			}

			result, ok := s.Submit(tt.validate, tt.hint)

			assert.Equal(t, tt.expectedResult, result)
			assert.Equal(t, tt.expectedOk, ok)
			assert.Equal(t, tt.expectedError, s.error)
		})
	}
}

func TestEstimateInputState_Reset(t *testing.T) {
	t.Parallel()

	s := &EstimateInputState{
		buffer:     "2d3h",
		error:      "some error",
		ReturnMode: EstimateInputMode,
	}

	s.Reset()

	assert.Equal(t, "", s.buffer)
	assert.Equal(t, "", s.error)
	assert.Equal(t, TaskFormMode, s.ReturnMode)
}
