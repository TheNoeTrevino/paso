package handler

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
)

// ============================================================================
// Test Helpers
// ============================================================================

// createTestCommand creates a mock cobra.Command with specified flags
func createTestCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "test",
		Run: func(cmd *cobra.Command, args []string) {},
	}
	return cmd
}

// createTestParser creates a FlagParser with a test command and formatter
func createTestParser(cmd *cobra.Command) *FlagParser {
	formatter := &cli.OutputFormatter{JSON: false, Quiet: false}
	return NewFlagParser(cmd, formatter)
}

// ============================================================================
// ParseTaskID Tests
// ============================================================================

func TestParseTaskID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		flagValue int
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "valid task ID",
			flagValue: 42,
			wantErr:   false,
		},
		{
			name:      "valid task ID = 1",
			flagValue: 1,
			wantErr:   false,
		},
		{
			name:      "zero task ID",
			flagValue: 0,
			wantErr:   true,
			errMsg:    "must be greater than 0",
		},
		{
			name:      "negative task ID",
			flagValue: -1,
			wantErr:   true,
			errMsg:    "must be greater than 0",
		},
		{
			name:      "large task ID",
			flagValue: 999999,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cmd := createTestCommand()
			cmd.Flags().Int("id", tt.flagValue, "task id")

			parser := createTestParser(cmd)

			result, err := parser.ParseTaskID("id")

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing '%s', got nil", tt.errMsg)
				} else if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error containing '%s', got '%s'", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result != tt.flagValue {
					t.Errorf("expected %d, got %d", tt.flagValue, result)
				}
			}
		})
	}
}

// ============================================================================
// ParseColumnID Tests
// ============================================================================

func TestParseColumnID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		flagValue int
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "valid column ID",
			flagValue: 5,
			wantErr:   false,
		},
		{
			name:      "valid column ID = 1",
			flagValue: 1,
			wantErr:   false,
		},
		{
			name:      "zero column ID",
			flagValue: 0,
			wantErr:   true,
			errMsg:    "must be greater than 0",
		},
		{
			name:      "negative column ID",
			flagValue: -10,
			wantErr:   true,
			errMsg:    "must be greater than 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cmd := createTestCommand()
			cmd.Flags().Int("column-id", tt.flagValue, "column id")

			parser := createTestParser(cmd)

			result, err := parser.ParseColumnID("column-id")

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing '%s', got nil", tt.errMsg)
				} else if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error containing '%s', got '%s'", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result != tt.flagValue {
					t.Errorf("expected %d, got %d", tt.flagValue, result)
				}
			}
		})
	}
}

// ============================================================================
// ParseLabelID Tests
// ============================================================================

func TestParseLabelID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		flagValue int
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "valid label ID",
			flagValue: 3,
			wantErr:   false,
		},
		{
			name:      "valid label ID = 1",
			flagValue: 1,
			wantErr:   false,
		},
		{
			name:      "zero label ID",
			flagValue: 0,
			wantErr:   true,
			errMsg:    "must be greater than 0",
		},
		{
			name:      "negative label ID",
			flagValue: -5,
			wantErr:   true,
			errMsg:    "must be greater than 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cmd := createTestCommand()
			cmd.Flags().Int("label-id", tt.flagValue, "label id")

			parser := createTestParser(cmd)

			result, err := parser.ParseLabelID("label-id")

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing '%s', got nil", tt.errMsg)
				} else if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error containing '%s', got '%s'", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result != tt.flagValue {
					t.Errorf("expected %d, got %d", tt.flagValue, result)
				}
			}
		})
	}
}

// ============================================================================
// ParseString Tests
// ============================================================================

func TestParseString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		flagValue string
		wantValue string
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "valid string",
			flagValue: "test-value",
			wantValue: "test-value",
			wantErr:   false,
		},
		{
			name:      "string with whitespace trimmed",
			flagValue: "  trimmed  ",
			wantValue: "trimmed",
			wantErr:   false,
		},
		{
			name:      "empty string",
			flagValue: "",
			wantErr:   true,
			errMsg:    "is required",
		},
		{
			name:      "whitespace only string",
			flagValue: "   ",
			wantErr:   true,
			errMsg:    "is required",
		},
		{
			name:      "string with special characters",
			flagValue: "test@#$%",
			wantValue: "test@#$%",
			wantErr:   false,
		},
		{
			name:      "string with newlines trimmed",
			flagValue: "\nvalue\n",
			wantValue: "value",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cmd := createTestCommand()
			cmd.Flags().String("name", "", "name flag")
			_ = cmd.Flags().Set("name", tt.flagValue)

			parser := createTestParser(cmd)

			result, err := parser.ParseString("name")

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing '%s', got nil", tt.errMsg)
				} else if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error containing '%s', got '%s'", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result != tt.wantValue {
					t.Errorf("expected '%s', got '%s'", tt.wantValue, result)
				}
			}
		})
	}
}

