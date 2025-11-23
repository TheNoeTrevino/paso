# Adding a New Entity - Projects

## Tabs

Right now, we are just showing "Project 1" etc...
we are going to implement a new entity called projects, and that will be displayed in the tabs.

We should add crud operations for projects. with keybindings.

ctrl+n should prompt the user to add a new project. This will 
also use 'huh'. ctrl+d should delete the current project, with a confirmation screen,
telling them how many tasks are about to be deleted.

### Project Form

When creating a new project, we should prompt the user for the fields of the model. 

This should also be done using 'huh'

## Project Model

The project model should have
- name
- date created
- date modified
- description

## Association 

The project is our top level container. Now, each column will have to be associated to a project.
We need to write tests for this, as it is going to compolicate things a bit. For example, when we
move a column around, we need to absolutely make sure, that next and prev are correctly set. 

When a project is initialized, it should come with 3 default columns. Todo, in progress, done.

## Ticket Names

Tickets should be show a number, but that number should increament pre project. 

There can be two ticket one's, one in project 1, and one in project 2. 

We should implement something like this:
``` sql
CREATE TABLE project_counters (
    project_id INT PRIMARY KEY,
    next_ticket_number INT DEFAULT 1
);

CREATE TABLE tickets (
    id UUID PRIMARY KEY,
    project_id INT,
    ticket_number INT,  -- just the number part
    title TEXT,
    ...
    UNIQUE(project_id, ticket_number)
);
```

## Tests

We also need to write tests for this. if a user creates 3 tickets in project 1,
then switches to project 2, and creates a ticket, the project 2 ticket should be ticket 1. 

Then, if we go back and create another ticket in project 1, it should be ticket 4.


