---
name: paso
description: Manage tasks, dependencies, and project workflows using the paso CLI. Use this skill to create, update, and track tasks in a paso project. Always use paso commands for task management instead of markdown TODOs or other tools. Follow the critical rules for tracking work and dependencies to ensure clear visibility into project status.
license: MIT
---

## What I do

Provide complete knowledge of the paso CLI for terminal-based kanban task management.
This skill covers all paso commands, workflows, and best practices for AI agents.

## CRITICAL TASK TRACKING RULES

**NEVER use the TodoWrite tool or markdown TODOs in a paso project!**

```
CORRECT:   paso task create -t "Fix bug" -p 1
WRONG:     TodoWrite tool
WRONG:     - [ ] Fix bug (markdown checkbox)
WRONG:     TODO comments for tracking work
```

**ALL work tracking MUST use paso commands. No exceptions.**

## Core Concepts

### Git Branch Project Association (IMPORTANT)

Paso can link projects to git branches. **When a branch has a linked project, you do NOT need
the `-p/--project` flag on most commands** — paso automatically detects the project.

**AI agents should:**
1. Always check if the current branch has a linked project: `paso project -s`
2. Link projects to branches whenever working on a feature: `paso project git-link <project-id>`
3. Omit `-p` flags on commands when the branch is linked — the CLI handles it automatically

```bash
# Link a project to the current branch
paso project git-link 5

# Link a specific project to a specific branch
paso project git-link 5 feature/auth

# Check which project is linked to current branch
paso project -s

# After linking, these commands work WITHOUT -p flag:
paso task list           # auto-detects project from branch
paso task ready          # auto-detects project from branch
paso task blocked        # auto-detects project from branch
paso task create -t "Fix bug"   # -p not needed if branch is linked!
```

**Commands that auto-detect project when branch is linked:**
- `paso task list`, `paso task ready`, `paso task blocked`, `paso task in-progress`
- `paso task create` (omit `-p` flag)
- `paso project tree` (can omit project-id positional arg)
- `paso label list`, `paso column list`

### Assignees

Paso tracks who is working on what. Agents should identify themselves.

```bash
paso assignee create -n "opencode"          # Create an assignee
paso assignee set -n "opencode"             # Set as active assignee
paso assignee whoami                        # Check active assignee
paso task assign 42 opencode                # Assign task to someone
paso task assign 42                         # Assign to active assignee
paso task assign 42 --clear                 # Remove assignee
paso task create -t "Fix bug" -p 1 -a opencode  # Assign on creation
```

### Dependencies and Relationships

The `--parent` and `--child` semantics with `--blocker` can be confusing:
- `--parent` = The task that IS BLOCKED (waits for the other)
- `--child` = The task DOING THE BLOCKING (must be completed first)

```bash
# Blocking: task 22 is blocked by task 26
paso task link -P 22 -C 26 -b

# Parent-child (non-blocking hierarchy)
paso task link -P 5 -C 3

# Related (non-blocking association)
paso task link -P 5 -C 3 -R
```

You can also set dependencies inline when creating tasks:

```bash
paso task create -t "Write tests" -p 1 -b 15     # blocked by task 15
paso task create -t "Epic task" -p 1 -B 20        # blocks task 20
paso task create -t "Subtask" -p 1 -P 5           # child of task 5
```

### Estimates

Tasks can have time estimates using compact duration format.

```bash
paso task estimate 42 2h        # 2 hours
paso task estimate 42 1d        # 1 day
paso task estimate 42 1w2d      # 1 week and 2 days
paso task estimate 42 --clear   # Remove estimate
```

You can also set estimates on creation:

```bash
paso task create -t "Fix bug" -p 1 -e 2h
```

## Essential Commands

### Projects
- `paso project list` - List all projects
- `paso project create -t "..." -d "..."` - Create project
- `paso project tree <project-id>` - Display hierarchical task tree
- `paso project git-link <project-id> [branch]` - Link project to git branch
- `paso project git-unlink <project-id>` - Unlink project from branch
- `paso project -s` - Show project linked to current git branch

