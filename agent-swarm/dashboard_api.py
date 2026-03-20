"""
HTTP client for the MCP Dashboard REST API (port 9091).

Used by the swarm to push file proposals to the VFS and manage agent
registration — operations that go through the dashboard command endpoint
rather than the MCP SSE transport.
"""

from __future__ import annotations

from typing import Any

import httpx

from config import MCP_HTTP_URL

_TIMEOUT = 15.0


async def _command(command: str, args: dict[str, Any] | None = None) -> dict:
    url = f"{MCP_HTTP_URL.rstrip('/')}/dashboard/command"
    payload = {"command": command, "args": args or {}}
    async with httpx.AsyncClient(timeout=_TIMEOUT) as client:
        resp = await client.post(url, json=payload)
        resp.raise_for_status()
        return resp.json()


async def propose_files(
    agent_id: str,
    files: list[dict[str, str]],
    session_id: str = "swarm-session",
    task_id: str = "swarm-task",
) -> dict:
    """Push file proposals directly into the MCP VFS."""
    return await _command(
        "propose_files",
        {
            "agent_id": agent_id,
            "session_id": session_id,
            "task_id": task_id,
            "files": files,
        },
    )


async def register_agent(
    agent_id: str,
    agent_type: str = "coder",
    display_name: str = "",
) -> dict:
    return await _command(
        "register_agent",
        {
            "id": agent_id,
            "type": agent_type,
            "display_name": display_name or agent_id,
        },
    )


async def clear_demo() -> dict:
    return await _command("clear_demo")


async def seed_demo() -> dict:
    return await _command("seed_demo")
