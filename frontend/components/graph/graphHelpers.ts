export type Node = {
  id: string
  label: string
  data: string
  connectsTo: string[]
  connectedFrom: string[]
}

function unique(values: string[]) {
  return Array.from(new Set(values))
}

export function addNode(nodes: Node[], parentId: string, newNode: Node) {
  const parent = nodes.find((node) => node.id === parentId)
  if (!parent) {
    return nodes
  }
  if (nodes.some((node) => node.id === newNode.id)) {
    return nodes
  }

  const updatedNodes = nodes.map((node) => {
    if (node.id !== parentId) {
      return node
    }
    return {
      ...node,
      connectsTo: unique([...node.connectsTo, newNode.id]),
    }
  })

  const normalizedNewNode: Node = {
    ...newNode,
    connectedFrom: unique([...newNode.connectedFrom, parentId]),
  }

  return [...updatedNodes, normalizedNewNode]
}

export function removeNode(nodes: Node[], nodeId: string) {
  const remaining = nodes.filter((node) => node.id !== nodeId)
  return remaining.map((node) => ({
    ...node,
    connectsTo: node.connectsTo.filter((id) => id !== nodeId),
    connectedFrom: node.connectedFrom.filter((id) => id !== nodeId),
  }))
}

export function connectNodes(nodes: Node[], sourceId: string, targetId: string) {
  const sourceExists = nodes.some((node) => node.id === sourceId)
  const targetExists = nodes.some((node) => node.id === targetId)
  if (!sourceExists || !targetExists) {
    return nodes
  }

  return nodes.map((node) => {
    if (node.id === sourceId) {
      return { ...node, connectsTo: unique([...node.connectsTo, targetId]) }
    }
    if (node.id === targetId) {
      return { ...node, connectedFrom: unique([...node.connectedFrom, sourceId]) }
    }
    return node
  })
}
