"use client"

import { useCallback, useEffect, useLayoutEffect, useRef, useState } from "react"
import { fetchTopology, ApiClientError } from "@/lib/api"
import { useSSE } from "@/hooks/use-sse"
import type { SSEEvent } from "@/hooks/use-sse"
import type { TopologyData, TopologyNode, TopologyNodeType } from "@/lib/types"
import { StatusDot } from "@/components/dashboard/status-badge"
import { NodeDetails } from "@/components/dashboard/node-details"
import { cn } from "@/lib/utils"

const TYPE_STYLES: Record<TopologyNodeType, { border: string; bg: string }> = {
  orchestrator: { border: "border-[#ea580c]", bg: "bg-[#ea580c]/10" },
  mcp: { border: "border-foreground", bg: "bg-muted/50" },
  backend: { border: "border-foreground", bg: "bg-muted/50" },
  agent: { border: "border-foreground/60", bg: "bg-background" },
  external: { border: "border-muted-foreground/40", bg: "bg-muted/30" },
}

const AGENT_TYPE_COLORS: Record<string, string> = {
  manager: "bg-purple-500",
  coder: "bg-blue-500",
  merge: "bg-amber-500",
  reviewer: "bg-green-500",
  external: "bg-gray-500",
}

const NODE_WIDTH = 170
const NODE_HEIGHT = 60

interface PositionedNode {
  node: TopologyNode
  x: number
  y: number
  cx: number
  cy: number
}

function layoutNodes(
  nodes: TopologyNode[],
  containerW: number,
  containerH: number,
): PositionedNode[] {
  const centerX = containerW / 2
  const centerY = containerH / 2

  const orca = nodes.find((n) => n.id === "orca")
  const mcp = nodes.find((n) => n.id === "mcp")
  const backend = nodes.find((n) => n.id === "backend")
  const watsonx = nodes.find((n) => n.id === "watsonx")
  const agentNodes = nodes.filter(
    (n) => !["orca", "mcp", "backend", "watsonx"].includes(n.id),
  )

  const positioned: PositionedNode[] = []

  function add(node: TopologyNode | undefined, x: number, y: number) {
    if (!node) return
    positioned.push({
      node,
      x: x - NODE_WIDTH / 2,
      y: y - NODE_HEIGHT / 2,
      cx: x,
      cy: y,
    })
  }

  add(orca, centerX, centerY)
  add(mcp, centerX + 280, centerY)
  add(backend, centerX - 280, centerY)
  add(watsonx, centerX + 280, centerY - 150)

  const agentCx = centerX + 280
  const agentCy = centerY
  const radius = 180
  agentNodes.forEach((agent, i) => {
    const count = agentNodes.length
    const startAngle = -Math.PI / 2
    const angle = startAngle + ((i + 0.5) / Math.max(count, 1)) * Math.PI
    const ax = agentCx + radius * Math.cos(angle)
    const ay = agentCy + radius * Math.sin(angle)
    add(agent, ax, ay)
  })

  return positioned
}

