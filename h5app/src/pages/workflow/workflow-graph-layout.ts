import type { WorkflowPublishedEdge, WorkflowPublishedNode } from '@/types/workflow'

export interface WorkflowGraphNodeBox {
  node: WorkflowPublishedNode
  x: number
  y: number
  width: number
  height: number
}

function centerX(item: WorkflowGraphNodeBox) {
  return item.x + item.width / 2
}

function connectedNodes(
  item: WorkflowGraphNodeBox,
  nodeMap: Map<string, WorkflowGraphNodeBox>,
  edges: WorkflowPublishedEdge[],
) {
  const connectedIds = item.node.type === 'start'
    ? edges.filter(edge => edge.source === item.node.id).map(edge => edge.target)
    : edges.filter(edge => edge.target === item.node.id).map(edge => edge.source)
  return connectedIds
    .map(id => nodeMap.get(id))
    .filter((node): node is WorkflowGraphNodeBox => Boolean(node))
}

export function alignWorkflowTerminalNodes(
  nodes: WorkflowGraphNodeBox[],
  edges: WorkflowPublishedEdge[],
): WorkflowGraphNodeBox[] {
  const nodeMap = new Map(nodes.map(item => [item.node.id, item]))

  return nodes.map((item) => {
    if (item.node.type !== 'start' && item.node.type !== 'end')
      return item

    const connected = connectedNodes(item, nodeMap, edges)
    if (connected.length === 0)
      return item

    const nonTerminal = connected.filter(node => node.node.type !== 'start' && node.node.type !== 'end')
    const references = nonTerminal.length > 0 ? nonTerminal : [item, ...connected]
    const targetCenterX = references.reduce((sum, node) => sum + centerX(node), 0) / references.length
    return {
      ...item,
      x: targetCenterX - item.width / 2,
    }
  })
}
