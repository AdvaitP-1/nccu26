"use client"

import { Node } from "@/components/graph/Node"
import type { Node as GraphNode } from "@/components/graph/graphHelpers"
import { useLayoutEffect, useRef, useState } from "react"

interface GraphProps {
  nodes: GraphNode[]
  selectedNodeId: string | null
  onSelectNode: (id: string) => void
  onDeselect: () => void
  topInset?: number
}

function isManagerNode(node: GraphNode) {
  return node.id === "manager" || node.label.toLowerCase() === "manager"
}

const NODE_SIZES = {
  manager: { width: 140, height: 72 },
  normal: { width: 110, height: 56 },
}

export function Graph({ nodes, selectedNodeId, onSelectNode, onDeselect, topInset = 0 }: GraphProps) {
  const containerRef = useRef<HTMLDivElement | null>(null)
  const [containerSize, setContainerSize] = useState({ width: 900, height: 520 })

  useLayoutEffect(() => {
    const element = containerRef.current
    if (!element) return

    const updateSize = () => {
      const rect = element.getBoundingClientRect()
      setContainerSize({ width: rect.width, height: rect.height })
    }

    updateSize()
    const observer = new ResizeObserver(updateSize)
    observer.observe(element)

    return () => observer.disconnect()
  }, [])

  const centerX = containerSize.width / 2
  const centerY = topInset + (containerSize.height - topInset) / 2

  const managerNode = nodes.find(isManagerNode) || null
  const otherNodes = nodes.filter((node) => !isManagerNode(node))

  const positions = nodes.map((node) => {
    const size = isManagerNode(node) ? NODE_SIZES.manager : NODE_SIZES.normal
    if (managerNode && node.id === managerNode.id) {
      const x = centerX - size.width / 2
      const y = centerY - size.height / 2
      return {
        id: node.id,
        x,
        y,
        width: size.width,
        height: size.height,
        centerX,
        centerY,
      }
    }

    const index = otherNodes.findIndex((item) => item.id === node.id)
    const ringSize = 6
    const ringIndex = Math.floor(index / ringSize)
    const positionInRing = index % ringSize
    const radius = 160 + ringIndex * 120
    const angle = (positionInRing / ringSize) * Math.PI * 2 - Math.PI / 2
    const x = centerX + radius * Math.cos(angle) - size.width / 2
    const y = centerY + radius * Math.sin(angle) - size.height / 2
    return {
      id: node.id,
      x,
      y,
      width: size.width,
      height: size.height,
      centerX: x + size.width / 2,
      centerY: y + size.height / 2,
    }
  })

  const positionMap = new Map(positions.map((pos) => [pos.id, pos]))

  return (
    <div
      ref={containerRef}
      className="relative h-full w-full overflow-hidden rounded-2xl border border-dashed border-blue-400/40 bg-[radial-gradient(circle_at_1px_1px,_rgba(59,130,246,0.35)_1px,_transparent_0)] [background-size:28px_28px]"
      onClick={onDeselect}
    >
      <svg className="pointer-events-none absolute inset-0 h-full w-full">
        {nodes.flatMap((node) =>
          node.connectsTo.map((targetId) => {
            const sourcePos = positionMap.get(node.id)
            const targetPos = positionMap.get(targetId)
            if (!sourcePos || !targetPos) {
              return null
            }
            return (
              <line
                key={`${node.id}-${targetId}`}
                x1={sourcePos.centerX}
                y1={sourcePos.centerY}
                x2={targetPos.centerX}
                y2={targetPos.centerY}
                stroke="#60a5fa"
                strokeWidth={2}
                strokeDasharray="6 6"
              />
            )
          })
        )}
      </svg>

      {nodes.map((node) => {
        const pos = positionMap.get(node.id)
        if (!pos) {
          return null
        }
        return (
          <div key={node.id} className="absolute" style={{ left: pos.x, top: pos.y }}>
            <Node node={node} isSelected={node.id === selectedNodeId} onClick={onSelectNode} />
          </div>
        )
      })}
    </div>
  )
}
