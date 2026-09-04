import type { WorkflowInstanceStatus, WorkflowTaskStatus } from './types'

export type WorkflowStatusTagType = 'primary' | 'success' | 'warning' | 'info' | 'danger'

export interface WorkflowStatusMeta {
  label: string
  type: WorkflowStatusTagType
}

const INSTANCE_STATUS_META: Record<WorkflowInstanceStatus, WorkflowStatusMeta> = {
  running: { label: '审批中', type: 'warning' },
  completed: { label: '已完成', type: 'success' },
  rejected: { label: '已驳回', type: 'danger' },
  cancelled: { label: '已取消', type: 'info' },
}

const TASK_STATUS_META: Record<WorkflowTaskStatus, WorkflowStatusMeta> = {
  waiting: { label: '待激活', type: 'warning' },
  pending: { label: '待处理', type: 'warning' },
  completed: { label: '已完成', type: 'success' },
  approved: { label: '已通过', type: 'success' },
  rejected: { label: '已驳回', type: 'danger' },
  returned: { label: '已退回', type: 'warning' },
  cancelled: { label: '已取消', type: 'info' },
}

function unknownStatusMeta(status?: string): WorkflowStatusMeta {
  return { label: status || '未知', type: 'info' }
}

export function workflowInstanceStatusMeta(status: WorkflowInstanceStatus): WorkflowStatusMeta {
  return INSTANCE_STATUS_META[status] || unknownStatusMeta(status)
}

export function workflowTaskStatusMeta(status: WorkflowTaskStatus): WorkflowStatusMeta {
  return TASK_STATUS_META[status] || unknownStatusMeta(status)
}
