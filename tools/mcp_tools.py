"""MCP task-tree tool bindings for IBM WatsonX Orchestrate ADK.

These tools wrap the FastAPI task-tree backend running at MCP_BASE_URL
(default http://localhost:8000). All agents in this system call these
tools to interact with the orchestration tree.

Endpoint reference (backend/app/api/tree_routes.py):
  GET  /tree/{node_id}            — fetch a node
  GET  /tree/{node_id}/siblings   — fetch siblings
  POST /tree                      — create a node
  POST /tree/{node_id}/status     — update status
  POST /tree/{node_id}/diff       — write diff
  POST /merge                     — merge diffs into a base file

NodeResponse fields: node_id, parent_id, task, agent_id, status, diff
Status values: "pending" | "in_progress" | "complete"
"""

from __future__ import annotations

import os
from typing import Any

import requests
from ibm_watsonx_orchestrate.agent_builder.tools import tool

_BASE = os.environ.get("MCP_BASE_URL", "http://localhost:8000")
_TIMEOUT = 10  # seconds


# ---------------------------------------------------------------------------
# get_node
# ---------------------------------------------------------------------------

@tool
def get_node(node_id: str) -> dict[str, Any]:
    """Fetch a single task-tree node by its ID.

    Call this FIRST when an agent begins work on an assigned node. It returns
    the full node record including the task description, current status, any
    diff already written, and the agent assigned to it.

    Use this tool whenever you need to:
    - Read the task description before starting implementation (coder_agent)
    - Check whether a node already has a diff written to it
    - Confirm the current status of a node before updating it
    - Collect a node's diff before triggering a merge (merge_agent)

    Args:
        node_id: The unique identifier of the node to fetch. This is the
                 string ID that was supplied when the node was created (e.g.
                 "feat-auth-login-handler"). Case-sensitive.

    Returns:
        A dict with the following keys:
            node_id    (str)       — same as the input node_id
            parent_id  (str|None)  — ID of the parent node, None for root nodes
            task       (str)       — human-readable description of the work to do
            agent_id   (str)       — which agent is assigned (may be empty string)
            status     (str)       — one of "pending", "in_progress", "complete"
            diff       (str|None)  — unified-diff string if work is done, else None

        On failure, returns {"error": "<message>"}.
    """
    try:
        resp = requests.get(f"{_BASE}/tree/{node_id}", timeout=_TIMEOUT)
        resp.raise_for_status()
        return resp.json()
    except requests.HTTPError as exc:
        return {
            "error": (
                f"Node '{node_id}' not found or server error "
                f"(HTTP {exc.response.status_code})"
            )
        }
    except requests.RequestException as exc:
        return {"error": f"Could not reach task-tree backend: {exc}"}


# ---------------------------------------------------------------------------
# get_node_siblings
# ---------------------------------------------------------------------------

@tool
def get_node_siblings(node_id: str) -> dict[str, Any]:
    """Fetch all sibling nodes — nodes that share the same parent.

    Use this to understand the full scope of parallel work happening at the
    same level of the task tree. The manager_agent uses this to:
    - Monitor whether all siblings are "complete" before triggering a merge
    - Detect if any sibling has failed or is stalled
    - Confirm the full sibling group before dispatching merge_agent

    IMPORTANT: Only the manager_agent should call this tool. Coder agents
    do not have visibility into sibling nodes — they work in strict isolation.

    Args:
        node_id: The ID of a node whose siblings you want. The response
                 excludes node_id itself — only the other children of the
                 same parent are returned. The node must exist.

    Returns:
        A dict with the following keys:
            node_id   (str)   — the queried node's ID
            siblings  (list)  — list of NodeResponse dicts, each with:
                                  node_id, parent_id, task, agent_id,
                                  status, diff

        On failure, returns {"error": "<message>"}.
    """
    try:
        resp = requests.get(
            f"{_BASE}/tree/{node_id}/siblings", timeout=_TIMEOUT
        )
        resp.raise_for_status()
        return resp.json()
    except requests.HTTPError as exc:
        return {
            "error": (
                f"Siblings query failed for node '{node_id}' "
                f"(HTTP {exc.response.status_code})"
            )
        }
    except requests.RequestException as exc:
        return {"error": f"Could not reach task-tree backend: {exc}"}


# ---------------------------------------------------------------------------
# get_node_children
# ---------------------------------------------------------------------------

