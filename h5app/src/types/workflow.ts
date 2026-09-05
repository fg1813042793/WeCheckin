export type WorkflowFieldType
  = | 'text'
    | 'textarea'
    | 'number'
    | 'select'
    | 'multi_select'
    | 'date'
    | 'datetime'
    | 'user'
    | 'department'
    | 'attachment'
    | 'boolean'
    | 'amount'
    | 'phone'
    | 'email'
    | 'radio'
    | 'checkbox'
    | 'time'
    | 'date_range'
    | 'user_multi'
    | 'department_multi'
    | 'detail_list'
    | 'group'
    | 'label'
    | 'description'
    | 'button'
    | string

export type WorkflowFieldAccess = 'hidden' | 'read' | 'write'
export type WorkflowFieldAction = 'add' | 'delete'
export type WorkflowFormData = Record<string, unknown>
export type WorkflowFieldAccessMap = Record<string, WorkflowFieldAccess>
export type WorkflowFieldActionsMap = Record<string, WorkflowFieldAction[]>

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
  type: 'static' | 'api' | string
  url?: string
  method?: string
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

export interface WorkflowFormRuleCondition {
  field: string
  operator: string
  value?: unknown
}

export interface WorkflowFormRule {
  id: string
  type: string
  min?: number
  max?: number
  precision?: number
  pattern?: string
  field?: string
  operator?: string
  when?: WorkflowFormRuleCondition
  message?: string
}

export interface WorkflowFormField {
  key: string
  label: string
  type: WorkflowFieldType
  required?: boolean
  default?: unknown
  placeholder?: string
  maxLength?: number
  min?: number
  max?: number
  options?: WorkflowFormOption[]
  optionSource?: WorkflowOptionSource
  span?: number
  rowKey?: string
  columns?: WorkflowFormField[]
  fields?: WorkflowFormField[]
  content?: string
  help?: WorkflowFormHelp
  minRows?: number
  maxRows?: number
  minVisibleRows?: number
  maxVisibleRows?: number
  rules?: WorkflowFormRule[]
}

export interface WorkflowFieldPermission {
  field: string
  access: WorkflowFieldAccess
  actions?: WorkflowFieldAction[]
}

export interface WorkflowInitiatorConfig {
  scope: 'all' | 'specified' | string
  userIds?: number[]
  departmentIds?: number[]
  excludedUserIds?: number[]
}

export type WorkflowStartAvailabilityMode = 'always' | 'fixed' | 'weekly' | 'monthly'
export type WorkflowStartAvailabilityStatus = 'available' | 'not_started' | 'expired' | 'outside_window'
export type WorkflowStartLimitMode = 'unlimited' | 'limited'
export type WorkflowStartLimitPeriod = 'total' | 'day' | 'week' | 'month' | 'availability'

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

export interface WorkflowPublishedNodePosition {
  x: number
  y: number
}

export interface WorkflowPublishedNode {
  id: string
  type: string
  name: string
  position?: WorkflowPublishedNodePosition
  approvalMode?: string
  gatewayMode?: string
  assigneeDisplay?: string
}

export interface WorkflowPublishedEdge {
  id: string
  source: string
  target: string
  sourceHandle?: string
  targetHandle?: string
  name?: string
  default?: boolean
}

export interface WorkflowPublishedDefinition {
  id: number
  key: string
  name: string
  description: string
  category: string
  logoUrl?: string
  version: number
  form: WorkflowFormField[]
  fieldPermissions: Record<string, WorkflowFieldPermission[]>
  startNodeId: string
  initiator: WorkflowInitiatorConfig
  availability: WorkflowStartAvailabilityConfig
  availabilityStatus: WorkflowStartAvailabilityStatus
  startLimit: WorkflowStartLimitConfig
  startLimitStatus: WorkflowStartLimitStatus
  nodes?: WorkflowPublishedNode[]
  edges?: WorkflowPublishedEdge[]
}

export type WorkflowInstanceStatus
  = | 'running'
    | 'completed'
    | 'rejected'
    | 'withdrawn'
    | 'cancelled'
    | string

export type WorkflowTaskStatus
  = | 'pending'
    | 'waiting'
    | 'approved'
    | 'rejected'
    | 'returned'
    | 'submitted'
    | 'cancelled'
    | string

export type WorkflowNodeProgressStatus
  = | 'completed'
    | 'processing'
    | 'not_started'
    | 'skipped'
    | 'returned'
    | 'terminated'

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
  currentNodeNames: string[]
  currentAssigneeNames: string[]
  status: WorkflowInstanceStatus
  startTime: number
  endTime: number
  formRevision: number
}

export interface WorkflowFormRevisionCapability {
  allowed: boolean
  revision: number
  fieldPermissions: WorkflowFieldPermission[]
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
  action: string
  comment: string
  images?: WorkflowAttachment[]
  handledBy: string
  handledByName: string
  handledAt: number
}

export interface WorkflowTokenSummary {
  id: string
  nodeId: string
  status: string
  branchGroup: string
  branchTotal: number
}

