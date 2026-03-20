#!/usr/bin/env python3
"""
Orca AI — Agent Swarm Demo
===========================

Connects multiple agents to the MCP server via SSE and walks through:

  1. Discovery  — list every MCP tool the server exposes
  2. Push phase — three coder agents push file changes in parallel
  3. VFS check  — inspect the virtual file system with all pending changes
  4. Review     — run structural overlap analysis across agents
  5. Merge      — attempt micro-commits, see policy block unsafe ones
  6. Final VFS  — verify committed files are cleared from the VFS

Run:
    cd agent-swarm
    pip install -r requirements.txt
    python run_swarm.py
"""

from __future__ import annotations

import asyncio
import sys
import time

from rich.console import Console
from rich.panel import Panel
from rich.table import Table
from rich.text import Text
from rich import box

import mcp_client
import dashboard_api
from config import MCP_SSE_URL, DEMO_BRANCH
from agents import CoderAgent, ReviewerAgent, MergeAgent

console = Console()

BANNER = r"""
   ____                     _    ___
  / __ \_____ _____ _      / \  |_ _|
 / / / / ___/ ___/ _` |   / _ \  | |
/ /_/ / /  / /__| (_| |  / ___ \ | |
\____/_/   \___/\__,_| /_/   \_\___|

      A G E N T   S W A R M   D E M O
"""


def header(title: str, step: int):
    console.print()
    console.rule(f"[bold white] STEP {step} — {title} ", style="bright_blue")
    console.print()


def render_vfs(vfs_data: dict) -> int:
    """Render VFS state as a table, return total file count."""
    pending = vfs_data.get("pending_changes", [])
    if not pending:
        console.print("  [dim]VFS is empty — no pending changes.[/dim]")
        return 0

    vfs_table = Table(
        title="Virtual File System",
        show_header=True,
        header_style="bold cyan",
        border_style="dim",
        box=box.SIMPLE_HEAVY,
    )
    vfs_table.add_column("Agent", style="bold")
    vfs_table.add_column("File", style="white")
    vfs_table.add_column("Language", style="dim")
    vfs_table.add_column("Lines", justify="right", style="dim")

    total = 0
    for change in pending:
        agent_id = change.get("agent_id", "?")
        files = change.get("files", [])
        for f in files:
            fp = f.get("path", "?")
            lang = f.get("language", "?")
            content = f.get("content", "")
            lines = len(content.splitlines()) if content else 0
            vfs_table.add_row(agent_id, fp, lang, str(lines))
            total += 1

    console.print(vfs_table)
    agent_count = len(pending)
    console.print(
        f"\n  [bold]{agent_count}[/bold] agent(s), "
        f"[bold]{total}[/bold] file(s) pending"
    )
    return total


