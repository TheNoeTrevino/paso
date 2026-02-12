package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock Types for Testing

type mockDataWithID struct {
	ID   int
	Name string
}

func (m mockDataWithID) GetID() int {
	return m.ID
}

type mockDataWithoutID struct {
	Name  string
	Value int
}

type mockDataWithPointerID struct {
	ID   int
	Data string
}

func (m *mockDataWithPointerID) GetID() int {
	return m.ID
}

// Success Method Tests - JSON Mode

func TestOutputFormatter_Success_JSON(t *testing.T) {
	tests := []struct {
		name     string
		data     any
		validate func(t *testing.T, result map[string]any)
	}{
		{
			name: "map data",
			data: map[string]any{"test": "value", "number": float64(42)},
			validate: func(t *testing.T, result map[string]any) {
				assert.True(t, result["success"].(bool))
				dataMap := result["data"].(map[string]any)
				assert.Equal(t, "value", dataMap["test"])
			},
		},
		{
			name: "struct with ID",
			data: mockDataWithID{ID: 123, Name: "Test"},
			validate: func(t *testing.T, result map[string]any) {
				assert.True(t, result["success"].(bool))
				dataMap := result["data"].(map[string]any)
				assert.Equal(t, "Test", dataMap["Name"])
			},
		},
		{
			name: "string data",
			data: "simple string",
			validate: func(t *testing.T, result map[string]any) {
				assert.True(t, result["success"].(bool))
				assert.Equal(t, "simple string", result["data"])
			},
		},
		{
			name: "integer data",
			data: 42,
			validate: func(t *testing.T, result map[string]any) {
				assert.True(t, result["success"].(bool))
				// JSON unmarshals numbers as float64
				assert.Equal(t, float64(42), result["data"].(float64))
			},
		},
		{
			name: "nil data",
			data: nil,
			validate: func(t *testing.T, result map[string]any) {
				assert.True(t, result["success"].(bool))
				assert.Nil(t, result["data"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, err := os.Pipe()
			require.NoError(t, err, "Failed to create pipe for test")
			os.Stdout = w

			formatter := &OutputFormatter{JSON: true, Quiet: false}
			err = formatter.Success(tt.data)
			assert.NoError(t, err, "Success method should not return error")

			_ = w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			_, _ = buf.ReadFrom(r)
			output := buf.String()

			// Verify JSON output
			var result map[string]any
			require.NoError(t, json.Unmarshal([]byte(output), &result), "Failed to parse JSON output")

			tt.validate(t, result)
		})
	}
}

// Success Method Tests - Quiet Mode with GetID

func TestOutputFormatter_Success_Quiet_WithID(t *testing.T) {
	tests := []struct {
		name       string
		data       any
		wantOutput string
	}{
		{
			name:       "value receiver with ID",
			data:       mockDataWithID{ID: 42, Name: "Test"},
			wantOutput: "42",
		},
		{
			name:       "pointer receiver with ID",
			data:       &mockDataWithPointerID{ID: 99, Data: "Test"},
			wantOutput: "99",
		},
		{
			name:       "pointer to value receiver with ID",
			data:       &mockDataWithID{ID: 55, Name: "Pointer"},
			wantOutput: "55",
		},
		{
			name:       "ID is zero",
			data:       mockDataWithID{ID: 0, Name: "Zero"},
			wantOutput: "0",
		},
		{
			name:       "negative ID",
			data:       mockDataWithID{ID: -1, Name: "Negative"},
			wantOutput: "-1",
		},
		{
			name:       "large ID",
			data:       mockDataWithID{ID: 999999, Name: "Large"},
			wantOutput: "999999",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, err := os.Pipe()
			require.NoError(t, err, "Failed to create pipe for test")
			os.Stdout = w

			formatter := &OutputFormatter{JSON: false, Quiet: true}
			err = formatter.Success(tt.data)
			assert.NoError(t, err, "Success method should not return error")

			// Close writer before checking error or reading output
			_ = w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			_, _ = buf.ReadFrom(r)
			output := strings.TrimSpace(buf.String())

			assert.Equal(t, tt.wantOutput, output)
		})
	}
}

// Success Method Tests - Quiet Mode without GetID (falls through to prettyPrint)

func TestOutputFormatter_Success_Quiet_WithoutID(t *testing.T) {
	tests := []struct {
		name          string
		data          any
		shouldContain string
	}{
		{
			name:          "struct without GetID",
			data:          mockDataWithoutID{Name: "Test", Value: 42},
			shouldContain: "Test",
		},
		{
			name:          "map without GetID",
			data:          map[string]string{"test": "value"},
			shouldContain: "test",
		},
		{
			name:          "string",
			data:          "plain string output",
			shouldContain: "plain string output",
		},
		{
			name:          "integer",
			data:          42,
			shouldContain: "42",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, err := os.Pipe()
			require.NoError(t, err, "Failed to create pipe for test")
			os.Stdout = w

			formatter := &OutputFormatter{JSON: false, Quiet: true}
			err = formatter.Success(tt.data)
			assert.NoError(t, err, "Success method should not return error")

			_ = w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			_, _ = buf.ReadFrom(r)
			output := buf.String()

			// Should fall through to pretty print when no GetID method
			assert.Contains(t, output, tt.shouldContain)
		})
	}
}

