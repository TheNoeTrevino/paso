package tutorial

import (
	_ "embed"
	"fmt"

	"github.com/spf13/cobra"
)

//go:embed tutorial.md
var tutorialContent string

// TutorialCmd returns the prime command
func TutorialCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tutorial",
		Short: "Output AI-optimized workflow context",
		Long: `Output essential paso workflow context in AI-optimized markdown format.

Designed for Claude Code hooks (SessionStart, PreCompact) to prevent
agents from forgetting paso workflow after context compaction.`,
		Run: func(cmd *cobra.Command, args []string) {
			outputTutorialContext()
		},
	}
	return cmd
}

func outputTutorialContext() {
	fmt.Print(tutorialContent)
}
