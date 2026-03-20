"use client"

import { Navbar } from "@/components/navbar"
import { useMemo, useState } from "react"
import { Graph } from "@/components/graph/Graph"
import { SidePanel } from "@/components/graph/SidePanel"
import { BackendPanel } from "@/components/backend-panel"
import { addNode, removeNode, type Node } from "@/components/graph/graphHelpers"

interface GraphState {
  nodes: Node[]
  selectedNodeId: string | null
}

const INITIAL_NODES: Node[] = [
  {
    id: "manager",
    label: "Manager",
    data: "Primary coordinator node.",
    connectsTo: ["node-1", "node-2"],
    connectedFrom: [],
  },
  {
    id: "node-1",
    label: "Node 1",
    data: "Worker node handling task A.",
    connectsTo: [],
    connectedFrom: ["manager"],
  },
  {
    id: "node-2",
    label: "Node 2",
    data: "Worker node handling task B.",
    connectsTo: [],
    connectedFrom: ["manager"],
  },
]

function getNextNodeId(nodes: Node[]) {
  let index = nodes.length + 1
  while (nodes.some((node) => node.id === `node-${index}`)) {
    index += 1
  }
  return `node-${index}`
}

export function GraphApp() {
  const [state, setState] = useState<GraphState>({
    nodes: INITIAL_NODES,
    selectedNodeId: null,
  })
  const navHeight = 72

  const selectedNode = useMemo(() => {
    return state.nodes.find((node) => node.id === state.selectedNodeId) || null
  }, [state.nodes, state.selectedNodeId])

  function handleSelectNode(id: string) {
    setState((prev) => ({ ...prev, selectedNodeId: id }))
  }

  function handleAddNode(parentId: string) {
    setState((prev) => {
      const newId = getNextNodeId(prev.nodes)
      const newNode: Node = {
        id: newId,
        label: `Node ${newId.split("-")[1]}`,
        data: "",
        connectsTo: [],
        connectedFrom: [],
      }

      return {
        ...prev,
        nodes: addNode(prev.nodes, parentId, newNode),
      }
    })
  }

  function handleRemoveNode(nodeId: string) {
    setState((prev) => ({
      nodes: removeNode(prev.nodes, nodeId),
      selectedNodeId: prev.selectedNodeId === nodeId ? null : prev.selectedNodeId,
    }))
  }

  function handleEditNode(nodeId: string, data: string) {
    setState((prev) => ({
      ...prev,
      nodes: prev.nodes.map((node) => (node.id === nodeId ? { ...node, data } : node)),
    }))
  }

  function handleDeselect() {
    setState((prev) => ({ ...prev, selectedNodeId: null }))
  }

  return (
    <div className="min-h-screen bg-slate-950 text-blue-50">
      <div className="absolute left-0 right-0 top-0 z-20">
        <Navbar />
      </div>
      <main className="relative h-screen w-full">
        <div className="absolute inset-0">
          <Graph
            nodes={state.nodes}
            selectedNodeId={state.selectedNodeId}
            onSelectNode={handleSelectNode}
            onDeselect={handleDeselect}
            topInset={navHeight}
          />
        </div>
        {selectedNode ? (
          <aside className="absolute right-6 top-6 w-[320px]">
            <SidePanel
              selectedNode={selectedNode}
              onAddNode={handleAddNode}
              onRemoveNode={handleRemoveNode}
              onEditNode={handleEditNode}
            />
          </aside>
        ) : null}
      </main>
      <BackendPanel />
    </div>
  )
}