// Success Method Tests - Human-Readable Mode

func TestOutputFormatter_Success_HumanReadable(t *testing.T) {
	tests := []struct {
		name          string
		data          any
		shouldContain string
	}{
		{
			name:          "struct with fields",
			data:          mockDataWithID{ID: 42, Name: "Test"},
			shouldContain: "42",
		},
		{
			name:          "map",
			data:          map[string]any{"key": "value", "num": 123},
			shouldContain: "key",
		},
		{
			name:          "string",
			data:          "human readable text",
			shouldContain: "human readable text",
		},
		{
			name:          "slice",
			data:          []string{"item1", "item2"},
			shouldContain: "item1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, err := os.Pipe()
			require.NoError(t, err, "Failed to create pipe for test")
			os.Stdout = w

			formatter := &OutputFormatter{JSON: false, Quiet: false}
			err = formatter.Success(tt.data)
			assert.NoError(t, err, "Success method should not return error")

			_ = w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			_, _ = buf.ReadFrom(r)
			output := buf.String()

			assert.Contains(t, output, tt.shouldContain)
		})
	}
}

// Error Method Tests - JSON Mode

func TestOutputFormatter_Error_JSON(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		message string
	}{
		{
			name:    "standard error",
			code:    "TEST_ERROR",
			message: "something went wrong",
		},
		{
			name:    "empty message",
			code:    "EMPTY_MSG",
			message: "",
		},
		{
			name:    "empty code",
			code:    "",
			message: "error without code",
		},
		{
			name:    "special characters in message",
			code:    "SPECIAL_CHAR",
			message: "error with \"quotes\" and \n newlines",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, err := os.Pipe()
			require.NoError(t, err, "Failed to create pipe for test")
			os.Stdout = w

			formatter := &OutputFormatter{JSON: true, Quiet: false}
			exitErr := formatter.Error(ExitError, tt.code, tt.message)
			assert.NotNil(t, exitErr, "Error method should return ExitErr")
			assert.Equal(t, ExitError, exitErr.Code)

			_ = w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			_, _ = buf.ReadFrom(r)
			output := buf.String()

			// Verify JSON output
			var result map[string]any
			require.NoError(t, json.Unmarshal([]byte(output), &result), "Failed to parse JSON output")

			assert.False(t, result["success"].(bool))

			errorData := result["error"].(map[string]any)
			assert.Equal(t, tt.code, errorData["code"])
			assert.Equal(t, tt.message, errorData["message"])

			// Verify no suggestion field when using Error() method
			_, hasSuggestion := errorData["suggestion"]
			assert.False(t, hasSuggestion)
		})
	}
}

// Error Method Tests - Quiet Mode

func TestOutputFormatter_Error_Quiet(t *testing.T) {
	// Capture stderr (should be empty in quiet mode)
	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	require.NoError(t, err, "Failed to create pipe for test")
	os.Stderr = w

	formatter := &OutputFormatter{JSON: false, Quiet: true}
	exitErr := formatter.Error(ExitError, "TEST_ERROR", "this should be suppressed")
	assert.NotNil(t, exitErr, "Error method should return ExitErr")
	assert.Equal(t, ExitError, exitErr.Code)

	_ = w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Quiet mode should suppress error output
	assert.Empty(t, output)
}

// Error Method Tests - Human-Readable Mode

func TestOutputFormatter_Error_HumanReadable(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		message string
	}{
		{
			name:    "standard error",
			code:    "TEST_ERROR",
			message: "something went wrong",
		},
		{
			name:    "long error message",
			code:    "LONG_ERROR",
			message: "this is a very long error message that explains what went wrong in detail",
		},
		{
			name:    "unicode in message",
			code:    "UNICODE_ERROR",
			message: "Error with unicode: 你好 мир",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stderr
			oldStderr := os.Stderr
			r, w, err := os.Pipe()
			require.NoError(t, err, "Failed to create pipe for test")
			os.Stderr = w

			formatter := &OutputFormatter{JSON: false, Quiet: false}
			exitErr := formatter.Error(ExitError, tt.code, tt.message)
			assert.NotNil(t, exitErr, "Error method should return ExitErr")
			assert.Equal(t, ExitError, exitErr.Code)

			_ = w.Close()
			os.Stderr = oldStderr

			var buf bytes.Buffer
			_, _ = buf.ReadFrom(r)
			output := buf.String()

			assert.Contains(t, output, tt.message)
			assert.Contains(t, output, "Error:")
			// Should NOT contain suggestion
			assert.NotContains(t, output, "Suggestion:")
		})
	}
}

