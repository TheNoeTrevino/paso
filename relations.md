Right now, we have related tasks:
Parent and child issues. 

I want to add a new feature to this, it will basically be a marker to HOW the issues are related. 

Here are the details

## Database Migration

We will need to add a field to the parent child join table. Here is what it will end up being:

parent_child_relations:
- parent_id (fk to tasks, already done)
- child_id (fk to tasks, already done)
- relation_id (fk to relation_types, new)

Relation Types Table (new):
- relation_id (pk)
- p_to_c_label ("Blocked By", "Parent"(default))
- c_to_p_label ("Blocker", "Child"(default))
- color (#xxxxxx)

## Integrating with the current TUI

In the task card, we will need to display this next to the task type. Right now, it shows: 
{task type}

it should now be 
{task type} - { c_to_p_label }

## Task Details

In the task details view, we will need to show the relation type next to each related task.

In the parent section, it should show:

{task_id} - {p_to_c_label(in the color)} - {task_title}
{task_id} - {c_to_p_label(in the color)} - {task_title}

### Execution
Scaffold well, and use the beads cli for task management.
