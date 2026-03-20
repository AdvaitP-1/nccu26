"""In-memory orchestration tree store.

Provides O(1) lookups by node ID and O(n) sibling queries (n = children of
the same parent).  Designed for easy replacement with a persistent backend
later.
"""

from __future__ import annotations

import threading
from typing import Sequence

from app.models.tree import NodeStatus, TreeNode


class TreeStore:
    """Thread-safe, dict-backed tree store."""

    def __init__(self) -> None:
        self._lock = threading.Lock()
        self._nodes: dict[str, TreeNode] = {}

    # ------------------------------------------------------------------
    # Reads
    # ------------------------------------------------------------------

    def get(self, node_id: str) -> TreeNode | None:
        with self._lock:
            return self._nodes.get(node_id)

    def siblings(self, node_id: str) -> list[TreeNode]:
        """Return all nodes sharing the same parent (excluding *node_id* itself)."""
        with self._lock:
            node = self._nodes.get(node_id)
            if node is None:
                return []
            return [
                n
                for n in self._nodes.values()
                if n.parent_id == node.parent_id and n.node_id != node_id
            ]

    # ------------------------------------------------------------------
    # Writes
    # ------------------------------------------------------------------

    def upsert(self, node: TreeNode) -> None:
        with self._lock:
            self._nodes[node.node_id] = node

    def set_status(self, node_id: str, status: NodeStatus) -> TreeNode | None:
        with self._lock:
            node = self._nodes.get(node_id)
            if node is not None:
                node.status = status
            return node

    def set_diff(self, node_id: str, diff: str) -> TreeNode | None:
        with self._lock:
            node = self._nodes.get(node_id)
            if node is not None:
                node.diff = diff
            return node

    # ------------------------------------------------------------------
    # Bulk helpers
    # ------------------------------------------------------------------

    def all_nodes(self) -> list[TreeNode]:
        with self._lock:
            return list(self._nodes.values())

    def clear(self) -> None:
        with self._lock:
            self._nodes.clear()


# Module-level singleton so all routes share one store.
store = TreeStore()
