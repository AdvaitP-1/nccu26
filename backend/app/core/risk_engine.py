"""File-level risk scoring derived from overlap analysis results.

Risk is a 0-100 score indicating how dangerous it is for agents to proceed
with their pending changes on a given file.  Stability is its complement.
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


def compute_file_risks(overlaps: list[Overlap]) -> list[FileRisk]:
    """Aggregate *overlaps* into per-file risk summaries."""
    if not overlaps:
        return []

    grouped: dict[str, list[Overlap]] = defaultdict(list)
    for o in overlaps:
        grouped[o.file_path].append(o)

    return [_score_file(path, file_overlaps) for path, file_overlaps in grouped.items()]


def _score_file(file_path: str, overlaps: list[Overlap]) -> FileRisk:
    raw_risk = sum(_SEVERITY_WEIGHTS.get(o.severity, 0) for o in overlaps)
    risk_score = min(raw_risk, settings.max_risk)
    stability_score = max(settings.base_stability - raw_risk, 0)

    severity_counts = _tally(overlaps)
    summary = (
        f"{len(overlaps)} overlap(s): "
        + ", ".join(f"{count} {sev.value}" for sev, count in severity_counts.items() if count)
    )

    return FileRisk(
        file_path=file_path,
        risk_score=risk_score,
        stability_score=stability_score,
        overlap_count=len(overlaps),
        summary=summary,
    )


def _tally(overlaps: list[Overlap]) -> dict[OverlapSeverity, int]:
    counts: dict[OverlapSeverity, int] = {s: 0 for s in OverlapSeverity}
    for o in overlaps:
        counts[o.severity] += 1
    return counts