@tool
def get_node_children(
    parent_node_id: str,
    any_child_node_id: str,
) -> dict[str, Any]:
    """Fetch all child nodes of a given parent node.

    Because the tree API exposes siblings (not children directly), this tool
    requires you to provide both the parent node ID and the ID of any one
    known child. It fetches that child, fetches its siblings, and combines
    them to reconstruct the full list of children.

    Use this tool as the manager_agent to:
    - Monitor overall progress of a decomposed task
    - Determine whether all leaf nodes under a parent are "complete"
    - Identify which children are stalled or failed

    IMPORTANT: Only manager_agent should call this. You must know at least
    one child of the parent — this is always true because manager_agent
    creates all nodes itself and tracks their IDs.

    Args:
        parent_node_id:    ID of the parent node whose children you want.
                           Used only to label the response — not queried
                           directly (no backend /children route exists).
        any_child_node_id: ID of any one known child of that parent. Both
                           GET /tree/{any_child_node_id} and
                           GET /tree/{any_child_node_id}/siblings are called
                           internally to reconstruct the full children list.

    Returns:
        A dict with the following keys:
            parent_id  (str)   — same as parent_node_id input
            children   (list)  — all child node records (NodeResponse dicts),
                                 including any_child_node_id itself

        On failure, returns {"error": "<message>"}.
    """
    try:
        # Fetch the known child
        child_resp = requests.get(
            f"{_BASE}/tree/{any_child_node_id}", timeout=_TIMEOUT
        )
        child_resp.raise_for_status()
        child_node = child_resp.json()

        # Fetch its siblings (other children of the same parent)
        sib_resp = requests.get(
            f"{_BASE}/tree/{any_child_node_id}/siblings", timeout=_TIMEOUT
        )
        sib_resp.raise_for_status()
        siblings_data = sib_resp.json()

        all_children = [child_node] + siblings_data.get("siblings", [])
        return {"parent_id": parent_node_id, "children": all_children}

    except requests.HTTPError as exc:
        return {
            "error": (
                f"Children query failed (HTTP {exc.response.status_code}) — "
                f"verify both '{parent_node_id}' and '{any_child_node_id}' exist"
            )
        }
    except requests.RequestException as exc:
        return {"error": f"Could not reach task-tree backend: {exc}"}


# ---------------------------------------------------------------------------
# update_node_status
# ---------------------------------------------------------------------------

@tool
def update_node_status(node_id: str, status: str) -> dict[str, Any]:
    """Update the lifecycle status of a task-tree node.

    Call this to signal progress through the node's work lifecycle:
      1. Set to "in_progress" immediately when you start working on a task,
         before writing any code.
      2. Set to "complete" ONLY AFTER you have successfully written the diff
         with update_node_diff. Never mark a node complete without a diff.

    Valid status transitions:
        "pending"     → "in_progress"   (agent starts working)
        "in_progress" → "complete"      (agent finished, diff written)
        "in_progress" → "pending"       (manager resets a stalled node for retry)

    Args:
        node_id: The ID of the node to update. The node must already exist
                 in the tree.
        status:  New status string. Must be exactly one of:
                   "pending"     — work has not started (default at creation)
                   "in_progress" — work is actively underway
                   "complete"    — work is finished and diff is written

    Returns:
        The updated NodeResponse dict (same shape as get_node return value),
        or {"error": "<message>"} if the node is not found or status value
        is invalid.
    """
    try:
        resp = requests.post(
            f"{_BASE}/tree/{node_id}/status",
            json={"status": status},
            timeout=_TIMEOUT,
        )
        resp.raise_for_status()
        return resp.json()
    except requests.HTTPError as exc:
        code = exc.response.status_code
        return {
            "error": (
                f"Status update failed for node '{node_id}' (HTTP {code}): "
                f"check the node exists and status is one of "
                f"'pending', 'in_progress', 'complete'"
            )
        }
    except requests.RequestException as exc:
        return {"error": f"Could not reach task-tree backend: {exc}"}


# ---------------------------------------------------------------------------
# update_node_diff
# ---------------------------------------------------------------------------

@tool
def update_node_diff(node_id: str, diff: str) -> dict[str, Any]:
    """Write a unified-diff string to a completed task node.

    Call this after you have implemented the task described by the node.
    The diff must represent the code changes needed to fulfil the node's
    task against the original (base) file. After this call succeeds, call
    update_node_status(node_id, "complete") to finalise the node.

    Correct coder_agent workflow:
        1. get_node(node_id)                         — read the task
        2. update_node_status(node_id, "in_progress")
        3. Implement the task, produce a unified diff string
        4. update_node_diff(node_id, diff)            ← this call
        5. update_node_status(node_id, "complete")   — ONLY after step 4 succeeds

    Diff format requirements:
        - Standard unified-diff format (output of `git diff` or
          `difflib.unified_diff`)
        - Must begin with --- / +++ header lines
        - Must contain one or more @@ hunk markers
        - Must touch ONLY the code required by the task description
        - An empty string is acceptable only if the task genuinely requires
          no file changes

    IMPORTANT: Only coder_agent should call this tool. manager_agent and
    merge_agent never write diffs — only read them.

    Args:
        node_id: ID of the node to write the diff to. The node must exist
                 and should be in "in_progress" status when this is called.
        diff:    The unified-diff string representing all code changes
                 needed to complete the node's task.

    Returns:
        The updated NodeResponse dict with the diff field populated,
        or {"error": "<message>"} if the node is not found.
    """
    try:
        resp = requests.post(
            f"{_BASE}/tree/{node_id}/diff",
            json={"diff": diff},
            timeout=_TIMEOUT,
        )
        resp.raise_for_status()
        return resp.json()
    except requests.HTTPError as exc:
        code = exc.response.status_code
        return {
            "error": (
                f"Diff write failed for node '{node_id}' (HTTP {code}): "
                f"check the node exists"
            )
        }
    except requests.RequestException as exc:
        return {"error": f"Could not reach task-tree backend: {exc}"}


