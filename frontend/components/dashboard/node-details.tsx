"use client"

import type { TopologyNode } from "@/lib/types"
import { StatusBadge } from "@/components/dashboard/status-badge"

const TYPE_LABELS: Record<string, string> = {
  orchestrator: "ORCHESTRATOR",
  mcp: "MCP SERVER",
  backend: "SERVICE",
  agent: "AGENT",
  external: "EXTERNAL",
}

export function NodeDetails({
  node,
  onClose,
}: {
  node: TopologyNode
  onClose: () => void
}) {
  return (
    <div className="border-2 border-foreground bg-background">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 border-b-2 border-foreground">
        <span className="text-[10px] tracking-[0.2em] uppercase text-muted-foreground font-mono">
          NODE DETAILS
        </span>
        <button
          type="button"
          onClick={onClose}
          className="text-xs font-mono text-muted-foreground hover:text-foreground transition-colors"
        >
          [X]
        </button>
      </div>

      <div className="p-4 space-y-4">
        {/* ID */}
        <div>
          <span className="text-[10px] tracking-[0.2em] uppercase text-muted-foreground font-mono block mb-1">
            ID
          </span>
          <span className="text-sm font-mono font-bold">{node.id}</span>
        </div>

        {/* Label */}
        <div>
          <span className="text-[10px] tracking-[0.2em] uppercase text-muted-foreground font-mono block mb-1">
            LABEL
          </span>
          <span className="text-sm font-mono">{node.label}</span>
        </div>

        {/* Type */}
        <div>
          <span className="text-[10px] tracking-[0.2em] uppercase text-muted-foreground font-mono block mb-1">
            TYPE
          </span>
          <span className="text-xs font-mono border border-foreground px-2 py-0.5">
            {TYPE_LABELS[node.type] || node.type}
          </span>
        </div>

        {/* Status */}
        <div>
          <span className="text-[10px] tracking-[0.2em] uppercase text-muted-foreground font-mono block mb-1">
            STATUS
          </span>
          <StatusBadge status={node.status} />
        </div>

        {/* Metadata */}
        {Object.keys(node.meta).length > 0 && (
          <div>
            <span className="text-[10px] tracking-[0.2em] uppercase text-muted-foreground font-mono block mb-2">
              METADATA
            </span>
            <div className="border border-border">
              {Object.entries(node.meta).map(([key, value]) => (
                <div
                  key={key}
                  className="flex items-center justify-between px-3 py-1.5 border-b border-border last:border-none"
                >
                  <span className="text-[10px] font-mono text-muted-foreground uppercase">
                    {key}
                  </span>
                  <span className="text-xs font-mono text-foreground">
                    {String(value)}
                  </span>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
