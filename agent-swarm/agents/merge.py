"""Merge agent — requests micro-commits for agents whose changes are safe."""

from __future__ import annotations

from typing import Any

from rich.panel import Panel
from rich import box

from .base import BaseAgent, console


class MergeAgent(BaseAgent):
    """Attempts a micro-commit for a given agent, respecting policy evaluation."""

    role = "merge"
    color = "green"

    async def run_for(self, target_agent_id: str) -> dict:
        self.log(f"Requesting micro-commit for [bold]{target_agent_id}[/bold]...")

        result = await self.client.request_micro_commit(
            agent_id=target_agent_id,
            message=f"micro-commit: {target_agent_id} changes",
        )

        allowed = result.get("allowed", False)
        reasons = result.get("reasons", [])

        if allowed:
            sha = result.get("sha", result.get("commit_id", "n/a"))
            self.log_success(
                f"Commit [bold green]ALLOWED[/bold green] for {target_agent_id}"
            )
            panel = Panel(
                f"[green]Agent:[/green] {target_agent_id}\n"
                f"[green]SHA:[/green]   {sha}\n"
                f"[green]Files:[/green] {result.get('files_committed', '?')}",
                title="Micro-Commit Success",
                border_style="green",
                box=box.ROUNDED,
                padding=(0, 2),
            )
            console.print(panel)
        else:
            self.log_warn(
                f"Commit [bold red]BLOCKED[/bold red] for {target_agent_id}"
            )
            if reasons:
                for r in reasons:
                    self.log(f"  → {r}")
            panel = Panel(
                f"[red]Agent:[/red]   {target_agent_id}\n"
                f"[red]Reasons:[/red]\n" + "\n".join(f"  • {r}" for r in reasons),
                title="Commit Blocked by Policy",
                border_style="red",
                box=box.ROUNDED,
                padding=(0, 2),
            )
            console.print(panel)

        return result

    async def run(self) -> Any:
        self.log_warn("MergeAgent.run() should not be called directly — use run_for()")
        return {}
