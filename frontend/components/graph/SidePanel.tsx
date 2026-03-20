"use client"

import { useEffect, useState } from "react"
import type { Node } from "@/components/graph/graphHelpers"

interface SidePanelProps {
  selectedNode: Node | null
  onAddNode: (parentId: string) => void
  onRemoveNode: (nodeId: string) => void
  onEditNode: (nodeId: string, data: string) => void
}

export function SidePanel({ selectedNode, onAddNode, onRemoveNode, onEditNode }: SidePanelProps) {
  const [dataValue, setDataValue] = useState("")

  useEffect(() => {
    setDataValue(selectedNode?.data ?? "")
  }, [selectedNode])

  if (!selectedNode) {
    return null
  }

  return (
    <div className="rounded-2xl border border-blue-400/30 bg-blue-950/80 p-6 text-blue-50 shadow-lg shadow-blue-950/40 backdrop-blur my-16">
      <h2 className="text-lg font-semibold">Node Details</h2>
      <p className="mt-1 text-sm text-blue-200/70">ID: {selectedNode.id}</p>

      <div className="mt-4 space-y-1">
        <p className="text-sm font-semibold text-blue-200">Label</p>
        <p className="text-sm text-blue-100/90">{selectedNode.label}</p>
      </div>

      <div className="mt-4">
        <label className="text-sm font-semibold text-blue-200" htmlFor="node-data">
          Data
        </label>
        <textarea
          id="node-data"
          className="mt-2 w-full rounded-xl border border-blue-400/40 bg-blue-950/60 px-3 py-2 text-sm text-blue-50 shadow-sm focus:border-blue-300 focus:outline-none focus:ring-2 focus:ring-blue-400/60"
          rows={4}
          value={dataValue}
          onChange={(event) => setDataValue(event.target.value)}
          onBlur={() => onEditNode(selectedNode.id, dataValue)}
        />
        <button
          type="button"
          onClick={() => onEditNode(selectedNode.id, dataValue)}
          className="mt-2 w-full rounded-full bg-blue-400 px-3 py-2 text-sm font-semibold text-blue-950 transition hover:bg-blue-300"
        >
          Save Data
        </button>
      </div>

      <div className="mt-5 space-y-2 text-sm text-blue-100/80">
        <div>
          <span className="font-semibold text-blue-200">Connected From: </span>
          {selectedNode.connectedFrom.length > 0 ? selectedNode.connectedFrom.join(", ") : "None"}
        </div>
        <div>
          <span className="font-semibold text-blue-200">Connects To: </span>
          {selectedNode.connectsTo.length > 0 ? selectedNode.connectsTo.join(", ") : "None"}
        </div>
      </div>

      <div className="mt-6 flex flex-col gap-2">
        <button
          type="button"
          onClick={() => onAddNode(selectedNode.id)}
          className="rounded-full border border-blue-400/40 px-3 py-2 text-sm font-semibold text-blue-100 transition hover:border-blue-300/70 hover:bg-blue-900/60"
        >
          Add Node
        </button>
        <button
          type="button"
          onClick={() => onRemoveNode(selectedNode.id)}
          className="rounded-full border border-red-300/60 px-3 py-2 text-sm font-semibold text-red-200 transition hover:border-red-200 hover:bg-red-900/30"
        >
          Remove Node
        </button>
      </div>
    </div>
  )
}
