interface WorkflowTaskAssignee {
  assigneeId?: unknown
}

function normalizedText(value: unknown) {
  return String(value ?? '').trim()
}

export function workflowTaskTabTitle(starterName: unknown, definitionName: unknown) {
  const starter = normalizedText(starterName) || '未知用户'
  const definition = normalizedText(definitionName) || '流程审批'
  return `${starter}-${definition}`
}

export function isWorkflowTaskAssignedToUser(task: WorkflowTaskAssignee | null | undefined, userId: unknown) {
  const assigneeId = normalizedText(task?.assigneeId)
  const currentUserId = normalizedText(userId)
  return Boolean(assigneeId && currentUserId && assigneeId === currentUserId)
}
