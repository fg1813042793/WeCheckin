export type AuditLogCategory = 'login' | 'operation'
export type AuditLogSource = 'admin' | 'client' | 'h5app' | 'system'
export type AuditLogOutcome = 'success' | 'failure' | 'blocked' | 'partial' | 'accepted' | 'unknown'
export type AuditActorType = 'admin' | 'user' | 'system' | 'anonymous'
export type AuditSnapshotStatus = 'pending' | 'recorded' | 'unchanged' | 'not_applicable' | 'not_configured' | 'operation_failed' | 'unavailable' | 'legacy_unknown'

export interface AuditLogChange {
  field: string
  before?: unknown
  after?: unknown
}

// Enrichment is optional for historical records; missing outcomes remain unknown.
export interface AuditLogRecord {
  id: number | string
  type?: number | string
  content?: string
  adminId?: number | null
  adminName?: string
  adminDesc?: string
  LOG_ADD_IP?: string
  _createTime?: number | string | null
  eventId?: string
  category?: AuditLogCategory
  source?: AuditLogSource
  module?: string
  action?: string
  outcome?: AuditLogOutcome
  reasonCode?: string
  reason?: string
  actorType?: AuditActorType
  attemptedAccount?: string
  targetType?: string
  targetId?: string | number
  targetName?: string
  targetIds?: string[]
  requestedCount?: number
  affectedCount?: number
  requestId?: string
  method?: string
  path?: string
  durationMs?: number
  userAgent?: string
  browser?: string
  deviceType?: string
  platformType?: string
  changes?: AuditLogChange[] | null
  snapshotStatus?: AuditSnapshotStatus
  snapshotReason?: string
  relatedType?: string
  relatedId?: string | number
}

export interface AuditLogQuery {
  search?: string
  source?: AuditLogSource
  module?: string
  action?: string
  outcome?: AuditLogOutcome
  actorId?: string
  ip?: string
  startTime?: string
  endTime?: string
  page: number
  pageSize: number
}

export interface AuditLogPage {
  list: AuditLogRecord[]
  total: number
  page: number
  pageSize: number
}

export interface AuditLogDeleteResult {
  deletedCount: number
}

export interface AuditLogFilters {
  search: string
  source: AuditLogSource | ''
  module: string
  action: string
  outcome: AuditLogOutcome | ''
  actorId: string
  ip: string
  range: [string, string] | null
}
