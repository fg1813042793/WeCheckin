export interface WorkflowNotificationOpenInput {
  sourceType: string
  sourceId: string
  type?: string
  title: string
}

export interface WorkflowNotificationDynamicTab {
  key: string
  label: string
  icon: string
  path: string
}

export interface WorkflowNotificationOpenDependencies {
  markRead: () => Promise<boolean>
  workflowInstanceContentKey: (sourceID: string) => string
  workflowFormRevisionDetailContentKey: (sourceID: string) => string
  openDynamicTab: (tab: WorkflowNotificationDynamicTab) => boolean | Promise<boolean>
  closePanel: () => void
}

const workflowNotificationLabels: Record<string, string> = {
  instance_form_revision_requested: '修订待确认',
  instance_form_revision_approved: '修订已生效',
  instance_form_revision_rejected: '修订已驳回',
  instance_form_revision_cancelled: '修订已取消',
  instance_form_revision_conflict: '修订冲突',
}

export function isWorkflowNotification(
  notification: Pick<WorkflowNotificationOpenInput, 'sourceType'>,
) {
  return notification.sourceType === 'workflow_instance'
    || notification.sourceType === 'workflow_form_revision'
}

export function workflowNotificationLabel(
  notification: Pick<WorkflowNotificationOpenInput, 'type'>,
) {
  return workflowNotificationLabels[String(notification.type || '').trim()] || '流程消息'
}

export function workflowNotificationDynamicTab(
  notification: WorkflowNotificationOpenInput,
  dependencies: Pick<
    WorkflowNotificationOpenDependencies,
    'workflowInstanceContentKey' | 'workflowFormRevisionDetailContentKey'
  >,
): WorkflowNotificationDynamicTab | null {
  const sourceID = String(notification.sourceId || '').trim()
  if (!sourceID)
    return null

  const revision = notification.sourceType === 'workflow_form_revision'
  const key = revision
    ? dependencies.workflowFormRevisionDetailContentKey(sourceID)
    : notification.sourceType === 'workflow_instance'
      ? dependencies.workflowInstanceContentKey(sourceID)
      : ''
  if (!key)
    return null
  return {
    key,
    label: String(notification.title || '').trim() || (revision ? '表单修订详情' : '流程详情'),
    icon: revision ? 'edit-pen' : 'file-text',
    path: `/pages/index/index?view=${encodeURIComponent(key)}`,
  }
}

export async function openWorkflowNotification(
  notification: WorkflowNotificationOpenInput,
  dependencies: WorkflowNotificationOpenDependencies,
) {
  if (!isWorkflowNotification(notification))
    return false

  const marked = await dependencies.markRead()
  if (!marked)
    return true

  const tab = workflowNotificationDynamicTab(notification, dependencies)
  if (!tab)
    return true
  const opened = await dependencies.openDynamicTab(tab)
  if (opened)
    dependencies.closePanel()
  return true
}
