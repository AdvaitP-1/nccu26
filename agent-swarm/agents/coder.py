"""Coder agent — pushes file changes to the MCP VFS via the dashboard API."""

from __future__ import annotations

from typing import Any

import dashboard_api
from .base import BaseAgent

# ---------------------------------------------------------------------------
# Simulated code payloads — each coder "writes" different files.
# Two coders intentionally touch a shared file to create a real overlap.
# ---------------------------------------------------------------------------

CODER_PAYLOADS: dict[str, list[dict[str, str]]] = {
    "coder-auth": [
        {
            "path": "backend/services/auth.py",
            "language": "python",
            "content": (
                "import jwt\nimport bcrypt\nfrom datetime import datetime, timedelta\n\n"
                "class AuthService:\n"
                "    SECRET = 'orca-demo-secret'\n"
                "    ALGORITHM = 'HS256'\n\n"
                "    def hash_password(self, raw: str) -> str:\n"
                "        return bcrypt.hashpw(raw.encode(), bcrypt.gensalt()).decode()\n\n"
                "    def verify_password(self, raw: str, hashed: str) -> bool:\n"
                "        return bcrypt.checkpw(raw.encode(), hashed.encode())\n\n"
                "    def create_token(self, user_id: str) -> str:\n"
                "        payload = {'sub': user_id, 'exp': datetime.utcnow() + timedelta(hours=1)}\n"
                "        return jwt.encode(payload, self.SECRET, algorithm=self.ALGORITHM)\n\n"
                "    def decode_token(self, token: str) -> dict:\n"
                "        return jwt.decode(token, self.SECRET, algorithms=[self.ALGORITHM])\n"
            ),
        },
        {
            "path": "backend/shared/config.py",
            "language": "python",
            "content": (
                "# Shared configuration — written by coder-auth\n\n"
                "class AppConfig:\n"
                "    APP_NAME = 'orca-platform'\n"
                "    AUTH_TOKEN_TTL = 3600\n"
                "    MAX_LOGIN_ATTEMPTS = 5\n\n"
                "def load_config() -> AppConfig:\n"
                "    return AppConfig()\n\n"
                "def validate_token(token: str) -> bool:\n"
                "    return len(token) > 10 and token.startswith('eyJ')\n"
            ),
        },
    ],
    "coder-api": [
        {
            "path": "backend/services/api_router.py",
            "language": "python",
            "content": (
                "from fastapi import APIRouter, Depends, HTTPException\n"
                "from pydantic import BaseModel\n\n"
                "router = APIRouter(prefix='/api/v1')\n\n"
                "class TaskCreate(BaseModel):\n"
                "    title: str\n"
                "    description: str = ''\n"
                "    priority: int = 1\n\n"
                "class TaskResponse(BaseModel):\n"
                "    id: str\n"
                "    title: str\n"
                "    status: str\n\n"
                "@router.post('/tasks', response_model=TaskResponse)\n"
                "async def create_task(payload: TaskCreate):\n"
                "    return TaskResponse(id='task-001', title=payload.title, status='pending')\n\n"
                "@router.get('/tasks/{task_id}', response_model=TaskResponse)\n"
                "async def get_task(task_id: str):\n"
                "    return TaskResponse(id=task_id, title='Demo Task', status='in_progress')\n"
            ),
        },
        {
            "path": "backend/shared/config.py",
            "language": "python",
            "content": (
                "# Shared configuration — written by coder-api\n\n"
                "class AppConfig:\n"
                "    APP_NAME = 'orca-platform'\n"
                "    API_RATE_LIMIT = 100\n"
                "    DEFAULT_PAGE_SIZE = 25\n\n"
                "def load_config() -> AppConfig:\n"
                "    return AppConfig()\n\n"
                "def validate_token(token: str) -> bool:\n"
                "    return bool(token) and '.' in token\n"
            ),
        },
    ],
    "coder-db": [
        {
            "path": "backend/services/database.py",
            "language": "python",
            "content": (
                "from typing import Any\nimport uuid\n\n"
                "class InMemoryDB:\n"
                "    def __init__(self):\n"
                "        self._store: dict[str, dict[str, Any]] = {}\n\n"
                "    def insert(self, collection: str, doc: dict) -> str:\n"
                "        doc_id = str(uuid.uuid4())\n"
                "        self._store.setdefault(collection, {})[doc_id] = doc\n"
                "        return doc_id\n\n"
                "    def get(self, collection: str, doc_id: str) -> dict | None:\n"
                "        return self._store.get(collection, {}).get(doc_id)\n\n"
                "    def list_all(self, collection: str) -> list[dict]:\n"
                "        return list(self._store.get(collection, {}).values())\n\n"
                "db = InMemoryDB()\n"
            ),
        },
    ],
}


class CoderAgent(BaseAgent):
    """Simulates a coding agent that pushes file changes to the MCP VFS."""

    role = "coder"
    color = "cyan"

    def __init__(self, agent_id: str, client: Any, branch: str):
        super().__init__(agent_id, client)
        self.branch = branch

    async def run(self) -> dict:
        files = CODER_PAYLOADS.get(self.agent_id, [])
        if not files:
            self.log_warn(f"No payload defined for {self.agent_id}")
            return {}

        file_paths = [f["path"] for f in files]
        self.log(f"Pushing {len(files)} file(s): {', '.join(file_paths)}")

        # Register the agent in the MCP agent registry
        await dashboard_api.register_agent(
            self.agent_id, "coder", f"Swarm {self.agent_id}"
        )

        # Propose files to the VFS via dashboard API
        result = await dashboard_api.propose_files(
            agent_id=self.agent_id,
            files=files,
            session_id=f"swarm-{self.agent_id}",
            task_id=f"task-{self.agent_id}",
        )

        if result.get("success"):
            self.log_success(f"Push registered — {len(files)} file(s) staged in VFS")
        else:
            self.log_error(f"Push failed: {result.get('error', 'unknown')}")

        return result
