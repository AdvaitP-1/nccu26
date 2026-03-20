/**
 * Client-side API utilities for calling the Next.js route handlers.
 *
 * These functions run in the browser and call /api/* which are proxied
 * server-side to the backend. The browser never sees BACKEND_BASE_URL.
 */

import type {
  HealthResponse,
  AnalyzeOverlapsRequest,
  AnalyzeOverlapsResponse,
  TreeNodeResponse,
  CreateNodeRequest,
  UpdateStatusRequest,
  WriteDiffRequest,
  SiblingsResponse,
  MergeRequest,
  MergeResponse,
  ApiError,
  VfsState,
  McpCommandRequest,
  McpCommandResponse,
  TopologyData,
  AgentInfo,
} from "./types";

// ---------------------------------------------------------------------------
// Error class
// ---------------------------------------------------------------------------

export class ApiClientError extends Error {
  constructor(
    message: string,
    public status: number,
    public details?: string,
  ) {
    super(message);
    this.name = "ApiClientError";
  }
}

// ---------------------------------------------------------------------------
// Core fetch wrapper
// ---------------------------------------------------------------------------

async function apiFetch<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(path, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      ...init?.headers,
    },
  });

  if (!res.ok) {
    const err: ApiError = await res.json().catch(() => ({
      error: `HTTP ${res.status}`,
    }));
    throw new ApiClientError(
      err.error || `HTTP ${res.status}`,
      res.status,
      err.details,
    );
  }

  return res.json() as Promise<T>;
}

// ---------------------------------------------------------------------------
// Health
// ---------------------------------------------------------------------------

export function fetchHealth(): Promise<HealthResponse> {
  return apiFetch<HealthResponse>("/api/health");
}

// ---------------------------------------------------------------------------
// Overlap Analysis
// ---------------------------------------------------------------------------

export function analyzeOverlaps(
  payload: AnalyzeOverlapsRequest,
): Promise<AnalyzeOverlapsResponse> {
  return apiFetch<AnalyzeOverlapsResponse>("/api/analyze/overlaps", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

// ---------------------------------------------------------------------------
// Tree
// ---------------------------------------------------------------------------

export function createTreeNode(
  payload: CreateNodeRequest,
): Promise<TreeNodeResponse> {
  return apiFetch<TreeNodeResponse>("/api/tree", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export function getTreeNode(nodeId: string): Promise<TreeNodeResponse> {
  return apiFetch<TreeNodeResponse>(`/api/tree/${encodeURIComponent(nodeId)}`);
}

export function getTreeNodeSiblings(nodeId: string): Promise<SiblingsResponse> {
  return apiFetch<SiblingsResponse>(
    `/api/tree/${encodeURIComponent(nodeId)}/siblings`,
  );
}

export function updateTreeNodeStatus(
  nodeId: string,
  payload: UpdateStatusRequest,
): Promise<TreeNodeResponse> {
  return apiFetch<TreeNodeResponse>(
    `/api/tree/${encodeURIComponent(nodeId)}/status`,
    { method: "POST", body: JSON.stringify(payload) },
  );
}

export function writeTreeNodeDiff(
  nodeId: string,
  payload: WriteDiffRequest,
): Promise<TreeNodeResponse> {
  return apiFetch<TreeNodeResponse>(
    `/api/tree/${encodeURIComponent(nodeId)}/diff`,
    { method: "POST", body: JSON.stringify(payload) },
  );
}

// ---------------------------------------------------------------------------
// Merge
// ---------------------------------------------------------------------------

export function requestMerge(payload: MergeRequest): Promise<MergeResponse> {
  return apiFetch<MergeResponse>("/api/merge", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

// ---------------------------------------------------------------------------
// MCP / Dashboard
// ---------------------------------------------------------------------------

export function fetchVfsState(): Promise<VfsState> {
  return apiFetch<VfsState>("/api/mcp/vfs");
}

export function sendMcpCommand(
  payload: McpCommandRequest,
): Promise<McpCommandResponse> {
  return apiFetch<McpCommandResponse>("/api/mcp/command", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export function fetchMcpStatus(): Promise<McpCommandResponse> {
  return apiFetch<McpCommandResponse>("/api/mcp/status");
}

export function fetchTopology(): Promise<TopologyData> {
  return apiFetch<TopologyData>("/api/mcp/topology");
}

export function fetchAgents(): Promise<AgentInfo[]> {
  return apiFetch<AgentInfo[]>("/api/mcp/agents");
}

export function seedDemo(): Promise<McpCommandResponse> {
  return apiFetch<McpCommandResponse>("/api/mcp/command", {
    method: "POST",
    body: JSON.stringify({ command: "seed_demo", args: {} }),
  });
}

export function clearDemo(): Promise<McpCommandResponse> {
  return apiFetch<McpCommandResponse>("/api/mcp/command", {
    method: "POST",
    body: JSON.stringify({ command: "clear_demo", args: {} }),
  });
}

// ---------------------------------------------------------------------------
// MCP Server Control
// ---------------------------------------------------------------------------

export interface McpServerStatus {
  status: "stopped" | "starting" | "running" | "error";
  pid: number | null;
  logs: string[];
  reachable: boolean;
  message?: string;
}

export function fetchMcpServerStatus(): Promise<McpServerStatus> {
  return apiFetch<McpServerStatus>("/api/mcp/server");
}

export function startMcpServer(): Promise<McpServerStatus> {
  return apiFetch<McpServerStatus>("/api/mcp/server", { method: "POST" });
}
