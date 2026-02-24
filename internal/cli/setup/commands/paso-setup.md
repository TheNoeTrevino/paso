---
description: Organize tasks into EPICs with dependency hierarchies using paso.
allowed-tools: Bash(paso:*)
argument-hint: <instructions>
---

Use the paso skill to understand the paso CLI commands and workflows.

$ARGUMENTS

Your goal is to organize the current tasks into a well-structured hierarchy:

1. Run `paso project tree <project-id>` to see the current state
2. Create overarching tasks with "EPIC: " in the title
3. Create subtasks using the `--blocks <epic_task_id>` flag (or `-B`) to establish
   blocking relationships so the EPIC is blocked until subtasks are done
4. If needed, create further subtasks under those with `--blocked-by` and `--blocks`

Use a bash for-loop or script to create tasks in bulk rather than one-by-one:

```bash
EPIC=$(paso task create -t "EPIC: Frontend Feature" -p 1 -T feature -q)
for title in "Add homepage layout" "Create Sidebar component" "Create Navbar component"; do
  paso task create -t "$title" -p 1 -B $EPIC -q
done
```

The resulting `paso project tree` should show a logical hierarchy:
```
EPIC: Frontend Feature
  L BLOCKER - Add homepage layout
    L BLOCKER - Create Sidebar component
    L BLOCKER - Create Navbar component
```

Follow the agile method of breaking down tasks into smaller, manageable pieces
with clear dependencies.

Once organized, add comments to each task explaining its purpose, related files,
and any relevant implementation details. Use `-a opencode` (or your agent name)
as the author.

Run `paso project tree <project-id>` to verify the structure, then suggest which
EPIC should be worked on first.
