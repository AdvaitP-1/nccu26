"""API route definitions.

These routes are internal — they are called only by the MCP layer,
never directly by agents or end-users.
"""

from __future__ import annotations

import logging

from fastapi import APIRouter, HTTPException

from app.core.overlap_service import detect_overlaps
from app.core.risk_engine import compute_file_risks
from app.schemas import AnalyzeOverlapsRequest, AnalyzeOverlapsResponse

logger = logging.getLogger(__name__)

router = APIRouter()


@router.get("/health")
async def health() -> dict[str, str]:
    return {"status": "ok"}


@router.post("/analyze/overlaps", response_model=AnalyzeOverlapsResponse)
async def analyze_overlaps(req: AnalyzeOverlapsRequest) -> AnalyzeOverlapsResponse:
    """Detect structural overlaps across agent changesets.

    Accepts one or more changesets, each containing files with source content.
    Returns per-symbol overlaps and per-file risk assessments.
    """
    if not req.changesets:
        return AnalyzeOverlapsResponse(overlaps=[], file_risks=[])

    overlaps = detect_overlaps(req)
    file_risks = compute_file_risks(overlaps, all_changesets=req.changesets)

    return AnalyzeOverlapsResponse(overlaps=overlaps, file_risks=file_risks)
