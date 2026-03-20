# MCP — Product Manager Assistant Control Plane

Central MCP (Model Context Protocol) server for the PM Assistant platform.

This is the **single system boundary** through which all agents and
orchestrators interact with the platform.  No external caller should reach
the backend directly.

## Architecture

```
IBM watsonx Orchestrate (or any MCP client)
        │
        ▼
      /mcp   ◄── this service
        │
        ├──► VFS (in-memory shadow workspace)
        ├──► Policy engine
        ├──► Commit coordinator
        └──► Analysis client ──► /backend (internal HTTP)
```

### Why MCP is the boundary

1. **Single entry point** — agents never call `/backend` or touch the repo
   directly.
2. **Policy enforcement** — every commit request passes through overlap
   analysis *and* the policy engine before anything is written.
3. **VFS isolation** — proposed changes live in a shadow workspace until
   explicitly approved, preventing uncoordinated writes.
4. **Orchestrator-ready** — the MCP protocol (JSON-RPC over SSE) is the
   transport that IBM watsonx Orchestrate uses to call remote tool servers.

## MCP Tools

| Tool | Description |
|---|---|
| `get_vfs_state` | Returns all pending file changes across agents |
| `identify_overlaps` | Gathers VFS state → calls backend `/analyze/overlaps` → returns overlaps + file risks |
| `request_micro_commit` | Runs analysis → evaluates policy → commits if safe (or blocks with reason) |

### `identify_overlaps` flow

```
1. Read all pending changesets from VFS
2. POST them to /backend/analyze/overlaps
3. Return the backend's overlap + file-risk response verbatim
```

### `request_micro_commit` flow

```
1. Verify the agent has pending changes in the VFS
2. Run identify_overlaps across the full VFS
3. Evaluate policy:
   a. Block if any overlap severity == "critical" (configurable)
   b. Block if any file risk > threshold (default 70)
4. If allowed → commit via Coordinator → clear agent's VFS entries
5. Return result with commit_id or block reason
```

## Package Structure

| Package | Responsibility |
|---|---|
| `cmd/server/` | Entrypoint — loads config, starts server |
| `internal/config/` | Env-var based configuration |
| `internal/models/` | Shared domain types (VFS, analysis, commit) |
| `internal/vfs/` | In-memory virtual file system |
| `internal/analysisclient/` | Typed HTTP client for the backend |
| `internal/policy/` | Commit-gating decision logic |
| `internal/commits/` | Micro-commit coordinator (git stub for v1) |
| `internal/tools/` | MCP tool definitions and handlers |
| `internal/server/` | MCP server wiring + SSE transport |
| `internal/logging/` | Structured logging setup |

## Local Development

### Prerequisites

- Go 1.22+
- Backend running on `http://localhost:8000` (see `/backend/README.md`)

### Run

```bash
cd mcp
go run ./cmd/server
```

The MCP server starts on `:9090` by default (SSE transport).

### Configuration

| Variable | Default | Description |
|---|---|---|
| `MCP_SERVER_ADDR` | `:9090` | Listen address |
| `MCP_BACKEND_URL` | `http://localhost:8000` | Backend base URL |
| `MCP_BACKEND_TIMEOUT` | `10s` | HTTP timeout for backend calls |
| `MCP_RISK_THRESHOLD` | `70` | Max safe file risk score (0-100) |
| `MCP_BLOCK_ON_CRITICAL` | `true` | Block commits with critical overlaps |
| `MCP_LOG_LEVEL` | `info` | Set to `debug` for verbose output |

## IBM watsonx Orchestrate Integration

watsonx Orchestrate connects to this server as a **remote MCP server**:

1. Deploy this service to **IBM Cloud Code Engine** (container).
2. Configure watsonx Orchestrate with the public SSE endpoint URL.
3. Orchestrate discovers tools via MCP `tools/list`.
4. Orchestrate calls tools via MCP `tools/call`.

The MCP protocol (JSON-RPC 2.0 over SSE) is natively supported by
watsonx Orchestrate's remote MCP server capability.

> **No IBM-specific SDK code is included.**  The integration relies entirely
> on the standard MCP protocol, which watsonx Orchestrate speaks natively.

### Code Engine deployment (v1 intent)

```
Container image → IBM Cloud Container Registry
Code Engine project → create service from image
Expose public HTTPS endpoint → configure in watsonx Orchestrate
Set env vars (MCP_BACKEND_URL, etc.) in Code Engine service config
```

No Terraform or deployment manifests are included for v1.

## Commit Coordinator — Stub Note

The commit coordinator (`internal/commits/`) currently generates a synthetic
commit ID and logs the commit.  **No actual git operations are performed.**

To add real git integration:
1. Replace the body of `(*Coordinator).Commit` in `commits.go`.
2. The interface, models, and caller code remain unchanged.
3. Consider calling an external git-service or shelling out to `git` with
   staged files from the VFS.

## Testing

```bash
cd mcp
go test ./...
```

No tests are included in v1 — the architecture is designed for easy unit
testing via dependency injection (every handler receives a `Deps` struct).
