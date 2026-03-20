"""Orchestration tree node model.

Each node represents a discrete task assigned to an agent.  Nodes are
organised in a parent/child hierarchy so siblings can be compared for
merge conflicts.
"""

from __future__ import annotations

from enum import Enum


class NodeStatus(str, Enum):
    PENDING = "pending"
    IN_PROGRESS = "in_progress"
    COMPLETE = "complete"


class TreeNode:
    """A single node in the orchestration tree."""

    __slots__ = (
        "node_id",
        "parent_id",
        "task",
        "status",
        "diff",
        "agent_id",
    )

    def __init__(
        self,
        node_id: str,
        parent_id: str | None,
        task: str,
        agent_id: str = "",
        status: NodeStatus = NodeStatus.PENDING,
        diff: str | None = None,
    ) -> None:
        self.node_id = node_id
        self.parent_id = parent_id
        self.task = task
        self.agent_id = agent_id
        self.status = status
        self.diff = diff

    def to_dict(self) -> dict:
        return {
            "node_id": self.node_id,
            "parent_id": self.parent_id,
            "task": self.task,
            "agent_id": self.agent_id,
            "status": self.status.value,
            "diff": self.diff,
        }
