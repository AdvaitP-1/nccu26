"""Pydantic request / response schemas for the analysis API."""

from __future__ import annotations

from enum import Enum

from pydantic import BaseModel, Field


# ---------------------------------------------------------------------------
# Enums
# ---------------------------------------------------------------------------

class OverlapSeverity(str, Enum):
    CRITICAL = "critical"
    HIGH = "high"
    MEDIUM = "medium"
    LOW = "low"


# ---------------------------------------------------------------------------
# Request
# ---------------------------------------------------------------------------

class FileInput(BaseModel):
    path: str
    language: str
    content: str


class ChangeSet(BaseModel):
    agent_id: str
    files: list[FileInput]


class AnalyzeOverlapsRequest(BaseModel):
    changesets: list[ChangeSet]


# ---------------------------------------------------------------------------
# Response
# ---------------------------------------------------------------------------

class Overlap(BaseModel):
    file_path: str
    symbol_name: str
    symbol_kind: str
    agent_a: str
    agent_b: str
    severity: OverlapSeverity
    reason: str
    start_line_a: int
    end_line_a: int
    start_line_b: int
    end_line_b: int


class FileRisk(BaseModel):
    file_path: str
    risk_score: int = Field(ge=0, le=100)
    stability_score: int = Field(ge=0, le=100)
    overlap_count: int
    contributors: list[str]
    contributors_count: int = Field(ge=0)
    pairwise_overlap_count: int = Field(ge=0)
    max_severity: OverlapSeverity | None = None
    is_hotspot: bool = False
    summary: str


class AnalyzeOverlapsResponse(BaseModel):
    overlaps: list[Overlap]
    file_risks: list[FileRisk]
