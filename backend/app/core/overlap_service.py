"""Structural overlap detection across agent changesets.

The algorithm:
  1. Parse every file submitted in every changeset.
  2. Group parsed symbols by ``(file_path, symbol_name, symbol_kind)``.
  3. For each group touched by more than one agent, compare line ranges
     pairwise across *all* agent combinations to determine overlap severity.

This correctly handles N agents — ``itertools.combinations`` generates every
unique pair regardless of how many agents are present.
"""

from __future__ import annotations

import logging
from collections import defaultdict
from itertools import combinations

from app.models.symbols import Symbol, SymbolKind
from app.parsers import get_parser
from app.schemas import (
    AnalyzeOverlapsRequest,
    ChangeSet,
    FileInput,
    Overlap,
    OverlapSeverity,
)

logger = logging.getLogger(__name__)

_SymbolKey = tuple[str, str, SymbolKind]

ADJACENT_THRESHOLD = 3  # lines


def detect_overlaps(request: AnalyzeOverlapsRequest) -> list[Overlap]:
    """Return all symbol-level overlaps found across *request*'s changesets."""
    agent_symbols = _parse_all_changesets(request.changesets)
    return _find_overlaps(agent_symbols)


# ---------------------------------------------------------------------------
# Parsing
# ---------------------------------------------------------------------------


def _parse_all_changesets(
    changesets: list[ChangeSet],
) -> dict[str, dict[_SymbolKey, list[Symbol]]]:
    """Parse every file in every changeset.

    Returns ``{agent_id: {symbol_key: [Symbol, …]}}``.
    """
    result: dict[str, dict[_SymbolKey, list[Symbol]]] = {}

    for cs in changesets:
        symbols_by_key: dict[_SymbolKey, list[Symbol]] = defaultdict(list)
        for fi in cs.files:
            for sym in _parse_file(fi):
                key: _SymbolKey = (sym.file_path, sym.name, sym.kind)
                symbols_by_key[key].append(sym)
        result[cs.agent_id] = symbols_by_key

    return result


def _parse_file(fi: FileInput) -> list[Symbol]:
    parser = get_parser(fi.path)
    if parser is None:
        logger.info("No parser for %s — skipping", fi.path)
        return []
    try:
        return parser.parse(fi.content, fi.path)
    except Exception:
        logger.warning("Unexpected parse failure for %s", fi.path, exc_info=True)
        return []


# ---------------------------------------------------------------------------
# Overlap detection — works for any number of agents
# ---------------------------------------------------------------------------


def _find_overlaps(
    agent_symbols: dict[str, dict[_SymbolKey, list[Symbol]]],
) -> list[Overlap]:
    """Compare every pair of agents that share a symbol key.

    For N agents this produces C(N,2) pairs — correct by construction.
    """
    overlaps: list[Overlap] = []
    agent_ids = sorted(agent_symbols.keys())

    for agent_a, agent_b in combinations(agent_ids, 2):
        syms_a = agent_symbols[agent_a]
        syms_b = agent_symbols[agent_b]

        shared_keys = set(syms_a.keys()) & set(syms_b.keys())
        for key in shared_keys:
            for sym_a in syms_a[key]:
                for sym_b in syms_b[key]:
                    overlap = _compare_symbols(sym_a, sym_b, agent_a, agent_b)
                    overlaps.append(overlap)

    return overlaps


def _compare_symbols(
    sym_a: Symbol,
    sym_b: Symbol,
    agent_a: str,
    agent_b: str,
) -> Overlap:
    severity, reason = _classify(sym_a, sym_b)

    return Overlap(
        file_path=sym_a.file_path,
        symbol_name=sym_a.name,
        symbol_kind=sym_a.kind.value,
        agent_a=agent_a,
        agent_b=agent_b,
        severity=severity,
        reason=reason,
        start_line_a=sym_a.start_line,
        end_line_a=sym_a.end_line,
        start_line_b=sym_b.start_line,
        end_line_b=sym_b.end_line,
    )


def _classify(a: Symbol, b: Symbol) -> tuple[OverlapSeverity, str]:
    """Determine severity from the line-range relationship of two symbols."""
    if a.overlaps_lines(b):
        return (
            OverlapSeverity.CRITICAL,
            f"Both agents modify '{a.name}' with overlapping line ranges",
        )

    gap = min(abs(a.start_line - b.end_line), abs(b.start_line - a.end_line))
    if gap <= ADJACENT_THRESHOLD:
        return (
            OverlapSeverity.HIGH,
            f"Both agents modify '{a.name}' with adjacent line ranges ({gap} lines apart)",
        )

    return (
        OverlapSeverity.MEDIUM,
        f"Both agents define '{a.name}' in the same file at non-adjacent locations",
    )
