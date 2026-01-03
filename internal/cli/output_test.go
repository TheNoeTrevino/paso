package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

// ============================================================================
// Mock Types for Testing
// ============================================================================

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

// ============================================================================
// Success Method Tests - JSON Mode
// ============================================================================

func TestOutputFormatter_Success_JSON(t *testing.T) {
	tests := []struct {
		name     string
		data     interface{}
		validate func(t *testing.T, result map[string]interface{})
	}{
		{
			name: "map data",
			data: map[string]interface{}{"test": "value", "number": float64(42)},
			validate: func(t *testing.T, result map[string]interface{}) {
				if !result["success"].(bool) {
					t.Error("Expected success to be true")
				}
				dataMap := result["data"].(map[string]interface{})
				if dataMap["test"] != "value" {
					t.Errorf("Expected data.test to be 'value', got %v", dataMap["test"])
				}
			},
		},
		{
			name: "struct with ID",
			data: mockDataWithID{ID: 123, Name: "Test"},
			validate: func(t *testing.T, result map[string]interface{}) {
				if !result["success"].(bool) {
					t.Error("Expected success to be true")
				}
				dataMap := result["data"].(map[string]interface{})
				if dataMap["Name"] != "Test" {
					t.Errorf("Expected data.Name to be 'Test', got %v", dataMap["Name"])
				}
			},
		},
		{
			name: "string data",
			data: "simple string",
			validate: func(t *testing.T, result map[string]interface{}) {
				if !result["success"].(bool) {
					t.Error("Expected success to be true")
				}
				if result["data"] != "simple string" {
					t.Errorf("Expected data to be 'simple string', got %v", result["data"])
				}
			},
		},
		{
			name: "integer data",
			data: 42,
			validate: func(t *testing.T, result map[string]interface{}) {
				if !result["success"].(bool) {
					t.Error("Expected success to be true")
				}
				// JSON unmarshals numbers as float64
				if result["data"].(float64) != 42 {
					t.Errorf("Expected data to be 42, got %v", result["data"])
				}
			},
		},
		{
			name: "nil data",
			data: nil,
			validate: func(t *testing.T, result map[string]interface{}) {
				if !result["success"].(bool) {
					t.Error("Expected success to be true")
				}
				if result["data"] != nil {
					t.Errorf("Expected data to be nil, got %v", result["data"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			formatter := &OutputFormatter{JSON: true, Quiet: false}
			err := formatter.Success(tt.data)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			_ = w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			_, _ = buf.ReadFrom(r)
			output := buf.String()

			// Verify JSON output
			var result map[string]interface{}
			if err := json.Unmarshal([]byte(output), &result); err != nil {
				t.Fatalf("Failed to parse JSON: %v\nOutput: %s", err, output)
			}

			tt.validate(t, result)
		})
	}
}

// ============================================================================
// Success Method Tests - Quiet Mode with GetID
// ============================================================================

func TestOutputFormatter_Success_Quiet_WithID(t *testing.T) {
	tests := []struct {
		name       string
		data       interface{}
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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			formatter := &OutputFormatter{JSON: false, Quiet: true}
			err := formatter.Success(tt.data)

			// Close writer before checking error or reading output
			_ = w.Close()
			os.Stdout = oldStdout

			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			var buf bytes.Buffer
			_, _ = buf.ReadFrom(r)
			output := strings.TrimSpace(buf.String())

			if output != tt.wantOutput {
				t.Errorf("Expected output '%s', got '%s'", tt.wantOutput, output)
			}
		})
	}
}

// ============================================================================
// Success Method Tests - Quiet Mode without GetID (falls through to prettyPrint)
// ============================================================================

func TestOutputFormatter_Success_Quiet_WithoutID(t *testing.T) {
	tests := []struct {
		name          string
		data          interface{}
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
			r, w, _ := os.Pipe()
			os.Stdout = w

			formatter := &OutputFormatter{JSON: false, Quiet: true}
			err := formatter.Success(tt.data)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			_ = w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			_, _ = buf.ReadFrom(r)
			output := buf.String()

			// Should fall through to pretty print when no GetID method
			if !strings.Contains(output, tt.shouldContain) {
				t.Errorf("Expected output to contain '%s', got '%s'", tt.shouldContain, output)
			}
		})
	}
}

// ============================================================================
// Success Method Tests - Human-Readable Mode
// ============================================================================