// ErrorWithSuggestion Method Tests - JSON Mode

func TestOutputFormatter_ErrorWithSuggestion_JSON(t *testing.T) {
	tests := []struct {
		name       string
		code       string
		message    string
		suggestion string
		hasSuggest bool
	}{
		{
			name:       "with suggestion",
			code:       "TEST_ERROR",
			message:    "something went wrong",
			suggestion: "try this instead",
			hasSuggest: true,
		},
		{
			name:       "without suggestion",
			code:       "NO_SUGGEST",
			message:    "error without suggestion",
			suggestion: "",
			hasSuggest: false,
		},
		{
			name:       "empty code and message",
			code:       "",
			message:    "",
			suggestion: "helpful tip",
			hasSuggest: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, err := os.Pipe()
			require.NoError(t, err, "Failed to create pipe for test")
			os.Stdout = w

			formatter := &OutputFormatter{JSON: true, Quiet: false}
			exitErr := formatter.ErrorWithSuggestion(ExitError, tt.code, tt.message, tt.suggestion)
			assert.NotNil(t, exitErr, "ErrorWithSuggestion method should return ExitErr")
			assert.Equal(t, ExitError, exitErr.Code)

			_ = w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			_, _ = buf.ReadFrom(r)
			output := buf.String()

			// Verify JSON output
			var result map[string]any
			require.NoError(t, json.Unmarshal([]byte(output), &result), "Failed to parse JSON output")

			assert.False(t, result["success"].(bool))

			errorData := result["error"].(map[string]any)
			assert.Equal(t, tt.code, errorData["code"])
			assert.Equal(t, tt.message, errorData["message"])

			if tt.hasSuggest {
				assert.Equal(t, tt.suggestion, errorData["suggestion"])
			} else {
				_, exists := errorData["suggestion"]
				assert.False(t, exists)
			}
		})
	}
}

// ErrorWithSuggestion Method Tests - Quiet Mode

func TestOutputFormatter_ErrorWithSuggestion_Quiet(t *testing.T) {
	tests := []struct {
		name       string
		suggestion string
	}{
		{
			name:       "with suggestion",
			suggestion: "try this",
		},
		{
			name:       "without suggestion",
			suggestion: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stderr (should be empty in quiet mode)
			oldStderr := os.Stderr
			r, w, err := os.Pipe()
			require.NoError(t, err, "Failed to create pipe for test")
			os.Stderr = w

			formatter := &OutputFormatter{JSON: false, Quiet: true}
			exitErr := formatter.ErrorWithSuggestion(ExitError, "ERR", "message", tt.suggestion)
			assert.NotNil(t, exitErr, "ErrorWithSuggestion method should return ExitErr")
			assert.Equal(t, ExitError, exitErr.Code)

			_ = w.Close()
			os.Stderr = oldStderr

			var buf bytes.Buffer
			_, _ = buf.ReadFrom(r)
			output := buf.String()

			// Quiet mode should suppress error output
			assert.Empty(t, output)
		})
	}
}

// ErrorWithSuggestion Method Tests - Human-Readable Mode

func TestOutputFormatter_ErrorWithSuggestion_HumanReadable(t *testing.T) {
	tests := []struct {
		name             string
		code             string
		message          string
		suggestion       string
		shouldContain    []string
		shouldNotContain string
	}{
		{
			name:          "with suggestion",
			code:          "TEST_ERROR",
			message:       "something went wrong",
			suggestion:    "try this instead",
			shouldContain: []string{"something went wrong", "try this instead", "Error:", "Suggestion:"},
		},
		{
			name:             "without suggestion",
			code:             "NO_SUGGEST",
			message:          "error without suggestion",
			suggestion:       "",
			shouldContain:    []string{"error without suggestion", "Error:"},
			shouldNotContain: "Suggestion:",
		},
		{
			name:          "long suggestion",
			code:          "LONG",
			message:       "error",
			suggestion:    "this is a very detailed suggestion with lots of helpful information",
			shouldContain: []string{"error", "this is a very detailed suggestion", "Suggestion:"},
		},
		{
			name:          "unicode characters",
			code:          "UNICODE",
			message:       "Unicode error: 你好",
			suggestion:    "Try again: مرحبا",
			shouldContain: []string{"你好", "مرحبا"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stderr
			oldStderr := os.Stderr
			r, w, err := os.Pipe()
			require.NoError(t, err, "Failed to create pipe for test")
			os.Stderr = w

			formatter := &OutputFormatter{JSON: false, Quiet: false}
			exitErr := formatter.ErrorWithSuggestion(ExitError, tt.code, tt.message, tt.suggestion)
			assert.NotNil(t, exitErr, "ErrorWithSuggestion method should return ExitErr")
			assert.Equal(t, ExitError, exitErr.Code)

			// Close writer before checking error
			_ = w.Close()
			os.Stderr = oldStderr

			var buf bytes.Buffer
			_, _ = buf.ReadFrom(r)
			output := buf.String()

			for _, expected := range tt.shouldContain {
				assert.Contains(t, output, expected)
			}

			if tt.shouldNotContain != "" {
				assert.NotContains(t, output, tt.shouldNotContain)
			}
		})
	}
}

