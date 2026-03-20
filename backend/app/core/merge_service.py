"""Merge service — applies a list of diffs to a base file.

For v1 the merge strategy is sequential text-patch application using
Python's ``difflib``.  Each diff is expected to be a unified-diff string.
If any hunk fails to apply cleanly, it is recorded as a conflict rather
than silently dropped.
"""

from __future__ import annotations

import difflib
import logging
from dataclasses import dataclass, field

logger = logging.getLogger(__name__)


@dataclass
class MergeConflict:
    diff_index: int
    reason: str


@dataclass
class MergeResult:
    success: bool
    merged_content: str
    conflicts: list[MergeConflict] = field(default_factory=list)


def merge_diffs(base_content: str, diffs: list[str]) -> MergeResult:
    """Apply *diffs* sequentially to *base_content*.

    Each entry in *diffs* is a unified-diff string.  They are applied in
    order; the output of one becomes the base for the next.

    Returns the final merged content and any conflicts encountered.
    """
    current = base_content
    conflicts: list[MergeConflict] = []

    for idx, diff_text in enumerate(diffs):
        if not diff_text.strip():
            continue

        patched = _apply_unified_diff(current, diff_text)
        if patched is None:
            conflicts.append(
                MergeConflict(diff_index=idx, reason="diff could not be applied cleanly")
            )
            logger.warning("Merge conflict at diff index %d", idx)
        else:
            current = patched

    return MergeResult(
        success=len(conflicts) == 0,
        merged_content=current,
        conflicts=conflicts,
    )


def _apply_unified_diff(base: str, diff_text: str) -> str | None:
    """Best-effort unified-diff application.

    Parses hunk headers and attempts line-level patching.  Returns ``None``
    when a hunk cannot be located in the base text.
    """
    base_lines = base.splitlines(keepends=True)
    result_lines = list(base_lines)
    diff_lines = diff_text.splitlines(keepends=True)

    offset = 0
    i = 0

    while i < len(diff_lines):
        line = diff_lines[i]

        if not line.startswith("@@"):
            i += 1
            continue

        hunk = _parse_hunk_header(line)
        if hunk is None:
            return None

        orig_start, orig_count = hunk
        orig_start -= 1  # 0-index
        i += 1

        remove_lines: list[str] = []
        add_lines: list[str] = []

        while i < len(diff_lines) and not diff_lines[i].startswith("@@"):
            dl = diff_lines[i]
            if dl.startswith("-"):
                remove_lines.append(dl[1:])
            elif dl.startswith("+"):
                add_lines.append(dl[1:])
            elif dl.startswith(" "):
                remove_lines.append(dl[1:])
                add_lines.append(dl[1:])
            i += 1

        pos = orig_start + offset
        end_pos = pos + len(remove_lines)

        if end_pos > len(result_lines):
            return None

        result_lines[pos:end_pos] = add_lines
        offset += len(add_lines) - len(remove_lines)

    return "".join(result_lines)


def _parse_hunk_header(line: str) -> tuple[int, int] | None:
    """Extract (start, count) for the original side of a ``@@ -s,c +s,c @@`` header."""
    try:
        parts = line.split("@@")[1].strip()
        old_range = parts.split(" ")[0]  # e.g. "-10,5"
        nums = old_range.lstrip("-").split(",")
        start = int(nums[0])
        count = int(nums[1]) if len(nums) > 1 else 1
        return start, count
    except (IndexError, ValueError):
        return None