func TestOutputFormatter_Success_HumanReadable(t *testing.T) {
	tests := []struct {
		name          string
		data          interface{}
		shouldContain string
	}{
		{
			name:          "struct with fields",
			data:          mockDataWithID{ID: 42, Name: "Test"},
			shouldContain: "42",
		},
		{
			name:          "map",
			data:          map[string]interface{}{"key": "value", "num": 123},
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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			formatter := &OutputFormatter{JSON: false, Quiet: false}
			err := formatter.Success(tt.data)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			_ = w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			_, _ = buf.ReadFrom(r)
			output := buf.String()

			if !strings.Contains(output, tt.shouldContain) {
				t.Errorf("Expected output to contain '%s', got '%s'", tt.shouldContain, output)
			}
		})
	}
}

// ============================================================================
// Error Method Tests - JSON Mode
// ============================================================================

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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			formatter := &OutputFormatter{JSON: true, Quiet: false}
			err := formatter.Error(tt.code, tt.message)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			_ = w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			_, _ = buf.ReadFrom(r)
			output := buf.String()

			// Verify JSON output
			var result map[string]interface{}
			if err := json.Unmarshal([]byte(output), &result); err != nil {
				t.Fatalf("Failed to parse JSON: %v\nOutput: %s", err, output)
			}

			if result["success"].(bool) {
				t.Error("Expected success to be false")
			}

			errorData := result["error"].(map[string]interface{})
			if errorData["code"] != tt.code {
				t.Errorf("Expected error code '%s', got '%s'", tt.code, errorData["code"])
			}
			if errorData["message"] != tt.message {
				t.Errorf("Expected message '%s', got '%s'", tt.message, errorData["message"])
			}

			// Verify no suggestion field when using Error() method
			if _, hasSuggestion := errorData["suggestion"]; hasSuggestion {
				t.Error("Expected no suggestion field in Error() output")
			}
		})
	}
}

// ============================================================================
// Error Method Tests - Quiet Mode
// ============================================================================

func TestOutputFormatter_Error_Quiet(t *testing.T) {
	// Capture stderr (should be empty in quiet mode)
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	formatter := &OutputFormatter{JSON: false, Quiet: true}
	err := formatter.Error("TEST_ERROR", "this should be suppressed")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	_ = w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Quiet mode should suppress error output
	if output != "" {
		t.Errorf("Expected no output in quiet mode, got '%s'", output)
	}
}

// ============================================================================
// Error Method Tests - Human-Readable Mode
// ============================================================================

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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			// Capture stderr
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			formatter := &OutputFormatter{JSON: false, Quiet: false}
			err := formatter.Error(tt.code, tt.message)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			_ = w.Close()
			os.Stderr = oldStderr

			var buf bytes.Buffer
			_, _ = buf.ReadFrom(r)
			output := buf.String()

			if !strings.Contains(output, tt.message) {
				t.Errorf("Expected output to contain error message '%s', got '%s'", tt.message, output)
			}
			if !strings.Contains(output, "Error:") {
				t.Errorf("Expected output to contain 'Error:', got '%s'", output)
			}
			// Should NOT contain suggestion
			if strings.Contains(output, "Suggestion:") {
				t.Errorf("Expected no suggestion in Error() output, got '%s'", output)
			}
		})
	}
}

// ============================================================================
// ErrorWithSuggestion Method Tests - JSON Mode
// ============================================================================

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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			formatter := &OutputFormatter{JSON: true, Quiet: false}
			err := formatter.ErrorWithSuggestion(tt.code, tt.message, tt.suggestion)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			_ = w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			_, _ = buf.ReadFrom(r)
			output := buf.String()

			// Verify JSON output
			var result map[string]interface{}
			if err := json.Unmarshal([]byte(output), &result); err != nil {
				t.Fatalf("Failed to parse JSON: %v\nOutput: %s", err, output)
			}

			if result["success"].(bool) {
				t.Error("Expected success to be false")
			}

			errorData := result["error"].(map[string]interface{})
			if errorData["code"] != tt.code {
				t.Errorf("Expected error code '%s', got '%s'", tt.code, errorData["code"])
			}
			if errorData["message"] != tt.message {
				t.Errorf("Expected message '%s', got '%s'", tt.message, errorData["message"])
			}

			if tt.hasSuggest {
				if errorData["suggestion"] != tt.suggestion {
					t.Errorf("Expected suggestion '%s', got '%v'", tt.suggestion, errorData["suggestion"])
				}
			} else {
				if _, exists := errorData["suggestion"]; exists {
					t.Error("Expected no suggestion field when suggestion is empty")
				}
			}
		})
	}
}