export interface WorkflowHistorySummary {
  id: string
  eventType: string
  nodeId: string
  taskId: string
  actorId: string
  actorName?: string
  message: string
  images?: WorkflowAttachment[]
  eventTime: number
}

export interface WorkflowNodeProgressSummary {
  nodeId: string
  nodeName: string
  nodeType: string
  gatewayMode?: string
  status: WorkflowNodeProgressStatus
}

export interface WorkflowReminderPolicy {
  cooldownSeconds: number
  dailyLimit: number
}

export interface WorkflowReminderNode {
  nodeId: string
  nodeName: string
  assigneeNames: string[]
  assigneeCount: number
  canRemind: boolean
  blockedReason: '' | 'cooldown' | 'daily_limit' | string
  lastRemindedAt: number
  nextAllowedAt: number
  todayCount: number
  remainingCount: number
}

export interface WorkflowInstanceDetail {
  instance: WorkflowInstanceSummary
  variables: Record<string, unknown>
  form: WorkflowFormField[]
  formData: WorkflowFormData
  fieldPermissions: Record<string, WorkflowFieldPermission[]>
  startNodeId: string
  nodeTypes: Record<string, string>
  nodes?: WorkflowPublishedNode[]
  edges?: WorkflowPublishedEdge[]
  tokens: WorkflowTokenSummary[]
  tasks: WorkflowTaskSummary[]
  history: WorkflowHistorySummary[]
  nodeProgress?: WorkflowNodeProgressSummary[]
  userNames: Record<string, string>
  reminderPolicy: WorkflowReminderPolicy
  reminderNodes: WorkflowReminderNode[]
  formRevision: WorkflowFormRevisionCapability
}

export interface WorkflowInstanceList {
  list: WorkflowInstanceSummary[]
  total: number
  page: number
  pageSize: number
}

export interface WorkflowOverview {
  pending: number
  handled: number
  started: number
  copied: number
}

export interface WorkflowTaskList {
  list: WorkflowTaskSummary[]
  total: number
  page: number
  pageSize: number
}

export interface WorkflowMutationTask {
  id: string
  nodeId: string
  nodeName: string
  assigneeId: string
  status: string
}

export interface WorkflowMutationResult {
  instanceId: string
  status: WorkflowInstanceStatus
  variables: Record<string, unknown>
  formData: WorkflowFormData
  formRevision: number
  pendingTasks: WorkflowMutationTask[]
}

export interface WorkflowRemindInstanceRequest {
  nodeId: string
}

export interface WorkflowRemindInstanceResult {
  nodeId: string
  nodeName: string
  remindedCount: number
  remindedAt: number
  nextAllowedAt: number
  remainingCount: number
}

export interface WorkflowStartRequest {
  definitionId: number
  definitionVersion: number
  businessType?: string
  businessKey?: string
  variables?: Record<string, unknown>
  formData: WorkflowFormData
}

export interface WorkflowStartDraft {
  definitionId: number
  definitionVersion: number
  formData: WorkflowFormData
  updatedAt: number
}

export interface WorkflowSaveStartDraftRequest {
  definitionVersion: number
  formData: WorkflowFormData
}

export interface WorkflowCompleteTaskRequest {
  action: 'approve' | 'reject' | 'return' | 'submit'
  comment?: string
  images?: WorkflowAttachment[]
  returnTargetNodeId?: string
  variables?: Record<string, unknown>
  formData?: WorkflowFormData
}

export interface WorkflowCommentRequest {
  comment?: string
  images?: WorkflowAttachment[]
  notification?: WorkflowCommentNotificationRequest
}

export type WorkflowNotificationChannel = 'in_app' | 'dingtalk_oa'

export interface WorkflowCommentNotificationRequest {
  userIds: string[]
  channels: WorkflowNotificationChannel[]
}

export interface WorkflowReviseFormRequest {
  expectedRevision: number
  formData: WorkflowFormData
  reason: string
  notification?: WorkflowCommentNotificationRequest
}

export interface WorkflowInstanceQuery {
  definitionId?: number
  definitionName?: string
  definitionCategory?: string
  starterName?: string
  status?: string
  businessType?: string
  businessKey?: string
  scope?: 'started' | 'handled' | 'copied'
  startTimeFrom?: number
  startTimeTo?: number
  endTimeFrom?: number
  endTimeTo?: number
  page?: number
  pageSize?: number
}

export type WorkflowSummaryExportFormat = 'pdf' | 'xlsx' | 'docx'

export interface WorkflowSummaryQuery {
  definitionId?: number
  definitionName?: string
  definitionVersion?: number
  starterName?: string
  status?: string
  startTimeFrom?: number
  startTimeTo?: number
  endTimeFrom?: number
  endTimeTo?: number
  page?: number
  pageSize?: 20 | 50
}

export interface WorkflowTaskQuery {
  instanceId?: string
  status?: string
  definitionName?: string
  definitionCategory?: string
  starterName?: string
  startTimeFrom?: number
  startTimeTo?: number
  page?: number
  pageSize?: number
}