# ---------------------------------------------------------------------------
# create_node
# ---------------------------------------------------------------------------

@tool
def create_node(
    node_id: str,
    task: str,
    parent_id: str | None = None,
    agent_id: str = "",
) -> dict[str, Any]:
    """Create a new task node in the orchestration tree.

    Use this during task decomposition to build the tree BEFORE dispatching
    any coder agents. Each node represents one discrete, independently
    implementable subtask. Build the entire tree first, then dispatch agents.

    Guidelines for task decomposition:
    - Each leaf node should map to a single, self-contained code change
      (e.g. one function, one class, one endpoint)
    - Sibling nodes (same parent_id) will be merged after all complete —
      design them to touch NON-OVERLAPPING lines of the target file
    - Choose node_id strings that are human-readable and unique, using
      kebab-case: e.g. "feat-auth-login-handler", "feat-auth-jwt-middleware"
    - The task string must be precise enough that a coding agent can produce
      a correct diff without any further clarification

    IMPORTANT: Only manager_agent should call this tool.

    Args:
        node_id:   Unique string identifier for this node. Must not already
                   exist in the tree. Use descriptive kebab-case names that
                   reflect the feature and subtask, e.g. "feat-cart-add-item".
        task:      Human-readable description of exactly what code must be
                   written or changed. Be precise: include function names,
                   file paths, and expected behaviour. This is the sole input
                   the coder_agent reads before implementing.
        parent_id: ID of the parent node, or None if this is a root node.
                   All nodes with the same parent_id are siblings and will
                   be merged together when all are complete.
        agent_id:  Optional agent instance identifier. Leave empty ("") to
                   let the manager assign dynamically during dispatch.

    Returns:
        The newly created NodeResponse dict, or {"error": "<message>"} if
        creation fails (e.g. node_id already exists, HTTP 4xx).
    """
    try:
        resp = requests.post(
            f"{_BASE}/tree",
            json={
                "node_id": node_id,
                "task": task,
                "parent_id": parent_id,
                "agent_id": agent_id,
            },
            timeout=_TIMEOUT,
        )
        resp.raise_for_status()
        return resp.json()
    except requests.HTTPError as exc:
        code = exc.response.status_code
        return {
            "error": (
                f"Node creation failed (HTTP {code}): "
                f"node_id '{node_id}' may already exist, or request is malformed"
            )
        }
    except requests.RequestException as exc:
        return {"error": f"Could not reach task-tree backend: {exc}"}


# ---------------------------------------------------------------------------
# trigger_merge
# ---------------------------------------------------------------------------

@tool
def trigger_merge(base_content: str, diffs: list[str]) -> dict[str, Any]:
    """Apply an ordered list of unified diffs sequentially to a base file.

    This is the FINAL step after all sibling nodes are confirmed complete.
    The merge engine applies diffs one at a time: the output of applying
    diffs[0] becomes the base for diffs[1], and so on. Only call this after
    ALL sibling nodes in the group have status "complete" and non-null diffs.

    Workflow for merge_agent:
        1. For each node_id, call get_node(node_id) and collect the diff field
        2. Verify every diff is non-null (stop if any is missing)
        3. Call trigger_merge(base_content, diffs=[...in order...])
        4. Return the result to the caller

    Diff ordering guidance:
        - Apply diffs in the order they were created by manager_agent (the
          order nodes appear in the decomposition tree, top-to-bottom)
        - Diffs from sibling nodes must touch non-overlapping code regions
          to merge cleanly — if they overlap, conflicts will be reported

    IMPORTANT: Only call this when ALL siblings are "complete". A single
    incomplete node must block the merge for the entire group.

    Args:
        base_content: The original file content as a string, before any
                      agent changes. This is the common baseline all sibling
                      diffs were authored against.
        diffs:        Ordered list of unified-diff strings to apply. Collect
                      these by calling get_node for each sibling and reading
                      the `diff` field. Order must match intended application
                      sequence.

    Returns:
        A dict with the following keys:
            success         (bool)  — True if all diffs applied without conflict
            merged_content  (str)   — Final file content after all diffs applied
            conflicts       (list)  — List of conflict dicts, each with:
                                        diff_index (int)  — which diff failed
                                        reason     (str)  — why it failed

        On network/server failure, returns {"error": "<message>"}.
    """
    try:
        resp = requests.post(
            f"{_BASE}/merge",
            json={"base_content": base_content, "diffs": diffs},
            timeout=_TIMEOUT,
        )
        resp.raise_for_status()
        return resp.json()
    except requests.HTTPError as exc:
        code = exc.response.status_code
        return {
            "error": (
                f"Merge request failed (HTTP {code}): "
                f"check that base_content is a non-empty string and "
                f"diffs is a non-empty list"
            )
        }
    except requests.RequestException as exc:
        return {"error": f"Could not reach task-tree backend: {exc}"}
