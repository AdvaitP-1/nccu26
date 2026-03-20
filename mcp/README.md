# MCP — Git State and Execution Service

MCP server exposing Git-side state management and execution capabilities
as composable MCP tools.

External systems (e.g. IBM watsonx Orchestrate) call these tools in
whatever sequence they decide.  This layer does **not** make merge
decisions, sequence operations, or embed orchestration logic.

## Architecture

```
External caller (watsonx Orchestrate, other services)
        │
        ▼
      /mcp   ◄── this service
        │
        ├──► Git state layer (branch-scoped per-file diff trees)
        │     ├── Storage (in-memory)
        │     ├── Diff engine (patch create/apply)
        │     ├── File tree manager
        │     └── Git executor (shell-based)
        ├──► VFS (in-memory shadow workspace)
        ├──► Policy engine
        ├──► Analysis client ──► /backend (internal HTTP)
        └──► HTTP REST API
```

### Boundary

This layer is responsible for:
- tracking branch-scoped file trees
- storing diff-only nodes
- registering pushes
- reconstructing file state
- exposing merge context (data, not decisions)
- accepting externally approved merge results
- creating Git commits
- pushing commits to remote

This layer is **NOT** responsible for:
- merge decision logic (external)
- operation sequencing / orchestration (external)
- UI integration
- AST parsing (handled by `/backend`)

---

## Core Concepts

### Per-File Diff Trees

```
Branch: feature/auth

File: auth.py                    File: user.py
┌─────────────────────┐          ┌─────────────────────┐
│ Base (accepted ver)  │          │ Base (accepted ver)  │
│ ├── A1 (Dev A diff)  │          │ └── A2 (Dev A diff)  │
│ ├── B1 (Dev B diff)  │          └─────────────────────┘
│ └── C1 (Dev C diff)  │
└─────────────────────┘
```

- Each file on each branch has its own tree
- Child nodes store **only diffs** (not full snapshots)
- One push creates nodes across multiple file trees (shared `push_id`)

### External Merge Flow

The merge flow is controlled by the caller, not this layer:

```
1. Caller calls  prepare_merge_context  → gets base + candidates + diffs
2. Caller makes merge decision (externally)
3. Caller calls  apply_merge_result     → merged content applied, head promoted
4. Caller calls  prepare_commit         → dry run, checks resolution
5. Caller calls  create_commit          → grouped multi-file commit
6. Caller calls  push_commit            → push to remote
```

### Grouped Multi-File Commits

After per-file merge results are applied, files belonging to the same
push are grouped into **one** Git commit:

```
Push P1 → auth.py (A1) + user.py (A2)
After apply_merge_result on both files:
  → create_commit groups both into a single Git commit
  → push_commit sends it to remote
```

---

## MCP Tools

### Git State and Execution Tools

| Tool | Description |
|---|---|
| `git_health` | Subsystem health (repo state, entity counts) |
| `register_push` | Register a push with multiple file changes |
| `get_branch_file_state` | Per-file tree state on a branch |
| `prepare_merge_context` | Base content, candidates, diffs — for external decision |
| `apply_merge_result` | Apply externally decided merge result |
| `prepare_commit` | Dry run: what would be committed, are all files resolved? |
| `create_commit` | Grouped multi-file Git commit |
| `push_commit` | Push commit to remote with retry |
| `get_commit_status` | Commit record and status lookup |

### Existing VFS Tools

| Tool | Description |
|---|---|
| `get_vfs_state` | Pending file changes across agents |
| `identify_overlaps` | Backend overlap analysis |
| `request_micro_commit` | Analysis → policy → commit |

---

## REST API

HTTP API on a separate port (default `:9091`).

| Method | Path | Description |
|---|---|---|
| `GET` | `/git/health` | Subsystem health |
| `POST` | `/git/pushes` | Register a push |
| `GET` | `/git/branches/{branch}/files/{path}` | File tree state |
| `POST` | `/git/merge/context` | Prepare merge context |
| `POST` | `/git/merge/apply` | Apply merge result |
| `POST` | `/git/commit/prepare` | Prepare commit (dry run) |
| `POST` | `/git/commit` | Create grouped commit |
| `POST` | `/git/push` | Push to remote |
| `GET` | `/git/commits/{id}` | Commit status |

---

## Package Structure

| Package | Responsibility |
|---|---|
| `cmd/server/` | Entrypoint |
| `internal/config/` | Env-var configuration |
| `internal/models/` | Domain types |
| `internal/storage/` | In-memory store for git state |
| `internal/diff/` | Diff creation and application |
| `internal/filetree/` | Per-file tree state management |
| `internal/gitcontrol/` | Git operations (shell-based) |
| `internal/service/` | Git state + execution service |
| `internal/httpapi/` | REST API handlers |
| `internal/tools/` | MCP tool definitions |
| `internal/server/` | Server wiring + SSE transport |
| `internal/vfs/` | VFS shadow workspace |
| `internal/analysisclient/` | Backend HTTP client |
| `internal/policy/` | Commit-gating logic |
| `internal/commits/` | Micro-commit coordinator |
| `internal/logging/` | Structured logging |

---

## Local Development

### Prerequisites

- Go 1.22+
- Backend running on `http://localhost:8000`
- Git (for real commit/push operations)

### Run

```bash
cd mcp
go mod tidy
go run ./cmd/server
```

### Test

```bash
cd mcp
go test ./...
```

### Configuration

| Variable | Default | Description |
|---|---|---|
| `MCP_SERVER_ADDR` | `:9090` | MCP SSE listen address |
| `MCP_HTTP_ADDR` | `:9091` | HTTP REST API address (empty to disable) |
| `MCP_BACKEND_URL` | `http://localhost:8000` | Backend base URL |
| `MCP_BACKEND_TIMEOUT` | `10s` | Backend HTTP timeout |
| `MCP_RISK_THRESHOLD` | `70` | Max safe file risk score |
| `MCP_BLOCK_ON_CRITICAL` | `true` | Block on critical overlaps |
| `MCP_GIT_REPO_PATH` | *(empty)* | Local repo path (empty = state-tracking only) |
| `MCP_GIT_REMOTE` | `origin` | Git remote name |
| `MCP_LOG_LEVEL` | `info` | `debug` for verbose |

---

## Tradeoffs

**Diff format:** diff-match-patch (`github.com/sergi/go-diff`).
Character-level patches — higher accuracy, larger text, deterministic
and serialisable.

**Git operations:** Shell-based (`os/exec`).  Keeps dependency tree
small, uses installed git version.  All commands use
`exec.CommandContext` for timeout safety.

**Storage:** In-memory (mutex-protected maps).  A persistent backend
can replace the maps without changing service or handler code.

**Merge decisions:** Entirely external.  This layer provides context
(`prepare_merge_context`) and accepts results (`apply_merge_result`).
It never decides what the merged content should be.
