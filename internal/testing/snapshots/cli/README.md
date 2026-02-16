# CLI Snapshot Tests

This package contains golden file tests for all CLI output. These tests detect regressions in the user-facing experience by comparing current output against stored snapshots.

## How It Works

Each test captures CLI output (help text, styled messages, command results), strips ANSI escape codes, and compares the plain text against a `.golden` file in `testdata/`. If the output differs, the test fails with a diff showing exactly what changed.

## Updating Snapshots

```bash
# Update all CLI snapshots
UPDATE_SNAPSHOTS=1 go test ./internal/testing/snapshots/cli/...

# Update only help snapshots
UPDATE_SNAPSHOTS=1 go test ./internal/testing/snapshots/cli/... -run TestGolden_HelpOutput

# Update only style snapshots
UPDATE_SNAPSHOTS=1 go test ./internal/testing/snapshots/cli/... -run TestGolden_StyleOutput

# Update only e2e snapshots
UPDATE_SNAPSHOTS=1 go test ./internal/testing/snapshots/cli/... -run TestGolden_E2E
```

## Help Output Tests (`help_test.go`)

Tests the `--help` output for every CLI command and subcommand. Uses auto-discovery: `buildRootCmd()` reconstructs the full command tree and `collectCommands()` recursively walks it to generate a test case for each command. When a new command is added to paso, the test picks it up automatically on the next run.

Golden files are named by the full command path: `paso-task-create.golden`, `paso-setup-claude.golden`, etc.

## Style Tests (`styles_test.go`)

Tests styled success messages across all color schemes (default, monochrome, wave, dragon, lotus) to ensure consistent formatting.

## E2E Command Output Tests (`e2e_test.go`)

Executes real CLI commands against an in-memory SQLite database and snapshots the full human-readable output users see. Each test group gets its own isolated database, seeds the required data, runs commands through the real handler pipeline, and compares captured output against golden files.

Covers 35 scenarios across all entity types:

- **Task** (19 tests): create, create with description, list, show, update, assign, estimate, comment, ready list, blocked list, in-progress list, move, in-progress move, to-ready, done, link, switch, delete
- **Project** (5 tests): create, create with description, list, tree, delete
- **Column** (4 tests): create, list, update, delete
- **Label** (6 tests): create, list, update, attach, detach, delete
- **Assignee** (3 tests): create, list, delete

`task show` output includes timestamps which are normalized to `[TIMESTAMP]` before comparison for determinism.

### Commands Not Covered

These commands depend on external state (filesystem, config files, user environment) and are not suitable for deterministic snapshot testing:

- `setup claude`, `setup opencode` -- read/write real filesystem paths
- `db add`, `db connect`, `db list`, `db remove` -- depend on paso config file
- `assignee set`, `assignee whoami` -- depend on paso config file
- `project git-link`, `project git-unlink` -- require mock git detector wiring (tested in integration tests instead)
