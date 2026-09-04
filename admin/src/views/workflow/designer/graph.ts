import type { WorkflowDraft, WorkflowEdge, WorkflowEdgeHandle, WorkflowInsertableNodeType, WorkflowNode, WorkflowNodeType, WorkflowNotificationConfig } from '../types'

let sequence = 0

function nextID(prefix: string) {
  sequence += 1
  return `${prefix}_${Date.now().toString(36)}_${sequence.toString(36)}`
}

function approvalNode(name = '审批节点'): WorkflowNode {
  return {
    id: nextID('approval'),
    type: 'approval',
    name,
    approvalMode: 'single',
    assignee: { type: 'manager', value: 'direct_manager' },
    notification: defaultNotificationConfig('approval'),
    resultNotification: defaultResultNotificationConfig(),
  }
}

export function defaultNotificationConfig(type: Extract<WorkflowNodeType, 'approval' | 'handle' | 'cc' | 'notify'>): WorkflowNotificationConfig {
  const content = type === 'cc'
    ? '{{starterName}} 发起的流程已抄送给你'
    : type === 'notify'
      ? '流程已到达 {{nodeName}}'
      : '你有一项待处理任务：{{nodeName}}'
  return {
    enabled: true,
    channels: ['in_app', 'dingtalk_oa'],
    title: '{{workflowName}}',
    content,
  }
}

export function defaultResultNotificationConfig(): WorkflowNotificationConfig {
  return {
    enabled: true,
    channels: ['in_app', 'dingtalk_oa'],
    title: '{{workflowName}}审批结果',
    content: '你发起的流程在“{{nodeName}}”节点{{result}}',
  }
}

function operationNode(type: Extract<WorkflowNodeType, 'handle' | 'cc' | 'notify' | 'automation' | 'timer'>): WorkflowNode {
  if (type === 'handle') {
    return {
      id: nextID(type), type, name: '办理/填写',
      assignee: { type: 'manager', value: 'direct_manager' },
      notification: defaultNotificationConfig(type),
    }
  }
  if (type === 'cc') {
    return {
      id: nextID(type), type, name: '抄送',
      assignee: { type: 'variable', value: 'ccUserIds' },
      notification: defaultNotificationConfig(type),
    }
  }
  if (type === 'notify') {
    return {
      id: nextID(type), type, name: '通知',
      assignee: { type: 'variable', value: 'notificationUserIds' },
      notification: defaultNotificationConfig(type),
    }
  }
  if (type === 'automation') {
    return { id: nextID(type), type, name: '自动动作', automation: { type: 'set_variables', variables: { processed: true } } }
  }
  return { id: nextID(type), type, name: '定时/等待', timer: { delaySeconds: 86400 } }
}

function edge(source: string, target: string, name = '', sourceHandle?: WorkflowEdgeHandle, targetHandle?: WorkflowEdgeHandle): WorkflowEdge {
  return {
    id: nextID('flow'),
    source,
    target,
    ...(name ? { name } : {}),
    ...(sourceHandle ? { sourceHandle } : {}),
    ...(targetHandle ? { targetHandle } : {}),
  }
}

function reconnectEdge(
  original: WorkflowEdge,
  source: string,
  target: string,
  handles?: { sourceHandle?: WorkflowEdgeHandle; targetHandle?: WorkflowEdgeHandle },
): WorkflowEdge {
  const sourceHandle = handles ? handles.sourceHandle : original.sourceHandle
  const targetHandle = handles ? handles.targetHandle : original.targetHandle
  return {
    ...original,
    id: nextID('flow'),
    source,
    target,
    sourceHandle,
    targetHandle,
    ...(original.condition ? { condition: { ...original.condition } } : {}),
  }
}

function insertionEdge(draft: WorkflowDraft, selectedID?: string) {
  if (selectedID) {
    const outgoing = draft.edges.filter(item => item.source === selectedID)
    if (outgoing.length === 1) return outgoing[0]
  }
  const end = draft.nodes.find(item => item.type === 'end')
  if (!end) return undefined
  const incoming = draft.edges.filter(item => item.target === end.id)
  return incoming.length === 1 ? incoming[0] : undefined
}

export function insertApproval(draft: WorkflowDraft, selectedID?: string) {
  const currentEdge = insertionEdge(draft, selectedID)
  return currentEdge ? insertNodeAtEdge(draft, currentEdge.id, 'approval') : null
}

