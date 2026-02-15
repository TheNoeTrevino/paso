# Contributing

First of all, thank you for considering contributing to this project! 

## Some Ceremonial Stuff

Before you start any work, make sure you have a issue open, and have it assigned to yourself.

This will stop two people from working on the same issue, and will also give us a place to discuss the implementation details of your contribution.

If you want to contribute, please open an issue first.

There is a chance your feature doesn't align with the vision of the project.
We want to avoid you doing work that might not be used.

## Git Workflow

1. Fork the repository
2. Create your feature/bugfix branch: `git checkout -b feature-123/your-feature`
  a. The 123 numbers should represent the issue you are working on. 
3. When committing, use the [conventional commit format](https://www.conventionalcommits.org/en/v1.0.0/).
  - You can use the `git log` for examples of previous commit messages.
  - Please try to have an understandable and followable commit history, open a new branch (don't PR changes from your main to the repository's main).
4. Before opening a PR, make sure to run tests locally, and to format your code with `gofmt -w .`
5. Open a PR to `main`

## Packages and Their Responsibilities

TODO: mention how each package has its own `docs.go` file. 
Give a quick overview of what each package is responsible for, libraries it uses, or quirks it has.

## Testing 

If you are making complex or significant changes, please consider adding tests.
We use testify for tests, and hand-roll mocks.

To run tests locally, you can use:
```bash
go test ./... -race
```

You must run this script before opening a PR. It will save everyone time, and my github minutes :)
