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
  | 'calculation'
  | 'group' | 'label' | 'description' | 'button'
export type WorkflowFormFieldSpan = 6 | 8 | 12 | 24
export type WorkflowFieldAccess = 'hidden' | 'read' | 'write'
export type WorkflowDetailRowAction = 'add' | 'delete'
export type WorkflowOptionSourceType = 'static' | 'api'
export type WorkflowCalculationDisplay = 'label' | 'field'
export type WorkflowInitiatorScope = 'all' | 'specified'
export type WorkflowStartAvailabilityMode = 'always' | 'fixed' | 'weekly' | 'monthly'
export type WorkflowStartAvailabilityStatus = 'available' | 'not_started' | 'expired' | 'outside_window'
export type WorkflowStartLimitMode = 'unlimited' | 'limited'
export type WorkflowStartLimitPeriod = 'total' | 'day' | 'week' | 'month' | 'availability'
export type WorkflowEdgeHandle = 'top' | 'right' | 'bottom' | 'left'
export type WorkflowNotificationChannel = 'in_app' | 'dingtalk_oa'
export type WorkflowNotificationResultType = 'approved' | 'rejected' | 'returned'
export type WorkflowValidationRuleType =
  | 'min_length' | 'max_length' | 'pattern' | 'number_range'
  | 'decimal_places' | 'selection_count' | 'compare_field' | 'column_sum' | 'conditional_required'
export type WorkflowValidationOperator = 'eq' | 'ne' | 'gt' | 'gte' | 'lt' | 'lte' | 'empty' | 'not_empty'

export interface WorkflowAttachment {
  id: string
  name: string
  url: string
  mimeType: string
  size: number
}

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
  minVisibleRows?: number
  maxVisibleRows?: number
  rules?: WorkflowValidationRule[]
  calculation?: WorkflowFormCalculation
}

export interface WorkflowFormCalculation {
  expression: string
  display?: WorkflowCalculationDisplay
  precision?: number
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

export type WorkflowDepartmentApprovalChainStopMode = 'root' | 'department'
export type WorkflowDepartmentApprovalChainMissingPolicy = 'skip' | 'error'

export interface WorkflowDepartmentApprovalChain {
  enabled: boolean
  stopMode: WorkflowDepartmentApprovalChainStopMode
  stopDepartmentId?: number
  missingAssigneePolicy: WorkflowDepartmentApprovalChainMissingPolicy
  skipStarter?: boolean
}

export interface WorkflowInitiatorConfig {
  scope: WorkflowInitiatorScope
  userIds?: number[]
  departmentIds?: number[]
  excludedUserIds?: number[]
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

export interface WorkflowStartLimitConfig {
  mode: WorkflowStartLimitMode
  period?: WorkflowStartLimitPeriod
  maxCount?: number
}

export interface WorkflowStartLimitStatus {
  allowed: boolean
  usedCount: number
  remainingCount: number
  resetsAt?: number
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
  resultTypes?: WorkflowNotificationResultType[]
}

export interface WorkflowPostHandleEditConfig {
  enabled: boolean
}

export interface WorkflowNode {
  id: string
  type: WorkflowNodeType
  name: string
  position?: WorkflowNodePosition
  approvalMode?: ApprovalMode
  assignee?: WorkflowAssignee
  departmentApprovalChain?: WorkflowDepartmentApprovalChain
  completionRate?: number
  gatewayMode?: GatewayMode
  formPermissions?: WorkflowFieldPermission[]
  postHandleEdit?: WorkflowPostHandleEditConfig
  automation?: WorkflowAutomationConfig
  timer?: WorkflowTimerConfig
  initiator?: WorkflowInitiatorConfig
  availability?: WorkflowStartAvailabilityConfig
  startLimit?: WorkflowStartLimitConfig
  notification?: WorkflowNotificationConfig
  resultNotification?: WorkflowNotificationConfig
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
  startLimit: WorkflowStartLimitConfig
  startLimitStatus: WorkflowStartLimitStatus
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
  name: string
  deploymentId: string
  publishedBy: number
  publishedByName: string
  publishedAt: number
  publishNote: string
  rollbackFromVersion: number
  changeBaseVersion: number
  changeHeadline: string
  changeCount: number
  instanceCount: number
  startDraftCount: number
  isCurrent: boolean
  canDelete: boolean
  deleteBlockedReason: string
}

export type WorkflowVersionChangeCategory = 'basic' | 'form' | 'node' | 'route' | 'start' | 'notification' | 'automation'
export type WorkflowVersionChangeAction = 'add' | 'update' | 'delete' | 'reorder'

export interface WorkflowVersionChangeItem {
  category: WorkflowVersionChangeCategory
  action: WorkflowVersionChangeAction
  title: string
  detail: string
}

export interface WorkflowVersionChangeSummary {
  baseVersion: number
  headline: string
  changeCount: number
  items: WorkflowVersionChangeItem[]
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
export type WorkflowTaskStatus = 'waiting' | 'pending' | 'completed' | 'approved' | 'rejected' | 'returned' | 'cancelled'
export type WorkflowTaskAction = 'approve' | 'reject' | 'return' | 'submit'

export interface WorkflowInstanceSummary {
  id: string
  definitionId: number
  definitionVersion: number
  definitionKey: string
  definitionName: string
  businessType: string
  businessKey: string
  starterId: string
  starterName: string
  operatorId: string
  operatorName: string
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
  definitionName: string
  starterId: string
  starterName: string
  assigneeId: string
  assigneeName: string
  approvalMode: string
  completionRate: number
  sequence: number
  total: number
  approvalChainKey?: string
  approvalLayer?: number
  approvalLayerTotal?: number
  sourceDepartmentId?: number
  sourceDepartmentName?: string
  status: WorkflowTaskStatus
  action: WorkflowTaskAction | ''
  comment: string
  handledBy: string
  handledByName: string
  handledAt: number
}

export interface WorkflowHistorySummary {
  id: string
  eventType: string
  nodeId: string
  taskId: string
  actorId: string
  actorName: string
  message: string
  eventTime: number
}

export type WorkflowNotificationKind
  = 'node_cc'
    | 'node_notify'
    | 'task_arrived'
    | 'task_reminder'
    | 'instance_commented'
    | 'instance_form_revised'
    | 'approval_result_approved'
    | 'approval_result_rejected'
    | 'approval_result_returned'
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
  recipientUserName: string
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
  userNames: Record<string, string>
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