export function insertGateway(draft: WorkflowDraft, type: Extract<WorkflowNodeType, 'exclusive' | 'parallel'>, selectedID?: string) {
  const currentEdge = insertionEdge(draft, selectedID)
  return currentEdge ? insertNodeAtEdge(draft, currentEdge.id, type) : null
}

export function insertNodeAtEdge(
  draft: WorkflowDraft,
  edgeID: string,
  type: WorkflowInsertableNodeType,
) {
  const currentEdge = draft.edges.find(item => item.id === edgeID)
  if (!currentEdge) return null
  if (type !== 'exclusive' && type !== 'parallel') {
    const node = type === 'approval' ? approvalNode() : operationNode(type)
    draft.edges = draft.edges.filter(item => item.id !== currentEdge.id)
    draft.nodes.push(node)
    draft.edges.push(
      reconnectEdge(currentEdge, currentEdge.source, node.id, { sourceHandle: currentEdge.sourceHandle }),
      edge(node.id, currentEdge.target, '', undefined, currentEdge.targetHandle),
    )
    return node
  }
  const label = type === 'exclusive' ? '条件分支' : '并行分支'
  const split: WorkflowNode = { id: nextID(type), type, name: label, gatewayMode: 'split' }
  const join: WorkflowNode = { id: nextID(`${type}_join`), type, name: `${label}汇聚`, gatewayMode: 'join' }
  const branchA = approvalNode('分支一审批')
  const branchB = approvalNode('分支二审批')
  const first = edge(split.id, branchA.id, type === 'exclusive' ? '满足条件' : '分支一')
  const second = edge(split.id, branchB.id, type === 'exclusive' ? '其他情况' : '分支二')
  if (type === 'exclusive') {
    first.condition = { field: 'approved', operator: 'eq', value: true }
    second.default = true
  }
  draft.edges = draft.edges.filter(item => item.id !== currentEdge.id)
  draft.nodes.push(split, branchA, branchB, join)
  draft.edges.push(
    reconnectEdge(currentEdge, currentEdge.source, split.id, { sourceHandle: currentEdge.sourceHandle }),
    first,
    second,
    edge(branchA.id, join.id),
    edge(branchB.id, join.id),
    edge(join.id, currentEdge.target, '', undefined, currentEdge.targetHandle),
  )
  return split
}

function reachableJoinDistances(draft: WorkflowDraft, startID: string) {
  const outgoing = new Map<string, WorkflowEdge[]>()
  draft.edges.forEach(item => outgoing.set(item.source, [...(outgoing.get(item.source) ?? []), item]))
  const distances = new Map<string, number>()
  const queue: Array<{ id: string; distance: number }> = [{ id: startID, distance: 0 }]
  const visited = new Set<string>()
  while (queue.length) {
    const current = queue.shift()!
    if (visited.has(current.id)) continue
    visited.add(current.id)
    const node = draft.nodes.find(item => item.id === current.id)
    if (node?.gatewayMode === 'join') distances.set(node.id, current.distance)
    ;(outgoing.get(current.id) ?? []).forEach(item => queue.push({ id: item.target, distance: current.distance + 1 }))
  }
  return distances
}

export function findPairedJoin(draft: WorkflowDraft, splitID: string) {
  const branchTargets = draft.edges.filter(item => item.source === splitID).map(item => item.target)
  if (branchTargets.length < 2) return undefined
  const candidates = branchTargets.map(target => reachableJoinDistances(draft, target))
  const common = Array.from(candidates[0].keys()).filter(id => candidates.every(items => items.has(id)))
  common.sort((left, right) => candidates.reduce((total, items) => total + (items.get(left) ?? Number.MAX_SAFE_INTEGER), 0)
    - candidates.reduce((total, items) => total + (items.get(right) ?? Number.MAX_SAFE_INTEGER), 0))
  const split = draft.nodes.find(item => item.id === splitID)
  return draft.nodes.find(item => item.id === common[0] && item.gatewayMode === 'join' && item.type === split?.type)
}

function findPairedSplit(draft: WorkflowDraft, joinID: string) {
  const reversed: WorkflowDraft = {
    ...draft,
    nodes: draft.nodes.map(item => ({
      ...item,
      gatewayMode: item.gatewayMode === 'join' ? 'split' : item.gatewayMode === 'split' ? 'join' : undefined,
    })),
    edges: draft.edges.map(item => ({
      ...item,
      source: item.target,
      target: item.source,
      sourceHandle: item.targetHandle,
      targetHandle: item.sourceHandle,
    })),
  }
  const paired = findPairedJoin(reversed, joinID)
  return draft.nodes.find(item => item.id === paired?.id && item.gatewayMode === 'split')
}

