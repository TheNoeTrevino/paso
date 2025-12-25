package tutorial

import (
	"fmt"

	"github.com/spf13/cobra"
)

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
	context := `# Paso Workflow Context

## Core Rules
- Use ` + "`paso task ready --project=<id>`" + ` to find actionable work (no blockers)
- Use ` + "`paso task blocked --project=<id>`" + ` to see what's waiting on dependencies
- Create blocking relationships with ` + "`paso task link --parent=<blocked> --child=<blocker> --blocker`" + `

## Essential Commands

### Projects
- ` + "`paso project list`" + ` - List all projects
- ` + "`paso project create --title=\"...\" --description=\"...\"`" + ` - Create project

### Tasks
- ` + "`paso task list --project=<id>`" + ` - List tasks in project
- ` + "`paso task ready --project=<id>`" + ` - Show ready tasks (no blockers)
- ` + "`paso task blocked --project=<id>`" + ` - Show blocked tasks
- ` + "`paso task create --project=<id> --title=\"...\" --type=task|feature --priority=medium`" + ` - Create task
- ` + "`paso task update --id=<id> --title=\"...\" --description=\"...\"`" + ` - Update task
- ` + "`paso task delete --id=<id>`" + ` - Delete task

### Dependencies
- ` + "`paso task link --parent=<id> --child=<id>`" + ` - Parent-child relationship
- ` + "`paso task link --parent=<id> --child=<id> --blocker`" + ` - Blocking dependency (parent blocked by child)
- ` + "`paso task link --parent=<id> --child=<id> --related`" + ` - Related tasks

### Columns
- ` + "`paso column list --project=<id>`" + ` - List columns
- ` + "`paso column create --project=<id> --name=\"...\"`" + ` - Create column

## Common Workflows

**Starting work:**
` + "```bash" + `
paso project list              # Find project ID
paso task ready --project=1    # Find available work
` + "```" + `

**Creating a task:**
` + "```bash" + `
paso task create --project=1 --title="Implement feature X" --type=feature
` + "```" + `

**Creating dependent work:**
` + "```bash" + `
FEATURE=$(paso task create --project=1 --title="Implement feature X" --quiet)
TESTS=$(paso task create --project=1 --title="Write tests for X" --quiet)
# Tests blocked by feature:
paso task link --parent=$TESTS --child=$FEATURE --blocker
` + "```" + `

## Output Flags
All commands support ` + "`--json`" + ` and ` + "`--quiet`" + ` flags for agent-friendly output:
- ` + "`--json`" + ` - Full JSON response
- ` + "`--quiet`" + ` - Minimal output (IDs only)
`
	fmt.Print(context)
}
