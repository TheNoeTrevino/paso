// Package handler provides command execution abstraction to reduce boilerplate
package handler

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/thenoetrevino/paso/internal/cli"
)

// Handler defines the interface for command execution
type Handler interface {
	// Execute runs the command with parsed arguments
	Execute(ctx context.Context, args *Arguments) (any, error)
}

// Arguments captures parsed CLI arguments and flags
type Arguments struct {
	Flags map[string]any
	Args  []string
	cmd   *cobra.Command
}

// GetCmd returns the cobra command for access to flag parsing utilities
func (a *Arguments) GetCmd() *cobra.Command {
	return a.cmd
}

// CommandConfig holds configuration for command execution
type CommandConfig struct {
	// Formatter for output
	Formatter *cli.OutputFormatter
}

// Command wraps common command execution logic
// Returns a cobra RunE compatible function
func Command(handler Handler, parseFlags func(*cobra.Command) error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		// Parse flags
		if err := parseFlags(cmd); err != nil {
			return err
		}

		// Get formatter from flags
		jsonOutput, _ := cmd.Flags().GetBool("json")
		quietMode, _ := cmd.Flags().GetBool("quiet")
		formatter := &cli.OutputFormatter{JSON: jsonOutput, Quiet: quietMode}

		// Build arguments map from all flags
		arguments := &Arguments{
			Flags: parseFlagsToMap(cmd),
			Args:  args,
			cmd:   cmd,
		}

		// Execute handler
		result, err := handler.Execute(ctx, arguments)
		if err != nil {
			return err
		}

		// Common output formatting
		return formatter.Success(result)
	}
}

// SimpleCommand wraps command execution with minimal setup
// Use this for commands that don't need complex flag parsing
func SimpleCommand(handler Handler) func(*cobra.Command, []string) error {
	return Command(handler, func(cmd *cobra.Command) error {
		return nil
	})
}

// parseFlagsToMap converts cobra command flags to a map
func parseFlagsToMap(cmd *cobra.Command) map[string]any {
	flags := make(map[string]any)

	// Visit all flags that were explicitly set
	cmd.Flags().Visit(func(f *pflag.Flag) {
		// Get the value based on flag type
		switch f.Value.Type() {
		case "string":
			if v, err := cmd.Flags().GetString(f.Name); err == nil {
				flags[f.Name] = v
			}
		case "int":
			if v, err := cmd.Flags().GetInt(f.Name); err == nil {
				flags[f.Name] = v
			}
		case "int64":
			if v, err := cmd.Flags().GetInt64(f.Name); err == nil {
				flags[f.Name] = v
			}
		case "bool":
			if v, err := cmd.Flags().GetBool(f.Name); err == nil {
				flags[f.Name] = v
			}
		case "float64":
			if v, err := cmd.Flags().GetFloat64(f.Name); err == nil {
				flags[f.Name] = v
			}
		case "stringSlice":
			if v, err := cmd.Flags().GetStringSlice(f.Name); err == nil {
				flags[f.Name] = v
			}
		case "intSlice":
			if v, err := cmd.Flags().GetIntSlice(f.Name); err == nil {
				flags[f.Name] = v
			}
		default:
			slog.Debug("unsupported flag type", "flag", f.Name, "type", f.Value.Type())
		}
	})

	return flags
}

// MustGetString retrieves a string flag and panics if it doesn't exist
func (a *Arguments) MustGetString(name string) string {
	v, ok := a.Flags[name]
	if !ok {
		slog.Error("flag not found", "flag", name)
		os.Exit(cli.ExitValidation)
	}
	val, ok := v.(string)
	if !ok {
		slog.Error("flag type assertion failed", "flag", name, "expected", "string", "got", fmt.Sprintf("%T", v))
		os.Exit(cli.ExitValidation)
	}
	return val
}

// GetString retrieves a string flag with default
func (a *Arguments) GetString(name string, defaultVal string) string {
	v, ok := a.Flags[name]
	if !ok {
		return defaultVal
	}
	val, ok := v.(string)
	if !ok {
		return defaultVal
	}
	return val
}

// MustGetInt retrieves an int flag and panics if it doesn't exist
func (a *Arguments) MustGetInt(name string) int {
	v, ok := a.Flags[name]
	if !ok {
		slog.Error("flag not found", "flag", name)
		os.Exit(cli.ExitValidation)
	}
	val, ok := v.(int)
	if !ok {
		slog.Error("flag type assertion failed", "flag", name, "expected", "int", "got", fmt.Sprintf("%T", v))
		os.Exit(cli.ExitValidation)
	}
	return val
}

// GetInt retrieves an int flag with default
func (a *Arguments) GetInt(name string, defaultVal int) int {
	v, ok := a.Flags[name]
	if !ok {
		return defaultVal
	}
	val, ok := v.(int)
	if !ok {
		return defaultVal
	}
	return val
}

// GetBool retrieves a bool flag
func (a *Arguments) GetBool(name string) bool {
	v, ok := a.Flags[name]
	if !ok {
		return false
	}
	val, ok := v.(bool)
	if !ok {
		return false
	}
	return val
}

// GetStringSlice retrieves a string slice flag with default
func (a *Arguments) GetStringSlice(name string, defaultVal []string) []string {
	v, ok := a.Flags[name]
	if !ok {
		return defaultVal
	}
	val, ok := v.([]string)
	if !ok {
		return defaultVal
	}
	return val
}

// GetIntSlice retrieves an int slice flag with default
func (a *Arguments) GetIntSlice(name string, defaultVal []int) []int {
	v, ok := a.Flags[name]
	if !ok {
		return defaultVal
	}
	val, ok := v.([]int)
	if !ok {
		return defaultVal
	}
	return val
}
