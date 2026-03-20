# Orca AI

**Multi-agent orchestration platform that prevents AI coding agents from breaking each other's work.**

When multiple AI agents edit the same codebase in parallel, they inevitably step on each other — two agents rewrite the same function, redefine a class with conflicting fields, or produce diffs that silently overwrite each other at merge time. Orca AI solves this by sitting between the agents and the repository, parsing every proposed change at the symbol level, detecting structural overlaps in real time, and blocking unsafe commits before they happen.

---

## How It Works

```
┌─────────────────────────────────────────────────────────────────────┐
│                        AI AGENTS (N parallel)                       │
│   ┌─────────┐   ┌─────────┐   ┌─────────┐   ┌──────────────────┐  │
│   │ Coder A │   │ Coder B │   │ Coder C │   │ External IDE/LLM │  │
│   └────┬────┘   └────┬────┘   └────┬────┘   └────────┬─────────┘  │
│        │             │             │                  │             │
│        └─────────────┴─────────────┴──────────────────┘             │
│                              │ MCP Protocol (SSE)                   │
│                              ▼                                      │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │                   MCP Go Server (:9090)                      │   │
│  │                                                              │   │
│  │  ┌─────────┐  ┌────────────┐  ┌────────┐  ┌─────────────┐  │   │
│  │  │   VFS   │  │  Policy    │  │ Agent  │  │   12 MCP    │  │   │
│  │  │ Manager │  │ Evaluator  │  │Registry│  │   Tools     │  │   │
│  │  └────┬────┘  └─────┬──────┘  └────────┘  └─────────────┘  │   │
│  │       │             │                                        │   │
│  │       ▼             ▼                                        │   │
│  │  Shadow Workspace:  Risk Threshold                           │   │
│  │  tracks every       + Block-on-Critical                      │   │
│  │  pending file       policy enforcement                       │   │
│  │  per agent                                                   │   │
│  └──────────────────────────┬───────────────────────────────────┘   │
│                             │ HTTP                                  │
│                             ▼                                       │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │              Python Analysis Backend (:8000)                 │   │
│  │                                                              │   │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐  │   │
│  │  │  AST Parser  │  │   Overlap    │  │   Risk Engine    │  │   │
│  │  │  (per-lang)  │  │  Detection   │  │   (per-file      │  │   │
│  │  │              │  │  (pairwise)  │  │    scoring)      │  │   │
│  │  │ Python │ Go  │  │              │  │                  │  │   │
│  │  │  TS/JS │     │  │              │  │                  │  │   │
│  │  └──────────────┘  └──────────────┘  └──────────────────┘  │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                     │
│                             ▲ Proxy                                 │
│                             │                                       │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │              Next.js Frontend (:3000)                        │   │
│  │                                                              │   │
│  │  Landing  │  Dashboard  │  Platform  │  Docs                │   │
│  │  Page     │  (VFS +     │  Overview  │  Reference           │   │
│  │           │  Terminal + │            │                      │   │
│  │           │  Topology)  │            │                      │   │
│  └──────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Core Concepts

### AST-Level Overlap Detection

Orca doesn't compare diffs line-by-line — it parses every proposed file change into an Abstract Syntax Tree and extracts symbols (functions, classes, methods, structs, interfaces). When two agents define or modify the same symbol in the same file, Orca flags the conflict at the symbol level with severity classification:

| Severity | Meaning |
|----------|---------|
| **critical** | Same symbol, overlapping line ranges — guaranteed merge conflict |
| **warning** | Same symbol, adjacent line ranges — likely merge conflict |
| **info** | Same file, different symbols — no conflict, but worth tracking |

Supported languages: **Python** (tree-sitter AST), **TypeScript/JavaScript** (tree-sitter AST), **Go** (regex-based parser).

### Virtual File System (VFS)

Every proposed file change from every agent lives in an in-memory shadow workspace before touching the real repo. This gives Orca a global view of all pending work across all agents simultaneously, enabling cross-agent analysis before any code is committed.

### Policy Gate

Before any micro-commit is allowed, Orca evaluates:
- Per-file **risk scores** (0–100) based on overlap count, contributor count, and severity
- **Block-on-critical** policy: if any critical overlap exists anywhere in the VFS, ALL commits are blocked until the conflict is resolved
- **Risk threshold**: files above the configured score (default 70) block commits

### Task Tree Orchestration

A hierarchical task tree decomposes features into atomic coding tasks:

```
feat-user-auth (root)
├── feat-user-auth-routes-py (file-level)
│   ├── feat-user-auth-login-endpoint (leaf → coder agent)
│   ├── feat-user-auth-logout-endpoint (leaf → coder agent)
│   └── feat-user-auth-jwt-middleware (leaf → coder agent)
└── feat-user-auth-models-py (file-level)
    └── feat-user-auth-user-model (leaf → coder agent)