// ============================================================================
// ParseStringOptional Tests
// ============================================================================

func TestParseStringOptional(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		flagValue string
		wantValue string
		wantErr   bool
	}{
		{
			name:      "valid string",
			flagValue: "optional-value",
			wantValue: "optional-value",
			wantErr:   false,
		},
		{
			name:      "empty string allowed",
			flagValue: "",
			wantValue: "",
			wantErr:   false,
		},
		{
			name:      "whitespace preserved",
			flagValue: "  spaces  ",
			wantValue: "  spaces  ",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cmd := createTestCommand()
			cmd.Flags().String("description", "", "description flag")
			_ = cmd.Flags().Set("description", tt.flagValue)

			parser := createTestParser(cmd)

			result, err := parser.ParseStringOptional("description")

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result != tt.wantValue {
					t.Errorf("expected '%s', got '%s'", tt.wantValue, result)
				}
			}
		})
	}
}

// ============================================================================
// ParseInt Tests
// ============================================================================

func TestParseInt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		flagValue int
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "valid int",
			flagValue: 10,
			wantErr:   false,
		},
		{
			name:      "valid int = 1",
			flagValue: 1,
			wantErr:   false,
		},
		{
			name:      "zero value",
			flagValue: 0,
			wantErr:   true,
			errMsg:    "must be greater than 0",
		},
		{
			name:      "negative value",
			flagValue: -1,
			wantErr:   true,
			errMsg:    "must be greater than 0",
		},
		{
			name:      "large value",
			flagValue: 1000000,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cmd := createTestCommand()
			cmd.Flags().Int("count", tt.flagValue, "count flag")

			parser := createTestParser(cmd)

			result, err := parser.ParseInt("count")

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing '%s', got nil", tt.errMsg)
				} else if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error containing '%s', got '%s'", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result != tt.flagValue {
					t.Errorf("expected %d, got %d", tt.flagValue, result)
				}
			}
		})
	}
}

// ============================================================================
// ParseIntOptional Tests
// ============================================================================

func TestParseIntOptional(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		flagValue int
		wantValue int
		wantErr   bool
	}{
		{
			name:      "positive int",
			flagValue: 42,
			wantValue: 42,
			wantErr:   false,
		},
		{
			name:      "zero value allowed",
			flagValue: 0,
			wantValue: 0,
			wantErr:   false,
		},
		{
			name:      "negative value allowed",
			flagValue: -10,
			wantValue: -10,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cmd := createTestCommand()
			cmd.Flags().Int("offset", tt.flagValue, "offset flag")

			parser := createTestParser(cmd)

			result, err := parser.ParseIntOptional("offset")

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result != tt.wantValue {
					t.Errorf("expected %d, got %d", tt.wantValue, result)
				}
			}
		})
	}
}

// ============================================================================
// ParseBool Tests
// ============================================================================

func TestParseBool(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		flagValue bool
		wantValue bool
		wantErr   bool
	}{
		{
			name:      "true value",
			flagValue: true,
			wantValue: true,
			wantErr:   false,
		},
		{
			name:      "false value",
			flagValue: false,
			wantValue: false,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cmd := createTestCommand()
			cmd.Flags().Bool("enabled", tt.flagValue, "enabled flag")

			parser := createTestParser(cmd)

			result, err := parser.ParseBool("enabled")

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result != tt.wantValue {
					t.Errorf("expected %v, got %v", tt.wantValue, result)
				}
			}
		})
	}
}

// ============================================================================
// ParseColor Tests
// ============================================================================

func TestParseColor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		flagValue string
		wantValue string
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "valid red color",
			flagValue: "#FF0000",
			wantValue: "#FF0000",
			wantErr:   false,
		},
		{
			name:      "valid blue color",
			flagValue: "#0000FF",
			wantValue: "#0000FF",
			wantErr:   false,
		},
		{
			name:      "valid lowercase color",
			flagValue: "#ff5733",
			wantValue: "#ff5733",
			wantErr:   false,
		},
		{
			name:      "valid mixed case color",
			flagValue: "#AbCdEf",
			wantValue: "#AbCdEf",
			wantErr:   false,
		},
		{
			name:      "missing # prefix",
			flagValue: "FF0000",
			wantErr:   true,
			errMsg:    "must be in hex format",
		},
		{
			name:      "too short",
			flagValue: "#FFF",
			wantErr:   true,
			errMsg:    "must be in hex format",
		},
		{
			name:      "too long",
			flagValue: "#FF00000",
			wantErr:   true,
			errMsg:    "must be in hex format",
		},
		{
			name:      "invalid hex characters",
			flagValue: "#GGGGGG",
			wantErr:   true,
			errMsg:    "must be in hex format",
		},
		{
			name:      "empty string",
			flagValue: "",
			wantErr:   true,
			errMsg:    "must be in hex format",
		},
		{
			name:      "only # symbol",
			flagValue: "#",
			wantErr:   true,
			errMsg:    "must be in hex format",
		},
		{
			name:      "contains space",
			flagValue: "#FF 000",
			wantErr:   true,
			errMsg:    "must be in hex format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cmd := createTestCommand()
			cmd.Flags().String("color", "", "color flag")
			_ = cmd.Flags().Set("color", tt.flagValue)

			parser := createTestParser(cmd)

			result, err := parser.ParseColor("color")

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing '%s', got nil", tt.errMsg)
				} else if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error containing '%s', got '%s'", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result != tt.wantValue {
					t.Errorf("expected '%s', got '%s'", tt.wantValue, result)
				}
			}
		})
	}
}

