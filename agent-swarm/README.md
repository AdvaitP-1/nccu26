# Orca AI — Agent Swarm Demo

A multi-agent demo that connects to the Orca MCP server via the Model Context Protocol (SSE transport) and walks through the full orchestration workflow:

1. **Discovery** — list every tool the MCP server exposes
2. **Push** — three coder agents push file changes in parallel
3. **VFS inspection** — view the virtual file system with all pending changes
4. **Overlap analysis** — detect structural conflicts across agents
5. **Micro-commit** — attempt commits through the policy gate
6. **Final state** — verify the VFS is cleaned up

## Prerequisites

- Python 3.11+
- The MCP server running on `localhost:9090` (SSE) / `localhost:9091` (HTTP)
- The Python backend running on `localhost:8000`

## Quick Start

```bash
# From the repo root, start the backend + MCP server first:
cd backend && python -m uvicorn app.main:app --reload --port 8000 &
cd mcp && MCP_GIT_REPO_PATH=$(pwd)/.. go run ./cmd/server &

# Then run the swarm:
cd agent-swarm
pip install -r requirements.txt
python run_swarm.py
```

## Configuration

Set these environment variables to override defaults:

| Variable | Default | Description |
|----------|---------|-------------|
| `MCP_SSE_URL` | `http://localhost:9090` | MCP SSE server URL |
| `MCP_HTTP_URL` | `http://localhost:9091` | MCP HTTP API URL |
| `BACKEND_URL` | `http://localhost:8000` | Python backend URL |

## Project Structure

```
agent-swarm/
├── run_swarm.py       # Entry point — orchestrates the full demo
├── config.py          # Environment-based configuration
├── mcp_client.py      # MCP SSE client wrapper
├── agents/
│   ├── base.py        # Shared agent lifecycle + logging
│   ├── coder.py       # Pushes file changes to VFS
│   ├── reviewer.py    # Runs overlap analysis
│   └── merge.py       # Requests micro-commits
└── requirements.txt
```

## What the Demo Proves

- **Multi-agent collaboration**: Multiple agents connect to a single MCP server and push changes independently
- **Conflict detection**: When two agents edit the same file, Orca's structural analysis catches overlapping symbols
- **Policy enforcement**: The micro-commit gate blocks unsafe commits while allowing clean ones through
- **VFS lifecycle**: Committed changes are automatically cleared from the shadow workspace
