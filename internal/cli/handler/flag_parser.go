// Package handler provides flag parsing utilities
package handler

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
)

// FlagParser provides common flag extraction patterns
type FlagParser struct {
	cmd       *cobra.Command
	formatter *cli.OutputFormatter
}

// NewFlagParser creates a new flag parser
func NewFlagParser(cmd *cobra.Command, formatter *cli.OutputFormatter) *FlagParser {
	return &FlagParser{
		cmd:       cmd,
		formatter: formatter,
	}
}

// ParseProjectID extracts project ID from --project flag or environment variable
// Returns the project ID or exits on error
func (p *FlagParser) ParseProjectID() (int, error) {
	projectID, err := cli.GetProjectID(p.cmd)
	if err != nil {
		if fmtErr := p.formatter.ErrorWithSuggestion("NO_PROJECT",
			err.Error(),
			"Set project with: eval $(paso use project <project-id>)"); fmtErr != nil {
			fmt.Fprintf(os.Stderr, "Error formatting error message: %v\n", fmtErr)
		}
		os.Exit(cli.ExitUsage)
	}
	return projectID, nil
}

// ParseTaskID extracts task ID from a flag
func (p *FlagParser) ParseTaskID(flagName string) (int, error) {
	taskID, err := p.cmd.Flags().GetInt(flagName)
	if err != nil {
		return 0, fmt.Errorf("failed to parse %s flag: %w", flagName, err)
	}
	if taskID <= 0 {
		return 0, fmt.Errorf("%s must be greater than 0", flagName)
	}
	return taskID, nil
}

// ParseColumnID extracts column ID from a flag
func (p *FlagParser) ParseColumnID(flagName string) (int, error) {
	columnID, err := p.cmd.Flags().GetInt(flagName)
	if err != nil {
		return 0, fmt.Errorf("failed to parse %s flag: %w", flagName, err)
	}
	if columnID <= 0 {
		return 0, fmt.Errorf("%s must be greater than 0", flagName)
	}
	return columnID, nil
}

// ParseLabelID extracts label ID from a flag
func (p *FlagParser) ParseLabelID(flagName string) (int, error) {
	labelID, err := p.cmd.Flags().GetInt(flagName)
	if err != nil {
		return 0, fmt.Errorf("failed to parse %s flag: %w", flagName, err)
	}
	if labelID <= 0 {
		return 0, fmt.Errorf("%s must be greater than 0", flagName)
	}
	return labelID, nil
}

// ParseString extracts a required string flag
func (p *FlagParser) ParseString(flagName string) (string, error) {
	value, err := p.cmd.Flags().GetString(flagName)
	if err != nil {
		return "", fmt.Errorf("failed to parse %s flag: %w", flagName, err)
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("%s is required", flagName)
	}
	return value, nil
}

// ParseStringOptional extracts an optional string flag
func (p *FlagParser) ParseStringOptional(flagName string) (string, error) {
	return p.cmd.Flags().GetString(flagName)
}

// ParseInt extracts a required int flag
func (p *FlagParser) ParseInt(flagName string) (int, error) {
	value, err := p.cmd.Flags().GetInt(flagName)
	if err != nil {
		return 0, fmt.Errorf("failed to parse %s flag: %w", flagName, err)
	}
	if value <= 0 {
		return 0, fmt.Errorf("%s must be greater than 0", flagName)
	}
	return value, nil
}

// ParseIntOptional extracts an optional int flag
func (p *FlagParser) ParseIntOptional(flagName string) (int, error) {
	return p.cmd.Flags().GetInt(flagName)
}

// ParseBool extracts a boolean flag
func (p *FlagParser) ParseBool(flagName string) (bool, error) {
	return p.cmd.Flags().GetBool(flagName)
}

// ParseColor extracts and validates a color flag
func (p *FlagParser) ParseColor(flagName string) (string, error) {
	color, err := p.cmd.Flags().GetString(flagName)
	if err != nil {
		return "", fmt.Errorf("failed to parse %s flag: %w", flagName, err)
	}
	if err := cli.ValidateColorHex(color); err != nil {
		return "", err
	}
	return color, nil
}

// OutputFormats extracts JSON and Quiet output flags
func (p *FlagParser) OutputFormats() (jsonOutput bool, quietMode bool, err error) {
	jsonOutput, err = p.cmd.Flags().GetBool("json")
	if err != nil {
		return false, false, fmt.Errorf("failed to parse json flag: %w", err)
	}

	quietMode, err = p.cmd.Flags().GetBool("quiet")
	if err != nil {
		return false, false, fmt.Errorf("failed to parse quiet flag: %w", err)
	}

	return jsonOutput, quietMode, nil
}
