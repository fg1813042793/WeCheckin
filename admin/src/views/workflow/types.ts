export type WorkflowNodeType = 'start' | 'approval' | 'exclusive' | 'parallel' | 'end'
export type ApprovalMode = 'single' | 'sequential' | 'parallel' | 'countersign'
export type AssigneeType = 'user' | 'role' | 'department_leader' | 'manager' | 'variable'
export type GatewayMode = 'split' | 'join'
export type ConditionOperator = 'eq' | 'ne' | 'gt' | 'gte' | 'lt' | 'lte'

export interface WorkflowAssignee {
  type: AssigneeType
  value: string
}

export interface WorkflowCondition {
  field: string
  operator: ConditionOperator
  value: string | number | boolean
}

export interface WorkflowNode {
  id: string
  type: WorkflowNodeType
  name: string
  position?: WorkflowNodePosition
  approvalMode?: ApprovalMode
  assignee?: WorkflowAssignee
  completionRate?: number
  gatewayMode?: GatewayMode
}

export interface WorkflowNodePosition {
  x: number
  y: number
}

export interface WorkflowEdge {
  id: string
  source: string
  target: string
  name?: string
  default?: boolean
  condition?: WorkflowCondition
}

export interface WorkflowDraft {
  schemaVersion: number
  key: string
  name: string
  nodes: WorkflowNode[]
  edges: WorkflowEdge[]
}

export interface WorkflowDefinitionSummary {
  id: number
  key: string
  name: string
  description: string
  category: string
  status: number
  currentVersion: number
  addUserId: number
  editUserId: number
  addTime: number
  editTime: number
}

export interface WorkflowDefinitionDetail extends WorkflowDefinitionSummary {
  draft: WorkflowDraft
}

export interface WorkflowValidationError {
  code: string
  message: string
  nodeId?: string
  edgeId?: string
}

export interface WorkflowVersion {
  id: number
  definitionId: number
  version: number
  deploymentId: string
  publishedBy: number
  publishedAt: number
}

export interface WorkflowValidationResult {
  valid: boolean
  errors: WorkflowValidationError[]
}

export interface WorkflowPublishResult {
  definitionId: number
  version: number
  bpmnXml: string
}

export type WorkflowInstanceStatus = 'running' | 'completed' | 'rejected' | 'cancelled'
export type WorkflowTaskStatus = 'waiting' | 'pending' | 'approved' | 'rejected' | 'cancelled'
export type WorkflowTaskAction = 'approve' | 'reject'

export interface WorkflowInstanceSummary {
  id: string
  definitionId: number
  definitionVersion: number
  definitionKey: string
  businessType: string
  businessKey: string
  starterId: string
  status: WorkflowInstanceStatus
  startTime: number
  endTime: number
}

export interface WorkflowTokenSummary {
  id: string
  nodeId: string
  status: string
  branchGroup: string
  branchTotal: number
}

export interface WorkflowTaskSummary {
  id: string
  instanceId: string
  nodeId: string
  nodeName: string
  assigneeId: string
  approvalMode: string
  completionRate: number
  sequence: number
  total: number
  status: WorkflowTaskStatus
  action: WorkflowTaskAction | ''
  comment: string
  handledBy: string
  handledAt: number
}

export interface WorkflowHistorySummary {
  id: string
  eventType: string
  nodeId: string
  taskId: string
  actorId: string
  message: string
  eventTime: number
}

export interface WorkflowInstanceDetail {
  instance: WorkflowInstanceSummary
  variables: Record<string, unknown>
  tokens: WorkflowTokenSummary[]
  tasks: WorkflowTaskSummary[]
  history: WorkflowHistorySummary[]
}

export function cloneDraft(draft: WorkflowDraft): WorkflowDraft {
  return JSON.parse(JSON.stringify(draft)) as WorkflowDraft
}
