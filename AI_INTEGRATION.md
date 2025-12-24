# AI Integration: Beads vs Paso Feature Comparison

> Research document comparing beads and paso to identify features that improve LLM/agent effectiveness.

## Executive Summary

**Paso** is a terminal-based kanban board for personal task management with CLI + TUI interfaces.

**Beads** is a dependency-aware issue tracker specifically designed for AI agents with persistent memory across sessions.

Both serve task management, but beads has several features that would significantly improve LLM/agent experience with paso.

---

## Features Beads Has That Paso Lacks

### 1. **Dependency Graph with Blocking Semantics** ⭐ HIGH IMPACT

**Beads:**
- `bd ready` - Shows only issues with no blockers (immediately actionable)
- `bd blocked` - Shows issues waiting on dependencies
- Dependency types: `blocks`, `related`, `parent-child`, `discovered-from`
- Automatic status tracking when blockers close

**Paso:**
- Only has `task link --parent --child` for subtask relationships
- No concept of "blocking" vs "related" dependencies
- No `ready` command to filter actionable work

**Why it matters for LLMs:** Agents need to know what they CAN work on right now. Without blocking semantics, agents may pick tasks that have unsatisfied dependencies.

---

### 2. **AI Context Injection (`bd prime`)** ⭐ HIGH IMPACT

*re("persistence").select()*Beads:**
- `bd prime` outputs optimized workflow context for AI agents
- Designed to survive context compaction/summarization
- Auto-injects via hooks (SessionStart, PreCompact)
- Token-efficient: ~1-2k tokens for CLI mode, ~50 tokens for MCP mode

**Paso:**
- No equivalent feature
- Agents must re-learn paso commands each session

**Why it matters for LLMs:** After context compaction or new sessions, agents forget workflow patterns. `prime` re-injects the essential workflow knowledge.

---

### 4. **Semantic Compaction / Memory Decay** ⭐ MEDIUM IMPACT

**Beads:**
- Summarizes old closed tasks to save context window
- Tier 1: 30 days closed → 70% size reduction
- Tier 2: 90 days closed → 95% size reduction
- AI-driven summarization workflow

**Paso:**
- No compaction mechanism
- Closed tasks remain at full size forever

**Why it matters for LLMs:** Context windows are limited. Old completed tasks consume tokens without providing value.

---

### 5. **Inter-Agent Messaging** ⭐ MEDIUM IMPACT

**Beads:**
- `bd mail send <recipient> -s "Subject" -m "Body"`
- `bd mail inbox` - Check messages
- `bd mail ack/reply` - Acknowledge/respond
- Enables async agent coordination

**Paso:**
- No messaging mechanism

**Why it matters for LLMs:** Multi-agent workflows need coordination. Messages allow agents to communicate without blocking.

---

### 6. **Duplicate Detection** ⭐ LOW IMPACT

**Beads:**
- `bd duplicates` finds content duplicates
- `bd duplicates --auto-merge` auto-merges

**Paso:**
- No duplicate detection

**Why it matters for LLMs:** Agents sometimes create duplicate tasks. Auto-detection keeps the database clean.

---

### 7. **MCP Server Integration** ⭐ MEDIUM IMPACT

**Beads:**
- Full MCP server in `integrations/beads-mcp/`
- Multi-project routing
- Per-project daemon isolation

**Paso:**
- No MCP server

**Why it matters for LLMs:** MCP enables richer tool integration with AI IDEs like Claude Desktop and Cursor.

---

### 8. **Hook System for Extensibility** ⭐ LOW IMPACT

**Beads:**
- `on_create`, `on_update`, `on_close`, `on_message` hooks
- Async/sync execution
- Git hooks for sync

**Paso:**
- No hook system

---

### 9. **Skills/Templates System** ⭐ LOW IMPACT

**Beads:**
- Structured SKILL.md with AI guidance
- Reference files for workflows
- When to use bd vs TodoWrite guidance

**Paso:**
- No built-in guidance

---

### 10. **Stealth Mode** ⭐ LOW IMPACT

**Beads:**
- `bd init --stealth` for local-only without git

