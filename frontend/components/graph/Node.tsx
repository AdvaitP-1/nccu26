"use client"

import type { MouseEvent } from "react"
import type { Node } from "@/components/graph/graphHelpers"

interface NodeProps {
  node: Node
  isSelected: boolean
  onClick: (id: string) => void
}

function isManagerNode(node: Node) {
  return node.id === "manager" || node.label.toLowerCase() === "manager"
}

export function Node({ node, isSelected, onClick }: NodeProps) {
  const isManager = isManagerNode(node)

  function handleClick(event: MouseEvent<HTMLButtonElement>) {
    event.stopPropagation()
    onClick(node.id)
  }

  return (
    <button
      type="button"
      onClick={handleClick}
      className={[
        "rounded-full border-2 px-4 font-semibold transition-all duration-200 ease-out",
        "shadow-sm hover:-translate-y-0.5 hover:shadow-md",
        "focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-blue-400/60",
        isManager ? "h-[72px] min-w-[140px] text-base" : "h-[56px] min-w-[110px] text-sm",
        isManager
          ? "border-red-700 bg-red-600 text-white"
          : "border-blue-400/50 bg-slate-900 text-blue-100 hover:border-blue-300/70",
        isSelected ? "ring-4 ring-blue-400/70" : "ring-0",
      ].join(" ")}
      aria-pressed={isSelected}
      data-node-id={node.id}
    >
      <span className="block text-center">{node.label}</span>
    </button>
  )
}
