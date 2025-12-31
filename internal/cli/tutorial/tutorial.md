# Paso Workflow Context

# 🚨 CRITICAL TASK TRACKING RULES 🚨

**NEVER use the TodoWrite tool or markdown TODOs in this project!**

```
✓ CORRECT:   paso task create --project=1 --title="Fix bug"
✗ WRONG:     TodoWrite tool
✗ WRONG:     - [ ] Fix bug (markdown checkbox)
✗ WRONG:     TODO comments for tracking work
```

**ALL work tracking MUST use paso commands. No exceptions.**


## Core Rules
- Use `paso task ready --project=<id>` to find actionable work (no blockers)
- Use `paso task blocked --project=<id>` to see what's waiting on dependencies
- Create blocking relationships with `paso task link --parent=<blocked> --child=<blocker> --blocker`

## Essential Commands

### Projects
- `paso project list` - List all projects
- `paso project create --title="..." --description="..."` - Create project

### Tasks
- `paso task list --project=<id>` - List tasks in project
- `paso task ready --project=<id>` - Show ready tasks (no blockers)
- `paso task blocked --project=<id>` - Show blocked tasks
- `paso task show <id>` - Display full task details (description, labels, relationships, metadata)
- `paso task create --project=<id> --title="..." --type=task|feature --priority=medium` - Create task
- `paso task create --project=<id> --title="..." --column="In Progress"` - Create in specific column
- `paso task update --id=<id> --title="..." --description="..." --priority=high` - Update task
- `paso task delete --id=<id>` - Delete task
- `paso task move --id=<id> next` - Move to next column
- `paso task move --id=<id> prev` - Move to previous column
- `paso task move --id=<id> "Done"` - Move to column by name (case-insensitive)

### Dependencies
- `paso task link --parent=<id> --child=<id>` - Parent-child relationship
- `paso task link --parent=<id> --child=<id> --blocker` - Blocking dependency (parent blocked by child)
- `paso task link --parent=<id> --child=<id> --related` - Related tasks

### Columns
- `paso column list --project=<id>` - List columns
- `paso column create --project=<id> --name="..."` - Create column
- `paso column update --id=<id> --name="..."` - Update column name
- `paso column update --id=<id> --ready` - Mark column as holding ready tasks
- `paso column update --id=<id> --completed` - Mark column as holding completed tasks
- `paso column update --id=<id> --in-progress` - Mark column as holding in-progress tasks

### Comments
- `paso task comment --id=<id> --message="..."` - Add comment to task (max 1000 chars)
- `paso task comment --id=<id> --message="..." --author="name"` - Add comment with specific author (always do this. E.g., claude/opencode/copilot)

### Labels
- `paso label list --project=<id>` - List project labels
- `paso label create --name="bug" --color="#FF0000" --project=<id>` - Create label
- `paso label attach --task=<id> --label=<id>` - Attach label to task
- `paso label detach --task=<id> --label=<id>` - Remove label from task
- `paso label update --id=<id> --name="..." --color="#RRGGBB"` - Update label
- `paso label delete --id=<id>` - Delete label

### Project Overview
- `paso project tree <project-id>` - Display hierarchical task tree showing parent-child and blocking relationships

## Getting Started

**Before beginning any work, always establish project context:**
1. If the user hasn't specified which project to work on, ask them explicitly via using the AskUser tool.
	Using the AskUser tool is MANDATORY if you are claude code.
2. Run `paso project list` to show available projects
3. Once confirmed, use that project ID for all subsequent operations

This prevents confusion and ensures all work is tracked in the correct project.

## AI Agent Workflow

**CRITICAL: Always move tasks to in-progress when you start working on them!**

**When working on tasks with paso:**
1. **IMMEDIATELY** use `paso task in-progress <id>` when you start working on a task
2. This applies to both existing tasks AND newly created tasks you're about to work on
3. This helps track active work across context compactions

**Creating and immediately working on a task:**
```bash
# Create task
TASK_ID=$(paso task create --project=1 --title="Fix bug" --quiet)

# IMMEDIATELY move to in-progress before starting work
paso task in-progress $TASK_ID

# Now start working...
```

**Working on existing task:**
```bash
# BEFORE you start working, move to in-progress
paso task in-progress 42

# Now work on the task...

# When done
paso task done 42
```

## Common Workflows

**Starting work:**
```bash
paso project list              # Find project ID
paso task ready --project=1    # Find available work
paso task show 42              # View full details before starting
```

**Creating a task:**
```bash
paso task create --project=1 --title="Implement feature X" --type=feature
```

**Creating dependent work:**
```bash
FEATURE=$(paso task create --project=1 --title="Implement feature X" --quiet)
TESTS=$(paso task create --project=1 --title="Write tests for X" --quiet)
# Tests blocked by feature:
paso task link --parent=$TESTS --child=$FEATURE --blocker
```

**Tracking progress with comments:**
```bash
paso task comment --id=42 --message="Started implementation, found edge case in auth flow"
paso task comment --id=42 --message="Edge case resolved, ready for review"
```

**Organizing with labels:**
```bash
# Create and attach labels for categorization
LABEL=$(paso label create --name="backend" --color="#4A90D9" --project=1 --quiet)
paso label attach --task=42 --label=$LABEL
```

**Moving tasks through workflow:**
```bash
paso task in-progress 42              # Start working on task
paso task move --id=42 next           # Move to next column
paso task move --id=42 "In Review"    # Move to specific column
paso task done 42                     # Mark as complete
```

**Viewing project structure:**
```bash
paso project tree 1    # See all tasks and their relationships in tree format
```

## Output Flags
All commands support `--json` and `--quiet` flags for agent-friendly output:
- `--json` - Full JSON response
- `--quiet` - Minimal output (IDs only)

**Note**: This project uses [paso](https://github.com/TheNoeTrevino/paso) for task management.
Use `paso` commands instead of markdown TODO's. See below for workflow details.