**Paso:**
- Already local-only (no git integration to hide)

---

## Recommended Features to Implement

Based on user feedback, focusing on **two high-impact features**:

### Feature 1: Dependency Blocking Semantics

**New Commands:**
- `paso task ready --project=<id>` - Show tasks with no blockers (immediately actionable)
- `paso task blocked --project=<id>` - Show blocked tasks with their blockers

**Database Changes:**
- Add `link_type` column to task links table (values: `blocks`, `related`, `parent`)
- Default existing links to `parent` type

**CLI Changes:**
- Extend `paso task link --parent=<id> --child=<id> --type=blocks|related|parent`

**TUI Changes:**
- Visual indicator for blocked tasks (e.g., lock icon, grayed out)
- Filter view to show only ready tasks

**How beads does it:** `beads/internal/storage/storage.go` has `GetBlockedByDependencies()` and `GetReadyIssues()` methods that filter based on dependency status.

---

### Feature 2: AI Context Injection (`paso prime`)

**New Command:**
- `paso prime` - Output workflow context for AI agents

**Output Content:**
- Available commands summary
- Common workflows (create project, add task, update status)
- Output format options (--json, --quiet)
- Session close protocol reminder

**Token Efficiency:**
- Target: ~500-1000 tokens for full context
- Structured markdown format
- No verbose explanations

**How beads does it:** `beads/cmd/bd/prime.go` outputs a compact workflow guide. It detects MCP mode and adjusts verbosity. Key sections:
- Core rules (brief)
- Essential commands (table format)
- Common workflows (code blocks)
- Session close protocol

---

## Implementation Files (Paso)

| File | Changes |
|------|---------|
| `internal/database/migrations.go` | Add `link_type` column |
| `internal/database/task_repository.go` | Add `GetReadyTasks()`, `GetBlockedTasks()` |
| `internal/cli/task.go` | Add `ready` and `blocked` subcommands |
| `cmd/prime.go` (new) | Implement `paso prime` command |
| `internal/tui/model.go` | Add blocked status display |

---

## NOT Implementing (Per User Feedback)

- Git-backed storage / JSONL sync
- Inter-agent messaging
- MCP server integration

---

## Additional Beads Architecture Details

### Beads Storage Architecture
- **Three-layer design**: Storage layer (SQLite/memory) → RPC layer (daemon) → CLI layer
- **Distributed database pattern**: SQLite (local, fast) ↔ JSONL (git-tracked) ↔ Remote JSONL
- **Auto-sync**: 5-second debounce between SQLite and JSONL

### Beads Daemon Architecture
- Per-workspace daemon for auto-sync and RPC operations
- Unix socket communication (`.beads/bd.sock`)
- Background sync with configurable intervals
- Event-driven mode for sub-500ms latency

### Beads Core Types
- **Issue Types**: Bug, Feature, Task, Epic, Chore
- **Statuses**: Open, InProgress, Blocked, Closed, Tombstone
- **Dependency Types**:
  1. `blocks` - Hard blocker
  2. `related` - Soft link
  3. `parent-child` - Hierarchical (epic/subtask)
  4. `discovered-from` - Provenance tracking

### Key Beads Design Patterns for AI

1. **Compaction Survival**: Issues persist in SQLite/JSONL when conversation history is deleted
2. **Dependency Graph**: Never work on blocked tasks; `bd ready` shows actionable work
3. **Dual Storage**: With daemon (RPC) or without daemon (direct SQLite)
4. **Multi-Repo Support**: Per-project daemon isolation, workspace routing
5. **Token-Optimized Context**: MCP mode (~50 tokens) vs CLI mode (~1-2k tokens)
6. **Inter-Agent Communication**: Sender/Ephemeral fields, mail inbox/send commands
7. **CRDT-Like Merge**: Content hash for deterministic dedup, tombstone support

---

## References

- Beads repository: `/home/noetrevino/projects/paso/feature/beads/`
- Beads documentation: `beads/README.md`, `beads/AGENTS.md`, `beads/CLAUDE.md`
- Beads MCP server: `beads/integrations/beads-mcp/`
- Beads skills: `beads/skills/beads/`
