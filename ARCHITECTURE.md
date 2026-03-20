# Architecture Overview

## Data Flow

```
Browser  →  Frontend UI  →  Next.js Route Handlers (/frontend/app/api/*)  →  Python Backend (:8000)

IBM watsonx Orchestrate  →  MCP Go Server (:3000)  →  Python Backend (:8000)
```

## Boundaries

| Layer | Runs | Talks to | Env var |
|-------|------|----------|---------|
| **Frontend UI** (browser) | Client-side | Next.js Route Handlers (`/api/*`) | — |
| **Next.js Route Handlers** (server) | Server-side (Vercel / local) | Python backend | `BACKEND_BASE_URL` |
| **Python Backend** | Server | — | — |
| **MCP Go Server** | Server | Python backend | `MCP_BACKEND_URL` |
| **watsonx Agents** | IBM SaaS | MCP (via ngrok in demo) | `MCP_BASE_URL` |

## Key Rules

- The frontend **never** calls the backend directly from browser code.
- The frontend **never** calls MCP.
- `BACKEND_BASE_URL` is a server-only env var (no `NEXT_PUBLIC_` prefix).
- ngrok is **only** used to expose the MCP server to IBM watsonx Orchestrate during demos.
- ngrok does **not** affect the frontend ↔ backend integration path.

## Route Handler Mapping

| Frontend route | Backend endpoint |
|----------------|-----------------|
| `GET /api/health` | `GET /health` |
| `POST /api/analyze/overlaps` | `POST /analyze/overlaps` |
| `POST /api/tree` | `POST /tree` |
| `GET /api/tree/[nodeId]` | `GET /tree/{node_id}` |
| `GET /api/tree/[nodeId]/siblings` | `GET /tree/{node_id}/siblings` |
| `POST /api/tree/[nodeId]/status` | `POST /tree/{node_id}/status` |
| `POST /api/tree/[nodeId]/diff` | `POST /tree/{node_id}/diff` |
| `POST /api/merge` | `POST /merge` |

## Running Locally

1. Start the Python backend:
   ```bash
   cd backend && uvicorn app.main:app --reload --port 8000
   ```

2. Start the frontend:
   ```bash
   cd frontend && cp .env.example .env && npm run dev
   ```

3. (Optional) Start the MCP server for watsonx integration:
   ```bash
   cd mcp && go run ./cmd/server
   ```

## Deploying to Vercel

1. Set `BACKEND_BASE_URL` in the Vercel project environment variables.
2. Deploy the frontend as a standard Next.js project.
3. The route handlers run server-side on Vercel's edge/serverless infrastructure.
