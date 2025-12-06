## Subtasks in Paso

Currently, tickets are pretty independent and they don't link to each other. 

I like how far we have gotten on the implementation of tasks. I think our code is structured well enough that we can add a new complex feature: subtasks.

Each feature can have a subtasks. This is simply an association between two tasks. 

Now, tasks will have two new components to display: 

- parent issues (issues that _shouldn't_ be completed until this one is done)
- child issues (tasks that _should_ be completed before this one can be done)


### Planning

Use the beads cli for task management when we are doing this. 

#### Data Model

This is the first and foremost to consider. We should make a join table. 

table task_subtasks
- parent_id
- child_id

These will be foreign keys to the tasks table.

#### TUI

When we open up a task, we should have two new sections:
- Parent Tasks
- Child Tasks

These should display like a chip except it is a row,
carrying the project name + task ID and its title.

Like this

Parent Issues:
`PROJ-12: Fix the thing`

Child Issues:
`PROJ-13: Fix the other thing`

#### Other Considerations

We should make the task edit/update/read view larger. It is starting to feel cramped.
