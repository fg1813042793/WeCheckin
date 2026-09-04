export type HandlerType = 'go' | 'workflow' | 'http' | 'shell' | 'sql'
export type CronPrecision = 'minute' | 'second'
export type MisfirePolicy = 'skip' | 'fire_once' | 'catch_up'
export type ConcurrencyPolicy = 'skip' | 'queue_once' | 'allow'

export interface ScheduledTask {
  id: number
  code: string
  name: string
  description: string
  handlerType: HandlerType
  handlerConfigJson: string
  cronExpression: string
  cronPrecision: CronPrecision
  timezone: string
  enabled: number
  misfirePolicy: MisfirePolicy
  maxCatchUp: number
  concurrencyPolicy: ConcurrencyPolicy
  timeoutSeconds: number
  maxRetries: number
  retryBackoffJson: string
  lastScheduledAt: number
  nextRunAt: number
  version: number
  addTime: number
  editTime: number
}

export interface ScheduledTaskPayload {
  code: string
  name: string
  description: string
  handlerType: HandlerType
  handlerConfigJson: Record<string, unknown>
  cronExpression: string
  cronPrecision: CronPrecision
  timezone: string
  enabled: boolean
  misfirePolicy: MisfirePolicy
  maxCatchUp: number
  concurrencyPolicy: ConcurrencyPolicy
  timeoutSeconds: number
  maxRetries: number
  retryBackoffJson: Record<string, unknown>
  version?: number
}

export interface HandlerMetadata {
  type: HandlerType
  name: string
  description: string
  riskLevel: 'low' | 'medium' | 'high' | 'critical'
  configSchema: {
    properties?: Record<string, {
      enum?: string[]
      'x-enum-labels'?: Record<string, string>
    }>
  }
}

export interface CronOccurrence {
  utcMillis: number
  localTime: string
}

export interface ScheduledTaskRun {
  id: string
  runKey: string
  taskId: number
  parentRunId: string
  triggerType: string
  status: string
  scheduledAt: number
  coalescedCount: number
  attempt: number
  workerId: string
  queuedAt: number
  startedAt: number
  finishedAt: number
  heartbeatAt: number
  nextRetryAt: number
  cancelRequestedAt: number
  resultSummary: string
  errorCode: string
  errorSummary: string
  addTime: number
}

export interface ScheduledTaskRunLog {
  id: number
  runId: string
  sequence: number
  level: string
  stage: string
  content: string
  logTime: number
}

export interface ScheduledTaskRunDetail {
  run: ScheduledTaskRun
  logs: ScheduledTaskRunLog[]
}

export interface ScheduledTaskWorker {
  workerId: string
  role: string
  version: string
  startedAt: number
  lastHeartbeat: number
  currentRuns: number
  workerCount: number
}

export interface PageData<T> {
  list: T[]
  total: number
  page: number
  pageSize: number
}
