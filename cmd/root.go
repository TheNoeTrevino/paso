package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "paso",
	Short: "Paso - A terminal-based kanban board",
	Long:  `Paso is a terminal-based kanban board for managing tasks and projects.`,
}

func Execute() error {
	return rootCmd.Execute()
}
