"""Reviewer agent — runs overlap detection across the VFS."""

from __future__ import annotations

from typing import Any

from rich.table import Table

from .base import BaseAgent, console


class ReviewerAgent(BaseAgent):
    """Scans all pending VFS changes and reports structural overlaps."""

    role = "reviewer"
    color = "yellow"

    async def run(self) -> dict:
        self.log("Running structural overlap analysis across all agents...")

        result = await self.client.identify_overlaps()

        overlaps = result.get("overlaps", [])
        file_risks = result.get("file_risks", [])
        note = result.get("note", "")

        if note:
            self.log(f"Note: {note}")
            return result

        # Display overlaps
        if overlaps:
            self.log_warn(f"Found {len(overlaps)} overlap(s)!")
            table = Table(
                title="Structural Overlaps",
                show_header=True,
                header_style="bold magenta",
                border_style="dim",
                padding=(0, 1),
            )
            table.add_column("File", style="white")
            table.add_column("Symbol", style="cyan")
            table.add_column("Agents", style="yellow")
            table.add_column("Severity", style="red")

            for o in overlaps:
                agent_a = o.get("agent_a", "")
                agent_b = o.get("agent_b", "")
                agents_str = f"{agent_a}, {agent_b}" if agent_a else ", ".join(o.get("agents", []))
                symbol = o.get("symbol_name", o.get("symbol", "?"))
                kind = o.get("symbol_kind", "")
                symbol_display = f"{symbol} ({kind})" if kind else symbol
                table.add_row(
                    o.get("file_path", "?"),
                    symbol_display,
                    agents_str,
                    o.get("severity", "?"),
                )

            console.print(table)
        else:
            self.log_success("No structural overlaps detected")

        # Display file risk scores
        if file_risks:
            risk_table = Table(
                title="File Risk Scores",
                show_header=True,
                header_style="bold blue",
                border_style="dim",
                padding=(0, 1),
            )
            risk_table.add_column("File", style="white")
            risk_table.add_column("Risk Score", justify="right", style="bold")
            risk_table.add_column("Agents", style="dim")

            for fr in file_risks:
                score = fr.get("risk_score", 0)
                score_style = "green" if score < 40 else ("yellow" if score < 70 else "red")
                contributors = fr.get("contributors", fr.get("agents", []))
                summary = fr.get("summary", "")
                risk_table.add_row(
                    fr.get("file_path", "?"),
                    f"[{score_style}]{score}[/{score_style}]",
                    ", ".join(contributors),
                )
                if summary:
                    console.print(f"  [dim]{summary}[/dim]")

            console.print(risk_table)

        return result
