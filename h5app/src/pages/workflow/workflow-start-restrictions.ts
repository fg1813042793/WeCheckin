import type {
  WorkflowPublishedDefinition,
  WorkflowStartAvailabilityConfig,
  WorkflowStartLimitConfig,
  WorkflowStartLimitStatus,
} from '../../types/workflow'

export interface WorkflowStartRestrictionItem {
  key: 'availability' | 'limit' | 'scope' | 'reset'
  label: string
  value: string
}

export interface WorkflowStartRestrictionSummary {
  allowed: boolean
  statusLabel: string
  statusTone: 'success' | 'warning' | 'danger'
  reason: string
  items: WorkflowStartRestrictionItem[]
}

const weekdayLabels: Record<number, string> = {
  1: '一',
  2: '二',
  3: '三',
  4: '四',
  5: '五',
  6: '六',
  7: '日',
}

const periodLabels: Record<string, string> = {
  total: '累计',
  day: '每天',
  week: '每周',
  month: '每月',
  availability: '每个开放周期',
}

function positiveInteger(value: unknown, fallback = 0) {
  const numberValue = Number(value)
  return Number.isFinite(numberValue) && numberValue > 0 ? Math.floor(numberValue) : fallback
}

function formatDateTime(timestamp: number | undefined, timezone = 'Asia/Shanghai') {
  if (!timestamp || !Number.isFinite(timestamp))
    return ''

  const parts = new Intl.DateTimeFormat('zh-CN', {
    timeZone: timezone,
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    hourCycle: 'h23',
  }).formatToParts(new Date(timestamp))
  const valueOf = (type: Intl.DateTimeFormatPartTypes) => parts.find(part => part.type === type)?.value || ''
  return `${valueOf('year')}-${valueOf('month')}-${valueOf('day')} ${valueOf('hour')}:${valueOf('minute')}`
}

function effectiveDateSuffix(availability: WorkflowStartAvailabilityConfig) {
  const start = availability.effectiveStartDate?.trim()
  const end = availability.effectiveEndDate?.trim()
  if (start && end)
    return `（${start} 至 ${end}）`
  if (start)
    return `（${start} 起）`
  if (end)
    return `（截至 ${end}）`
  return ''
}

function dailyTimeRange(availability: WorkflowStartAvailabilityConfig) {
  const start = availability.dailyStartTime?.trim()
  const end = availability.dailyEndTime?.trim()
  if (start && end)
    return ` ${start}-${end}`
  if (start)
    return ` ${start} 起`
  if (end)
    return ` 截至 ${end}`
  return ''
}

function availabilityText(availability: WorkflowStartAvailabilityConfig) {
  const timezone = availability.timezone || 'Asia/Shanghai'
  switch (availability.mode) {
    case 'fixed': {
      const start = formatDateTime(availability.startsAt, timezone)
      const end = formatDateTime(availability.endsAt, timezone)
      if (start && end)
        return `${start} 至 ${end}`
      if (start)
        return `${start} 起`
      if (end)
        return `截至 ${end}`
      return '指定时间段'
    }
    case 'weekly': {
      const weekdays = [...new Set(availability.weekdays || [])]
        .sort((left, right) => left - right)
        .map(day => weekdayLabels[day])
        .filter(Boolean)
      const days = weekdays.length > 0 ? weekdays.join('、周') : '未设置'
      return `每周${days}${dailyTimeRange(availability)}${effectiveDateSuffix(availability)}`
    }
    case 'monthly': {
      const monthDays = [...new Set(availability.monthDays || [])]
        .filter(day => Number.isInteger(day) && day > 0 && day <= 31)
        .sort((left, right) => left - right)
        .map(day => `${day}日`)
      if (availability.lastDayOfMonth)
        monthDays.push('最后一天')
      const days = monthDays.length > 0 ? monthDays.join('、') : '未设置日期'
      return `每月${days}${dailyTimeRange(availability)}${effectiveDateSuffix(availability)}`
    }
    case 'monthly_window': {
      const startDay = positiveInteger(availability.windowStartDayFromEnd, 1)
      const endDay = positiveInteger(availability.windowEndDay, 1)
      const startTime = availability.windowStartTime?.trim() || '00:00'
      const endTime = availability.windowEndTime?.trim() || '23:59'
      return `每月倒数第${startDay}天 ${startTime} 至次月${endDay}日 ${endTime}${effectiveDateSuffix(availability)}`
    }
    case 'always':
    default:
      return '长期有效'
  }
}

function limitText(limit: WorkflowStartLimitConfig, status: WorkflowStartLimitStatus) {
  if (limit.mode !== 'limited')
    return '不限次数'

  const period = periodLabels[limit.period || 'total'] || '每个周期'
  const maxCount = positiveInteger(limit.maxCount)
  const usedCount = Math.max(0, positiveInteger(status.usedCount))
  const remainingCount = Math.max(0, positiveInteger(status.remainingCount))
  return `${period}最多 ${maxCount} 次，已用 ${usedCount} 次，剩余 ${remainingCount} 次`
}

function currentStatus(definition: WorkflowPublishedDefinition) {
  switch (definition.availabilityStatus) {
    case 'not_started':
      return {
        allowed: false,
        statusLabel: '尚未开放',
        statusTone: 'warning' as const,
        reason: '流程尚未到允许发起时间',
      }
    case 'expired':
      return {
        allowed: false,
        statusLabel: '开放时间已结束',
        statusTone: 'danger' as const,
        reason: '流程已超过允许发起时间',
      }
    case 'outside_window':
      return {
        allowed: false,
        statusLabel: '当前时段不可发起',
        statusTone: 'warning' as const,
        reason: '当前不在流程允许发起时间内',
      }
    default:
      if (definition.startLimitStatus?.allowed === false) {
        return {
          allowed: false,
          statusLabel: '次数已用完',
          statusTone: 'danger' as const,
          reason: '当前周期的流程发起次数已用完',
        }
      }
      return {
        allowed: true,
        statusLabel: '当前可发起',
        statusTone: 'success' as const,
        reason: '',
      }
  }
}

export function workflowStartRestrictionSummary(
  definition: WorkflowPublishedDefinition,
): WorkflowStartRestrictionSummary {
  const status = currentStatus(definition)
  const items: WorkflowStartRestrictionItem[] = [
    { key: 'availability', label: '允许发起时间', value: availabilityText(definition.availability) },
    { key: 'limit', label: '发起次数', value: limitText(definition.startLimit, definition.startLimitStatus) },
    { key: 'scope', label: '发起范围', value: '当前账号可发起' },
  ]

  if (definition.startLimit.mode === 'limited'
    && definition.startLimit.period !== 'total'
    && definition.startLimitStatus.resetsAt) {
    const resetTime = formatDateTime(
      definition.startLimitStatus.resetsAt,
      definition.availability.timezone || 'Asia/Shanghai',
    )
    if (resetTime)
      items.push({ key: 'reset', label: '次数重置时间', value: resetTime })
  }

  return { ...status, items }
}
