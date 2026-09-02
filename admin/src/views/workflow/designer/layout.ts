import dagre from '@dagrejs/dagre'
import type { WorkflowDraft, WorkflowNode, WorkflowNodePosition } from '../types'

export const WORKFLOW_NODE_WIDTH = 238
export const WORKFLOW_NODE_HEIGHT = 90

export function layoutWorkflow(draft: WorkflowDraft): Map<string, WorkflowNodePosition> {
  const graph = new dagre.graphlib.Graph()
  graph.setDefaultEdgeLabel(() => ({}))
  graph.setGraph({
    rankdir: 'TB',
    align: 'UL',
    nodesep: 92,
    ranksep: 118,
    marginx: 72,
    marginy: 48,
  })

  draft.nodes.forEach((node) => {
    const dimensions = nodeDimensions(node)
    graph.setNode(node.id, dimensions)
  })
  draft.edges.forEach(edge => graph.setEdge(edge.source, edge.target))
  dagre.layout(graph)

  return new Map(draft.nodes.map((node) => {
    const point = graph.node(node.id)
    const dimensions = nodeDimensions(node)
    return [node.id, {
      x: Math.round(point.x - dimensions.width / 2),
      y: Math.round(point.y - dimensions.height / 2),
    }]
  }))
}

function nodeDimensions(node: WorkflowNode) {
  if (node.type === 'start') return { width: 120, height: 76 }
  if (node.type === 'end') return { width: 120, height: 76 }
  if (node.gatewayMode) return { width: 136, height: 136 }
  return { width: WORKFLOW_NODE_WIDTH, height: WORKFLOW_NODE_HEIGHT }
}
