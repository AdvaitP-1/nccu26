/**
 * Typed models mirroring the backend Pydantic schemas.
 *
 * Sources:
 *   - backend/app/schemas.py       (overlap analysis)
 *   - backend/app/schemas_tree.py  (tree + merge)
 */

// ---------------------------------------------------------------------------
// Health
// ---------------------------------------------------------------------------

export interface HealthResponse {
  status: string;
}

// ---------------------------------------------------------------------------
// Overlap Analysis
// ---------------------------------------------------------------------------

export interface FileInput {
  path: string;
  language: string;
  content: string;
}

export interface ChangeSet {
  agent_id: string;
  files: FileInput[];
}

export interface AnalyzeOverlapsRequest {
  changesets: ChangeSet[];
}

export type OverlapSeverity = "critical" | "high" | "medium" | "low";

export interface Overlap {
  file_path: string;
  symbol_name: string;
  symbol_kind: string;
  agent_a: string;
  agent_b: string;
  severity: OverlapSeverity;
  reason: string;
  start_line_a: number;
  end_line_a: number;
  start_line_b: number;
  end_line_b: number;
}

export interface FileRisk {
  file_path: string;
  risk_score: number;
  stability_score: number;
  overlap_count: number;
  contributors: string[];
  contributors_count: number;
  pairwise_overlap_count: number;
  max_severity: OverlapSeverity | null;
  is_hotspot: boolean;
  summary: string;
}

export interface AnalyzeOverlapsResponse {
  overlaps: Overlap[];
  file_risks: FileRisk[];
}

// ---------------------------------------------------------------------------
// Tree
// ---------------------------------------------------------------------------

export type NodeStatus = "pending" | "in_progress" | "complete";

export interface TreeNodeResponse {
  node_id: string;
  parent_id: string | null;
  task: string;
  agent_id: string;
  status: NodeStatus;
  diff: string | null;
}

export interface CreateNodeRequest {
  node_id: string;
  parent_id?: string | null;
  task: string;
  agent_id?: string;
}

export interface UpdateStatusRequest {
  status: NodeStatus;
}

export interface WriteDiffRequest {
  diff: string;
}

export interface SiblingsResponse {
  node_id: string;
  siblings: TreeNodeResponse[];
}

// ---------------------------------------------------------------------------
// Merge
// ---------------------------------------------------------------------------

export interface MergeRequest {
  base_content: string;
  diffs: string[];
}

export interface MergeConflict {
  diff_index: number;
  reason: string;
}

export interface MergeResponse {
  success: boolean;
  merged_content: string;
  conflicts: MergeConflict[];
}

// ---------------------------------------------------------------------------
// Shared error shape returned by route handlers
// ---------------------------------------------------------------------------

export interface ApiError {
  error: string;
  details?: string;
}

// ---------------------------------------------------------------------------
// VFS (Virtual File System) — mirrors MCP models
// ---------------------------------------------------------------------------

export interface VfsFileSnapshot {
  path: string;
  language: string;
  content: string;
}

export interface VfsPendingChange {
  agent_id: string;
  session_id?: string;
  task_id?: string;
  files: VfsFileSnapshot[];
  created_at: string;
  updated_at: string;
}

export interface VfsState {
  pending_changes: VfsPendingChange[];
  total_files: number;
  total_agents: number;
}

// ---------------------------------------------------------------------------
// MCP Command Interface
// ---------------------------------------------------------------------------

export interface McpCommandRequest {
  command: string;
  args: Record<string, unknown>;
}

export interface McpCommandResponse {
  success: boolean;
  data?: unknown;
  error?: string;
}

// ---------------------------------------------------------------------------
// Dashboard Topology
// ---------------------------------------------------------------------------

export type TopologyNodeType =
  | "orchestrator"
  | "mcp"
  | "backend"
  | "agent"
  | "external";

export type TopologyNodeStatus = "online" | "active" | "idle" | "offline" | "unknown";

export interface TopologyNode {
  id: string;
  label: string;
  type: TopologyNodeType;
  status: TopologyNodeStatus;
  meta: Record<string, string | number>;
}

export interface TopologyEdge {
  source: string;
  target: string;
  label?: string;
}

export interface TopologyData {
  nodes: TopologyNode[];
  edges: TopologyEdge[];
}

// ---------------------------------------------------------------------------
// Agent Registry
// ---------------------------------------------------------------------------

export type AgentType = "manager" | "coder" | "merge" | "reviewer" | "external";

export type AgentStatus = "active" | "idle" | "disconnected";

export interface AgentInfo {
  id: string;
  type: AgentType;
  display_name: string;
  status: AgentStatus;
  meta?: Record<string, string>;
  connected_at: string;
  last_seen_at: string;
}

// ---------------------------------------------------------------------------
// SSE Events
// ---------------------------------------------------------------------------

export type SSEEventType =
  | "vfs_update"
  | "agent_registered"
  | "agent_removed"
  | "command_result";

export interface SSEEvent {
  type: SSEEventType;
  timestamp: string;
  data: unknown;
}

// ---------------------------------------------------------------------------
// Terminal History
// ---------------------------------------------------------------------------

export interface CommandHistoryEntry {
  id: string;
  input: string;
  output: string;
  success: boolean;
  timestamp: number;
}