// ============================================================================
// OutputFormats Tests
// ============================================================================

func TestOutputFormats(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		jsonFlag  bool
		quietFlag bool
		wantJSON  bool
		wantQuiet bool
		wantErr   bool
	}{
		{
			name:      "both false",
			jsonFlag:  false,
			quietFlag: false,
			wantJSON:  false,
			wantQuiet: false,
			wantErr:   false,
		},
		{
			name:      "json true",
			jsonFlag:  true,
			quietFlag: false,
			wantJSON:  true,
			wantQuiet: false,
			wantErr:   false,
		},
		{
			name:      "quiet true",
			jsonFlag:  false,
			quietFlag: true,
			wantJSON:  false,
			wantQuiet: true,
			wantErr:   false,
		},
		{
			name:      "both true",
			jsonFlag:  true,
			quietFlag: true,
			wantJSON:  true,
			wantQuiet: true,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cmd := createTestCommand()
			cmd.Flags().Bool("json", tt.jsonFlag, "json output")
			cmd.Flags().Bool("quiet", tt.quietFlag, "quiet mode")

			parser := createTestParser(cmd)

			jsonOutput, quietMode, err := parser.OutputFormats()

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if jsonOutput != tt.wantJSON {
					t.Errorf("expected json=%v, got %v", tt.wantJSON, jsonOutput)
				}
				if quietMode != tt.wantQuiet {
					t.Errorf("expected quiet=%v, got %v", tt.wantQuiet, quietMode)
				}
			}
		})
	}
}

// ============================================================================
// OutputFormats Error Tests
// ============================================================================

func TestOutputFormats_MissingFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupFlags  func(*cobra.Command)
		wantErr     bool
		errContains string
	}{
		{
			name: "missing json flag",
			setupFlags: func(cmd *cobra.Command) {
				cmd.Flags().Bool("quiet", false, "quiet mode")
			},
			wantErr:     true,
			errContains: "json",
		},
		{
			name: "missing quiet flag",
			setupFlags: func(cmd *cobra.Command) {
				cmd.Flags().Bool("json", false, "json output")
			},
			wantErr:     true,
			errContains: "quiet",
		},
		{
			name: "both flags missing",
			setupFlags: func(cmd *cobra.Command) {
				// Don't add any flags
			},
			wantErr:     true,
			errContains: "json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cmd := createTestCommand()
			tt.setupFlags(cmd)

			parser := createTestParser(cmd)

			_, _, err := parser.OutputFormats()

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing '%s', got nil", tt.errContains)
				} else if !contains(err.Error(), tt.errContains) {
					t.Errorf("expected error containing '%s', got '%s'", tt.errContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// ============================================================================
// NewFlagParser Tests
// ============================================================================

func TestNewFlagParser(t *testing.T) {
	t.Parallel()

	cmd := createTestCommand()
	formatter := &cli.OutputFormatter{JSON: false, Quiet: false}

	parser := NewFlagParser(cmd, formatter)

	if parser == nil {
		t.Fatal("expected non-nil parser")
	}

	if parser.cmd != cmd {
		t.Error("parser cmd field not set correctly")
	}

	if parser.formatter != formatter {
		t.Error("parser formatter field not set correctly")
	}
}

// ============================================================================
// Edge Case Tests
// ============================================================================

func TestParseString_NonExistentFlag(t *testing.T) {
	t.Parallel()

	cmd := createTestCommand()
	parser := createTestParser(cmd)

	_, err := parser.ParseString("non-existent")

	if err == nil {
		t.Error("expected error for non-existent flag, got nil")
	}
}

func TestParseInt_NonExistentFlag(t *testing.T) {
	t.Parallel()

	cmd := createTestCommand()
	parser := createTestParser(cmd)

	_, err := parser.ParseInt("non-existent")

	if err == nil {
		t.Error("expected error for non-existent flag, got nil")
	}
}

func TestParseBool_NonExistentFlag(t *testing.T) {
	t.Parallel()

	cmd := createTestCommand()
	parser := createTestParser(cmd)

	_, err := parser.ParseBool("non-existent")

	if err == nil {
		t.Error("expected error for non-existent flag, got nil")
	}
}

// ============================================================================
// Utility Functions
// ============================================================================

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