async def main():
    console.print(Panel(
        Text(BANNER, style="bold cyan"),
        border_style="bright_blue",
        box=box.DOUBLE,
        padding=(0, 4),
    ))
    console.print(f"  [dim]MCP Server:[/dim] {MCP_SSE_URL}")
    console.print(f"  [dim]Branch:[/dim]     {DEMO_BRANCH}")
    console.print()

    # ------------------------------------------------------------------
    # Clear any existing demo data first
    # ------------------------------------------------------------------
    console.print("  [bold]Clearing existing demo data...[/bold]")
    try:
        await dashboard_api.clear_demo()
        console.print("  [green]✓[/green] Clean slate\n")
    except Exception as exc:
        console.print(f"  [yellow]⚠[/yellow] Could not clear: {exc}\n")

    # ------------------------------------------------------------------
    # Connect to MCP via SSE
    # ------------------------------------------------------------------
    console.print("  [bold]Connecting to MCP server via SSE...[/bold]")
    t0 = time.monotonic()

    try:
        async with mcp_client.connect(MCP_SSE_URL) as client:
            elapsed = time.monotonic() - t0
            console.print(f"  [green]✓[/green] Connected in {elapsed:.2f}s\n")

            # ==========================================================
            # STEP 1 — Tool Discovery
            # ==========================================================
            header("TOOL DISCOVERY", 1)
            tools = await client.list_tools()
            table = Table(
                title=f"MCP Tools ({len(tools)})",
                show_header=True,
                header_style="bold magenta",
                border_style="dim",
                box=box.SIMPLE_HEAVY,
            )
            table.add_column("#", justify="right", style="dim")
            table.add_column("Tool Name", style="bold white")
            for i, name in enumerate(tools, 1):
                table.add_row(str(i), name)
            console.print(table)

            # ==========================================================
            # STEP 2 — Coder Agents Push Changes (via Dashboard API → VFS)
            # ==========================================================
            header("CODER AGENTS — PARALLEL PUSH", 2)

            coder_ids = ["coder-auth", "coder-api", "coder-db"]
            coders = [
                CoderAgent(agent_id=cid, client=client, branch=DEMO_BRANCH)
                for cid in coder_ids
            ]

            results = await asyncio.gather(
                *(c.run() for c in coders), return_exceptions=True
            )
            for cid, res in zip(coder_ids, results):
                if isinstance(res, Exception):
                    console.print(f"  [red]✗[/red] {cid} failed: {res}")

            await asyncio.sleep(0.5)

            # ==========================================================
            # STEP 3 — VFS State Inspection (via MCP tool)
            # ==========================================================
            header("VFS STATE — PENDING CHANGES", 3)

            vfs = await client.get_vfs_state()
            render_vfs(vfs)

            await asyncio.sleep(0.5)

            # ==========================================================
            # STEP 4 — Overlap Analysis (via MCP tool)
            # ==========================================================
            header("OVERLAP ANALYSIS", 4)

            reviewer = ReviewerAgent(agent_id="reviewer-0", client=client)
            await reviewer.run()

            await asyncio.sleep(0.5)

            # ==========================================================
            # STEP 5 — Micro-Commit Attempts (via MCP tool)
            # ==========================================================
            header("MICRO-COMMIT — POLICY GATE", 5)

            merger = MergeAgent(agent_id="merge-agent", client=client)

            console.print("  [dim]Attempting commit for coder-db (no overlaps expected)...[/dim]\n")
            await merger.run_for("coder-db")
            await asyncio.sleep(0.3)

            console.print("\n  [dim]Attempting commit for coder-auth (overlapping file)...[/dim]\n")
            await merger.run_for("coder-auth")
            await asyncio.sleep(0.3)

            console.print("\n  [dim]Attempting commit for coder-api (overlapping file)...[/dim]\n")
            await merger.run_for("coder-api")

            await asyncio.sleep(0.5)

            # ==========================================================
            # STEP 6 — Final VFS State
            # ==========================================================
            header("FINAL VFS STATE", 6)

            final_vfs = await client.get_vfs_state()
            remaining = render_vfs(final_vfs)

            if remaining == 0:
                console.print("  [green]✓[/green] VFS is empty — all changes committed or cleared")

            # ==========================================================
            # Done
            # ==========================================================
            console.print()
            console.rule("[bold green] DEMO COMPLETE ", style="green")
            console.print()
            console.print(
                Panel(
                    "[bold]What you just saw:[/bold]\n\n"
                    "  1. Three coder agents pushed changes in parallel via MCP\n"
                    "  2. Two agents edited the same file (backend/shared/config.py)\n"
                    "  3. Orca detected 3 symbol-level overlaps (class + 2 functions)\n"
                    "  4. Policy gate blocked ALL commits while critical conflicts exist\n"
                    "  5. Resolution requires agents to coordinate before any can commit\n\n"
                    "[dim]This is the core safety loop: detect → block → resolve → commit.\n"
                    "No agent can merge unsafe code, even if their own files are clean.[/dim]",
                    title="Summary",
                    border_style="bright_blue",
                    box=box.ROUNDED,
                    padding=(1, 3),
                )
            )

    except ConnectionRefusedError:
        console.print(f"[red]✗ Could not connect to MCP server at {MCP_SSE_URL}[/red]")
        console.print("  Make sure the MCP server is running:")
        console.print("    cd mcp && MCP_GIT_REPO_PATH=.. go run ./cmd/server")
        sys.exit(1)
    except Exception as exc:
        console.print(f"[red]✗ Fatal error: {exc}[/red]")
        import traceback
        traceback.print_exc()
        sys.exit(1)


if __name__ == "__main__":
    asyncio.run(main())