function collectNodesBeforeJoin(draft: WorkflowDraft, splitID: string, joinID: string) {
  const removed = new Set<string>([splitID, joinID])
  const queue = draft.edges.filter(item => item.source === splitID).map(item => item.target)
  while (queue.length) {
    const current = queue.shift()!
    if (current === joinID || removed.has(current)) continue
    removed.add(current)
    draft.edges.filter(item => item.source === current).forEach(item => queue.push(item.target))
  }
  return removed
}

export function removeNode(draft: WorkflowDraft, nodeID: string) {
  const node = draft.nodes.find(item => item.id === nodeID)
  if (!node || node.type === 'start' || node.type === 'end') return false

  if (node.gatewayMode === 'split' || node.gatewayMode === 'join') {
    const split = node.gatewayMode === 'split' ? node : findPairedSplit(draft, node.id)
    const join = split ? findPairedJoin(draft, split.id) : undefined
    if (!split || !join) return false
    const before = draft.edges.find(item => item.target === split.id)
    const after = draft.edges.find(item => item.source === join.id)
    if (!before || !after) return false
    const removed = collectNodesBeforeJoin(draft, split.id, join.id)
    draft.nodes = draft.nodes.filter(item => !removed.has(item.id))
    draft.edges = draft.edges.filter(item => !removed.has(item.source) && !removed.has(item.target))
    draft.edges.push(edge(before.source, after.target, '', before.sourceHandle, after.targetHandle))
    return true
  }

  const incoming = draft.edges.filter(item => item.target === nodeID)
  const outgoing = draft.edges.filter(item => item.source === nodeID)
  if (incoming.length !== 1 || outgoing.length !== 1) return false
  const parent = draft.nodes.find(item => item.id === incoming[0].source)
  const child = draft.nodes.find(item => item.id === outgoing[0].target)
  if (parent?.gatewayMode === 'split' && child?.gatewayMode === 'join') {
    const branchCount = draft.edges.filter(item => item.source === parent.id).length
    if (branchCount <= 2) return false
    draft.nodes = draft.nodes.filter(item => item.id !== nodeID)
    draft.edges = draft.edges.filter(item => item.source !== nodeID && item.target !== nodeID)
    return true
  }
  draft.nodes = draft.nodes.filter(item => item.id !== nodeID)
  draft.edges = draft.edges.filter(item => item.source !== nodeID && item.target !== nodeID)
  draft.edges.push(reconnectEdge(incoming[0], incoming[0].source, outgoing[0].target, {
    sourceHandle: incoming[0].sourceHandle,
    targetHandle: outgoing[0].targetHandle,
  }))
  return true
}

export function addBranch(draft: WorkflowDraft, splitID: string) {
  const split = draft.nodes.find(item => item.id === splitID && item.gatewayMode === 'split')
  const join = split ? findPairedJoin(draft, split.id) : undefined
  if (!split || !join) return null
  const node = approvalNode(`分支${draft.edges.filter(item => item.source === splitID).length + 1}审批`)
  const branch = edge(split.id, node.id, `分支${draft.edges.filter(item => item.source === splitID).length + 1}`)
  if (split.type === 'exclusive') branch.condition = { field: 'approved', operator: 'eq', value: true }
  draft.nodes.push(node)
  draft.edges.push(branch, edge(node.id, join.id))
  return node
}

export function removeBranch(draft: WorkflowDraft, splitID: string, branchEdgeID: string) {
  const split = draft.nodes.find(item => item.id === splitID && item.gatewayMode === 'split')
  const join = split ? findPairedJoin(draft, split.id) : undefined
  const branches = draft.edges.filter(item => item.source === splitID)
  const branch = branches.find(item => item.id === branchEdgeID)
  if (!split || !join || !branch || branches.length <= 2) return false

  const removedNodes = new Set<string>()
  const queue = [branch.target]
  while (queue.length) {
    const current = queue.shift()!
    if (current === join.id || removedNodes.has(current)) continue
    removedNodes.add(current)
    draft.edges.filter(item => item.source === current).forEach(item => queue.push(item.target))
  }

  draft.nodes = draft.nodes.filter(item => !removedNodes.has(item.id))
  draft.edges = draft.edges.filter(item => item.id !== branch.id
    && !removedNodes.has(item.source)
    && !removedNodes.has(item.target))

  if (split.type === 'exclusive' && branch.default) {
    const replacement = draft.edges.find(item => item.source === splitID)
    if (replacement) {
      replacement.default = true
      delete replacement.condition
    }
  }
  return true
}
