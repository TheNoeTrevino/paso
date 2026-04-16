---
description: Delegate ready paso tasks to subagents for parallel completion.
agent: build
---

Use the paso skill to understand the paso CLI commands and workflows.

$ARGUMENTS

1. Run `paso project tree <project-id>` to identify the task hierarchy
2. Run `paso task ready -p <project-id>` to find unblocked tasks that can be started
3. For each ready task, send a subagent to complete it. Give each subagent:
   - A detailed prompt on what to do and which files to look at
   - Instructions to run `paso task show <task-id>` for full task details
   - Instructions to run `paso task in-progress <task-id>` before starting work
4. Ensure subagents work on different tasks to avoid file conflicts

Have the subagents move the tasks to done when they finish their work.
Do not wait until they are _all_ done.

After all subagents finish:
- Lint, test, and build the project to ensure everything works
- Run any other standard code quality checks for the project
- Mark completed tasks with `paso task done <task-id> -q`

Provide a summary of completed tasks and important changes.
