"use client"

import { motion } from "framer-motion"
import { useEffect, useState } from "react"

interface Node {
  id: string
  label: string
  x: number
  y: number
}

const NODES: Node[] = [
  { id: "task", label: "TASK", x: 80, y: 180 },
  { id: "decompose", label: "DECOMPOSE", x: 260, y: 80 },
  { id: "agents", label: "AGENTS", x: 440, y: 180 },
  { id: "analyze", label: "ANALYZE", x: 260, y: 280 },
  { id: "merge", label: "MERGE", x: 620, y: 180 },
]

const EDGES: [string, string][] = [
  ["task", "decompose"],
  ["decompose", "agents"],
  ["task", "analyze"],
  ["analyze", "agents"],
  ["agents", "merge"],
]

function getNode(id: string) {
  return NODES.find((n) => n.id === id)!
}

function DataPacket({ x1, y1, x2, y2, delay }: { x1: number; y1: number; x2: number; y2: number; delay: number }) {
  return (
    <motion.circle
      r={4}
      fill="#ea580c"
      initial={{ cx: x1 + 60, cy: y1 + 20 }}
      animate={{
        cx: [x1 + 60, x2 + 60],
        cy: [y1 + 20, y2 + 20],
      }}
      transition={{
        duration: 2,
        delay,
        repeat: Infinity,
        repeatDelay: 1.5,
        ease: "linear",
      }}
    />
  )
}

export function TopologyGraph() {
  const [mounted, setMounted] = useState(false)

  useEffect(() => {
    setMounted(true)
  }, [])

  if (!mounted) {
    return <div className="h-[360px] w-full border-2 border-foreground" />
  }

  return (
    <div className="relative w-full border-2 border-foreground bg-background">
      <svg viewBox="0 0 760 360" className="w-full h-auto" role="img" aria-label="Agent orchestration topology graph showing data flow from Task through Decompose and Analyze to Agents and Merge">
        {Array.from({ length: 32 }).map((_, i) => (
          <line
            key={`vline-${i}`}
            x1={i * 24}
            y1={0}
            x2={i * 24}
            y2={360}
            stroke="hsl(var(--border))"
            strokeWidth={0.5}
          />
        ))}
        {Array.from({ length: 16 }).map((_, i) => (
          <line
            key={`hline-${i}`}
            x1={0}
            y1={i * 24}
            x2={760}
            y2={i * 24}
            stroke="hsl(var(--border))"
            strokeWidth={0.5}
          />
        ))}

        {EDGES.map(([fromId, toId], i) => {
          const from = getNode(fromId)
          const to = getNode(toId)
          return (
            <motion.line
              key={`edge-${i}`}
              x1={from.x + 60}
              y1={from.y + 20}
              x2={to.x + 60}
              y2={to.y + 20}
              stroke="hsl(var(--foreground))"
              strokeWidth={1.5}
              strokeDasharray="6 3"
              initial={{ pathLength: 0, opacity: 0 }}
              animate={{ pathLength: 1, opacity: 1 }}
              transition={{ duration: 0.8, delay: i * 0.15, ease: "linear" }}
            />
          )
        })}

        {EDGES.map(([fromId, toId], i) => {
          const from = getNode(fromId)
          const to = getNode(toId)
          return (
            <DataPacket
              key={`packet-${i}`}
              x1={from.x}
              y1={from.y}
              x2={to.x}
              y2={to.y}
              delay={i * 0.4}
            />
          )
        })}

        {NODES.map((node, i) => (
          <motion.g
            key={node.id}
            initial={{ opacity: 0, scale: 0.8 }}
            animate={{ opacity: 1, scale: 1 }}
            transition={{ duration: 0.01, delay: i * 0.1 }}
          >
            <rect
              x={node.x}
              y={node.y}
              width={120}
              height={40}
              fill="hsl(var(--foreground))"
              stroke="hsl(var(--foreground))"
              strokeWidth={2}
            />
            <text
              x={node.x + 60}
              y={node.y + 24}
              textAnchor="middle"
              fill="hsl(var(--background))"
              fontSize={11}
              fontFamily="var(--font-mono), monospace"
              fontWeight={600}
              letterSpacing="0.1em"
            >
              {node.label}
            </text>
            <circle cx={node.x + 110} cy={node.y + 10} r={3} fill="#ea580c">
              <animate
                attributeName="opacity"
                values="1;0.3;1"
                dur="2s"
                begin={`${i * 0.3}s`}
                repeatCount="indefinite"
              />
            </circle>
          </motion.g>
        ))}
      </svg>
    </div>
  )
}