// Edge Cases and Integration Tests

func TestOutputFormatter_QuietModeGetIDPrecedence(t *testing.T) {
	// When Quiet is true and data has GetID(), it should output ID only
	// even if JSON is also true (Quiet check happens first)
	t.Run("quiet takes precedence over JSON when GetID exists", func(t *testing.T) {
		oldStdout := os.Stdout
		r, w, err := os.Pipe()
		require.NoError(t, err, "Failed to create pipe for test")
		os.Stdout = w

		formatter := &OutputFormatter{JSON: true, Quiet: true}
		data := mockDataWithID{ID: 42, Name: "Test"}
		err = formatter.Success(data)
		assert.NoError(t, err, "Success method should not return error")

		_ = w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		output := strings.TrimSpace(buf.String())

		// Should output just the ID, not JSON
		assert.Equal(t, "42", output)
	})

	// When Quiet is true but data has no GetID(), it falls through to JSON
	t.Run("quiet without GetID falls through to JSON", func(t *testing.T) {
		oldStdout := os.Stdout
		r, w, err := os.Pipe()
		require.NoError(t, err, "Failed to create pipe for test")
		os.Stdout = w

		formatter := &OutputFormatter{JSON: true, Quiet: true}
		data := mockDataWithoutID{Name: "Test", Value: 42}
		err = formatter.Success(data)
		assert.NoError(t, err, "Success method should not return error")

		_ = w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		output := buf.String()

		// Should output JSON
		var result map[string]any
		require.NoError(t, json.Unmarshal([]byte(output), &result), "Expected JSON output when Quiet=true without GetID()")
	})

	// JSON should take precedence for Error
	t.Run("jSON takes precedence over Quiet for Error", func(t *testing.T) {
		oldStdout := os.Stdout
		r, w, err := os.Pipe()
		require.NoError(t, err, "Failed to create pipe for test")
		os.Stdout = w

		formatter := &OutputFormatter{JSON: true, Quiet: true}
		exitErr := formatter.Error(ExitError, "TEST", "message")
		assert.NotNil(t, exitErr, "Error method should return ExitErr")
		assert.Equal(t, ExitError, exitErr.Code)

		_ = w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		output := buf.String()

		// Should output JSON
		var result map[string]any
		require.NoError(t, json.Unmarshal([]byte(output), &result), "Expected JSON output when JSON=true")
	})
}

func TestOutputFormatter_NilData(t *testing.T) {
	tests := []struct {
		name  string
		json  bool
		quiet bool
	}{
		{"JSON mode", true, false},
		{"Quiet mode", false, true},
		{"Human mode", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldStdout := os.Stdout
			r, w, err := os.Pipe()
			require.NoError(t, err, "Failed to create pipe for test")
			os.Stdout = w

			formatter := &OutputFormatter{JSON: tt.json, Quiet: tt.quiet}
			err = formatter.Success(nil)
			assert.NoError(t, err, "Success method should not return error with nil data")

			_ = w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			_, _ = buf.ReadFrom(r)
			output := buf.String()

			// Should handle nil gracefully - just verify no panic occurred
			if tt.json {
				var result map[string]any
				require.NoError(t, json.Unmarshal([]byte(output), &result), "Failed to parse JSON with nil data")
			}
		})
	}
}

func TestOutputFormatter_ErrorCallsErrorWithSuggestion(t *testing.T) {
	// Verify that Error() correctly calls ErrorWithSuggestion() with empty suggestion
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err, "Failed to create pipe for test")
	os.Stdout = w

	formatter := &OutputFormatter{JSON: true, Quiet: false}
	exitErr := formatter.Error(ExitError, "CODE", "message")
	assert.NotNil(t, exitErr, "Error method should return ExitErr")
	assert.Equal(t, ExitError, exitErr.Code)

	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	var result map[string]any
	require.NoError(t, json.Unmarshal([]byte(output), &result), "Failed to parse JSON output")

	errorData := result["error"].(map[string]any)
	// Verify no suggestion field is present
	_, exists := errorData["suggestion"]
	assert.False(t, exists)
}
