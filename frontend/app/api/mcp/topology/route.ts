import { NextResponse } from "next/server";
import type { TopologyData, TopologyNode, TopologyEdge, AgentInfo } from "@/lib/types";

/**
 * GET /api/mcp/topology
 *
 * Constructs the agent/MCP topology by fetching:
 *  - Agent registry from MCP (first-class agent types)
 *  - VFS state for file counts
 *  - Health checks for MCP and backend
 *
 * Agent types now come from the MCP agent registry instead of
 * name-based heuristics.
 */
export async function GET() {
  const mcpUrl = process.env.MCP_BASE_URL?.replace(/\/+$/, "");
  const backendUrl = process.env.BACKEND_BASE_URL?.replace(/\/+$/, "");

  const nodes: TopologyNode[] = [];
  const edges: TopologyEdge[] = [];

  nodes.push({
    id: "orca",
    label: "ORCA AI",
    type: "orchestrator",
    status: "online",
    meta: { role: "orchestrator" },
  });

  // MCP health
  let mcpOnline = false;
  if (mcpUrl) {
    try {
      const res = await fetch(`${mcpUrl}/git/health`, {
        signal: AbortSignal.timeout(5000),
      });
      mcpOnline = res.ok;
    } catch {
      /* offline */
    }
  }
  nodes.push({
    id: "mcp",
    label: "MCP SERVER",
    type: "mcp",
    status: mcpOnline ? "online" : "offline",
    meta: { url: mcpUrl || "not configured" },
  });
  edges.push({ source: "orca", target: "mcp", label: "orchestrate" });

  // Backend health
  let backendOnline = false;
  if (backendUrl) {
    try {
      const res = await fetch(`${backendUrl}/health`, {
        signal: AbortSignal.timeout(5000),
      });
      backendOnline = res.ok;
    } catch {
      /* offline */
    }
  }
  nodes.push({
    id: "backend",
    label: "ANALYSIS BACKEND",
    type: "backend",
    status: backendOnline ? "online" : "offline",
    meta: { url: backendUrl || "not configured" },
  });
  edges.push({ source: "orca", target: "backend", label: "analyze" });

  // Agent nodes — prefer registry (first-class types), fall back to VFS
  if (mcpUrl) {
    let registryAgents: AgentInfo[] = [];

    try {
      const res = await fetch(`${mcpUrl}/dashboard/agents`, {
        signal: AbortSignal.timeout(5000),
      });
      if (res.ok) {
        registryAgents = await res.json();
      }
    } catch {
      /* registry unavailable */
    }

    // Build file-count map from VFS
    const fileCounts = new Map<string, number>();
    try {
      const res = await fetch(`${mcpUrl}/dashboard/vfs`, {
        signal: AbortSignal.timeout(5000),
      });
      if (res.ok) {
        const vfs = await res.json();
        const changes: Array<{ agent_id: string; files: unknown[] }> =
          vfs.pending_changes || [];
        for (const pc of changes) {
          fileCounts.set(pc.agent_id, pc.files?.length ?? 0);
        }
      }
    } catch {
      /* VFS fetch failed */
    }

    const TYPE_LABELS: Record<string, string> = {
      manager: "MGR",
      coder: "CODER",
      merge: "MERGE",
      reviewer: "REVIEW",
      external: "EXT",
    };

    if (registryAgents.length > 0) {
      for (const agent of registryAgents) {
        const prefix = TYPE_LABELS[agent.type] || "AGENT";
        const status = agent.status === "active" ? "active" as const
                     : agent.status === "idle" ? "idle" as const
                     : "offline" as const;

        nodes.push({
          id: agent.id,
          label: `${prefix}: ${agent.display_name}`,
          type: "agent",
          status,
          meta: {
            agent_type: agent.type,
            files: fileCounts.get(agent.id) ?? 0,
            ...(agent.meta || {}),
          },
        });
        edges.push({ source: agent.id, target: "mcp", label: "push" });
      }
    } else {
      // Fallback: derive agents from VFS when registry is empty
      for (const [agentId, count] of fileCounts) {
        nodes.push({
          id: agentId,
          label: `AGENT: ${agentId}`,
          type: "agent",
          status: "active",
          meta: { files: count, agent_type: "unknown" },
        });
        edges.push({ source: agentId, target: "mcp", label: "push" });
      }
    }
  }

  nodes.push({
    id: "watsonx",
    label: "WATSONX ORCHESTRATE",
    type: "external",
    status: "unknown",
    meta: { role: "external orchestrator" },
  });
  edges.push({ source: "watsonx", target: "mcp", label: "SSE" });

  const topology: TopologyData = { nodes, edges };
  return NextResponse.json(topology);
}
