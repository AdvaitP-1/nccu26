# Structural Analysis Backend

Internal analysis service for the Product Manager Assistant platform.
Parses source code, detects structural overlaps between agent changesets, and scores file-level risk.

> **This service is called only by the MCP layer.**
> It should never be exposed directly to agents, users, or external orchestrators.

## Architecture

```
MCP (Go)
  └─ POST /analyze/overlaps ──► this backend
```

### Components

| Package | Responsibility |
|---|---|
| `app/parsers/` | Language-specific AST parsing (Python via `ast`, TS/JS via tree-sitter) |
| `app/models/` | Normalised `Symbol` model shared across all parsers |
| `app/core/overlap_service.py` | Cross-agent structural overlap detection |
| `app/core/risk_engine.py` | Per-file risk / stability scoring |
| `app/api/routes.py` | FastAPI route definitions |
| `app/schemas.py` | Pydantic request / response models |
| `app/config.py` | Environment-variable based configuration |

## Endpoints

| Method | Path | Description |
|---|---|---|
| `GET` | `/health` | Liveness check |
| `POST` | `/analyze/overlaps` | Accept changesets → return overlaps + file risks |

### `POST /analyze/overlaps`

**Request:**
```json
{
  "changesets": [
    {
      "agent_id": "agent-alpha",
      "files": [
        { "path": "src/auth.py", "language": "python", "content": "..." }
      ]
    },
    {
      "agent_id": "agent-beta",
      "files": [
        { "path": "src/auth.py", "language": "python", "content": "..." }
      ]
    }
  ]
}
```

**Response:**
```json
{
  "overlaps": [
    {
      "file_path": "src/auth.py",
      "symbol_name": "validate_token",
      "symbol_kind": "function",
      "agent_a": "agent-alpha",
      "agent_b": "agent-beta",
      "severity": "critical",
      "reason": "Both agents modify 'validate_token' with overlapping line ranges",
      "start_line_a": 10,
      "end_line_a": 25,
      "start_line_b": 12,
      "end_line_b": 30
    }
  ],
  "file_risks": [
    {
      "file_path": "src/auth.py",
      "risk_score": 40,
      "stability_score": 60,
      "overlap_count": 1,
      "summary": "1 overlap(s): 1 critical"
    }
  ]
}
```

## Overlap Severity

| Severity | Condition |
|---|---|
| `critical` | Same symbol, overlapping line ranges |
| `high` | Same symbol, line ranges within 3 lines of each other |
| `medium` | Same symbol, non-adjacent locations in the same file |

## Supported Languages

| Language | Parser | Extensions |
|---|---|---|
| Python | Built-in `ast` module | `.py` |
| TypeScript | tree-sitter | `.ts`, `.tsx` |
| JavaScript | tree-sitter | `.js`, `.jsx` |

Extracted symbol kinds: `function`, `async_function`, `class`, `method`, `import`, `variable`, `constant`.

## Local Development

```bash
cd backend
python3 -m venv .venv && source .venv/bin/activate
pip install -r requirements.txt
uvicorn app.main:app --reload --port 8000
```

API docs at http://localhost:8000/docs

## Configuration

All values via environment variables with `BACKEND_` prefix:

| Variable | Default | Description |
|---|---|---|
| `BACKEND_HOST` | `0.0.0.0` | Bind host |
| `BACKEND_PORT` | `8000` | Bind port |
| `BACKEND_CRITICAL_WEIGHT` | `40` | Risk points per critical overlap |
| `BACKEND_HIGH_WEIGHT` | `25` | Risk points per high overlap |
| `BACKEND_MEDIUM_WEIGHT` | `15` | Risk points per medium overlap |
| `BACKEND_LOW_WEIGHT` | `5` | Risk points per low overlap |
| `BACKEND_BASE_STABILITY` | `100` | Starting stability for any file |
| `BACKEND_MAX_RISK` | `100` | Ceiling for risk scores |

## Deployment

Intended for **Vercel** (Python serverless functions or Docker).
No deployment config is included for v1 — local development only.
