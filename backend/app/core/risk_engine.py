"""File-level risk scoring derived from overlap analysis results.

Risk is a 0-100 score indicating how dangerous it is for agents to proceed
with their pending changes on a given file.  Stability is its complement.

For N-agent scenarios the engine also tracks:
  - which agents contribute to a file (contributors)
  - how many unique pairwise overlaps exist
  - the maximum severity found
  - whether the file qualifies as a hotspot
"""

from __future__ import annotations

from collections import defaultdict

from app.config import settings
from app.schemas import FileRisk, Overlap, OverlapSeverity

_SEVERITY_WEIGHTS: dict[OverlapSeverity, int] = {
    OverlapSeverity.CRITICAL: settings.critical_weight,
    OverlapSeverity.HIGH: settings.high_weight,
    OverlapSeverity.MEDIUM: settings.medium_weight,
    OverlapSeverity.LOW: settings.low_weight,
}

_SEVERITY_RANK: dict[OverlapSeverity, int] = {
    OverlapSeverity.CRITICAL: 4,
    OverlapSeverity.HIGH: 3,
    OverlapSeverity.MEDIUM: 2,
    OverlapSeverity.LOW: 1,
}

HOTSPOT_CONTRIBUTOR_THRESHOLD = 3
HOTSPOT_RISK_THRESHOLD = 50


def compute_file_risks(
    overlaps: list[Overlap],
    all_changesets: list | None = None,
) -> list[FileRisk]:
    """Aggregate *overlaps* into per-file risk summaries.

    *all_changesets* (optional) is the original request's changeset list.
    When provided it is used to compute accurate contributor lists even for
    files that have no overlaps but are touched by multiple agents.
    """
    if not overlaps and not all_changesets:
        return []

    grouped: dict[str, list[Overlap]] = defaultdict(list)
    for o in overlaps:
        grouped[o.file_path].append(o)

    file_contributors: dict[str, set[str]] = defaultdict(set)
    if all_changesets is not None:
        for cs in all_changesets:
            for fi in cs.files:
                file_contributors[fi.path].add(cs.agent_id)

    # Score files that have overlaps.
    results: list[FileRisk] = []
    for path, file_overlaps in grouped.items():
        contributors = _collect_contributors(file_overlaps) | file_contributors.get(path, set())
        results.append(_score_file(path, file_overlaps, sorted(contributors)))

    # Include files touched by multiple agents but with no overlaps.
    for path, agents in file_contributors.items():
        if path not in grouped and len(agents) >= 2:
            results.append(_score_file(path, [], sorted(agents)))

    return results


def _collect_contributors(overlaps: list[Overlap]) -> set[str]:
    agents: set[str] = set()
    for o in overlaps:
        agents.add(o.agent_a)
        agents.add(o.agent_b)
    return agents


def _score_file(
    file_path: str,
    overlaps: list[Overlap],
    contributors: list[str],
) -> FileRisk:
    raw_risk = sum(_SEVERITY_WEIGHTS.get(o.severity, 0) for o in overlaps)
    risk_score = min(raw_risk, settings.max_risk)
    stability_score = max(settings.base_stability - raw_risk, 0)

    pairwise_pairs: set[tuple[str, str]] = set()
    max_sev: OverlapSeverity | None = None
    max_rank = 0
    severity_counts = _tally(overlaps)

    for o in overlaps:
        pair = (min(o.agent_a, o.agent_b), max(o.agent_a, o.agent_b))
        pairwise_pairs.add(pair)
        rank = _SEVERITY_RANK.get(o.severity, 0)
        if rank > max_rank:
            max_rank = rank
            max_sev = o.severity

    is_hotspot = (
        len(contributors) >= HOTSPOT_CONTRIBUTOR_THRESHOLD
        or risk_score >= HOTSPOT_RISK_THRESHOLD
    )

    summary_parts = []
    if overlaps:
        summary_parts.append(
            f"{len(overlaps)} overlap(s): "
            + ", ".join(f"{c} {s.value}" for s, c in severity_counts.items() if c)
        )
    summary_parts.append(f"{len(contributors)} contributor(s)")
    if is_hotspot:
        summary_parts.append("HOTSPOT")

    return FileRisk(
        file_path=file_path,
        risk_score=risk_score,
        stability_score=stability_score,
        overlap_count=len(overlaps),
        contributors=contributors,
        contributors_count=len(contributors),
        pairwise_overlap_count=len(pairwise_pairs),
        max_severity=max_sev,
        is_hotspot=is_hotspot,
        summary=" | ".join(summary_parts),
    )


def _tally(overlaps: list[Overlap]) -> dict[OverlapSeverity, int]:
    counts: dict[OverlapSeverity, int] = {s: 0 for s in OverlapSeverity}
    for o in overlaps:
        counts[o.severity] += 1
    return counts
