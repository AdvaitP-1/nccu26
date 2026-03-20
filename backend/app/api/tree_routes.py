"""Orchestration tree and merge endpoints.

These let agents (via the MCP) manage a task tree:
  - fetch a node and its diff
  - query siblings for coordination
  - update status as work progresses
  - write diffs when work is done
  - merge multiple diffs against a base file
"""

from __future__ import annotations

import logging

from fastapi import APIRouter, HTTPException

from app.core.merge_service import merge_diffs
from app.core.tree_store import store
from app.models.tree import NodeStatus, TreeNode
from app.schemas_tree import (
    CreateNodeRequest,
    MergeConflictResponse,
    MergeRequest,
    MergeResponse,
    NodeResponse,
    NodeStatusEnum,
    SiblingsResponse,
    UpdateStatusRequest,
    WriteDiffRequest,
)

logger = logging.getLogger(__name__)

router = APIRouter()


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

def _node_to_response(node: TreeNode) -> NodeResponse:
    return NodeResponse(
        node_id=node.node_id,
        parent_id=node.parent_id,
        task=node.task,
        agent_id=node.agent_id,
        status=NodeStatusEnum(node.status.value),
        diff=node.diff,
    )


def _require_node(node_id: str) -> TreeNode:
    node = store.get(node_id)
    if node is None:
        raise HTTPException(status_code=404, detail=f"Node '{node_id}' not found")
    return node


# ---------------------------------------------------------------------------
# Routes
# ---------------------------------------------------------------------------

@router.post("/tree", response_model=NodeResponse, status_code=201)
async def create_node(req: CreateNodeRequest) -> NodeResponse:
    """Seed a new task node in the tree."""
    node = TreeNode(
        node_id=req.node_id,
        parent_id=req.parent_id,
        task=req.task,
        agent_id=req.agent_id,
    )
    store.upsert(node)
    logger.info("Node created: %s", req.node_id)
    return _node_to_response(node)


@router.get("/tree/{node_id}", response_model=NodeResponse)
async def get_node(node_id: str) -> NodeResponse:
    """Fetch a node's task, status, and diff."""
    return _node_to_response(_require_node(node_id))


@router.get("/tree/{node_id}/siblings", response_model=SiblingsResponse)
async def get_siblings(node_id: str) -> SiblingsResponse:
    """Get all sibling nodes and their statuses."""
    _require_node(node_id)
    siblings = store.siblings(node_id)
    return SiblingsResponse(
        node_id=node_id,
        siblings=[_node_to_response(s) for s in siblings],
    )


@router.post("/tree/{node_id}/status", response_model=NodeResponse)
async def update_status(node_id: str, req: UpdateStatusRequest) -> NodeResponse:
    """Update a node's status."""
    updated = store.set_status(node_id, NodeStatus(req.status.value))
    if updated is None:
        raise HTTPException(status_code=404, detail=f"Node '{node_id}' not found")
    logger.info("Node %s status → %s", node_id, req.status.value)
    return _node_to_response(updated)


@router.post("/tree/{node_id}/diff", response_model=NodeResponse)
async def write_diff(node_id: str, req: WriteDiffRequest) -> NodeResponse:
    """Write a diff result to a node."""
    updated = store.set_diff(node_id, req.diff)
    if updated is None:
        raise HTTPException(status_code=404, detail=f"Node '{node_id}' not found")
    logger.info("Diff written to node %s (%d chars)", node_id, len(req.diff))
    return _node_to_response(updated)


@router.post("/merge", response_model=MergeResponse)
async def merge(req: MergeRequest) -> MergeResponse:
    """Submit a merge job — apply a list of diffs to a base file."""
    if not req.diffs:
        raise HTTPException(status_code=422, detail="At least one diff is required")

    result = merge_diffs(req.base_content, req.diffs)

    return MergeResponse(
        success=result.success,
        merged_content=result.merged_content,
        conflicts=[
            MergeConflictResponse(diff_index=c.diff_index, reason=c.reason)
            for c in result.conflicts
        ],
    )