### Tasks
- `paso task list [project-id]` - List tasks (project auto-detected from branch if linked)
- `paso task ready -p <id>` - Show ready tasks (no blockers)
- `paso task blocked -p <id>` - Show blocked tasks
- `paso task show <id>` - Full task details (description, labels, relationships, metadata)
- `paso task create -t "..." -p <id> [flags]` - Create task (see create flags below)
- `paso task update <id> -t "..." -d "..." -r high` - Update task
- `paso task delete <id> -f` - Delete task (use `-f` to skip confirmation)
- `paso task assign <id> [name]` - Assign task to someone
- `paso task estimate <id> <duration>` - Set time estimate

### Task Create Flags
```
-t, --title          Task title (required)
-p, --project        Project ID (auto-detected from branch if linked)
-d, --description    Task description (use - for stdin)
-T, --type           task or feature (default: task)
-r, --priority       trivial, low, medium, high, critical (default: medium)
-c, --column         Column name (defaults to first column)
-a, --assignee       Assignee name (defaults to active assignee)
-e, --estimate       Time estimate (e.g. 2h, 30m, 1d)
-b, --blocked-by     Task ID that blocks this task
-B, --blocks         Task ID that is blocked by this task
-P, --parent         Parent task ID (creates hierarchy)
-q, --quiet          Minimal output (ID only, for bash capture)
-j, --json           Full JSON output
```

### Task Movement
- `paso task in-progress <id>` - Move to in-progress column
- `paso task done <id>` - Move to completed column
- `paso task to-ready <id>` - Move to ready column
- `paso task move -i <id> next` - Move to next column
- `paso task move -i <id> prev` - Move to previous column
- `paso task move -i <id> "Column Name"` - Move to specific column (case-insensitive)

### Dependencies
- `paso task link -P <parent> -C <child>` - Parent-child relationship
- `paso task link -P <parent> -C <child> -b` - Blocking dependency (parent blocked by child)
- `paso task link -P <parent> -C <child> -R` - Related tasks

### Comments
- `paso task comment -i <id> -m "..." -a "opencode"` - Add comment with author

### Assignees
- `paso assignee list` - List all assignees
- `paso assignee create -n "..."` - Create assignee
- `paso assignee set -n "..."` - Set active assignee
- `paso assignee whoami` - Show active assignee
- `paso assignee delete <id>` - Delete assignee

### Labels
- `paso label list -p <id>` - List project labels
- `paso label create -n "bug" --color="#FF0000" -p <id>` - Create label
- `paso label attach --task=<id> --label=<id>` - Attach label to task
- `paso label detach --task=<id> --label=<id>` - Remove label from task

### Columns
- `paso column list -p <id>` - List columns
- `paso column create -p <id> -n "..."` - Create column
- `paso column update -i <id> -r` - Mark as ready column
- `paso column update -i <id> -c` - Mark as completed column
- `paso column update -i <id> -I` - Mark as in-progress column

## Output Flags

All commands support `--json`/`-j` and `--quiet`/`-q` flags:
- `--json` - Full JSON response
- `--quiet` - Minimal output (IDs only, useful for bash variable capture)

## Getting Started

**Before beginning any work, always establish project context:**
1. If the user hasn't specified which project to work on, ask them explicitly via using the AskUser tool.
   Using the AskUser tool is MANDATORY if you are claude code.
2. Run `paso project list` to show available projects
3. **Check if the current branch is linked:** `paso project -s`
   - If linked: You're ready! Omit `-p` flags from all commands.
   - If not linked: Run `paso project git-link <project-id>` to link the project to the current branch.
4. After linking, you can use commands without `-p` flags (e.g., `paso task create -t "Fix bug"` instead of `paso task create -t "Fix bug" -p 1`)

## AI Agent Workflow

**CRITICAL: Always move tasks to in-progress when you start working on them!**

