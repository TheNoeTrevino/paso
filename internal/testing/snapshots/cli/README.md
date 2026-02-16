# CLI Snapshot Tests

This package contains golden file tests for all CLI output. These tests detect regressions in the user-facing experience by comparing current output against stored snapshots.

## How It Works

Each test captures CLI output (help text, styled messages, etc.), strips ANSI escape codes, and compares the plain text against a `.golden` file in `testdata/`. If the output differs, the test fails with a diff showing exactly what changed.

## Updating Snapshots

```bash
# Update all CLI snapshots
UPDATE_SNAPSHOTS=1 go test ./internal/testing/snapshots/cli/...

# Update only help snapshots
UPDATE_SNAPSHOTS=1 go test ./internal/testing/snapshots/cli/... -run TestGolden_HelpOutput

# Update only style snapshots
UPDATE_SNAPSHOTS=1 go test ./internal/testing/snapshots/cli/... -run TestGolden_StyleOutput
```

## Help Output Tests (`help_test.go`)

Tests the `--help` output for every CLI command and subcommand. Uses auto-discovery: `buildRootCmd()` reconstructs the full command tree and `collectCommands()` recursively walks it to generate a test case for each command. When a new command is added to paso, the test picks it up automatically on the next run.

Golden files are named by the full command path: `paso-task-create.golden`, `paso-setup-claude.golden`, etc.

## Style Tests (`styles_test.go`)

Tests styled success messages across all color schemes (default, monochrome, wave, dragon, lotus) to ensure consistent formatting.
