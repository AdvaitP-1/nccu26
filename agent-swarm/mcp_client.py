"""
MCP client wrapper that connects to the Orca AI MCP server via SSE.

Provides a thin async context manager around the official Python MCP SDK's
SSE transport, exposing helper methods that map 1:1 to the MCP tools the
Go server registers.
"""

from __future__ import annotations

import json
from contextlib import asynccontextmanager
from typing import Any, AsyncIterator

from mcp import ClientSession
from mcp.client.sse import sse_client

from config import MCP_SSE_URL


class OrcaMcpClient:
    """High-level async wrapper around a single MCP SSE session."""

    def __init__(self, session: ClientSession):
        self._session = session

    # ------------------------------------------------------------------
    # Generic tool call
    # ------------------------------------------------------------------

    async def call_tool(self, name: str, arguments: dict[str, Any] | None = None) -> str:
        result = await self._session.call_tool(name, arguments or {})
        parts = []
        for block in result.content:
            if hasattr(block, "text"):
                parts.append(block.text)
        return "\n".join(parts)

    async def call_tool_json(self, name: str, arguments: dict[str, Any] | None = None) -> Any:
        raw = await self.call_tool(name, arguments)
        if not raw or not raw.strip():
            return {"error": "empty response from tool"}
        try:
            return json.loads(raw)
        except json.JSONDecodeError:
            return {"raw_text": raw}

    # ------------------------------------------------------------------
    # Tool-specific helpers
    # ------------------------------------------------------------------

    async def list_tools(self) -> list[str]:
        resp = await self._session.list_tools()
        return [t.name for t in resp.tools]

    async def get_vfs_state(self) -> Any:
        return await self.call_tool_json("get_vfs_state")

    async def register_push(
        self,
        branch_name: str,
        user_id: str,
        files: list[dict[str, str]],
        message: str = "",
    ) -> Any:
        return await self.call_tool_json(
            "register_push",
            {
                "branch_name": branch_name,
                "user_id": user_id,
                "files_json": json.dumps(files),
                "message": message,
            },
        )

    async def identify_overlaps(self) -> Any:
        return await self.call_tool_json("identify_overlaps")

    async def request_micro_commit(self, agent_id: str, message: str = "") -> Any:
        return await self.call_tool_json(
            "request_micro_commit",
            {"agent_id": agent_id, "message": message},
        )

    async def git_health(self) -> Any:
        return await self.call_tool_json("git_health")


@asynccontextmanager
async def connect(url: str = MCP_SSE_URL) -> AsyncIterator[OrcaMcpClient]:
    """Connect to the Orca MCP SSE server and yield a ready-to-use client."""
    sse_url = f"{url.rstrip('/')}/sse"
    async with sse_client(sse_url) as (read_stream, write_stream):
        async with ClientSession(read_stream, write_stream) as session:
            await session.initialize()
            yield OrcaMcpClient(session)