**When working on tasks with paso:**
1. **IMMEDIATELY** use `paso task in-progress <id>` when you start working on a task
2. This applies to both existing tasks AND newly created tasks you're about to work on
3. This helps track active work across context compactions

**Creating and immediately working on a task (with git-linked project):**
```bash
# Assuming project is already linked to branch, no -p needed!
TASK_ID=$(paso task create -t "Fix bug" -a opencode -q)
paso task in-progress $TASK_ID
# Now start working...
```

**If project is not linked yet:**
```bash
paso project git-link 1  # Link project 1 to current branch
TASK_ID=$(paso task create -t "Fix bug" -a opencode -q)  # No -p needed now!
paso task in-progress $TASK_ID
```

**Working on existing task:**
```bash
paso task in-progress 42
# Work on the task...
paso task done 42
```

## Bulk Task Creation with Bash Loops

When creating many tasks at once, use bash for-loops instead of running commands one at a time.

**Recommended: Use git-linking to avoid -p flags:**
```bash
# Link project to current branch first
paso project git-link 1

# Now create tasks without -p flag!
for title in "Setup database schema" "Create API endpoints" "Write unit tests" "Add input validation"; do
  paso task create -t "$title" -q
done
```

**Creating tasks with dependencies (EPIC pattern with git-linking):**
```bash
# Assuming project is already linked to branch
EPIC=$(paso task create -t "EPIC: User Authentication" -T feature -q)

for title in "Implement login endpoint" "Add password hashing" "Create session management"; do
  TASK_ID=$(paso task create -t "$title" -B $EPIC -q)
done
```

**Creating tasks with assignees and estimates (git-linked):**
```bash
# Assuming project is already linked to branch
declare -A TASKS=(
  ["Setup CI pipeline"]="2h"
  ["Write integration tests"]="4h"
  ["Deploy to staging"]="1h"
)

for title in "${!TASKS[@]}"; do
  paso task create -t "$title" -e "${TASKS[$title]}" -a opencode -q
done
```

**If you cannot use git-linking (fallback):**
```bash
PROJECT=1

for title in "Setup database schema" "Create API endpoints" "Write unit tests"; do
  paso task create -t "$title" -p $PROJECT -q
done
```

## Common Workflows

**Starting work (recommended workflow with git-linking):**
```bash
paso project list                  # Find project ID
paso project git-link 1            # Link project to current branch
paso task ready                    # Find available work (no -p needed!)
paso task show 42                  # View full details before starting
paso task in-progress 42           # Mark as in-progress
```

**Starting work (without git-linking, if needed):**
```bash
paso project list                  # Find project ID
paso task ready -p 1               # Find available work
paso task show 42                  # View full details before starting
paso task in-progress 42           # Mark as in-progress
```

**Creating dependent work (EPIC pattern with git-linked project):**
```bash
# Assuming project is already linked to branch
EPIC=$(paso task create -t "EPIC: User Authentication" -T feature -q)
LOGIN=$(paso task create -t "Implement login endpoint" -B $EPIC -q)
HASH=$(paso task create -t "Add password hashing" -B $EPIC -q)

# Result: EPIC shows "Blocked By: LOGIN, HASH"
```

**Tracking progress with comments:**
```bash
paso task comment -i 42 -m "Started implementation, found edge case in auth flow" -a opencode
paso task comment -i 42 -m "Edge case resolved, ready for review" -a opencode
```

**Moving tasks through workflow:**
```bash
paso task in-progress 42              # Start working on task
paso task move -i 42 next             # Move to next column
paso task move -i 42 "In Review"      # Move to specific column
paso task done 42                     # Mark as complete
```

**Viewing project structure:**
```bash
paso project tree 1    # See all tasks and their relationships in tree format
```

## When to use me

Use this when you are managing tasks with paso, instead of a todo list.

**Note**: This project uses [paso](https://github.com/TheNoeTrevino/paso) for task management.
Use `paso` commands instead of markdown TODO's.
