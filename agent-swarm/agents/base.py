"""Base agent class with shared lifecycle and logging."""

from __future__ import annotations

import asyncio
from typing import Any

from rich.console import Console

from mcp_client import OrcaMcpClient

console = Console()


class BaseAgent:
    """Shared foundation for every agent in the swarm."""

    role: str = "base"
    color: str = "white"

    def __init__(self, agent_id: str, client: OrcaMcpClient):
        self.agent_id = agent_id
        self.client = client
        self._log_prefix = f"[bold {self.color}][{self.role.upper():>8}][/bold {self.color}]"

    # ------------------------------------------------------------------
    # Logging helpers
    # ------------------------------------------------------------------

    def log(self, message: str):
        console.print(f"  {self._log_prefix}  {message}")

    def log_success(self, message: str):
        console.print(f"  {self._log_prefix}  [green]✓[/green] {message}")

    def log_warn(self, message: str):
        console.print(f"  {self._log_prefix}  [yellow]⚠[/yellow] {message}")

    def log_error(self, message: str):
        console.print(f"  {self._log_prefix}  [red]✗[/red] {message}")

    # ------------------------------------------------------------------
    # Lifecycle
    # ------------------------------------------------------------------

    async def run(self) -> Any:
        """Override in subclasses to perform the agent's work."""
        raise NotImplementedError

    async def sleep(self, seconds: float):
        """Visible pause for demo pacing."""
        await asyncio.sleep(seconds)