export function TopologyPanel() {
  const [topology, setTopology] = useState<TopologyData | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [selectedId, setSelectedId] = useState<string | null>(null)
  const [containerSize, setContainerSize] = useState({ width: 900, height: 600 })
  const containerRef = useRef<HTMLDivElement>(null)

  const load = useCallback(async () => {
    try {
      const data = await fetchTopology()
      setTopology(data)
      setError(null)
    } catch (err) {
      setError(
        err instanceof ApiClientError ? err.message : "Failed to fetch topology",
      )
    }
  }, [])

  // SSE triggers a re-fetch when topology-relevant events arrive
  const handleSSE = useCallback(
    (evt: SSEEvent) => {
      if (
        evt.type === "vfs_update" ||
        evt.type === "agent_registered" ||
        evt.type === "agent_removed"
      ) {
        load()
      }
    },
    [load],
  )

  const { connected } = useSSE({
    url: "/api/mcp/events",
    onEvent: handleSSE,
  })

  // Fallback polling when SSE is down
  useEffect(() => {
    load()
    if (!connected) {
      const id = setInterval(load, 5000)
      return () => clearInterval(id)
    }
  }, [load, connected])

  useLayoutEffect(() => {
    const el = containerRef.current
    if (!el) return
    const update = () => {
      const r = el.getBoundingClientRect()
      setContainerSize({ width: r.width, height: r.height })
    }
    update()
    const obs = new ResizeObserver(update)
    obs.observe(el)
    return () => obs.disconnect()
  }, [])

  const selectedNode = topology?.nodes.find((n) => n.id === selectedId) ?? null
  const positions = topology
    ? layoutNodes(topology.nodes, containerSize.width, containerSize.height)
    : []
  const posMap = new Map(positions.map((p) => [p.node.id, p]))

  return (
    <div className="flex h-full">
      {/* Graph area */}
      <div
        ref={containerRef}
        className="flex-1 relative overflow-hidden"
        onClick={() => setSelectedId(null)}
        style={{
          backgroundImage:
            "radial-gradient(circle, hsl(var(--border)) 1px, transparent 1px)",
          backgroundSize: "28px 28px",
        }}
      >
        {/* Header overlay */}
        <div className="absolute top-0 left-0 right-0 z-10 flex items-center gap-4 px-5 py-3">
          <span className="text-[10px] tracking-[0.2em] uppercase text-muted-foreground font-mono">
            // TOPOLOGY: AGENT_GRAPH
          </span>
          <div className="flex-1 border-t border-border" />
          <span
            className={cn(
              "text-[10px] tracking-[0.2em] uppercase font-mono",
              connected ? "text-green-500" : "text-muted-foreground",
            )}
          >
            {connected ? "SSE" : "POLL"}
          </span>
          <StatusDot status={error ? "offline" : "online"} />
          <span className="text-[10px] tracking-[0.2em] uppercase text-muted-foreground font-mono">
            {topology ? `${topology.nodes.length} NODES` : "LOADING"}
          </span>
        </div>

        {error && (
          <div className="absolute inset-0 flex items-center justify-center">
            <div className="border-2 border-red-500/30 p-6 bg-background/80">
              <p className="text-xs font-mono text-red-400">{error}</p>
            </div>
          </div>
        )}

        {/* Edges */}
        <svg className="absolute inset-0 w-full h-full pointer-events-none">
          {topology?.edges.map((edge) => {
            const src = posMap.get(edge.source)
            const tgt = posMap.get(edge.target)
            if (!src || !tgt) return null
            return (
              <line
                key={`${edge.source}-${edge.target}`}
                x1={src.cx}
                y1={src.cy}
                x2={tgt.cx}
                y2={tgt.cy}
                stroke="hsl(var(--border))"
                strokeWidth={1.5}
                strokeDasharray="6 4"
              />
            )
          })}
          {topology?.edges.map((edge, i) => {
            const src = posMap.get(edge.source)
            const tgt = posMap.get(edge.target)
            if (!src || !tgt) return null
            return (
              <circle
                key={`pkt-${edge.source}-${edge.target}`}
                r={3}
                fill="#ea580c"
                opacity={0.7}
              >
                <animate
                  attributeName="cx"
                  values={`${src.cx};${tgt.cx}`}
                  dur="2.5s"
                  begin={`${i * 0.4}s`}
                  repeatCount="indefinite"
                />
                <animate
                  attributeName="cy"
                  values={`${src.cy};${tgt.cy}`}
                  dur="2.5s"
                  begin={`${i * 0.4}s`}
                  repeatCount="indefinite"
                />
              </circle>
            )
          })}
        </svg>

        {/* Nodes */}
        {positions.map((pos) => {
          const agentType = pos.node.meta?.agent_type as string | undefined
          const typeDot = agentType ? AGENT_TYPE_COLORS[agentType] : null

          return (
            <div
              key={pos.node.id}
              className="absolute"
              style={{
                left: pos.x,
                top: pos.y,
                width: NODE_WIDTH,
                height: NODE_HEIGHT,
              }}
            >
              <button
                type="button"
                onClick={(e) => {
                  e.stopPropagation()
                  setSelectedId(pos.node.id)
                }}
                className={cn(
                  "w-full h-full border-2 flex flex-col items-center justify-center transition-all",
                  TYPE_STYLES[pos.node.type].border,
                  TYPE_STYLES[pos.node.type].bg,
                  "hover:bg-muted/60",
                  selectedId === pos.node.id && "ring-2 ring-[#ea580c]",
                )}
              >
                <span className="text-[10px] font-mono font-bold tracking-wide uppercase truncate px-2">
                  {pos.node.label}
                </span>
                <div className="flex items-center gap-1.5 mt-1">
                  {typeDot && (
                    <span className={cn("h-1.5 w-1.5 rounded-full", typeDot)} />
                  )}
                  <StatusDot status={pos.node.status} />
                  <span className="text-[9px] font-mono text-muted-foreground uppercase">
                    {pos.node.status}
                  </span>
                </div>
              </button>
            </div>
          )
        })}

        {/* Legend */}
        <div className="absolute bottom-3 left-5 flex items-center gap-4">
          {Object.entries(AGENT_TYPE_COLORS).map(([type, color]) => (
            <div key={type} className="flex items-center gap-1.5">
              <span className={cn("h-2 w-2 rounded-full", color)} />
              <span className="text-[9px] font-mono text-muted-foreground uppercase">
                {type}
              </span>
            </div>
          ))}
        </div>
      </div>

      {/* Side panel */}
      {selectedNode && (
        <div className="w-[300px] border-l-2 border-foreground overflow-auto">
          <NodeDetails
            node={selectedNode}
            onClose={() => setSelectedId(null)}
          />
        </div>
      )}
    </div>
  )
}
