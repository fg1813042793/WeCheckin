import type { WorkflowDraft, WorkflowEdge, WorkflowNode } from '../types'
import { findPairedJoin } from './graph'

export interface WorkflowTreeSequence {
  items: WorkflowTreeItem[]
}

export interface WorkflowTreeNodeItem {
  kind: 'node'
  node: WorkflowNode
  outgoingEdgeId?: string
}

export interface WorkflowTreeBranch {
  edge: WorkflowEdge
  entryEdgeId: string
  sequence: WorkflowTreeSequence
}

export interface WorkflowTreeBranchItem {
  kind: 'branch'
  split: WorkflowNode
  join: WorkflowNode
  branches: WorkflowTreeBranch[]
  outgoingEdgeId?: string
}

export type WorkflowTreeItem = WorkflowTreeNodeItem | WorkflowTreeBranchItem

export function buildWorkflowTree(draft: WorkflowDraft): WorkflowTreeSequence {
  const start = draft.nodes.find(item => item.type === 'start')
  if (!start) return { items: [] }
  const nodeMap = new Map(draft.nodes.map(item => [item.id, item]))
  const outgoing = new Map<string, WorkflowEdge[]>()
  draft.edges.forEach(item => outgoing.set(item.source, [...(outgoing.get(item.source) ?? []), item]))

  function buildSequence(startID: string, stopID?: string, ancestors = new Set<string>()): WorkflowTreeSequence {
    const items: WorkflowTreeItem[] = []
    let currentID = startID
    const visited = new Set(ancestors)
    while (currentID && currentID !== stopID && !visited.has(currentID)) {
      visited.add(currentID)
      const node = nodeMap.get(currentID)
      if (!node) break

      if (node.gatewayMode === 'split') {
        const join = findPairedJoin(draft, node.id)
        if (!join) {
          items.push({ kind: 'node', node })
          break
        }
        const branchEdges = outgoing.get(node.id) ?? []
        const after = (outgoing.get(join.id) ?? [])[0]
        items.push({
          kind: 'branch',
          split: node,
          join,
          branches: branchEdges.map(branch => ({
            edge: branch,
            entryEdgeId: branch.id,
            sequence: buildSequence(branch.target, join.id, visited),
          })),
          ...(after ? { outgoingEdgeId: after.id } : {}),
        })
        if (!after) break
        currentID = after.target
        continue
      }

      const next = (outgoing.get(node.id) ?? [])[0]
      items.push({ kind: 'node', node, ...(next ? { outgoingEdgeId: next.id } : {}) })
      if (!next) break
      currentID = next.target
    }
    return { items }
  }

  return buildSequence(start.id)
}

export function workflowPermissionNodes(draft: WorkflowDraft): WorkflowNode[] {
  const result: WorkflowNode[] = []
  const included = new Set<string>()

  function append(sequence: WorkflowTreeSequence) {
    for (const item of sequence.items) {
      if (item.kind === 'branch') {
        item.branches.forEach(branch => append(branch.sequence))
        continue
      }
      if (!['start', 'approval', 'handle'].includes(item.node.type) || included.has(item.node.id)) continue
      included.add(item.node.id)
      result.push(item.node)
    }
  }

  append(buildWorkflowTree(draft))
  return result
}
