package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/tui"
)

var rootCmd = &cobra.Command{
	Use:   "paso",
	Short: "Terminal-based Kanban board with CLI and TUI",
	Long: `Paso is a zero-setup, terminal-based kanban board for personal task management.

Use 'paso tui' to launch the interactive TUI.
Use 'paso task create ...' for CLI commands.`,
	// No Run function - shows help text by default
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// Add TUI subcommand
	tuiCmd := &cobra.Command{
		Use:   "tui",
		Short: "Launch the interactive TUI",
		Long:  "Launch the interactive terminal user interface for managing tasks visually.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return tui.Launch()
		},
	}
	rootCmd.AddCommand(tuiCmd)
}
