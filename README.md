<div align="center">
    <h1 align="center">Paso 👣</h1>
    <h4 align="center">
      A terminal-based kanban board for managing your projects and tasks.
    </h4>
    <p>
      Dual interface (CLI + TUI) &bull; SQLite or PostgreSQL &bull; Git-aware &bull; AI agent friendly
    </p>
</div>

TODO: cool photo of a spanish dude walking slowly
> Paso is Spanish for "step"

Paso is a zero-setup task management tool that runs entirely in your terminal.
Use the interactive TUI for a visual kanban board, or the CLI for scripting and automation.
Tasks, projects, labels, comments, and relationships are all stored locally in SQLite — or remotely in PostgreSQL if you prefer.

## Features

- **Kanban board and list view** — Toggle between a visual board and a sortable table view
- **Full task management** — Types (task, feature, bug), five priority levels, time estimates, due dates, and assignees
- **Task relationships** — Parent-child, blocking, and related-to links between tasks, with a tree view to visualize dependency chains
- **GitHub and Jira import** — Pull in issues, with their comments, directly as Paso tasks
- **Labels** — Color-coded, project-scoped tags you can attach to any task
- **Comments and activity feed** — Add notes to tasks and see a timeline of all changes
- **Filtering and search** — Filter your board by priority, type, assignee, label, or free-text search
- **Project spaces** — Organize work into separate projects, each with their own columns and labels
- **Git branch linking** — Associate a project with a git branch so Paso knows what you're working on
- **Multiple databases** — Save and switch between SQLite and PostgreSQL connections
- **AI agent integration** — TODO: Install the slash commands and skills with setup....
- **Real-time sync** — A lightweight daemon keeps multiple terminal sessions in sync via Unix sockets
- **Configurable keybindings** — Remap every TUI keybinding to your preference
- **Themes** — Five built-in color presets (default, monochrome, wave, dragon, lotus) plus full custom theming
- **Shell completions** — Tab completion for bash, zsh, fish, and powershell

## Installation

### Prerequisites

- [Go](https://golang.org/dl/) 1.26+

Optional:
- [gh CLI](https://cli.github.com/) — for GitHub issue import
- [jira CLI](https://github.com/ankitpokhrel/jira-cli) — for Jira issue import
- systemd — for the real-time sync daemon (Linux only)

### From Source

```bash
git clone https://github.com/TheNoeTrevino/paso.git
cd paso
./scripts/install.sh
```

The install script will:
1. Build and install the `paso` binary to `~/.local/bin`
2. Optionally set up shell completions
3. Optionally install and enable the systemd service for real-time sync

TODO: pre-built binaries via github releases

## Quick Start

```bash
# Create your first project
paso project create --title="My Project"

# Create some tasks
paso task create --project=1 --title="Set up the database" --priority=high
paso task create --project=1 --title="Write the tests" --type=feature

# See what's ready to work on
paso task ready --project=1

# Or launch the interactive TUI
paso tui
```

## CLI Commands

Every command supports `--json` and `--quiet` flags for scripting and agent use.

| Command | Description |
|---|---|
| `paso tui` | Launch the interactive kanban board |
| `paso task` | Create, list, show, update, delete, move, link, comment on tasks |
| `paso project` | Create and manage projects, view task trees, link to git branches |
| `paso column` | Create, rename, delete, and configure board columns |
| `paso label` | Create, attach, detach, and manage color-coded labels |
| `paso assignee` | Manage assignees, set active identity with `whoami` and `set` |
| `paso db` | Add, connect, list, and remove database connections |
| `paso gh import` | Import a GitHub issue as a task |
| `paso jira import` | Import a Jira issue as a task |
| `paso daemon` | Set up, start, stop, and check status of the real-time sync daemon |
| `paso setup` | Configure integrations with Claude Code and OpenCode |
| `paso tutorial` | Print workflow context for AI agents |
| `paso completion` | Generate shell completions for bash, zsh, fish, or powershell |

Run `paso --help` or `paso <command> --help` for full details on any command.

## Configuration

Paso stores its configuration at `~/.config/paso/config.yaml` (follows XDG conventions).

### Themes

Set a built-in preset or define your own colors:

```yaml
theme:
  preset: dragon  # default, monochrome, wave, dragon, lotus
```

You can also override individual colors or point to an external theme file with the `PASO_THEME_FILE` environment variable.

### Keybindings

All TUI keys are remappable. 

See [`example_config.yaml`](example_config.yaml) for every available option with defaults.

### Database Connections

Manage connections through the CLI or config file:

```bash
# Add a SQLite database
paso db add --name=local --connection="~/.paso/tasks.db" --type=sqlite

# Add a remote PostgreSQL database
paso db add --name=remote --connection="postgres://user:pass@host/db" --type=postgres

# Switch between them
paso db connect --name=remote
```

## External Integrations

### GitHub

Import issues directly into your Paso board. Requires the [gh CLI](https://cli.github.com/).

```bash
paso gh import 101 --project=1
```

### Jira

Import Jira issues the same way. Requires [jira-cli](https://github.com/ankitpokhrel/jira-cli).

```bash
paso jira import PROJ-123 --project=1
```

### Real-time Sync Daemon

Paso includes a lightweight daemon that broadcasts events over a Unix socket, keeping multiple terminal sessions (CLI and TUI) in sync. The daemon is managed entirely through the `paso` CLI — no separate binary required.

```bash
# Install and enable the daemon as a systemd user service (starts on login)
paso daemon setup

# Check if the daemon is running
paso daemon status

# Stop the daemon
paso daemon stop

# Remove the systemd service
paso daemon setup --remove

# Run in the foreground for debugging
paso daemon start
```

Requires systemd (Linux only).

### AI Agents

Paso has strong support for AI coding assistants. The `--json` and `--quiet` flags on every command make it easy for agents to read output.
The setup commands install tools for you to reduce repetitive motions when using Paso with an AI buddy:

```bash
# Set up OpenCode skills and slash commands
paso setup opencode

# Set up Claude Code skills and slash commands
paso setup claude
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup, coding conventions, and how to submit changes.

## License

TBD
