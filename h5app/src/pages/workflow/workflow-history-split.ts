export const WORKFLOW_HISTORY_SPLIT_MIN = 25
export const WORKFLOW_HISTORY_SPLIT_MAX = 75
export const WORKFLOW_HISTORY_SPLIT_DEFAULT = 50

export function clampWorkflowHistorySplit(value: number) {
  const normalized = Number(value)
  if (!Number.isFinite(normalized))
    return WORKFLOW_HISTORY_SPLIT_DEFAULT
  return Math.min(WORKFLOW_HISTORY_SPLIT_MAX, Math.max(WORKFLOW_HISTORY_SPLIT_MIN, normalized))
}

export function workflowHistorySplitFromPointer(clientY: number, containerTop: number, containerHeight: number) {
  const height = Number(containerHeight)
  if (!Number.isFinite(height) || height <= 0)
    return WORKFLOW_HISTORY_SPLIT_DEFAULT
  const percentage = ((Number(clientY) - Number(containerTop)) / height) * 100
  return Math.round(clampWorkflowHistorySplit(percentage) * 10) / 10
}

export function workflowHistorySplitFromKey(current: number, key: string, largeStep: boolean) {
  const step = largeStep ? 10 : 4
  if (key === 'ArrowUp')
    return clampWorkflowHistorySplit(current - step)
  if (key === 'ArrowDown')
    return clampWorkflowHistorySplit(current + step)
  if (key === 'Home')
    return WORKFLOW_HISTORY_SPLIT_MIN
  if (key === 'End')
    return WORKFLOW_HISTORY_SPLIT_MAX
  return null
}
