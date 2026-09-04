export interface WorkflowHistoryDateFilters {
  startDateFrom?: string
  startDateTo?: string
  endDateFrom?: string
  endDateTo?: string
}

export interface WorkflowHistoryTimeQuery {
  startTimeFrom?: number
  startTimeTo?: number
  endTimeFrom?: number
  endTimeTo?: number
}

function dateTimestamp(value: string | undefined, endOfDay: boolean) {
  if (!value)
    return undefined
  const match = /^(\d{4})-(\d{2})-(\d{2})$/.exec(value)
  if (!match)
    return undefined
  const year = Number(match[1])
  const month = Number(match[2])
  const day = Number(match[3])
  const date = new Date(year, month - 1, day)
  if (date.getFullYear() !== year || date.getMonth() !== month - 1 || date.getDate() !== day)
    return undefined
  if (endOfDay)
    return new Date(year, month - 1, day + 1).getTime() - 1
  return date.getTime()
}

export function buildWorkflowHistoryTimeQuery(filters: WorkflowHistoryDateFilters): WorkflowHistoryTimeQuery | null {
  const startTimeFrom = dateTimestamp(filters.startDateFrom, false)
  const startTimeTo = dateTimestamp(filters.startDateTo, true)
  const endTimeFrom = dateTimestamp(filters.endDateFrom, false)
  const endTimeTo = dateTimestamp(filters.endDateTo, true)

  if ((filters.startDateFrom && startTimeFrom === undefined)
    || (filters.startDateTo && startTimeTo === undefined)
    || (filters.endDateFrom && endTimeFrom === undefined)
    || (filters.endDateTo && endTimeTo === undefined)
    || (startTimeFrom !== undefined && startTimeTo !== undefined && startTimeFrom > startTimeTo)
    || (endTimeFrom !== undefined && endTimeTo !== undefined && endTimeFrom > endTimeTo)) {
    return null
  }

  return {
    ...(startTimeFrom === undefined ? {} : { startTimeFrom }),
    ...(startTimeTo === undefined ? {} : { startTimeTo }),
    ...(endTimeFrom === undefined ? {} : { endTimeFrom }),
    ...(endTimeTo === undefined ? {} : { endTimeTo }),
  }
}
