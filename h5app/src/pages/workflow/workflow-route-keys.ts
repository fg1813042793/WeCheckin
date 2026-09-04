const workflowStartPrefix = 'workflow:start:'
const workflowInstancePrefix = 'workflow:instance:'
const workflowTaskPrefix = 'workflow:task:'
const workflowTaskInstanceSeparator = ':instance:'

export function workflowStartContentKey(definitionId: number) {
  return `${workflowStartPrefix}${definitionId}`
}

export function workflowDefinitionIdFromContentKey(key: string) {
  if (!key.startsWith(workflowStartPrefix))
    return 0
  const value = Number(key.slice(workflowStartPrefix.length))
  return Number.isInteger(value) && value > 0 ? value : 0
}

export function workflowInstanceContentKey(instanceId: string) {
  const normalizedInstanceId = String(instanceId || '').trim()
  return normalizedInstanceId
    ? `${workflowInstancePrefix}${encodeURIComponent(normalizedInstanceId)}`
    : ''
}

export function workflowInstanceIdFromContentKey(key: string) {
  if (!key.startsWith(workflowInstancePrefix))
    return ''
  return decodeWorkflowRouteValue(key.slice(workflowInstancePrefix.length))
}

export function workflowTaskContentKey(taskId: string, instanceId: string) {
  const normalizedTaskId = String(taskId || '').trim()
  const normalizedInstanceId = String(instanceId || '').trim()
  if (!normalizedTaskId || !normalizedInstanceId)
    return ''
  return `${workflowTaskPrefix}${encodeURIComponent(normalizedTaskId)}${workflowTaskInstanceSeparator}${encodeURIComponent(normalizedInstanceId)}`
}

export function workflowTaskIdFromContentKey(key: string) {
  const payload = workflowTaskPayloadFromContentKey(key)
  return payload ? decodeWorkflowRouteValue(payload.taskId) : ''
}

export function workflowTaskInstanceIdFromContentKey(key: string) {
  const payload = workflowTaskPayloadFromContentKey(key)
  return payload ? decodeWorkflowRouteValue(payload.instanceId) : ''
}

function workflowTaskPayloadFromContentKey(key: string) {
  if (!key.startsWith(workflowTaskPrefix))
    return null
  const payload = key.slice(workflowTaskPrefix.length)
  const separatorIndex = payload.indexOf(workflowTaskInstanceSeparator)
  if (separatorIndex <= 0)
    return null
  return {
    taskId: payload.slice(0, separatorIndex),
    instanceId: payload.slice(separatorIndex + workflowTaskInstanceSeparator.length),
  }
}

function decodeWorkflowRouteValue(value: string) {
  try {
    return decodeURIComponent(value).trim()
  }
  catch {
    return ''
  }
}
