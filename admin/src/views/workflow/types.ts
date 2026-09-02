export type WorkflowNodeType = 'start' | 'approval' | 'handle' | 'cc' | 'notify' | 'automation' | 'timer' | 'exclusive' | 'parallel' | 'end'
export type WorkflowInsertableNodeType = Exclude<WorkflowNodeType, 'start' | 'end'>
export type ApprovalMode = 'single' | 'sequential' | 'parallel' | 'countersign'
export type AssigneeType = 'initiator' | 'user' | 'role' | 'department_leader' | 'manager' | 'variable' | 'org_identity'
export type GatewayMode = 'split' | 'join'
export type ConditionOperator = 'eq' | 'ne' | 'gt' | 'gte' | 'lt' | 'lte'
export type WorkflowFormFieldType =
  | 'text' | 'textarea' | 'number' | 'amount' | 'phone' | 'email' | 'boolean'
  | 'select' | 'multi_select' | 'radio' | 'checkbox'
  | 'date' | 'datetime' | 'time' | 'date_range'
  | 'user' | 'user_multi' | 'department' | 'department_multi' | 'attachment'
  | 'detail_list'
  | 'group' | 'label' | 'description' | 'button'
export type WorkflowFormFieldSpan = 6 | 8 | 12 | 24
export type WorkflowFieldAccess = 'hidden' | 'read' | 'write'
export type WorkflowDetailRowAction = 'add' | 'delete'
export type WorkflowOptionSourceType = 'static' | 'api'
export type WorkflowInitiatorScope = 'all' | 'specified'
export type WorkflowStartAvailabilityMode = 'always' | 'fixed' | 'weekly' | 'monthly'
export type WorkflowStartAvailabilityStatus = 'available' | 'not_started' | 'expired' | 'outside_window'
export type WorkflowEdgeHandle = 'top' | 'right' | 'bottom' | 'left'
export type WorkflowNotificationChannel = 'in_app' | 'dingtalk_oa'
export type WorkflowValidationRuleType =
  | 'min_length' | 'max_length' | 'pattern' | 'number_range'
  | 'decimal_places' | 'selection_count' | 'compare_field' | 'column_sum' | 'conditional_required'
export type WorkflowValidationOperator = 'eq' | 'ne' | 'gt' | 'gte' | 'lt' | 'lte' | 'empty' | 'not_empty'

export interface WorkflowFormOption {
  label: string
  value: string
  children?: WorkflowFormOption[]
}

export interface WorkflowOptionSource {
  type: WorkflowOptionSourceType
  url?: string
  method?: 'GET' | 'POST'
  responsePath?: string
  labelField?: string
  valueField?: string
  childrenField?: string
}

export interface WorkflowFormHelp {
  buttonText?: string
  title: string
  content: string
}

export interface WorkflowValidationCondition {
  field: string
  operator: WorkflowValidationOperator
  value?: unknown
}

export interface WorkflowValidationRule {
  id: string
  type: WorkflowValidationRuleType
  min?: number
  max?: number
  precision?: number
  pattern?: string
  field?: string
  column?: string
  operator?: WorkflowValidationOperator
  value?: number
  when?: WorkflowValidationCondition
  message?: string
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
  optionSource?: WorkflowOptionSource
  span?: WorkflowFormFieldSpan
  rowKey?: string
  columns?: WorkflowFormField[]
  fields?: WorkflowFormField[]
  content?: string
  help?: WorkflowFormHelp
  minRows?: number
  maxRows?: number
  rules?: WorkflowValidationRule[]
}

export interface WorkflowFieldPermission {
  field: string
  access: WorkflowFieldAccess
  actions?: WorkflowDetailRowAction[]
}

export interface WorkflowAssignee {
  type: AssigneeType
  value: string
}

export interface WorkflowInitiatorConfig {
  scope: WorkflowInitiatorScope
  userIds?: number[]
  departmentIds?: number[]
}

export interface WorkflowStartAvailabilityConfig {
  mode: WorkflowStartAvailabilityMode
  timezone?: string
  startsAt?: number
  endsAt?: number
  effectiveStartDate?: string
  effectiveEndDate?: string
  weekdays?: number[]
  monthDays?: number[]
  lastDayOfMonth?: boolean
  dailyStartTime?: string
  dailyEndTime?: string
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
  status?: number
}

export interface WorkflowCondition {
  field: string
  operator: ConditionOperator
  value: string | number | boolean
}

export interface WorkflowAutomationConfig {
  type: 'set_variables'
  variables: Record<string, unknown>
}

export interface WorkflowTimerConfig {
  delaySeconds: number
}

export interface WorkflowNotificationConfig {
  enabled: boolean
  channels: WorkflowNotificationChannel[]
  title: string
  content: string
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
  automation?: WorkflowAutomationConfig
  timer?: WorkflowTimerConfig
  initiator?: WorkflowInitiatorConfig
  availability?: WorkflowStartAvailabilityConfig
  notification?: WorkflowNotificationConfig
}

export interface WorkflowNodePosition {
  x: number
  y: number
}

export interface WorkflowEdge {
  id: string
  source: string
  target: string
  sourceHandle?: WorkflowEdgeHandle
  targetHandle?: WorkflowEdgeHandle
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
  logoUrl: string
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

export interface WorkflowPublishedDefinition {
  id: number
  key: string
  name: string
  description: string
  category: string
  version: number
  form: WorkflowFormField[]
  fieldPermissions: Record<string, WorkflowFieldPermission[]>
  startNodeId: string
  initiator: WorkflowInitiatorConfig
  availability: WorkflowStartAvailabilityConfig
  availabilityStatus: WorkflowStartAvailabilityStatus
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
export type WorkflowTaskStatus = 'waiting' | 'pending' | 'completed' | 'approved' | 'rejected' | 'cancelled'
export type WorkflowTaskAction = 'approve' | 'reject' | 'submit'

export interface WorkflowInstanceSummary {
  id: string
  definitionId: number
  definitionVersion: number
  definitionKey: string
  businessType: string
  businessKey: string
  starterId: string
  operatorId: string
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

export type WorkflowNotificationKind = 'node_cc' | 'node_notify' | 'task_arrived'
export type WorkflowNotificationStatus = 'pending' | 'sending' | 'sent' | 'failed' | 'dead'

export interface WorkflowNotificationPayload {
  title: string
  content: string
  workflowName: string
  nodeName: string
  starterId: string
  starterName: string
  instanceId: string
  taskId: string
  recipientUserId: string
  kind: WorkflowNotificationKind
}

export interface WorkflowNotificationRecord {
  id: string
  instanceId: string
  nodeId: string
  taskId: string
  recipientUserId: string
  kind: WorkflowNotificationKind
  channel: WorkflowNotificationChannel
  status: WorkflowNotificationStatus
  payload: WorkflowNotificationPayload
  corpId: string
  providerMessageId: string
  attempts: number
  nextRetryAt: number
  lastError: string
  sentAt: number
  addTime: number
  editTime: number
}

export interface WorkflowNotificationList {
  list: WorkflowNotificationRecord[]
  total: number
  page: number
  pageSize: number
}

export interface WorkflowInstanceDetail {
  instance: WorkflowInstanceSummary
  variables: Record<string, unknown>
  form: WorkflowFormField[]
  formData: Record<string, unknown>
  fieldPermissions: Record<string, WorkflowFieldPermission[]>
  startNodeId: string
  nodeTypes: Record<string, WorkflowNodeType>
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
