export type WorkflowNodeType = 'start' | 'approval' | 'exclusive' | 'parallel' | 'end'
export type ApprovalMode = 'single' | 'sequential' | 'parallel' | 'countersign'
export type AssigneeType = 'user' | 'role' | 'department_leader' | 'manager' | 'variable' | 'org_identity'
export type GatewayMode = 'split' | 'join'
export type ConditionOperator = 'eq' | 'ne' | 'gt' | 'gte' | 'lt' | 'lte'
export type WorkflowFormFieldType =
  | 'text' | 'textarea' | 'number' | 'amount' | 'phone' | 'email' | 'boolean'
  | 'select' | 'multi_select' | 'radio' | 'checkbox'
  | 'date' | 'datetime' | 'time' | 'date_range'
  | 'user' | 'user_multi' | 'department' | 'department_multi' | 'attachment'
export type WorkflowFormFieldSpan = 6 | 8 | 12 | 24
export type WorkflowFieldAccess = 'hidden' | 'read' | 'write'

export interface WorkflowFormOption {
  label: string
  value: string
}

export interface WorkflowFormField {
  key: string
  label: string
  type: WorkflowFormFieldType
  required?: boolean
  default?: unknown
  placeholder?: string
  maxLength?: number
  min?: number
  max?: number
  options?: WorkflowFormOption[]
  span?: WorkflowFormFieldSpan
}

export interface WorkflowFieldPermission {
  field: string
  access: WorkflowFieldAccess
}

export interface WorkflowAssignee {
  type: AssigneeType
  value: string
}

export interface WorkflowOrgApproverIdentity {
  id?: number
  code: string
  name: string
  sort?: number
  status?: number
}

export interface WorkflowAssigneeUser {
  id: number
  name: string
  mobile?: string
  deptIds?: number[]
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
  formPermissions?: WorkflowFieldPermission[]
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
  form: WorkflowFormField[]
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
  const cloned = JSON.parse(JSON.stringify(draft)) as WorkflowDraft
  cloned.form = Array.isArray(cloned.form) ? cloned.form : []
  cloned.nodes = Array.isArray(cloned.nodes) ? cloned.nodes : []
  cloned.edges = Array.isArray(cloned.edges) ? cloned.edges : []
  cloned.nodes.forEach((node) => {
    node.formPermissions = Array.isArray(node.formPermissions) ? node.formPermissions : []
  })
  return cloned
}
