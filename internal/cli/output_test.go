package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

// ============================================================================
// OutputFormatter Tests
// ============================================================================

type mockDataWithID struct {
	ID   int
	Name string
}

func (m mockDataWithID) GetID() int {
	return m.ID
}

func TestOutputFormatter_Success_JSON(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	formatter := &OutputFormatter{JSON: true, Quiet: false}
	data := map[string]interface{}{"test": "value"}

	err := formatter.Success(data)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify JSON output
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if !result["success"].(bool) {
		t.Error("Expected success to be true")
	}
}

func TestOutputFormatter_Success_Quiet_WithID(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	formatter := &OutputFormatter{JSON: false, Quiet: true}
	data := mockDataWithID{ID: 42, Name: "Test"}

	err := formatter.Success(data)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := strings.TrimSpace(buf.String())

	if output != "42" {
		t.Errorf("Expected output '42', got '%s'", output)
	}
}

func TestOutputFormatter_Success_Quiet_WithoutID(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	formatter := &OutputFormatter{JSON: false, Quiet: true}
	data := map[string]string{"test": "value"}

	err := formatter.Success(data)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Should fall through to pretty print when no GetID method
	if !strings.Contains(output, "test") {
		t.Errorf("Expected output to contain 'test', got '%s'", output)
	}
}

func TestOutputFormatter_Error_JSON(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	formatter := &OutputFormatter{JSON: true, Quiet: false}

	err := formatter.Error("TEST_ERROR", "something went wrong")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify JSON output
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if result["success"].(bool) {
		t.Error("Expected success to be false")
	}

	errorData := result["error"].(map[string]interface{})
	if errorData["code"] != "TEST_ERROR" {
		t.Errorf("Expected error code 'TEST_ERROR', got '%s'", errorData["code"])
	}
	if errorData["message"] != "something went wrong" {
		t.Errorf("Expected message 'something went wrong', got '%s'", errorData["message"])
	}
}

func TestOutputFormatter_ErrorWithSuggestion_JSON(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	formatter := &OutputFormatter{JSON: true, Quiet: false}

	err := formatter.ErrorWithSuggestion("TEST_ERROR", "something went wrong", "try this instead")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify JSON output includes suggestion
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	errorData := result["error"].(map[string]interface{})
	if errorData["suggestion"] != "try this instead" {
		t.Errorf("Expected suggestion 'try this instead', got '%s'", errorData["suggestion"])
	}
}

func TestOutputFormatter_Error_HumanReadable(t *testing.T) {
	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	formatter := &OutputFormatter{JSON: false, Quiet: false}

	err := formatter.Error("TEST_ERROR", "something went wrong")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "something went wrong") {
		t.Errorf("Expected output to contain error message, got '%s'", output)
	}
}

func TestOutputFormatter_ErrorWithSuggestion_HumanReadable(t *testing.T) {
	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	formatter := &OutputFormatter{JSON: false, Quiet: false}

	err := formatter.ErrorWithSuggestion("TEST_ERROR", "something went wrong", "try this instead")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "something went wrong") {
		t.Errorf("Expected output to contain error message, got '%s'", output)
	}
	if !strings.Contains(output, "try this instead") {
		t.Errorf("Expected output to contain suggestion, got '%s'", output)
	}
}
