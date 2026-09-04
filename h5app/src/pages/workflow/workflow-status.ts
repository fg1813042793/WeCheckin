import type {
  WorkflowInstanceStatus,
  WorkflowNodeProgressStatus,
  WorkflowTaskStatus,
} from '@/types/workflow'

export type WorkflowStatusTagType = 'primary' | 'success' | 'error' | 'warning' | 'info'

export interface WorkflowStatusMeta {
  label: string
  type: WorkflowStatusTagType
}

const INSTANCE_STATUS_META: Record<string, WorkflowStatusMeta> = {
  running: { label: '审批中', type: 'warning' },
  completed: { label: '已完成', type: 'success' },
  rejected: { label: '已驳回', type: 'error' },
  withdrawn: { label: '已撤回', type: 'info' },
  cancelled: { label: '已取消', type: 'info' },
}

const TASK_STATUS_META: Record<string, WorkflowStatusMeta> = {
  waiting: { label: '待激活', type: 'warning' },
  pending: { label: '待处理', type: 'warning' },
  approved: { label: '已通过', type: 'success' },
  submitted: { label: '已提交', type: 'success' },
  completed: { label: '已完成', type: 'success' },
  rejected: { label: '已驳回', type: 'error' },
  returned: { label: '已退回', type: 'warning' },
  cancelled: { label: '已取消', type: 'info' },
}

const NODE_PROGRESS_STATUS_META: Record<string, WorkflowStatusMeta> = {
  completed: { label: '已完成', type: 'success' },
  processing: { label: '处理中', type: 'warning' },
  not_started: { label: '未开始', type: 'info' },
  skipped: { label: '已跳过', type: 'info' },
  returned: { label: '已退回', type: 'warning' },
  terminated: { label: '已终止', type: 'error' },
}

function unknownStatusMeta(status?: string): WorkflowStatusMeta {
  return { label: status || '未知', type: 'info' }
}

export function workflowInstanceStatusMeta(status: WorkflowInstanceStatus = ''): WorkflowStatusMeta {
  return INSTANCE_STATUS_META[status] || unknownStatusMeta(status)
}

export function workflowTaskStatusMeta(status: WorkflowTaskStatus = ''): WorkflowStatusMeta {
  return TASK_STATUS_META[status] || unknownStatusMeta(status)
}

export function workflowNodeProgressStatusMeta(
  status: WorkflowNodeProgressStatus,
  instanceStatus: WorkflowInstanceStatus = '',
): WorkflowStatusMeta {
  if (status === 'terminated' && (instanceStatus === 'withdrawn' || instanceStatus === 'cancelled'))
    return { label: '已终止', type: 'info' }
  return NODE_PROGRESS_STATUS_META[status] || unknownStatusMeta(status)
}
