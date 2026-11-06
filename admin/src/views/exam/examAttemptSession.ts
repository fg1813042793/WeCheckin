type UnknownRecord = Record<string, unknown>

export interface ExamCountdownInput {
  now: number
  startAt: number
  timeLimit: number
  maxSubmitMinutes: number
  endTime: number
  submitting: boolean
  submitted: boolean
}

export interface ExamCountdownState {
  remaining: number
  shouldSchedule: boolean
  shouldAutoSubmit: boolean
}

function asRecord(value: unknown): UnknownRecord | null {
  return value && typeof value === 'object' && !Array.isArray(value) ? value as UnknownRecord : null
}

function finiteNumber(value: unknown): number {
  const number = Number(value)
  return Number.isFinite(number) ? number : 0
}

function parseSettings(value: unknown): UnknownRecord {
  if (typeof value === 'string') {
    try { return asRecord(JSON.parse(value)) || {} } catch { return {} }
  }
  return asRecord(value) || {}
}

function attemptDeadline(startAt: number, timeLimit: number, maxSubmitMinutes: number, endTime: number): number {
  if (startAt <= 0) return 0
  const deadlines: number[] = []
  if (timeLimit > 0) deadlines.push(startAt + timeLimit * 60_000)
  if (maxSubmitMinutes > 0) deadlines.push(startAt + maxSubmitMinutes * 60_000)
  if (endTime > 0) deadlines.push(endTime)
  return deadlines.length ? Math.min(...deadlines) : 0
}

export function examCountdownState(input: ExamCountdownInput): ExamCountdownState {
  const deadline = attemptDeadline(input.startAt, input.timeLimit, input.maxSubmitMinutes, input.endTime)
  if (deadline <= 0) return { remaining: 0, shouldSchedule: false, shouldAutoSubmit: false }
  const remaining = Math.max(0, deadline - input.now)
  return {
    remaining,
    shouldSchedule: remaining > 0,
    shouldAutoSubmit: remaining <= 0 && !input.submitting && !input.submitted,
  }
}

export function shouldRenewPublicExamSession(detailValue: unknown, requestedSession: string, now = Date.now()): boolean {
  if (!requestedSession) return false
  const detail = asRecord(detailValue)
  if (!detail || String(detail.session || '') !== requestedSession) return false
  const exam = asRecord(detail.exam) || detail
  if (!Object.prototype.hasOwnProperty.call(exam, 'visibility') || !Object.prototype.hasOwnProperty.call(exam, 'maxAttempts')) return false
  if (finiteNumber(exam.visibility) !== 0 || finiteNumber(exam.maxAttempts) !== 0) return false
  const settings = parseSettings(detail.settings ?? exam.settings)
  if (finiteNumber(settings.deviceLimit) > 0 || finiteNumber(settings.ipLimit) > 0) return false
  const endTime = finiteNumber(exam.endTime)
  if (endTime > 0 && now >= endTime) return false
  const deadline = attemptDeadline(
    finiteNumber(detail.startAt),
    finiteNumber(settings.timeLimit),
    finiteNumber(settings.maxSubmitMinutes),
    endTime,
  )
  return deadline > 0 && now >= deadline
}
