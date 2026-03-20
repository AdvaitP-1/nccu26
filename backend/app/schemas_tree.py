"""Pydantic schemas for the orchestration tree and merge APIs."""

from __future__ import annotations

from enum import Enum

from pydantic import BaseModel, Field


# ---------------------------------------------------------------------------
# Enums
# ---------------------------------------------------------------------------

class NodeStatusEnum(str, Enum):
    PENDING = "pending"
    IN_PROGRESS = "in_progress"
    COMPLETE = "complete"


# ---------------------------------------------------------------------------
# Shared node representation
# ---------------------------------------------------------------------------

class NodeResponse(BaseModel):
    node_id: str
    parent_id: str | None
    task: str
    agent_id: str
    status: NodeStatusEnum
    diff: str | None


# ---------------------------------------------------------------------------
# POST /tree — create a node
# ---------------------------------------------------------------------------

class CreateNodeRequest(BaseModel):
    node_id: str
    parent_id: str | None = None
    task: str
    agent_id: str = ""


# ---------------------------------------------------------------------------
# POST /tree/{node_id}/status
# ---------------------------------------------------------------------------

class UpdateStatusRequest(BaseModel):
    status: NodeStatusEnum


# ---------------------------------------------------------------------------
# POST /tree/{node_id}/diff
# ---------------------------------------------------------------------------

class WriteDiffRequest(BaseModel):
    diff: str


# ---------------------------------------------------------------------------
# GET /tree/{node_id}/siblings
# ---------------------------------------------------------------------------

class SiblingsResponse(BaseModel):
    node_id: str
    siblings: list[NodeResponse]


# ---------------------------------------------------------------------------
# POST /merge
# ---------------------------------------------------------------------------

class MergeRequest(BaseModel):
    base_content: str
    diffs: list[str]


class MergeConflictResponse(BaseModel):
    diff_index: int
    reason: str


class MergeResponse(BaseModel):
    success: bool
    merged_content: str
    conflicts: list[MergeConflictResponse] = Field(default_factory=list)
