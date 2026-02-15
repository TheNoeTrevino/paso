# Paso

Paso is a dual-interface task management CLI/TUI

## Tech Stack

- Go
- BubbleTea (TUI Framework)
- Cobra (CLI Framework)
- Sqlite/Postgres (Database)
- Sqlc (DB code generation)
- Github

## Workflow

Before starting any work, make sure we are on a branch, and are tracking an issue.

For example: `feature-123/your-feature`, where 123 is the issue number you are working on.

If we are not on a branch like that, ask me what issue we are working on.
Using `gh`, find that issue on github.
If there is no issue, recommend the user to open one, and assign it to themselves. 

Ask the user if they want to create and checkout a new branch for that issue.

Follow the Git workflow outlined in the `CONTRIBUTING.md` file.

## Coding Practices

Expose struct fields directly opening visibility if needed, unless you need validation or computed logic.
Then you can create getter and setter methods for those fields, and make them private.
Prefer public fields if needed outside of the package and they don't required complex validation/error handling. 

Use `any` instead of `interface{}` for fields that can hold any type, it is a new go feature

Use interfaces to avoid coupling packages together, and to allow for easier testing and mocking. 
Define interfaces in the package that uses them, not in the package that implements them. (idiomatic to go)

Run `golangci-lint run ./...` to maintain code quality and catch potential issues early

Run `go test ./...` to ensure your code is working as expected and to catch any regressions

Use practices like the guard pattern, early returns, fail fasts, to ensure our code is easy to read. 

Avoid the N+1 query problem by fetching what you need in one query, never EVER fetch inside a for loop.

## Docs

In the `./docs` folder, we have these entries:
- `ARCHITECTURE.md`
- `TESTING.md`

Use them if needed, but load lazily, not eagerly at the start of your work