// ============================================================================
// ErrorWithSuggestion Method Tests - Quiet Mode
// ============================================================================

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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			// Capture stderr (should be empty in quiet mode)
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			formatter := &OutputFormatter{JSON: false, Quiet: true}
			err := formatter.ErrorWithSuggestion("ERR", "message", tt.suggestion)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			_ = w.Close()
			os.Stderr = oldStderr

			var buf bytes.Buffer
			_, _ = buf.ReadFrom(r)
			output := buf.String()

			// Quiet mode should suppress error output
			if output != "" {
				t.Errorf("Expected no output in quiet mode, got '%s'", output)
			}
		})
	}
}

// ============================================================================
// ErrorWithSuggestion Method Tests - Human-Readable Mode
// ============================================================================

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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			// Capture stderr
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			formatter := &OutputFormatter{JSON: false, Quiet: false}
			err := formatter.ErrorWithSuggestion(tt.code, tt.message, tt.suggestion)

			// Close writer before checking error
			_ = w.Close()
			os.Stderr = oldStderr

			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			var buf bytes.Buffer
			_, _ = buf.ReadFrom(r)
			output := buf.String()

			for _, expected := range tt.shouldContain {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain '%s', got '%s'", expected, output)
				}
			}

			if tt.shouldNotContain != "" && strings.Contains(output, tt.shouldNotContain) {
				t.Errorf("Expected output to NOT contain '%s', got '%s'", tt.shouldNotContain, output)
			}
		})
	}
}

// ============================================================================
// Edge Cases and Integration Tests
// ============================================================================

func TestOutputFormatter_QuietModeGetIDPrecedence(t *testing.T) {
	// When Quiet is true and data has GetID(), it should output ID only
	// even if JSON is also true (Quiet check happens first)
	t.Run("Quiet takes precedence over JSON when GetID exists", func(t *testing.T) {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		formatter := &OutputFormatter{JSON: true, Quiet: true}
		data := mockDataWithID{ID: 42, Name: "Test"}
		err := formatter.Success(data)

		_ = w.Close()
		os.Stdout = oldStdout

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		output := strings.TrimSpace(buf.String())

		// Should output just the ID, not JSON
		if output != "42" {
			t.Errorf("Expected output '42' when Quiet=true with GetID(), got: %s", output)
		}
	})

	// When Quiet is true but data has no GetID(), it falls through to JSON
	t.Run("Quiet without GetID falls through to JSON", func(t *testing.T) {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		formatter := &OutputFormatter{JSON: true, Quiet: true}
		data := mockDataWithoutID{Name: "Test", Value: 42}
		err := formatter.Success(data)

		_ = w.Close()
		os.Stdout = oldStdout

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		output := buf.String()

		// Should output JSON
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(output), &result); err != nil {
			t.Fatalf("Expected JSON output when Quiet=true without GetID(), got: %s", output)
		}
	})

	// JSON should take precedence for Error
	t.Run("JSON takes precedence over Quiet for Error", func(t *testing.T) {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		formatter := &OutputFormatter{JSON: true, Quiet: true}
		err := formatter.Error("TEST", "message")

		_ = w.Close()
		os.Stdout = oldStdout

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		output := buf.String()

		// Should output JSON
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(output), &result); err != nil {
			t.Fatalf("Expected JSON output when JSON=true, got: %s", output)
		}
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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			formatter := &OutputFormatter{JSON: tt.json, Quiet: tt.quiet}
			err := formatter.Success(nil)
			if err != nil {
				t.Errorf("Expected no error with nil data, got %v", err)
			}

			_ = w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			_, _ = buf.ReadFrom(r)
			output := buf.String()

			// Should handle nil gracefully - just verify no panic occurred
			if tt.json {
				var result map[string]interface{}
				if err := json.Unmarshal([]byte(output), &result); err != nil {
					t.Fatalf("Failed to parse JSON with nil data: %v", err)
				}
			}
		})
	}
}

func TestOutputFormatter_ErrorCallsErrorWithSuggestion(t *testing.T) {
	// Verify that Error() correctly calls ErrorWithSuggestion() with empty suggestion
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	formatter := &OutputFormatter{JSON: true, Quiet: false}
	err := formatter.Error("CODE", "message")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	errorData := result["error"].(map[string]interface{})
	// Verify no suggestion field is present
	if _, exists := errorData["suggestion"]; exists {
		t.Error("Expected no suggestion field when calling Error()")
	}
}