```

The **Manager Agent** decomposes tasks, the **Coder Agents** implement them in isolation (each sees only its own node), and the **Merge Agent** combines diffs when all siblings are complete.

---

## Project Structure

```
nccu26/
├── frontend/          Next.js 16 — landing page, PM dashboard, docs
├── backend/           FastAPI — AST parsing, overlap detection, risk scoring
├── mcp/               Go — MCP server, VFS, policy engine, git integration
├── agent-swarm/       Python — multi-agent demo that exercises the full pipeline
├── agents/            IBM watsonx Orchestrate agent YAML configs
├── tools/             IBM watsonx ADK tool bindings for the task tree
└── ARCHITECTURE.md    Data flow and boundary rules
```

### Frontend (`frontend/`)

Next.js 16 app with four main pages:

| Route | Purpose |
|-------|---------|
| `/` | Landing page — product overview, feature grid, integrations |
| `/dashboard` | PM Dashboard — live VFS state, terminal console, agent topology graph |
| `/platform` | Platform architecture overview |
| `/docs` | Full documentation — setup, architecture, MCP tool reference, API reference |

The dashboard includes:
- **VFS Panel** — real-time view of every pending file change per agent
- **Terminal Console** — command interface to interact with the MCP server (seed demo, identify overlaps, request micro-commits, register agents)
- **Topology Panel** — visual graph of all agents, MCP server, and backend with live status
- **SSE + Polling** — real-time updates via Server-Sent Events with automatic polling fallback
- **MCP Server Control** — start the MCP server directly from the dashboard UI

**Tech:** React 19, Tailwind CSS, Framer Motion, shadcn/ui, Recharts.

### Backend (`backend/`)

Python FastAPI server providing structural analysis:

| Endpoint | Purpose |
|----------|---------|
| `POST /analyze/overlaps` | Parse files, extract symbols, detect pairwise overlaps |
| `POST /tree` | Create a task tree node |
| `GET /tree/{id}` | Fetch a node |
| `GET /tree/{id}/siblings` | Fetch sibling nodes |
| `POST /tree/{id}/status` | Update node lifecycle status |
| `POST /tree/{id}/diff` | Write a unified diff to a node |
| `POST /merge` | Sequentially apply diffs to a base file |

The overlap detection algorithm:
1. Parse every file in every changeset using language-specific AST parsers
2. Group symbols by `(file_path, symbol_name, symbol_kind)`
3. For each group touched by more than one agent, compare line ranges pairwise
4. Classify severity: **critical** (overlapping), **warning** (adjacent), **info** (same file)
5. Score each file's risk on a 0–100 scale

**Tech:** FastAPI, Pydantic, tree-sitter (Python/TS/JS), regex (Go).

### MCP Server (`mcp/`)

Go server implementing the [Model Context Protocol](https://modelcontextprotocol.io/) — the open standard for connecting AI agents to tools. Exposes 12 tools via SSE transport:

| Tool | Purpose |
|------|---------|
| `get_vfs_state` | Snapshot of all pending changes across all agents |
| `identify_overlaps` | Run structural overlap analysis on the full VFS |
| `request_micro_commit` | Commit an agent's changes (policy-gated) |
| `register_push` | Ingest a push event with file changes |
| `git_health` | Git repository health check |
| `get_branch_file_state` | File state on a specific branch |
| `prepare_merge_context` | Gather merge context for a file |
| `apply_merge_result` | Apply a resolved merge |
| `prepare_commit` | Stage files for commit |
| `create_commit` | Create a git commit |
| `push_commit` | Push a commit to remote |
| `get_commit_status` | Check commit status |

Also exposes a REST API on `:9091` for the dashboard (VFS state, command execution, SSE event stream, agent registry).

**Tech:** Go 1.23, [mark3labs/mcp-go](https://github.com/mark3labs/mcp-go).

### Agent Swarm (`agent-swarm/`)

Self-contained Python demo that connects to the live MCP server and walks through the entire Orca workflow in ~4 seconds:

1. Connects via MCP SSE protocol
2. Three coder agents push file changes in parallel
3. Two agents intentionally edit the same file (`backend/shared/config.py`)
4. Overlap analysis detects 3 symbol-level conflicts
5. Policy gate blocks all commits while critical overlaps exist
6. Rich terminal output with tables, panels, and color-coded results

```bash
cd agent-swarm && pip install -r requirements.txt && python run_swarm.py
```

### IBM watsonx Orchestrate Agents (`agents/`)

Three agent definitions for IBM watsonx Orchestrate (YAML configs with full prompt engineering):

| Agent | Role | Tools |
|-------|------|-------|
| **manager_agent** | Decomposes features into task trees, dispatches coders, polls status, triggers merges | `create_node`, `get_node`, `get_node_siblings`, `get_node_children`, `update_node_status`, `trigger_merge` |
| **coder_agent** | Implements a single atomic task in isolation, produces a unified diff | `get_node`, `update_node_status`, `update_node_diff` |
| **merge_agent** | Collects diffs from completed siblings, applies them sequentially to a base file | `get_node`, `trigger_merge` |

Each agent has strict isolation rules — coder agents cannot see sibling work, and the manager never writes code.

### Tool Bindings (`tools/`)

Python functions decorated with `@tool` from the IBM watsonx Orchestrate ADK, wrapping the backend's REST API for use by the watsonx agents: `get_node`, `get_node_siblings`, `get_node_children`, `create_node`, `update_node_status`, `update_node_diff`, `trigger_merge`.

---

## Running Locally

### Prerequisites

- **Node.js** 18+ and npm
- **Python** 3.11+ and pip
- **Go** 1.23+

### 1. Start the Python backend

```bash
cd backend
pip install -r requirements.txt
python -m uvicorn app.main:app --reload --port 8000
```

### 2. Start the MCP server

```bash
cd mcp
MCP_GIT_REPO_PATH=$(pwd)/.. go run ./cmd/server
```

This starts:
- SSE server on `:9090` (MCP protocol for agent connections)
- HTTP API on `:9091` (REST for dashboard)

### 3. Start the frontend

```bash
cd frontend
cp .env.example .env
npm install && npm run dev
```

Open [http://localhost:3000](http://localhost:3000) for the landing page, [http://localhost:3000/dashboard](http://localhost:3000/dashboard) for the PM dashboard.

### 4. Run the agent swarm demo

```bash
cd agent-swarm
pip install -r requirements.txt
python run_swarm.py
```

### 5. (Optional) Expose MCP for external agents

```bash
ngrok http 9090
```

Use the ngrok URL as the MCP endpoint in IBM watsonx Orchestrate or any MCP-compatible IDE.

---

## Environment Variables

| Variable | Default | Used By | Purpose |
|----------|---------|---------|---------|
| `BACKEND_BASE_URL` | `http://localhost:8000` | Frontend | Backend proxy target |
| `MCP_BASE_URL` | `http://localhost:9091` | Frontend | MCP HTTP API proxy target |
| `MCP_BACKEND_URL` | `http://localhost:8000` | MCP | Backend for overlap analysis |
| `MCP_GIT_REPO_PATH` | *(empty)* | MCP | Git repo path (enables real git ops) |
| `MCP_SERVER_ADDR` | `:9090` | MCP | SSE server listen address |
| `MCP_HTTP_ADDR` | `:9091` | MCP | HTTP API listen address |
| `MCP_RISK_THRESHOLD` | `70` | MCP | Max risk score before blocking commits |
| `MCP_BLOCK_ON_CRITICAL` | `true` | MCP | Block all commits if any critical overlap exists |

---

## Key Design Decisions

- **Frontend never calls backend directly** — all requests go through Next.js Route Handlers (server-side), keeping backend URLs off the client
- **MCP as the integration layer** — any MCP-compatible tool (Cursor, Claude Code, watsonx, etc.) can connect to Orca without custom integrations
- **Symbol-level, not line-level** — overlap detection operates on AST symbols (functions, classes, methods), not raw line diffs, reducing false positives
- **Global policy enforcement** — when critical overlaps exist, ALL agents are blocked, not just the conflicting ones, preventing race conditions
- **VFS before repo** — every change lives in memory first, giving the system a complete picture before any code touches the real repository

---

## License

Built for NCCU Hackathon 2026.
