package task

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateCreateTitle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		title         string
		expectedError string
	}{
		{
			name:  "valid title",
			title: "Fix authentication bug",
		},
		{
			name:  "valid title with special characters",
			title: "Add OAuth2.0 support #123",
		},
		{
			name:  "title with leading/trailing spaces",
			title: "  Valid Title  ",
		},
		{
			name:          "empty string",
			title:         "",
			expectedError: "title is required",
		},
		{
			name:          "whitespace only",
			title:         "   ",
			expectedError: "title is required",
		},
		{
			name:          "tabs only",
			title:         "\t\t\t",
			expectedError: "title is required",
		},
		{
			name:          "newlines only",
			title:         "\n\n",
			expectedError: "title is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateCreateTitle(tt.title)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestReadDescription(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		description   string
		reader        io.Reader
		expected      string
		expectedError string
	}{
		{
			name:        "normal description - not from stdin",
			description: "This is a regular description",
			reader:      nil,
			expected:    "This is a regular description",
		},
		{
			name:        "empty description - not from stdin",
			description: "",
			reader:      nil,
			expected:    "",
		},
		{
			name:        "read from stdin - simple text",
			description: "-",
			reader:      strings.NewReader("Description from stdin"),
			expected:    "Description from stdin",
		},
		{
			name:        "read from stdin - multiline",
			description: "-",
			reader:      strings.NewReader("Line 1\nLine 2\nLine 3"),
			expected:    "Line 1\nLine 2\nLine 3",
		},
		{
			name:        "read from stdin - empty",
			description: "-",
			reader:      strings.NewReader(""),
			expected:    "",
		},
		{
			name:          "stdin flag but no reader",
			description:   "-",
			reader:        nil,
			expectedError: "no reader provided",
		},
		{
			name:          "stdin flag with error reader",
			description:   "-",
			reader:        &errorReader{},
			expectedError: "failed to read stdin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := ReadDescription(tt.description, tt.reader)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Empty(t, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestBuildCreateDescription(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		description    string
		stdinAvailable bool
		reader         io.Reader
		expected       string
		expectedError  string
	}{
		{
			name:           "normal description",
			description:    "Regular description",
			stdinAvailable: false,
			reader:         nil,
			expected:       "Regular description",
		},
		{
			name:           "stdin available and requested",
			description:    "-",
			stdinAvailable: true,
			reader:         strings.NewReader("Stdin content"),
			expected:       "Stdin content",
		},
		{
			name:           "stdin requested but not available",
			description:    "-",
			stdinAvailable: false,
			reader:         nil,
			expectedError:  "stdin not available",
		},
		{
			name:           "stdin available with multiline content",
			description:    "-",
			stdinAvailable: true,
			reader:         strings.NewReader("Multi\nLine\nContent"),
			expected:       "Multi\nLine\nContent",
		},
		{
			name:           "empty description",
			description:    "",
			stdinAvailable: false,
			reader:         nil,
			expected:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := BuildCreateDescription(tt.description, tt.stdinAvailable, tt.reader)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Empty(t, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestReadDescriptionWithRealBuffer(t *testing.T) {
	t.Parallel()

	t.Run("large content", func(t *testing.T) {
		t.Parallel()

		largeContent := strings.Repeat("A", 10000)
		reader := strings.NewReader(largeContent)

		result, err := ReadDescription("-", reader)

		require.NoError(t, err)
		assert.Equal(t, largeContent, result)
	})

	t.Run("binary content", func(t *testing.T) {
		t.Parallel()

		binaryContent := []byte{0x00, 0x01, 0x02, 0xFF, 0xFE}
		reader := bytes.NewReader(binaryContent)

		result, err := ReadDescription("-", reader)

		require.NoError(t, err)
		assert.Equal(t, string(binaryContent), result)
	})
}

// errorReader is a helper that always returns an error when read
type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("simulated read error")
}
