"""Tests for the risk engine — file-level aggregation and N-agent scoring."""

from __future__ import annotations

import pytest

from app.core.risk_engine import compute_file_risks
from app.schemas import Overlap, OverlapSeverity


def _overlap(
    file_path: str = "f.py",
    symbol: str = "fn",
    kind: str = "function",
    agent_a: str = "a",
    agent_b: str = "b",
    severity: OverlapSeverity = OverlapSeverity.CRITICAL,
) -> Overlap:
    return Overlap(
        file_path=file_path,
        symbol_name=symbol,
        symbol_kind=kind,
        agent_a=agent_a,
        agent_b=agent_b,
        severity=severity,
        reason="test",
        start_line_a=1, end_line_a=5,
        start_line_b=2, end_line_b=6,
    )


class TestRiskScoring:
    def test_no_overlaps_returns_empty(self):
        assert compute_file_risks([]) == []

    def test_single_critical_overlap(self):
        risks = compute_file_risks([_overlap(severity=OverlapSeverity.CRITICAL)])
        assert len(risks) == 1
        r = risks[0]
        assert r.risk_score == 40  # default critical weight
        assert r.max_severity == OverlapSeverity.CRITICAL
        assert r.contributors_count == 2
        assert "a" in r.contributors
        assert "b" in r.contributors

    def test_risk_clamped_at_100(self):
        overlaps = [_overlap(severity=OverlapSeverity.CRITICAL) for _ in range(10)]
        risks = compute_file_risks(overlaps)
        assert risks[0].risk_score == 100

    def test_stability_floors_at_zero(self):
        overlaps = [_overlap(severity=OverlapSeverity.CRITICAL) for _ in range(10)]
        risks = compute_file_risks(overlaps)
        assert risks[0].stability_score == 0


class TestContributorAggregation:
    def test_three_agents(self):
        overlaps = [
            _overlap(agent_a="a", agent_b="b"),
            _overlap(agent_a="a", agent_b="c"),
            _overlap(agent_a="b", agent_b="c"),
        ]
        risks = compute_file_risks(overlaps)
        assert risks[0].contributors_count == 3
        assert set(risks[0].contributors) == {"a", "b", "c"}

    def test_four_agents_pairwise_count(self):
        overlaps = [
            _overlap(agent_a="a", agent_b="b"),
            _overlap(agent_a="a", agent_b="c"),
            _overlap(agent_a="a", agent_b="d"),
            _overlap(agent_a="b", agent_b="c"),
            _overlap(agent_a="b", agent_b="d"),
            _overlap(agent_a="c", agent_b="d"),
        ]
        risks = compute_file_risks(overlaps)
        assert risks[0].pairwise_overlap_count == 6
        assert risks[0].contributors_count == 4


class TestHotspotDetection:
    def test_three_contributors_is_hotspot(self):
        overlaps = [
            _overlap(agent_a="a", agent_b="b", severity=OverlapSeverity.LOW),
            _overlap(agent_a="a", agent_b="c", severity=OverlapSeverity.LOW),
            _overlap(agent_a="b", agent_b="c", severity=OverlapSeverity.LOW),
        ]
        risks = compute_file_risks(overlaps)
        assert risks[0].is_hotspot is True

    def test_high_risk_is_hotspot(self):
        overlaps = [
            _overlap(severity=OverlapSeverity.CRITICAL),
            _overlap(severity=OverlapSeverity.HIGH),
        ]
        risks = compute_file_risks(overlaps)
        assert risks[0].is_hotspot is True

    def test_two_agents_low_risk_not_hotspot(self):
        overlaps = [_overlap(severity=OverlapSeverity.LOW)]
        risks = compute_file_risks(overlaps)
        assert risks[0].is_hotspot is False


class TestMaxSeverity:
    def test_max_severity_is_critical(self):
        overlaps = [
            _overlap(severity=OverlapSeverity.LOW),
            _overlap(severity=OverlapSeverity.CRITICAL),
            _overlap(severity=OverlapSeverity.MEDIUM),
        ]
        risks = compute_file_risks(overlaps)
        assert risks[0].max_severity == OverlapSeverity.CRITICAL

    def test_max_severity_medium_only(self):
        overlaps = [_overlap(severity=OverlapSeverity.MEDIUM)]
        risks = compute_file_risks(overlaps)
        assert risks[0].max_severity == OverlapSeverity.MEDIUM

    def test_no_overlaps_no_max_severity(self):
        risks = compute_file_risks([])
        assert risks == []


class TestMultiFileRisks:
    def test_two_files_scored_independently(self):
        overlaps = [
            _overlap(file_path="a.py", agent_a="x", agent_b="y", severity=OverlapSeverity.CRITICAL),
            _overlap(file_path="b.py", agent_a="x", agent_b="z", severity=OverlapSeverity.LOW),
        ]
        risks = compute_file_risks(overlaps)
        by_path = {r.file_path: r for r in risks}
        assert by_path["a.py"].risk_score > by_path["b.py"].risk_score
        assert by_path["a.py"].max_severity == OverlapSeverity.CRITICAL
        assert by_path["b.py"].max_severity == OverlapSeverity.LOW
